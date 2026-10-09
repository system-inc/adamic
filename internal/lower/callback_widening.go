package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Intrinsic calls supply element slots directly. A scalar slot cannot be read
// as the heap-backed union promised by a wider callback parameter.
func (l *lowering) checkArrayCallbackRepresentation(node, receiver *ast.Node, name string, written []*ast.Node, element ir.Type) error {
	positions := []int{0}
	switch name {
	case "map", "filter", "forEach", "find", "findIndex", "some", "every", "flatMap", "findLast", "findLastIndex":
	case "reduce":
		positions = []int{1}
	case "sort", "toSorted":
		positions = []int{0, 1}
	default:
		return nil
	}
	if len(written) == 0 {
		return nil
	}
	signatures := l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(written[0]), checker.SignatureKindCall)
	if len(signatures) != 1 {
		return nil
	}
	source := l.checker.GetElementTypeOfArrayType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(receiver)))
	for _, position := range positions {
		if position >= len(signatures[0].Parameters()) {
			continue
		}
		parameter := signatures[0].Parameters()[position]
		target := l.concrete(l.checker.GetTypeOfSymbol(parameter))
		takes, known := l.representation(target)
		if known && takes == ir.Union && element != ir.Union && !element.IsReference() {
			return &Refused{Where: l.program.Where(written[0]), What: "a " + name + " callback parameter " + parameter.Name + " of type " + l.checker.TypeToString(target) + " receiving array element type " + l.checker.TypeToString(source) + " without a representation adapter", Fix: "annotate parameter " + parameter.Name + " as the element type " + l.checker.TypeToString(source)}
		}
	}
	return nil
}
