package react

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers this package's `no-unused-class-component-methods` rule.
//
// No `Decode`, because `meta.schema` is the empty array and the rule takes no options.
func init() {
	rule.Register(rule.Registration{Rule: NoUnusedClassComponentMethods})
}
