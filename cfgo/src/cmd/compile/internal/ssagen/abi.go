// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssagen

import (
	"fmt"
	"internal/buildcfg"
	"log"
	"os"
	"strings"

	"cmd/compile/internal/base"
	"cmd/compile/internal/ir"
	"cmd/compile/internal/typecheck"
	"cmd/compile/internal/types"
	"cmd/internal/obj"
)

// SymABIs records information provided by the assembler about symbol
// definition ABIs and reference ABIs.
type SymABIs struct {
	defs map[string]obj.ABI
	refs map[string]obj.ABISet
}

func NewSymABIs() *SymABIs {
	return &SymABIs{
		defs: make(map[string]obj.ABI),
		refs: make(map[string]obj.ABISet),
	}
}

// canonicalize returns the canonical name used for a linker symbol in
// s's maps. Symbols in this package may be written either as "".X or
// with the package's import path already in the symbol. This rewrites
// both to use the full path, which matches compiler-generated linker
// symbol names.
func (s *SymABIs) canonicalize(linksym string) string {
	// If the symbol is already prefixed with "", rewrite it to start
	// with LocalPkg.Prefix.
	//
	// TODO(mdempsky): Have cmd/asm stop writing out symbols like this.
	if strings.HasPrefix(linksym, `"".`) {
		return types.LocalPkg.Prefix + linksym[2:]
	}
	return linksym
}

// ReadSymABIs reads a symabis file that specifies definitions and
// references of text symbols by ABI.
//
// The symabis format is a set of lines, where each line is a sequence
// of whitespace-separated fields. The first field is a verb and is
// either "def" for defining a symbol ABI or "ref" for referencing a
// symbol using an ABI. For both "def" and "ref", the second field is
// the symbol name and the third field is the ABI name, as one of the
// named cmd/internal/obj.ABI constants.
func (s *SymABIs) ReadSymABIs(file string) {
	data, err := os.ReadFile(file)
	if err != nil {
		log.Fatalf("-symabis: %v", err)
	}

	for lineNum, line := range strings.Split(string(data), "\n") {
		lineNum++ // 1-based
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Fields(line)
		switch parts[0] {
		case "def", "ref":
			// Parse line.
			if len(parts) != 3 {
				log.Fatalf(`%s:%d: invalid symabi: syntax is "%s sym abi"`, file, lineNum, parts[0])
			}
			sym, abistr := parts[1], parts[2]
			abi, valid := obj.ParseABI(abistr)
			if !valid {
				log.Fatalf(`%s:%d: invalid symabi: unknown abi "%s"`, file, lineNum, abistr)
			}

			sym = s.canonicalize(sym)

			// Record for later.
			if parts[0] == "def" {
				s.defs[sym] = abi
			} else {
				s.refs[sym] |= obj.ABISetOf(abi)
			}
		default:
			log.Fatalf(`%s:%d: invalid symabi type "%s"`, file, lineNum, parts[0])
		}
	}
}

// GenABIWrappers applies ABI information to Funcs and generates ABI
// wrapper functions where necessary.
func (s *SymABIs) GenABIWrappers() {
	// For cgo exported symbols, we tell the linker to export the
	// definition ABI to C. That also means that we don't want to
	// create ABI wrappers even if there's a linkname.
	//
	// TODO(austin): Maybe we want to create the ABI wrappers, but
	// ensure the linker exports the right ABI definition under
	// the unmangled name?
	cgoExports := make(map[string][]*[]string)
	for i, prag := range typecheck.Target.CgoPragmas {
		switch prag[0] {
		case "cgo_export_static", "cgo_export_dynamic":
			symName := s.canonicalize(prag[1])
			pprag := &typecheck.Target.CgoPragmas[i]
			cgoExports[symName] = append(cgoExports[symName], pprag)
		}
	}

	// Apply ABI defs and refs to Funcs and generate wrappers.
	//
	// This may generate new decls for the wrappers, but we
	// specifically *don't* want to visit those, lest we create
	// wrappers for wrappers.
	for _, fn := range typecheck.Target.Decls {
		if fn.Op() != ir.ODCLFUNC {
			continue
		}
		fn := fn.(*ir.Func)
		nam := fn.Nname
		if ir.IsBlank(nam) {
			continue
		}
		sym := nam.Sym()

		symName := sym.Linkname
		if symName == "" {
			symName = sym.Pkg.Prefix + "." + sym.Name
		}
		symName = s.canonicalize(symName)

		// Apply definitions.
		defABI, hasDefABI := s.defs[symName]
		if hasDefABI {
			if len(fn.Body) != 0 {
				base.ErrorfAt(fn.Pos(), "%v defined in both Go and assembly", fn)
			}
			fn.ABI = defABI
		}

		if fn.Pragma&ir.CgoUnsafeArgs != 0 {
			// CgoUnsafeArgs indicates the function (or its callee) uses
			// offsets to dispatch arguments, which currently using ABI0
			// frame layout. Pin it to ABI0.
			fn.ABI = obj.ABI0
		}

		// If cgo-exported, add the definition ABI to the cgo
		// pragmas.
		cgoExport := cgoExports[symName]
		for _, pprag := range cgoExport {
			// The export pragmas have the form:
			//
			//   cgo_export_* <local> [<remote>]
			//
			// If <remote> is omitted, it's the same as
			// <local>.
			//
			// Expand to
			//
			//   cgo_export_* <local> <remote> <ABI>
			if len(*pprag) == 2 {
				*pprag = append(*pprag, (*pprag)[1])
			}
			// Add the ABI argument.
			*pprag = append(*pprag, fn.ABI.String())
		}

		// Apply references.
		if abis, ok := s.refs[symName]; ok {
			fn.ABIRefs |= abis
		}
		// Assume all functions are referenced at least as
		// ABIInternal, since they may be referenced from
		// other packages.
		fn.ABIRefs.Set(obj.ABIInternal, true)

		// If a symbol is defined in this package (either in
		// Go or assembly) and given a linkname, it may be
		// referenced from another package, so make it
		// callable via any ABI. It's important that we know
		// it's defined in this package since other packages
		// may "pull" symbols using linkname and we don't want
		// to create duplicate ABI wrappers.
		//
		// However, if it's given a linkname for exporting to
		// C, then we don't make ABI wrappers because the cgo
		// tool wants the original definition.
		hasBody := len(fn.Body) != 0
		if sym.Linkname != "" && (hasBody || hasDefABI) && len(cgoExport) == 0 {
			fn.ABIRefs |= obj.ABISetCallable
		}

		// Double check that cgo-exported symbols don't get
		// any wrappers.
		if len(cgoExport) > 0 && fn.ABIRefs&^obj.ABISetOf(fn.ABI) != 0 {
			base.Fatalf("cgo exported function %v cannot have ABI wrappers", fn)
		}

		if !buildcfg.Experiment.RegabiWrappers {
			continue
		}

		forEachWrapperABI(fn, makeABIWrapper)
	}
}

func forEachWrapperABI(fn *ir.Func, cb func(fn *ir.Func, wrapperABI obj.ABI)) {
	need := fn.ABIRefs &^ obj.ABISetOf(fn.ABI)
	if need == 0 {
		return
	}

	for wrapperABI := obj.ABI(0); wrapperABI < obj.ABICount; wrapperABI++ {
		if !need.Get(wrapperABI) {
			continue
		}
		cb(fn, wrapperABI)
	}
}

// makeABIWrapper creates a new function that will be called with
// wrapperABI and calls "f" using f.ABI.
func makeABIWrapper(f *ir.Func, wrapperABI obj.ABI) {
	if base.Debug.ABIWrap != 0 {
		fmt.Fprintf(os.Stderr, "=-= %v to %v wrapper for %v\n", wrapperABI, f.ABI, f)
	}

	// Q: is this needed?
	savepos := base.Pos
	savedclcontext := typecheck.DeclContext
	savedcurfn := ir.CurFunc

	base.Pos = base.AutogeneratedPos
	typecheck.DeclContext = ir.PEXTERN

	// At the moment we don't support wrapping a method, we'd need machinery
	// below to handle the receiver. Panic if we see this scenario.
	ft := f.Nname.Type()
	if ft.NumRecvs() != 0 {
		base.ErrorfAt(f.Pos(), "makeABIWrapper support for wrapping methods not implemented")
		return
	}

	// Reuse f's types.Sym to create a new ODCLFUNC/function.
	fn := typecheck.DeclFunc(f.Nname.Sym(), nil,
		typecheck.NewFuncParams(ft.Params(), true),
		typecheck.NewFuncParams(ft.Results(), false))
	fn.ABI = wrapperABI

	fn.SetABIWrapper(true)
	fn.SetDupok(true)

	// ABI0-to-ABIInternal wrappers will be mainly loading params from
	// stack into registers (and/or storing stack locations back to
	// registers after the wrapped call); in most cases they won't
	// need to allocate stack space, so it should be OK to mark them
	// as NOSPLIT in these cases. In addition, my assumption is that
	// functions written in assembly are NOSPLIT in most (but not all)
	// cases. In the case of an ABIInternal target that has too many
	// parameters to fit into registers, the wrapper would need to
	// allocate stack space, but this seems like an unlikely scenario.
	// Hence: mark these wrappers NOSPLIT.
	//
	// ABIInternal-to-ABI0 wrappers on the other hand will be taking
	// things in registers and pushing them onto the stack prior to
	// the ABI0 call, meaning that they will always need to allocate
	// stack space. If the compiler marks them as NOSPLIT this seems
	// as though it could lead to situations where the linker's
	// nosplit-overflow analysis would trigger a link failure. On the
	// other hand if they not tagged NOSPLIT then this could cause
	// problems when building the runtime (since there may be calls to
	// asm routine in cases where it's not safe to grow the stack). In
	// most cases the wrapper would be (in effect) inlined, but are
	// there (perhaps) indirect calls from the runtime that could run
	// into trouble here.
	// FIXME: at the moment all.bash does not pass when I leave out
	// NOSPLIT for these wrappers, so all are currently tagged with NOSPLIT.
	fn.Pragma |= ir.Nosplit

	// Generate call. Use tail call if no params and no returns,
	// but a regular call otherwise.
	//
	// Note: ideally we would be using a tail call in cases where
	// there are params but no returns for ABI0->ABIInternal wrappers,
	// provided that all params fit into registers (e.g. we don't have
	// to allocate any stack space). Doing this will require some
	// extra work in typecheck/walk/ssa, might want to add a new node
	// OTAILCALL or something to this effect.
	tailcall := fn.Type().NumResults() == 0 && fn.Type().NumParams() == 0 && fn.Type().NumRecvs() == 0
	if base.Ctxt.Arch.Name == "ppc64le" && base.Ctxt.Flag_dynlink {
		// cannot tailcall on PPC64 with dynamic linking, as we need
		// to restore R2 after call.
		tailcall = false
	}
	if base.Ctxt.Arch.Name == "amd64" && wrapperABI == obj.ABIInternal {
		// cannot tailcall from ABIInternal to ABI0 on AMD64, as we need
		// to special registers (X15) when returning to ABIInternal.
		tailcall = false
	}

	var tail ir.Node
	call := ir.NewCallExpr(base.Pos, ir.OCALL, f.Nname, nil)
	call.Args = ir.ParamNames(fn.Type())
	call.IsDDD = fn.Type().IsVariadic()
	tail = call
	if tailcall {
		tail = ir.NewTailCallStmt(base.Pos, call)
	} else if fn.Type().NumResults() > 0 {
		n := ir.NewReturnStmt(base.Pos, nil)
		n.Results = []ir.Node{call}
		tail = n
	}
	fn.Body.Append(tail)

	typecheck.FinishFuncBody()
	if base.Debug.DclStack != 0 {
		types.CheckDclstack()
	}

	typecheck.Func(fn)
	ir.CurFunc = fn
	typecheck.Stmts(fn.Body)

	typecheck.Target.Decls = append(typecheck.Target.Decls, fn)

	// Restore previous context.
	base.Pos = savepos
	typecheck.DeclContext = savedclcontext
	ir.CurFunc = savedcurfn
}
