package nexus

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers nexus/concurrency-no-lost-update.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common.
//
// No decoder: the rule takes no options. The judgment is structural, and an option naming targets to
// ignore would be an allowance.
func init() {
	rule.Register(rule.Registration{Rule: ConcurrencyNoLostUpdate})
}
