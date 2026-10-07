// Functions and declarations moved unchanged from emit.go.
package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

var cMath = map[string]string{
	"abs": "fabs", "ceil": "ceil", "floor": "floor", "trunc": "trunc", "sqrt": "sqrt",
	"round": "adamic_math_round", "sign": "adamic_math_sign", "pow": "adamic_power",
}

// numberFormat emits toExponential and toPrecision (runtime/dtoa.c) and toString with a radix
// (runtime/radix.c), the receiver before the argument, as JavaScript reads them.
func (e *emitter) numberFormat(format ir.NumberFormat) string {
	value := e.value(format.Value)
	if format.Method == "toString" {
		if format.Argument == nil {
			return e.own(ir.String, fmt.Sprintf("adamic_string_from_number(%s)", value))
		}
		return e.own(ir.String, fmt.Sprintf("adamic_number_to_radix(%s, %s)", value, e.value(format.Argument)))
	}
	argument, hasArgument := "0.0", "false"
	if format.Argument != nil {
		argument, hasArgument = e.value(format.Argument), "true"
	}
	function := map[string]string{"toExponential": "adamic_number_to_exponential", "toPrecision": "adamic_number_to_precision"}[format.Method]
	return e.own(ir.String, fmt.Sprintf("%s(%s, %s, %s)", function, value, argument, hasArgument))
}

// cIeee754 are the Math functions ported from V8 (runtime/ieee754.c), each adamic_math_<name>. C's
// own sin, exp and the rest differ from V8's in the last bit, so they're never called.
var cIeee754 = map[string]bool{
	"sin": true, "cos": true, "tan": true, "asin": true, "acos": true, "atan": true, "atan2": true,
	"sinh": true, "cosh": true, "tanh": true, "asinh": true, "acosh": true, "atanh": true,
	"exp": true, "expm1": true, "log": true, "log1p": true, "log2": true, "log10": true, "cbrt": true,
}

func (e *emitter) mathCall(call ir.MathCall) string {
	if value, known := e.libraryMathCall(call); known {
		return value
	}
	if call.Spread != nil {
		return e.snapshot(ir.Number, fmt.Sprintf("adamic_math_%s_of(%s)", call.Function, e.value(call.Spread)))
	}
	arguments := make([]string, 0, len(call.Arguments))
	for _, argument := range call.Arguments {
		arguments = append(arguments, e.value(argument))
	}
	if function, isDirect := cMath[call.Function]; isDirect {
		return fmt.Sprintf("%s(%s)", function, strings.Join(arguments, ", "))
	}
	if cIeee754[call.Function] {
		return fmt.Sprintf("adamic_math_%s(%s)", call.Function, strings.Join(arguments, ", "))
	}
	if call.Function == "hypot" {
		// Math.hypot() is +0; otherwise the values go as an array, already evaluated in order.
		if len(arguments) == 0 {
			return "0.0"
		}
		return fmt.Sprintf("adamic_math_hypot(%d, (const double[]){%s})", len(arguments), strings.Join(arguments, ", "))
	}
	// max and min take any number of arguments: none gives -Infinity (max) or Infinity (min), one
	// gives itself, and more fold pairwise, every argument already evaluated, as JavaScript does.
	switch len(arguments) {
	case 0:
		if call.Function == "max" {
			return "(-HUGE_VAL)"
		}
		return "HUGE_VAL"
	case 1:
		return "(+" + arguments[0] + ")"
	}
	folded := arguments[0]
	for _, argument := range arguments[1:] {
		folded = fmt.Sprintf("adamic_math_%s(%s, %s)", call.Function, folded, argument)
	}
	return folded
}
