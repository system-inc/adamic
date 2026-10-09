package fresh_test

import (
	"testing"

	"github.com/system-inc/adamic/internal/fresh"
	"github.com/system-inc/adamic/internal/ir"
)

func TestArrayHolesResultsAreFresh(t *testing.T) {
	t.Parallel()
	program := &ir.Program{
		Locals: []ir.Local{{Type: ir.Object, Function: 0}, {Type: ir.Array, Function: 0}},
		Functions: []ir.Function{{Name: "holes", Closure: true, Parameters: []int{0}, Body: []ir.Statement{
			ir.Declare{Local: 1, Value: ir.ArrayHoles{Length: ir.NumberConstant{Value: 1}, Element: ir.Object}},
			ir.SetIndex{Array: ir.Read{Local: 1, Of: ir.Array}, Index: ir.NumberConstant{Value: 0}, Value: ir.Read{Local: 0, Of: ir.Object}, Element: ir.Object, Site: 1},
			ir.Return{},
		}}},
	}
	writes := fresh.ProveWrites(program)
	if len(writes) != 1 || writes[0].Kind != fresh.WriteElement || writes[0].Site != 1 || !writes[0].Proven {
		t.Fatalf("empty fresh array should hold an outside value safely: %+v", writes)
	}
}

func TestArrayHolesBackEdgeRemainsUnproven(t *testing.T) {
	t.Parallel()
	program := &ir.Program{
		Locals: []ir.Local{{Type: ir.Array, Function: -1}, {Type: ir.Object, Function: -1}},
		Main: []ir.Statement{
			ir.Declare{Local: 0, Value: ir.ArrayHoles{Length: ir.NumberConstant{Value: 1}, Element: ir.Object}},
			ir.Declare{Local: 1, Value: ir.ObjectLiteral{Fields: []ir.Field{{Name: "back", Value: ir.Read{Local: 0, Of: ir.Array}}}}},
			ir.SetIndex{Array: ir.Read{Local: 0, Of: ir.Array}, Index: ir.NumberConstant{Value: 0}, Value: ir.Read{Local: 1, Of: ir.Object}, Element: ir.Object, Site: 1},
		},
	}
	writes := fresh.ProveWrites(program)
	if len(writes) != 1 || writes[0].Kind != fresh.WriteElement || writes[0].Site != 1 || writes[0].Proven {
		t.Fatalf("a node that points back to its holey array would close a cycle: %+v", writes)
	}
}

func TestArrayHolesOperandsStillJudgeCycles(t *testing.T) {
	t.Parallel()
	closes := ir.ArrayPush{Array: ir.Property{Object: ir.Read{Local: 0, Of: ir.Object}, Name: "nodes", Of: ir.Array}, Value: ir.Read{Local: 0, Of: ir.Object}, Element: ir.Object, Site: 2}
	for name, expression := range map[string]ir.Expression{
		"length":   ir.ArrayHoles{Length: closes, Element: ir.Number},
		"identity": ir.ArrayRangeErrorIs{Value: ir.Comma{Left: closes, Right: ir.Read{Local: 0, Of: ir.Object}}},
	} {
		t.Run(name, func(t *testing.T) {
			program := regexTreeProgram(expression)
			program.Locals = append(program.Locals, ir.Local{Type: ir.Object, Function: -1})
			program.Main = []ir.Statement{
				ir.Declare{Local: 1, Value: ir.ObjectLiteral{Fields: []ir.Field{{Name: "nodes", Value: ir.ArrayLiteral{Element: ir.Object}}}}},
				ir.Evaluate{Value: ir.Call{Function: 0, Arguments: []ir.Expression{ir.Read{Local: 1, Of: ir.Object}}}},
			}
			found := false
			for _, write := range fresh.ProveWrites(program) {
				if write.Kind == fresh.WriteUnknown {
					t.Fatalf("unknown operand: %s", write.Why)
				}
				found = found || write.Site == 2 && write.Kind == fresh.WriteElement && !write.Proven
			}
			if !found {
				t.Fatal("operand evaluation lost its cycle-closing write")
			}
		})
	}
}

func TestArrayRangeErrorIdentityDoesNotEscape(t *testing.T) {
	t.Parallel()
	writes := fresh.ProveWrites(regexTreeProgram(ir.ArrayRangeErrorIs{Value: ir.Read{Local: 0, Of: ir.Object}}))
	if len(writes) != 1 || writes[0].Kind != fresh.WriteField || !writes[0].Proven {
		t.Fatalf("borrowed identity inspection poisoned a safe fresh write: %+v", writes)
	}
}
