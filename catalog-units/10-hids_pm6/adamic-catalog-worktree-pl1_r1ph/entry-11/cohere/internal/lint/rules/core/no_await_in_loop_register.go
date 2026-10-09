package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers no-await-in-loop.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common.
//
// No decoder: upstream declares `schema: []`, so the rule takes no options at all.
func init() {
	rule.Register(rule.Registration{Rule: NoAwaitInLoop})
}
