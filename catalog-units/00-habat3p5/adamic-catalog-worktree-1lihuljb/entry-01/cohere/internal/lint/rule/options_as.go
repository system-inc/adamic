package rule

import (
	"fmt"
	"sync/atomic"
)

// OptionsAsReads counts every OptionsAs call, so the registry's decoder agreement test can tell a rule
// whose options read was reached from one that ran without reaching it (#zwd43jn). One atomic add per
// options read, contended across the parallel file walk. Measured on ahra's full lint, five runs each:
// a median lint phase of 3.88s with the counter and 4.07s without, against a run-to-run spread of 3.4s
// to 5.8s, so its cost is below what a run resolves rather than shown to be zero.
var OptionsAsReads atomic.Int64

// OptionsAs reads the options a rule's Run is handed as the type its registration's decoder
// returns, in the same two-value shape as the comma-ok assertion it replaces.
//
//	settings, ok := rule.OptionsAs[NoConcatenatedClassesOptions](options)
//	if !ok {
//		settings = defaultNoConcatenatedClassesSettings()
//	}
//
// Nil is the one honest miss: a rule configured as a bare `"error"` is handed no options, so it
// gets the zero value and false, and the rule's own default arm runs exactly as it did before.
//
// Any other mismatch panics, because it can only mean the decoder and the assertion disagree. Run
// takes `options any`, so the compiler never compares the type a Decode returns with the type a rule
// asserts, and a comma-ok assertion cannot fail loudly: `unified-signatures` asserted a pointer over
// the value `DecodeOptionsInto` returns, the assertion missed on every configured run, and both of
// its options read false whatever the config said, with every fixture green because the fixtures
// handed Run the pointer directly (#qmvkf83). A panic turns that into a crashed file in the run and a
// failing test for any fixture that configures the rule.
//
// The message names both package-qualified types, which together name the rule.
func OptionsAs[Options any](options any) (Options, bool) {
	OptionsAsReads.Add(1)
	var zero Options
	if options == nil {
		return zero, false
	}
	typed, matches := options.(Options)
	if !matches {
		panic(fmt.Sprintf("the rule reads its options as %T, but its decoder handed it %T: the "+
			"registration's Decode and the rule's OptionsAs must name the same type, or every "+
			"configured option is ignored", zero, options))
	}
	return typed, true
}
