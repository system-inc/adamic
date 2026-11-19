package lower

import (
	"testing"
)

// The cycle finder's use walk is bounded by typeNesting (cycles.go). Its programs, a generic that
// instantiates itself deeper used with no view, are stopped first by the expansive-generic pre-pass,
// whose pins hold them (expansive_generics_test.go), so the bound is defense with no lowering witness
// (cycles-nesting.mutants.md).

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
