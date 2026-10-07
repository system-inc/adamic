// Functions and declarations moved unchanged from emit.go.
package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"regexp"
	"slices"
)

// releaseScopes releases the string locals of every scope from depth inward.
func (e *emitter) releaseScopes(depth int) {
	for scope := len(e.scopes) - 1; scope >= depth; scope-- {
		for index := len(e.scopes[scope]) - 1; index >= 0; index-- {
			if e.mostlyNull[e.scopes[scope][index]] {
				// Held only on the rare path (a fallback's owner): NULL, the common case, needs no call.
				e.line("if (%s != NULL) {", e.scopes[scope][index])
				e.line("\tadamic_release(%s);", e.scopes[scope][index])
				e.line("}")
				continue
			}
			e.line("adamic_release(%s);", e.scopes[scope][index])
		}
	}
}

// hold makes the innermost scope release a reference when it ends.
func (e *emitter) hold(name string) {
	e.scopes[len(e.scopes)-1] = append(e.scopes[len(e.scopes)-1], name)
}

// end releases what the statement just emitted owned.
func (e *emitter) end() {
	for index := len(e.owned) - 1; index >= 0; index-- {
		e.line("adamic_release(%s);", e.owned[index])
	}
	e.owned = nil
}

// releaseGlobals lets go of every global main declared, the last declared first, once main has run
// to its end. Nothing reads them after that, and a program that let go of everything is one the leak
// check can hold to account: a reference still held at exit, or still written in a global, hides
// whatever it reaches. A panic exits
// where it stands, as Node does, and never gets here.
func (e *emitter) releaseGlobals() {
	for index := len(e.initialized) - 1; index >= 0; index-- {
		if e.program.Locals[e.initialized[index]].Type.IsReference() {
			// And forgotten: the leak check counts what a global points at as alive, so a count the
			// global's own object never gave back, or a cycle the global led into, would hide.
			e.line("adamic_release(%s);", e.localName(e.initialized[index]))
			e.line("%s = NULL;", e.localName(e.initialized[index]))
		}
	}
}

// snapshot reads a value that something later in the statement could change (a field, a length, a
// map's size) into a temporary, at the moment JavaScript reads it.
func (e *emitter) snapshot(valueType ir.Type, value string) string {
	name := e.temporary()
	e.line("%s %s = %s;", cType(valueType), name, value)
	return name
}

// retained keeps a reference. Immortal constants and the null pointer need no count.
func retained(value string) string {
	if staticallyImmortal(value) || constantUndefined.MatchString(value) {
		return value
	}
	return "adamic_retain(" + value + ")"
}

// staticallyImmortal recognizes only addresses of statics with immortal headers.
// It runs before ownership dispatch; arbitrary expressions and heap temporaries
// must still take their existing retain or graph boundary path.
func staticallyImmortal(value string) bool {
	return immortalStaticAddress.MatchString(value)
}

var immortalStaticAddress = regexp.MustCompile(`^\(*(\([a-z_]+ \*\)\(*)?&adamic_(string_([0-9]+|empty|true|false)|box_(true|false)(\.heap)?|typeof_(number|string|boolean|undefined|object|function))\)*$`)

// constantUndefined matches C that is the null pointer constant, as Undefined and a missing argument
// are emitted: NULL, perhaps parenthesized and cast to a pointer type.
var constantUndefined = regexp.MustCompile(`^\(*(\([a-z_]+ \*\)\(*)?NULL\)*$`)

// kept is a reference handed to a place that holds it past the statement's end and can't let go of
// it before: a field of an object the statement made, or what a function returns. A temporary the
// statement owns is moved there, its count with it, and isn't let go of when the statement ends;
// anything else is retained. A variable isn't such a place: a call later in the statement could
// assign it, and let go of the value while the statement still reads it.
func (e *emitter) kept(value string) string {
	if index := slices.Index(e.owned, value); index >= 0 {
		e.owned = slices.Delete(e.owned, index, index+1)
		return value
	}
	return retained(value)
}

// taken gives up the statement's own count of a temporary, when value is one, to whatever is about
// to hold it without a retain: true says it was given up.
func (e *emitter) taken(value string) bool {
	index := slices.Index(e.owned, value)
	if index < 0 {
		return false
	}
	e.owned = slices.Delete(e.owned, index, index+1)
	return true
}

// own puts a reference the statement owns in a temporary, released when the statement ends.
func (e *emitter) own(valueType ir.Type, value string) string {
	name := e.temporary()
	e.line("%s %s = %s;", cType(valueType), name, value)
	e.owned = append(e.owned, name)
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
