package react

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers this package's `prefer-es6-class` rule.
//
// A file of its own rather than a line in `register.go`, so a port is three owned files and no
// shared edit. The decoder is hand-written rather than `rule.DecodeOptionsInto` because the option
// is a bare enum string rather than an object; see DecodePreferEs6ClassOptions.
func init() {
	rule.Register(rule.Registration{Rule: PreferEs6Class, Decode: DecodePreferEs6ClassOptions})
}
