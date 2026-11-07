package native

import "github.com/system-inc/adamic/internal/ir"

// Undefined payload admission is separate from optional field presence.
func (nativeEmitter *emitter) viewStringUndefined(property ir.Property) bool {
	id := property.ViewContract
	return property.Of == ir.String &&
		id > 0 &&
		int(id) <= len(nativeEmitter.program.ViewContracts) &&
		nativeEmitter.program.ViewContracts[id-1].Undefined &&
		nativeEmitter.program.ViewContracts[id-1].Unsupported == ""
}
