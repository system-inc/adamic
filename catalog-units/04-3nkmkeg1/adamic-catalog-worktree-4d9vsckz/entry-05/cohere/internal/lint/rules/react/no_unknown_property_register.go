package react

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers this package's `no-unknown-property` rule.
//
// A file of its own rather than a line in `register.go`, so a port is three owned files and no
// shared edit.
//
// The decoder is hand-rolled rather than `rule.DecodeOptionsInto` so an empty body answers with
// upstream's defaults instead of an error. Both defaults happen to match Go's zero values here, and
// the decoder exists to make that explicit rather than accidental.
func init() {
	rule.Register(rule.Registration{
		Rule:   NoUnknownProperty,
		Decode: DecodeNoUnknownPropertyOptions,
	})
}
