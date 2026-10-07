package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/adamic/internal/ir"
)

// nonNull uses the same nullish test and terminal panic as ?? panic(...).
// Recover presence checks from nullable loads, but keep a checked union narrowing
// in its result representation. Calls may invalidate the original narrowing.
func (l *lowering) nonNull(node *ast.Node) (ir.Expression, error) {
	operand := node.AsNonNullExpression().Expression
	value, err := l.expression(operand)
	if err != nil {
		return nil, err
	}
	for {
		switch narrowed := value.(type) {
		case ir.Unwrap:
			value = narrowed.Value
		case ir.Defined:
			value = narrowed.Value
		default:
			goto stored
		}
	}
stored:
	weakOperand := false
	if target, ok := value.(ir.WeakTarget); ok {
		// A narrowed Weak may have cleared since the narrowing. This assertion owns
		// the terminal check and its diagnostic, rather than the generic Weak load.
		target.Present = false
		value, weakOperand = target, true
	}
	if property, ok := value.(ir.Property); ok && ast.SkipParentheses(operand).Kind == ast.KindPropertyAccessExpression {
		if symbol := l.checker.GetSymbolAtLocation(ast.SkipParentheses(operand).Name()); symbol != nil {
			if declared, known := l.representation(l.checker.GetTypeOfSymbol(symbol)); known {
				property.Of = declared
				value = property
			}
		}
	}
	of, err := l.typeOf(node)
	if err != nil && l.uninitializedInitializer(node) {
		of = ir.Object
		if contextual := l.checker.GetContextualType(node, checker.ContextFlagsNone); contextual != nil {
			if representation, known := l.representation(contextual); known {
				of = representation
			}
		}
		if of.IsMaybe() {
			of = of.Present()
		}
		if of == ir.Weak {
			of = ir.Object
		}
		if of == ir.Number || of == ir.Boolean {
			value = ir.MaybeOf{Of: ir.Maybe(of)}
		} else {
			value = ir.Undefined{Of: of}
		}
		err = nil
	}
	if err != nil {
		return nil, err
	}
	// A checked narrowing helper can already return a plain scalar. Its result
	// has no nullish representation and must never receive a pointer test.
	if value.Type() == ir.Number || value.Type() == ir.Boolean {
		return value, nil
	}
	proven := l.checker.GetTypeAtLocation(operand)
	if !weakOperand && !l.includesUndefined(proven) && !l.includesNull(proven) && !l.narrowedAway(ast.SkipParentheses(operand)) && !value.Type().IsMaybe() {
		if value.Type() == ir.Union && of != ir.Union {
			return ir.Narrow{Value: value, To: of}, nil
		}
		return value, nil
	}
	file := ast.GetSourceFileOfNode(node)
	text := file.Text()[scanner.GetTokenPosOfNode(node, file, false):node.End()]
	message := ir.StringConstant{Index: l.constant("non-null assertion failed: " + text + " is null or undefined")}
	result := ir.Expression(ir.Coalesce{Value: value, Panic: message, Of: value.Type().Present()})
	if result.Type() == ir.Union && of != ir.Union {
		result = ir.Narrow{Value: result, To: of}
	}
	return result, nil
}
