package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers no-else-return.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common.
//
// The decoder is hand-rolled because `allowElseIf` defaults to TRUE. A generic decoder would hand
// back a zero-value struct with it false, and false is not a no-op: it switches the rule from
// judging an `else if` chain as a whole to judging each link, which reports chains upstream
// deliberately allows.
func init() {
	rule.Register(rule.Registration{
		Rule:   NoElseReturn,
		Decode: DecodeNoElseReturnOptions,
	})
}
