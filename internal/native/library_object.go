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
		return e.own(ir.Array, fmt.Sprintf("adamic_object_values(%s, %t, %t)", arguments[0], call.Element.IsReference(), call.Method == "entries"))
	case "assign":
		for _, source := range arguments[1:] {
			e.line("adamic_object_assign(%s, %s);", arguments[0], source)
		}
		return e.own(ir.Object, fmt.Sprintf("adamic_retain(%s)", arguments[0]))
	}
	panic("native: unknown Object method")
}
