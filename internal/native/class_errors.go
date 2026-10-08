package native

import (
	"fmt"

	"github.com/system-inc/adamic/internal/ir"
)

// A named error keeps the existing Error ownership and uncaught formatting convention.
func (e *emitter) makeError(expression ir.MakeError) string {
	if expression.Family != "" {
		kind := 1
		for index, name := range []string{"Error", "TypeError", "RangeError", "SyntaxError", "ReferenceError", "EvalError", "URIError", "AggregateError"} {
			if name == expression.Family {
				kind = index + 1
			}
		}
		errors := "NULL"
		if expression.Errors != nil {
			errors = e.value(expression.Errors)
		}
		message := e.value(expression.Message)
		cause := "NULL"
		if expression.Cause != nil {
			cause = e.value(expression.Cause)
		}
		frames, limit := e.value(expression.Frames), e.value(expression.Limit)
		return e.own(ir.Object, fmt.Sprintf("adamic_error_builtin_new(%d, %s, %t, %s, %t, %s, %d, %s, %s)", kind, message, !expression.MessageAbsent, cause, expression.Cause != nil, errors, expression.ErrorElement, frames, limit))
	}
	object := e.own(ir.Object, fmt.Sprintf("adamic_error_new(%s)", e.value(expression.Message)))
	if expression.Name != nil {
		name := e.value(expression.Name)
		slot := e.temporary()
		e.line("adamic_value *%s = adamic_object_field(%s, \"name\", &%s);", slot, object, e.cache())
		e.line("adamic_release(%s->reference);", slot)
		e.line("%s->reference = %s;", slot, e.kept(name))
	}
	return object
}
