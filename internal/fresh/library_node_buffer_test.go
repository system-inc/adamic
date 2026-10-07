package fresh_test

import (
	"github.com/system-inc/adamic/internal/fresh"
	"github.com/system-inc/adamic/internal/ir"
	"testing"
)

func TestNodeBufferHashUpdateKeepsAlias(t *testing.T) {
	t.Parallel()
	program := &ir.Program{
		Strings: []string{"abc"},
		Locals:  []ir.Local{{Type: ir.Object, Function: -1}, {Type: ir.Object, Function: -1}},
		Main: []ir.Statement{
			ir.Declare{Local: 0, Value: ir.NodeBufferCall{Function: "hash_new", Returns: ir.Object}},
			ir.Declare{Local: 1, Value: ir.NodeBufferCall{Function: "hash_update", Returns: ir.Object, Arguments: []ir.Expression{ir.Read{Local: 0, Of: ir.Object}, ir.StringConstant{Index: 0}}}},
			ir.SetProperty{Object: ir.Read{Local: 0, Of: ir.Object}, Name: "next", Value: ir.Read{Local: 1, Of: ir.Object}, Site: 1},
		},
	}
	writes := fresh.ProveWrites(program)
	if len(writes) != 1 || writes[0].Proven || writes[0].Kind != fresh.WriteField {
		t.Fatalf("Hash.update lost its receiver alias: %+v", writes)
	}
}
