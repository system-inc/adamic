package adamic

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers adamic/no-definite-assignment. No decoder: an unchecked claim is one whatever it names.
func init() {
	rule.Register(rule.Registration{Rule: NoDefiniteAssignment})
}
