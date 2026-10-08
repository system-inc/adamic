package javascript

import (
	"encoding/json"

	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) viewCallableLogicalProducer(property ir.Property, recorded string) string {
	if property.ViewContract == 0 {
		return recorded
	}
	contract := e.program.ViewContracts[property.ViewContract-1]
	if contract.Kind != ir.ViewCallable || !contract.ProducerCertified {
		return recorded
	}
	functions := append([]int{}, contract.Functions...)
	encoded, _ := json.Marshal(functions)
	return "((recorded) => " + string(encoded) + ".includes(recorded?.function) ? recorded : undefined)(" + recorded + ")"
}
