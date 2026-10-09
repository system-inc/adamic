package typescript

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers no-misused-spread.
//
// Its own file rather than a line in the package's register.go: that shared list is one file every
// concurrent port has to edit, so two ports landing at once conflict on a file neither is really
// changing.
//
// The Decode entry is required because `allow` arrives as one heterogeneous array of strings and
// specifier objects and leaves as two typed fields, which no struct tag can express. Unlike
// restrict-plus-operands, the zero value is already correct, so nil options need no fallback.
func init() {
	rule.Register(rule.Registration{Rule: NoMisusedSpread, Decode: DecodeNoMisusedSpreadOptions})
}
