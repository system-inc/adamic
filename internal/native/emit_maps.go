// Functions and declarations moved unchanged from emit.go.
package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

// newMap makes a map, or a Set's map, for its keys: strings by their text, numbers by
// SameValueZero, booleans as booleans, and anything else held by reference by its identity.
func newMap(key ir.Type, referenceValues bool) string {
	if key == ir.MaybeNumber {
		return fmt.Sprintf("adamic_map_new_maybe_numbers(%t)", referenceValues)
	}
	if key == ir.Boolean {
		return fmt.Sprintf("adamic_map_new_booleans(%t)", referenceValues)
	}
	if key.IsReference() && key != ir.String {
		return fmt.Sprintf("adamic_map_new_identity(%t)", referenceValues)
	}
	return fmt.Sprintf("adamic_map_new(%t, %t)", key == ir.String, referenceValues)
}

// mapForEach emits map.forEach and set.forEach as for...of's loop over the map, live as it is. The key
// is borrowed: closure parameters retain on entry, before user code can delete or clear it.
// A named callback goes through an owned closure forwarder, which holds the key while the named
// function borrows it; a reassigned parameter takes an additional count in that function.
// After the callback returns (or throws), the loop never reads that key again. The iterator owns
// the map and its iterating count prevents tombstone compaction; next reads the current entries
// array by index, skipping deleted entries even after growth. A Set passes the key twice, and
// each callback parameter owns its own count. Keep the value hold, whose release reads it after
// the call. The Map and Set foreach_keys.a and foreach_named_keys.a fixtures hold this with
// counted keys. Detached methods through interfaces are refused as unbound-method.
func (e *emitter) mapForEach(visit ir.MapForEach) string {
	collection := e.value(visit.Map)
	callback := e.value(visit.Callback)
	iterator, key, value, answer := e.temporary(), e.temporary(), e.temporary(), e.temporary()
	e.line("adamic_map_iterator *%s = adamic_map_iterate(%s);", iterator, collection)
	e.line("adamic_value %s, %s;", key, value)
	e.line("while (adamic_map_iterator_next(%s, &%s, &%s)) {", iterator, key, value)
	e.indent++
	hold := func(slot string, of ir.Type, how string) {
		if of.IsReference() {
			e.line("%s(%s.reference);", how, slot)
		}
	}
	first := value
	if visit.Set {
		first = key
	} else {
		hold(value, visit.Value, "adamic_retain")
	}
	call := fmt.Sprintf("%s->code(%s, 3, (adamic_value[]){%s, %s, {.reference = %s}})", callback, callback, first, key, collection)
	// A throw lets go of the value held across the call, and the iterator.
	holds := []string{}
	if !visit.Set && visit.Value.IsReference() {
		holds = append(holds, value+".reference")
	}
	holds = append(holds, iterator)
	if visit.Returns.IsReference() {
		// A callback's result comes back owned, and forEach has no use for it.
		e.line("adamic_value %s = %s;", answer, call)
		e.closureThrown(holds...)
		e.line("adamic_release(%s.reference);", answer)
	} else {
		e.line("%s;", call)
		e.closureThrown(holds...)
	}
	if !visit.Set {
		hold(value, visit.Value, "adamic_release")
	}
	e.indent--
	e.line("}")
	e.line("adamic_release(%s);", iterator)
	return "0"
}
