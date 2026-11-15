package lower

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// The receiver exemption reads the declared type, never a generic instantiation, a union's
// common member or an intersection's library arm. The source need not lower to test those types.
func TestRegExpServedReceiverExemption(t *testing.T) {
	t.Parallel()
	const source = `function probe<T extends RegExp>(typed: RegExp, union: RegExp | { test?: (text: string) => boolean }, intersection: RegExp & { readonly label: string }, parameter: T, unknown: unknown, any: any): void {}`
	path := filepath.Join(t.TempDir(), "main.ts")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	types, release := program.Checker(t.Context(), program.Files()[0])
	defer release()
	found := map[string]*checker.Type{}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindParameter {
			found[node.Name().Text()] = types.GetTypeAtLocation(node.Name())
		}
		return node.ForEachChild(visit)
	}
	program.Files()[0].AsNode().ForEachChild(visit)
	for _, name := range []string{"typed", "union", "intersection", "parameter", "unknown", "any"} {
		if found[name] == nil {
			t.Fatalf("missing receiver %s", name)
		}
		if got, want := regexpServedReceiverIsRegExp(found[name]), name == "typed"; got != want {
			t.Errorf("%s exempt = %t, want %t", name, got, want)
		}
	}
}

func TestRegExpServedPresenceKeepsHostFrontiers(t *testing.T) {
	t.Parallel()
	for _, row := range []struct {
		name    string
		value   ir.Expression
		unknown bool
		sites   int
	}{
		{"regexp", ir.RegExpNew{}, false, 0},
		{"object", ir.ObjectLiteral{GraphTypes: []int{-1}}, false, 1},
		{"host", ir.NodeFSFile{Operation: "stat", Of: ir.Object}, true, 0},
		{"mixed_host", ir.Conditional{WhenTrue: ir.RegExpNew{}, WhenNot: ir.NodeFSFile{Operation: "stat", Of: ir.Object}, Of: ir.Object}, true, 0},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			graph := newAllocationFlowGraph(&ir.Program{})
			graph.entry = regexpPresenceEntry
			reach := graph.ReachingAllocations(row.value)
			remaining := withoutRegExpPresenceSite(reach.Sites)
			if reach.Unknown != row.unknown || len(remaining) != row.sites {
				t.Fatalf("reach: %+v; remaining %v", reach, remaining)
			}
			if row.name == "regexp" && len(reach.Sites) != 1 {
				t.Fatal("proof edited its memoized sites")
			}
		})
	}
}
