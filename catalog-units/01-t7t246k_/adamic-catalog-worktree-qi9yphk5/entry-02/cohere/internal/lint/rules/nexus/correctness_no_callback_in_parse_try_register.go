package nexus

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers nexus/correctness-no-callback-in-parse-try.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common.
//
// No decoder: the rule takes no options. Its parsers are the two read and verified to throw on
// malformed text, and an option adding to them could name one that returns an outcome instead,
// which would report a try that guards nothing.
func init() {
	rule.Register(rule.Registration{Rule: CorrectnessNoCallbackInParseTry})
}
