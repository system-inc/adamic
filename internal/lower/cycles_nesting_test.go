package lower

import (
	"errors"
	"strings"
	"testing"
)

const cyclesNestingWords = "finding cycles through a type that instantiates itself deeper at each level"

// A generic that instantiates itself deeper at each level, used with no view at all, made the cycle
// finder's use walk meet a fresh type at every field (Own<Own<Own<number>>> and on), so where never
// repeated and lowering never ended (#ykd89v9, J-pp7vwgw on main c56b0229; Node prints 1). typeNesting
// bounds the walk now, and the program stops as not yet, in .a and .ts, through the generic itself,
// beside a Map member, and through a tuple. The test returning is the proof of bounded time. Mutant
// (cycles-nesting.mutants.md): with the enter forced ok, every row runs until Go's test timeout.
func TestCyclesNestingStopsOnAGenericThatDeepensItself(t *testing.T) {
	t.Parallel()
	for _, row := range []struct{ name, source string }{
		{"generic", "interface Own<T> { readonly next: Own<Own<T>> | undefined; readonly v: T; }\nconst o: Own<number> = { next: undefined, v: 1 };\nconsole.log(String(o.v));\n"},
		{"beside a map", "interface Own<T> { readonly next: Own<Own<T>> | undefined; readonly m: Map<string, T>; }\nconst o: Own<number> = { next: undefined, m: new Map<string, number>([['a', 1]]) };\nconsole.log(String(o.m.has('a')));\n"},
		{"tuple", "interface Own<T> { readonly next: Own<[T, T]> | undefined; readonly v: T; }\nconst o: Own<number> = { next: undefined, v: 1 };\nconsole.log(String(o.v));\n"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(row.name+extension, func(t *testing.T) {
				t.Parallel()
				program, err := lowerSourceAs(t, row.source, extension)
				var notYet *NotYet
				if !errors.As(err, &notYet) || !strings.Contains(notYet.What, cyclesNestingWords) || !strings.Contains(notYet.Where, "main"+extension+":") {
					t.Fatalf("got %v, want a NotYet in main%s saying %q", err, extension, cyclesNestingWords)
				}
				if program != nil {
					t.Fatal("a stopped program came back lowered")
				}
				t.Logf("STOP %s%s: %s", row.name, extension, notYet.Where[strings.LastIndex(notYet.Where, "main"):])
			})
		}
	}
}

// The control: a generic that holds itself without growing, Box<T> { next: Box<T> }, is one type at
// every level, so the walk ends by where alone and the program lowers.
func TestCyclesNestingControlLowers(t *testing.T) {
	t.Parallel()
	for _, extension := range []string{".a", ".ts"} {
		t.Run("box"+extension, func(t *testing.T) {
			t.Parallel()
			source := "interface Box<T> { readonly next: Box<T> | undefined; readonly v: T; }\nconst o: Box<number> = { next: undefined, v: 1 };\nconsole.log(String(o.v));\n"
			if _, err := lowerSourceAs(t, source, extension); err != nil {
				t.Fatalf("got %v, want it lowered", err)
			}
		})
	}
}
