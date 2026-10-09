package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers no-underscore-dangle.
//
// Its own file rather than a line in the package's register.go: that shared list is one file every
// concurrent port has to edit, so two ports landing at once conflict on a file neither is changing.
//
// The Decode entry is required rather than defensive. This rule has nine options, and without a
// decoder a configured object reaches the rule as a raw value the type assertion rejects, leaving
// every flag at its default while the config says otherwise.
func init() {
	rule.Register(rule.Registration{
		Rule:   NoUnderscoreDangle,
		Decode: DecodeNoUnderscoreDangleOptions,
	})
}
