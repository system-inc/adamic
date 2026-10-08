package lower

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// Only the fs host argument positions proven to inspect without retaining or writing
// may consume a structural options view; the host lowerer validates its exact layout.
func (l *lowering) nodeHostConsumesArgument(call *ast.Node, index int) bool {
	arguments := call.AsCallExpression().Arguments.Nodes
	return index >= 0 && index < len(arguments) && l.nodeFSFileReadOnlyArgument(arguments[index])
}
