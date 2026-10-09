package native

import (
	"fmt"

	"github.com/system-inc/adamic/internal/ir"
)

// A named error keeps the existing Error ownership and uncaught formatting convention.
func (e *emitter) makeError(expression ir.MakeError) string {
	constructor := expression.Constructor
	if constructor == "" {
		constructor = "Error"
	}
	object := e.own(ir.Object, fmt.Sprintf("adamic_error_new_kind(%s, %s)", e.value(expression.Message), cString(constructor)))
	if expression.Name != nil {
		name := e.value(expression.Name)
		slot := e.temporary()
		e.line("adamic_value *%s = adamic_object_field(%s, \"name\", &%s);", slot, object, e.cache())
		e.line("adamic_release(%s->reference);", slot)
		e.line("%s->reference = %s;", slot, e.kept(name))
	}
	return object
}
