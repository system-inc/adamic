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

// nestingBoundsProgram loads one .ts source and returns a bare lowering over it, its modules in
// moduleOrder, and the type of each named function parameter.
func nestingBoundsProgram(t *testing.T, source string) (*lowering, []*ast.SourceFile, map[string]*checker.Type) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "main.ts")
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
	l := &lowering{program: program, checker: checked, entry: file}
	modules, err := l.moduleOrder(file)
	if err != nil {
		t.Fatal(err)
	}
	parameters := map[string]*checker.Type{}
	for _, statement := range file.Statements.Nodes {
		if statement.Kind != ast.KindFunctionDeclaration {
			continue
		}
		for _, parameter := range statement.AsFunctionDeclaration().Parameters.Nodes {
			parameters[parameter.Name().Text()] = checked.GetTypeAtLocation(parameter.Name())
		}
	}
	return l, modules, parameters
}

// The expansive-generic pre-pass (#ykd89v9) refuses a generic that instantiates itself deeper before any
// walk meets it, so no program reaches the bounds collectViewedElements (#k031ey0) and the cycle finder's
// use walk keep. Those bounds stay as defense, and a defense nothing tests is the code the next edit
// deletes without noticing (Compiler, on #ykd89v9). These tests call each walk directly, past the
// pre-pass, as TestRecordStorageNestingBoundsTheWalk calls sameRecordStorage, so each bound keeps a red
// witness. Mutants (nesting-bounds-direct.mutants.md): each walk's enter forced ok, and its test runs
// until Go's timeout.

// collectViewedElements, called on its own over a predicate view whose target reaches a growing generic,
// stops at that view in its own words.
func TestCollectViewedElementsBoundsAGrowingGenericDirectly(t *testing.T) {
	t.Parallel()
	l, modules, _ := nestingBoundsProgram(t, `interface Grow<T> { readonly value?: T; readonly next?: Grow<Grow<T>> }
interface N { readonly kind: number; readonly grow?: Grow<number> }
interface Id extends N { readonly kind: 1; readonly text: string }
function isId(n: N): n is Id { return n.kind === 1; }
const id: Id = { kind: 1, text: "a" };
const n: N = id;
if (isId(n)) { console.log(n.text); }
`)
	err := l.collectViewedElements(modules)
	var notYet *NotYet
	if !errors.As(err, &notYet) || !strings.Contains(notYet.What, "instantiates itself deeper at each step") {
		t.Fatalf("got %v, want the view pre-pass's own stop at its bound", err)
	}
}

// The cycle finder's use walk, handed Own<number> on its own, notes the first type past its bound and
// stops, having noted a handful of types rather than one for every level.
func TestCycleFinderUseBoundsAGrowingGenericDirectly(t *testing.T) {
	t.Parallel()
	l, _, parameters := nestingBoundsProgram(t, `interface Own<T> { readonly next: Own<Own<T>> | undefined; readonly v: T; }
function hold(own: Own<number>): void {}
`)
	finder := &cycleFinder{l: l, where: map[*checker.Type]*ast.Node{}}
	finder.use(parameters["own"], nil)
	if finder.deepest == nil {
		t.Fatal("the use walk ended without meeting its bound, so it isn't the bound that ended it")
	}
	if len(finder.where) > 64 {
		t.Errorf("%d types noted, want at most 64", len(finder.where))
	}
}
