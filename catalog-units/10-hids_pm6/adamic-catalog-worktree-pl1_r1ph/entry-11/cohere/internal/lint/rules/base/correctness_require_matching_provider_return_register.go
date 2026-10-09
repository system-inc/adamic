package base

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers base/correctness-require-matching-provider-return.
//
// A file of its own rather than a line in a shared `register.go`, so two authors landing at once
// touch nothing in common.
//
// Registered and left unenabled, the same as every other base rule here. `base` is api-phi-health's
// lint layer rather than ahra's, and ahra's config names no `base/` rule at all, so which trees
// enforce it is Kirk's call rather than a porter's.
func init() {
	rule.Register(rule.Registration{Rule: CorrectnessRequireMatchingProviderReturn})
}
