package lower

import (
	"errors"
	"testing"
)

func TestViewPreflightPreservesOtherRefusals(t *testing.T) {
	for _, source := range []string{
		"interface User { readonly name: string } function read(x: unknown): User { return x as User; }",
		"const n = ('wrong' as unknown) as number;",
	} {
		_, err := lowerSource(t, source)
		var refused *Refused
		if !errors.As(err, &refused) {
			t.Fatalf("want refusal, got %v", err)
		}
		want := "a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)"
		if got := refused.What + "; " + refused.Fix; got != want {
			t.Fatalf("changed diagnostic: %s", got)
		}
	}
}

func TestViewArrayConsumerRefusals(t *testing.T) {
	for _, source := range []string{
		`const empty: never[] = [] as never[]; function mutable<T>(flag: boolean, values: T[]): T[] { return flag ? empty : values; }`,
		`const empty: never[] = [] as never[]; function mutable<T>(flag: boolean, values: T[]): T[] { return (flag ? empty : values) as readonly T[] as T[]; }`,
	} {
		_, err := lowerSource(t, source)
		var refused *Refused
		if !errors.As(err, &refused) {
			t.Fatalf("want mutable target refusal, got %v", err)
		}
		want := "a value of type never seen as T, a type parameter whose constraint any can be written, so it can write what never can't hold; take it as never, or constrain T to something readonly, which can't write (adamic/invariant-mutable)"
		if got := refused.What + "; " + refused.Fix; got != want {
			t.Fatalf("changed mutable target diagnostic: %s", got)
		}
	}
	_, err := lowerSource(t, `function nonnullable<T, U>(array: readonly T[] | undefined): readonly U[] { return array as unknown[] as U[]; }`)
	var refused *Refused
	if !errors.As(err, &refused) || refused.What != "a cast the runtime can't check" || refused.Fix != castRepair {
		t.Fatalf("want nonnullable target refusal, got %v", err)
	}
}

func TestViewArrayOptionalComparatorRefusal(t *testing.T) {
	_, err := lowerSource(t, `function sorted(values: (string | undefined)[], comparer?: (a: string | undefined, b: string | undefined) => number): (string | undefined)[] { return values.sort(comparer); }`)
	var unsupported *NotYet
	if !errors.As(err, &unsupported) || unsupported.What != "an optional comparator over undefined reference elements" {
		t.Fatalf("want undefined-reference comparator stop, got %v", err)
	}
}
