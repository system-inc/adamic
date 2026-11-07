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

func TestCycleScalarLiteralShapesOnce(test *testing.T) {
	test.Parallel()
	const elements = 3072
	var source strings.Builder
	source.WriteString("const values = [")
	for index := 0; index < elements; index++ {
		fmt.Fprintf(&source, "{value:%d,done:false},", index)
	}
	source.WriteString("]; console.log(String(values.length));")
	lowering, _ := lowerCountedImpliedTargets(test, source.String())
	finder := &cycleFinder{lowering: lowering, where: map[*checker.Type]*ast.Node{}}
	literals := 0
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindObjectLiteralExpression {
			literals++
			finder.shape(lowering.checker.GetTypeAtLocation(node), node)
			if len(finder.shapes) != 1 {
				test.Fatalf("scalar literal shapes = %d after %d literals, want 1", len(finder.shapes), literals)
			}
		}
		return node.ForEachChild(visit)
	}
	lowering.program.Files()[0].AsNode().ForEachChild(visit)
	if literals != elements {
		test.Fatalf("literals = %d, want %d", literals, elements)
	}
	// Run the complete cycle analysis too, after pinning the bounded graph size.
	if findCyclesError := lowering.findCycles(lowering.program.Files()); findCyclesError != nil {
		test.Fatal(findCyclesError)
	}
}

func TestCycleScalarShapeKeysPreserveDistinctFields(test *testing.T) {
	test.Parallel()
	source := `const a={value:1};const b={other:1};const c={value:"one"};const d={value:1 as const};const e={value:2 as const};const nested={value:{back:1}};const spread={...a};const method={value(){return 1;}};const computed={["value"]:1};const frozen={value:1} as const;console.log(String(a.value));`
	path := filepath.Join(test.TempDir(), "main.a")
	if writeError := os.WriteFile(path, []byte(source), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	program, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	checked, release := program.Checker(context.Background(), program.Files()[0])
	test.Cleanup(release)
	lowering := &lowering{program: program, checker: checked}
	finder := &cycleFinder{lowering: lowering, where: map[*checker.Type]*ast.Node{}}
	keys := map[string]bool{}
	keyed, unkeyed := 0, 0
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindObjectLiteralExpression {
			proven := lowering.checker.GetTypeAtLocation(node)
			key, ok := finder.scalarShapeKey(proven, node)
			if ok {
				keyed++
				if keys[key] {
					test.Fatalf(
						"distinct fields or literal values shared a shape key: %s (parent %v)",
						lowering.checker.TypeToString(proven),
						node.Parent.Kind,
					)
				}
				keys[key] = true
			} else {
				unkeyed++
			}
		}
		return node.ForEachChild(visit)
	}
	lowering.program.Files()[0].AsNode().ForEachChild(visit)
	if keyed != 7 || unkeyed != 4 {
		test.Fatalf("keyed/unkeyed = %d/%d, want 7/4", keyed, unkeyed)
	}
}
