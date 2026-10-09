package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers no-unneeded-ternary.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common.
//
// The decoder is hand-rolled because the only option defaults to TRUE. `rule.DecodeOptionsInto`
// errors on empty input, the config layer turns that into nil, and a nil behind a plain bool field
// yields false, which would silently switch the default-assignment judgment ON for every consumer.
// The field behind the decoder is a pointer so an absent key stays distinguishable from an explicit
// false.
func init() {
	rule.Register(rule.Registration{
		Rule:   NoUnneededTernary,
		Decode: DecodeNoUnneededTernaryOptions,
	})
}
