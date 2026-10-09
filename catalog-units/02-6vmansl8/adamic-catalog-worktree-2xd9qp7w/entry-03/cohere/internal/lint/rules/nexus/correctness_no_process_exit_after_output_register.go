package nexus

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers nexus/correctness-no-process-exit-after-output.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common.
//
// No decoder: the rule takes no options. What counts as a write and as the exit is decided by the
// symbols `@types/node` declares, and an option widening either would let a configuration name a
// logger that writes to a file, which is the false positive the doc comment declines on purpose.
func init() {
	rule.Register(rule.Registration{Rule: CorrectnessNoProcessExitAfterOutput})
}
