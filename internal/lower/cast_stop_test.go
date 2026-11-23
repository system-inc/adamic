package lower

import (
	"errors"
	"strings"
	"testing"
)

// castStopProbes is one program per castCause a small program reaches, each with the words its stop says in
// .ts and whether that stop is a refusal the ledger makes or a check nobody has built (#b5w3ycg, Nov 22).
var castStopProbes = []struct {
	name, source string
	refused      bool
	words        string
}{
	{"double", "function f(n: number): string { return (n as unknown) as string; }", true,
		"a cast the runtime can't check: an `as unknown as` pair from number to string proves nothing about the value; write a checked view"},
	{"from unknown", "interface User { readonly name: string } function read(x: unknown): User { return x as User; }", true,
		"a cast the runtime can't check: a cast from unknown to User; narrow it first"},
	{"primitive", "function f(x: number | string | boolean): number { return x as number; }", true,
		"a cast the runtime can't check: a cast from string | number | boolean to number, which holds a primitive no view can check; compare the primitive's typeof first"},
	{"class arguments", "class Base {} class Box<T> extends Base { readonly value: T; constructor(value: T) { super(); this.value = value; } } function f(x: Base): Box<number> { return x as Box<number>; }", true,
		"a cast the runtime can't check: a downcast from Base to Box<number> whose type arguments Base's class ancestry doesn't fix"},
	{"class ancestry", "class A { readonly n: number = 1; } class B { readonly n: number = 1; readonly m: number = 2; } function f(x: A): B { return x as B; }", false,
		"stage 0 can't lower a checked class cast from A to B by instanceof yet"},
	{"class target", "class A { readonly n = 1; } function f(x: { readonly n: number }): A { return x as A; }", false,
		"stage 0 can't lower a checked class cast from { readonly n: number; } to A by instanceof yet"},
	{"refinement", "type A = { readonly kind: 'a'; readonly n: number }; type B = { readonly kind: 'b'; readonly n: number }; function f(x: A | B): { readonly kind: 'a'; readonly n: 1 } { return x as { readonly kind: 'a'; readonly n: 1 }; }", true,
		"which is narrower than a member of A | B without being one"},
	{"union write", "type A = { kind: 'a'; readonly n: number }; type B = { kind: 'b'; readonly n: number }; function f(x: A | B): A { return x as A; }", true,
		"a cast the runtime can't check: a cast over A | B, whose writable field kind some member can't hold every value of"},
	{"ambiguous tag", "type A = { readonly kind: 'a'; readonly n: number }; type B = { readonly kind: 'a'; readonly s: string }; function f(x: A | B): A { return x as A; }", true,
		"whose tag field kind holds a literal in every member but doesn't tell each apart"},
	{"callable", "function cast(x: (n: number | string) => number | string): (n: number) => number { return x as (n: number) => number; }", true,
		"which asserts a call signature; prove the callable body"},
	{"index signature", "function f(x: object): Record<string, number> { return x as Record<string, number>; }", false,
		"stage 0 can't lower a checked dictionary view of Record<string, number> over object yet"},
	{"intersection", "interface A { readonly a: number } interface B { readonly b: number } function f(x: A): A & B { return x as A & B; }", false,
		"stage 0 can't lower a checked view of the intersection A & B over A yet"},
	{"intersection with a callable", "interface A { readonly a: number } function f(x: A): A & (() => void) { return x as A & (() => void); }", true,
		"a cast the runtime can't check: a cast from A to () => void, which asserts a call signature; prove the callable body"},
	{"constrained type parameter", "interface A { readonly a: number } function f<T extends { readonly n: number }>(x: T): A & T { return x as A & T; }", false,
		"stage 0 can't lower a checked view of A & T over the type parameter T, constrained to { readonly n: number; } yet"},
	{"unconstrained type parameter", "interface A { readonly a: number } function f<T>(x: T): A & T { return x as A & T; }", true,
		"a cast the runtime can't check: a cast from the type parameter T to A & T, which nothing at run time can test, since T may be any type; constrain T to an object type"},
	{"untagged union", "interface A { readonly a: number } interface B { readonly b: string } function f(x: A | B): A { return x as A; }", false,
		"stage 0 can't lower a checked view of A over the union A | B, whose members no literal tag tells apart yet"},
	{"union target", "interface Base { readonly flags: number } interface A extends Base { readonly a: number } interface B extends Base { readonly b: string } function f(x: Base): A | B { return x as A | B; }", false,
		"stage 0 can't lower a checked view of the union A | B over Base, whose members no literal tag tells apart yet"},
	// A cast that keeps the source's own undefined lowers through its present road (cast_keeps_nullish_test.go);
	// a source's null kept as undefined is past it, and still stops with this cause.
	{"keeps undefined", "interface Base { readonly n: number } interface Sub extends Base { readonly m: number } function f(x: Base | null): Sub | undefined { return x as Sub | undefined; }", false,
		"stage 0 can't lower a checked cast from Base | null to Sub | undefined, which keeps undefined yet"},
	{"from object", "interface B { readonly b: number } function f(x: object): B { return x as B; }", false,
		"stage 0 can't lower a checked view of B over a value typed object, which names no fields yet"},
	// A cast to exactly the source without plain undefined or null is presentCast's road and lowers
	// (cast_drops_nullish_test.go); one whose source holds void is past it, and still stops with this cause.
	{"removes undefined", "interface S { readonly n: number } function f(x: S | void): S { return x as S; }", false,
		"stage 0 can't lower a checked cast from void | S to S, which removes undefined yet"},
	{"mixed source", "interface O { readonly n: number } function f(x: O | string): O { return x as O; }", false,
		"stage 0 can't lower a checked cast from string | O, which holds primitives, to O, which needs a typeof test before its view yet"},
}

// TestCastStopNamesItsCause holds each cause's words and kind in .ts, and holds .a to main's one sentence,
// since item 8 refuses a checked downcast there whatever stopped it.
func TestCastStopNamesItsCause(t *testing.T) {
	t.Parallel()
	for _, probe := range castStopProbes {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, probe.source, ".ts")
			var refused *Refused
			var notYet *NotYet
			switch {
			case err == nil:
				t.Fatalf(".ts: lowered, want a stop saying %q", probe.words)
			case probe.refused && !errors.As(err, &refused), !probe.refused && !errors.As(err, &notYet):
				t.Fatalf(".ts: want refused=%v, got %T: %v", probe.refused, err, err)
			case !strings.Contains(err.Error(), probe.words):
				t.Fatalf(".ts: want %q in the stop, got %v", probe.words, err)
			}
			_, err = lowerSourceAs(t, probe.source, ".a")
			if !errors.As(err, &refused) || !strings.Contains(err.Error(), "Adamic 0.1 refuses a cast the runtime can't check; ") {
				t.Fatalf(".a: want main's refusal word for word, got %v", err)
			}
		})
	}
}
