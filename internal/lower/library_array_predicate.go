package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Port of V8's ArrayFilter/Every/SomeLoopContinuation (Node 24.19.0).
// The receiver is an ordinary dense array: HasProperty is its current length
// check. Holes, prototypes, species and explicit thisArg remain unsupported.
func (l *lowering) libraryArrayPredicate(node *ast.Node, array ir.Expression, element ir.Type, name string) (ir.Expression, bool, error) {
	written := node.AsCallExpression().Arguments.Nodes
	if len(written) != 1 {
		return nil, true, l.notYet(node, name+" with other than one callback")
	}
	signatures := l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(written[0]), checker.SignatureKindCall)
	if len(signatures) != 1 || len(signatures[0].Parameters()) > 3 {
		return nil, true, l.notYet(node, name+" with an overloaded callback or more than three parameters")
	}
	returned := l.checker.GetReturnTypeOfSignature(signatures[0])
	returns, known := l.representation(returned)
	if known && returns == ir.Boolean {
		return l.arrayVisit(node, array, element, name)
	}
	isVoid := returned.Flags()&checker.TypeFlagsVoid != 0
	if isVoid {
		return nil, true, l.notYet(node, name+" with a void callback result (void can hide a returned value)")
	}
	if !known || returns == ir.Union || returns == ir.Weak {
		return nil, true, l.notYet(node, name+" with an unrepresented callback truth value")
	}
	if returns == ir.Object {
		// Structural object views such as {} also admit primitives. A pointer
		// presence test would incorrectly treat a boxed zero as truthy.
		for _, primitive := range []*checker.Type{l.checker.GetNumberType(), l.checker.GetStringType(), l.checker.GetBooleanType()} {
			if l.checker.IsTypeAssignableTo(primitive, returned) {
				return nil, true, l.notYet(node, name+" with an object callback result that can hide primitive truth values")
			}
		}
	}
	callback, err := l.expression(written[0])
	if err != nil {
		return nil, true, err
	}
	if callback.Type() != ir.Closure {
		return nil, true, l.notYet(node, name+" with a callback that isn't a function")
	}
	b := l.libraryArrayBuilder([]ir.Expression{array, callback})
	source := b.read(b.parameters[0])
	length := b.declare("length", ir.Length{Array: source})
	result := -1
	if name == "filter" {
		result = b.declare("filtered", ir.ArrayLiteral{Element: element})
	}
	index := b.declare("index", ir.NumberConstant{})
	input := ir.Expression(ir.ArrayIndex{Array: source, Index: b.read(index), Element: element})
	if input.Type().IsMaybe() && input.Type() != element {
		input = ir.Unwrap{Value: input}
	}
	item := b.local("element", element)
	// Keep kValue before the callback; filter writes that value even if the
	// callback replaces or removes the receiver's element.
	body := []ir.Statement{ir.Declare{Local: item, Value: input}}
	arguments := []ir.Expression{b.read(item), b.read(index), source}
	for i, parameter := range signatures[0].Parameters() {
		of, known := l.representation(l.checker.GetTypeOfSymbol(parameter))
		if !known || slotless(of) {
			return nil, true, l.notYet(node, name+" with an unrepresented callback parameter")
		}
		arguments[i] = fit(arguments[i], of)
		if arguments[i].Type() != of {
			return nil, true, l.notYet(node, name+" with an incompatible callback parameter")
		}
	}
	call := ir.CallClosure{Closure: b.read(b.parameters[1]), Arguments: arguments, Returns: returns}
	selected := ir.Expression(ir.BooleanConstant{Value: false})
	answer := b.local("answer", returns)
	body = append(body, ir.Declare{Local: answer, Value: call})
	if returned.Flags()&(checker.TypeFlagsNull|checker.TypeFlagsUndefined) == 0 {
		selected = libraryArrayToBoolean(b.read(answer))
	}
	if name == "filter" {
		body = append(body, ir.If{Condition: selected, Then: []ir.Statement{ir.Evaluate{Value: ir.ArrayPush{Array: b.read(result), Value: b.read(item), Element: element, Site: l.writeSite(node)}}}})
	} else {
		answer := name == "some"
		if !answer {
			selected = ir.Unary{Operator: ir.Not, Operand: selected}
		}
		body = append(body, ir.If{Condition: selected, Then: []ir.Statement{ir.Return{Value: ir.BooleanConstant{Value: answer}}}})
	}
	b.body = append(b.body, ir.Loop{Condition: ir.Binary{Operator: ir.Less, Left: b.read(index), Right: b.read(length)}, Body: []ir.Statement{ir.If{Condition: ir.Binary{Operator: ir.Less, Left: b.read(index), Right: ir.Length{Array: source}}, Then: body}}, Update: []ir.Statement{ir.Assign{Local: index, Value: ir.Binary{Operator: ir.Add, Left: b.read(index), Right: ir.NumberConstant{Value: 1}}}}})
	final := ir.Expression(ir.BooleanConstant{Value: name == "every"})
	if name == "filter" {
		final = b.read(result)
	}
	return b.finish("array_predicate_"+name, final), true, nil
}

// ToBoolean has no user coercion hooks. It never calls an object's valueOf.
func libraryArrayToBoolean(value ir.Expression) ir.Expression {
	if value.Type().IsMaybe() {
		return ir.Conditional{Condition: ir.IsUndefined{Value: value}, WhenTrue: ir.BooleanConstant{Value: false}, WhenNot: libraryArrayToBoolean(ir.Unwrap{Value: value})}
	}
	switch value.Type() {
	case ir.Boolean:
		return value
	case ir.Number:
		return ir.Binary{Operator: ir.And, Left: ir.Binary{Operator: ir.NotEqual, Left: value, Right: ir.NumberConstant{}}, Right: ir.Unary{Operator: ir.Not, Operand: ir.NumberCall{Function: "isNaN", Arguments: []ir.Expression{value}}}}
	case ir.String:
		return ir.Conditional{Condition: ir.IsUndefined{Value: value}, WhenTrue: ir.BooleanConstant{Value: false}, WhenNot: ir.Binary{Operator: ir.Greater, Left: ir.StringLength{Value: value}, Right: ir.NumberConstant{}}}
	default:
		return ir.Unary{Operator: ir.Not, Operand: ir.IsUndefined{Value: value}}
	}
}
