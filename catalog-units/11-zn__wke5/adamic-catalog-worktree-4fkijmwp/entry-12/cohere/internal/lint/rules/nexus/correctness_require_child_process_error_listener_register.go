package nexus

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers nexus/correctness-require-child-process-error-listener.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common.
//
// No decoder: the rule takes no options. Its producers are the two `node:child_process` functions
// read and confirmed to return a child with no `'error'` listener of its own, and an option adding to
// them could name `execFile`, which attaches one, and make every use of it a false finding.
func init() {
	rule.Register(rule.Registration{Rule: CorrectnessRequireChildProcessErrorListener})
}
