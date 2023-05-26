// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file implements initialization and assignment checks.

package types

import (
	"fmt"
	"go/ast"
	"go/token"
	. "internal/types/errors"
	"strings"
)

// assignment reports whether x can be assigned to a variable of type T,
// if necessary by attempting to convert untyped values to the appropriate
// type. context describes the context in which the assignment takes place.
// Use T == nil to indicate assignment to an untyped blank identifier.
// x.mode is set to invalid if the assignment failed.
func (check *Checker) assignment(x *operand, T Type, context string) {
	check.singleValue(x)

	switch x.mode {
	case invalid:
		return // error reported before
	case constant_, variable, mapindex, value, commaok, commaerr:
		// ok
	default:
		// we may get here because of other problems (issue #39634, crash 12)
		// TODO(gri) do we need a new "generic" error code here?
		check.errorf(x, IncompatibleAssign, "cannot assign %s to %s in %s", x, T, context)
		return
	}

	if isUntyped(x.typ) {
		target := T
		// spec: "If an untyped constant is assigned to a variable of interface
		// type or the blank identifier, the constant is first converted to type
		// bool, rune, int, float64, complex128 or string respectively, depending
		// on whether the value is a boolean, rune, integer, floating-point,
		// complex, or string constant."
		if T == nil || isNonTypeParamInterface(T) {
			if T == nil && x.typ == Typ[UntypedNil] {
				check.errorf(x, UntypedNilUse, "use of untyped nil in %s", context)
				x.mode = invalid
				return
			}
			target = Default(x.typ)
		}
		newType, val, code := check.implicitTypeAndValue(x, target)
		if code != 0 {
			msg := check.sprintf("cannot use %s as %s value in %s", x, target, context)
			switch code {
			case TruncatedFloat:
				msg += " (truncated)"
			case NumericOverflow:
				msg += " (overflows)"
			default:
				code = IncompatibleAssign
			}
			check.error(x, code, msg)
			x.mode = invalid
			return
		}
		if val != nil {
			x.val = val
			check.updateExprVal(x.expr, val)
		}
		if newType != x.typ {
			x.typ = newType
			check.updateExprType(x.expr, newType, false)
		}
	}

	// A generic (non-instantiated) function value cannot be assigned to a variable.
	if sig, _ := under(x.typ).(*Signature); sig != nil && sig.TypeParams().Len() > 0 {
		check.errorf(x, WrongTypeArgCount, "cannot use generic function %s without instantiation in %s", x, context)
	}

	// spec: "If a left-hand side is the blank identifier, any typed or
	// non-constant value except for the predeclared identifier nil may
	// be assigned to it."
	if T == nil {
		return
	}

	cause := ""
	if ok, code := x.assignableTo(check, T, &cause); !ok {
		if cause != "" {
			check.errorf(x, code, "cannot use %s as %s value in %s: %s", x, T, context, cause)
		} else {
			check.errorf(x, code, "cannot use %s as %s value in %s", x, T, context)
		}
		x.mode = invalid
	}
}

func (check *Checker) initConst(lhs *Const, x *operand) {
	if x.mode == invalid || x.typ == Typ[Invalid] || lhs.typ == Typ[Invalid] {
		if lhs.typ == nil {
			lhs.typ = Typ[Invalid]
		}
		return
	}

	// rhs must be a constant
	if x.mode != constant_ {
		check.errorf(x, InvalidConstInit, "%s is not constant", x)
		if lhs.typ == nil {
			lhs.typ = Typ[Invalid]
		}
		return
	}
	assert(isConstType(x.typ))

	// If the lhs doesn't have a type yet, use the type of x.
	if lhs.typ == nil {
		lhs.typ = x.typ
	}

	check.assignment(x, lhs.typ, "constant declaration")
	if x.mode == invalid {
		return
	}

	lhs.val = x.val
}

func (check *Checker) initVar(lhs *Var, x *operand, context string) Type {
	if x.mode == invalid || x.typ == Typ[Invalid] || lhs.typ == Typ[Invalid] {
		if lhs.typ == nil {
			lhs.typ = Typ[Invalid]
		}
		return nil
	}

	// If the lhs doesn't have a type yet, use the type of x.
	if lhs.typ == nil {
		typ := x.typ
		if isUntyped(typ) {
			// convert untyped types to default types
			if typ == Typ[UntypedNil] {
				check.errorf(x, UntypedNilUse, "use of untyped nil in %s", context)
				lhs.typ = Typ[Invalid]
				return nil
			}
			typ = Default(typ)
		}
		lhs.typ = typ
	}

	check.assignment(x, lhs.typ, context)
	if x.mode == invalid {
		return nil
	}

	return x.typ
}

func (check *Checker) assignVar(lhs ast.Expr, x *operand) Type {
	if x.mode == invalid || x.typ == Typ[Invalid] {
		check.useLHS(lhs)
		return nil
	}

	// Determine if the lhs is a (possibly parenthesized) identifier.
	ident, _ := unparen(lhs).(*ast.Ident)

	// Don't evaluate lhs if it is the blank identifier.
	if ident != nil && ident.Name == "_" {
		check.recordDef(ident, nil)
		check.assignment(x, nil, "assignment to _ identifier")
		if x.mode == invalid {
			return nil
		}
		return x.typ
	}

	// If the lhs is an identifier denoting a variable v, this assignment
	// is not a 'use' of v. Remember current value of v.used and restore
	// after evaluating the lhs via check.expr.
	var v *Var
	var v_used bool
	if ident != nil {
		if obj := check.lookup(ident.Name); obj != nil {
			// It's ok to mark non-local variables, but ignore variables
			// from other packages to avoid potential race conditions with
			// dot-imported variables.
			if w, _ := obj.(*Var); w != nil && w.pkg == check.pkg {
				v = w
				v_used = v.used
			}
		}
	}

	var z operand
	check.expr(&z, lhs)
	if v != nil {
		v.used = v_used // restore v.used
	}

	if z.mode == invalid || z.typ == Typ[Invalid] {
		return nil
	}

	// spec: "Each left-hand side operand must be addressable, a map index
	// expression, or the blank identifier. Operands may be parenthesized."
	switch z.mode {
	case invalid:
		return nil
	case variable, mapindex:
		// ok
	default:
		if sel, ok := z.expr.(*ast.SelectorExpr); ok {
			var op operand
			check.expr(&op, sel.X)
			if op.mode == mapindex {
				check.errorf(&z, UnaddressableFieldAssign, "cannot assign to struct field %s in map", ExprString(z.expr))
				return nil
			}
		}
		check.errorf(&z, UnassignableOperand, "cannot assign to %s", &z)
		return nil
	}

	check.assignment(x, z.typ, "assignment")
	if x.mode == invalid {
		return nil
	}

	return x.typ
}

// operandTypes returns the list of types for the given operands.
func operandTypes(list []*operand) (res []Type) {
	for _, x := range list {
		res = append(res, x.typ)
	}
	return res
}

// varTypes returns the list of types for the given variables.
func varTypes(list []*Var) (res []Type) {
	for _, x := range list {
		res = append(res, x.typ)
	}
	return res
}

// typesSummary returns a string of the form "(t1, t2, ...)" where the
// ti's are user-friendly string representations for the given types.
// If variadic is set and the last type is a slice, its string is of
// the form "...E" where E is the slice's element type.
func (check *Checker) typesSummary(list []Type, variadic bool) string {
	var res []string
	for i, t := range list {
		var s string
		switch {
		case t == nil:
			fallthrough // should not happen but be cautious
		case t == Typ[Invalid]:
			s = "<T>"
		case isUntyped(t):
			if isNumeric(t) {
				// Do not imply a specific type requirement:
				// "have number, want float64" is better than
				// "have untyped int, want float64" or
				// "have int, want float64".
				s = "number"
			} else {
				// If we don't have a number, omit the "untyped" qualifier
				// for compactness.
				s = strings.Replace(t.(*Basic).name, "untyped ", "", -1)
			}
		case variadic && i == len(list)-1:
			s = check.sprintf("...%s", t.(*Slice).elem)
		}
		if s == "" {
			s = check.sprintf("%s", t)
		}
		res = append(res, s)
	}
	return "(" + strings.Join(res, ", ") + ")"
}

func measure(x int, unit string) string {
	if x != 1 {
		unit += "s"
	}
	return fmt.Sprintf("%d %s", x, unit)
}

func (check *Checker) assignError(rhs []ast.Expr, nvars, nvals int) {
	vars := measure(nvars, "variable")
	vals := measure(nvals, "value")
	rhs0 := rhs[0]

	if len(rhs) == 1 {
		if call, _ := unparen(rhs0).(*ast.CallExpr); call != nil {
			check.errorf(rhs0, WrongAssignCount, "assignment mismatch: %s but %s returns %s", vars, call.Fun, vals)
			return
		}
	}
	check.errorf(rhs0, WrongAssignCount, "assignment mismatch: %s but %s", vars, vals)
}

// If returnStmt != nil, initVars is called to type-check the assignment
// of return expressions, and returnStmt is the return statement.
func (check *Checker) initVars(lhs []*Var, origRHS []ast.Expr, returnStmt ast.Stmt) {
	rhs, commaOk := check.exprList(origRHS, len(lhs) == 2 && returnStmt == nil)

	if len(lhs) != len(rhs) {
		// invalidate lhs
		for _, obj := range lhs {
			obj.used = true // avoid declared and not used errors
			if obj.typ == nil {
				obj.typ = Typ[Invalid]
			}
		}
		// don't report an error if we already reported one
		for _, x := range rhs {
			if x.mode == invalid {
				return
			}
		}
		if returnStmt != nil {
			var at positioner = returnStmt
			qualifier := "not enough"
			if len(rhs) > len(lhs) {
				at = rhs[len(lhs)].expr // report at first extra value
				qualifier = "too many"
			} else if len(rhs) > 0 {
				at = rhs[len(rhs)-1].expr // report at last value
			}
			err := newErrorf(at, WrongResultCount, "%s return values", qualifier)
			err.errorf(token.NoPos, "have %s", check.typesSummary(operandTypes(rhs), false))
			err.errorf(token.NoPos, "want %s", check.typesSummary(varTypes(lhs), false))
			check.report(err)
			return
		}
		check.assignError(origRHS, len(lhs), len(rhs))
		return
	}

	context := "assignment"
	if returnStmt != nil {
		context = "return statement"
	}

	if commaOk {
		var a [2]Type
		for i := range a {
			a[i] = check.initVar(lhs[i], rhs[i], context)
		}
		check.recordCommaOkTypes(origRHS[0], a)
		return
	}

	for i, lhs := range lhs {
		check.initVar(lhs, rhs[i], context)
	}
}

func (check *Checker) assignVars(lhs, origRHS []ast.Expr) {
	rhs, commaOk := check.exprList(origRHS, len(lhs) == 2)

	if len(lhs) != len(rhs) {
		check.useLHS(lhs...)
		// don't report an error if we already reported one
		for _, x := range rhs {
			if x.mode == invalid {
				return
			}
		}
		check.assignError(origRHS, len(lhs), len(rhs))
		return
	}

	if commaOk {
		var a [2]Type
		for i := range a {
			a[i] = check.assignVar(lhs[i], rhs[i])
		}
		check.recordCommaOkTypes(origRHS[0], a)
		return
	}

	for i, lhs := range lhs {
		check.assignVar(lhs, rhs[i])
	}
}

func (check *Checker) shortVarDecl(pos positioner, lhs, rhs []ast.Expr) {
	top := len(check.delayed)
	scope := check.scope

	// collect lhs variables
	seen := make(map[string]bool, len(lhs))
	lhsVars := make([]*Var, len(lhs))
	newVars := make([]*Var, 0, len(lhs))
	hasErr := false
	for i, lhs := range lhs {
		ident, _ := lhs.(*ast.Ident)
		if ident == nil {
			check.useLHS(lhs)
			// TODO(rFindley) this is redundant with a parser error. Consider omitting?
			check.errorf(lhs, BadDecl, "non-name %s on left side of :=", lhs)
			hasErr = true
			continue
		}

		name := ident.Name
		if name != "_" {
			if seen[name] {
				check.errorf(lhs, RepeatedDecl, "%s repeated on left side of :=", lhs)
				hasErr = true
				continue
			}
			seen[name] = true
		}

		// Use the correct obj if the ident is redeclared. The
		// variable's scope starts after the declaration; so we
		// must use Scope.Lookup here and call Scope.Insert
		// (via check.declare) later.
		if alt := scope.Lookup(name); alt != nil {
			check.recordUse(ident, alt)
			// redeclared object must be a variable
			if obj, _ := alt.(*Var); obj != nil {
				lhsVars[i] = obj
			} else {
				check.errorf(lhs, UnassignableOperand, "cannot assign to %s", lhs)
				hasErr = true
			}
			continue
		}

		// declare new variable
		obj := NewVar(ident.Pos(), check.pkg, name, nil)
		lhsVars[i] = obj
		if name != "_" {
			newVars = append(newVars, obj)
		}
		check.recordDef(ident, obj)
	}

	// create dummy variables where the lhs is invalid
	for i, obj := range lhsVars {
		if obj == nil {
			lhsVars[i] = NewVar(lhs[i].Pos(), check.pkg, "_", nil)
		}
	}

	check.initVars(lhsVars, rhs, nil)

	// process function literals in rhs expressions before scope changes
	check.processDelayed(top)

	if len(newVars) == 0 && !hasErr {
		check.softErrorf(pos, NoNewVar, "no new variables on left side of :=")
		return
	}

	// declare new variables
	// spec: "The scope of a constant or variable identifier declared inside
	// a function begins at the end of the ConstSpec or VarSpec (ShortVarDecl
	// for short variable declarations) and ends at the end of the innermost
	// containing block."
	scopePos := rhs[len(rhs)-1].End()
	for _, obj := range newVars {
		check.declare(scope, nil, obj, scopePos) // id = nil: recordDef already called
	}
}
