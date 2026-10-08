package native

import "github.com/system-inc/adamic/internal/ir"

// Common read and call dispatch deliberately leaves family admission to later
// slices. A descriptor with no proof must never select an unchecked operation.
func (e *emitter) viewRead(property ir.Property) string {
	e.requireViewFrame(property.ViewContract)
	return e.viewField(property)
}

func (e *emitter) requireViewFrame(id ir.ViewContractID) {
	if id == 0 {
		return
	} // the base's checked scalar path has no descriptor
	contract := e.program.ViewContracts[id-1]
	if contract.Unsupported != "" || (contract.Kind != ir.ViewScalar && contract.Kind != ir.ViewObject) {
		panic("compiler bug: unsupported checked-view family reached native dispatch")
	}
}
