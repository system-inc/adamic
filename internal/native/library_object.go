package native

import (
	"fmt"
	"strings"

	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) objectCall(call ir.ObjectCall) string {
	arguments := make([]string, 0, len(call.Arguments))
	for _, argument := range call.Arguments {
		arguments = append(arguments, e.value(argument))
	}
	switch call.Method {
	case "catchProperty":
		return e.caughtProperty(call, arguments)
	case "optionalFunctionStorage":
		e.line("extern bool adamic_optional_function_storage(const adamic_closure *, const char *);")
		return e.snapshot(ir.Boolean, fmt.Sprintf("adamic_optional_function_storage(%s, %s)", arguments[0], cString(call.Readiness)))
	case "optionalArrayPresence":
		e.line("extern bool adamic_optional_array_presence(const adamic_array *, bool, const adamic_array *, bool, const char *);")
		return e.snapshot(ir.Boolean, fmt.Sprintf("adamic_optional_array_presence(%s, %s, %s, %s, %s)", arguments[0], arguments[1], arguments[2], arguments[3], cString(call.Readiness)))
	case "optionalViewStorage":
		e.line("extern bool adamic_optional_view_storage(const adamic_object *, bool, const char *);")
		return e.snapshot(ir.Boolean, fmt.Sprintf("adamic_optional_view_storage(%s, %s, %s)", arguments[0], arguments[1], cString(call.Readiness)))
	case "optionalSpreadKeys":
		e.line("extern adamic_array *adamic_optional_spread_keys(const adamic_object *);")
		return e.own(ir.Array, fmt.Sprintf("adamic_optional_spread_keys(%s)", arguments[0]))
	case "optionalSpreadPresence":
		e.line("extern bool adamic_optional_spread_presence(const adamic_object *, const adamic_array *, const char *);")
		return e.snapshot(ir.Boolean, fmt.Sprintf("adamic_optional_spread_presence(%s, %s, %s)", arguments[0], arguments[1], cString(call.Readiness)))
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
		if call.Checked {
			e.declarations = append(e.declarations, "adamic_array *adamic_object_values_checked(adamic_object *, int, const char *, bool, const adamic_value *, size_t);")
			allowed := "NULL"
			if len(call.Allowed) > 0 {
				allowed = e.temporary()
				values := make([]string, 0, len(call.Allowed))
				for _, value := range call.Allowed {
					values = append(values, fmt.Sprintf("{.%s = %s}", member(call.Element), e.value(value)))
				}
				e.line("const adamic_value %s[] = {%s};", allowed, strings.Join(values, ", "))
			}
			return e.own(ir.Array, fmt.Sprintf("adamic_object_values_checked(%s, %d, %s, %t, %s, %d)", arguments[0], call.Element, cString(call.ElementName), call.Method == "entries", allowed, len(call.Allowed)))
		}
		return e.own(ir.Array, fmt.Sprintf("adamic_object_values(%s, %t, %t)", arguments[0], call.Element.IsReference(), call.Method == "entries"))
	case "assign":
		for _, source := range arguments[1:] {
			e.line("adamic_object_assign(%s, %s);", arguments[0], source)
		}
		return e.own(ir.Object, fmt.Sprintf("adamic_retain(%s)", arguments[0]))
	}
	panic("native: unknown Object method")
}
