package lower

import (
	"github.com/system-inc/adamic/internal/ir"
	"strings"
	"testing"
)

func TestViewFallbackScopesRetainWiderHelperGuard(t *testing.T) {
	for _, name := range []string{"unrelated never", "viewed never in wider helper", "ordinary only", "unknown callback", "untracked origin"} {
		t.Run(name, func(t *testing.T) {
			object := func() ir.ObjectLiteral {
				return ir.ObjectLiteral{Fields: []ir.Field{{Name: "name", Value: ir.NumberConstant{Value: 42}}}}
			}
			origin := ir.Read{Local: 0, Of: ir.Object}
			program := &ir.Program{
				Locals:        []ir.Local{{Type: ir.Object}, {Type: ir.Object}, {Type: ir.Object}},
				CheckedFields: map[string]bool{"name": true}, ViewOrigins: []ir.Expression{origin}, ViewContractTypes: map[int]ir.ViewContractID{42: 1},
				ViewContracts: []ir.ViewContract{{Kind: ir.ViewScalar, Of: ir.String}, {Kind: ir.ViewUnknown, Name: "never", Unsupported: "never"}, {Kind: ir.ViewObject, Fields: []ir.ViewFieldContract{{Name: "name", Contract: 1}}}, {Kind: ir.ViewObject, Name: "unreachable ArrowFunction", Fields: []ir.ViewFieldContract{{Name: "name", Contract: 2}}}},
				Main:          []ir.Statement{ir.Declare{Local: 0, Value: object()}, ir.Declare{Local: 2, Value: object()}, ir.Evaluate{Value: ir.CheckedCast{Value: origin, ViewContract: 3}}},
				Functions:     []ir.Function{{Parameters: []int{1}, Body: []ir.Statement{ir.Evaluate{Value: ir.Property{Object: ir.Read{Local: 1, Of: ir.Object}, Name: "name", Of: ir.String, ViewTypeID: 42, ViewWhere: "helper.a:9:69"}}}}},
			}
			argument := origin
			if name == "viewed never in wider helper" || name == "ordinary only" {
				program.Main[2] = ir.Evaluate{Value: ir.CheckedCast{Value: origin, ViewContract: 4}}
			}
			if name == "ordinary only" {
				argument = ir.Read{Local: 2, Of: ir.Object}
			}
			program.Main = append(program.Main, ir.Evaluate{Value: ir.Call{Function: 0, Arguments: []ir.Expression{argument}}})
			if name == "unknown callback" {
				program.Main = append(program.Main, ir.Evaluate{Value: ir.MakeClosure{Function: 0}})
			}
			if name == "untracked origin" {
				program.Main[2] = ir.Evaluate{Value: origin}
			}
			err := (&lowering{result: program}).checkLazyViewReads()
			if name == "unrelated never" || name == "ordinary only" {
				if err != nil {
					t.Fatalf("unrelated descriptor poisoned known receiver: %v", err)
				}
			} else if err == nil || !strings.Contains(err.Error(), "field name with unsupported never") || !strings.Contains(err.Error(), "helper.a:9:69") {
				t.Fatalf("wider/unknown helper lost guard: %v", err)
			}
		})
	}
}

func TestViewAggregateNullishConstantsAreNotAllocations(t *testing.T) {
	for _, value := range []ir.Expression{ir.Null{}, ir.Null{Of: ir.Array}, ir.Undefined{}, ir.Undefined{Of: ir.Object}} {
		if viewAggregate(value) {
			t.Fatalf("nullish constant invented aggregate allocation: %#v", value)
		}
	}
	for _, value := range []ir.Expression{ir.Read{Local: 0, Of: ir.Object}, ir.Read{Local: 0, Of: ir.Array}, ir.ObjectLiteral{}, ir.ArrayLiteral{}} {
		if !viewAggregate(value) {
			t.Fatalf("actual or unknown aggregate was dropped: %#v", value)
		}
	}
}
