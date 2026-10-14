package javascript

import "github.com/system-inc/adamic/internal/ir"

func objectPrototypeValue(expression ir.Expression) (string, bool) {
	if call, ok := expression.(ir.ObjectCall); ok && call.Method == "intrinsicObjectPrototype" {
		return "Object.prototype", true
	}
	return "", false
}
