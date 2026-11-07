package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Recognition is bound to a declared function and a callable first parameter.
// Naming a function this way grants no trust: its unknown result is checked.
func (lowering *lowering) speculativeCast(node *ast.Node) bool {
	call := ast.SkipParentheses(node.AsAsExpression().Expression)
	if call.Kind != ast.KindCallExpression || len(call.AsCallExpression().Arguments.Nodes) == 0 {
		return false
	}
	callee := ast.SkipParentheses(call.AsCallExpression().Expression)
	symbol := lowering.symbol(callee)
	if symbol == nil || len(symbol.Declarations) != 1 || symbol.Declarations[0].Kind != ast.KindFunctionDeclaration {
		return false
	}
	switch symbol.Name {
	case "lookAhead", "tryParse", "speculationHelper":
	default:
		return false
	}
	signature := lowering.checker.GetResolvedSignature(call)
	if signature == nil || len(signature.Parameters()) == 0 || len(
		lowering.checker.GetSignaturesOfType(
			lowering.checker.GetTypeOfSymbol(signature.Parameters()[0]),
			checker.SignatureKindCall,
		),
	) != 1 {
		return false
	}
	return lowering.concrete(lowering.checker.GetTypeAtLocation(call)).Flags()&checker.TypeFlagsUnknown != 0
}

// Existing union storage, checked narrowing and field views carry the check to
// both backends. The helper receives the completed speculation exactly once.
func (lowering *lowering) checkedSpeculativeResult(node *ast.Node) (ir.Expression, error) {
	call := ast.SkipParentheses(node.AsAsExpression().Expression)
	target := lowering.concrete(lowering.checker.GetTypeAtLocation(node))
	present := lowering.checker.GetNonNullableType(target)
	held, known := lowering.representation(present)
	if !known ||
		(held != ir.Number && held != ir.Boolean && held != ir.String && held != ir.Object) ||
		lowering.includesNull(target) {
		return nil, lowering.notYet(node, "a checked speculative result of type "+lowering.checker.TypeToString(target))
	}
	if held == ir.Object {
		if lowering.checker.GetIndexTypeOfType(present, lowering.checker.GetStringType()) != nil ||
			lowering.checker.GetIndexTypeOfType(present, lowering.checker.GetNumberType()) != nil ||
			isClassInstance(present) ||
			len(lowering.checker.GetPropertiesOfType(present)) == 0 ||
			len(lowering.checker.GetSignaturesOfType(present, checker.SignatureKindCall)) != 0 ||
			len(lowering.checker.GetSignaturesOfType(present, checker.SignatureKindConstruct)) != 0 {
			return nil, lowering.notYet(node, "a speculative object result without a fixed scalar field contract")
		}
		for _, field := range lowering.checker.GetPropertiesOfType(present) {
			if field.Flags&ast.SymbolFlagsOptional != 0 ||
				!interfaceScalar(lowering.checker.GetTypeOfSymbol(field)) ||
				accessorSymbol(field) {
				return nil, lowering.notYet(node, "a speculative object result with an unreifiable field "+field.Name)
			}
		}
	}
	value, expressionError := lowering.expression(call)
	if expressionError != nil {
		return nil, expressionError
	}
	message := "speculative result failed: callback " +
		sourceExpression(call.AsCallExpression().Arguments.Nodes[0]) +
		" at " +
		lowering.program.Where(call) +
		"; expected " +
		lowering.checker.TypeToString(target)
	b := lowering.libraryArrayBuilder([]ir.Expression{fit(value, ir.Union)})
	observed := b.read(b.parameters[0])
	if lowering.includesUndefined(target) {
		absent := ir.Expression(ir.Undefined{Of: held})
		if held == ir.Number || held == ir.Boolean {
			absent = ir.MaybeOf{Of: ir.Maybe(held)}
		}
		b.body = append(
			b.body,
			ir.If{Condition: ir.IsUndefined{Value: observed}, Then: []ir.Statement{ir.Return{Value: absent}}},
		)
	}
	checked := ir.Expression(ir.Narrow{Value: observed, To: held, Checked: true, Message: message})
	if held == ir.Object {
		local := b.declare("speculative_result", checked)
		checked = b.read(local)
		// The speculation's own result is unknown, so the view vouches for nothing before it.
		if _, viewError := lowering.view(node, checked, nil, present); viewError != nil {
			return nil, viewError
		}
		for _, field := range lowering.checker.GetPropertiesOfType(present) {
			of, _ := lowering.representation(lowering.checker.GetTypeOfSymbol(field))
			b.body = append(b.body, ir.Evaluate{Value: ir.Property{
				Object:      checked,
				Name:        field.Name,
				Of:          of,
				View:        message,
				ViewType:    lowering.checker.TypeToString(lowering.checker.GetTypeOfSymbol(field)),
				ViewAllowed: lowering.viewLiterals(lowering.checker.GetTypeOfSymbol(field)),
			}})
		}
	} else if allowed := lowering.viewLiterals(present); len(allowed) != 0 {
		local := b.declare("speculative_result", checked)
		checked = b.read(local)
		var fits ir.Expression
		for _, literal := range allowed {
			test := ir.Expression(ir.Binary{Operator: ir.Equal, Left: checked, Right: literal})
			if fits == nil {
				fits = test
			} else {
				fits = ir.Binary{Operator: ir.Or, Left: fits, Right: test}
			}
		}
		b.body = append(b.body, ir.If{
			Condition: ir.Unary{Operator: ir.Not, Operand: fits},
			Then:      []ir.Statement{ir.Panic{Message: ir.StringConstant{Index: lowering.constant(message)}}},
		})
	}
	if lowering.includesUndefined(target) && (held == ir.Number || held == ir.Boolean) {
		checked = ir.MaybeOf{Of: ir.Maybe(held), Value: checked}
	}
	return b.finish("checked_speculative_result", checked), nil
}
