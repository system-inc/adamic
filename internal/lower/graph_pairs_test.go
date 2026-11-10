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

// unplannedPair has no live road (internal/oracle/map_entries_pairs_test.go: every pair held as
// another type tried has its own type selected), so its words are pinned here directly, on
// graphAdoptionBoundary with synthetic sites (#dps0kp6, Runtime's condition): a pair whose type isn't a
// graph type but is assignable to one stops, a pair the checker didn't name stops, and a planned pair,
// or a program with nothing selected, doesn't.
//
// Mutant: unplannedPair returns nil, and "held as another type" and "unnamed" go red.
func TestGraphPairBoundary(t *testing.T) {
	t.Parallel()
	source := `interface Node { name: string; next: Node[] }
const node: Node = { name: 'a' + '', next: [] };
const pair: [string, Node] = ['a', node];
const held: readonly [string, Node] = pair;
console.log(held[0]);
`
	path := filepath.Join(t.TempDir(), "main.a")
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
	pairType, heldType := checked.GetTypeAtLocation(declared["pair"]), checked.GetTypeAtLocation(declared["held"])
	if pairType == heldType {
		t.Fatal("the readonly tuple is the pair's own type: the test needs two")
	}
	boundary := func(selected []*checker.Type, planned bool, sites ...graphPairSite) error {
		l := &lowering{program: program, checker: checked, result: &ir.Program{GraphTypes: map[int]bool{}}, graphPairs: sites}
		for _, each := range selected {
			l.result.GraphTypes[int(each.Id())] = true
		}
		if planned {
			l.result.GraphTypes[l.cycleTypeID(pairType)] = true
		}
		return (&cycleFinder{l: l, selected: selected}).graphAdoptionBoundary()
	}
	stops := func(name string, err error, want string) {
		var notYet *NotYet
		if !errors.As(err, &notYet) || !strings.Contains(notYet.What, want) {
			t.Errorf("%s: got %v, want NotYet naming %q", name, err, want)
		}
	}
	site := graphPairSite{node: declared["pair"], pair: pairType}
	stops("held as another type", boundary([]*checker.Type{heldType}, false, site), "pair held as another graph type in a graph program")
	stops("unnamed", boundary([]*checker.Type{heldType}, false, graphPairSite{node: declared["pair"]}), "pair in a graph program, whose type the checker didn't name")
	if err := boundary([]*checker.Type{heldType}, true, site); err != nil {
		t.Errorf("planned: got %v, want nil", err)
	}
	if err := boundary(nil, false, site); err != nil {
		t.Errorf("nothing selected: got %v, want nil", err)
	}
}
