package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers eqeqeq.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common.
//
// DecodeOptionList rather than Decode, because upstream's option surface is a two-element list:
// `["error", "always", {"null": "ignore"}]`. The decoder is hand-rolled so an unknown mode or null
// policy fails loudly. Every arm of this rule
// is selected by string equality, so `rule.DecodeOptionsInto` leaving an unrecognized string in the
// field would pick a silent fourth behaviour of reporting nothing rather than producing an error.
// The same function is also where upstream's mode-dependent default lives: the null policy defaults
// to Always under the Always mode and is forced to Ignore under Smart.
func init() {
	rule.Register(rule.Registration{
		Rule:             Eqeqeq,
		DecodeOptionList: DecodeEqeqeqOptions,
	})
}
