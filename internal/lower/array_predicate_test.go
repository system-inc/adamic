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
		`function guard(value: unknown): value is readonly unknown[] { return Array.isArray(value); }`,
		`function count(value: unknown): number { if (Array.isArray(value)) return value.length; return -1; }`,
		`const value: unknown = [1, 2]; if (Array.isArray(value)) console.log(String(value.length));`,
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
		`function guard(value: unknown): value is readonly number[] { return Array.isArray(value); }`,
		`function guard(value: unknown): value is unknown[] { return Array.isArray(value); }`,
		`function guard(value: unknown): value is readonly unknown[] { return true; }`,
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

func TestUnknownArrayPredicateRefusesUnrepresentedObservations(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`function test(value: unknown): string { return typeof value; }`,
		`function test(value: unknown): boolean { return value === null; }`,
		`function test(value: unknown): unknown { return value; }`,
		`function test(value: unknown): unknown { if (Array.isArray(value)) return value[0]; return undefined; }`,
		`function test(value: unknown): boolean { return Array.isArray(value); } const pair: [number, string] = [1, 'x']; console.log(String(test(pair)));`,
	} {
		_, err := lowerSource(t, source)
		var notYet *NotYet
		if !errors.As(err, &notYet) {
			t.Fatalf("%s: want named representation refusal, got %v", source, err)
		}
	}
}
