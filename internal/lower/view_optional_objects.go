package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// An optional own data slot supplies either its object or undefined. Its shape
// retains presence; reading it never invents or removes an own property. The
// recursively registered child fields keep their checks through saved aliases.
func (l *lowering) optionalObjectViewRead(node *ast.Node, field *ast.Symbol) bool {
	if field.Flags&ast.SymbolFlagsOptional == 0 || accessorSymbol(field) || node.AsPropertyAccessExpression().QuestionDotToken != nil {
		return false
	}
	declared := l.checker.GetTypeOfSymbol(field)
	if declared.Flags()&checker.TypeFlagsUnion == 0 {
		return false
	}
	objects, undefined := 0, false
	for _, member := range declared.Types() {
		if member.Flags()&checker.TypeFlagsUndefined != 0 {
			undefined = true
			continue
		}
		of, known := l.representation(member)
		if !known || of != ir.Object || member.Flags()&checker.TypeFlagsObject == 0 || isClassInstance(member) || len(l.checker.GetSignaturesOfType(member, checker.SignatureKindCall)) != 0 {
			return false
		}
		objects++
	}
	return undefined && objects == 1
}
