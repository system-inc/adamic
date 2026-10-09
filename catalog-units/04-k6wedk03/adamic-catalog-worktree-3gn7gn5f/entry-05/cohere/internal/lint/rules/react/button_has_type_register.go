package react

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers this package's `button-has-type` rule.
//
// A file of its own rather than a line in `register.go`, so a port is three owned files and no
// shared edit. The decoder is hand-written rather than `rule.DecodeOptionsInto` because all three
// options default to TRUE, which is not Go's zero value; see DecodeButtonHasTypeOptions.
func init() {
	rule.Register(rule.Registration{Rule: ButtonHasType, Decode: DecodeButtonHasTypeOptions})
}
