package nexus

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers nexus/correctness-no-import-cycle-load-time-read.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common.
//
// No decoder: the rule takes no options. What counts as a runtime edge and as a read at load is the
// language's, not a project's, so there is nothing for a configuration to say.
func init() {
	rule.Register(rule.Registration{Rule: CorrectnessNoImportCycleLoadTimeRead})
}
