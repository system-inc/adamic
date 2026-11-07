package flow

import (
	"github.com/system-inc/adamic/internal/ir"
	"testing"
)

func TestFlowTypedArraySortUint8(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkTypedArraySortEffects(test, ir.Uint8Array)
	checkAllFlowProgram(test, "../oracle/testdata/typed_arrays_sort_uint8array.a")
}

func TestFlowTypedArraySortUint16(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkTypedArraySortEffects(test, ir.Uint16Array)
	checkAllFlowProgram(test, "../oracle/testdata/typed_arrays_sort_uint16array.a")
}

func TestFlowTypedArraySortInt32(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkTypedArraySortEffects(test, ir.Int32Array)
	checkAllFlowProgram(test, "../oracle/testdata/typed_arrays_sort_int32array.a")
}

func TestFlowTypedArraySortFloat64(test *testing.T) {
	test.Parallel()
	beginFlowUnit(test)
	checkTypedArraySortEffects(test, ir.Float64Array)
	checkAllFlowProgram(test, "../oracle/testdata/typed_arrays_sort_float64array.a")
}

// Runtime output cannot distinguish precise alias metadata from a conservative
// unknown call. Pin both the receiver mutation and the returned identity here.
func checkTypedArraySortEffects(test *testing.T, kind ir.Type) {
	test.Helper()
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
			test.Fatal("missing sort operand or result")
		}
		receiver, result := instruction.Uses[0], instruction.Defines[0]
		mutation, alias := false, false
		for _, effect := range effects.Get(instruction.Id) {
			mutation = mutation || effect.Kind == AliasingEffectMutate && effect.Into == receiver
			alias = alias || effect.Kind == AliasingEffectAssign &&
				effect.HasFrom &&
				effect.From == receiver &&
				effect.Into == result
		}
		if !mutation || !alias {
			test.Fatalf("sort must mutate receiver and return its identity: %+v", effects.Get(instruction.Id))
		}
		return
	}
	test.Fatal("missing sort declaration")
}
