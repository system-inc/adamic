package nexus

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers nexus/concurrency-no-check-then-write.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common.
//
// No decoder: the rule takes no options. Its judgment is a proof that one path is checked and then
// written, and an option widening the writes or the path shapes it accepts would reach the guesses
// its doc comment declines on purpose.
func init() {
	rule.Register(rule.Registration{Rule: ConcurrencyNoCheckThenWrite})
}
