package react

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers this package's `jsx-fragments` rule.
//
// The decoder is hand-written rather than `rule.DecodeOptionsInto` because the option is a bare
// enum string rather than an object; see DecodeJsxFragmentsOptions.
func init() {
	rule.Register(rule.Registration{Rule: JsxFragments, Decode: DecodeJsxFragmentsOptions})
}
