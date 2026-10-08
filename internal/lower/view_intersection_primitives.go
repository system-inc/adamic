package lower

import "github.com/system-inc/adamic/internal/ir"

// A complete primitive descendant has its own selector at the actual field
// read. It cannot make an otherwise structural receiver an eager refusal.
// Unknown, object and reference alternatives retain their existing guards.
func (l *lowering) primitiveIntersectionDescendant(id ir.ViewContractID) bool {
	contract := l.result.ViewContracts[id-1]
	if contract.Unsupported != "" {
		return false
	}
	if contract.Kind == ir.ViewNullable {
		id = contract.Element
	}
	members, ok := ir.PrimitiveViewMembers(l.result, id)
	return ok && len(members) != 0
}
