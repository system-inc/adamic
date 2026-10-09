package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers this package's `dot-notation` rule.
//
// A file of its own rather than a line in `register.go`, so a port is three owned files and no
// shared edit. The decoder is hand-written because `allowKeywords` defaults to TRUE and the generic
// helper would decode an absent option to false, inverting the rule; see DecodeDotNotationOptions.
func init() {
	rule.Register(rule.Registration{Rule: DotNotation, Decode: DecodeDotNotationOptions})
}
