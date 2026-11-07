package native

import (
	"fmt"
	"strings"

	"github.com/system-inc/adamic/internal/ir"
)

func (nativeEmitter *emitter) objectCall(call ir.ObjectCall) string {
	arguments := make([]string, 0, len(call.Arguments))
	for _, argument := range call.Arguments {
		arguments = append(arguments, nativeEmitter.value(argument))
	}
	switch call.Method {
	case "is":
		return nativeEmitter.snapshot(ir.Boolean, fmt.Sprintf("adamic_object_is(%s, %s)", arguments[0], arguments[1]))
	case "isFrozen":
		return nativeEmitter.snapshot(ir.Boolean, fmt.Sprintf("adamic_object_is_frozen(%s)", arguments[0]))
	case "freeze":
		return nativeEmitter.own(ir.Object, fmt.Sprintf("adamic_object_freeze(%s)", arguments[0]))
	case "hasOwn":
		return nativeEmitter.snapshot(
			ir.Boolean,
			fmt.Sprintf("adamic_object_has_own(%s, %s)", arguments[0], arguments[1]),
		)
	case "keys":
		if call.Arguments[0].Type() == ir.Array {
			return nativeEmitter.own(ir.Array, fmt.Sprintf("adamic_node_array_keys(%s)", arguments[0]))
		}
		return nativeEmitter.own(ir.Array, fmt.Sprintf("adamic_object_keys(%s)", arguments[0]))
	case "values", "entries":
		if call.Checked {
			nativeEmitter.declarations = append(
				nativeEmitter.declarations,
				"adamic_array *adamic_object_values_checked(adamic_object *, int, const char *, bool, const adamic_value *, size_t);",
			)
			allowed := "NULL"
			if len(call.Allowed) > 0 {
				allowed = nativeEmitter.temporary()
				values := make([]string, 0, len(call.Allowed))
				for _, value := range call.Allowed {
					values = append(values, fmt.Sprintf("{.%s = %s}", member(call.Element), nativeEmitter.value(value)))
				}
				nativeEmitter.line("const adamic_value %s[] = {%s};", allowed, strings.Join(values, ", "))
			}
			return nativeEmitter.own(ir.Array, fmt.Sprintf(
				"adamic_object_values_checked(%s, %d, %s, %t, %s, %d)",
				arguments[0],
				call.Element,
				cString(call.ElementName),
				call.Method == "entries",
				allowed,
				len(call.Allowed),
			))
		}
		return nativeEmitter.own(ir.Array, fmt.Sprintf(
			"adamic_object_values(%s, %s, %t)",
			arguments[0],
			arrayElement(call.Element),
			call.Method == "entries",
		))
	case "assign":
		for _, source := range arguments[1:] {
			nativeEmitter.line("adamic_object_assign(%s, %s);", arguments[0], source)
			// A frozen target leaves Node's TypeError pending at its first key, and writes no other.
			if nativeEmitter.program.FrozenWriteMayThrow(call) {
				nativeEmitter.checkThrown()
			}
		}
		return nativeEmitter.own(ir.Object, fmt.Sprintf("adamic_retain(%s)", arguments[0]))
	}
	panic("native: unknown Object method")
}
