package native

import (
	"fmt"

	"github.com/system-inc/adamic/internal/ir"
)

// arrayFrom emits Array.from({ length }, callback): the length and the callback evaluated in that
// order, the length checked as JavaScript checks it, then the callback called once per index with
// undefined and the index, each result pushed (it comes back owned, so the array takes it as it is).
func (e *emitter) arrayFrom(from ir.ArrayFrom) string {
	length := e.value(from.Length)
	callback := e.value(from.Callback)
	count, index := e.temporary(), e.temporary()
	e.line("size_t %s = adamic_array_from_length(%s);", count, length)
	made := e.own(ir.Array, fmt.Sprintf("adamic_array_new(%s, %t)", count, from.Element.IsReference()))
	e.line("for (size_t %s = 0; %s < %s; %s++) {", index, index, count, index)
	// Undefined as the first parameter holds it: a null reference, or a packed word.
	undefined := "{.reference = NULL}"
	if from.First == ir.MaybeNumber {
		undefined = fmt.Sprintf("{.number = %s}", slotted(ir.MaybeNumber, zero(ir.MaybeNumber)))
	}
	e.indent++
	element := e.temporary()
	e.line("adamic_value %s = %s->code(%s, (adamic_value[]){%s, {.number = (double)%s}});", element, callback, callback, undefined, index)
	// What's made so far is the statement's, let go with its temporaries.
	e.closureThrown()
	e.line("adamic_array_push(%s, %s);", made, element)
	e.indent--
	e.line("}")
	return made
}
