package typescript

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers @typescript-eslint/no-base-to-string.
//
// Its own file rather than a line in the package's register.go: that shared list is one file every
// concurrent port has to edit, so two ports landing at once conflict on a file neither is really
// changing.
//
// The Decode entry is required rather than defensive here. This rule's ignoredTypeNames defaults to
// four builtin names rather than to a zero value, so the generic helper would hand the rule an empty
// slice for an absent key and it would start reporting on every Error.
func init() {
	rule.Register(rule.Registration{
		Rule:   NoBaseToString,
		Decode: DecodeNoBaseToStringOptions,
	})
}
