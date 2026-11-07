package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestCensusMarkerResultIsAssignable(test *testing.T) {
	test.Parallel()
	_, err := lowerSource(test, `function value(): number { return 1; }
const marker: (...args: never[]) => void = value;
console.log(typeof marker);`)
	var refused *Refused
	if !errors.As(err, &refused) {
		test.Fatalf("expected result relation refusal, got %v", err)
	}
}

func TestCensusMarkerZeroCallIsNotAssumedSafe(test *testing.T) {
	test.Parallel()
	_, err := lowerSource(test, `function required(value: number): void { console.log(String(value === undefined)); }
const marker: (...args: never[]) => void = required;
marker();`)
	var notYet *NotYet
	if !errors.As(err, &notYet) || !strings.Contains(notYet.What, "call through an erased never-rest callable marker") {
		test.Fatalf("expected explicit erased-call boundary, got %v", err)
	}
}
