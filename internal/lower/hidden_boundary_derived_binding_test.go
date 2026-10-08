package lower

import (
	"errors"
	"testing"
)

func TestHiddenDerivedSelfInitializerStaysNotYet(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `function enclosing(): () => number { const derived: () => number = () => derived(); return derived; } enclosing()();`)
	var notYet *NotYet
	if !errors.As(err, &notYet) || notYet.What != "a function value that captures the variable its own initializer declares" {
		t.Fatalf("got %v, want self-initializer capture refusal", err)
	}
}
