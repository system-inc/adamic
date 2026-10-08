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
			e.checkThrown()
		}
		return e.own(ir.Object, fmt.Sprintf("adamic_retain(%s)", arguments[0]))
	}
	panic("native: unknown Object method")
}
