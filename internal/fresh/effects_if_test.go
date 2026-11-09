package fresh_test

import (
	"testing"

	"github.com/system-inc/adamic/internal/fresh"
	"github.com/system-inc/adamic/internal/ir"
)

// An assignment on one path cannot erase a possible self-cycle on the other.
// An indirect closure is judged here, rather than deferred to absent direct callers.
func TestEffectsIfJoinsPossibleCycles(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name                     string
		initial, then, otherwise int
		proven                   bool
	}{
		{"skipped then retains self", 1, 2, -1, false},
		{"then retains self beside else", 2, 1, 2, false},
		{"both paths retain independent value", 2, 2, 2, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			branch := ir.If{Condition: ir.Read{Local: 0, Of: ir.Boolean}, Then: []ir.Statement{ir.Assign{Local: 3, Value: ir.Read{Local: test.then, Of: ir.Object}}}}
			if test.otherwise >= 0 {
				branch.Else = []ir.Statement{ir.Assign{Local: 3, Value: ir.Read{Local: test.otherwise, Of: ir.Object}}}
			}
			program := &ir.Program{
				Locals: []ir.Local{{Type: ir.Boolean, Function: 0}, {Type: ir.Object, Function: 0}, {Type: ir.Object, Function: 0}, {Type: ir.Object, Function: 0}},
				Functions: []ir.Function{{Name: "probe", Closure: true, Parameters: []int{0}, Body: []ir.Statement{
					ir.Declare{Local: 1, Value: ir.ObjectLiteral{}},
					ir.Declare{Local: 2, Value: ir.ObjectLiteral{}},
					ir.Declare{Local: 3, Value: ir.Read{Local: test.initial, Of: ir.Object}},
					ir.Evaluate{Value: ir.Effects{Body: []ir.Statement{branch}, Result: ir.Undefined{}}},
					ir.SetProperty{Object: ir.Read{Local: 1, Of: ir.Object}, Name: "child", Value: ir.Read{Local: 3, Of: ir.Object}, Site: 1},
				}}},
			}
			writes := fresh.ProveWrites(program)
			if len(writes) != 1 || writes[0].Kind != fresh.WriteField || writes[0].Site != 1 || writes[0].Proven != test.proven {
				t.Fatalf("want one field write, proven=%v; got %+v", test.proven, writes)
			}
		})
	}
}
