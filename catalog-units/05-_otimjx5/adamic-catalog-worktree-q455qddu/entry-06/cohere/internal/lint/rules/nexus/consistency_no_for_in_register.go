package nexus

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers nexus/consistency-no-for-in.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common.
//
// No decoder: the rule takes no options. Every `for...in` is the shape being retired, and an option
// naming loops that may keep it would be the allowance the ruling declined.
func init() {
	rule.Register(rule.Registration{Rule: ConsistencyNoForIn})
}
