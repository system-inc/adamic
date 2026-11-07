package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

// nonNull uses the same nullish test and terminal panic as ?? panic(...).
// Recover presence checks from nullable loads, but keep a checked union narrowing
// in its result representation. Calls may invalidate the original narrowing.
func (lowering *lowering) nonNull(node *ast.Node) (ir.Expression, error) {
	operand := node.AsNonNullExpression().Expression
	value, expressionError := lowering.expression(operand)
	if expressionError != nil {
		return nil, expressionError
	}
	return lowering.nonNullValue(node, value)
}

// nonNullValue also checks an already-held assignment target read.
func (lowering *lowering) nonNullValue(node *ast.Node, value ir.Expression) (ir.Expression, error) {
	operand := node.AsNonNullExpression().Expression
	var typeOfError error
	for {
		switch narrowed := value.(type) {
		case ir.Unwrap:
			value = narrowed.Value
		case ir.Defined:
			value = narrowed.Value
		default:
			// Narrow and narrowing calls retain their representation and checks.
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
	if property, ok := value.(ir.Property); ok &&
		ast.SkipParentheses(operand).Kind == ast.KindPropertyAccessExpression {
		if symbol := lowering.checker.GetSymbolAtLocation(ast.SkipParentheses(operand).Name()); symbol != nil {
			if declared, known := lowering.representation(lowering.checker.GetTypeOfSymbol(symbol)); known {
				property.Of = declared
				value = property
			}
		}
	}
	of, typeOfError := lowering.typeOf(node)
	flags := lowering.checker.GetTypeAtLocation(operand).Flags()
	if typeOfError != nil && (flags == checker.TypeFlagsUndefined || flags == checker.TypeFlagsNull) {
		of = ir.Object
		if contextual := lowering.slotType(node); contextual != nil {
			if representation, known := lowering.representation(contextual); known {
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
		typeOfError = nil
	}
	if typeOfError != nil {
		return nil, typeOfError
	}
	// A checked narrowing helper can already return a plain scalar. Its result
	// has no nullish representation and must never receive a pointer test.
	if value.Type() == ir.Number || value.Type() == ir.Boolean {
		lowering.recordNonNullCheck(node, true)
		return value, nil
	}
	proven := lowering.checker.GetTypeAtLocation(operand)
	if !weakOperand &&
		!lowering.includesUndefined(proven) &&
		!lowering.includesNull(proven) &&
		!lowering.narrowedAway(ast.SkipParentheses(operand)) &&
		!value.Type().IsMaybe() {
		if value.Type() == ir.Union && of != ir.Union {
			lowering.recordNonNullCheck(node, true)
			return ir.Narrow{Value: value, To: of}, nil
		}
		lowering.recordNonNullCheck(node, true)
		return value, nil
	}
	file := ast.GetSourceFileOfNode(node)
	text := file.Text()[scanner.GetTokenPosOfNode(node, file, false):node.End()]
	message := ir.StringConstant{Index: lowering.constant(
		"non-null assertion failed at " + lowering.program.Where(node) + ": " + text + " is null or undefined",
	)}
	result := ir.Expression(ir.Coalesce{Value: value, Panic: message, Of: value.Type().Present()})
	if result.Type() == ir.Union && of != ir.Union {
		result = ir.Narrow{Value: result, To: of}
	}
	lowering.recordNonNullCheck(node, false)
	return result, nil
}

func (lowering *lowering) checkedAssertionSource(node *ast.Node) bool {
	return strings.HasSuffix(lowering.program.FileName(ast.GetSourceFileOfNode(node)), ".ts")
}

func (lowering *lowering) recordNonNullCheck(node *ast.Node, proven bool) {
	counts := &lowering.result.NonNullChecks
	if proven {
		counts.Proven++
	} else {
		counts.Checked++
	}
	counts.Sites = append(
		counts.Sites,
		ir.NonNullCheck{Where: lowering.program.Where(node), Expression: sourceExpression(node), Proven: proven},
	)
}
