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
	e.adoptProgram(made, "sizeof *"+made, from.ProgramRegion)
	e.line("for (size_t %s = 0; %s < %s; %s++) {", index, index, count, index)
	// Undefined as the first parameter holds it: a null reference, or a packed word.
	undefined := "{.reference = NULL}"
	if from.First.IsMaybe() {
		undefined = fmt.Sprintf("{.%s = %s}", member(from.First), slotted(from.First, zero(from.First)))
	}
	e.indent++
	element := e.temporary()
	call := e.callbackCall(callback, from.Callback, from.CallbackType, undefined, fmt.Sprintf("{.number = (double)%s}", index))
	e.line("adamic_value %s = %s;", element, call)
	// What's made so far is the statement's, let go with its temporaries.
	e.closureThrown()
	e.line("adamic_array_push(%s, %s);", made, element)
	e.indent--
	e.line("}")
	return made
}
