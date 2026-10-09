package react

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers this package's `no-access-state-in-setstate` rule.
func init() {
	rule.Register(rule.Registration{Rule: NoAccessStateInSetstate})
}
