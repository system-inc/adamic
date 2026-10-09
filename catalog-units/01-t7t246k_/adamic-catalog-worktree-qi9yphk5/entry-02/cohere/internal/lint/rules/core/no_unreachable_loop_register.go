package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers no-unreachable-loop.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors
// landing at once touch nothing in common.
func init() {
	rule.Register(rule.Registration{
		Rule:   NoUnreachableLoop,
		Decode: rule.DecodeOptionsInto[NoUnreachableLoopOptions](),
	})
}
