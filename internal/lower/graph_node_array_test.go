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
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// unplannedNodeArray has no live road today: tsc's shape, a NodeArray of nodes each naming its parent,
// is refused first as an unchecked cast (#djwd1t2, row 3, Codex 6af4093f). So its words are pinned here
// directly, on graphAdoptionBoundary with synthetic sites: a NodeArray whose type is a graph type
// stops, and so does one that can be held as a selected graph type; one that can't, or a program with
// nothing selected, doesn't.
//
// Mutant: unplannedNodeArray returns nil, and "a graph type" and "held as one" go red.
func TestGraphNodeArrayBoundary(t *testing.T) {
	t.Parallel()
	source := `interface TextRange { pos: number; end: number }
interface Node { kind: number; parent: Node | undefined }
interface NodeArray<T extends Node> extends Array<T>, TextRange { hasTrailingComma: boolean }
declare const nodes: NodeArray<Node>;
declare const plain: Node[];
declare const counts: number[];
console.log(String(nodes.length + plain.length + counts.length));
`
	path := filepath.Join(t.TempDir(), "main.ts")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	checked, release := program.Checker(context.Background(), program.Files()[0])
	t.Cleanup(release)
	declared := map[string]*ast.Node{}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindVariableDeclaration && ast.IsIdentifier(node.Name()) {
			declared[node.Name().Text()] = node.Name()
		}
		return node.ForEachChild(visit)
	}
	program.Files()[0].AsNode().ForEachChild(visit)
	nodesType, plainType, countsType := checked.GetTypeAtLocation(declared["nodes"]), checked.GetTypeAtLocation(declared["plain"]), checked.GetTypeAtLocation(declared["counts"])
	boundary := func(selected []*checker.Type) error {
		l := &lowering{program: program, checker: checked, result: &ir.Program{GraphTypes: map[int]bool{}}}
		l.graphNodeArrays = []graphNodeArraySite{{node: declared["nodes"], array: nodesType}}
		for _, each := range selected {
			l.result.GraphTypes[l.cycleTypeID(each)] = true
		}
		return (&cycleFinder{l: l, selected: selected}).graphAdoptionBoundary()
	}
	stops := func(name string, err error) {
		var notYet *NotYet
		if !errors.As(err, &notYet) || !strings.Contains(notYet.What, "a NodeArray on a cycle in a graph program") {
			t.Errorf("%s: got %v, want the NodeArray stop", name, err)
		}
	}
	stops("a graph type", boundary([]*checker.Type{nodesType}))
	stops("held as one", boundary([]*checker.Type{plainType}))
	if err := boundary([]*checker.Type{countsType}); err != nil {
		t.Errorf("not assignable: got %v, want nil", err)
	}
	if err := boundary(nil); err != nil {
		t.Errorf("nothing selected: got %v, want nil", err)
	}
}
