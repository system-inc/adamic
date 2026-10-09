package base

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers base/consistency-require-pagination-argument-name.
//
// A file of its own rather than a line in a shared `register.go`, so two authors landing at once
// touch nothing in common.
//
// Registered and left unenabled. `base` is api-phi-health's layer rather than ahra's, api-phi-health
// carries no CohereSettings.json of its own, and ahra's config names no `base/` rule. Where these
// get turned on is Kirk's call rather than a porter's.
func init() {
	rule.Register(rule.Registration{Rule: ConsistencyRequirePaginationArgumentName})
}
