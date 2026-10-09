package typescript

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers prefer-promise-reject-errors.
//
// Its own file rather than a line in the package's register.go: that shared list is one file every
// concurrent port has to edit, so two ports landing at once conflict on a file neither is really
// changing.
//
// The Decode entry is required because `allow` arrives as one heterogeneous array of strings and
// specifier objects and leaves as two typed fields, which no struct tag can express. The three
// booleans would bind straight through, but they all default to false, so a rule handed nil options
// already behaves correctly without them.
func init() {
	rule.Register(rule.Registration{Rule: PreferPromiseRejectErrors, Decode: DecodePreferPromiseRejectErrorsOptions})
}
