package lower

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

func TestProgramRegionSCCOwningEdges(t *testing.T) {
	t.Parallel()
	edges := map[string][]string{"left": {"right"}, "right": {"left", "leaf"}, "leaf": {}, "view1": {"view2"}, "view2": {"view1"}}
	owning := map[string][]string{"left": {"right"}, "right": {"leaf"}}
	selected := ProgramRegionSCC(edges, owning)
	if !selected["left"] || !selected["right"] || selected["leaf"] || selected["view1"] || selected["view2"] {
		t.Fatalf("wrong owning SCC: %v", selected)
	}
}
func TestProgramRegionShapeKeySeparators(t *testing.T) {
	t.Parallel()
	if programShapeKey([]string{"a\x00b"}, nil) == programShapeKey([]string{"a", "b"}, nil) {
		t.Fatal("distinct property shapes collide")
	}
	if programShapeKey([]string{"a\x01b"}, nil) == programShapeKey([]string{"a"}, []string{"b\x01"}) {
		t.Fatal("required and available shape boundary collides")
	}
}

func TestProgramRegionAcyclicContainerPayloads(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "payloads.a")
	source := `interface Node { kind:number; next:Node|undefined; }
interface Leaf { path:string; circular:boolean; }
interface Nested { leaf:Leaf; }
interface Backlink { node:Node; }
interface Brand { readonly brand:any; }
interface OpaqueMethod { close():void; }
interface OpaqueValue { value:unknown; }
type PendingId = number & Brand;
type Pending = PendingId | [PendingId] | [PendingId,number];
const leaves:Leaf[]=[];
const nested:Nested[][]=[];
const pending:Pending[]=[];
const nodes:Node[]=[];
const backlinks:Backlink[]=[];
const maps = new Map<Node,Leaf>();
const methods:OpaqueMethod[]=[];
const opaque:OpaqueValue[]=[];`
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
	finder := &cycleFinder{l: &lowering{checker: typeChecker, result: &ir.Program{}}}
	selected := map[cycleNode]bool{}
	containers := map[string]*checker.Type{}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindInterfaceDeclaration {
			declared := typeChecker.GetDeclaredTypeOfSymbol(typeChecker.GetSymbolAtLocation(node.Name()))
			// Model the structural SCC's conservative admission of scalar records too.
			selected[cycleNode{proven: declared}] = true
		}
		if node.Kind == ast.KindVariableDeclaration && node.Name() != nil {
			containers[node.Name().Text()] = typeChecker.GetTypeAtLocation(node.Name())
		}
		return node.ForEachChild(visit)
	}
	entry.AsNode().ForEachChild(visit)
	for name, want := range map[string]bool{"leaves": false, "nested": false, "pending": false, "nodes": true, "backlinks": true, "maps": true, "methods": true, "opaque": true} {
		proven := containers[name]
		if proven == nil {
			t.Fatalf("missing container %s", name)
		}
		if got := finder.programContainerMember(proven, selected); got != want {
			t.Fatalf("container %s member=%t, want %t", name, got, want)
		}
	}
}
