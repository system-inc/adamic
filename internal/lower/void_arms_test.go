package lower

import (
	"errors"
	"strings"
	"testing"
)

// A union source member is related only to the target arms it isn't discriminated apart from
// (reachableArms, #wcva20f and #54c1r7k, Compiler's disjointness rule as Adamic ruled it on Nov 5).
// Mutants: reachableArms returning every arm turns the first test red (each program is refused again
// at its initializer), and discriminatedApart replaced by tsc's assignability turns the second test's
// overlapping and undiscriminated rows red (each is admitted).

// An arm a discriminant keeps a value out of is never related to it: kind "b" is never read through
// kind "a", so its void-returning run isn't seen as one returning number there. A void field would be
// the void container rule's to judge (TestVoidFieldHoldersAreRead), so the pin carries the void in a
// function's result, as an iterator's next() does. A generator's IteratorReturnResult<void> (done: true)
// against IteratorYieldResult<number> (done?: false) is the same pairing, and is no longer this rule's
// stop.
func TestVoidWidenedSkipsArmsADiscriminantRulesOut(t *testing.T) {
	t.Parallel()
	_, err := lowerSourceAs(t, `function* g(): Generator<number, void, unknown> { yield 1; yield 2; }
const values: Iterable<number, void> = g();
console.log([...values].join(","));
`, ".ts")
	if err != nil && strings.Contains(err.Error(), "no-void-widening") {
		t.Errorf("Iterable<number, void> taking a function*'s call: want no void-widening stop, got %v", err)
	}
	const tagged = `interface A { readonly kind: "a"; readonly run: () => number; }
interface B { readonly kind: "b"; readonly run: () => void; }
function noop(): void { }
const b: B = { kind: "b", run: noop };
const either: A | B = b;
console.log(either.kind);
`
	for _, extension := range []string{".a", ".ts"} {
		lowersAndAgreesWithNodeAs(t, tagged, extension)
	}
}

// Every pairing no discriminant rules out is still related, so each of these is refused at its
// initializer: an arm whose discriminant overlaps the source's, an arm no discriminant separates from
// it (a program can select it with a predicate or an in test), a void function seen through a union of
// a callable and a string, and a generator's void TReturn seen as unknown, which a program can read.
func TestVoidWidenedKeepsArmsNoDiscriminantRulesOut(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, want string }{
		{"overlapping discriminant", `interface A { readonly kind: "a" | "b"; readonly value: number; }
interface B { readonly kind: "b"; readonly value: void; }
const b: B = { kind: "b", value: undefined };
const either: A | B = b;
console.log(either.kind);`, "a void value at the initializer site through .value seen as number"},
		{"no discriminant", `interface N { readonly n: number; readonly value: number; }
interface V { readonly tag: string; readonly value: void; }
const v: V = { tag: "t", value: undefined };
const either: N | V = v;
function isN(x: N | V): x is N { return (x as N).n !== undefined; }
console.log(String(isN(either)));`, "a void value at the initializer site through .value seen as number"},
		{"void function through a union", `function make(): void { }
const f: (() => unknown) | string = make;
console.log(typeof f);`, "a void-returning function at the initializer site"},
		{"void TReturn seen as unknown", `function* g(): Generator<number, void, unknown> { yield 1; }
const it: Iterator<number, unknown> = g();
console.log(typeof it);`, "seen as unknown"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, probe.source+"\n", ".ts")
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(refused.What, probe.want) {
				t.Fatalf("want Refused with %q, got %v", probe.want, err)
			}
		})
	}
}
