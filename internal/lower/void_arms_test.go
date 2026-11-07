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
func TestVoidWidenedSkipsArmsADiscriminantRulesOut(test *testing.T) {
	test.Parallel()
	_, err := lowerSourceAs(test, `function* g(): Generator<number, void, unknown> { yield 1; yield 2; }
const values: Iterable<number, void> = g();
console.log([...values].join(","));
`, ".ts")
	if err != nil && strings.Contains(err.Error(), "no-void-widening") {
		test.Errorf("Iterable<number, void> taking a function*'s call: want no void-widening stop, got %v", err)
	}
	const tagged = `interface A { readonly kind: "a"; readonly run: () => number; }
interface B { readonly kind: "b"; readonly run: () => void; }
function noop(): void { }
const b: B = { kind: "b", run: noop };
const either: A | B = b;
console.log(either.kind);
`
	for _, extension := range []string{".a", ".ts"} {
		lowersAndAgreesWithNodeAs(test, tagged, extension)
	}
}

// Every pairing no discriminant rules out is still related, so each of these is refused at its
// initializer: an arm whose discriminant overlaps the source's, an arm no discriminant separates from
// it (a program can select it with a predicate or an in test), a void function seen through a union of
// a callable and a string, and a generator's void TReturn seen as unknown, which a program can read.
func TestVoidWidenedKeepsArmsNoDiscriminantRulesOut(test *testing.T) {
	test.Parallel()
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
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			_, err := lowerSourceAs(test, probe.source+"\n", ".ts")
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(refused.What, probe.want) {
				test.Fatalf("want Refused with %q, got %v", probe.want, err)
			}
		})
	}
}

// A numeric enum is open (rulings.md:216), so its whole type proves no arm apart, and a union naming
// every member is that type (tsc flattens it). A member literal, or a union of some members, is closed
// (rulings.md:220, Language on #wcva20f) and keeps its proof. An optional discriminant can be absent on
// both sides, which reads undefined, so it keeps the arm. Mutant: holdsOpenNumericEnum answering false
// turns the open rows red, and the optional undefined dropped turns the optional row red.
func TestVoidWidenedArmsUnderEnumsAndOptionals(test *testing.T) {
	test.Parallel()
	const enums = `enum Kind { A, B }
enum Other { X = 5 }
function noop(): void { }
`
	refused := "a void-returning function at the initializer site through .run seen as () => number"
	for _, probe := range []struct {
		name, source, extension string
		lowers                  bool
	}{
		{
			"open enum against another enum's member, .ts",
			enums + `interface S { readonly kind: Kind; readonly run: () => void; }
interface N { readonly kind: Other.X; readonly run: () => number; }
const s: S = { kind: Kind.A, run: noop };
const either: N | S = s;
console.log(String(either.kind));`,
			".ts",
			false,
		},
		{
			"open enum against another enum's member, .a",
			enums + `interface S { readonly kind: Kind; readonly run: () => void; }
interface N { readonly kind: Other.X; readonly run: () => number; }
const s: S = { kind: Kind.A, run: noop };
const either: N | S = s;
console.log(String(either.kind));`,
			".a",
			false,
		},
		{
			"every member named is the open enum",
			enums + `interface S { readonly kind: Kind.A | Kind.B; readonly run: () => void; }
interface N { readonly kind: Other.X; readonly run: () => number; }
const s: S = { kind: Kind.A, run: noop };
const either: N | S = s;
console.log(String(either.kind));`,
			".ts",
			false,
		},
		{"a member literal is closed", enums + `interface S { readonly kind: Kind.A; readonly run: () => void; }
interface N { readonly kind: Kind.B; readonly run: () => number; }
const s: S = { kind: Kind.A, run: noop };
const either: N | S = s;
console.log(String(either.kind));`, ".ts", true},
		{"optional discriminants both absent", `function noop(): void { }
interface S { readonly kind?: "a"; readonly run: () => void; }
interface P { readonly kind?: "a"; readonly run: () => void; }
interface N { readonly kind?: "b"; readonly run: () => number; }
const s: S = { run: noop };
const either: P | N = s;
console.log(typeof either.kind);`, ".ts", false},
	} {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			if probe.lowers {
				lowersAndAgreesWithNodeAs(test, probe.source+"\n", probe.extension)
				return
			}
			_, err := lowerSourceAs(test, probe.source+"\n", probe.extension)
			var stop *Refused
			if !errors.As(err, &stop) || !strings.Contains(stop.What, refused) {
				test.Fatalf("want Refused with %q, got %v", refused, err)
			}
		})
	}
}
