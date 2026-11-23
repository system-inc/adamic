package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// A member cast whose source holds undefined beside a represented union (#b5w3ycg slice 2, absentArm) lowers
// in .ts to the absent check, then the tag check: one CheckedCast whose value is the Coalesce that panics with
// the cast's own message.
func TestAbsentArmCastsLower(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, message string }{
		{"member", "enum K { A = 1, B = 2 } type A = { readonly kind: K.A; readonly n: number }; type B = { readonly kind: K.B; readonly n: number }; function f(x: A | B | undefined): A { return x as A; }", "cast failed: this A | B | undefined is not a A"},
		{"sub-union", "enum K { A = 1, B = 2, C = 3 } type A = { readonly kind: K.A }; type B = { readonly kind: K.B }; type C = { readonly kind: K.C }; function f(x: A | B | C | undefined): A | C { return x as A | C; }", "cast failed: this A | B | C | undefined is not a A | C"},
		{"string tags", "type A = { readonly kind: 'a'; readonly n: number }; type B = { readonly kind: 'b'; readonly s: string }; function f(x: A | B | undefined): B { return x as B; }", "cast failed: this A | B | undefined is not a B"},
		{"optional parameter", "enum K { A = 1, B = 2 } type A = { readonly kind: K.A }; type B = { readonly kind: K.B }; function f(x?: A | B): A { return x as A; }", "cast failed: this A | B | undefined is not a A"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, err := lowerSourceAs(t, probe.source, ".ts")
			if err != nil {
				t.Fatal(err)
			}
			casts := 0
			for _, function := range program.Functions {
				walk(function.Body, func(node any) bool {
					cast, ok := node.(ir.CheckedCast)
					if !ok {
						return true
					}
					casts++
					absent, ok := cast.Value.(ir.Coalesce)
					if !ok || absent.Panic == nil || absent.Fallback != nil {
						t.Fatalf("the cast's value is %T, want the Coalesce that panics before the tag is read", cast.Value)
					}
					if text := program.Strings[absent.Panic.(ir.StringConstant).Index]; text != probe.message || cast.Message != probe.message {
						t.Fatalf("absent message %q, tag message %q, want both %q", text, cast.Message, probe.message)
					}
					return true
				})
			}
			if casts != 1 {
				t.Fatalf("got %d checked casts, want 1", casts)
			}
		})
	}
}

// What absentArm leaves as it was: a null beside the members, a single member beside undefined, a target that
// keeps undefined (CheckedCast can't pass an absent value through yet), a writable tag, and a tag the checker's
// duplicates can't tell apart.
func TestAbsentArmCastsStayRefused(t *testing.T) {
	t.Parallel()
	members := "enum K { A = 1, B = 2 } type A = { readonly kind: K.A }; type B = { readonly kind: K.B }; "
	for _, probe := range []struct{ name, source, extension string }{
		{"null beside", members + "function f(x: A | B | null): A { return x as A; }", ".ts"},
		{"null and undefined beside", members + "function f(x: A | B | null | undefined): A { return x as A; }", ".ts"},
		{"single member", members + "function f(x: A | undefined): A { return x as A; }", ".ts"},
		{"target keeps undefined", members + "function f(x: A | B | undefined): A | undefined { return x as A | undefined; }", ".ts"},
		{"writable tag", "enum K { A = 1, B = 2 } type A = { kind: K.A }; type B = { kind: K.B }; function f(x: A | B | undefined): A { return x as A; }", ".ts"},
		{"duplicate tags", "enum K { A = 1, B = 1 } type A = { readonly kind: K.A; readonly n: number }; type B = { readonly kind: K.B; readonly s: string }; function f(x: A | B | undefined): A { return x as A; }", ".ts"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, probe.source, probe.extension)
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(err.Error(), "adamic/no-unchecked-cast") {
				t.Fatalf("want an unchecked-cast refusal, got %v", err)
			}
		})
	}
}

// In .a the same cast is a member cast, refused as #m8dvvmf refuses every one, and the narrowing its fix offers
// reads the tag through ?., so that it type-checks on an operand that may be undefined and fails it too.
func TestAbsentArmCastInDotAIsRefusedWithItsNarrowing(t *testing.T) {
	t.Parallel()
	_, err := lowerSourceAs(t, "enum K { A = 1, B = 2 } type A = { readonly kind: K.A }; type B = { readonly kind: K.B }; function f(x: A | B | undefined): A { return x as A; }", ".a")
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), "in .a whose checks no proof erases") || !strings.Contains(err.Error(), "if (x?.kind !== K.A) panic('why')") {
		t.Fatalf("want .a's member-cast refusal offering x?.kind, got %v", err)
	}
}

// The absent field's premise (Language's v03 on #ha4hbqt): a tag read through a getter stops at the checked view
// before either test runs, with and without an undefined beside the members, so the absent test and the tag test
// can't disagree through one. If that rule lifts, this goes red and the premise needs reading again.
func TestAbsentArmGetterTagStopsFirst(t *testing.T) {
	t.Parallel()
	for _, source := range []string{"A | B | undefined", "A | B"} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			program := "enum K { A = 1, B = 2 } interface A { readonly kind: K.A; readonly n: number } interface B { readonly kind: K.B; readonly n: number } class Lying { get kind(): K.A { return K.A; } readonly n = 5; } function f(x: " + source + "): A { return x as A; } console.log(String(f(new Lying()).n));"
			_, err := lowerSourceAs(t, program, ".ts")
			if err == nil || !strings.Contains(err.Error(), "a getter in a checked field contract") {
				t.Fatalf("want the stop at a getter in a checked field contract, got %v", err)
			}
		})
	}
}
