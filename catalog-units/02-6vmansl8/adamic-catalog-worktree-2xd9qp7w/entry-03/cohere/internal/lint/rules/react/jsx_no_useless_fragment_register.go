package react

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers this package's `jsx-no-useless-fragment` rule.
//
// The decoder is hand-written rather than `rule.DecodeOptionsInto` because the generic one errors on
// empty input, which is what a bare `"error"` configuration hands it.
func init() {
	rule.Register(rule.Registration{Rule: JsxNoUselessFragment, Decode: DecodeJsxNoUselessFragmentOptions})
}
