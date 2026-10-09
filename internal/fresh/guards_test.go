package fresh_test

import (
	"testing"

	"github.com/system-inc/adamic/internal/fresh"
	"github.com/system-inc/adamic/internal/ir"
)

func requireUnprovenField(t *testing.T, program *ir.Program, function int, name string) {
	t.Helper()
	writes := fresh.ProveWrites(program)
	if len(writes) != 1 || writes[0].Kind != fresh.WriteField || writes[0].Site != 1 || writes[0].Function != function || writes[0].Name != name || writes[0].Proven || writes[0].Why == "" {
		t.Fatalf("want one unproven %s field write at site 1 in function %d, got %+v", name, function, writes)
	}
}

func TestStatAtimeSelfCycleRemainsUnproven(t *testing.T) {
	t.Parallel()
	// Reading the independently allocated atime must preserve its identity.
	program := &ir.Program{
		Locals: []ir.Local{{Type: ir.Object, Function: -1}, {Type: ir.Object, Function: -1}},
		Main: []ir.Statement{
			ir.Declare{Local: 0, Value: ir.NodeFSFile{Operation: "stat", Of: ir.Object}},
			ir.Declare{Local: 1, Value: ir.Property{Object: ir.Read{Local: 0, Of: ir.Object}, Name: "atime", Of: ir.Object}},
			ir.SetProperty{Object: ir.Read{Local: 1, Of: ir.Object}, Name: "next", Value: ir.Read{Local: 1, Of: ir.Object}, Site: 1},
		},
	}
	requireUnprovenField(t, program, -1, "next")
}

func TestReplacementCallbackEscapeRemainsUnproven(t *testing.T) {
	t.Parallel()
	// A replacement callback can keep its operands. A later outside value may
	// reach the escaped holder, so this store must not receive a fresh proof.
	program := &ir.Program{
		Locals: []ir.Local{{Type: ir.Object, Function: 0}, {Type: ir.Object, Function: 0}},
		Functions: []ir.Function{{Name: "callback", Closure: true, Parameters: []int{0}, Body: []ir.Statement{
			ir.Declare{Local: 1, Value: ir.ObjectLiteral{}},
			ir.Evaluate{Value: ir.RegExpCall{Value: ir.Read{Local: 1, Of: ir.Object}, Method: "replace", Returns: ir.String, Replacement: &ir.RegExpReplacement{}}},
			ir.SetProperty{Object: ir.Read{Local: 1, Of: ir.Object}, Name: "saved", Value: ir.Read{Local: 0, Of: ir.Object}, Site: 1},
			ir.Return{},
		}}},
	}
	requireUnprovenField(t, program, 0, "saved")
}

func TestFutureBufferOperationRemainsUnknown(t *testing.T) {
	t.Parallel()
	program := &ir.Program{Main: []ir.Statement{ir.Evaluate{Value: ir.NodeBufferCall{Function: "future_buffer", Returns: ir.Object}}}}
	writes := fresh.ProveWrites(program)
	if len(writes) != 1 || writes[0].Kind != fresh.WriteUnknown || writes[0].Proven || writes[0].Why == "" {
		t.Fatalf("an unrecognized buffer operation must keep its unknown write, got %+v", writes)
	}
}

func TestClassMethodOutsideSelfCycleRemainsUnproven(t *testing.T) {
	t.Parallel()
	// Method calls do not supply the direct-call argument proof. Their parameters
	// remain outside even when no ir.Call to this method appears in the program.
	program := &ir.Program{
		Classes: []ir.Class{{Methods: []int{0}}},
		Locals:  []ir.Local{{Type: ir.Object, Function: 0}},
		Functions: []ir.Function{{Name: "method", Parameters: []int{0}, Body: []ir.Statement{
			ir.SetProperty{Object: ir.Read{Local: 0, Of: ir.Object}, Name: "saved", Value: ir.Read{Local: 0, Of: ir.Object}, Site: 1},
			ir.Return{},
		}}},
	}
	requireUnprovenField(t, program, 0, "saved")
}
