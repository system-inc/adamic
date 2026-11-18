package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

// recordStorageNestingShapes are the two self-expanding generics of #fzx0j5z, Own seen as Deep: each
// next is one level deeper than the type that holds it, through the generic itself or through a tuple.
var recordStorageNestingShapes = []struct{ name, deep, own string }{
	{"generic", "Deep<Deep<T>>", "Own<Own<T>>"},
	{"tuple", "Deep<[T, T]>", "Own<[T, T]>"},
}

// recordStorageNestingTypes declares Deep and Own with a shape's next.
func recordStorageNestingTypes(deep, own string) string {
	return "interface Deep<T> { readonly next: " + deep + " | undefined; readonly m: ReadonlyMap<string, T>; }\n" +
		"interface Own<T> { readonly next: " + own + " | undefined; readonly m: Map<string, T>; }\n"
}

const recordStorageNestingWords = "comparing record storage through a type that instantiates itself deeper at each level (Own<number> seen as Deep<number>)"

// Two distinct generics that each instantiate themselves deeper at every level, Own<number> seen as
// Deep<number>, made sameRecordStorage walk a fresh pair at every step, so its visited set never
// repeated and lowering never ended (#fzx0j5z; Node prints true). typeNesting bounds the walk now, and
// the view stops as not yet in the bound's words, not as different storage and not refused. The test
// returning is the proof of bounded time. Mutant (records-storage-nesting.mutants.md): with the enter
// checks forced ok, it runs until Go's test timeout.
func TestRecordStorageNestingStopsAtTheView(t *testing.T) {
	t.Parallel()
	for _, shape := range recordStorageNestingShapes {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(shape.name+extension, func(t *testing.T) {
				t.Parallel()
				source := recordStorageNestingTypes(shape.deep, shape.own) +
					"const o: Own<number> = { next: undefined, m: new Map<string, number>([['a', 1]]) };\n" +
					"const d: Deep<number> = o;\n" +
					"console.log(String(d.m.has('a')));\n"
				program, err := lowerSourceAs(t, source, extension)
				var notYet *NotYet
				var refused *Refused
				if errors.As(err, &refused) {
					t.Fatalf("got refused %v, want not yet: the program is valid", err)
				}
				if !errors.As(err, &notYet) || !strings.Contains(notYet.What, recordStorageNestingWords) {
					t.Fatalf("got %v, want a NotYet saying %q", err, recordStorageNestingWords)
				}
				if program != nil {
					t.Fatal("a stopped program came back lowered")
				}
			})
		}
	}
}

// The walk itself stops by a frame count, not a clock: sameRecordStorage over Own<number> and
// Deep<number> answers false with bounded set, having visited a few pairs for each of typeNestingLimit
// levels. Mutant: with the enter checks forced ok, the walk never returns.
func TestRecordStorageNestingBoundsTheWalk(t *testing.T) {
	t.Parallel()
	for _, shape := range recordStorageNestingShapes {
		t.Run(shape.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "main.a")
			source := recordStorageNestingTypes(shape.deep, shape.own) + "function view(own: Own<number>, deep: Deep<number>): void {}\n"
			if err := os.WriteFile(path, []byte(source), 0600); err != nil {
				t.Fatal(err)
			}
			program, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			file := program.Files()[0]
			checked, release := program.Checker(context.Background(), file)
			t.Cleanup(release)
			l := &lowering{program: program, checker: checked}
			parameters := map[string]*checker.Type{}
			for _, statement := range file.Statements.Nodes {
				if statement.Kind != ast.KindFunctionDeclaration {
					continue
				}
				for _, parameter := range statement.AsFunctionDeclaration().Parameters.Nodes {
					parameters[parameter.Name().Text()] = checked.GetTypeAtLocation(parameter.Name())
				}
			}
			walk := newRecordStorageWalk()
			if l.sameRecordStorage(parameters["own"], parameters["deep"], walk) || !walk.bounded {
				t.Errorf("same storage with bounded %v, want false at the bound", walk.bounded)
			}
			if len(walk.visited) > 64 {
				t.Errorf("%d pairs visited, want at most 64", len(walk.visited))
			}
		})
	}
}

// The control: one self-expanding generic seen as itself with the same argument is one type, so the
// walk never starts and the program lowers.
func TestRecordStorageNestingControlLowers(t *testing.T) {
	t.Parallel()
	for _, extension := range []string{".a", ".ts"} {
		t.Run("own"+extension, func(t *testing.T) {
			t.Parallel()
			source := recordStorageNestingTypes("Deep<Deep<T>>", "Own<Own<T>>") +
				"const o: Own<number> = { next: undefined, m: new Map<string, number>([['a', 1]]) };\n" +
				"const d: Own<number> = o;\n" +
				"console.log(String(d.m.has('a')));\n"
			if _, err := lowerSourceAs(t, source, extension); err != nil {
				t.Fatalf("got %v, want it lowered", err)
			}
		})
	}
}
