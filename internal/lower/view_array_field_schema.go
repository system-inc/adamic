package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Array-write certificates belong to array payload views. An unrelated array
// field sharing a scalar view's name does not acquire that obligation.
func (l *lowering) viewSchemaArrayField(target *checker.Type, name string) bool {
	seen := map[ir.ViewContractID]bool{}
	var containsArray func(ir.ViewContractID) bool
	containsArray = func(id ir.ViewContractID) bool {
		if id == 0 || seen[id] {
			return false
		}
		seen[id] = true
		contract := l.result.ViewContracts[id-1]
		if contract.Kind == ir.ViewArray {
			return true
		}
		for _, member := range contract.Members {
			if containsArray(member) {
				return true
			}
		}
		return false
	}
	visited := map[ir.ViewContractID]bool{}
	var find func(ir.ViewContractID) bool
	find = func(id ir.ViewContractID) bool {
		if id == 0 || visited[id] {
			return false
		}
		visited[id] = true
		contract := l.result.ViewContracts[id-1]
		for _, field := range contract.Fields {
			if field.Name == name && containsArray(field.Contract) {
				return true
			}
			if find(field.Contract) {
				return true
			}
		}
		for _, member := range contract.Members {
			if find(member) {
				return true
			}
		}
		return false
	}
	return find(l.result.ViewContractTypes[int(target.Id())])
}
