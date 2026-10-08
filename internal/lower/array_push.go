package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// All arguments run before the first intrinsic write. Keep the existing
// one-value IR; a multi-value push binds its operands before pushing in order.
func (l *lowering) libraryArrayPush(node, receiver *ast.Node, array ir.Expression, element ir.Type) (ir.Expression, bool, error) {
	arguments := []ir.Expression{array}
	for _, written := range node.AsCallExpression().Arguments.Nodes {
		value, err := l.expression(written)
		if err != nil {
			return nil, true, err
		}
		value = fit(value, element)
		if value.Type() != element {
			return nil, true, l.notYet(written, "push with a value of another stored element type")
		}
		arguments = append(arguments, value)
	}
	site := l.writeSite(receiver)
	if len(arguments) == 2 {
		return ir.ArrayPush{Array: array, Value: arguments[1], Element: element, Site: site}, true, nil
	}
	b := l.libraryArrayBuilder(arguments)
	source := b.read(b.parameters[0])
	for _, parameter := range b.parameters[1:] {
		b.body = append(b.body, ir.Evaluate{Value: ir.ArrayPush{Array: source, Value: b.read(parameter), Element: element, Site: site}})
	}
	return b.finish("array_push_values", ir.Length{Array: source}), true, nil
}
