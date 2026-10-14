package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (e *emitter) objectPrototypeCall(call ir.ObjectCall, arguments []string) (string, bool) {
	switch call.Method {
	case "intrinsicObjectPrototype":
		e.declarations = append(e.declarations, "#include \"library_prototypes.h\"")
		return e.own(ir.Object, "adamic_object_intrinsic_prototype()"), true
	case "create", "setPrototypeOf", "getPrototypeOf":
		e.declarations = append(e.declarations, "#include \"library_prototypes.h\"")
		method := map[string]string{"create": "create", "setPrototypeOf": "set", "getPrototypeOf": "get"}[call.Method]
		value := e.own(ir.Object, fmt.Sprintf("adamic_object_%s_prototype(%s)", method, strings.Join(arguments, ", ")))
		if call.Method == "setPrototypeOf" {
			e.checkThrown()
		}
		return value, true
	}
	return "", false
}
