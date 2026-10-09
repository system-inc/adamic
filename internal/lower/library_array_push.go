package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// GenericArrayPush in V8's src/builtins/builtins-array.cc appends every supplied
// value in order and returns the new length. Call arguments must all finish
// before the first append, including arguments that mutate the receiver. Dense
// array spreads are snapshotted at their argument position, before later operands
// run; this also preserves push(...receiver). Other iterator dispatch stays refused.
func (l *lowering) libraryArrayPush(node, receiver *ast.Node, array ir.Expression, element ir.Type) (ir.Expression, bool, error) {
	arguments := []ir.Expression{array}
	spreads := map[int]ir.Type{}
	for _, written := range node.AsCallExpression().Arguments.Nodes {
		if written.Kind == ast.KindSpreadElement {
			source := written.AsSpreadElement().Expression
			value, err := l.expression(source)
			if err != nil {
				return nil, true, err
			}
			if value.Type() != ir.Array {
				return nil, true, l.notYet(written, "push non-array spread (compiler iterator argument dispatch is not lowered)")
			}
			item, err := l.elementType(source)
			if err != nil {
				return nil, true, err
			}
			if fit(ir.Read{Of: item}, element).Type() != element {
				return nil, true, l.notYet(written, "push spread with another element representation")
			}
			spreads[len(arguments)] = item
			arguments = append(arguments, ir.ArraySlice{Array: value})
			continue
		}
		value, err := l.expression(written)
		if err != nil {
			return nil, true, err
		}
		value = fit(value, element)
		if value.Type() != element {
			return nil, true, l.notYet(written, "push with a value of another element representation")
		}
		arguments = append(arguments, value)
	}
	b := l.libraryArrayBuilder(arguments)
	source := b.read(b.parameters[0])
	for index, parameter := range b.parameters[1:] {
		if item, spread := spreads[index+1]; spread {
			local := b.local("spread_item", item)
			b.body = append(b.body, ir.ForOf{Iterable: b.read(parameter), Local: local, Element: item, Body: []ir.Statement{ir.Evaluate{Value: ir.ArrayPush{Array: source, Value: fit(b.read(local), element), Element: element, Site: l.writeSite(receiver)}}}})
		} else {
			b.body = append(b.body, ir.Evaluate{Value: ir.ArrayPush{Array: source, Value: b.read(parameter), Element: element, Site: l.writeSite(receiver)}})
		}
	}
	return b.finish("array_push_many", ir.Length{Array: source}), true, nil
}
