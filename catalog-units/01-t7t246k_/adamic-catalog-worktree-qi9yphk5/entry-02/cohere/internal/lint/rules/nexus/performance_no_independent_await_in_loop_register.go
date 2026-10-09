package nexus

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers nexus/performance-no-independent-await-in-loop.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common.
//
// No decoder: the rule takes no options. The judgment is structural, and a knob for "which names
// count as output" would be the first thing tuned until the rule matched nothing.
func init() {
	rule.Register(rule.Registration{Rule: PerformanceNoIndependentAwaitInLoop})
}
