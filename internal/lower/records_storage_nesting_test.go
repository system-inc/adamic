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

// Two distinct generics that each instantiate themselves deeper at every level, Own<number> seen as
// Deep<number>, made sameRecordStorage walk a fresh pair at every step, so its visited set never
// repeated and lowering never ended (#fzx0j5z; Node prints true). typeNesting bounds the walk, and since
// #ykd89v9 the expansive-generic pre-pass stops the program first, where Deep is declared, refused as
// polymorphic recursion, which docs/0.1.md refuses. TestRecordStorageNestingBoundsTheWalk calls the walk itself, so it's the bound's
// witness (records-storage-nesting.mutants.md).
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
				var refused *Refused
				if !errors.As(err, &refused) || !strings.Contains(refused.What, expansiveGenericsWords) {
					t.Fatalf("got %v, want a refusal saying %q", err, expansiveGenericsWords)
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

// The control: two distinct generics that don't grow, OwnBox<number> seen as DeepBox<number>, are
// walked member by member, as Own and Deep are, and each member's pair is a level the bound never
// reaches, so the walk answers alike and the program lowers. A self-expanding generic seen as itself
// can't serve here: Own<number> alone stops first in the cycle finder (#ykd89v9).
func TestRecordStorageNestingControlLowers(t *testing.T) {
	t.Parallel()
	for _, extension := range []string{".a", ".ts"} {
		t.Run("box"+extension, func(t *testing.T) {
			t.Parallel()
			source := "interface DeepBox<T> { readonly inner: T; readonly m: ReadonlyMap<string, T>; }\n" +
				"interface OwnBox<T> { readonly inner: T; readonly m: Map<string, T>; }\n" +
				"const o: OwnBox<number> = { inner: 1, m: new Map<string, number>([['a', 1]]) };\n" +
				"const d: DeepBox<number> = o;\n" +
				"console.log(String(d.m.has('a')));\n"
			if _, err := lowerSourceAs(t, source, extension); err != nil {
				t.Fatalf("got %v, want it lowered", err)
			}
		})
	}
}

// A promise is compared by what it settles to: lib's Promise and PromiseLike store alike when their type
// argument does (records.go). Their then members meet a fresh instantiation at every level, so before
// that arm, Promise<number> seen as PromiseLike<number> walked into typeNesting's bound and stopped
// oracle's async adoption and callback pins as not yet (J-vsqgd66). A record inside still decides by its
// storage, as different, never as bounded. Mutant (records-storage-nesting.mutants.md): drop the
// Promise arm, and the alike rows answer false with bounded set.
func TestRecordStorageComparesPromisesByWhatTheySettleTo(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "main.a")
	source := "function view(promise: Promise<number>, like: PromiseLike<number>, dictionary: Promise<Record<string, number>>, fixed: PromiseLike<{ readonly a: number }>): void {}\n"
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
	for _, parameter := range file.Statements.Nodes[0].AsFunctionDeclaration().Parameters.Nodes {
		parameters[parameter.Name().Text()] = checked.GetTypeAtLocation(parameter.Name())
	}
	for _, pair := range []struct {
		from, to string
		same     bool
	}{
		{"promise", "like", true},
		{"like", "promise", true},
		{"dictionary", "fixed", false},
	} {
		walk := newRecordStorageWalk()
		if got := l.sameRecordStorage(parameters[pair.from], parameters[pair.to], walk); got != pair.same || walk.bounded {
			t.Errorf("%s to %s: same storage %v with bounded %v, want %v unbounded", pair.from, pair.to, got, walk.bounded, pair.same)
		}
		if len(walk.visited) > 16 {
			t.Errorf("%s to %s: %d pairs visited, want at most 16", pair.from, pair.to, len(walk.visited))
		}
	}
}

// The Promise arm compares lib's Promise and PromiseLike by their type argument, so it reaches only a
// value whose type is one of those two on each side. A program-made thenable is typed PromiseLike, and
// PromiseLike is never assignable to Promise (no catch or finally), so the one road that would hand
// the arm a thenable as a Promise is a cast, which no-unchecked-cast refuses. The thenable's other roads
// into Promise storage stop on their own rules: Promise.resolve of it and an async function returning
// it (J-jnkcqek, Compiler's question on #fzx0j5z). The thenable here lowers on its own: its then throws.
func TestRecordStoragePromiseArmMeetsNoThenable(t *testing.T) {
	t.Parallel()
	thenable := "class Later implements PromiseLike<number> { then<A = number, B = never>(): PromiseLike<A | B> { throw new Error('unused'); } }\n" +
		"const thenable: PromiseLike<number> = new Later();\n"
	for _, extension := range []string{".a", ".ts"} {
		for _, row := range []struct {
			name, tail, where, words string
			refused                  bool
		}{
			{"cast", "const promised = thenable as Promise<number>;\nconsole.log(typeof promised);\n", ":3:18", "a cast the runtime can't check", true},
			{"resolve", "const promised: Promise<number> = Promise.resolve(thenable);\npromised.then((v) => { console.log(String(v)); });\n", ":3:35", "Promise thenables", false},
			{"async return", "async function lift(): Promise<number> { return thenable; }\nlift().then((v) => { console.log(String(v)); });\n", ":3:49", "return of thenables", false},
		} {
			t.Run(row.name+extension, func(t *testing.T) {
				t.Parallel()
				_, err := lowerSourceAs(t, thenable+row.tail, extension)
				want := "main" + extension + row.where
				var notYet *NotYet
				var refused *Refused
				if row.refused {
					if !errors.As(err, &refused) || !strings.HasSuffix(refused.Where, want) || !strings.Contains(refused.What, row.words) {
						t.Fatalf("got %v, want refused at %s saying %q", err, want, row.words)
					}
					return
				}
				if !errors.As(err, &notYet) || !strings.HasSuffix(notYet.Where, want) || !strings.Contains(notYet.What, row.words) {
					t.Fatalf("got %v, want a NotYet at %s saying %q", err, want, row.words)
				}
			})
		}
	}
}
