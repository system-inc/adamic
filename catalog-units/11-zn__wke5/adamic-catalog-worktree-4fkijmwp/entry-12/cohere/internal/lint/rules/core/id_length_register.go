package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers id-length.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors
// landing at once touch nothing in common.
//
// The decoder is hand rolled because two of upstream's defaults are non-zero -- a minimum of 2 and
// properties "always" -- so a zero-value struct would be a different rule rather than a weaker one.
//
// `RequiresOptions` is deliberately NOT set: upstream's default is a minimum length of 2, so an
// unconfigured rule enforces that rather than doing nothing.
func init() {
	rule.Register(rule.Registration{
		Rule:   IdLength,
		Decode: DecodeIdLengthOptions,
	})
}
