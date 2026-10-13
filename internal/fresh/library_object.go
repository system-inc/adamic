package fresh

import "github.com/system-inc/adamic/internal/ir"

// Scalar reflection and assignment do not capture mutable values. Prototype operations
// conservatively capture their ownership graph until fresh analysis models its links.
func (a *analysis) objectCall(call ir.ObjectCall) value {
	if call.Method == "create" || call.Method == "setPrototypeOf" || call.Method == "getPrototypeOf" {
		return a.call(a.operands(call), call.Type())
	}
	var target value
	for index, argument := range call.Arguments {
		held := a.value(argument)
		if index == 0 {
			target = held
		}
	}
	switch call.Method {
	case "freeze", "assign":
		return target
	case "keys", "values", "entries":
		return a.fresh(elementKey, value{})
	}
	return value{}
}
