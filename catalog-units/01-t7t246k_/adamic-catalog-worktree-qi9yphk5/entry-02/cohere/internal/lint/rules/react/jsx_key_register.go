package react

import "github.com/system-inc/cohere/internal/lint/rule"

// Registered with a hand-written decoder rather than `rule.DecodeOptionsInto`.
//
// Every default here is false, so the struct needs no translation; what the generic decoder cannot
// do is accept EMPTY input, which it turns into an error. A rule configured as a bare `"error"` is
// handed exactly that, and upstream accepts it, so the nil path has to answer the defaults.
func init() {
	rule.Register(rule.Registration{Rule: JsxKey, Decode: DecodeJsxKeyOptions})
}
