package lower

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// The source inventory prevents a newly introduced allocation from escaping the
// traversal probes. Each real allocation type is tested behind each container,
// independently of whether today's lowering puts it there.
func TestGraphAllocationSiteContainers(t *testing.T) {
	allocations := []ir.Expression{
		ir.ObjectLiteral{}, ir.ArrayLiteral{}, ir.MapNew{}, ir.SetNew{},
		ir.ArrayMap{}, ir.ArrayFrom{}, ir.ArrayFill{}, ir.ArraySplice{},
		ir.ArrayConcat{}, ir.MapEntries{}, ir.ArraySlice{}, ir.MapKeys{},
		ir.MapValues{}, ir.SetValues{}, ir.ArrayVisit{},
	}
	covered := map[string]bool{}
	for _, allocation := range allocations {
		covered[reflect.TypeOf(allocation).Name()] = true
	}
	files, err := filepath.Glob("../ir/*.go")
	if err != nil {
		t.Fatal(err)
	}
	found := map[string]bool{}
	for _, path := range files {
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(file, func(node ast.Node) bool {
			declaration, ok := node.(*ast.TypeSpec)
			if !ok {
				return true
			}
			structure, ok := declaration.Type.(*ast.StructType)
			if !ok {
				return true
			}
			for _, field := range structure.Fields.List {
				for _, name := range field.Names {
					if name.Name == "GraphTypes" {
						// Program's ownership index is not an allocation site.
						if declaration.Name.Name != "Program" {
							found[declaration.Name.Name] = true
							if !covered[declaration.Name.Name] {
								t.Errorf("IR allocation %s lacks pointer/map traversal coverage", declaration.Name.Name)
							}
						}
					}
				}
			}
			return true
		})
	}
	for _, allocation := range allocations {
		kind := reflect.TypeOf(allocation)
		if !found[kind.Name()] {
			t.Errorf("stale allocation traversal probe %s", kind.Name())
		}
		t.Run(kind.Name(), func(t *testing.T) {
			value := reflect.New(kind).Elem()
			value.Set(reflect.ValueOf(allocation))
			value.FieldByName("GraphTypes").Set(reflect.ValueOf([]int{17}))
			pointer := reflect.New(kind)
			pointer.Elem().Set(value)
			inputs := map[string]any{
				"pointer":     pointer.Interface(),
				"map_value":   map[string]any{"site": value.Interface()},
				"map_pointer": map[string]any{"site": pointer.Interface()},
				"map_key":     map[any]string{pointer.Interface(): "site"},
			}
			for name, input := range inputs {
				t.Run(name, func(t *testing.T) {
					next := 0
					output := graphAllocationSites(reflect.ValueOf(input), &next)
					if next != -1 {
						t.Fatalf("allocation behind %s was skipped: next=%d", name, next)
					}
					if output.Kind() == reflect.Map {
						entries := output.MapRange()
						entries.Next()
						if name == "map_key" {
							output = entries.Key()
						} else {
							output = entries.Value()
						}
					}
					for output.Kind() == reflect.Ptr || output.Kind() == reflect.Interface {
						output = output.Elem()
					}
					ids := output.FieldByName("GraphTypes").Interface().([]int)
					if !reflect.DeepEqual(ids, []int{17, -1}) {
						t.Fatalf("allocation IDs = %v, want checker ID and site ID", ids)
					}
					if !reflect.DeepEqual(pointer.Elem().FieldByName("GraphTypes").Interface(), []int{17}) {
						t.Fatal("site walk mutated the input IR")
					}
				})
			}
		})
	}
	for _, input := range []any{(*ir.ObjectLiteral)(nil), map[string]ir.Expression(nil)} {
		next := 0
		output := graphAllocationSites(reflect.ValueOf(input), &next)
		if !output.IsNil() || next != 0 {
			t.Fatalf("nil container acquired a site: %v, %d", output, next)
		}
	}
}

// Load/check is outside the timer. The same warmed checker is reused for both
// the ordinary build and the scratch overlay that removes only the site walk.
func BenchmarkGraphLowering(b *testing.B) {
	path := os.Getenv("ADAMIC_GRAPH_BENCH_SOURCE")
	if path == "" {
		path = "../../stage1/typescript/parser/main.ts"
	}
	path, err := filepath.Abs(path)
	if err != nil {
		b.Fatal(err)
	}
	var checked *load.Program
	if os.Getenv("ADAMIC_GRAPH_BENCH_CYCLE") == "1" {
		// These stage-1 programs prove their cycles. Append one real unproven
		// self link to measure the walk when the graph pass is enabled.
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			b.Fatal(readErr)
		}
		source := string(data) + `
interface GraphWalkNode { next: GraphWalkNode | undefined }
function graphWalkMake() { return { next: undefined as GraphWalkNode | undefined }; }
function graphWalkCycle(): GraphWalkNode {
    const node: GraphWalkNode = graphWalkMake();
    node.next = node;
    return node;
}
const graphWalkNode = graphWalkCycle();
console.log(graphWalkNode.next === graphWalkNode ? 'cycle' : 'missing');
`
		checked, err = load.LoadOverlay([]string{path}, map[string]string{path: source})
	} else {
		checked, err = load.Load([]string{path})
	}
	if err != nil {
		b.Fatal(err)
	}
	if os.Getenv("ADAMIC_GRAPH_BENCH_TSGO") == "1" {
		checked.EnableTSGo()
	}
	program, err := Lower(context.Background(), checked)
	if err != nil {
		b.Fatal(err)
	}
	if os.Getenv("ADAMIC_GRAPH_BENCH_CYCLE") == "1" && len(program.GraphTypes) == 0 {
		b.Fatal("benchmark graph component did not enable graph flow")
	}
	b.Logf("input=%s functions=%d locals=%d graph IDs=%d", path, len(program.Functions), len(program.Locals), len(program.GraphTypes))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Lower(context.Background(), checked); err != nil {
			b.Fatal(err)
		}
	}
}

// Isolate site assignment from checker/SCC work and total-lowering timing noise.
func BenchmarkGraphAllocationSiteWalk(b *testing.B) {
	path := os.Getenv("ADAMIC_GRAPH_BENCH_SOURCE")
	if path == "" {
		path = "../../stage1/typescript/parser/main.ts"
	}
	checked, err := load.Load([]string{path})
	if err != nil {
		b.Fatal(err)
	}
	if os.Getenv("ADAMIC_GRAPH_BENCH_TSGO") == "1" {
		checked.EnableTSGo()
	}
	program, err := Lower(context.Background(), checked)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		next := 0
		graphAllocationSites(reflect.ValueOf(program.Main), &next)
		for _, function := range program.Functions {
			graphAllocationSites(reflect.ValueOf(function.Body), &next)
		}
	}
}
