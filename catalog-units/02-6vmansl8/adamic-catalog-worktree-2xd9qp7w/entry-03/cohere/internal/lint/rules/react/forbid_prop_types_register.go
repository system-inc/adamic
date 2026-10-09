package react

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers this package's `forbid-prop-types` rule.
//
// The decoder is hand-written rather than `rule.DecodeOptionsInto` because `forbid` has to be
// distinguishable as absent; see DecodeForbidPropTypesOptions.
func init() {
	rule.Register(rule.Registration{Rule: ForbidPropTypes, Decode: DecodeForbidPropTypesOptions})
}
