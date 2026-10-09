package lower

import (
	"errors"
	"testing"
)

func TestCensusPredicateMarkerKeepsProofBoundaries(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`function isText(value: string | undefined): value is string { return true; }
const marker: (...args: never[]) => void = isText; console.log(typeof marker);`,
		`function describe(test: (value: string | number) => value is string): boolean { return test(1); }`,
		`function describe(test: (value: string | number) => value is string): typeof test { return test; }`,
		`function hold(marker: (...args: number[]) => void): void {} function describe(test: (value: number | string) => value is string): void { hold(test); }`,
	} {
		_, err := lowerSource(t, source)
		var refused *Refused
		if !errors.As(err, &refused) {
			t.Fatalf("unobservable marker exception escaped its boundary: %v", err)
		}
	}
}

func TestCensusPredicateMarkerCallStaysNotYet(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `function isText(value: string | number): value is string { return typeof value === 'string'; }
const marker: (...args: never[]) => void = isText;
marker();`)
	var notYet *NotYet
	if !errors.As(err, &notYet) {
		t.Fatalf("expected erased marker call refusal, got %v", err)
	}
}
