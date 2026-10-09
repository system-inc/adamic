package nexus

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers nexus/correctness-no-write-only-collection.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common.
//
// No decoder: the rule takes no options, because the writer methods are the default library's, read
// and fixed, and an option adding a project method would let a configuration name one that also reads.
func init() {
	rule.Register(rule.Registration{Rule: CorrectnessNoWriteOnlyCollection})
}
