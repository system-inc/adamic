package javascript

import "github.com/system-inc/adamic/internal/ir"

func (e *emitter) dateCall(call ir.DateCall) string {
	args := e.values(call.Arguments)
	switch call.Method {
	case "nullableTypeOf":
		return "typeof (" + e.value(call.Receiver) + ")"
	case "nullableNumber":
		return "Number(" + e.value(call.Receiver) + ")"
	case "new", "components", "newString":
		return "new Date(" + args + ")"
	case "copy":
		return "new Date(" + e.value(call.Receiver) + ")"
	case "UTC", "parse":
		return "Date." + call.Method + "(" + args + ")"
	}
	return "(" + e.value(call.Receiver) + ")." + call.Method + "(" + args + ")"
}
