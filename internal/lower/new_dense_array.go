package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// A sole numeric argument creates holes, not an element or initialized storage.
// Dense forms use existing ArrayLiteral IR and its ordered operand evaluation.
func (l *lowering) newDenseArray(node *ast.Node) (ir.Expression, error) {
	element, err := l.elementType(node)
	if err != nil {
		return nil, err
	}
	written := nodesOf(node.AsNewExpression().Arguments)
	for _, argument := range written {
		if argument.Kind == ast.KindSpreadElement {
			return nil, l.notYet(node, "a spread into an Array constructor")
		}
	}
	if len(written) == 1 {
		argument, err := l.expression(written[0])
		if err != nil {
			return nil, err
		}
		if argument.Type() == ir.Number || argument.Type() == ir.MaybeNumber {
			return nil, l.notYet(node, "an Array constructor creating holes")
		}
		if argument = fit(argument, element); argument.Type() != element {
			return nil, l.notYet(node, "an Array constructor element with a different native representation")
		}
		return ir.ArrayLiteral{Element: element, Elements: []ir.Expression{argument}}, nil
	}
	literal := ir.ArrayLiteral{Element: element}
	for _, argument := range written {
		value, err := l.expression(argument)
		if err != nil {
			return nil, err
		}
		if value = fit(value, element); value.Type() != element {
			return nil, l.notYet(node, "an Array constructor element with a different native representation")
		}
		literal.Elements = append(literal.Elements, value)
	}
	return literal, nil
}
