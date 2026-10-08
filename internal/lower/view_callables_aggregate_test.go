package lower

import (
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// A returned allocation is demanded even when the closure owner's flow cannot
// be bounded. It must not disappear simply because it is not a stored field of
// the viewed record. The ordinary-call control stays disjoint.
func TestViewCallableAggregateReturnedDemand(t *testing.T) {
	for _, callable := range []bool{false, true} {
		t.Run(map[bool]string{false: "ordinary-disjoint", true: "unknown-callable-owner"}[callable], func(t *testing.T) {
			program := &ir.Program{
				Locals:            []ir.Local{{Type: ir.Object}, {Type: ir.Closure, ConstantClosure: 1}, {Type: ir.Object}},
				ViewOrigins:       []ir.Expression{ir.Read{Local: 0, Of: ir.Object}},
				CheckedFields:     map[string]bool{"opaque": true},
				ViewContracts:     []ir.ViewContract{{Kind: ir.ViewUnknown, Unsupported: "dictionary"}},
				ViewContractTypes: map[int]ir.ViewContractID{42: 1},
				Functions:         []ir.Function{{Returns: ir.Object, Body: []ir.Statement{ir.Return{Value: ir.ObjectLiteral{Fields: []ir.Field{{Name: "opaque", Value: ir.NumberConstant{Value: 1}}}}}}}},
			}
			var call ir.Expression = ir.Call{Function: 0, Returns: ir.Object}
			if callable {
				call = ir.CallClosure{Closure: ir.Read{Local: 1, Of: ir.Closure}, Returns: ir.Object}
			}
			program.Main = []ir.Statement{
				ir.Declare{Local: 0, Value: ir.ObjectLiteral{}},
				ir.Declare{Local: 2, Value: call},
				ir.Evaluate{Value: ir.Property{Object: ir.Read{Local: 2, Of: ir.Object}, Name: "opaque", Of: ir.Number, ViewTypeID: 42, ViewWhere: "returned.a:12:3"}},
			}
			err := (&lowering{result: program}).checkLazyViewReads()
			if !callable {
				if err != nil {
					t.Fatalf("ordinary disjoint allocation refused: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), "field opaque with unsupported dictionary") || !strings.Contains(err.Error(), "returned.a:12:3") {
				t.Fatalf("returned allocation lost demanded refusal: %v", err)
			}
		})
	}
}
