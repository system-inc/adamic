package native

import (
	"context"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func TestBorrowChainDeclarations(test *testing.T) {
	test.Parallel()
	for _, fixture := range []string{"walk", "write", "reassigned", "capture"} {
		loaded, loadError := load.Load([]string{"../oracle/testdata/borrow_chain_" + fixture + ".a"})
		if loadError != nil {
			test.Fatal(loadError)
		}
		program, lowerError := lower.Lower(context.Background(), loaded)
		if lowerError != nil {
			test.Fatal(lowerError)
		}
		plan, _ := planElementBorrows(program)
		found := map[string]bool{}
		for statement := range plan {
			// The plan also holds borrowing loop bindings (ir.ForOf); this test is about declarations.
			declare, ok := (*statement).(ir.Declare)
			if !ok {
				continue
			}
			local := program.Locals[declare.Local]
			if program.Functions[local.Function].Name == "inspect" ||
				program.Functions[local.Function].Name == "visit" {
				found[local.Name] = true
			}
		}
		if fixture == "walk" {
			for _, name := range []string{"parent", "kind", "children"} {
				if !found[name] {
					test.Errorf("walk: %s did not borrow", name)
				}
			}
		} else if len(found) != 0 {
			test.Errorf("%s borrowed unsafe declarations: %v", fixture, found)
		}
	}
}

func TestBorrowChainTargets(test *testing.T) {
	test.Parallel()
	program := &ir.Program{Locals: []ir.Local{{Type: ir.Closure}}, Functions: []ir.Function{
		{Body: []ir.Statement{ir.Return{Value: ir.NumberConstant{Value: 1}}}},
		{Body: []ir.Statement{ir.SetProperty{Name: "parent", Value: ir.Undefined{Of: ir.Object}}}},
	}, MethodTargets: map[int][]int{0: {0, 1}}}
	names := map[string]bool{"parent": true}
	if !chainUnchanged(program, program.Functions[0].Body, names) {
		test.Error("read-only body changes a field")
	}
	if chainUnchanged(program, []ir.Statement{ir.Return{Value: ir.Call{Function: 0, Virtual: 1}}}, names) {
		test.Error("ignored writing override")
	}
	if chainUnchanged(
		program,
		[]ir.Statement{ir.Return{Value: ir.CallClosure{Closure: ir.Read{Of: ir.Closure}}}},
		names,
	) {
		test.Error("unknown callback considered read-only")
	}
}
