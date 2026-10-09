package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers this package's `no-continue` rule.
//
// A file of its own rather than a line in `register.go`, so a port is three owned files and no
// shared edit. No `Decode`: upstream's schema is `[]`, so there is nothing to configure.
func init() {
	rule.Register(rule.Registration{Rule: NoContinue})
}
