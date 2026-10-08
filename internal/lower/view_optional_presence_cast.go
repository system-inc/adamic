package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// A required scalar slot admitting undefined may be read as an optional slot
// with the same read type. Exact optional assignability rejects the presence
// difference, but it is not evidence that a runtime read can be trusted.
// Keep the existing view checks and accept no change to a source slot's type.
func (l *lowering) optionalPresenceViewCandidate(_ *ast.Node, source, target *checker.Type) bool {
	if source.Flags()&checker.TypeFlagsObject == 0 || target.Flags()&checker.TypeFlagsObject == 0 || isClassInstance(source) || isClassInstance(target) || l.callableViewContract(source) || l.callableViewContract(target) || len(l.checker.GetIndexInfosOfType(source)) != 0 || len(l.checker.GetIndexInfosOfType(target)) != 0 || l.checker.IsArrayType(source) || l.checker.IsArrayType(target) || checker.IsTupleType(source) || checker.IsTupleType(target) {
		return false
	}
	changed := false
	for _, original := range l.checker.GetPropertiesOfType(source) {
		viewed := l.checker.GetPropertyOfType(target, original.Name)
		if viewed == nil || accessorSymbol(original) || accessorSymbol(viewed) || original.Flags&ast.SymbolFlagsMethod != 0 || viewed.Flags&ast.SymbolFlagsMethod != 0 || l.checker.IsReadonlySymbol(original) != l.checker.IsReadonlySymbol(viewed) {
			return false
		}
		from, to := l.checker.GetTypeOfSymbol(original), l.checker.GetTypeOfSymbol(viewed)
		if from != to && (!interfaceScalar(from) || !interfaceScalar(to) || !l.checker.IsTypeAssignableTo(from, to) || !l.checker.IsTypeAssignableTo(to, from)) {
			return false
		}
		if original.Flags&ast.SymbolFlagsOptional == 0 && viewed.Flags&ast.SymbolFlagsOptional != 0 {
			if !interfaceScalar(from) || !l.includesUndefined(from) {
				return false
			}
			changed = true
		} else if original.Flags&ast.SymbolFlagsOptional != viewed.Flags&ast.SymbolFlagsOptional {
			return false
		}
	}
	return changed
}

func (l *lowering) optionalPresenceViewCast(node *ast.Node, value ir.Expression, source, target *checker.Type) (ir.Expression, error) {
	if value.Type() != ir.Object || !l.optionalPresenceViewCandidate(node, source, target) {
		return nil, nil
	}
	return l.view(node, value, target)
}
