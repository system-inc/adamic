package lower

import "github.com/system-inc/adamic/internal/ir"

// Flat conjunctions retain every merged field contract. Their scalar/undefined
// alternatives use the existing plain matcher rather than a shared tag guess.
func (l *lowering) completeMixedArrayContract(id ir.ViewContractID) bool {
	if !ir.MixedArrayContract(l.result, id) {
		return false
	}
	root := &l.result.ViewContracts[id-1]
	root.Unsupported = ""
	for _, member := range root.Members {
		c := &l.result.ViewContracts[member-1]
		if c.Kind == ir.ViewObject && c.Unsupported == "compound intersection payload" {
			c.Unsupported = ""
		}
	}
	return true
}
