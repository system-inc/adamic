package fresh

import "github.com/system-inc/adamic/internal/ir"

// These methods never capture mutable values. The lowering admits only scalar values and entries,
// and assign only copies numbers, booleans and strings into fields the target already has.
func (a *analysis) objectCall(call ir.ObjectCall) value {
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
