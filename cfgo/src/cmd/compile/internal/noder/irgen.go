// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package noder

import (
	"fmt"
	"regexp"
	"sort"

	"cmd/compile/internal/base"
	"cmd/compile/internal/dwarfgen"
	"cmd/compile/internal/ir"
	"cmd/compile/internal/syntax"
	"cmd/compile/internal/typecheck"
	"cmd/compile/internal/types"
	"cmd/compile/internal/types2"
	"cmd/internal/src"
)

var versionErrorRx = regexp.MustCompile(`requires go[0-9]+\.[0-9]+ or later`)

// checkFiles configures and runs the types2 checker on the given
// parsed source files and then returns the result.
func checkFiles(noders []*noder) (posMap, *types2.Package, *types2.Info) {
	if base.SyntaxErrors() != 0 {
		base.ErrorExit()
	}

	// setup and syntax error reporting
	var m posMap
	files := make([]*syntax.File, len(noders))
	for i, p := range noders {
		m.join(&p.posMap)
		files[i] = p.file
	}

	// typechecking
	ctxt := types2.NewContext()
	importer := gcimports{
		ctxt:     ctxt,
		packages: make(map[string]*types2.Package),
	}
	conf := types2.Config{
		Context:            ctxt,
		GoVersion:          base.Flag.Lang,
		IgnoreBranchErrors: true, // parser already checked via syntax.CheckBranches mode
		Error: func(err error) {
			terr := err.(types2.Error)
			msg := terr.Msg
			// if we have a version error, hint at the -lang setting
			if versionErrorRx.MatchString(msg) {
				msg = fmt.Sprintf("%s (-lang was set to %s; check go.mod)", msg, base.Flag.Lang)
			}
			base.ErrorfAt(m.makeXPos(terr.Pos), "%s", msg)
		},
		Importer:               &importer,
		Sizes:                  &gcSizes{},
		OldComparableSemantics: base.Flag.OldComparable, // default is new comparable semantics
	}
	info := &types2.Info{
		StoreTypesInSyntax: true,
		Defs:               make(map[*syntax.Name]types2.Object),
		Uses:               make(map[*syntax.Name]types2.Object),
		Selections:         make(map[*syntax.SelectorExpr]*types2.Selection),
		Implicits:          make(map[syntax.Node]types2.Object),
		Scopes:             make(map[syntax.Node]*types2.Scope),
		Instances:          make(map[*syntax.Name]types2.Instance),
		// expand as needed
	}

	pkg, err := conf.Check(base.Ctxt.Pkgpath, files, info)

	// Check for anonymous interface cycles (#56103).
	if base.Debug.InterfaceCycles == 0 {
		var f cycleFinder
		for _, file := range files {
			syntax.Inspect(file, func(n syntax.Node) bool {
				if n, ok := n.(*syntax.InterfaceType); ok {
					if f.hasCycle(n.GetTypeInfo().Type.(*types2.Interface)) {
						base.ErrorfAt(m.makeXPos(n.Pos()), "invalid recursive type: anonymous interface refers to itself (see https://go.dev/issue/56103)")

						for typ := range f.cyclic {
							f.cyclic[typ] = false // suppress duplicate errors
						}
					}
					return false
				}
				return true
			})
		}
	}

	// Implementation restriction: we don't allow not-in-heap types to
	// be used as type arguments (#54765).
	{
		type nihTarg struct {
			pos src.XPos
			typ types2.Type
		}
		var nihTargs []nihTarg

		for name, inst := range info.Instances {
			for i := 0; i < inst.TypeArgs.Len(); i++ {
				if targ := inst.TypeArgs.At(i); isNotInHeap(targ) {
					nihTargs = append(nihTargs, nihTarg{m.makeXPos(name.Pos()), targ})
				}
			}
		}
		sort.Slice(nihTargs, func(i, j int) bool {
			ti, tj := nihTargs[i], nihTargs[j]
			return ti.pos.Before(tj.pos)
		})
		for _, targ := range nihTargs {
			base.ErrorfAt(targ.pos, "cannot use incomplete (or unallocatable) type as a type argument: %v", targ.typ)
		}
	}

	base.ExitIfErrors()
	if err != nil {
		base.FatalfAt(src.NoXPos, "conf.Check error: %v", err)
	}

	return m, pkg, info
}

// check2 type checks a Go package using types2, and then generates IR
// using the results.
func check2(noders []*noder) {
	m, pkg, info := checkFiles(noders)

	g := irgen{
		target: typecheck.Target,
		self:   pkg,
		info:   info,
		posMap: m,
		objs:   make(map[types2.Object]*ir.Name),
		typs:   make(map[types2.Type]*types.Type),
	}
	g.generate(noders)
}

// Information about sub-dictionary entries in a dictionary
type subDictInfo struct {
	// Call or XDOT node that requires a dictionary.
	callNode ir.Node
	// Saved CallExpr.X node (*ir.SelectorExpr or *InstExpr node) for a generic
	// method or function call, since this node will get dropped when the generic
	// method/function call is transformed to a call on the instantiated shape
	// function. Nil for other kinds of calls or XDOTs.
	savedXNode ir.Node
}

// dictInfo is the dictionary format for an instantiation of a generic function with
// particular shapes. shapeParams, derivedTypes, subDictCalls, itabConvs, and methodExprClosures
// describe the actual dictionary entries in order, and the remaining fields are other info
// needed in doing dictionary processing during compilation.
type dictInfo struct {
	// Types substituted for the type parameters, which are shape types.
	shapeParams []*types.Type
	// All types derived from those typeparams used in the instantiation.
	derivedTypes []*types.Type
	// Nodes in the instantiation that requires a subdictionary. Includes
	// method and function calls (OCALL), function values (OFUNCINST), method
	// values/expressions (OXDOT).
	subDictCalls []subDictInfo
	// Nodes in the instantiation that are a conversion from a typeparam/derived
	// type to a specific interface.
	itabConvs []ir.Node
	// Method expression closures. For a generic type T with method M(arg1, arg2) res,
	// these closures are func(rcvr T, arg1, arg2) res.
	// These closures capture no variables, they are just the generic version of ·f symbols
	// that live in the dictionary instead of in the readonly globals section.
	methodExprClosures []methodExprClosure

	// Mapping from each shape type that substitutes a type param, to its
	// type bound (which is also substituted with shapes if it is parameterized)
	shapeToBound map[*types.Type]*types.Type

	// For type switches on nonempty interfaces, a map from OTYPE entries of
	// HasShape type, to the interface type we're switching from.
	type2switchType map[ir.Node]*types.Type

	startSubDict            int // Start of dict entries for subdictionaries
	startItabConv           int // Start of dict entries for itab conversions
	startMethodExprClosures int // Start of dict entries for closures for method expressions
	dictLen                 int // Total number of entries in dictionary
}

type methodExprClosure struct {
	idx  int    // index in list of shape parameters
	name string // method name
}

// instInfo is information gathered on an shape instantiation of a function.
type instInfo struct {
	fun       *ir.Func // The instantiated function (with body)
	dictParam *ir.Name // The node inside fun that refers to the dictionary param

	dictInfo *dictInfo
}

type irgen struct {
	target *ir.Package
	self   *types2.Package
	info   *types2.Info

	posMap
	objs   map[types2.Object]*ir.Name
	typs   map[types2.Type]*types.Type
	marker dwarfgen.ScopeMarker

	// laterFuncs records tasks that need to run after all declarations
	// are processed.
	laterFuncs []func()
	// haveEmbed indicates whether the current node belongs to file that
	// imports "embed" package.
	haveEmbed bool

	// exprStmtOK indicates whether it's safe to generate expressions or
	// statements yet.
	exprStmtOK bool

	// types which we need to finish, by doing g.fillinMethods.
	typesToFinalize []*typeDelayInfo

	// True when we are compiling a top-level generic function or method. Use to
	// avoid adding closures of generic functions/methods to the target.Decls
	// list.
	topFuncIsGeneric bool

	// The context during type/function/method declarations that is used to
	// uniquely name type parameters. We need unique names for type params so we
	// can be sure they match up correctly between types2-to-types1 translation
	// and types1 importing.
	curDecl string
}

// genInst has the information for creating needed instantiations and modifying
// functions to use instantiations.
type genInst struct {
	dnum int // for generating unique dictionary variables

	// Map from the names of all instantiations to information about the
	// instantiations.
	instInfoMap map[*types.Sym]*instInfo

	// Dictionary syms which we need to finish, by writing out any itabconv
	// or method expression closure entries.
	dictSymsToFinalize []*delayInfo

	// New instantiations created during this round of buildInstantiations().
	newInsts []ir.Node
}

func (g *irgen) later(fn func()) {
	g.laterFuncs = append(g.laterFuncs, fn)
}

type delayInfo struct {
	gf     *ir.Name
	targs  []*types.Type
	sym    *types.Sym
	off    int
	isMeth bool
}

type typeDelayInfo struct {
	typ  *types2.Named
	ntyp *types.Type
}

func (g *irgen) generate(noders []*noder) {
	types.LocalPkg.Name = g.self.Name()
	typecheck.TypecheckAllowed = true

	// Prevent size calculations until we set the underlying type
	// for all package-block defined types.
	types.DeferCheckSize()

	// At this point, types2 has already handled name resolution and
	// type checking. We just need to map from its object and type
	// representations to those currently used by the rest of the
	// compiler. This happens in a few passes.

	// 1. Process all import declarations. We use the compiler's own
	// importer for this, rather than types2's gcimporter-derived one,
	// to handle extensions and inline function bodies correctly.
	//
	// Also, we need to do this in a separate pass, because mappings are
	// instantiated on demand. If we interleaved processing import
	// declarations with other declarations, it's likely we'd end up
	// wanting to map an object/type from another source file, but not
	// yet have the import data it relies on.
	declLists := make([][]syntax.Decl, len(noders))
Outer:
	for i, p := range noders {
		g.pragmaFlags(p.file.Pragma, ir.GoBuildPragma)
		for j, decl := range p.file.DeclList {
			switch decl := decl.(type) {
			case *syntax.ImportDecl:
				g.importDecl(p, decl)
			default:
				declLists[i] = p.file.DeclList[j:]
				continue Outer // no more ImportDecls
			}
		}
	}

	// 2. Process all package-block type declarations. As with imports,
	// we need to make sure all types are properly instantiated before
	// trying to map any expressions that utilize them. In particular,
	// we need to make sure type pragmas are already known (see comment
	// in irgen.typeDecl).
	//
	// We could perhaps instead defer processing of package-block
	// variable initializers and function bodies, like noder does, but
	// special-casing just package-block type declarations minimizes the
	// differences between processing package-block and function-scoped
	// declarations.
	for _, declList := range declLists {
		for _, decl := range declList {
			switch decl := decl.(type) {
			case *syntax.TypeDecl:
				g.typeDecl((*ir.Nodes)(&g.target.Decls), decl)
			}
		}
	}
	types.ResumeCheckSize()

	// 3. Process all remaining declarations.
	for i, declList := range declLists {
		old := g.haveEmbed
		g.haveEmbed = noders[i].importedEmbed
		g.decls((*ir.Nodes)(&g.target.Decls), declList)
		g.haveEmbed = old
	}
	g.exprStmtOK = true

	// 4. Run any "later" tasks. Avoid using 'range' so that tasks can
	// recursively queue further tasks. (Not currently utilized though.)
	for len(g.laterFuncs) > 0 {
		fn := g.laterFuncs[0]
		g.laterFuncs = g.laterFuncs[1:]
		fn()
	}

	if base.Flag.W > 1 {
		for _, n := range g.target.Decls {
			s := fmt.Sprintf("\nafter noder2 %v", n)
			ir.Dump(s, n)
		}
	}

	for _, p := range noders {
		// Process linkname and cgo pragmas.
		p.processPragmas()

		// Double check for any type-checking inconsistencies. This can be
		// removed once we're confident in IR generation results.
		syntax.Crawl(p.file, func(n syntax.Node) bool {
			g.validate(n)
			return false
		})
	}

	if base.Flag.Complete {
		for _, n := range g.target.Decls {
			if fn, ok := n.(*ir.Func); ok {
				if fn.Body == nil && fn.Nname.Sym().Linkname == "" {
					base.ErrorfAt(fn.Pos(), "missing function body")
				}
			}
		}
	}

	// Check for unusual case where noder2 encounters a type error that types2
	// doesn't check for (e.g. notinheap incompatibility).
	base.ExitIfErrors()

	typecheck.DeclareUniverse()

	// Create any needed instantiations of generic functions and transform
	// existing and new functions to use those instantiations.
	BuildInstantiations()

	// Remove all generic functions from g.target.Decl, since they have been
	// used for stenciling, but don't compile. Generic functions will already
	// have been marked for export as appropriate.
	j := 0
	for i, decl := range g.target.Decls {
		if decl.Op() != ir.ODCLFUNC || !decl.Type().HasTParam() {
			g.target.Decls[j] = g.target.Decls[i]
			j++
		}
	}
	g.target.Decls = g.target.Decls[:j]

	base.Assertf(len(g.laterFuncs) == 0, "still have %d later funcs", len(g.laterFuncs))
}

func (g *irgen) unhandled(what string, p poser) {
	base.FatalfAt(g.pos(p), "unhandled %s: %T", what, p)
	panic("unreachable")
}

// delayTransform returns true if we should delay all transforms, because we are
// creating the nodes for a generic function/method.
func (g *irgen) delayTransform() bool {
	return g.topFuncIsGeneric
}

func (g *irgen) typeAndValue(x syntax.Expr) syntax.TypeAndValue {
	tv := x.GetTypeInfo()
	if tv.Type == nil {
		base.FatalfAt(g.pos(x), "missing type for %v (%T)", x, x)
	}
	return tv
}

func (g *irgen) type2(x syntax.Expr) syntax.Type {
	tv := x.GetTypeInfo()
	if tv.Type == nil {
		base.FatalfAt(g.pos(x), "missing type for %v (%T)", x, x)
	}
	return tv.Type
}

// A cycleFinder detects anonymous interface cycles (go.dev/issue/56103).
type cycleFinder struct {
	cyclic map[*types2.Interface]bool
}

// hasCycle reports whether typ is part of an anonymous interface cycle.
func (f *cycleFinder) hasCycle(typ *types2.Interface) bool {
	// We use Method instead of ExplicitMethod to implicitly expand any
	// embedded interfaces. Then we just need to walk any anonymous
	// types, keeping track of *types2.Interface types we visit along
	// the way.
	for i := 0; i < typ.NumMethods(); i++ {
		if f.visit(typ.Method(i).Type()) {
			return true
		}
	}
	return false
}

// visit recursively walks typ0 to check any referenced interface types.
func (f *cycleFinder) visit(typ0 types2.Type) bool {
	for { // loop for tail recursion
		switch typ := typ0.(type) {
		default:
			base.Fatalf("unexpected type: %T", typ)

		case *types2.Basic, *types2.Named, *types2.TypeParam:
			return false // named types cannot be part of an anonymous cycle
		case *types2.Pointer:
			typ0 = typ.Elem()
		case *types2.Array:
			typ0 = typ.Elem()
		case *types2.Chan:
			typ0 = typ.Elem()
		case *types2.Map:
			if f.visit(typ.Key()) {
				return true
			}
			typ0 = typ.Elem()
		case *types2.Slice:
			typ0 = typ.Elem()

		case *types2.Struct:
			for i := 0; i < typ.NumFields(); i++ {
				if f.visit(typ.Field(i).Type()) {
					return true
				}
			}
			return false

		case *types2.Interface:
			// The empty interface (e.g., "any") cannot be part of a cycle.
			if typ.NumExplicitMethods() == 0 && typ.NumEmbeddeds() == 0 {
				return false
			}

			// As an optimization, we wait to allocate cyclic here, after
			// we've found at least one other (non-empty) anonymous
			// interface. This means when a cycle is present, we need to
			// make an extra recursive call to actually detect it. But for
			// most packages, it allows skipping the map allocation
			// entirely.
			if x, ok := f.cyclic[typ]; ok {
				return x
			}
			if f.cyclic == nil {
				f.cyclic = make(map[*types2.Interface]bool)
			}
			f.cyclic[typ] = true
			if f.hasCycle(typ) {
				return true
			}
			f.cyclic[typ] = false
			return false

		case *types2.Signature:
			return f.visit(typ.Params()) || f.visit(typ.Results())
		case *types2.Tuple:
			for i := 0; i < typ.Len(); i++ {
				if f.visit(typ.At(i).Type()) {
					return true
				}
			}
			return false
		}
	}
}
