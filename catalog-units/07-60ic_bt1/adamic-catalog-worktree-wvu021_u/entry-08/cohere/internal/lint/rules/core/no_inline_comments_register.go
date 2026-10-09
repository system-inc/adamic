package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers this package's `no-inline-comments` rule.
//
// A file of its own rather than a line in `register.go`, so a port is three owned files and no
// shared edit. The decoder is hand-written so a bare `"error"` configuration resolves to the
// documented default; see DecodeNoInlineCommentsOptions.
func init() {
	rule.Register(rule.Registration{Rule: NoInlineComments, Decode: DecodeNoInlineCommentsOptions})
}
