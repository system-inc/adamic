package native

import (
	"fmt"

	"github.com/system-inc/adamic/internal/ir"
)

// maybe is a Maybe pair (number | undefined, boolean | undefined) as a C expression: present, holding
// value, or missing when value is "".
func maybe(of ir.Type, value string) string {
	if value == "" {
		return zero(of)
	}
	return fmt.Sprintf("((%s){true, %s})", cType(of), value)
}

// maybeSlot is the Maybe pair for an element's slot that may be NULL (missing), holding a value of
// the element's type: a number or a boolean, or number | undefined, packed.
func maybeSlot(element ir.Type, slot string) string {
	of := ir.Maybe(element)
	held := slot + "->" + member(element)
	if element.IsMaybe() {
		held = unslotted(element, held)
	} else {
		held = maybe(of, held)
	}
	return fmt.Sprintf("%s == NULL ? %s : %s", slot, zero(of), held)
}

// maybeToString is String(value) for a Maybe pair: what it holds, written as String() writes it, or
// "undefined". A number's text is made, and owned; the rest are constants.
func (e *emitter) maybeToString(value ir.Expression) string {
	pair := e.value(value)
	if value.Type() == ir.MaybeBoolean {
		return fmt.Sprintf("(!(%s).present ? &adamic_string_undefined : (%s).boolean ? &adamic_string_true : &adamic_string_false)", pair, pair)
	}
	return e.own(ir.String, fmt.Sprintf("adamic_string_from_maybe_number(%s)", pair))
}

// pairEquality is === or !== on two Maybe pairs of one type.
func pairEquality(operator ir.Operator, pair ir.Type, left string, right string) string {
	function := "adamic_maybe_number_equal"
	if pair == ir.MaybeBoolean {
		function = "adamic_maybe_boolean_equal"
	}
	if operator == ir.NotEqual {
		return fmt.Sprintf("(!%s(%s, %s))", function, left, right)
	}
	return fmt.Sprintf("%s(%s, %s)", function, left, right)
}
