package react

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers this package's `forbid-foreign-prop-types` rule.
//
// A file of its own rather than a line in `register.go`, so a port is three owned files and no
// shared edit.
//
// The generic decoder is safe for this option surface, which is one boolean defaulting to FALSE.
// A rule configured as a bare severity is handed nil options, `DecodeOptionsInto` errors on empty
// input, the config layer turns that into nil for a non-required rule, and `options.(T)` on nil
// yields the zero struct. Here the zero struct IS upstream's default, so the nil path and the
// explicit `{"allowInPropTypes": false}` path agree. A fixture bypasses the decoder to pin that,
// because a default-true option decoded this way would silently invert the rule and every fixture
// built from a struct would pass anyway.
func init() {
	rule.Register(rule.Registration{
		Rule:   ForbidForeignPropTypes,
		Decode: rule.DecodeOptionsInto[ForbidForeignPropTypesOptions](),
	})
}
