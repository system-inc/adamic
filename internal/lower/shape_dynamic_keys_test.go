package lower

import (
	"github.com/system-inc/adamic/internal/ir"
	"reflect"
	"strings"
	"testing"
)

func TestShapeDynamicKeys(t *testing.T) {
	for _, mode := range []string{"finite", "open", "unbounded", "absent", "mutation", "store", "array", "array-mutation", "production-array", "opaque-store", "depth", "spread"} {
		t.Run(mode, func(t *testing.T) {
			first := ir.ObjectLiteral{GraphTypes: []int{-2}}
			second := ir.ObjectLiteral{GraphTypes: []int{-3}}
			receiver := ir.Read{Local: 0, Of: ir.Object}
			key := ir.Read{Local: 1, Of: ir.String}
			program := &ir.Program{Strings: []string{"first", "second", "absent"}, Locals: make([]ir.Local, 3), Main: []ir.Statement{
				ir.Declare{Local: 0, Value: ir.ObjectLiteral{GraphTypes: []int{-1}, Fields: []ir.Field{{Name: "first", Value: first}, {Name: "second", Value: second}}}},
				ir.Declare{Local: 1, Value: ir.Conditional{WhenTrue: ir.StringConstant{Index: 0}, WhenNot: ir.StringConstant{Index: 1}}},
			}}
			unknown := false
			expected := []int{-3, -2}
			switch mode {
			case "open":
				program.Main = append(program.Main, ir.Assign{Local: 1, Value: ir.Read{Local: 2}})
				unknown = true
			case "unbounded":
				program.Main[1] = ir.Declare{Local: 1, Value: ir.Read{Local: 2}}
				unknown = true
			case "absent":
				program.Main = append(program.Main, ir.Assign{Local: 1, Value: ir.StringConstant{Index: 2}})
				unknown = true
			case "mutation":
				program.Main = append(program.Main, ir.Evaluate{Value: shapeDynamicMutation{Value: ir.Read{Local: 2}}})
				unknown = true
			case "opaque-store":
				program.Main = append(program.Main, ir.SetProperty{Object: ir.Read{Local: 2, Of: ir.Object}, Name: "first", Value: ir.ObjectLiteral{GraphTypes: []int{-4}}})
				expected = []int{-4, -3, -2}
				unknown = true
			case "store":
				program.Main = append(program.Main, ir.SetProperty{Object: receiver, Name: "first", Value: ir.ObjectLiteral{GraphTypes: []int{-4}}})
				expected = []int{-4, -3, -2}
			case "array", "array-mutation", "production-array":
				program.Main[0] = ir.Declare{Local: 0, Value: ir.ArrayLiteral{GraphTypes: []int{-1}, Elements: []ir.Expression{first, second}}}
				program.Main[1] = ir.Declare{Local: 1, Value: ir.Conditional{WhenTrue: ir.NumberConstant{Value: 0}, WhenNot: ir.NumberConstant{Value: 1}}}
				if mode == "array-mutation" {
					program.Main = append(program.Main, ir.Evaluate{Value: ir.ArrayPush{Array: receiver}})
					unknown = true
				}
			case "spread":
				literal := program.Main[0].(ir.Declare).Value.(ir.ObjectLiteral)
				literal.Spread = ir.Read{Local: 2}
				program.Main[0] = ir.Declare{Local: 0, Value: literal}
				unknown = true
			}
			var query ir.Expression = shapeDynamicProjection{Receiver: receiver, Key: key}
			if mode == "production-array" {
				query = ir.ArrayIndex{Array: receiver, Index: key, Element: ir.Object}
			}
			graph := newAllocationFlowGraph(program)
			if mode == "depth" {
				graph.ReachingAllocations(query) // Prepare producer facts before exercising the limit.
				sources, reasons, recognized := graph.dynamicProjectionSources(receiver, key, 32)
				if len(sources) != 0 || len(reasons) != 1 || reasons[0] != "recursive dynamic projection depth not certified" || !recognized {
					t.Fatalf("depth lost keys, slots or Unknown: %v %v %t", sources, reasons, recognized)
				}
				return
			}
			got := graph.ReachingAllocations(query)
			// The second query also pins the memoized result.
			got = graph.ReachingAllocations(query)
			if got.Unknown != unknown || !reflect.DeepEqual(got.Sites, expected) {
				t.Fatalf("%s lost keys, slots or Unknown: %+v", mode, got)
			}
			if unknown && len(got.Reasons) == 0 {
				t.Fatal("unnamed Unknown")
			}
			if mode == "open" && !strings.Contains(strings.Join(got.Reasons, ";"), "dynamic key membership not proven") {
				t.Fatalf("missing membership cause: %+v", got)
			}
		})
	}
}
