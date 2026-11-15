package native

import (
	"fmt"

	"github.com/system-inc/adamic/internal/ir"
)

// Presence asks whether a callable is there without materializing a prototype function.
// Own fields keep the ordinary readiness check, and the receiver is evaluated once.
func (e *emitter) optionalFunctionPresence(value ir.Expression) (string, bool) {
	if !e.program.RegExpServedView {
		return "", false
	}
	property, ok := value.(ir.Property)
	if !ok || property.Of != ir.Closure || !property.Absent || property.View != "" {
		return "", false
	}
	object := e.value(property.Object)
	readiness := "NULL"
	if property.Readiness != "" {
		readiness = cString(property.Readiness)
	}
	return e.snapshot(ir.Boolean, fmt.Sprintf("adamic_object_optional_function_present(%s, %s, &%s, %s)", object, cString(property.Name), e.cache(), readiness)), true
}
