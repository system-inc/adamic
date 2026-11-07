package lower

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// The source inventory prevents a newly introduced allocation from escaping the
// traversal probes. Each real allocation type is tested behind each container,
// independently of whether today's lowering puts it there.
func TestGraphAllocationSiteContainers(test *testing.T) {
	test.Parallel()
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
	// Reads: internal/ir/*.go
	files, globError := filepath.Glob("../ir/*.go")
	if globError != nil {
		test.Fatal(globError)
	}
	found := map[string]bool{}
	for _, path := range files {
		file, parseFileError := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if parseFileError != nil {
			test.Fatal(parseFileError)
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
								test.Errorf(
									"IR allocation %s lacks pointer/map traversal coverage",
									declaration.Name.Name,
								)
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
			test.Errorf("stale allocation traversal probe %s", kind.Name())
		}
		test.Run(kind.Name(), func(test *testing.T) {
			value := reflect.New(kind).Elem()
			value.Set(reflect.ValueOf(allocation))
			value.FieldByName("GraphTypes").Set(reflect.ValueOf([]int{17}))
			pointer := reflect.New(kind)
			pointer.Elem().Set(value)
			inputs := map[string]any{
				"pointer":     pointer.Interface(),
				"slice":       []any{value.Interface()},
				"map_value":   map[string]any{"site": value.Interface()},
				"map_pointer": map[string]any{"site": pointer.Interface()},
				"map_key":     map[any]string{pointer.Interface(): "site"},
			}
			for name, input := range inputs {
				test.Run(name, func(test *testing.T) {
					next := 0
					output := graphAllocationSites(reflect.ValueOf(input), &next)
					if next != -1 {
						test.Fatalf("allocation behind %s was skipped: next=%d", name, next)
					}
					if output.Kind() == reflect.Slice {
						output = output.Index(0)
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
						test.Fatalf("allocation IDs = %v, want checker ID and site ID", ids)
					}
					if !reflect.DeepEqual(pointer.Elem().FieldByName("GraphTypes").Interface(), []int{17}) {
						test.Fatal("site walk mutated the input IR")
					}
				})
			}
		})
	}
	for _, input := range []any{(*ir.ObjectLiteral)(nil), map[string]ir.Expression(nil)} {
		next := 0
		output := graphAllocationSites(reflect.ValueOf(input), &next)
		if !output.IsNil() || next != 0 {
			test.Fatalf("nil container acquired a site: %v, %d", output, next)
		}
	}
}

// Graph metadata extends ownership selection, not the view proof's vocabulary.
// These results were Unknown to views before graph regions and must stay so.
func TestGraphMetadataKeepsViewAllocationsUnknown(test *testing.T) {
	test.Parallel()
	for _, expression := range []ir.Expression{
		ir.MapNew{}, ir.SetNew{}, ir.ArrayMap{}, ir.ArrayFrom{}, ir.ArrayFill{},
		ir.ArraySplice{}, ir.ArrayConcat{}, ir.MapEntries{}, ir.ArraySlice{},
		ir.MapKeys{}, ir.MapValues{}, ir.SetValues{}, ir.ArrayVisit{},
	} {
		next := 0
		value := graphAllocationSites(reflect.ValueOf(expression), &next).Interface().(ir.Expression)
		known, unknown := false, false
		graph := newAllocationFlowGraph(&ir.Program{})
		graph.follow(value, func(int) { known = true }, func(int) {}, func(string) { unknown = true })
		if known || !unknown {
			test.Fatalf("graph metadata certified a view allocation: %T", expression)
		}
	}
}
