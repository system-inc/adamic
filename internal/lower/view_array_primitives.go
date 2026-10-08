package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// This exemption belongs only to reads. A union cannot authorize changing an
// allocation's physical element slots or a write through a different view.
func (l *lowering) primitiveArrayIndexRead(node *ast.Node, array ir.Expression) (ir.Expression, bool, error) {
	access := node.AsElementAccessExpression()
	declared := l.untaggedArrayElement(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(access.Expression)))
	if declared == nil {
		return nil, false, nil
	}
	if of, known := l.representation(declared); !known || of != ir.Union {
		return nil, false, nil
	}
	if _, err := l.viewContract(node, declared); err != nil {
		return nil, true, err
	}
	if !l.viewPrimitiveUnionRead(declared) {
		return nil, false, nil
	}
	index, err := l.expression(access.ArgumentExpression)
	if err != nil {
		return nil, true, err
	}
	if index.Type() != ir.Number {
		return nil, true, l.notYet(node, "an array index that isn't a number")
	}
	l.result.PrimitiveArrayReads = true
	read := l.markViewArrayRead(node, ir.ArrayIndex{Array: array, Index: index, Element: ir.Union})
	return l.defined(node, read), true, nil
}
