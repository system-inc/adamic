package lower

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

func TestProgramRegionMembership(test *testing.T) {
	test.Parallel()
	path := filepath.Join(test.TempDir(), "main.a")
	source := `import type { Weak } from 'adamic';
interface Node { id: string; children: Node[]; parent: Weak<Node>; }
interface Leaf { text: string; }
function run():void {
 const root:Node={id:'r',children:[],parent:undefined};
 const child:Node={id:'c',children:[],parent:root};root.children.push(child);
 const leaf:Leaf={text:'leaf'};console.log(leaf.text);
 const cache=new Map<string,number>();cache.set('r',1);
 const scratch:number[]=[1,2];console.log('scratch');
 console.log('tree');
}run();`
	if writeError := os.WriteFile(path, []byte(source), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	loaded, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	off, lowerWithOptionsError := LowerWithOptions(context.Background(), loaded, Options{})
	if lowerWithOptionsError != nil {
		test.Fatal(lowerWithOptionsError)
	}
	if off.ProgramRegion || len(off.ProgramTypes) != 0 {
		test.Fatal("prototype enabled by default")
	}
	on, lowerWithOptionsError := LowerWithOptions(context.Background(), loaded, Options{ProgramRegion: true})
	if lowerWithOptionsError != nil {
		test.Fatal(lowerWithOptionsError)
	}
	markedObjects, markedArrays := 0, 0
	for _, function := range on.Functions {
		walk(function.Body, func(node any) bool {
			switch node := node.(type) {
			case ir.ObjectLiteral:
				for _, field := range node.Fields {
					if field.Name == "id" {
						if !node.ProgramRegion {
							test.Fatal("recursive Node left counted")
						}
						markedObjects++
					}
					if field.Name == "text" && node.ProgramRegion {
						test.Fatal("acyclic Leaf made a member")
					}
				}
			case ir.MapNew:
				if node.ProgramRegion {
					test.Fatal("scalar relation scratch made a member")
				}
			case ir.ArrayLiteral:
				if node.Element == ir.Number && node.ProgramRegion {
					test.Fatal("scalar temporary array made a member")
				}
				if node.ProgramRegion {
					markedArrays++
				}
			}
			return true
		})
	}
	if markedObjects != 2 || markedArrays != 2 {
		test.Fatalf("membership objects=%d arrays=%d", markedObjects, markedArrays)
	}
}

func TestProgramRegionViewCycleIsNotOwnership(test *testing.T) {
	test.Parallel()
	edges := map[string][]string{"a": {"b"}, "b": {"a"}, "string": nil}
	if got := ProgramRegionSCC(edges, map[string][]string{"a": {"string"}}); len(got) != 0 {
		test.Fatalf("view-only SCC selected: %v", got)
	}
	edges["a"] = append(edges["a"], "a")
	if got := ProgramRegionSCC(
		edges,
		map[string][]string{"a": {"a", "string"}},
	); !got["a"] || !got["b"] || got["string"] {
		test.Fatal(got)
	}
}

func TestProgramRegionTaskBoundary(test *testing.T) {
	test.Parallel()
	path := filepath.Join(test.TempDir(), "main.a")
	source := `import { parallelMap } from 'adamic';
interface Tree { readonly children: readonly Tree[]; }
const tree: Tree = {children:[]};
console.log('start');
const items: readonly number[] = [1,2];
const values = parallelMap(items, (value:number):number => value+1);
console.log('end');`
	if writeError := os.WriteFile(path, []byte(source), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	loaded, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	if _, lowerWithOptionsError := LowerWithOptions(
		context.Background(),
		loaded,
		Options{ProgramRegion: true},
	); lowerWithOptionsError == nil || !strings.Contains(lowerWithOptionsError.Error(), "Program-region members") {
		test.Fatalf("task boundary: %v", lowerWithOptionsError)
	}
}

// Containers follow the ruled element/key/value membership at concrete allocation sites.
func TestProgramRegionAllocationSiteContainers(test *testing.T) {
	test.Parallel()
	path := filepath.Join(test.TempDir(), "main.a")
	source := `interface Node { children: Node[]; id: number; }
function wrap<T>(item:T):T[] { return [item]; }
function run():void {
 const node:Node = {children:[],id:1};
 const values = new Map<string,Node>(); values.set('node',node);
 const keys = new Map<Node,string>(); keys.set(node,'node');
 const nodes = new Set<Node>(); nodes.add(node);
 const scalar = new Map<string,number>(); scalar.set('one',1);
 const array = wrap<Node>(node);
 const tuple:[Node,number] = [node,1];
 console.log('containers');
} run();`
	if writeError := os.WriteFile(path, []byte(source), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	loaded, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	program, lowerWithOptionsError := LowerWithOptions(context.Background(), loaded, Options{ProgramRegion: true})
	if lowerWithOptionsError != nil {
		test.Fatal(lowerWithOptionsError)
	}
	memberMaps, countedMaps, memberSets, genericArrays, memberTuples := 0, 0, 0, 0, 0
	for _, function := range program.Functions {
		walk(function.Body, func(expression any) bool {
			switch expression := expression.(type) {
			case ir.ObjectLiteral:
				if expression.Tuple && expression.ProgramRegion {
					memberTuples++
				}
			case ir.MapNew:
				if expression.ProgramRegion {
					memberMaps++
				} else {
					countedMaps++
				}
			case ir.SetNew:
				if expression.ProgramRegion {
					memberSets++
				}
			case ir.ArrayLiteral:
				if expression.ProgramRegion && len(expression.Elements) == 1 {
					genericArrays++
				}
			}
			return true
		})
	}
	if memberMaps != 2 || countedMaps != 1 || memberSets != 1 || genericArrays != 1 || memberTuples != 1 {
		test.Fatalf(
			"containers after monomorphization: member Maps=%d counted Maps=%d member Sets=%d generic Node arrays=%d member tuples=%d",
			memberMaps,
			countedMaps,
			memberSets,
			genericArrays,
			memberTuples,
		)
	}
}

func TestProgramRegionEntryPairsRefused(test *testing.T) {
	test.Parallel()
	path := filepath.Join(test.TempDir(), "main.a")
	source := `interface Node {next:Node|undefined;} const node:Node={next:undefined};node.next=node;const items=new Map<Node,string>();items.set(node,'x');const pairs=[...items.entries()];console.log(String(pairs.length));`
	if writeError := os.WriteFile(path, []byte(source), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	loaded, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	_, lowerWithOptionsError := LowerWithOptions(context.Background(), loaded, Options{ProgramRegion: true})
	if lowerWithOptionsError == nil || !strings.Contains(lowerWithOptionsError.Error(), "Map entry pair allocation") {
		test.Fatalf("want unpublished-pair refusal, got %v", lowerWithOptionsError)
	}
}

func programLowerSource(test *testing.T, source string) (*ir.Program, error) {
	test.Helper()
	path := filepath.Join(test.TempDir(), "main.a")
	if writeError := os.WriteFile(path, []byte(source), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	loaded, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	return LowerWithOptions(context.Background(), loaded, Options{ProgramRegion: true})
}

func TestProgramRegionAcyclicLayoutsStayCounted(test *testing.T) {
	test.Parallel()
	program, programLowerSourceError := programLowerSource(
		test,
		`interface Leaf {text:string;} interface Branch {left:Leaf;right:Leaf;}
 interface Graph {text:string;next:Graph|undefined;} function run():void {
 const leaf:Leaf={text:'leaf'};const tree:Branch={left:leaf,right:leaf};
 const graph:Graph={text:'graph',next:undefined};graph.next=graph;
 console.log(tree.left.text+graph.text); }run();`,
	)
	if programLowerSourceError != nil {
		test.Fatal(programLowerSourceError)
	}
	marked := 0
	for _, function := range program.Functions {
		walk(function.Body, func(n any) bool {
			if literal, ok := n.(ir.ObjectLiteral); ok {
				for _, field := range literal.Fields {
					if field.Name == "next" {
						if !literal.ProgramRegion {
							test.Fatal("cyclic static type left counted")
						}
						marked++
					}
					if field.Name == "left" && literal.ProgramRegion {
						test.Fatal("acyclic tree admitted through a structural view")
					}

				}
			}
			return true
		})
	}
	if marked != 1 {
		test.Fatalf("cyclic allocations %d, want 1", marked)
	}
}

func TestProgramRegionAsyncBoundary(test *testing.T) {
	test.Parallel()
	_, programLowerSourceError := programLowerSource(
		test,
		`async function work():Promise<number>{return 1;} console.log(String(await work()));`,
	)
	if programLowerSourceError == nil ||
		!strings.Contains(programLowerSourceError.Error(), "Program region with async work") {
		test.Fatalf("async boundary: %v", programLowerSourceError)
	}
}

func TestProgramCanonicalConstructorBoundary(test *testing.T) {
	test.Parallel()
	program, programLowerSourceError := programLowerSource(
		test,
		"// a-check: refused adamic/cycle-capable\nclass Reader {\n next: Reader | undefined = undefined;\n read: () => string;\n constructor(label: string) {\n  let saved: typeof read | undefined = undefined;\n  this.read = read;\n  const owner: Reader = this;\n  function read(): string { return saved === read && owner.next === owner ? label : \"broken\"; }\n  saved = read;\n  this.next = this;\n  console.log(String(read === read));\n }\n}\nconst first = new Reader(`node${1}`);\nconst held = first.read;\nconsole.log(String(held === first.read));\nconsole.log(held());\nconst second = new Reader(`node${2}`);\nconsole.log(String(held === second.read));\nconsole.log(held());\nconsole.log(second.read());\n",
	)
	if programLowerSourceError != nil {
		test.Fatal(programLowerSourceError)
	}
	if len(program.ProgramTypes) != 0 || len(program.GraphTypes) == 0 {
		test.Fatal("canonical closures must retain the accepted graph fallback before caching")
	}
}

func TestProgramCanonicalCountedConstructorBoundary(test *testing.T) {
	test.Parallel()
	program, programLowerSourceError := programLowerSource(
		test,
		"// a-check: refused adamic/cycle-capable\nclass Reader {\n next: Reader | undefined = undefined;\n read: (suffix?: string) => string;\n constructor(label: string) {\n  let saved: typeof read | undefined = undefined;\n  this.read = read;\n  const owner: Reader = this;\n  function read(suffix?: string): string {\n   return `${label}:${arguments.length}:${saved === read && owner.next === owner}:${suffix === undefined}`;\n  }\n  saved = read;\n  this.next = this;\n  console.log(String(read === read));\n }\n}\nconst first = new Reader(`node${1}`);\nconst held = first.read;\nconsole.log(String(held === first.read));\nconsole.log(held());\nconsole.log(held(\"!\"));\nconst second = new Reader(`node${2}`);\nconsole.log(String(held === second.read));\nconsole.log(held());\nconsole.log(second.read());\n",
	)
	if programLowerSourceError != nil {
		test.Fatal(programLowerSourceError)
	}
	if len(program.ProgramTypes) != 0 || len(program.GraphTypes) == 0 {
		test.Fatal("canonical closures must retain the accepted graph fallback before caching")
	}
}

// Independent acyclic analogues of the K60 and K144 shapes, tested alongside a
// recursive object with overlapping fields. These labels never enter selection.
func TestProgramRegionAcyclicContainerLayouts(test *testing.T) {
	test.Parallel()
	program, programLowerSourceError := programLowerSource(
		test,
		`interface Project {path:string;originalPath?:string;prepend?:boolean;}
 interface Recursive {path:string;next:Recursive|undefined;}
 type Pending = number | readonly [number,number];
 function run():void {
 const project:Project={path:'a'};const projects:Project[]=[project];
 const pending:Pending[]=[1,[2,3]];
 const recursive:Recursive={path:'r',next:undefined};recursive.next=recursive;
 console.log(String(projects.length)+String(pending.length)+recursive.path);
 }run();`,
	)
	if programLowerSourceError != nil {
		test.Fatal(programLowerSourceError)
	}
	counted := 0
	for _, function := range program.Functions {
		walk(function.Body, func(n any) bool {
			if object, ok := n.(ir.ObjectLiteral); ok {
				recursive := false
				for _, field := range object.Fields {
					recursive = recursive || field.Name == "next"
				}
				if !recursive && object.ProgramRegion {
					test.Fatal("acyclic container element was selected")
				}
			}
			if literal, ok := n.(ir.ArrayLiteral); ok {
				if literal.ProgramRegion {
					test.Fatal("acyclic container joined Program region")
				}
				counted++
			}
			return true
		})
	}
	if counted != 2 {
		test.Fatalf("counted containers %d, want 2", counted)
	}
}
