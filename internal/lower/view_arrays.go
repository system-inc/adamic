package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// The shared contract builder owns interning and recursion. This hook builds only
// the declared element contract; it never inspects array contents at a cast/read.
// An error from a nested object/union/callable contract must not be discarded.
func (l *lowering) viewArrayContract(node *ast.Node, target *checker.Type, buildElement func(*checker.Type) error) (bool, error) {
	if checker.IsTupleType(target) {
		return true, l.notYet(node, "a checked tuple view with per-position optional and rest contracts")
	}
	element := l.viewArrayElementType(target)
	if element == nil {
		return false, nil
	}
	if buildElement == nil {
		return true, l.notYet(node, "an array view without a recursive element contract builder")
	}
	return true, buildElement(element)
}

func (l *lowering) viewArrayFields(node *ast.Node, target *checker.Type, fields map[string]bool, seen map[*checker.Type]bool) error {
	if seen[target] {
		return nil
	}
	seen[target] = true
	if err := l.viewArrayUnsupportedUses(node); err != nil {
		return err
	}
	if checker.IsTupleType(target) {
		return l.notYet(node, "a checked tuple view")
	}
	element := l.viewArrayElementType(target)
	if element == nil {
		return l.notYet(node, "an array view without an element type")
	}
	if l.callableViewContract(element) {
		return l.viewCallableCall(node, l.checker.TypeToString(target)+"[element]", false)
	}
	if interfaceScalar(element) {
		return nil
	}
	if element.Flags()&checker.TypeFlagsObject != 0 && l.checker.IsArrayType(element) {
		return l.viewArrayFields(node, element, fields, seen)
	}
	return l.viewObjectFields(node, element, fields, seen, true)
}

// This is recorded before all casts have necessarily been lowered. Readiness's
// final pass enables it only in a program containing an admitted array view.
func (l *lowering) markViewArrayRead(node *ast.Node, read ir.ArrayIndex) ir.ArrayIndex {
	array := node.AsElementAccessExpression().Expression
	element := l.viewArrayElementType(l.checker.GetTypeAtLocation(array))
	if element != nil {
		read.View = sourceExpression(node)
		read.ViewType = l.checker.TypeToString(element)
		read.ViewTypeID = int(element.Id())
		read.UndefinedAllowed = l.includesUndefined(element)
		read.ViewAllowed = l.viewContractLiterals(element)
		read.TupleUnion = l.tupleScalarUnionType(element)
	}
	return read
}

func markProgramViewArrayRead(program *ir.Program, read ir.ArrayIndex) ir.ArrayIndex {
	if !ir.HasArrayViews(program) {
		read.View = ""
		read.TupleUnion = false
		read.ViewAllowed = nil
		return read
	}
	read.ViewContract = program.ViewContractTypes[read.ViewTypeID]
	if read.TupleUnion {
		_, read.TupleUnion = ir.TupleViewMembers(program, read.ViewContract)
	}
	return read
}

func (l *lowering) viewArrayUse(node, array *ast.Node, of ir.Type, required bool) ir.ArrayViewRead {
	element := l.viewArrayElementType(l.checker.GetTypeAtLocation(array))
	if element == nil {
		return ir.ArrayViewRead{}
	}
	return ir.ArrayViewRead{TupleUnion: l.tupleScalarUnionType(element), UndefinedAllowed: l.includesUndefined(element), Element: of, View: sourceExpression(array) + "[element]", ViewType: l.checker.TypeToString(element), ViewTypeID: int(element.Id()), ViewAllowed: l.viewContractLiterals(element), Required: required && !l.includesUndefined(element)}
}

// Fail closed for consumers whose element extraction/conversion is not wired.
// This scan is conservative, as lane 1's global field-name policy is. Production
// expressions and indexed writes are supported; they retain original storage.
func (l *lowering) viewArrayUnsupportedUses(node *ast.Node) error {
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return err
	}
	var found error
	var visit ast.Visitor
	visit = func(part *ast.Node) bool {
		if found != nil {
			return true
		}
		if part.Kind == ast.KindCallExpression {
			callee := ast.SkipParentheses(part.AsCallExpression().Expression)
			if callee.Kind == ast.KindPropertyAccessExpression && l.viewArrayBase(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(callee.AsPropertyAccessExpression().Expression))) != nil {
				if callee.Name().Text() != "forEach" && l.tupleScalarUnionType(l.viewArrayElementType(l.checker.GetTypeAtLocation(callee.AsPropertyAccessExpression().Expression))) {
					found = l.notYet(part, "tuple union array consumer requiring a certified ownership transfer")
					return true
				}
				switch callee.Name().Text() {
				case "map", "forEach", "filter", "some", "every", "find", "findIndex", "reduce", "slice", "at", "pop", "push", "indexOf", "includes", "lastIndexOf":
				case "join":
					element := l.viewArrayElementType(l.checker.GetTypeAtLocation(callee.AsPropertyAccessExpression().Expression))
					if l.checker.IsArrayType(l.checker.GetNonNullableType(element)) {
						found = l.notYet(part, "checked nested array view consumer .join requiring recursive element conversion")
					}
				default:
					found = l.notYet(part, "checked array view consumer ."+callee.Name().Text()+" requiring element conversion")
				}
			}
		}
		if part.Kind == ast.KindForOfStatement {
			if element := l.viewArrayElementType(l.checker.GetTypeAtLocation(part.AsForInOrOfStatement().Expression)); l.tupleScalarUnionType(element) {
				found = l.notYet(part, "tuple union array iteration requiring a certified ownership transfer")
			}
		}
		if part.Kind == ast.KindSpreadElement && l.checker.IsArrayType(l.checker.GetTypeAtLocation(part.AsSpreadElement().Expression)) {
			found = l.notYet(part, "spreading a checked array requiring storage conversion")
		}
		part.ForEachChild(visit)
		return found != nil
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	return found
}

func markProgramViewArrayUse(program *ir.Program, read ir.ArrayViewRead) ir.ArrayViewRead {
	mapped := markProgramViewArrayRead(program, read.Index())
	read.View, read.ViewAllowed, read.ViewContract, read.TupleUnion = mapped.View, mapped.ViewAllowed, mapped.ViewContract, mapped.TupleUnion
	return read
}

func (l *lowering) viewArrayCast(node *ast.Node, value ir.Expression, source, target *checker.Type) (ir.Expression, error) {
	if value.Type() != ir.Array || !l.checker.IsArrayType(l.withoutUndefined(source)) || !l.checker.IsArrayType(target) {
		return nil, nil
	}
	// Readonly scalar elements introduce no writable slot; their type is checked
	// lazily when read, including an unknown[] bridge into a readonly result.
	readonlyScalar := l.isLibraryType(target, "ReadonlyArray") && interfaceScalar(l.concrete(l.viewArrayElementType(target)))
	consumer := l.readonlyArrayConsumer(node)
	readonlyConsumer := consumer != nil && interfaceScalar(l.concrete(l.viewArrayElementType(target)))
	if !readonlyScalar && !readonlyConsumer {
		if err := l.widened(source, target, map[[2]*checker.Type]bool{}); err != nil {
			return nil, l.notYet(node, "a writable array view requiring source contract certification")
		}
	}
	return l.view(node, value, target)
}

// Attach the declared element contract at the consuming join, never at a field read.
func (l *lowering) viewArrayJoin(node, receiver *ast.Node, array, separator ir.Expression, element ir.Type) ir.ArrayJoin {
	return ir.ArrayJoin{Array: array, Separator: separator, Element: element, ViewRead: l.viewArrayUse(node, receiver, element, false)}
}

// Only lift the hole guard for consumers whose sparse runtime paths are wired.
// Other holey programs keep the library lane's existing refusals.
func (l *lowering) viewArrayHolesMethod(name string) bool {
	if !ir.HasArrayViews(l.result) {
		return false
	}
	return name == "push" || name == "pop" || name == "slice"
}

func (l *lowering) viewArrayHolesConsumer(node any) bool {
	if !ir.HasArrayViews(l.result) {
		return false
	}
	switch node.(type) {
	case ir.ArrayPush, ir.ArrayPop, ir.ArraySlice:
		return true
	}
	return false
}

func (l *lowering) viewArraySearch(node, receiver *ast.Node, array, value ir.Expression, element ir.Type, includes, last bool) ir.ArraySearch {
	return ir.ArraySearch{Array: array, Value: value, Element: element, Includes: includes, Last: last, ViewRead: l.viewArrayUse(node, receiver, element, false)}
}
