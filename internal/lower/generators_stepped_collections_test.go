package lower

import (
	"errors"
	"strings"
	"testing"
)

// new Map and new Set join spread and Array.from as stepped consumers (Adamic, on #qk8rztp, S2b-1c):
// each steps a generator to its end and drops its completion value, so a plain function*'s void
// TReturn is never read, and the void rules let a direct call through when it is the constructor's
// one argument and what it yields holds no void. Everything else stays judged: a yield that holds
// void (a void call, or a pair either of whose elements is void), a stored generator, a user class
// named Map, and a subclass of the built-in.
//
// Mutants: generatorStepped's new Map and new Set arm taken out (the lowering rows are refused
// again); widened to any Generator-typed expression at the site (the stored rows go red); its
// yielded-type test taken out (the void rows go red); the constructor judged by its name (the
// shadowed row goes red).
func TestGeneratorVoidReturnIsNeverReadByNewMapOrNewSet(t *testing.T) {
	t.Parallel()
	// Plain function*s, so each infers Generator<T, void, unknown>. The pair's tuple type comes
	// from a typed local, since an array literal yielded alone would infer an array.
	const pairs = "function* pairs(count: number) { for (let index = 0; index < count; index++) { const pair: [string, number] = ['k' + String(index), index]; yield pair; } }\n"
	const words = "function* words(count: number) { for (let index = 0; index < count; index++) { yield 'w' + String(index); } }\n"
	lowers := []struct{ name, source string }{
		{"a plain function* into new Map", pairs + "console.log(String(new Map(pairs(3)).size));"},
		{"a plain function* into new Set", words + "console.log(String(new Set(words(3)).size));"},
		{"tsc's mapIterator into new Map", "function* mapIt<T, U>(xs: readonly T[], f: (x: T) => U): Generator<U, void, unknown> { for (const x of xs) { yield f(x); } }\nconst flip = (entry: readonly [string, number]): [string, string] => ['' + String(entry[1]), entry[0]];\nconst entries: (readonly [string, number])[] = [['a', 1], ['b', 2]];\nconsole.log(String(new Map(mapIt(entries, flip)).size));"},
	}
	for _, probe := range lowers {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			if _, err := lowerSource(t, probe.source); err != nil {
				t.Fatalf("%v", err)
			}
		})
	}
	const voidCall = "function f(): void {}\n"
	refused := []struct{ name, source string }{
		{"a void-yielding generator into new Set", voidCall + "function* g() { yield f(); }\nconsole.log(String(new Set(g()).size));"},
		{"a pair whose value is void into new Map", "function* g(): Generator<[string, void], void, unknown> { yield ['k' + 'a', undefined]; }\nconsole.log(String(new Map(g()).size));"},
		{"a pair whose key is void into new Map", "function* g(): Generator<[void, string], void, unknown> { yield [undefined, 'v' + 'a']; }\nconsole.log(String(new Map(g()).size));"},
		{"a stored generator into new Map", pairs + "const stored = pairs(3);\nconsole.log(String(new Map(stored).size));"},
		{"a stored generator into new Set", words + "const stored = words(3);\nconsole.log(String(new Set(stored).size));"},
		{"a user class named Map", "export {};\nclass Map { constructor(readonly entries: Iterable<[string, number]>) {} }\n" + pairs + "const made = new Map(pairs(3));\nconsole.log('made');"},
	}
	for _, probe := range refused {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var refusal *Refused
			if !errors.As(err, &refusal) || !strings.Contains(refusal.Error(), "adamic/no-void") {
				t.Fatalf("got %v, want a void rule's refusal", err)
			}
		})
	}
	// A subclass of the built-in Map runs its own set once per pair, program code, so it is no
	// stepped consumer. It is pinned only as never lowering, whichever stop it meets first (a void
	// rule, or the class itself), so that admitting it is a change someone makes on purpose.
	t.Run("a subclass of the built-in Map", func(t *testing.T) {
		t.Parallel()
		if _, err := lowerSource(t, "class Tally extends Map<string, number> {}\n"+pairs+"console.log(String(new Tally(pairs(3)).size));"); err == nil {
			t.Fatal("a subclass of Map consumed a plain function*'s call and lowered; want it judged, not exempt")
		}
	})
}
