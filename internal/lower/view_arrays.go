package lower

import "github.com/system-inc/adamic/internal/ir"

// Mutability belongs to the array type, independently of whether its containing
// property is readonly. Optional and union fields retain every array arm's rule.
func mutableViewArrayContract(program *ir.Program, id ir.ViewContractID, seen map[ir.ViewContractID]bool) bool {
	if id <= 0 || int(id) > len(program.ViewContracts) || seen[id] {
		return false
	}
	seen[id] = true
	contract := program.ViewContracts[id-1]
	if contract.Kind == ir.ViewArray {
		return !contract.ArrayReadonly
	}
	for _, member := range contract.Members {
		if mutableViewArrayContract(program, member, seen) {
			return true
		}
	}
	return false
}
