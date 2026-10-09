package react

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers this package's `no-set-state` rule.
//
// A file of its own rather than a line in `register.go`, so a port is three owned files and no
// shared edit. No decoder, because upstream's schema is empty.
func init() {
	rule.Register(rule.Registration{Rule: NoSetState})
}
