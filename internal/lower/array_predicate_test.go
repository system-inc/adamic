package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestArrayPredicatePreservesDeclaredElementContract(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`function guard(value: string | number[]): value is number[] { return Array.isArray(value); }`,
		`function guard(value: string | readonly number[]): value is readonly number[] { if (Array.isArray(value)) { return true; } return false; }`,
	} {
		if _, err := lowerSource(t, source); err != nil {
			t.Fatalf("%s: %v", source, err)
		}
	}
}

func TestArrayPredicateCannotInventAnElementContract(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`function guard(value: string | (number | string)[]): value is number[] { return Array.isArray(value); }`,
		`function guard(value: any): value is readonly unknown[] { return Array.isArray(value); }`,
		`function guard(value: {} | number[]): value is number[] { return Array.isArray(value); }`,
		`function guard(value: string | number[]): value is number[] { const Array = { isArray(value: string | number[]): boolean { return true; } }; return Array.isArray(value); }`,
	} {
		_, err := lowerSource(t, source)
		var refusal *Refused
		if !errors.As(err, &refusal) || !strings.Contains(err.Error(), "return is not proven") {
			t.Fatalf("%s: want unproven predicate refusal, got %v", source, err)
		}
	}
}

func TestArrayPredicateDoesNotMisclassifyNativeTuples(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`const tuple: [number, string] = [1, 'x']; console.log(String(Array.isArray(tuple)));`,
		`function test(value: {}): boolean { return Array.isArray(value); }`,
	} {
		_, err := lowerSource(t, source)
		var notYet *NotYet
		if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "tuple or an erased") {
			t.Fatalf("%s: want an explicit representation refusal, got %v", source, err)
		}
	}
}
