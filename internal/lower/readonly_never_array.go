package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Only a literal with no slots needs no element representation. Its array
// identity is real; invariance prevents a shared never[] from gaining writable
// inhabited elements. A contextual concrete array retains its normal layout.
func (l *lowering) emptyNeverArrayLiteral(node *ast.Node) bool {
	if len(node.AsArrayLiteralExpression().Elements.Nodes) != 0 {
		return false
	}
	proven := l.concrete(l.checker.GetContextualType(node, checker.ContextFlagsNone))
	if proven == nil {
		proven = l.concrete(l.checker.GetTypeAtLocation(node))
	}
	return l.checker.IsArrayType(proven) && l.checker.GetElementTypeOfArrayType(proven).Flags()&checker.TypeFlagsNever != 0
}
