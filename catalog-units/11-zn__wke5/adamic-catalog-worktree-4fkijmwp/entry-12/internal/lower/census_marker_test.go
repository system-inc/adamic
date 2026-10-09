package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestCensusMarkerResultIsAssignable(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `function value(): number { return 1; }
const marker: (...args: never[]) => void = value;
console.log(typeof marker);`)
	var refused *Refused
	if !errors.As(err, &refused) {
		t.Fatalf("expected result relation refusal, got %v", err)
	}
}

func TestCensusMarkerZeroCallIsNotAssumedSafe(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `function required(value: number): void { console.log(String(value === undefined)); }
const marker: (...args: never[]) => void = required;
marker();`)
	var notYet *NotYet
	if !errors.As(err, &notYet) || !strings.Contains(notYet.What, "call through an erased never-rest callable marker") {
		t.Fatalf("expected explicit erased-call boundary, got %v", err)
	}
}
