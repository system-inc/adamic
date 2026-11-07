package lower

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/load"
)

// programRegionBeginName is a begin's function as a reader names it: a declaration's own name, the
// variable an anonymous function is kept in, or its class for a constructor.
func programRegionBeginName(function *ast.Node) string {
	switch {
	case function.Kind == ast.KindConstructor:
		return programRegionBeginName(function.Parent) + ".constructor"
	case function.Name() != nil:
		return function.Name().Text()
	case function.Parent != nil && function.Parent.Kind == ast.KindVariableDeclaration:
		return function.Parent.Name().Text()
	}
	return "anonymous"
}

// programRegionRoots lowers source with Program discovery and names the functions a begin goes in,
// under rule (1) and under rule (2).
func programRegionRoots(t *testing.T, source string) ([]string, []string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "main.a")
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	entry := loaded.Files()[0]
	typeChecker, release := loaded.Checker(context.Background(), entry)
	defer release()
	discovered, err := lowerChecked(loaded, typeChecker, entry, nil, true, false)
	if err != nil {
		t.Fatal(err)
	}
	modules, err := discovered.moduleOrder(entry)
	if err != nil {
		t.Fatal(err)
	}
	if len(discovered.programPlan.tops) == 0 {
		t.Fatal("no top component")
	}
	allocating, returning := []string{}, []string{}
	for _, begin := range discovered.programRegionBegins(modules, discovered.programPlan.tops) {
		allocating = append(allocating, programRegionBeginName(begin.function))
		if begin.returned {
			returning = append(returning, programRegionBeginName(begin.function))
		}
	}
	slices.Sort(allocating)
	slices.Sort(returning)
	return allocating, returning
}

func programRegionRootsWant(t *testing.T, rule string, got, want []string) {
	t.Helper()
	if !slices.Equal(got, want) {
		t.Fatalf("rule %s: got %v, want %v", rule, got, want)
	}
}

// Program and File hold each other, so they're one owning component, and File owns a parse tree,
// a component of its own below it. The tree is entered from File, so only Program's component is
// a top. An array of Programs reaches Program without being a source. A helper that makes a File
// and returns nothing gets a begin under rule (1) only.
func TestProgramRegionRootsTopComponent(t *testing.T) {
	t.Parallel()
	allocating, returning := programRegionRoots(t, `interface Tree { parent: Tree | undefined; children: Tree[]; }
interface File { program: Program; root: Tree; }
interface Program { files: File[]; name: string; }
function parse(): Tree {
 const node: Tree = { parent: undefined, children: [] };
 const child: Tree = { parent: node, children: [] };
 node.children.push(child);
 return node;
}
function addFile(program: Program, root: Tree): void { const file: File = { program, root }; program.files.push(file); }
function makeFile(program: Program): File { const file: File = { program, root: parse() }; return file; }
function createProgram(name: string): Program {
 const program: Program = { files: [], name };
 addFile(program, parse());
 program.files.push(makeFile(program));
 return program;
}
const programs: Program[] = [];
programs.push(createProgram('a'));
programs.push(createProgram('b'));
console.log(programs.length);
`)
	programRegionRootsWant(t, "1", allocating, []string{"addFile", "createProgram", "makeFile"})
	programRegionRootsWant(t, "2", returning, []string{"createProgram", "makeFile"})
}

// A Session that keeps its previous Session is a component of its own, and it owns a Program, so
// Program's component is entered and Session's is the top.
func TestProgramRegionRootsHolderAbove(t *testing.T) {
	t.Parallel()
	allocating, returning := programRegionRoots(t, `interface File { program: Program; }
interface Program { files: File[]; }
interface Session { program: Program; previous: Session | undefined; }
function createProgram(): Program { const program: Program = { files: [] }; const file: File = { program }; program.files.push(file); return program; }
function createSession(previous: Session | undefined): Session { const session: Session = { program: createProgram(), previous }; return session; }
const first = createSession(undefined);
const second = createSession(first);
console.log(second.program.files.length);
`)
	programRegionRootsWant(t, "1", allocating, []string{"createSession"})
	programRegionRootsWant(t, "2", returning, []string{"createSession"})
}

// A nested function that builds the root gets the begin at its own entry, and the function that
// makes it doesn't.
func TestProgramRegionRootsNestedFunction(t *testing.T) {
	t.Parallel()
	allocating, returning := programRegionRoots(t, `interface File { program: Program; }
interface Program { files: File[]; }
function outer(): number {
 const build = (): Program => { const program: Program = { files: [] }; const file: File = { program }; program.files.push(file); return program; };
 return build().files.length + build().files.length;
}
console.log(outer());
`)
	programRegionRootsWant(t, "1", allocating, []string{"build"})
	programRegionRootsWant(t, "2", returning, []string{"build"})
}

// ProgramRegionSCC is the union of ProgramRegionComponents, which keeps a component only when an
// owning edge stays inside it.
func TestProgramRegionComponents(t *testing.T) {
	t.Parallel()
	edges := map[int][]int{1: {2}, 2: {1, 3}, 3: {4}, 4: {3}, 5: {6}, 6: {5}}
	owning := map[int][]int{1: {2}, 2: {3}, 3: {4}, 5: {}}
	components := ProgramRegionComponents(edges, owning)
	if len(components) != 2 {
		t.Fatalf("components %v, want {1,2} and {3,4}", components)
	}
	for _, component := range components {
		if !(component[1] && component[2] && len(component) == 2) && !(component[3] && component[4] && len(component) == 2) {
			t.Fatalf("unexpected component %v", component)
		}
	}
	if selected := ProgramRegionSCC(edges, owning); len(selected) != 4 || selected[5] || selected[6] {
		t.Fatalf("selection %v", selected)
	}
}
