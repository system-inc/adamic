package lower

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// Library replaces this hook when its node host argument consumption proof lands.
func (l *lowering) nodeHostConsumesArgument(call *ast.Node, index int) bool { return false }
