package native

import "github.com/system-inc/adamic/internal/ir"

func (e *emitter) viewCallableNestedSignature(parameters []ir.Type, result ir.Type, name string, masks []uint16, contracts []ir.ViewContractID) string {
	children := []string{}
	active := false
	for _, id := range contracts {
		child := "NULL"
		if id != 0 && e.program.ViewContracts[id-1].Kind == ir.ViewCallable {
			child = e.viewCallableExpected(ir.Property{ViewContract: id})
			active = true
		}
		children = append(children, child)
	}
	if !active {
		children = nil
	}
	return e.viewCallableSignatureChildren(parameters, result, name, masks, children)
}
