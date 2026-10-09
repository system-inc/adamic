package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers strict.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common.
//
// The decoder is hand-rolled so an unknown mode fails loudly. Every arm of this rule is selected by
// string equality, so `rule.DecodeOptionsInto` leaving an unrecognized string in the field would
// pick a silent fifth behaviour of reporting nothing rather than producing an error.
func init() {
	rule.Register(rule.Registration{
		Rule:   Strict,
		Decode: DecodeStrictOptions,
	})
}
