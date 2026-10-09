package nexus

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers nexus/correctness-no-discarded-pure-result.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common.
//
// No decoder: the rule takes no options, because the method sets are the default library's side-
// effect-free methods, read one by one, and an option adding to them would let a configuration name
// one that is not.
func init() {
	rule.Register(rule.Registration{Rule: CorrectnessNoDiscardedPureResult})
}
