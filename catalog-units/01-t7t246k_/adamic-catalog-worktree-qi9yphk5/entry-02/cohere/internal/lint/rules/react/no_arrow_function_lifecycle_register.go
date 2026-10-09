package react

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers this package's `no-arrow-function-lifecycle` rule.
//
// A file of its own rather than a line in `register.go`, so a port is three owned files and no
// shared edit.
//
// No `Decode`, because upstream declares `schema: []` and the rule takes no options.
func init() {
	rule.Register(rule.Registration{Rule: NoArrowFunctionLifecycle})
}
