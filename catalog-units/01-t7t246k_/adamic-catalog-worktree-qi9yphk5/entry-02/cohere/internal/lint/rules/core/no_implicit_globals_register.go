package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers no-implicit-globals.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common.
//
// The decoder is hand-rolled so an explicit `false` stays distinguishable from an absent key. That
// costs nothing today, since upstream's default for `lexicalBindings` is already false, and it is
// what keeps the shape correct if the default ever moves.
func init() {
	rule.Register(rule.Registration{
		Rule:   NoImplicitGlobals,
		Decode: DecodeNoImplicitGlobalsOptions,
	})
}
