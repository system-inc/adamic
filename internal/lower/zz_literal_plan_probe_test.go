package lower

// Throwaway probe for #a9e21hc, never to land: each object literal's Record flag, GraphTypes ids and
// whether each id is a graph type, for the programs in $LITERAL_PROBES.

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

func TestZzLiteralPlanProbe(t *testing.T) {
	directory := os.Getenv("LITERAL_PROBES")
	if directory == "" {
		t.Skip("LITERAL_PROBES unset")
	}
	paths, _ := filepath.Glob(filepath.Join(directory, "s7*.a"))
	sort.Strings(paths)
	for _, path := range paths {
		fmt.Printf("\n######## %s\n", filepath.Base(path))
		program, err := load.Load([]string{path})
		if err != nil {
			fmt.Printf("LOAD ERROR: %v\n", err)
			continue
		}
		result, err := Lower(context.Background(), program)
		if err != nil {
			fmt.Printf("LOWER: %v\n", err)
			continue
		}
		fmt.Printf("GraphTypes: %d\n", len(result.GraphTypes))
		inspect := func(node any) bool {
			switch value := node.(type) {
			case ir.ObjectLiteral:
				names := []string{}
				for _, field := range value.Fields {
					names = append(names, field.Name)
				}
				graph := []bool{}
				for _, id := range value.GraphTypes {
					graph = append(graph, result.GraphTypes[id])
				}
				fmt.Printf("ObjectLiteral fields=%v record=%t class=%d graphTypes=%v graph=%v program=%t\n", names, value.Record, value.Class, value.GraphTypes, graph, value.ProgramRegion)
			case ir.ArrayLiteral:
				graph := []bool{}
				for _, id := range value.GraphTypes {
					graph = append(graph, result.GraphTypes[id])
				}
				fmt.Printf("ArrayLiteral elements=%d graphTypes=%v graph=%v\n", len(value.Elements), value.GraphTypes, graph)
			case ir.ArraySlice:
				graph := []bool{}
				for _, id := range value.GraphTypes {
					graph = append(graph, result.GraphTypes[id])
				}
				fmt.Printf("ArraySlice graphTypes=%v graph=%v\n", value.GraphTypes, graph)
			}
			return true
		}
		walk(result.Main, inspect)
		for _, function := range result.Functions {
			walk(function.Body, inspect)
		}
	}
}
