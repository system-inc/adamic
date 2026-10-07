package native

import (
	"context"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func TestBorrowChainDeclarations(t *testing.T) {
	t.Parallel()
	for _, fixture := range []string{"walk", "write", "reassigned", "capture"} {
		loaded, err := load.Load([]string{"../oracle/testdata/borrow_chain_" + fixture + ".a"})
		if err != nil {
			t.Fatal(err)
		}
		program, err := lower.Lower(context.Background(), loaded)
		if err != nil {
			t.Fatal(err)
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
			if program.Functions[local.Function].Name == "inspect" || program.Functions[local.Function].Name == "visit" {
				found[local.Name] = true
			}
		}
		if fixture == "walk" {
			for _, name := range []string{"parent", "kind", "children"} {
				if !found[name] {
					t.Errorf("walk: %s did not borrow", name)
				}
			}
		} else if len(found) != 0 {
			t.Errorf("%s borrowed unsafe declarations: %v", fixture, found)
		}
	}
}

func TestBorrowChainTargets(t *testing.T) {
	t.Parallel()
	program := &ir.Program{Locals: []ir.Local{{Type: ir.Closure}}, Functions: []ir.Function{
		{Body: []ir.Statement{ir.Return{Value: ir.NumberConstant{Value: 1}}}},
		{Body: []ir.Statement{ir.SetProperty{Name: "parent", Value: ir.Undefined{Of: ir.Object}}}},
	}, MethodTargets: map[int][]int{0: {0, 1}}}
	names := map[string]bool{"parent": true}
	if !chainUnchanged(program, program.Functions[0].Body, names) {
		t.Error("read-only body changes a field")
	}
	if chainUnchanged(program, []ir.Statement{ir.Return{Value: ir.Call{Function: 0, Virtual: 1}}}, names) {
		t.Error("ignored writing override")
	}
	if chainUnchanged(program, []ir.Statement{ir.Return{Value: ir.CallClosure{Closure: ir.Read{Of: ir.Closure}}}}, names) {
		t.Error("unknown callback considered read-only")
	}
}
