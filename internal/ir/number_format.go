package ir

import "math"

// A number format given an argument out of its range throws Node's RangeError, which a catch takes
// (runtime/number.c, dtoa.c and radix.c set it on the pending word). Lowering counts the calls that
// can, the flow graph gives them an exception edge, and the emitter tests the word after each one.

// MayThrow reports whether toFixed's digits can be out of 0 to 100: unless they're a constant within.
func (call ToFixed) MayThrow() bool {
	return !ConstantWithin(call.Digits, 0, 100)
}

// MayThrow reports whether the argument can be out of the method's range. Left out, it can't: each
// method has a meaning of its own for that (NumberFormat).
func (call NumberFormat) MayThrow() bool {
	bounds := formatArguments[call.Method]
	return call.Argument != nil && !ConstantWithin(call.Argument, bounds[0], bounds[1])
}

// formatArguments are the arguments each number format takes without throwing.
var formatArguments = map[string][2]float64{"toExponential": {0, 100}, "toPrecision": {1, 100}, "toString": {2, 36}}

// ConstantWithin reports whether a value is a constant integer from low to high, which a call taking
// it can't fail on.
func ConstantWithin(value Expression, low float64, high float64) bool {
	constant, isConstant := value.(NumberConstant)
	return isConstant && constant.Value >= low && constant.Value <= high && (high == math.MaxFloat64 || constant.Value == math.Trunc(constant.Value)) && !math.IsInf(constant.Value, 0)
}
