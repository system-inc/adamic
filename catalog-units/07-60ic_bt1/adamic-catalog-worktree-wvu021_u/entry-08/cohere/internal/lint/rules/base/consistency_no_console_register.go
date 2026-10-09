package base

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers base/consistency-no-console.
//
// Its own file rather than a line in a shared register.go, so two ports landing at once conflict on
// nothing.
//
// No Decode entry: `schema: []` upstream and the rule reads no options.
func init() {
	rule.Register(rule.Registration{Rule: ConsistencyNoConsole})
}
