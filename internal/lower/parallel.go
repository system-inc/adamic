package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

func (l *lowering) parallelMap(node *ast.Node) (ir.Expression, error) {
	return nil, l.notYet(node, "parallelMap without its Shareable and effect proof")
}
