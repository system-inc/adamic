package typescript

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers @typescript-eslint/consistent-return.
//
// Its own file rather than a line in the package's register.go: that shared list is one file every
// concurrent port has to edit, so two ports landing at once conflict on a file neither is really
// changing.
//
// The Decode entry is required rather than defensive. Without it a configured
// `{"treatUndefinedAsUnspecified": true}` would reach the rule as a raw value the type assertion
// rejects, and the rule would silently run with the option off -- a rule that registers, reports,
// and enforces something other than what the config says.
func init() {
	rule.Register(rule.Registration{
		Rule:   ConsistentReturn,
		Decode: DecodeConsistentReturnOptions,
	})
}
