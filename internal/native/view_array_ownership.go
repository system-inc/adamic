package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

// Array payload owners preserve retained union boxes across callback reads.
func (e *emitter) viewArrayReadOwner(element ir.Type) string {
	if element != ir.Union {
		return "NULL"
	}
	return e.own(ir.Array, "adamic_array_new(0, true)")
}

// V3 keeps the base's acyclic reference-counted allocation policy. The graph
// runtime slice installs graph-region adoption separately.
func (e *emitter) graphArray(code string, types []int) string {
	return e.own(ir.Array, code)
}
func (e *emitter) heldIn(holder string, of ir.Type, value string) string {
	if of.IsReference() {
		return fmt.Sprintf("(adamic_value){.reference = %s}", retained(value))
	}
	return borrowed(of, value)
}
