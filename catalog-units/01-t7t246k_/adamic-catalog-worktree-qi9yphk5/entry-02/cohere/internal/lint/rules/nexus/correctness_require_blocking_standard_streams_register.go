package nexus

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers nexus/correctness-require-blocking-standard-streams.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common.
//
// No decoder: the rule takes no options. What blocks is decided by Nexus's declaration and by what
// the program's imports can reach, and an option naming another blocking function would let a
// configuration silence an entry that does not block, or report one that does.
func init() {
	rule.Register(rule.Registration{Rule: CorrectnessRequireBlockingStandardStreams})
}
