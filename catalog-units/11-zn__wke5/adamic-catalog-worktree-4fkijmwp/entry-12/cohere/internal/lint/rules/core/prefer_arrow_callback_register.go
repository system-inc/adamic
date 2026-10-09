package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers prefer-arrow-callback.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors
// landing at once touch nothing in common.
//
// The decoder is hand rolled because `allowUnboundThis` defaults to TRUE, so a zero-value struct
// would be a stricter rule rather than an unconfigured one.
func init() {
	rule.Register(rule.Registration{
		Rule:   PreferArrowCallback,
		Decode: DecodePreferArrowCallbackOptions,
	})
}
