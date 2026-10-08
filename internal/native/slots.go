package native

import (
	"fmt"

	"github.com/system-inc/adamic/internal/ir"
)

// slotted is a value as an adamic_value holds it, in a field, an element, a cell, or a function
// value's argument or result. number | undefined is packed into the one double (maybe.c): undefined
// as the reserved NaN, any other NaN made the ordinary one. Optional booleans use one byte
// with separate tags for missing, false and true. Everything else is held as it is.
func slotted(valueType ir.Type, value string) string {
	if valueType == ir.MaybeNumber {
		return fmt.Sprintf("adamic_maybe_number_pack(%s)", value)
	}
	if valueType == ir.MaybeBoolean {
		return fmt.Sprintf("adamic_maybe_boolean_pack(%s)", value)
	}
	return value
}

// unslotted is what a slot of the type holds, given the slot's member (slot.number), read back.
func unslotted(valueType ir.Type, member string) string {
	if valueType == ir.MaybeNumber {
		return fmt.Sprintf("adamic_maybe_number_unpack(%s)", member)
	}
	if valueType == ir.MaybeBoolean {
		return fmt.Sprintf("adamic_maybe_boolean_unpack(%s)", member)
	}
	return member
}
