package javascript

import "github.com/system-inc/adamic/internal/ir"

func (e *emitter) objectDescriptorCall(call ir.ObjectCall) (string, bool) {
	switch call.Method {
	case "propertyIsEnumerable":
		return "(" + e.value(call.Arguments[0]) + ").propertyIsEnumerable(" + e.value(call.Arguments[1]) + ")", true
	case "descriptorFlag", "descriptorValue", "descriptorAbsent":
		return "(" + e.value(call.Arguments[0]) + ")?.[" + e.value(call.Arguments[1]) + "]", true
	}
	return "", false
}
