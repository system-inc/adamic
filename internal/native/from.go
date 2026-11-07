package native

import (
	"fmt"

	"github.com/system-inc/adamic/internal/ir"
)

// arrayFrom emits Array.from({ length }, callback): the length and the callback evaluated in that
// order, the length checked as JavaScript checks it, then the callback called once per index with
// undefined and the index, each result pushed (it comes back owned, so the array takes it as it is).
func (nativeEmitter *emitter) arrayFrom(from ir.ArrayFrom) string {
	length := nativeEmitter.value(from.Length)
	callback := nativeEmitter.value(from.Callback)
	count, index := nativeEmitter.temporary(), nativeEmitter.temporary()
	nativeEmitter.line("size_t %s = adamic_array_from_length(%s);", count, length)
	// A length past 2^32 - 1 leaves a RangeError pending, tested before the array is made.
	if from.MayThrow() {
		nativeEmitter.checkThrown()
	}
	made := nativeEmitter.own(ir.Array, fmt.Sprintf("adamic_array_make(%s, %s)", count, arrayElement(from.Element)))
	nativeEmitter.adoptGraph(made, "sizeof *"+made, nativeEmitter.graphTypes(from.GraphTypes) && !from.ProgramRegion)
	nativeEmitter.programArray(made, from.ProgramRegion)
	nativeEmitter.line("for (size_t %s = 0; %s < %s; %s++) {", index, index, count, index)
	// Undefined as the first parameter holds it: a null reference, or a packed word.
	undefined := "{.reference = NULL}"
	if from.First.IsMaybe() {
		undefined = fmt.Sprintf("{.%s = %s}", member(from.First), slotted(from.First, zero(from.First)))
	}
	nativeEmitter.indent++
	element := nativeEmitter.temporary()
	call := nativeEmitter.callbackCall(
		callback,
		from.Callback,
		from.CallbackType,
		undefined,
		fmt.Sprintf("{.number = (double)%s}", index),
	)
	nativeEmitter.line("adamic_value %s = %s;", element, call)
	// What's made so far is the statement's, let go with its temporaries.
	nativeEmitter.closureThrown()
	if nativeEmitter.graphTypes(from.GraphTypes) && from.Element.IsReference() {
		nativeEmitter.line("%s.reference = adamic_graph_take(%s, %s.reference);", element, made, element)
	}
	nativeEmitter.line("adamic_array_push(%s, %s);", made, element)
	nativeEmitter.indent--
	nativeEmitter.line("}")
	return made
}
