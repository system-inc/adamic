package ir

import "os"

// An instrument for #xpkpx8p, never to land. It prices design (a) for a catchable "RangeError:
// Invalid string length" (#c3bk0fp): every operation that makes a string gets an exception edge.
// Under ADAMIC_MEASURE_STRING_CHECKS=1, lower's throwsOut and flow's CanThrow count the nodes
// MeasuredStringOperation names as able to throw, and the native emitter tests the pending word after
// each one's result. The runtime never sets the word from them: this prices the checks, the cleanup
// code and the lost elision, not the throws.

// MeasureStringChecks is ADAMIC_MEASURE_STRING_CHECKS=1, read once when the compiler starts.
var MeasureStringChecks = os.Getenv("ADAMIC_MEASURE_STRING_CHECKS") == "1"

// MeasuredStringOperation reports whether node is one the instrument counts: a concatenation (+ on
// strings, a template, concat, and text = text + ...), slice (substring and substr lower to it),
// padStart, padEnd, repeat, toUpperCase, toLowerCase, normalize, join, and String(x) or number to
// string where it makes a string. A number format whose argument can already throw is counted by its
// own MayThrow, and checked once.
func MeasuredStringOperation(node any) bool {
	switch node := node.(type) {
	case Concat, ArrayJoin, NumberToString, UnionToString:
		return true
	case MaybeToString:
		return node.Value.Type() == MaybeNumber
	case NumberFormat:
		return node.Method == "toString" && !node.MayThrow()
	case StringCall:
		switch node.Method {
		case "slice", "padStart", "padEnd", "repeat", "toUpperCase", "toLowerCase", "normalize":
			return true
		}
	}
	return false
}
