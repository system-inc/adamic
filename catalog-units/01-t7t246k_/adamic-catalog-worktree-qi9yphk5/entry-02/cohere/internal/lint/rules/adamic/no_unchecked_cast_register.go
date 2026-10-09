package adamic

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers adamic/no-unchecked-cast. No decoder: which casts are checked is Adamic's decision
// (#bw3xg7c), not a project's.
func init() {
	rule.Register(rule.Registration{Rule: NoUncheckedCast})
}
