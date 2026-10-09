package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers object-shorthand.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors
// landing at once touch nothing in common.
//
// DecodeOptionList rather than Decode, because upstream's schema is an `anyOf` over three list
// shapes, `["error", "always", {"avoidQuotes": true}]` among them. The decoder is hand rolled
// because of those shapes and because its default mode is "always", so a zero-value struct would
// carry an empty mode matching none of the six and silently report nothing.
func init() {
	rule.Register(rule.Registration{
		Rule:             ObjectShorthand,
		DecodeOptionList: DecodeObjectShorthandOptions,
	})
}
