package native

import "github.com/system-inc/adamic/internal/ir"

// Undefined payload admission is separate from optional field presence.
func (e *emitter) viewStringUndefined(property ir.Property) bool {
	id := property.ViewContract
	return property.Of == ir.String && id > 0 && int(id) <= len(e.program.ViewContracts) && e.program.ViewContracts[id-1].Undefined && e.program.ViewContracts[id-1].Unsupported == ""
}
