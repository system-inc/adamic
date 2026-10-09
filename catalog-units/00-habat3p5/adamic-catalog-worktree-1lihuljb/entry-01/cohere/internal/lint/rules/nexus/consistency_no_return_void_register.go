package nexus

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers nexus/consistency-no-return-void.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common.
//
// No decoder: the rule takes no options. An option widening it to arrow bodies would reach the
// `() => void promise()` callback idiom the ruling left alone, and one narrowing it would be an
// allowance.
func init() {
	rule.Register(rule.Registration{Rule: ConsistencyNoReturnVoid})
}
