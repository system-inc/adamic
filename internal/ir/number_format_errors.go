package ir

import "math"

// NumberFormatMayThrow keeps the emitter, exception inference and flow edges in agreement.
// Constant arguments whose ToIntegerOrInfinity is in range need no exceptional edge.
func NumberFormatMayThrow(value any) bool {
	var argument Expression
	low, high := 0.0, 100.0
	switch value := value.(type) {
	case ToFixed:
		argument = value.Digits
	case NumberFormat:
		argument = value.Argument
		if value.Method == "toPrecision" {
			low = 1
		}
		if value.Method == "toString" {
			low, high = 2, 36
		}
	default:
		return false
	}
	if argument == nil {
		return false
	}
	number, constant := argument.(NumberConstant)
	if !constant {
		return true
	}
	integer := math.Trunc(number.Value)
	if math.IsNaN(integer) {
		integer = 0
	}
	return integer < low || integer > high
}
