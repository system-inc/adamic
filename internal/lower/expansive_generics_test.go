package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

const expansiveGenericsWords = "a generic instantiated without end (polymorphic recursion)"

// A generic whose own members instantiate it deeper is a new type at every level, and each walk over
// it whose only stop was a set of the types it had seen ran without end (#k031ey0, #fzx0j5z, #ykd89v9).
// expansiveGenerics refuses it where it's declared, before lowering walks any type: through the generic
// itself, a tuple, an array, an object type, beside a Map member, through a second generic, a type
// alias and a base, in .a and .ts. The test returning is the proof that nothing hangs. Mutant
// (expansive-generics.mutants.md): with every edge plain, the rows no bound holds run until Go's test
// timeout.
func TestExpansiveGenericsStopsWhereTheGenericIsDeclared(t *testing.T) {
	t.Parallel()
	for _, row := range []struct{ name, source string }{
		{"in itself", "interface Own<T> { readonly next: Own<Own<T>> | undefined; readonly v: T; }\nconst o: Own<number> = { next: undefined, v: 1 };\nconsole.log(String(o.v));\n"},
		{"in a tuple", "interface Own<T> { readonly next: Own<[T, T]> | undefined; readonly v: T; }\nconst o: Own<number> = { next: undefined, v: 1 };\nconsole.log(String(o.v));\n"},
		{"in an array", "interface Own<T> { readonly next: Own<T[]> | undefined; readonly v: T; }\nconst o: Own<number> = { next: undefined, v: 1 };\nconsole.log(String(o.v));\n"},
		{"in an object type", "interface Own<T> { readonly next: Own<{ readonly v: T }> | undefined; readonly v: T; }\nconst o: Own<number> = { next: undefined, v: 1 };\nconsole.log(String(o.v));\n"},
		{"beside a map", "interface Own<T> { readonly next: Own<Own<T>> | undefined; readonly m: Map<string, T>; }\nconst o: Own<number> = { next: undefined, m: new Map<string, number>([['a', 1]]) };\nconsole.log(String(o.m.has('a')));\n"},
		// The growth goes through a second generic, which only holds the first plainly.
		{"through a second generic", "interface A<T> { readonly b: B<A<T>> | undefined; readonly v: T; }\ninterface B<U> { readonly a: A<U> | undefined; }\nconst x: A<number> = { b: undefined, v: 1 };\nconsole.log(String(x.v));\n"},
		{"a type alias", "type Grow<T> = { readonly next: Grow<Grow<T>> | undefined; readonly v: T };\nconst g: Grow<number> = { next: undefined, v: 1 };\nconsole.log(String(g.v));\n"},
		// The base is no cycle of its own, Base<T> holds T plainly; Grow<Grow<T>> in its argument is.
		{"through a base", "interface Base<T> { readonly v: T | undefined; }\ninterface Grow<T> extends Base<Grow<Grow<T>>> {}\nconst g: Grow<number> = { v: undefined };\nconsole.log(String(g.v === undefined));\n"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(row.name+extension, func(t *testing.T) {
				t.Parallel()
				program, err := lowerSourceAs(t, row.source, extension)
				var refused *Refused
				if !errors.As(err, &refused) || !strings.Contains(refused.What, expansiveGenericsWords) || !strings.Contains(refused.Where, "main"+extension+":") {
					t.Fatalf("got %v, want a refusal in main%s saying %q", err, extension, expansiveGenericsWords)
				}
				if program != nil {
					t.Fatal("a stopped program came back lowered")
				}
				t.Logf("STOP %s%s: %s (%s)", row.name, extension, refused.Where[strings.LastIndex(refused.Where, "main"):], refused.What)
			})
		}
	}
}

// The cycle crosses modules (Compiler's condition 3): the entry's Near holds the second module's Far
// plainly, and Far grows Near, so the stop is in the second module, which moduleOrder walks first.
func TestExpansiveGenericsStopsAcrossTwoModules(t *testing.T) {
	t.Parallel()
	for _, extension := range []string{".a", ".ts"} {
		t.Run("two modules"+extension, func(t *testing.T) {
			t.Parallel()
			// A .ts import resolves without its extension, and an .a one names it, as
			// module_namespace_test.go's do.
			second, entry := "./second", "./main"
			if extension == ".a" {
				second, entry = second+extension, entry+extension
			}
			directory := t.TempDir()
			for name, source := range map[string]string{
				"main" + extension:   "import { seed } from '" + second + "';\nimport type { Far } from '" + second + "';\nexport interface Near<T> { readonly far: Far<T> | undefined; }\nconst n: Near<number> = { far: { back: undefined, v: seed } };\nconsole.log(String(n.far === undefined));\n",
				"second" + extension: "import type { Near } from '" + entry + "';\nexport interface Far<T> { readonly back: Near<[T, T]> | undefined; readonly v: T; }\nexport const seed = 1;\n",
			} {
				if err := os.WriteFile(filepath.Join(directory, name), []byte(source), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			checked, err := load.Load([]string{filepath.Join(directory, "main"+extension)})
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			program, err := Lower(context.Background(), checked)
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(refused.What, expansiveGenericsWords) || !strings.Contains(refused.Where, "second"+extension+":") {
				t.Fatalf("got %v, want a refusal in second%s saying %q", err, extension, expansiveGenericsWords)
			}
			if program != nil {
				t.Fatal("a stopped program came back lowered")
			}
			t.Logf("STOP two modules%s: %s (%s)", extension, refused.Where[strings.LastIndex(refused.Where, "second"):], refused.What)
		})
	}
}

// The controls: a generic that holds itself without growing is one type at every level; one that
// grows an argument to Array has no cycle back to it; a written nest of arrays is a finite type. Each
// lowers.
func TestExpansiveGenericsControlsLower(t *testing.T) {
	t.Parallel()
	for _, row := range []struct{ name, source string }{
		{"a plain self edge", "interface Box<T> { readonly next: Box<T> | undefined; readonly v: T; }\nconst o: Box<number> = { next: undefined, v: 1 };\nconsole.log(String(o.v));\n"},
		{"an array with no cycle back", "interface Pair<T> { readonly a: T[]; readonly b: Pair<T> | undefined; }\nconst p: Pair<number> = { a: [1], b: undefined };\nconsole.log(String(p.a.length));\n"},
		{"a written nest of arrays", "const grid: string[][][] = [[['a']]];\nconsole.log(String(grid.length));\n"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(row.name+extension, func(t *testing.T) {
				t.Parallel()
				if _, err := lowerSourceAs(t, row.source, extension); err != nil {
					t.Fatalf("got %v, want it lowered", err)
				}
			})
		}
	}
}

// DeepReadonly<T[K]> instantiates DeepReadonly with a member of T, never with something holding T, so
// its edge is plain, and keyof T and the K that indexes make none (Compiler's condition 2). It must
// not hang and must not stop at this rule; a stop at another rule is logged for reading.
func TestExpansiveGenericsLeavesDeepReadonly(t *testing.T) {
	t.Parallel()
	source := "type DeepReadonly<T> = { readonly [K in keyof T]: DeepReadonly<T[K]> };\ninterface Leaf { readonly n: number }\ninterface Tree { readonly leaf: Leaf }\nconst t: DeepReadonly<Tree> = { leaf: { n: 1 } };\nconsole.log(String(t.leaf.n));\n"
	for _, extension := range []string{".a", ".ts"} {
		t.Run("deep readonly"+extension, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, source, extension)
			if err != nil && strings.Contains(err.Error(), expansiveGenericsWords) {
				t.Fatalf("got %v, want DeepReadonly past this rule", err)
			}
			if err != nil {
				t.Logf("STOP deep readonly%s: %v", extension, err)
			}
		})
	}
}
