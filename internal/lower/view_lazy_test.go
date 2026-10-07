package lower

import (
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestLazyViewDemandUsesSharedFlow(t *testing.T) {
	for _, name := range []string{"helper merge", "wider helper", "ordinary only", "callback unknown", "missing producer"} {
		t.Run(name, func(t *testing.T) {
			object := func() ir.ObjectLiteral {
				return ir.ObjectLiteral{Fields: []ir.Field{{Name: "name", Value: ir.NumberConstant{Value: 1}}}}
			}
			program := &ir.Program{
				Locals:            []ir.Local{{Type: ir.Object}, {Type: ir.Object}, {Type: ir.Object}},
				CheckedFields:     map[string]bool{"opaque": true},
				ViewContracts:     []ir.ViewContract{{Kind: ir.ViewCallable, Unsupported: "callable"}, {Kind: ir.ViewObject, Fields: []ir.ViewFieldContract{{Name: "opaque", Contract: 1}}}},
				ViewContractTypes: map[int]ir.ViewContractID{42: 1},
				ViewOrigins:       []ir.Expression{ir.Read{Local: 0, Of: ir.Object}},
				Main:              []ir.Statement{ir.Declare{Local: 0, Value: object()}, ir.Declare{Local: 2, Value: object()}},
				Functions:         []ir.Function{{Parameters: []int{1}, Body: []ir.Statement{ir.Evaluate{Value: ir.Property{Object: ir.Read{Local: 1, Of: ir.Object}, Name: "opaque", Of: ir.Closure, ViewTypeID: 42, ViewWhere: "helper.a:9:69"}}}}},
			}
			// All variants supply an ordinary value of the same declared type.
			program.Main = append(program.Main, ir.Evaluate{Value: ir.Call{Function: 0, Arguments: []ir.Expression{ir.Read{Local: 2, Of: ir.Object}}}})
			switch name {
			case "helper merge", "wider helper":
				if name == "wider helper" {
					program.Functions[0].Body[0] = ir.Evaluate{Value: ir.Property{Object: ir.Read{Local: 1, Of: ir.Object}, Name: "opaque", Of: ir.Closure, ViewTypeID: 43, ViewWhere: "helper.a:9:69"}}
				}
				program.Main = append(program.Main, ir.Evaluate{Value: ir.Call{Function: 0, Arguments: []ir.Expression{ir.Read{Local: 0, Of: ir.Object}}}})
			case "callback unknown":
				program.Main = append(program.Main, ir.Evaluate{Value: ir.MakeClosure{Function: 0}})
			case "missing producer":
				program.ViewOrigins = []ir.Expression{ir.Read{Local: 1, Of: ir.Object}}
				program.Main = program.Main[:2]
			}
			l := &lowering{result: program}
			err := l.checkLazyViewReads()
			if name == "ordinary only" {
				if err != nil {
					t.Fatalf("disjoint allocations refused: %v", err)
				}
			} else if err == nil || !strings.Contains(err.Error(), "field opaque with unsupported callable") || !strings.Contains(err.Error(), "helper.a:9:69") {
				t.Fatalf("want refusal at demanded read, got %v", err)
			}
		})
	}
}

func TestLazyViewArrayDemand(t *testing.T) {
	for _, method := range []string{"index", "map", "visit", "reduce", "pop", "join", "for of"} {
		t.Run(method, func(t *testing.T) {
			array := ir.Read{Local: 0, Of: ir.Array}
			metadata := ir.ArrayViewRead{ViewTypeID: 42, View: "values[element]", Element: ir.Closure}
			var consumer ir.Statement
			switch method {
			case "index":
				consumer = ir.Evaluate{Value: ir.ArrayIndex{Array: array, Element: ir.Closure, ViewTypeID: 42, View: "values[0]"}}
			case "map":
				consumer = ir.Evaluate{Value: ir.ArrayMap{Array: array, ViewRead: metadata}}
			case "visit":
				consumer = ir.Evaluate{Value: ir.ArrayVisit{Array: array, ViewRead: metadata}}
			case "reduce":
				consumer = ir.Evaluate{Value: ir.ArrayReduce{Array: array, ViewRead: metadata}}
			case "pop":
				consumer = ir.Evaluate{Value: ir.ArrayPop{Array: array, ViewRead: metadata}}
			case "join":
				consumer = ir.Evaluate{Value: ir.ArrayJoin{Array: array, ViewRead: metadata}}
			case "for of":
				consumer = ir.ForOf{Iterable: array, ViewRead: metadata}
			}
			program := &ir.Program{Locals: []ir.Local{{Type: ir.Array}}, ViewOrigins: []ir.Expression{array}, ViewContractTypes: map[int]ir.ViewContractID{42: 1}, ViewContracts: []ir.ViewContract{{Kind: ir.ViewCallable, Unsupported: "callable"}}, Main: []ir.Statement{ir.Declare{Local: 0, Value: ir.ArrayLiteral{Element: ir.Closure}}, consumer}}
			err := (&lowering{result: program}).checkLazyViewReads()
			if err == nil || !strings.Contains(err.Error(), "field [element] with unsupported callable") {
				t.Fatalf("want named element refusal, got %v", err)
			}
		})
	}
}
