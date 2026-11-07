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
func (lowering *lowering) arrayFrom(node *ast.Node) (ir.Expression, bool, error) {
	arguments := node.AsCallExpression().Arguments.Nodes
	// A call to a program's generator is collected by its origin, before the declared-iterator
	// paths below can read its type (#qk8rztp, Compiler's risk 2).
	if len(arguments) > 0 {
		if function, isGenerator, generatorCalleeError := lowering.generatorCallee(
			arguments[0],
		); generatorCalleeError != nil {
			return nil, true, generatorCalleeError
		} else if isGenerator {
			if len(arguments) > 1 {
				return nil, true, lowering.generatorNotYet(node, "consumed by Array.from with a mapper")
			}
			element, elementTypeError := lowering.elementType(node)
			if elementTypeError != nil {
				return nil, true, elementTypeError
			}
			value, yields, collectGeneratorError := lowering.collectGenerator(arguments[0], function)
			if collectGeneratorError == nil && yields != element {
				return nil, true, lowering.generatorNotYet(
					node,
					"consumed by Array.from into an array held differently from what it yields",
				)
			}
			return value, true, collectGeneratorError
		}
	}
	// Collection iterators already yield fitted slots. Preserve those references
	// when copying, including boxed union values.
	if len(arguments) == 1 {
		if element, iterator := lowering.libraryIteratorElement(arguments[0]); iterator {
			value, expressionError := lowering.expression(arguments[0])
			if expressionError != nil {
				return nil, true, expressionError
			}
			return lowering.libraryIteratorArray(node, value, element), true, nil
		}
	}
	if len(arguments) > 0 {
		if plan, planIterationError := lowering.planIteration(arguments[0]); planIterationError != nil {
			return nil, true, planIterationError
		} else if plan != nil {
			if len(arguments) > 2 {
				return nil, true, lowering.notYet(node, "Array.from with thisArg")
			}
			element, elementTypeError := lowering.elementType(node)
			if elementTypeError != nil {
				return nil, true, elementTypeError
			}
			var callback *ast.Node
			if len(arguments) == 2 {
				callback = arguments[1]
			}
			value, collectIterationError := lowering.collectIteration(node, arguments[0], callback, plan, element)
			return value, true, collectIterationError
		}
	}
	if len(arguments) != 2 {
		return nil, true, lowering.notYet(node, "Array.from with other than { length } and a callback")
	}
	source := ast.SkipParentheses(arguments[0])
	if source.Kind != ast.KindObjectLiteralExpression || len(source.AsObjectLiteralExpression().Properties.Nodes) != 1 {
		return nil, true, lowering.notYet(arguments[0], "Array.from of anything but { length }")
	}
	property := source.AsObjectLiteralExpression().Properties.Nodes[0]
	if (property.Kind != ast.KindPropertyAssignment && property.Kind != ast.KindShorthandPropertyAssignment) ||
		!ast.IsIdentifier(property.Name()) ||
		property.Name().Text() != "length" {
		return nil, true, lowering.notYet(arguments[0], "Array.from of anything but { length }")
	}
	element, err := lowering.elementType(node)
	if err != nil {
		return nil, true, err
	}
	var length ir.Expression
	if property.Kind == ast.KindPropertyAssignment {
		length, err = lowering.expression(property.AsPropertyAssignment().Initializer)
	} else {
		length, err = lowering.shorthand(property)
	}
	if err != nil {
		return nil, true, err
	}
	if length.Type() != ir.Number {
		return nil, true, lowering.notYet(property, "Array.from with a length that isn't a number")
	}
	callback := ast.SkipParentheses(arguments[1])
	if callback.Kind != ast.KindArrowFunction && callback.Kind != ast.KindFunctionExpression {
		// A function value from elsewhere would be typed (value: unknown, index: number), and an
		// unknown parameter isn't something stage 0 holds.
		return nil, true, lowering.notYet(
			arguments[1],
			"Array.from with a callback that isn't an arrow function written in place",
		)
	}
	if parameters := callback.Parameters(); len(parameters) > 0 && ast.IsIdentifier(parameters[0].Name()) {
		received := lowering.checker.GetTypeAtLocation(parameters[0].Name())
		switch {
		case parameters[0].AsParameterDeclaration().Initializer != nil:
			// The incoming undefined selects the default before the body reads this parameter.
		case received.Flags()&(checker.TypeFlagsUnknown|checker.TypeFlagsUndefined) != 0:
			if lowering.alwaysUndefined == nil {
				lowering.alwaysUndefined = map[*ast.Symbol]bool{}
			}
			lowering.alwaysUndefined[lowering.symbol(parameters[0].Name())] = true
		case !lowering.includesUndefined(received):
			// The checker reads { length } as an array-like of whatever the parameter says, but
			// it has no elements: the parameter is undefined every time, whatever its type claims.
			return nil, true, &Refused{
				Where: lowering.program.Where(parameters[0]),
				What: "a first Array.from parameter typed " +
					lowering.checker.TypeToString(received) +
					", which is undefined every time",
				Fix: "name it _ and leave it untyped, (_, index) => ..., or type it undefined (the type would be a lie the checker can't see)",
			}
		}
	}
	mapped, err := lowering.expression(callback)
	if err != nil {
		return nil, true, err
	}
	// What the callback receives first is undefined, held as its first parameter holds undefined: a
	// null reference, or a maybe scalar's packed word.
	var first ir.Type
	if closure, isClosure := mapped.(ir.MakeClosure); isClosure {
		if parameters := lowering.result.Functions[closure.Function].Parameters; len(parameters) > 0 {
			first = lowering.result.Locals[parameters[0]].Type
		}
	}
	return ir.ArrayFrom{
		Length:       length,
		Callback:     mapped,
		Element:      element,
		First:        first,
		CallbackType: int(lowering.concrete(lowering.checker.GetTypeAtLocation(callback)).Id()),
	}, true, nil
}
