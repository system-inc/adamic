package nexus

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers nexus/correctness-require-response-status-check.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common.
//
// No decoder: the rule takes no options. Its set of producers is the set read and verified not to
// throw on an HTTP error, and an option adding to it would let a configuration name a wrapper that
// already throws, which is the false positive the doc comment declines on purpose.
func init() {
	rule.Register(rule.Registration{Rule: CorrectnessRequireResponseStatusCheck})
}
