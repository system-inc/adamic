package adamic

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers adamic/invariant-mutable. No decoder: the rule takes no options, since a mutable slot
// is either invariant or a hole.
func init() {
	rule.Register(rule.Registration{Rule: InvariantMutable})
}
