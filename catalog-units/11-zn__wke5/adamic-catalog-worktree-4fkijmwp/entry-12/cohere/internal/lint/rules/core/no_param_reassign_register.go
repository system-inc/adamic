package core

import "github.com/system-inc/cohere/internal/lint/rule"

// The decoder is hand-rolled rather than `rule.DecodeOptionsInto`, so that empty input decodes to the
// default rather than erroring into a nil the caller reads as a zero value. Nothing inverts here --
// every option's zero value is its default -- but the two shapes should not be told apart by whether
// the decoder happened to run.
func init() {
	rule.Register(rule.Registration{
		Rule:   NoParamReassign,
		Decode: DecodeNoParamReassignOptions,
	})
}
