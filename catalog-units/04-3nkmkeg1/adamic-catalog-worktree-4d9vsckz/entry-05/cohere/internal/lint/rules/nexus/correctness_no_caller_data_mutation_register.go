package nexus

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers nexus/correctness-no-caller-data-mutation.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common.
//
// No decoder: the rule takes no options, because the line it draws is where a type is declared, and an
// option naming parameters to excuse is the allowance list the rule exists to replace.
func init() {
	rule.Register(rule.Registration{Rule: CorrectnessNoCallerDataMutation})
}
