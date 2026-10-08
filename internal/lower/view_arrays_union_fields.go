package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Reverse assignability and writable-slot checks apply to every arm. Admission
// only installs lazy views; it neither selects an arm nor inspects its payload.
func (l *lowering) arrayCommonUnionCastCandidate(source, target *checker.Type) bool {
	if source.Flags()&checker.TypeFlagsObject == 0 || target.Flags()&checker.TypeFlagsUnion == 0 || l.callableViewContract(source) {
		return false
	}
	for _, member := range target.Types() {
		if member.Flags()&checker.TypeFlagsObject == 0 || isClassInstance(member) || l.callableViewContract(member) || !l.checker.IsTypeAssignableTo(member, source) || l.widened(member, source, map[[2]*checker.Type]bool{}) != nil {
			return false
		}
	}
	for _, field := range l.checker.GetPropertiesOfType(target) {
		if field.Flags&ast.SymbolFlagsOptional == 0 && !accessorSymbol(field) && l.viewArrayBase(l.checker.GetTypeOfSymbol(field)) != nil {
			return true
		}
	}
	return false
}
func (l *lowering) viewArrayCommonUnionCast(node *ast.Node, value ir.Expression, source, target *checker.Type) (ir.Expression, error) {
	if value.Type() != ir.Object || !l.arrayCommonUnionCastCandidate(source, target) {
		return nil, nil
	}
	return l.view(node, value, target)
}
