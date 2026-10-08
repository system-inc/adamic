package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// A primitive-union array declaration owns boxed storage even when its first
// initializer happens to be homogeneous. Later writes through the declaration
// can change primitive kind, and every existing alias must see that mutation.
func (l *lowering) boxedPrimitiveArrayContext(node *ast.Node, array *checker.Type) bool {
	element := l.concrete(l.checker.GetElementTypeOfArrayType(array))
	of, known := l.representation(element)
	if !known || of != ir.Union || !interfaceScalar(element) {
		return false
	}
	id, err := l.viewContract(node, element)
	return err == nil && ir.PrimitiveArrayContract(l.result, id)
}
