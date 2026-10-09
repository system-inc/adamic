package core

import "github.com/system-inc/cohere/internal/lint/rule"

// The decoder is hand-rolled rather than `rule.DecodeOptionsInto`, and that is the whole reason this
// registration is worth reading. The generic helper errors on empty input, the config layer turns
// that error into nil for a non-required rule, and a nil handed to a rule reading a bool field yields
// `false` -- which for this rule's only option is the inverted default. `DecodeNoSequencesOptions`
// returns the default on empty input instead, and the field behind it is a pointer so an absent key
// stays distinguishable from an explicit `false`.
func init() {
	rule.Register(rule.Registration{
		Rule:   NoSequences,
		Decode: DecodeNoSequencesOptions,
	})
}
