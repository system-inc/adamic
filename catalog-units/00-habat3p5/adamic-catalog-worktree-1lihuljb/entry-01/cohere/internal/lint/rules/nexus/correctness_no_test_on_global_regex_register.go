package nexus

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers nexus/correctness-no-test-on-global-regex.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common.
//
// No decoder: the rule takes no options, because the exemptions are the shapes that use `g` on
// purpose, and an option widening them would let a configuration silence the shared-regex case this
// rule exists for.
func init() {
	rule.Register(rule.Registration{Rule: CorrectnessNoTestOnGlobalRegex})
}
