package react

import "github.com/system-inc/cohere/internal/lint/rule"

// Registered with a hand-written decoder rather than `rule.DecodeOptionsInto`.
//
// Two of the five options have a default that is not Go's zero value: `enforceDynamicLinks` is a
// string enum whose absent value is `always`, and `links` defaults to true. The generic decoder
// also errors on empty input, where this rule must answer upstream's defaults instead.
func init() {
	rule.Register(rule.Registration{Rule: JsxNoTargetBlank, Decode: DecodeJsxNoTargetBlankOptions})
}
