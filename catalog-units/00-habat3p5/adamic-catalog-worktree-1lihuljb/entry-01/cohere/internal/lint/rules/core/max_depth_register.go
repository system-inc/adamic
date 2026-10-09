package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers max-depth.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors
// landing at once touch nothing in common.
//
// The decoder is hand rolled because upstream's schema is a `oneOf` over a bare integer and an
// object carrying `maximum` or `max`, which no single Go struct decodes.
//
// `RequiresOptions` is deliberately NOT set: upstream's default is a depth of 4, so an unconfigured
// rule enforces that rather than doing nothing.
func init() {
	rule.Register(rule.Registration{
		Rule:   MaxDepth,
		Decode: DecodeMaxDepthOptions,
	})
}
