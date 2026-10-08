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
	made := e.own(ir.Array, fmt.Sprintf("adamic_array_new_typed(%s, %t, %s)", count, from.Element.IsReference(), jsonStorage(from.Element)))
	e.adoptGraph(made, "sizeof *"+made, e.graphTypes(from.GraphTypes))
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
	if e.graphTypes(from.GraphTypes) && from.Element.IsReference() {
		e.line("%s.reference = adamic_graph_take(%s, %s.reference);", element, made, element)
	}
	e.line("adamic_array_push_typed(%s, %s, %s);", made, element, jsonStorage(from.Element))
	e.indent--
	e.line("}")
	return made
}
