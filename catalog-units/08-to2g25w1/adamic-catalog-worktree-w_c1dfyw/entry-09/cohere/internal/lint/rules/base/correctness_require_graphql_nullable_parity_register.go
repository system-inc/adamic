package base

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers base/correctness-require-graphql-nullable-parity.
//
// Registered and deliberately NOT enabled, like every base rule landing tonight: nothing in this
// tree carries `base/` keys, because base is api-phi-health's own lint layer, so where these run is
// a decision nobody has made.
//
// No Decode entry: the original's `schema: []` means no option surface.
func init() {
	rule.Register(rule.Registration{Rule: GraphQlNullableParity})
}
