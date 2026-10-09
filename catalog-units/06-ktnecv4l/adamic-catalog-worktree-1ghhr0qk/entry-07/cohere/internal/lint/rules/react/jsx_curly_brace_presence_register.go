package react

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers this package's `jsx-curly-brace-presence` rule.
//
// The decoder is hand-written rather than `rule.DecodeOptionsInto` because the option surface is
// sometimes a string and sometimes an object, and because two of its three defaults are not Go zero
// values. See `DecodeJsxCurlyBracePresenceOptions`.
func init() {
	rule.Register(rule.Registration{Rule: JsxCurlyBracePresence, Decode: DecodeJsxCurlyBracePresenceOptions})
}
