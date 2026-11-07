// Functions and declarations moved unchanged from emit.go.
package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"regexp"
	"slices"
)

// releaseScopes releases the string locals of every scope from depth inward.
func (nativeEmitter *emitter) releaseScopes(depth int) {
	for scope := len(nativeEmitter.scopes) - 1; scope >= depth; scope-- {
		for index := len(nativeEmitter.scopes[scope]) - 1; index >= 0; index-- {
			if nativeEmitter.closes[nativeEmitter.scopes[scope][index]] {
				// A view's for-of left: a generator it holds is finished, as Node's return() leaves
				// it, so a later pass over it finds it done and its cells are let go of now.
				nativeEmitter.line("adamic_iterator_close(%s);", nativeEmitter.scopes[scope][index])
			}
			if nativeEmitter.regions != nil {
				if cleanup := nativeEmitter.regions.cleanups[nativeEmitter.scopes[scope][index]]; cleanup != "" {
					nativeEmitter.line("%s", cleanup)
					continue
				}
			}
			if nativeEmitter.mostlyNull[nativeEmitter.scopes[scope][index]] {
				// Held only on the rare path (a fallback's owner): NULL, the common case, needs no call.
				nativeEmitter.line("if (%s != NULL) {", nativeEmitter.scopes[scope][index])
				nativeEmitter.line("\tadamic_release(%s);", nativeEmitter.scopes[scope][index])
				nativeEmitter.line("}")
				continue
			}
			nativeEmitter.line("adamic_release(%s);", nativeEmitter.scopes[scope][index])
		}
	}
}

// hold makes the innermost scope release a reference when it ends.
func (nativeEmitter *emitter) hold(name string) {
	nativeEmitter.scopes[len(nativeEmitter.scopes)-1] = append(nativeEmitter.scopes[len(nativeEmitter.scopes)-1], name)
}

// end releases what the statement just emitted owned, and settles its checks' ladder after.
func (nativeEmitter *emitter) end() {
	for index := len(nativeEmitter.owned) - 1; index >= 0; index-- {
		nativeEmitter.line("adamic_release(%s);", nativeEmitter.owned[index])
	}
	nativeEmitter.owned = nil
	nativeEmitter.settleThrows()
}

// releaseGlobals lets go of every global main declared, the last declared first, once main has run
// to its end. Nothing reads them after that, and a program that let go of everything is one the leak
// check can hold to account: a reference still held at exit, or still written in a global, hides
// whatever it reaches. A panic exits
// where it stands, as Node does, and never gets here.
func (nativeEmitter *emitter) releaseGlobals() {
	for index := len(nativeEmitter.initialized) - 1; index >= 0; index-- {
		if nativeEmitter.program.Locals[nativeEmitter.initialized[index]].Type.IsReference() {
			// And forgotten: the leak check counts what a global points at as alive, so a count the
			// global's own object never gave back, or a cycle the global led into, would hide.
			nativeEmitter.line("adamic_release(%s);", nativeEmitter.localName(nativeEmitter.initialized[index]))
			nativeEmitter.line("%s = NULL;", nativeEmitter.localName(nativeEmitter.initialized[index]))
		}
	}
}

// snapshot reads a value that something later in the statement could change (a field, a length, a
// map's size) into a temporary, at the moment JavaScript reads it.
func (nativeEmitter *emitter) snapshot(valueType ir.Type, value string) string {
	name := nativeEmitter.temporary()
	nativeEmitter.line("%s %s = %s;", cType(valueType), name, value)
	return name
}

// retained is a reference retained, for a place that keeps it: the constant undefined (the null
// pointer, however it's cast) needs no retain, since retain passes over it, and costs a call.
func retained(value string) string {
	if constantUndefined.MatchString(value) {
		return value
	}
	return "adamic_retain(" + value + ")"
}

// constantUndefined matches C that is the null pointer constant, as Undefined and a missing argument
// are emitted: NULL, perhaps parenthesized and cast to a pointer type.
var constantUndefined = regexp.MustCompile(`^\(*(\([a-z_]+ \*\)\(*)?NULL\)*$`)

// kept is a reference handed to a place that holds it past the statement's end and can't let go of
// it before: a field of an object the statement made, or what a function returns. A temporary the
// statement owns is moved there, its count with it, and isn't let go of when the statement ends;
// anything else is retained. A variable isn't such a place: a call later in the statement could
// assign it, and let go of the value while the statement still reads it.
func (nativeEmitter *emitter) kept(value string) string {
	if index := slices.Index(nativeEmitter.owned, value); index >= 0 {
		nativeEmitter.owned = slices.Delete(nativeEmitter.owned, index, index+1)
		return value
	}
	return retained(value)
}

// taken gives up the statement's own count of a temporary, when value is one, to whatever is about
// to hold it without a retain: true says it was given up.
func (nativeEmitter *emitter) taken(value string) bool {
	index := slices.Index(nativeEmitter.owned, value)
	if index < 0 {
		return false
	}
	nativeEmitter.owned = slices.Delete(nativeEmitter.owned, index, index+1)
	return true
}

// own puts a reference the statement owns in a temporary, released when the statement ends.
func (nativeEmitter *emitter) own(valueType ir.Type, value string) string {
	name := nativeEmitter.temporary()
	nativeEmitter.line("%s %s = %s;", cType(valueType), name, value)
	nativeEmitter.owned = append(nativeEmitter.owned, name)
	return name
}

// held is a value as an adamic_value whose reference, if it has one, the receiver keeps.
func held(valueType ir.Type, value string) string {
	if valueType.IsReference() {
		return fmt.Sprintf("(adamic_value){.reference = %s}", retained(value))
	}
	return fmt.Sprintf("(adamic_value){.%s = %s}", member(valueType), slotted(valueType, value))
}

// borrowed is a value as an adamic_value the receiver only looks at.
func borrowed(valueType ir.Type, value string) string {
	return fmt.Sprintf("(adamic_value){.%s = %s}", member(valueType), slotted(valueType, value))
}
