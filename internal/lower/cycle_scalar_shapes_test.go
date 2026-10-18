package lower

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

func TestCycleScalarLiteralShapesOnce(t *testing.T) {
	t.Parallel()
	const elements = 3072
	var source strings.Builder
	source.WriteString("const values = [")
	for i := 0; i < elements; i++ {
		fmt.Fprintf(&source, "{value:%d,done:false},", i)
	}
	source.WriteString("]; console.log(String(values.length));")
	l, _ := lowerCountedImpliedTargets(t, source.String())
	f := &cycleFinder{l: l, where: map[*checker.Type]*ast.Node{}}
	literals := 0
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindObjectLiteralExpression {
			literals++
			f.shape(l.checker.GetTypeAtLocation(node), node)
			if len(f.shapes) != 1 {
				t.Fatalf("scalar literal shapes = %d after %d literals, want 1", len(f.shapes), literals)
			}
		}
		return node.ForEachChild(visit)
	}
	l.program.Files()[0].AsNode().ForEachChild(visit)
	if literals != elements {
		t.Fatalf("literals = %d, want %d", literals, elements)
	}
	// Run the complete cycle analysis too, after pinning the bounded graph size.
	if err := l.findCycles(l.program.Files()); err != nil {
		t.Fatal(err)
	}
}

func TestCycleScalarShapeKeysPreserveDistinctFields(t *testing.T) {
	t.Parallel()
	source := `const a={value:1};const b={other:1};const c={value:"one"};const d={value:1 as const};const e={value:2 as const};const nested={value:{back:1}};const spread={...a};const method={value(){return 1;}};const computed={["value"]:1};const frozen={value:1} as const;console.log(String(a.value));`
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
	l := &lowering{program: program, checker: checked}
	f := &cycleFinder{l: l, where: map[*checker.Type]*ast.Node{}}
	keys := map[string]bool{}
	keyed, unkeyed := 0, 0
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindObjectLiteralExpression {
			proven := l.checker.GetTypeAtLocation(node)
			key, ok := f.scalarShapeKey(proven, node)
			if ok {
				keyed++
				if keys[key] {
					t.Fatalf("distinct fields or literal values shared a shape key: %s (parent %v)", l.checker.TypeToString(proven), node.Parent.Kind)
				}
				keys[key] = true
			} else {
				unkeyed++
			}
		}
		return node.ForEachChild(visit)
	}
	l.program.Files()[0].AsNode().ForEachChild(visit)
	if keyed != 7 || unkeyed != 4 {
		t.Fatalf("keyed/unkeyed = %d/%d, want 7/4", keyed, unkeyed)
	}
}
