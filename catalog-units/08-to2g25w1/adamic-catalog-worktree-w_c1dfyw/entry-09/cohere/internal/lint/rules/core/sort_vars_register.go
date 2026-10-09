package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers sort-vars.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors
// landing at once touch nothing in common.
//
// The decoder is hand rolled only because `rule.DecodeOptionsInto` errors on empty input, and a
// rule configured as a bare "error" is handed nil: upstream's default is a working configuration
// rather than a disabled one, so nil has to decode rather than fail.
//
// `RequiresOptions` is deliberately NOT set: the one option defaults false and the rule enforces
// sorting without it.
func init() {
	rule.Register(rule.Registration{
		Rule:   SortVars,
		Decode: DecodeSortVarsOptions,
	})
}
