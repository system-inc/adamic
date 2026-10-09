package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

func (l *lowering) libraryArrayIsArray(node *ast.Node) (ir.Expression, bool, error) {
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if callee.Kind != ast.KindPropertyAccessExpression || callee.Name().Text() != "isArray" || !l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Array") {
		return nil, false, nil
	}
	written := node.AsCallExpression().Arguments.Nodes
	if len(written) != 1 || written[0].Kind == ast.KindSpreadElement {
		return nil, true, l.notYet(node, "Array.isArray without one explicit value")
	}
	argument := ast.SkipParentheses(written[0])
	// These intrinsic identities cannot be replaced and evaluating them has no effects.
	for _, global := range []string{"Object", "Array", "String", "Number", "JSON", "Math", "Map", "Set"} {
		if l.isLibraryGlobal(argument, global) {
			return ir.BooleanConstant{Value: false}, true, nil
		}
	}
	if argument.Kind == ast.KindPropertyAccessExpression && argument.Name().Text() == "prototype" && l.isLibraryGlobal(argument.AsPropertyAccessExpression().Expression, "Array") {
		return ir.BooleanConstant{Value: true}, true, nil
	}
	// An empty literal has no element representation and no operand effects.
	if argument.Kind == ast.KindArrayLiteralExpression && len(argument.AsArrayLiteralExpression().Elements.Nodes) == 0 {
		return ir.BooleanConstant{Value: true}, true, nil
	}
	proven := l.checker.GetTypeAtLocation(argument)
	isArray := l.checker.IsArrayType(proven) || l.tupleType(argument) != nil
	value, err := l.expression(argument)
	if err != nil {
		return nil, true, err
	}
	if !isArray && value.Type() == ir.Object {
		if _, missing := value.(ir.Undefined); !missing {
			if _, null := value.(ir.Null); !null && !l.libraryArrayExactShape(argument) {
				return nil, true, l.notYet(node, "Array.isArray on an object view that can hide an array")
			}
		}
	}
	if value.Type() == ir.Union || value.Type() == ir.Array && !isArray || l.includesUndefined(proven) && isArray || l.includesNull(proven) && isArray {
		return nil, true, l.notYet(node, "Array.isArray on a mixed or optional representation")
	}
	// Bind the operand before returning the proven fact. This keeps effects and
	// ownership without inventing a pointer test for a primitive representation.
	b := l.libraryArrayBuilder([]ir.Expression{value})
	return b.finish("array_is_array", ir.BooleanConstant{Value: isArray}), true, nil
}
