package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Only the complete primitive domain justifies the existing intrinsic signature.
// A literal, brand, enum or aggregate must not borrow its physical representation
// as a logical parameter certificate. Unknown domains remain demand-time failures.
func (l *lowering) viewSetCallableElement(node *ast.Node) ir.Type {
	arguments := l.typeArguments(l.concrete(l.checker.GetTypeAtLocation(node)))
	if len(arguments) != 1 {
		return 0
	}
	element := l.concrete(arguments[0])
	for _, candidate := range []struct {
		proven *checker.Type
		of     ir.Type
	}{
		{l.checker.GetNumberType(), ir.Number},
		{l.checker.GetBooleanType(), ir.Boolean},
		{l.checker.GetStringType(), ir.String},
	} {
		if checker.Checker_isTypeIdenticalTo(l.checker, element, candidate.proven) {
			return candidate.of
		}
	}
	return 0
}
