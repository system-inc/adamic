package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Port of V8 ArrayReduceLoopContinuation and ArrayReduceRightLoopContinuation
// (src/builtins/array-reduce.tq and array-reduce-right.tq). Dense arrays need no
// hole sentinel after the first element seeds an omitted initial accumulator.
func (l *lowering) libraryArrayReduce(node *ast.Node, array ir.Expression, element ir.Type, name string) (ir.Expression, bool, error) {
	written := node.AsCallExpression().Arguments.Nodes
	if len(written) < 1 || len(written) > 2 {
		return nil, true, l.notYet(node, name+" without a callback and optional initial value")
	}
	callback, err := l.expression(written[0])
	if err != nil {
		return nil, true, err
	}
	signatures := l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(written[0]), checker.SignatureKindCall)
	if callback.Type() != ir.Closure || len(signatures) != 1 || len(signatures[0].Parameters()) > 4 {
		return nil, true, l.notYet(node, name+" with this callback signature")
	}
	result, err := l.typeOf(node)
	if err != nil {
		return nil, true, err
	}
	returned, known := l.representation(l.checker.GetReturnTypeOfSignature(signatures[0]))
	if !known || slotless(result) || fit(ir.Read{Of: returned}, result).Type() != result {
		return nil, true, l.notYet(node, name+" with an incompatible accumulator representation")
	}
	arguments := []ir.Expression{array, callback}
	if len(written) == 2 {
		initial, err := l.expression(written[1])
		if err != nil {
			return nil, true, err
		}
		initial = fit(initial, result)
		if initial.Type() != result {
			return nil, true, l.notYet(node, name+" with an incompatible initial value")
		}
		arguments = append(arguments, initial)
	} else if element != result {
		return nil, true, l.notYet(node, name+" without an initial value of the element representation")
	}
	b := l.libraryArrayBuilder(arguments)
	source := b.read(b.parameters[0])
	length := b.declare("length", ir.Length{Array: source})
	zero, one := ir.NumberConstant{}, ir.NumberConstant{Value: 1}
	last := name == "reduceRight"
	start := ir.Expression(zero)
	if last {
		start = ir.Binary{Operator: ir.Subtract, Left: b.read(length), Right: one}
	}
	index := b.declare("index", start)
	readElement := func() ir.Expression {
		value := ir.Expression(ir.ArrayIndex{Array: source, Index: b.read(index), Element: element})
		if value.Type().IsMaybe() && value.Type() != element {
			value = ir.Unwrap{Value: value}
		}
		return value
	}
	operator := ir.Add
	if last {
		operator = ir.Subtract
	}
	advance := ir.Assign{Local: index, Value: ir.Binary{Operator: operator, Left: b.read(index), Right: one}}
	initial := ir.Expression(nil)
	if len(arguments) == 3 {
		initial = b.read(b.parameters[2])
	} else {
		b.body = append(b.body, ir.If{Condition: ir.Binary{Operator: ir.Equal, Left: b.read(length), Right: zero}, Then: b.typeError(node, "Reduce of empty array with no initial value")})
		initial = readElement()
	}
	accumulator := b.declare("accumulator", initial)
	if len(arguments) == 2 {
		b.body = append(b.body, advance)
	}
	item := b.local("element", element)
	callArguments := []ir.Expression{b.read(accumulator), b.read(item), b.read(index), source}
	for index, parameter := range signatures[0].Parameters() {
		takes, known := l.representation(l.checker.GetTypeOfSymbol(parameter))
		if !known || slotless(takes) {
			return nil, true, l.notYet(node, name+" with this callback parameter representation")
		}
		callArguments[index] = fit(callArguments[index], takes)
		if callArguments[index].Type() != takes {
			return nil, true, l.notYet(node, name+" with incompatible callback parameters")
		}
	}
	condition := ir.Expression(ir.Binary{Operator: ir.Less, Left: b.read(index), Right: b.read(length)})
	if last {
		condition = ir.Binary{Operator: ir.GreaterOrEqual, Left: b.read(index), Right: zero}
	}
	// HasProperty: shrinking skips missing indexes; growing cannot extend the saved length.
	b.body = append(b.body, ir.Loop{Condition: condition, Body: []ir.Statement{ir.If{
		Condition: ir.Binary{Operator: ir.Less, Left: b.read(index), Right: ir.Length{Array: source}},
		Then:      []ir.Statement{ir.Declare{Local: item, Value: readElement()}, ir.Assign{Local: accumulator, Value: fit(ir.CallClosure{Closure: b.read(b.parameters[1]), Arguments: callArguments, Returns: returned}, result)}},
	}}, Update: []ir.Statement{advance}})
	return b.finish("array_"+name, b.read(accumulator)), true, nil
}

func (b *libraryArrayBuilder) typeError(node *ast.Node, message string) []ir.Statement {
	return b.typeErrorMessage(node, ir.StringConstant{Index: b.l.constant(message)})
}

func (b *libraryArrayBuilder) typeErrorMessage(node *ast.Node, message ir.Expression) []ir.Statement {
	local := b.local("type_error", ir.Object)
	return []ir.Statement{
		ir.Declare{Local: local, Value: ir.BuiltinError{Message: message, Kind: 1}},
		ir.Throw{Value: b.read(local)},
	}
}
