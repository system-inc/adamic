package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers no-bitwise.
//
// Its own file rather than a line in the package's register.go: that shared list is one file every
// concurrent port has to edit, so two ports landing at once conflict on a file neither is changing.
//
// The Decode entry is required rather than defensive. Without it a configured `allow` list would
// reach the rule as a raw value the type assertion rejects, and the rule would run exempting
// nothing while the config said otherwise.
func init() {
	rule.Register(rule.Registration{Rule: NoBitwise, Decode: DecodeNoBitwiseOptions})
}
