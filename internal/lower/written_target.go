package lower

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// writtenTarget reports whether node is written where it stands: an assignment target by any road
// ast.IsAssignmentTarget follows (=, compound and logical assignments, ++ and --, destructuring targets,
// for-in and for-of heads, through parentheses and non-null), or one wrapped in a type assertion, which
// IsAssignmentTarget stops at though `(node as T) = other` and `[node satisfies T] = [other]` write node
// too, or the operand of delete, which removes it. Each wrapper is climbed in turn, as overloadCallbackAssigned
// does (Language's attack on the namespace export write stop, #8n7eet3).
func writtenTarget(node *ast.Node) bool {
	for target := node; target.Parent != nil; target = target.Parent {
		if ast.IsAssignmentTarget(target) || target.Parent.Kind == ast.KindDeleteExpression {
			return true
		}
		switch target.Parent.Kind {
		case ast.KindParenthesizedExpression, ast.KindNonNullExpression, ast.KindAsExpression, ast.KindSatisfiesExpression, ast.KindTypeAssertionExpression:
			continue
		}
		return false
	}
	return false
}
