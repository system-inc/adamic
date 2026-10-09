package adamic

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers adamic/no-optional-widening. No decoder: the rule takes no options.
func init() {
	rule.Register(rule.Registration{Rule: NoOptionalWidening})
}
