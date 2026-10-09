package react

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers this package's `no-invalid-html-attribute` rule.
//
// A file of its own rather than a line in `register.go`, so a port is three owned files and no
// shared edit. The decoder is hand-written because the option is a bare list rather than an
// object; see DecodeNoInvalidHtmlAttributeOptions.
func init() {
	rule.Register(rule.Registration{Rule: NoInvalidHtmlAttribute, Decode: DecodeNoInvalidHtmlAttributeOptions})
}
