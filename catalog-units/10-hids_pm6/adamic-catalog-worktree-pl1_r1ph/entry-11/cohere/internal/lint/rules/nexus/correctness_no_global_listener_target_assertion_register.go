package nexus

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers nexus/correctness-no-global-listener-target-assertion.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common.
//
// No decoder: the rule takes no options. The general element types it leaves alone are the four the
// default library declares, and an option widening them would only hide the lie it reports.
func init() {
	rule.Register(rule.Registration{Rule: CorrectnessNoGlobalListenerTargetAssertion})
}
