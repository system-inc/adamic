package structure

import "github.com/microsoft/TypeScript/tsc/shim/ast"

func AdamicJsx(mode string, node *ast.Node) bool {
	if mode == "value" {
		return isJsxValue(node)
	}
	if mode == "after" {
		return jsxAfterParentheses(node)
	}
	return returnArgumentLooksLikeJsx(node)
}
