package lower

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// A fresh literal cannot escape this call or expose its lastIndex. Only a string
// literal is admitted; evaluating a binding or expression could change the input.
func regexLiteralTestInput(node *ast.Node) (string, bool) {
	if node.Kind != ast.KindRegularExpressionLiteral {
		return "", false
	}
	property := node.Parent
	if property == nil || property.Kind != ast.KindPropertyAccessExpression || property.Name().Text() != "test" ||
		property.AsPropertyAccessExpression().Expression != node || property.AsPropertyAccessExpression().QuestionDotToken != nil {
		return "", false
	}
	call := property.Parent
	if call == nil || call.Kind != ast.KindCallExpression || call.AsCallExpression().Expression != property || call.AsCallExpression().QuestionDotToken != nil || len(call.AsCallExpression().Arguments.Nodes) != 1 {
		return "", false
	}
	input := ast.SkipParentheses(call.AsCallExpression().Arguments.Nodes[0])
	if input.Kind != ast.KindStringLiteral && input.Kind != ast.KindNoSubstitutionTemplateLiteral {
		return "", false
	}
	return input.Text(), true
}
