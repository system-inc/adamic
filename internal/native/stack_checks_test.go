package native

import (
	"reflect"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// Exhaust all three-vertex directed graphs and compare to transitive closure,
// which does not use component numbers, low links or Tarjan's active stack.
func TestRecursiveStackComponentsAgainstReachability(t *testing.T) {
	t.Parallel()
	for edges := 0; edges < 1<<9; edges++ {
		graph := make([][]int, 3)
		var reaches [3][3]bool
		for from := range graph {
			for to := range graph {
				if edges&(1<<(from*3+to)) != 0 {
					graph[from] = append(graph[from], to)
					reaches[from][to] = true
				}
			}
		}
		for via := range graph {
			for from := range graph {
				for to := range graph {
					reaches[from][to] = reaches[from][to] || reaches[from][via] && reaches[via][to]
				}
			}
		}
		got := recursiveStackComponents(graph)
		for function := range graph {
			if got[function] != reaches[function][function] {
				t.Fatalf("edges=%09b function=%d: candidate=%t recursive=%t", edges, function, got[function], reaches[function][function])
			}
		}
	}
}

func TestStackCheckCallTargets(t *testing.T) {
	t.Parallel()
	evaluate := func(expression ir.Expression) ir.Statement { return ir.Evaluate{Value: expression} }
	for _, test := range []struct {
		name    string
		program ir.Program
		want    []bool
	}{
		{"acyclic descendants", ir.Program{Functions: []ir.Function{
			{Body: []ir.Statement{evaluate(ir.Call{Function: 1})}},
			{Body: []ir.Statement{evaluate(ir.Call{Function: 0}), evaluate(ir.Call{Function: 2})}},
			{Body: []ir.Statement{evaluate(ir.Call{Function: 3})}}, {},
		}}, []bool{true, true, false, false}},
		{"virtual implementation", ir.Program{MethodTargets: map[int][]int{1: {1, 2}}, Functions: []ir.Function{
			{Body: []ir.Statement{evaluate(ir.Call{Function: 1, Virtual: 1})}}, {},
			{Body: []ir.Statement{evaluate(ir.Call{Function: 0})}},
		}}, []bool{true, false, true}},
		{"closure literal", ir.Program{Functions: []ir.Function{
			{Body: []ir.Statement{evaluate(ir.CallClosure{Closure: ir.MakeClosure{Function: 1}})}},
			{Closure: true, Body: []ir.Statement{evaluate(ir.Call{Function: 0})}}, {},
		}}, []bool{true, true, false}},
		{"constant closure", ir.Program{Locals: []ir.Local{{ConstantClosure: 2}}, Functions: []ir.Function{
			{Body: []ir.Statement{evaluate(ir.CallClosure{Closure: ir.Read{Local: 0, Of: ir.Closure}})}},
			{Closure: true, Body: []ir.Statement{evaluate(ir.Call{Function: 0})}}, {},
		}}, []bool{true, true, false}},
		{"unresolved top level", ir.Program{Main: []ir.Statement{evaluate(ir.CallClosure{Closure: ir.Read{Of: ir.Closure}})}, Locals: []ir.Local{{}}, Functions: []ir.Function{{}, {}, {}}}, []bool{true, true, true}},
		{"nested exceptional body", ir.Program{Functions: []ir.Function{
			{Body: []ir.Statement{ir.Try{Finally: []ir.Statement{ir.If{Then: []ir.Statement{evaluate(ir.Call{Function: 0})}}}}}}, {},
		}}, []bool{true, false}},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := stackCheckCandidates(&test.program); !reflect.DeepEqual(got, test.want) {
				t.Fatalf("got %v, want %v", got, test.want)
			}
		})
	}
}

func TestStackCheckBuiltinCallbacks(t *testing.T) {
	t.Parallel()
	for _, callback := range []ir.Expression{
		ir.ArrayMap{Callback: ir.MakeClosure{Function: 1}},
		ir.ArrayVisit{Callback: ir.MakeClosure{Function: 1}},
		ir.ArrayReduce{Callback: ir.MakeClosure{Function: 1}},
		ir.ArrayFrom{Callback: ir.MakeClosure{Function: 1}},
		ir.MapForEach{Callback: ir.MakeClosure{Function: 1}},
		ir.ArraySort{Callback: ir.MakeClosure{Function: 1}},
		ir.ArraySort{Comparator: 1},
	} {
		program := &ir.Program{Functions: []ir.Function{
			{Body: []ir.Statement{ir.Evaluate{Value: callback}}},
			{Body: []ir.Statement{ir.Evaluate{Value: ir.Call{Function: 0}}}}, {},
		}}
		if got := stackCheckCandidates(program); !reflect.DeepEqual(got, []bool{true, true, false}) {
			t.Fatalf("%T: got %v", callback, got)
		}
	}
	program := &ir.Program{Locals: []ir.Local{{}}, Functions: []ir.Function{
		{Body: []ir.Statement{ir.Evaluate{Value: ir.ArrayMap{Callback: ir.Read{Local: 0, Of: ir.Closure}}}}}, {}, {},
	}}
	if got := stackCheckCandidates(program); !reflect.DeepEqual(got, []bool{true, true, true}) {
		t.Fatalf("unresolved map callback: got %v", got)
	}
}

func TestStackCheckExactInterfaceReceiver(t *testing.T) {
	t.Parallel()
	program := &ir.Program{
		Locals:  []ir.Local{{Type: ir.Object}, {Type: ir.Object}},
		Classes: []ir.Class{{Constructor: 0}},
		Main:    []ir.Statement{ir.Declare{Local: 0, Value: ir.Call{Function: 0, Returns: ir.Object}}},
		Functions: []ir.Function{
			{Body: []ir.Statement{ir.Declare{Local: 1, Value: ir.ObjectLiteral{Class: 1, Methods: []ir.Method{{Name: "run", Function: 1}}}}}},
			{Body: []ir.Statement{ir.Evaluate{Value: ir.CallClosure{Closure: ir.Property{Object: ir.Read{Local: 0, Of: ir.Object}, Name: "run", Method: true}}}}}, {},
		},
	}
	if got := stackCheckCandidates(program); !reflect.DeepEqual(got, []bool{false, true, false}) {
		t.Fatalf("exact method: got %v", got)
	}
	program.Main = append(program.Main, ir.Assign{Local: 0, Value: ir.Undefined{Of: ir.Object}})
	if got := stackCheckCandidates(program); !reflect.DeepEqual(got, []bool{true, true, true}) {
		t.Fatalf("assigned interface receiver: got %v", got)
	}
}
