package ir

import "math"

// An array made with a length out of range throws Node's RangeError, which a catch takes
// (runtime/array.c and array_from.c set it on the pending word). Lowering counts the calls that can,
// the flow graph gives them an exception edge, and the emitter tests the word after each one.

// MayThrow reports whether new Array(length) can be given a length that isn't an integer from 0 to
// 2^32 - 1: unless it's a constant within. A fill of an array already made can't throw.
func (call ArrayFill) MayThrow() bool {
	return call.Array == nil && !ConstantWithin(call.Length, 0, 4294967295)
}

// MayThrow reports whether Array.from({ length }) can be given a length past 2^32 - 1: unless it's a
// constant integer up to it.
func (call ArrayFrom) MayThrow() bool {
	return !ConstantWithin(call.Length, math.Inf(-1), 4294967295)
}
