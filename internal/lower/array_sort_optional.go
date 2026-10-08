package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Undefined selects the host's existing default string order. Capture both
// operands before testing presence so an effectful callback expression runs once.
func (l *lowering) optionalStringSort(node *ast.Node, array ir.Expression, element ir.Type) (ir.Expression, bool, error) {
	written := node.AsCallExpression().Arguments.Nodes
	if len(written) != 1 || !l.includesUndefined(l.checker.GetTypeAtLocation(written[0])) {
		return nil, false, nil
	}
	if element != ir.String {
		return nil, true, l.notYet(node, "an optional sort comparator on non-string elements; narrow the comparator to a present function before sorting")
	}
	proven := l.concrete(l.checker.GetTypeAtLocation(written[0]))
	if l.includesNull(proven) {
		return nil, true, l.notYet(written[0], "a sort comparator that may be null")
	}
	present := l.checker.GetNonNullableType(proven)
	signatures := l.checker.GetSignaturesOfType(present, checker.SignatureKindCall)
	onlyUndefined := proven.Flags()&checker.TypeFlagsUndefined != 0
	if !onlyUndefined {
		if len(signatures) != 1 {
			return nil, true, l.notYet(written[0], "an optional sort comparator without one callable signature")
		}
		if returns, known := l.representation(l.checker.GetReturnTypeOfSignature(signatures[0])); !known || returns != ir.Number {
			return nil, true, l.notYet(written[0], "an optional sort comparator that does not return a number")
		}
		if len(signatures[0].Parameters()) > 2 {
			return nil, true, l.notYet(written[0], "an optional sort comparator requiring more than two elements")
		}
		for _, parameter := range signatures[0].Parameters() {
			if of, known := l.representation(l.checker.GetTypeOfSymbol(parameter)); !known || of != element {
				return nil, true, l.notYet(written[0], "an optional sort comparator with a different parameter representation")
			}
		}
	}
	callback, err := l.expression(written[0])
	if err != nil {
		return nil, true, err
	}
	callback = fit(callback, ir.Closure)
	if callback.Type() != ir.Closure {
		return nil, true, l.notYet(written[0], "an optional sort comparator without closure storage")
	}
	defaultComparator := l.libraryArrayComparator(element)
	b := l.libraryArrayBuilder([]ir.Expression{array, callback})
	source, selected := b.read(b.parameters[0]), b.read(b.parameters[1])
	fallback := ir.ArraySort{Array: source, Element: element, Comparator: defaultComparator, DefaultStrings: true}
	value := ir.Conditional{Condition: ir.IsUndefined{Value: selected}, WhenTrue: fallback, WhenNot: ir.ArraySort{Array: source, Element: element, Callback: selected}}
	return b.finish("array_sort_optional_string_comparator", value), true, nil
}
