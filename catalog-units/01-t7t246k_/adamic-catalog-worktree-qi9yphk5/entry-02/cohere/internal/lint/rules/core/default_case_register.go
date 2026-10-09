package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers this package's `default-case` rule.
//
// A file of its own rather than a line in `register.go`, so a port is three owned files and no
// shared edit. The decoder is hand-written rather than `rule.DecodeOptionsInto` so a bare `"error"`
// configuration resolves to the documented default; see DecodeDefaultCaseOptions.
func init() {
	rule.Register(rule.Registration{Rule: DefaultCase, Decode: DecodeDefaultCaseOptions})
}
