package react

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers this package's `require-optimization` rule.
//
// A file of its own rather than a line in `register.go`, so a port is three owned files and no
// shared edit. The decoder is hand-written rather than `rule.DecodeOptionsInto` because the generic
// one errors on empty input, which is what a bare `"error"` configuration hands it.
func init() {
	rule.Register(rule.Registration{Rule: RequireOptimization, Decode: DecodeRequireOptimizationOptions})
}
