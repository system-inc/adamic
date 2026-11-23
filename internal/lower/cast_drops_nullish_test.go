package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// A .ts cast whose target is exactly its source without undefined, null or both (presentCast, tsc's
// `args[0] as Expression` of an `Expression | undefined`, #b5w3ycg) lowers to the check a non-null assertion
// makes: one Coalesce that panics, named for the cast, counted as nonNullValue counts every check it makes, so the
// cast and `!` share one presence check rather than each having its own.
// Mutant: the drops-nullish road removed (the cast stops as before), presentCast answering false, and every row
// here stops with castDropsNullish's NotYet.
func TestDropsNullishCastLowersAsPresenceCheck(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"undefined", "interface S { readonly n: number } function f(x: S | undefined): S { return x as S; }"},
		{"null", "interface S { readonly n: number } function f(x: S | null): S { return x as S; }"},
		{"undefined and null", "interface S { readonly n: number } function f(x: S | null | undefined): S { return x as S; }"},
		{"a call's result", "interface S { readonly n: number } function find(x: S, found: boolean): S | undefined { return found ? x : undefined; } function f(x: S): number { return (find(x, true) as S).n; }"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, err := lowerSourceAs(t, probe.source, ".ts")
			if err != nil {
				t.Fatal(err)
			}
			checks := 0
			for _, function := range program.Functions {
				walk(function.Body, func(node any) bool {
					present, ok := node.(ir.Coalesce)
					if !ok || present.Panic == nil {
						return true
					}
					checks++
					if present.Fallback != nil {
						t.Fatalf("the check is %#v, want a Coalesce that panics", present)
					}
					if text := program.Strings[present.Panic.(ir.StringConstant).Index]; !strings.HasPrefix(text, "cast failed: this ") || !strings.HasSuffix(text, " is not a S") {
						t.Fatalf("message %q, want the cast named as written", text)
					}
					return true
				})
			}
			if checks != 1 || program.NonNullChecks.Checked != 1 || program.NonNullChecks.Proven != 0 {
				t.Fatalf("got %d presence checks and %+v, want one checked site", checks, program.NonNullChecks)
			}
		})
	}
}

// What presentCast leaves as castProof judged it: a source holding undefined and null cast to a target that keeps
// one of them (the target must be the full non-nullable type), stopping with its current words, and the honest
// cast in .a, which keeps main's refusal word for word, as `!` is refused there.
// Mutant: the drops-nullish road removed (the cast stops as before) leaves this test green, and
// TestDropsNullishCastLowersAsPresenceCheck red. presentCast's .a test dropped (checkedAssertionSource) reds the .a
// row. The kept rows are held twice, by the identity test and by castDropsNullish, so neither dropped alone reds them.
func TestDropsNullishCastOutsideItsShapeStops(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, extension, words string }{
		{"keeps null", "interface S { readonly n: number } function f(x: S | null | undefined): S | null { return x as S | null; }", ".ts",
			"to S | null, which keeps null yet"},
		{"keeps undefined", "interface S { readonly n: number } function f(x: S | null | undefined): S | undefined { return x as S | undefined; }", ".ts",
			"to S | undefined, which keeps undefined yet"},
		{"in .a", "interface S { readonly n: number } function f(x: S | undefined): S { return x as S; }", ".a",
			"Adamic 0.1 refuses a cast the runtime can't check; "},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, probe.source, probe.extension)
			var refused *Refused
			var notYet *NotYet
			switch {
			case err == nil:
				t.Fatalf("lowered, want a stop saying %q", probe.words)
			case probe.extension == ".a" && !errors.As(err, &refused), probe.extension == ".ts" && !errors.As(err, &notYet):
				t.Fatalf("got %T: %v", err, err)
			case !strings.Contains(err.Error(), probe.words):
				t.Fatalf("want %q in the stop, got %v", probe.words, err)
			}
		})
	}
}
