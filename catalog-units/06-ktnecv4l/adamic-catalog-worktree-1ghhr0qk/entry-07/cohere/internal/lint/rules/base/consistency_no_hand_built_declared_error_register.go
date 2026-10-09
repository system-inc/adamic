package base

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers base/consistency-no-hand-built-declared-error.
//
// Its own file rather than a line in a shared register.go, so two ports landing at once do not
// conflict on a file neither is really changing.
//
// The rule is registered and DELIBERATELY NOT ENABLED. api-phi-health has no CohereSettings.json of
// its own, and ahra's config carries no `base/` keys because base is that project's layer rather
// than ahra's, so where this gets enabled is a decision rather than an omission.
//
// No Decode entry: the original's `schema: []` means the rule has no option surface at all.
func init() {
	rule.Register(rule.Registration{Rule: ConsistencyNoHandBuiltDeclaredError})
}
