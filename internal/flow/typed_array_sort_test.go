package flow

import (
	"github.com/system-inc/adamic/internal/ir"
	"testing"
)

func TestFlowTypedArraySortUint8(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkTypedArraySortEffects(t, ir.Uint8Array)
	checkAllFlowProgram(t, "../oracle/testdata/typed_arrays_sort_uint8array.a")
}

func TestFlowTypedArraySortUint16(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkTypedArraySortEffects(t, ir.Uint16Array)
	checkAllFlowProgram(t, "../oracle/testdata/typed_arrays_sort_uint16array.a")
}

func TestFlowTypedArraySortInt32(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkTypedArraySortEffects(t, ir.Int32Array)
	checkAllFlowProgram(t, "../oracle/testdata/typed_arrays_sort_int32array.a")
}

func TestFlowTypedArraySortFloat64(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	checkTypedArraySortEffects(t, ir.Float64Array)
	checkAllFlowProgram(t, "../oracle/testdata/typed_arrays_sort_float64array.a")
}

// Runtime output cannot distinguish precise alias metadata from a conservative
// unknown call. Pin both the receiver mutation and the returned identity here.
func checkTypedArraySortEffects(t *testing.T, kind ir.Type) {
	t.Helper()
	program := &ir.Program{
		Locals: []ir.Local{{Type: kind, Function: 0}, {Type: kind, Function: 0}},
		Functions: []ir.Function{{Name: "sort", Returns: kind, Body: []ir.Statement{
			ir.Declare{Local: 0, Value: ir.TypedArrayNew{Of: kind, Source: ir.NumberConstant{Value: 3}}},
			ir.Declare{Local: 1, Value: ir.TypedArraySort{Array: ir.Read{Local: 0, Of: kind}}},
			ir.Return{Value: ir.Read{Local: 1, Of: kind}},
		}}},
	}
	graph := Build(program, 0)
	Construct(graph)
	effects := InferAliasingEffects(graph)
	for _, instruction := range graph.Instructions {
		declaration, ok := (*instruction.At).(ir.Declare)
		if !ok || declaration.Local != 1 {
			continue
		}
		if len(instruction.Uses) != 1 || len(instruction.Defines) != 1 {
			t.Fatal("missing sort operand or result")
		}
		receiver, result := instruction.Uses[0], instruction.Defines[0]
		mutation, alias := false, false
		for _, effect := range effects.Get(instruction.Id) {
			mutation = mutation || effect.Kind == AliasingEffectMutate && effect.Into == receiver
			alias = alias || effect.Kind == AliasingEffectAssign && effect.HasFrom && effect.From == receiver && effect.Into == result
		}
		if !mutation || !alias {
			t.Fatalf("sort must mutate receiver and return its identity: %+v", effects.Get(instruction.Id))
		}
		return
	}
	t.Fatal("missing sort declaration")
}
