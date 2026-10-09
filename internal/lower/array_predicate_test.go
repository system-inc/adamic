package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestArrayPredicatePreservesDeclaredElementContract(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`function guard(value: string | number[]): value is number[] { return Array.isArray(value); } console.log(String(guard([1, 2]))); console.log(String(guard('x')));`,
		`function guard(value: unknown): value is readonly unknown[] { return Array.isArray(value); } console.log(String(guard([1, 2]))); console.log(String(guard('x')));`,
		`function count(value: unknown): number { if (Array.isArray(value)) return value.length; return -1; } console.log(String(count([1, 2]))); console.log(String(count('x')));`,
		`const value: unknown = [1, 2]; if (Array.isArray(value)) console.log(String(value.length));`,
		`function guard(value: string | readonly number[]): value is readonly number[] { if (Array.isArray(value)) { return true; } return false; } console.log(String(guard([1, 2]))); console.log(String(guard('x')));`,
	} {
		lowersAndAgreesWithNode(t, source)
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

func TestArrayPredicateCoexistsWithUnknownReflection(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`function test(value: unknown): string { return typeof value; } console.log(test(1)); console.log(test('x'));`,
		`function test(value: unknown): boolean { return value === null; } console.log(String(test(null))); console.log(String(test(1)));`,
		`function test(value: unknown): unknown { return value; } console.log(String(test(7) === 7));`,
		`function errorCode(error: unknown): string | undefined { return typeof error === 'object' && error !== null && 'code' in error && typeof error.code === 'string' ? error.code : undefined; } console.log(errorCode({ code: 'ok' }) ?? 'missing'); console.log(errorCode(null) ?? 'missing');`,
		`function isArray(value: unknown): value is readonly unknown[] { return Array.isArray(value); } function length(value: unknown): number { if (isArray(value)) return value.length; return -1; } console.log(String(length([1, 2]))); console.log(String(length('x')));`,
	} {
		lowersAndAgreesWithNode(t, source)
	}
}
