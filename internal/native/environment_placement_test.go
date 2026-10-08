package native

import (
	"context"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func TestEnvironmentPlacement(t *testing.T) {
	t.Parallel()
	for _, row := range []struct {
		name      string
		placement environmentPlacement
	}{
		{"direct", environmentStack}, {"callbacks", environmentStack}, {"loop", environmentStack}, {"siblings", environmentStack}, {"local_call", environmentStack}, {"exits", environmentStack}, {"large", environmentRegion},
		{"returned", environmentHeap}, {"field", environmentHeap}, {"array", environmentHeap}, {"map", environmentHeap}, {"set", environmentHeap}, {"global", environmentHeap}, {"capture", environmentHeap}, {"keeping_call", environmentHeap}, {"unknown_call", environmentHeap}, {"callback_escape", environmentHeap}, {"virtual_call", environmentHeap}, {"unknown_callback", environmentHeap},
	} {
		t.Run(row.name, func(t *testing.T) {
			source, err := load.Load([]string{"../oracle/testdata/environment_" + row.name + ".a"})
			if err != nil {
				t.Fatal(err)
			}
			program, err := lower.Lower(context.Background(), source)
			if err != nil {
				t.Fatal(err)
			}
			plan := planRegions(program)
			sites := 0
			for _, function := range program.Functions {
				for index := range function.Body {
					if _, ok := function.Body[index].(ir.AllocateEnvironment); !ok {
						continue
					}
					sites++
					if actual := plan.environments[&function.Body[index]]; actual != row.placement {
						t.Errorf("%s: placement %d, want %d", function.Name, actual, row.placement)
					}
				}
			}
			if sites == 0 {
				t.Fatal("fixture has no explicit environment site")
			}
		})
	}
}

// A thrown carrier is an escape even when source throws only Error objects.
func TestThrownEnvironmentStaysOnHeap(t *testing.T) {
	t.Parallel()
	program := &ir.Program{
		Locals: []ir.Local{{Type: ir.Number, Captured: true, EnvironmentCell: true}},
		Functions: []ir.Function{
			{FrameEnvironment: []int{0}, Body: []ir.Statement{ir.AllocateEnvironment{Cells: []int{0}}, ir.Throw{Value: ir.MakeClosure{Function: 1}}}},
			{Closure: true, Environment: []int{0}},
		},
	}
	plan := planRegions(program)
	if plan.environments[&program.Functions[0].Body[0]] != environmentHeap {
		t.Fatal("throw placed its carrier in the frame")
	}
}

func TestEnvironmentParameterAllTargets(t *testing.T) {
	t.Parallel()
	program := &ir.Program{
		Locals: []ir.Local{{Type: ir.Closure}, {Type: ir.Closure}, {Type: ir.Closure, Global: true}},
		Functions: []ir.Function{
			{Parameters: []int{0}, Body: []ir.Statement{ir.Evaluate{Value: ir.CallClosure{Closure: ir.Read{Local: 0, Of: ir.Closure}, Returns: ir.String}}}},
			{Parameters: []int{1}, Body: []ir.Statement{ir.Assign{Local: 2, Value: ir.Read{Local: 1, Of: ir.Closure}}}},
		},
	}
	plan := planRegions(program)
	if plan.environmentCallEscapes(ir.FunctionTargets{Functions: []int{0}}, 0) {
		t.Error("synchronous helper retains its callback")
	}
	if !plan.environmentCallEscapes(ir.FunctionTargets{Functions: []int{0, 1}}, 0) {
		t.Error("keeping override ignored")
	}
	if !plan.environmentCallEscapes(ir.FunctionTargets{Unknown: true}, 0) {
		t.Error("unknown target treated as synchronous")
	}
}

// An operation added later (for example a pool handoff or async frame transfer)
// must get a lifetime proof before it can borrow an environment carrier.
type unknownEnvironmentTransfer struct{ Value ir.Expression }

func (unknownEnvironmentTransfer) Type() ir.Type { return ir.Closure }

func TestUnknownEnvironmentTransferStaysOnHeap(t *testing.T) {
	t.Parallel()
	program := &ir.Program{
		Locals: []ir.Local{{Type: ir.Number, Captured: true, EnvironmentCell: true}},
		Functions: []ir.Function{
			{FrameEnvironment: []int{0}, Body: []ir.Statement{ir.AllocateEnvironment{Cells: []int{0}}, ir.Evaluate{Value: unknownEnvironmentTransfer{Value: ir.MakeClosure{Function: 1}}}}},
			{Closure: true, Environment: []int{0}},
		},
	}
	plan := planRegions(program)
	if plan.environments[&program.Functions[0].Body[0]] != environmentHeap {
		t.Fatal("unknown handoff placed its carrier in the frame")
	}
}

func TestAsyncEnvironmentStaysCounted(t *testing.T) {
	t.Parallel()
	program := &ir.Program{
		Locals: []ir.Local{{Type: ir.String, Function: 0, Captured: true, EnvironmentCell: true}},
		Functions: []ir.Function{
			{Async: true, FrameEnvironment: []int{0}, Body: []ir.Statement{ir.AllocateEnvironment{Cells: []int{0}}}},
		},
	}
	plan := planRegions(program)
	if plan.environments[&program.Functions[0].Body[0]] != environmentHeap {
		t.Fatal("async cells placed in call storage")
	}
}

func TestAsyncCarrierAndArgumentsEscape(t *testing.T) {
	t.Parallel()
	for _, callback := range []bool{false, true} {
		program := &ir.Program{
			Locals: []ir.Local{{Type: ir.String, Function: 0, Captured: true, EnvironmentCell: true}, {Type: ir.Closure, Function: 1}},
			Functions: []ir.Function{
				{FrameEnvironment: []int{0}, Body: []ir.Statement{ir.AllocateEnvironment{Cells: []int{0}}}},
				{Async: true, Closure: true, Parameters: []int{1}, Environment: []int{0}},
			},
		}
		var call ir.Expression = ir.CallClosure{Closure: ir.MakeClosure{Function: 1}, Returns: ir.Promise}
		if callback {
			call = ir.ArrayVisit{Array: ir.ArrayLiteral{Element: ir.Number}, Callback: ir.MakeClosure{Function: 1}}
		}
		program.Functions[0].Body = append(program.Functions[0].Body, ir.Evaluate{Value: call})
		plan := planRegions(program)
		if plan.environments[&program.Functions[0].Body[0]] != environmentHeap {
			t.Fatal("async carrier placed in call storage")
		}
		if !plan.environmentCallEscapes(ir.FunctionTargets{Functions: []int{1}}, 0) {
			t.Fatal("async argument treated as borrowed")
		}
	}
}
