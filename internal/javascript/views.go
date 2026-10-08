package javascript

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) viewRead(property ir.Property) string {
	e.requireViewFrame(property.ViewContract)
	expected := property.ViewType
	if expected == "" {
		expected = map[ir.Type]string{ir.Number: "number", ir.Boolean: "boolean", ir.String: "string", ir.Object: "object", ir.Array: "array", ir.Map: "Map"}[property.Of]
	}
	return fmt.Sprintf("adamicViewField(%s, %s, %s, %d, %s, [%s])", e.value(property.Object), quote(property.Name), quote(property.View), property.Of, quote(expected), e.values(property.ViewAllowed))
}

// V1 supplies the shared call boundary without changing the ordinary counted,
// rest or receiver ABI. Unsupported signatures never authorize a call.
func (e *emitter) requireViewFrame(id ir.ViewContractID) {
	if id == 0 {
		return
	}
	contract := e.program.ViewContracts[id-1]
	if contract.Unsupported != "" || (contract.Kind != ir.ViewScalar && contract.Kind != ir.ViewObject) {
		panic("compiler bug: unsupported checked-view family reached JavaScript dispatch")
	}
}
