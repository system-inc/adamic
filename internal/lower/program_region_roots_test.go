package lower

import (
	"context"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

// programRegionBeginName is a begin's function as a reader names it: a declaration's own name, the
// variable an anonymous function is kept in, or its class for a constructor.
func programRegionBeginName(function *ast.Node) string {
	switch {
	case function.Kind == ast.KindConstructor:
		return programRegionBeginName(function.Parent) + ".constructor"
	case function.Name() != nil:
		return programRegionNameText(function.Name())
	case function.Parent != nil && function.Parent.Kind == ast.KindVariableDeclaration:
		return programRegionNameText(function.Parent.Name())
	case function.Parent != nil && function.Parent.Kind == ast.KindPropertyAssignment:
		return programRegionNameText(function.Parent.Name())
	}
	return "anonymous"
}

// programRegionNameText is a declaration name's text, or "computed" for one tsc works out.
func programRegionNameText(name *ast.Node) string {
	if ast.IsIdentifier(name) || name.Kind == ast.KindStringLiteral || name.Kind == ast.KindNumericLiteral || name.Kind == ast.KindPrivateIdentifier {
		return name.Text()
	}
	return "computed"
}

// programRegionDiscovered lowers source with Program discovery, and returns the discovery lowering
// and its modules. The checker is released when the test ends.
func programRegionDiscovered(t *testing.T, source string) (*lowering, []*ast.SourceFile) {
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
	t.Cleanup(release)
	discovered, err := lowerChecked(loaded, typeChecker, entry, nil, true, false)
	if err != nil {
		t.Fatal(err)
	}
	modules, err := discovered.moduleOrder(entry)
	if err != nil {
		t.Fatal(err)
	}
	return discovered, modules
}

// programRegionBeginNames names the functions a begin goes in, under rule (1) and under rule (2).
func programRegionBeginNames(discovered *lowering, modules []*ast.SourceFile) ([]string, []string) {
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

// programRegionRoots lowers source with Program discovery and names the functions a begin goes in,
// under rule (1) and under rule (2).
func programRegionRoots(t *testing.T, source string) ([]string, []string) {
	t.Helper()
	discovered, modules := programRegionDiscovered(t, source)
	if len(discovered.programPlan.tops) == 0 {
		t.Fatal("no top component")
	}
	return programRegionBeginNames(discovered, modules)
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
console.log(programs.length.toString());
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
console.log(second.program.files.length.toString());
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
console.log(outer().toString());
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

// programRegionRebuilt rebuilds a discovery lowering's finder, seeded as findCycles and
// selectProgramMembership seed it, so a test can read the owning components the plan was chosen
// from. Its tops must equal the plan's, or the copy has drifted from the seeding it mirrors.
func programRegionRebuilt(t *testing.T, discovered *lowering, modules []*ast.SourceFile) (*cycleFinder, programRegionOwning) {
	t.Helper()
	finder := &cycleFinder{l: discovered, where: map[*checker.Type]*ast.Node{}}
	for _, module := range modules {
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if ast.IsPartOfTypeNode(node) {
				return false
			}
			finder.libraryIteratorMade(node)
			switch {
			case node.Kind == ast.KindObjectLiteralExpression:
				finder.shape(discovered.checker.GetTypeAtLocation(node), node)
			case node.Kind == ast.KindArrayLiteralExpression || node.Kind == ast.KindCallExpression || node.Kind == ast.KindNewExpression:
				finder.made(discovered.checker.GetTypeAtLocation(node), node)
			case ast.IsIdentifier(node) || node.Kind == ast.KindThisKeyword || node.Kind == ast.KindPropertyAccessExpression || node.Kind == ast.KindElementAccessExpression:
				if ast.IsIdentifier(node) && !ast.IsDeclarationName(node) && !ast.IsExpressionNode(node) {
					break
				}
				finder.use(discovered.checker.GetTypeAtLocation(node), node)
			}
			if ast.IsExpressionNode(node) {
				finder.use(discovered.checker.GetTypeAtLocation(node), node)
				if contextual := discovered.slotType(node); contextual != nil {
					finder.use(contextual, node)
				}
			}
			return node.ForEachChild(visit)
		}
		module.AsNode().ForEachChild(visit)
	}
	for _, proven := range discovered.instantiated {
		finder.use(proven, discovered.classNodeFor(proven))
	}
	if err := finder.graphTypes(modules); err != nil {
		t.Fatal(err)
	}
	for _, closure := range discovered.closureRecords {
		finder.use(closure.proven, closure.node)
	}
	for proven, node := range discovered.programCandidates {
		finder.use(proven, node)
	}
	for _, local := range slices.Sorted(maps.Keys(discovered.localTypes)) {
		if proven := discovered.localTypes[local]; proven != nil {
			finder.use(proven, discovered.localNodes[local])
		}
	}
	_, owning := finder.programRegionOwningMode(true)
	if tops := finder.programRegionTops(owning, discovered.programPlan.types); !maps.Equal(tops, discovered.programPlan.tops) {
		t.Fatalf("rebuilt tops %v differ from the plan's %v", tops, discovered.programPlan.tops)
	}
	return finder, owning
}

// tsc's shape, where every interface is methods only: the checker is an object of arrow
// properties capturing its host, and the Program an object of createProgram's nested functions,
// as program.ts lists them, whose getChecker builds the checker lazily from the Program itself and
// keeps it in a captured let. A function value can't capture the variable its own initializer
// declares, which is why tsc's getTypeChecker is a nested declaration too. deriveFields
// drops method signatures, so whether the interfaces or only the literals' own types are members,
// and where the root falls, is what this pins.
const programRegionMethodOnlySource = `interface SourceFile { fileName: string; text: string; }
interface Host {
 getSourceFile(fileName: string): SourceFile | undefined;
 getFileNames(): string[];
}
interface Checker {
 getFileCount(): number;
 getHost(): Host;
}
interface Program extends Host {
 getChecker(): Checker;
}
function createChecker(host: Host): Checker {
 const checker: Checker = {
  getFileCount: () => host.getFileNames().length,
  getHost: () => host,
 };
 return checker;
}
function createProgram(names: string[]): Program {
 const files: SourceFile[] = [];
 for (const name of names) {
  files.push({ fileName: name, text: 'text of ' + name });
 }
 let checker: Checker | undefined = undefined;
 function getSourceFile(fileName: string): SourceFile | undefined {
  return files.find((file) => file.fileName === fileName);
 }
 function getFileNames(): string[] {
  return files.map((file) => file.fileName);
 }
 function getChecker(): Checker {
  if (checker === undefined) {
   checker = createChecker(program);
  }
  return checker;
 }
 const program: Program = { getSourceFile, getFileNames, getChecker };
 return program;
}
const first = createProgram(['a.ts', 'b.ts']);
console.log(first.getChecker().getFileCount().toString());
const second = createProgram(['c.ts']);
console.log(second.getChecker().getHost().getFileNames().join(','));
const found = first.getSourceFile('a.ts');
console.log(found === undefined ? 'missing' : found.text);
`

func TestProgramRegionRootsMethodOnly(t *testing.T) {
	t.Parallel()
	discovered, modules := programRegionDiscovered(t, programRegionMethodOnlySource)
	_, owning := programRegionRebuilt(t, discovered, modules)
	format := func(proven *checker.Type) string {
		return discovered.checker.TypeToStringEx(proven, nil, checker.TypeFormatFlagsNoTruncation, nil)
	}
	component := map[*checker.Type]int{}
	for index, nodes := range owning.components {
		labels, top := []string{}, false
		for node := range nodes {
			if node.proven == nil {
				labels = append(labels, "cell "+discovered.result.Locals[node.cell-1].Name)
				continue
			}
			component[node.proven] = index
			label := format(node.proven)
			if discovered.programPlan.types[int(node.proven.Id())] {
				label += " (member)"
			}
			if discovered.programPlan.tops[int(node.proven.Id())] {
				label += " (top)"
				top = true
			}
			labels = append(labels, label)
		}
		slices.Sort(labels)
		t.Logf("component %d top=%t: %s", index, top, strings.Join(labels, "; "))
	}
	members := []string{}
	for id := range discovered.programPlan.types {
		for proven := range component {
			if int(proven.Id()) == id {
				members = append(members, format(proven))
			}
		}
	}
	slices.Sort(members)
	t.Logf("member types (%d, %d outside any component): %s", len(discovered.programPlan.types), len(discovered.programPlan.types)-len(members), strings.Join(members, "; "))
	for local, member := range discovered.programPlan.cells {
		if member {
			t.Logf("member cell %s", discovered.result.Locals[local].Name)
		}
	}
	// The object literal each of createProgram and createChecker keeps, by its own type and its slot's.
	literals := map[string][]*checker.Type{}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindObjectLiteralExpression && node.Parent.Kind == ast.KindVariableDeclaration {
			name := node.Parent.Name().Text()
			literals[name] = append(literals[name], discovered.checker.GetTypeAtLocation(node), discovered.slotType(node))
		}
		return node.ForEachChild(visit)
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	for _, name := range []string{"program", "checker"} {
		for _, proven := range literals[name] {
			index, in := component[proven]
			if !in {
				index = -1
			}
			t.Logf("literal %s: %s in component %d, member %t, top %t", name, format(proven), index, discovered.programPlan.types[int(proven.Id())], discovered.programPlan.tops[int(proven.Id())])
		}
	}
	allocating, returning := programRegionBeginNames(discovered, modules)
	t.Logf("rule 1: %v", allocating)
	t.Logf("rule 2: %v", returning)
}
