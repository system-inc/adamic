package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers no-implicit-coercion.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common.
//
// The decoder is hand-rolled because three of the four booleans default to TRUE. The generic helper
// errors on empty input, the config layer turns that into nil, and a zero-valued struct would leave
// every one of those three false, disabling three of the four judgments.
func init() {
	rule.Register(rule.Registration{
		Rule:   NoImplicitCoercion,
		Decode: DecodeNoImplicitCoercionOptions,
	})
}
