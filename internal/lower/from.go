package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// arrayFrom lowers Array.from({ length }, (_, index) => ...), the one form of Array.from 0.1 has
// (docs/0.1.md): a new array of length elements, each what the callback returns for its index.
//
// The object is never made. It has no elements, so what the callback receives first is always
// undefined; the checker types that parameter unknown, and it's lowered as a reference that's always
// missing (alwaysUndefined), which is exactly what it holds.
func (l *lowering) arrayFrom(node *ast.Node) (ir.Expression, bool, error) {
	arguments := node.AsCallExpression().Arguments.Nodes
	if len(arguments) != 2 {
		return nil, true, l.notYet(node, "Array.from with other than { length } and a callback")
	}
	source := ast.SkipParentheses(arguments[0])
	if source.Kind != ast.KindObjectLiteralExpression || len(source.AsObjectLiteralExpression().Properties.Nodes) != 1 {
		return nil, true, l.notYet(arguments[0], "Array.from of anything but { length }")
	}
	property := source.AsObjectLiteralExpression().Properties.Nodes[0]
	if (property.Kind != ast.KindPropertyAssignment && property.Kind != ast.KindShorthandPropertyAssignment) || !ast.IsIdentifier(property.Name()) || property.Name().Text() != "length" {
		return nil, true, l.notYet(arguments[0], "Array.from of anything but { length }")
	}
	element, err := l.elementType(node)
	if err != nil {
		return nil, true, err
	}
	var length ir.Expression
	if property.Kind == ast.KindPropertyAssignment {
		length, err = l.expression(property.AsPropertyAssignment().Initializer)
	} else {
		length, err = l.shorthand(property)
	}
	if err != nil {
		return nil, true, err
	}
	if length.Type() != ir.Number {
		return nil, true, l.notYet(property, "Array.from with a length that isn't a number")
	}
	callback := ast.SkipParentheses(arguments[1])
	if callback.Kind != ast.KindArrowFunction {
		// A function value from elsewhere would be typed (value: unknown, index: number), and an
		// unknown parameter isn't something stage 0 holds.
		return nil, true, l.notYet(arguments[1], "Array.from with a callback that isn't an arrow function written in place")
	}
	if parameters := callback.Parameters(); len(parameters) > 0 && ast.IsIdentifier(parameters[0].Name()) {
		if received := l.checker.GetTypeAtLocation(parameters[0].Name()); received.Flags()&(checker.TypeFlagsUnknown|checker.TypeFlagsUndefined) != 0 {
			if l.alwaysUndefined == nil {
				l.alwaysUndefined = map[*ast.Symbol]bool{}
			}
			l.alwaysUndefined[l.symbol(parameters[0].Name())] = true
		}
	}
	mapped, err := l.expression(callback)
	if err != nil {
		return nil, true, err
	}
	return ir.ArrayFrom{Length: length, Callback: mapped, Element: element}, true, nil
}
