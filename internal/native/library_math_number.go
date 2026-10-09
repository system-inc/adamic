package native

import (
	"fmt"
	"strings"

	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) libraryMathCall(call ir.MathCall) (string, bool) {
	switch call.Function {
	case "clz32", "fround", "imul":
	default:
		return "", false
	}
	arguments := make([]string, len(call.Arguments))
	for index, argument := range call.Arguments {
		arguments[index] = e.snapshot(ir.Number, e.value(argument))
	}
	return e.snapshot(ir.Number, "adamic_math_"+call.Function+"("+strings.Join(arguments, ", ")+")"), true
}

func (e *emitter) libraryNumberCall(call ir.NumberCall) (string, bool) {
	if call.Function == "toBoolean" {
		return e.censusToBoolean(call.Arguments[0]), true
	}
	if call.Function == "hasOwnProperty" || call.Function == "prototypeHasOwnProperty" {
		prototype := "false"
		if call.Function == "prototypeHasOwnProperty" {
			prototype = "true"
		}
		return e.snapshot(ir.Boolean, "adamic_number_has_own_property("+e.value(call.Arguments[0])+", "+prototype+")"), true
	}
	if call.Function != "convert" {
		return "", false
	}
	argument := call.Arguments[0]
	value := e.value(argument)
	switch argument.Type() {
	case ir.Boolean:
		value = "(double)(" + value + ")"
	case ir.String:
		value = "adamic_number_from_string(" + value + ")"
	case ir.Union:
		value = "adamic_number_from_union(" + value + ")"
	case ir.MaybeNumber, ir.MaybeBoolean:
		value = e.snapshot(argument.Type(), value)
		field := "number"
		if argument.Type() == ir.MaybeBoolean {
			field = "boolean"
		}
		value = fmt.Sprintf("(%s.present ? (double)%s.%s : NAN)", value, value, field)
	}
	return e.snapshot(ir.Number, value), true
}

// A format can leave through a RangeError, after both operands were evaluated in order.
func (e *emitter) libraryNumberFormat(method string, receiver, argument ir.Expression) string {
	value := e.snapshot(ir.Number, e.value(receiver))
	digits := e.snapshot(ir.Number, e.value(argument))
	function := map[string]string{"fixed": "fixed", "toString": "radix", "toExponential": "exponential", "toPrecision": "precision"}[method]
	result := e.own(ir.String, fmt.Sprintf("adamic_number_checked_%s(%s, %s)", function, value, digits))
	e.checkThrown()
	return result
}
