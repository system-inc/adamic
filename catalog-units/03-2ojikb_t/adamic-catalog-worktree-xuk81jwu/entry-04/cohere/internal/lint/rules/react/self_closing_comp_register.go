package react

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers this package's `self-closing-comp` rule.
//
// The decoder is hand-written rather than `rule.DecodeOptionsInto` because both options default to
// TRUE; see DecodeSelfClosingCompOptions.
func init() {
	rule.Register(rule.Registration{Rule: SelfClosingComp, Decode: DecodeSelfClosingCompOptions})
}
