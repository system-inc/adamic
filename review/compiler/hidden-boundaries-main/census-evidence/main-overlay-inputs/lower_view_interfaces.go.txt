package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Interface inheritance is a structural contract, never a construction proof.
// Re-enter the data walker with the same memo table so recursive interfaces keep
// checks without expanding their schema indefinitely.
func (l *lowering) viewInterfaceFields(node *ast.Node, target *checker.Type, fields map[string]bool, seen map[*checker.Type]bool) error {
	return l.viewObjectFields(node, target, fields, seen, false)
}
