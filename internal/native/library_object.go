package native

import (
	"fmt"

	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) objectCall(call ir.ObjectCall) string {
	arguments := make([]string, 0, len(call.Arguments))
	for _, argument := range call.Arguments {
		arguments = append(arguments, e.value(argument))
	}
	switch call.Method {
	case "optionalWritePresence":
		e.line("extern bool adamic_optional_write_presence(const adamic_object *, const adamic_string *, const char *);")
		return e.snapshot(ir.Boolean, fmt.Sprintf("adamic_optional_write_presence(%s, %s, %s)", arguments[0], arguments[1], cString(call.Readiness)))
	case "is":
		return e.snapshot(ir.Boolean, fmt.Sprintf("adamic_object_is(%s, %s)", arguments[0], arguments[1]))
	case "isFrozen":
		return e.snapshot(ir.Boolean, fmt.Sprintf("adamic_object_is_frozen(%s)", arguments[0]))
	case "freeze":
		return e.own(ir.Object, fmt.Sprintf("adamic_object_freeze(%s)", arguments[0]))
	case "optionalDelete":
		return e.snapshot(ir.Boolean, fmt.Sprintf("adamic_object_delete(%s, %s)", arguments[0], arguments[1]))
	case "hasOwn", "optionalIn":
		return e.snapshot(ir.Boolean, fmt.Sprintf("adamic_object_has_own(%s, %s)", arguments[0], arguments[1]))
	case "keys":
		return e.own(ir.Array, fmt.Sprintf("adamic_object_keys(%s)", arguments[0]))
	case "values", "entries":
		if call.Readiness != "" {
			return e.own(ir.Array, fmt.Sprintf("adamic_object_values_checked(%s, %t, %t, %s)", arguments[0], call.Element.IsReference(), call.Method == "entries", cString(call.Readiness)))
		}
		return e.own(ir.Array, fmt.Sprintf("adamic_object_values(%s, %t, %t)", arguments[0], call.Element.IsReference(), call.Method == "entries"))
	case "assign":
		for _, source := range arguments[1:] {
			if call.Readiness != "" {
				e.line("adamic_object_assign_checked(%s, %s, %s);", arguments[0], source, cString(call.Readiness))
			} else {
				e.line("adamic_object_assign(%s, %s);", arguments[0], source)
			}
		}
		return e.own(ir.Object, fmt.Sprintf("adamic_retain(%s)", arguments[0]))
	}
	panic("native: unknown Object method")
}
