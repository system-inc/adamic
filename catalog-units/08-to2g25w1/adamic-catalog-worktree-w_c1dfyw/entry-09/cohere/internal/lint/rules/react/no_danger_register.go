package react

import "github.com/system-inc/cohere/internal/lint/rule"

// Registered with a hand-written decoder rather than `rule.DecodeOptionsInto`.
//
// The generic helper errors on empty input, and a rule configured as a bare severity is handed nil.
// For a non-required rule that error becomes nil and the assertion yields the zero value, which
// happens to be the right answer here, so the generic decoder would work by accident. Answering the
// documented default explicitly makes the unconfigured path a decision instead.
func init() {
	rule.Register(rule.Registration{Rule: NoDanger, Decode: DecodeNoDangerOptions})
}
