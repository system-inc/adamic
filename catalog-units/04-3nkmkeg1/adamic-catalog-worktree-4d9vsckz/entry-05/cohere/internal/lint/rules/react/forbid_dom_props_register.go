package react

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers this package's `forbid-dom-props` rule.
//
// A file of its own rather than a line in `register.go`, so a port is three owned files and no
// shared edit.
//
// The decoder is hand-rolled rather than `rule.DecodeOptionsInto`, because the schema's list items
// are a union of a string and an object, which Go has no single shape for, and because an absent
// `disallowedFor` and an empty one mean opposite things that a plain slice cannot separate. It also
// has to answer an empty body with an empty list rather than an error: an unconfigured rule
// declines everything, which is what the installed build does.
func init() {
	rule.Register(rule.Registration{
		Rule:   ForbidDomProps,
		Decode: DecodeForbidDomPropsOptions,
	})
}
