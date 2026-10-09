package nexus

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers nexus/consistency-no-hand-rolled-delay.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common.
//
// No decoder: the rule takes no options. Which timer shapes count as a plain wait is the rule's
// judgment, and an option to exempt one would be an allowance.
func init() {
	rule.Register(rule.Registration{Rule: ConsistencyNoHandRolledDelay})
}
