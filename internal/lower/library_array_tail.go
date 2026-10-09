package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// V8 SortCompareDefault converts each dense primitive element to a string and
// compares UTF-16 code units. Reuse the comparator already held for toSorted;
// sort mutates the original array and returns that same reference.
func (l *lowering) libraryArrayDefaultSort(node, receiver *ast.Node, array ir.Expression, element ir.Type) (ir.Expression, bool, error) {
	written := node.AsCallExpression().Arguments.Nodes
	if len(written) > 1 {
		return nil, true, l.notYet(node, "sort with extra arguments")
	}
	defaultComparator := len(written) == 0
	if len(written) == 1 {
		defaultComparator = l.checker.GetTypeAtLocation(written[0]).Flags()&checker.TypeFlagsUndefined != 0
	}
	if !defaultComparator {
		return l.arraySort(node, array, element)
	}
	if element != ir.Number && element != ir.Boolean && element != ir.String || l.includesUndefined(l.checker.GetElementTypeOfArrayType(l.checker.GetTypeAtLocation(receiver))) {
		return nil, true, l.notYet(node, "default sort requires dense primitive elements without undefined (object coercion and undefined ordering are not represented)")
	}
	comparator := l.libraryArrayComparator(element)
	if len(written) == 0 {
		return ir.ArraySort{Array: array, Element: element, Comparator: comparator}, true, nil
	}
	argument, err := l.expression(written[0])
	if err != nil {
		return nil, true, err
	}
	// Arguments run before sorting, including an expression returning undefined.
	b := l.libraryArrayBuilder([]ir.Expression{array, argument})
	return b.finish("array_default_sort", ir.ArraySort{Array: b.read(b.parameters[0]), Element: element, Comparator: comparator}), true, nil
}

// V8 ArrayReduceRightLoop snapshots length, visits existing indices in descending
// order and supplies accumulator, element, index and receiver to the callback.
// Dense arrays can shrink through pop/splice, so recheck length before each read.
func (l *lowering) libraryArrayReduceRight(node *ast.Node, array ir.Expression, element ir.Type) (ir.Expression, bool, error) {
	written := node.AsCallExpression().Arguments.Nodes
	if len(written) != 2 {
		return nil, true, l.notYet(node, "reduceRight requires a callback and an initial value (empty reduction needs a catchable TypeError)")
	}
	callback, err := l.expression(written[0])
	if err != nil {
		return nil, true, err
	}
	signatures := l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(written[0]), checker.SignatureKindCall)
	if callback.Type() != ir.Closure || len(signatures) != 1 || len(signatures[0].Parameters()) > 4 {
		return nil, true, l.notYet(node, "reduceRight requires one callback with at most four parameters")
	}
	result, err := l.typeOf(node)
	if err != nil {
		return nil, true, err
	}
	returned, known := l.representation(l.checker.GetReturnTypeOfSignature(signatures[0]))
	if !known || returned != result || slotless(result) {
		return nil, true, l.notYet(node, "reduceRight with an unrepresented accumulator result")
	}
	initial, err := l.expression(written[1])
	if err != nil {
		return nil, true, err
	}
	initial = fit(initial, result)
	if initial.Type() != result {
		return nil, true, l.notYet(node, "reduceRight with an unrepresented initial accumulator")
	}
	b := l.libraryArrayBuilder([]ir.Expression{array, callback, initial})
	source := b.read(b.parameters[0])
	accumulator := b.declare("accumulator", b.read(b.parameters[2]))
	index := b.declare("index", ir.Binary{Operator: ir.Subtract, Left: ir.Length{Array: source}, Right: ir.NumberConstant{Value: 1}})
	item := ir.Expression(ir.ArrayIndex{Array: source, Index: b.read(index), Element: element})
	if item.Type().IsMaybe() && item.Type() != element {
		item = ir.Unwrap{Value: item}
	}
	arguments := []ir.Expression{b.read(accumulator), item, b.read(index), source}
	for index, parameter := range signatures[0].Parameters() {
		takes, known := l.representation(l.checker.GetTypeOfSymbol(parameter))
		if !known || slotless(takes) {
			return nil, true, l.notYet(node, "reduceRight with an unrepresented callback parameter")
		}
		arguments[index] = fit(arguments[index], takes)
		if arguments[index].Type() != takes {
			return nil, true, l.notYet(node, "reduceRight with a callback parameter of another representation")
		}
	}
	call := ir.CallClosure{Closure: b.read(b.parameters[1]), Arguments: arguments, Returns: result}
	b.body = append(b.body, ir.Loop{Condition: ir.Binary{Operator: ir.GreaterOrEqual, Left: b.read(index), Right: ir.NumberConstant{}}, Body: []ir.Statement{
		ir.If{Condition: ir.Binary{Operator: ir.Less, Left: b.read(index), Right: ir.Length{Array: source}}, Then: []ir.Statement{ir.Assign{Local: accumulator, Value: call}}},
	}, Update: []ir.Statement{ir.Assign{Local: index, Value: ir.Binary{Operator: ir.Subtract, Left: b.read(index), Right: ir.NumberConstant{Value: 1}}}}})
	return b.finish("array_reduce_right", b.read(accumulator)), true, nil
}
