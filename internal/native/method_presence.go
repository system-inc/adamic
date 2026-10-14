package native

import (
	"github.com/system-inc/adamic/internal/ir"
	"slices"
)

// methodPresence checks own storage before prototype methods. It never constructs,
// retains or returns a closure, and evaluates the receiver exactly once.
func (e *emitter) methodPresence(expression ir.MethodPresence) string {
	if !slices.Contains(e.declarations, "#include <string.h>") {
		e.declarations = append(e.declarations, "#include <string.h>")
	}
	object := e.snapshot(ir.Object, e.value(expression.Object))
	present, found, index := e.temporary(), e.temporary(), e.temporary()
	e.line("bool %s = false;", present)
	e.line("bool %s = false;", found)
	if expression.Optional {
		e.line("if (%s != NULL) {", object)
		e.indent++
	}
	e.line("for (size_t %s = 0; %s < %s->shape->count; %s++) {", index, index, object, index)
	e.line("\tif (strcmp(%s->shape->names[%s], %s) == 0) {", object, index, cString(expression.Name))
	e.line("\t\t%s = true;", found)
	e.line("\t\t%s = %s->slots[%s].reference != NULL;", present, object, index)
	e.line("\t\tbreak;")
	e.line("\t}")
	e.line("}")
	e.line("for (size_t %s = 0; !%s && %s->shape->methods != NULL && %s < %s->shape->methods->count; %s++) {", index, found, object, index, object, index)
	e.line("\tif (strcmp(%s->shape->methods->names[%s], %s) == 0) {", object, index, cString(expression.Name))
	e.line("\t\t%s = true;", present)
	e.line("\t\tbreak;")
	e.line("\t}")
	e.line("}")
	if expression.Optional {
		e.indent--
		e.line("}")
	}
	return present
}
