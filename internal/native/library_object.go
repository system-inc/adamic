package native

import (
	"fmt"
	"strings"

	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) objectCall(call ir.ObjectCall) string {
	if value, handled := e.objectDescriptorCall(call); handled {
		return value
	}
	arguments := make([]string, 0, len(call.Arguments))
	for _, argument := range call.Arguments {
		arguments = append(arguments, e.value(argument))
	}
	switch call.Method {
	case "intrinsicObjectPrototype":
		e.declarations = append(e.declarations, "#include \"library_prototypes.h\"")
		return e.own(ir.Object, "adamic_object_intrinsic_prototype()")
	case "create", "setPrototypeOf", "getPrototypeOf":
		e.declarations = append(e.declarations, "#include \"library_prototypes.h\"")
		function := map[string]string{"create": "create", "setPrototypeOf": "set", "getPrototypeOf": "get"}[call.Method]
		value := e.own(ir.Object, fmt.Sprintf("adamic_object_%s_prototype(%s)", function, strings.Join(arguments, ", ")))
		if call.Method == "setPrototypeOf" {
			e.checkThrown()
		}
		return value
	case "errorIsType":
		e.declarations = append(e.declarations, "#include \"library_errors.h\"")
		return e.snapshot(ir.Boolean, fmt.Sprintf("adamic_library_error_is_type(%s, %s)", arguments[0], arguments[1]))
	case "errorCaptureStack":
		return fmt.Sprintf("adamic_error_capture_stack(%s)", arguments[0])
	case "errorReadStack":
		return e.own(ir.String, fmt.Sprintf("adamic_error_read_stack(%s)", arguments[0]))
	case "is":
		return e.snapshot(ir.Boolean, fmt.Sprintf("adamic_object_is(%s, %s)", arguments[0], arguments[1]))
	case "isFrozen":
		return e.snapshot(ir.Boolean, fmt.Sprintf("adamic_object_test_integrity(%s, true)", arguments[0]))
	case "isSealed":
		return e.snapshot(ir.Boolean, fmt.Sprintf("adamic_object_test_integrity(%s, false)", arguments[0]))
	case "isExtensible":
		return e.snapshot(ir.Boolean, fmt.Sprintf("adamic_object_is_extensible(%s)", arguments[0]))
	case "seal", "preventExtensions", "freeze":
		return e.own(call.Returns, fmt.Sprintf("adamic_receiver_set_integrity((adamic_heap *)%s, %t, %t, %d)", arguments[0], call.Method != "preventExtensions", call.Method == "freeze", call.IntegrityShape))
	case "hasOwn":
		return e.snapshot(ir.Boolean, fmt.Sprintf("adamic_object_has_own(%s, %s)", arguments[0], arguments[1]))
	case "keys", "getOwnPropertyNames":
		return e.own(ir.Array, fmt.Sprintf("adamic_object_names(%s, %t)", arguments[0], call.Method == "getOwnPropertyNames"))
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
			e.checkThrown()
		}
		return e.own(ir.Object, fmt.Sprintf("adamic_retain(%s)", arguments[0]))
	}
	panic("native: unknown Object method")
}
