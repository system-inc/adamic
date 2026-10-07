package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// The shared dispatcher must store these ids as a conjunction, never a union.
// This helper does not enable cast admission or certify any runtime payload.
func (l *lowering) viewIntersectionContracts(node *ast.Node, target *checker.Type, build viewContractBuilder) ([]ir.ViewContractID, error) {
	if target.Flags()&checker.TypeFlagsIntersection == 0 || l.phantomBase(target) != nil {
		return nil, l.notYet(node, "a nonprimitive intersection checked-view contract")
	}
	if build == nil {
		return nil, l.notYet(node, "an intersection checked view without a member builder")
	}
	var members []ir.ViewContractID
	for _, part := range target.Types() {
		if l.viewIntersectionPhantom(part) {
			continue
		}
		child, err := build(part)
		if err != nil {
			return nil, err
		}
		if child <= 0 || int(child) > len(l.result.ViewContracts) || l.result.ViewContracts[int(child)-1].Kind == ir.ViewUnknown {
			return nil, l.notYet(node, "an unavailable intersection member contract for "+l.checker.TypeToString(part))
		}
		members = append(members, child)
	}
	if len(members) == 0 {
		return nil, l.notYet(node, "an intersection without a runtime constituent")
	}
	return members, nil
}

// Reuse the brand lane's field rule. An empty object is not evidence of a brand,
// and callable/indexed objects must retain their runtime obligations.
func (l *lowering) viewIntersectionPhantom(target *checker.Type) bool {
	if target.Flags()&checker.TypeFlagsObject == 0 || len(l.checker.GetSignaturesOfType(target, checker.SignatureKindCall)) != 0 || len(l.checker.GetSignaturesOfType(target, checker.SignatureKindConstruct)) != 0 || len(l.checker.GetIndexInfosOfType(target)) != 0 {
		return false
	}
	fields := l.checker.GetPropertiesOfType(target)
	if len(fields) == 0 {
		return false
	}
	for _, field := range fields {
		if !phantomField(l.checker.GetTypeOfSymbol(field), field.Flags&ast.SymbolFlagsOptional != 0) {
			return false
		}
	}
	return true
}
