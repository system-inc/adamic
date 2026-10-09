package nexus

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers nexus/correctness-no-discarded-outcome.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common.
//
// No decoder: the rule takes no options. Its outcome types are the ones read in Nexus and named in the
// doc comment, and an option naming more would let a configuration point at a type nobody has read,
// or at one whose every arm is a success.
func init() {
	rule.Register(rule.Registration{Rule: CorrectnessNoDiscardedOutcome})
}
