package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// logicalNullLeft is what left && x gives where the left is falsy and the result is a union, when
// the left is held as a bare reference that may be null: the boxed null, as nullableObjectBoundary
// writes it. A reference whose type holds null and not undefined holds its null as the null
// pointer, which a union reads as undefined, so ir.Logical keeping the left made
// `none && none.ready` over an Item | null that's null read as undefined: === null answered false
// natively where Node answers true, and a read narrowed to boolean past it panicked (#ayvxxpc).
// A reference whose type holds both can't say which its null pointer is, so it's NotYet. It's nil,
// with no error, when the left isn't such a reference, and ir.Logical keeps it as before: a string
// or a union says its own falsy value, and a reference that may be only undefined holds it as the
// null pointer a union reads as undefined too.
func (l *lowering) logicalNullLeft(leftNode *ast.Node, left ir.Expression) (ir.Expression, error) {
	held := left.Type()
	if !held.IsReference() || held == ir.String || held == ir.Union || held == ir.Weak {
		return nil, nil
	}
	own := l.concrete(l.checker.GetTypeAtLocation(leftNode))
	if !l.includesNull(own) {
		return nil, nil
	}
	if l.includesUndefined(own) {
		return nil, l.notYet(leftNode, "a logical "+typeName(held)+" operand that may be null or undefined, kept in a union result")
	}
	return ir.Box{Value: ir.Null{}}, nil
}
