package react

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers this package's `forbid-elements` rule.
//
// A file of its own rather than a line in `register.go`, so a port is three owned files and no
// shared edit.
//
// The decoder is hand-rolled rather than `rule.DecodeOptionsInto`, because the schema's list items
// are `anyOf: [string, object]` and Go has no single shape for that union. It also has to answer an
// empty body with an empty list rather than an error: upstream reads `configuration.forbid || []`,
// so an unconfigured rule declines everything instead of failing.
func init() {
	rule.Register(rule.Registration{
		Rule:   ForbidElements,
		Decode: DecodeForbidElementsOptions,
	})
}
