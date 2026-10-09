package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// This is recorded before all casts have necessarily been lowered. Readiness's
// final pass enables it only in a program containing an admitted array view.
func (l *lowering) markViewArrayRead(node *ast.Node, read ir.ArrayIndex) ir.ArrayIndex {
	array := node.AsElementAccessExpression().Expression
	element := l.untaggedArrayElement(l.checker.GetNonNullableType(l.concrete(l.checker.GetTypeAtLocation(array))))
	if element != nil {
		element = l.concrete(element)
		_, _ = l.viewContract(node, element)
		read.View = sourceExpression(node)
		read.ViewType = l.checker.TypeToString(element)
		read.ViewTypeID = int(element.Id())
		read.UndefinedAllowed = l.includesUndefined(element)
		read.ViewAllowed = l.viewContractLiterals(element)
	}
	return read
}

func markProgramViewArrayRead(program *ir.Program, read ir.ArrayIndex) ir.ArrayIndex {
	if !ir.HasArrayViews(program) {
		read.View = ""
		read.ViewAllowed = nil
		return read
	}
	read.ViewContract = program.ViewContractTypes[read.ViewTypeID]
	return read
}

func (l *lowering) viewArrayUse(node, array *ast.Node, of ir.Type, required bool) ir.ArrayViewRead {
	element := l.untaggedArrayElement(l.checker.GetNonNullableType(l.concrete(l.checker.GetTypeAtLocation(array))))
	if element == nil {
		return ir.ArrayViewRead{}
	}
	element = l.concrete(element)
	_, _ = l.viewContract(node, element)
	return ir.ArrayViewRead{UndefinedAllowed: l.includesUndefined(element), Element: of, View: sourceExpression(array) + "[element]", ViewType: l.checker.TypeToString(element), ViewTypeID: int(element.Id()), ViewAllowed: l.viewContractLiterals(element), Required: required && !l.includesUndefined(element)}
}

func markProgramViewArrayUse(program *ir.Program, read ir.ArrayViewRead) ir.ArrayViewRead {
	mapped := markProgramViewArrayRead(program, read.Index())
	read.View, read.ViewAllowed, read.ViewContract = mapped.View, mapped.ViewAllowed, mapped.ViewContract
	return read
}

func (l *lowering) viewArrayString(node, span *ast.Node, value ir.Expression) (ir.Expression, error) {
	declared := false
	for _, c := range l.result.ViewContracts {
		declared = declared || c.Kind == ir.ViewArray
	}
	if !declared {
		return nil, l.notYet(span, "a template interpolating an object, an array, a map, a function or undefined")
	}
	l.result.ArrayViewTemplateWhere = l.program.Where(span)
	element, err := l.elementType(node)
	if err != nil {
		return nil, err
	}
	join := ir.ArrayJoin{Array: value, Separator: ir.StringConstant{Index: l.constant(",")}, Element: element, ViewRead: l.viewArrayUse(node, node, element, false)}
	if !l.includesUndefined(l.checker.GetTypeAtLocation(node)) {
		return join, nil
	}
	local := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "array_string_receiver", Type: ir.Array, Function: l.functionIndex})
	held := ir.Read{Local: local, Of: ir.Array}
	join.Array = held
	return ir.Effects{Body: []ir.Statement{ir.Declare{Local: local, Value: value}}, Result: ir.Conditional{Condition: ir.Truthy{Value: held}, WhenTrue: join, WhenNot: ir.StringConstant{Index: l.constant("undefined")}, Of: ir.String}}, nil
}

// A joined logical element type cannot choose an arbitrary physical member.
// Same-storage object arms can share checked reads; scalar/reference conversion
// still needs a boxing adapter and keeps the existing refusal.
func (l *lowering) sameArrayMemberStorage(target *checker.Type) bool {
	if target.Flags()&checker.TypeFlagsUnion == 0 {
		return true
	}
	held := ir.Type(0)
	for _, member := range target.Types() {
		element := l.viewArrayElementType(member)
		if element == nil {
			return false
		}
		storage, known := l.kept(element)
		if !known || held != 0 && storage != held {
			return false
		}
		held = storage
	}
	return held != 0
}
