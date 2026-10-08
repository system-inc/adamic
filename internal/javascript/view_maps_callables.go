package javascript

import "github.com/system-inc/adamic/internal/ir"

func (e *emitter) mapEntryCallableCertificate(id ir.ViewContractID, value, where string) string {
	if id == 0 {
		return value
	}
	contract := e.program.ViewContracts[id-1]
	if contract.Kind == ir.ViewNullable {
		if contract.Element == 0 || e.program.ViewContracts[contract.Element-1].Kind != ir.ViewCallable {
			return value
		}
		return "((entry) => entry === null || entry === undefined ? entry : " + e.mapEntryCallableCertificate(contract.Element, "entry", where) + ")(" + value + ")"
	}
	if contract.Kind != ir.ViewCallable {
		return value
	}
	return e.emitViewCallableCertificate(ir.Property{Of: ir.Closure, ViewContract: id, View: where}, value)
}
