package adamic

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers adamic/no-type-predicate. No decoder: every written predicate is trusted, so every one
// is reported.
func init() {
	rule.Register(rule.Registration{Rule: NoTypePredicate})
}
