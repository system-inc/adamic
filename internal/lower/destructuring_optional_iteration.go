package lower

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// Map and Set constructors treat undefined as an empty input. Ordinary iteration
// does not, so this fallback belongs only to an argument of the library constructor.
func (l *lowering) collectionAllowsAbsent(node *ast.Node) bool {
	for node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression {
		node = node.Parent
	}
	parent := node.Parent
	if parent == nil || parent.Kind != ast.KindNewExpression {
		return false
	}
	callee := ast.SkipParentheses(parent.AsNewExpression().Expression)
	return l.isLibraryGlobal(callee, "Map") || l.isLibraryGlobal(callee, "Set")
}
