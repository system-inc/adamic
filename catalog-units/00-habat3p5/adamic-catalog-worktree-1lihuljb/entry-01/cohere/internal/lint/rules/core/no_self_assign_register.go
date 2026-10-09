package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers this package's `no-self-assign` rule.
//
// A file of its own rather than a line in `register.go`, so a port is three owned files and no
// shared edit. The decoder is hand-written so a bare `"error"` and `{}` both resolve to the
// documented default of `props: true`; see DecodeNoSelfAssignOptions.
func init() {
	rule.Register(rule.Registration{Rule: NoSelfAssign, Decode: DecodeNoSelfAssignOptions})
}
