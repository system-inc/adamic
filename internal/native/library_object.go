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
	case "errorToString":
		return e.own(ir.String, fmt.Sprintf("adamic_error_to_string(%s)", arguments[0]))
	case "errorMember":
		return e.own(ir.String, fmt.Sprintf("adamic_error_member(%s, (int)%s)", arguments[0], arguments[1]))
	case "errorSetMember":
		return e.own(ir.String, fmt.Sprintf("adamic_error_set_member(%s, (int)%s, %s)", arguments[0], arguments[1], arguments[2]))
	case "errorCause":
		return e.own(ir.Union, fmt.Sprintf("adamic_error_cause(%s)", arguments[0]))
	case "errorErrors":
		return e.own(ir.Array, fmt.Sprintf("adamic_error_errors(%s)", arguments[0]))
	case "errorPrototype":
		return e.own(ir.Object, fmt.Sprintf("adamic_retain(adamic_error_prototype((int)%s))", arguments[0]))
	case "errorGetPrototype":
		return e.own(ir.Union, fmt.Sprintf("adamic_retain(adamic_error_get_prototype(%s))", arguments[0]))
	case "errorIsPrototypeOf":
		return e.snapshot(ir.Boolean, fmt.Sprintf("adamic_error_is_prototype_of(%s, %s)", arguments[0], arguments[1]))
	case "errorInstanceOf":
		return e.snapshot(ir.Boolean, fmt.Sprintf("adamic_error_instanceof(%s, (int)%s)", arguments[0], arguments[1]))
	case "errorEnumerable":
		return e.snapshot(ir.Boolean, fmt.Sprintf("adamic_error_enumerable(%s, %s)", arguments[0], arguments[1]))
	case "errorCaptureStack":
		if len(arguments) == 3 {
			return fmt.Sprintf("adamic_error_capture_at(%s, %s, %s)", arguments[0], arguments[1], arguments[2])
		}
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
	case "seal", "preventExtensions":
		return e.own(ir.Object, fmt.Sprintf("adamic_object_set_integrity(%s, %t)", arguments[0], call.Method == "seal"))
	case "freeze":
		return e.own(ir.Object, fmt.Sprintf("adamic_object_freeze(%s)", arguments[0]))
	case "hasOwn":
		return e.snapshot(ir.Boolean, fmt.Sprintf("adamic_object_has_own(%s, %s)", arguments[0], arguments[1]))
	case "keys", "getOwnPropertyNames":
		return e.own(ir.Array, fmt.Sprintf("adamic_object_names(%s, %t)", arguments[0], call.Method == "getOwnPropertyNames"))
	case "values", "entries":
		return e.own(ir.Array, fmt.Sprintf("adamic_object_values(%s, %t, %t)", arguments[0], call.Element.IsReference(), call.Method == "entries"))
	case "assign":
		for _, source := range arguments[1:] {
			e.line("adamic_object_assign(%s, %s);", arguments[0], source)
		}
		return e.own(ir.Object, fmt.Sprintf("adamic_retain(%s)", arguments[0]))
	}
	panic("native: unknown Object method")
}
