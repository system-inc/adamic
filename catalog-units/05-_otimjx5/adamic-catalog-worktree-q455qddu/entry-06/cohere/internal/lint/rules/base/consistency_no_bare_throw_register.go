package base

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers base/consistency-no-bare-throw.
//
// Its own file rather than a line in a shared register.go: that list is one file every concurrent
// port has to edit, so two ports landing at once conflict on a file neither is really changing.
//
// No Decode entry, because the original's `schema` is `[]` and the rule reads no options at all.
//
// Registered and NOT enabled. `api-phi-health` carries no CohereSettings.json of its own, and ahra's
// config names no `base/` keys because base is that project's layer rather than ahra's. Where these
// rules get turned on is a decision for whoever owns that configuration, and a port that enabled
// itself would be making it silently.
func init() {
	rule.Register(rule.Registration{Rule: ConsistencyNoBareThrow})
}
