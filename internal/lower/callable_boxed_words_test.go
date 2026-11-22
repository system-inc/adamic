package lower

import (
	"errors"
	"strings"
	"testing"
)

// TestCallableBoxedSplitNamesTheSlot holds the words for a callable seen through a signature whose slot one
// side holds as a boxed union and the other unboxed: the NotYet names that slot and both representations,
// never the weak sentence, which named the wrong cause at five tsc sites (#bjtwhg8, Adamic on #bbssdcy, family
// 34). The shapes are tsc's: a callback returning boolean handed where boolean | "quit" is returned
// (utilities.ts:2325 and :2332), and a context-local function whose parameter is wider than the shared
// one's (transformer.ts:679).
func TestCallableBoxedSplitNamesTheSlot(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, words string }{
		{"result unboxed in the function, boxed in the view",
			"function run(n: number): void {\n\tconst f: () => boolean | 'quit' = (): boolean => n > 0;\n\tconsole.log(`${f()}`);\n}\nrun(1);\n",
			"a () => boolean seen as a () => \"quit\" | boolean, whose result is boolean, held unboxed, in one and \"quit\" | boolean, held as a boxed union, in the other (a callable's boxed and unboxed representations differ, and the adapter between them isn't built)"},
		{"parameter boxed in the function, unboxed in the view",
			"function run(n: number): void {\n\tconst g: (x: number) => string = (x: number | string): string => `${x}`;\n\tconsole.log(g(n));\n}\nrun(1);\n",
			"a (x: string | number) => string seen as a (x: number) => string, whose parameter x is string | number, held as a boxed union, in one and number, held unboxed, in the other (a callable's boxed and unboxed representations differ, and the adapter between them isn't built)"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var later *NotYet
			if !errors.As(err, &later) || later.What != probe.words {
				t.Fatalf("got %v, want the NotYet %q", err, probe.words)
			}
			if strings.Contains(err.Error(), "weakly") {
				t.Fatalf("the weak sentence is back: %v", err)
			}
		})
	}
}
