package adamic

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers adamic/single-spread. No decoder: a later spread is a hole whatever the project prefers.
func init() {
	rule.Register(rule.Registration{Rule: SingleSpread})
}
