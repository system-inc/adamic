package nexus

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers nexus/correctness-no-mock-on-module-namespace.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common.
//
// No decoder: the rule takes no options. What it reports always throws, so there is nothing for a
// configuration to tune.
func init() {
	rule.Register(rule.Registration{Rule: CorrectnessNoMockOnModuleNamespace})
}
