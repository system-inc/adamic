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
	if len(arguments) > 0 {
		if plan, err := l.planIteration(arguments[0]); err != nil {
			return nil, true, err
		} else if plan != nil {
			if len(arguments) > 2 {
				return nil, true, l.notYet(node, "Array.from with thisArg")
			}
			element, err := l.elementType(node)
			if err != nil {
				return nil, true, err
			}
			var callback *ast.Node
			if len(arguments) == 2 {
				callback = arguments[1]
			}
			value, err := l.collectIteration(node, arguments[0], callback, plan, element)
			return value, true, err
		}
	}
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
	if callback.Kind != ast.KindArrowFunction && callback.Kind != ast.KindFunctionExpression {
		// A function value from elsewhere would be typed (value: unknown, index: number), and an
		// unknown parameter isn't something stage 0 holds.
		return nil, true, l.notYet(arguments[1], "Array.from with a callback that isn't an arrow function written in place")
	}
	if parameters := callback.Parameters(); len(parameters) > 0 && ast.IsIdentifier(parameters[0].Name()) {
		received := l.checker.GetTypeAtLocation(parameters[0].Name())
		switch {
		case received.Flags()&(checker.TypeFlagsUnknown|checker.TypeFlagsUndefined) != 0:
			if l.alwaysUndefined == nil {
				l.alwaysUndefined = map[*ast.Symbol]bool{}
			}
			l.alwaysUndefined[l.symbol(parameters[0].Name())] = true
		case !l.includesUndefined(received):
			// The checker reads { length } as an array-like of whatever the parameter says, but
			// it has no elements: the parameter is undefined every time, whatever its type claims.
			return nil, true, &Refused{Where: l.program.Where(parameters[0]), What: "a first Array.from parameter typed " + l.checker.TypeToString(received) + ", which is undefined every time", Fix: "name it _ and leave it untyped, (_, index) => ..., or type it undefined (the type would be a lie the checker can't see)"}
		}
	}
	mapped, err := l.expression(callback)
	if err != nil {
		return nil, true, err
	}
	// What the callback receives first is undefined, held as its first parameter holds undefined: a
	// null reference, or number | undefined's packed word.
	var first ir.Type
	if closure, isClosure := mapped.(ir.MakeClosure); isClosure {
		if parameters := l.result.Functions[closure.Function].Parameters; len(parameters) > 0 {
			first = l.result.Locals[parameters[0]].Type
		}
	}
	return ir.ArrayFrom{Length: length, Callback: mapped, Element: element, First: first}, true, nil
}
