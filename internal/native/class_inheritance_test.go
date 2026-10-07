package native

import (
	"context"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func inheritanceMemoryProgram(t *testing.T) *ir.Program {
	t.Helper()
	checked, err := load.Load([]string{"../oracle/testdata/class_inheritance_memory.a"})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), checked)
	if err != nil {
		t.Fatal(err)
	}
	return program
}

// These are optimization and ownership promises: a hierarchy must not disable unrelated
// regions or reuse, and a call must account for overrides that store the argument.
func TestInheritanceMemoryPlans(t *testing.T) {
	t.Parallel()
	program := inheritanceMemoryProgram(t)
	regions := planRegions(program)
	_, lending := planElementBorrows(program)
	reuse := planReuse(program, lending)
	if len(regions.statements) == 0 {
		t.Fatal("hierarchy lost every region")
	}
	if len(reuse.spreads) == 0 {
		t.Fatal("hierarchy lost every reused spread")
	}
	for _, class := range program.Classes {
		if class.Name == "ChildReader" && !regions.fresh[class.Constructor] {
			t.Error("class allocator lost its region variant")
		}
	}
	checkedEscape, checkedConvention := false, false
	walkExpressions(program, func(expression ir.Expression) {
		call, ok := expression.(ir.Call)
		if !ok || call.Virtual == 0 {
			return
		}
		targets := program.CallTargets(call)
		if len(targets) < 2 {
			return
		}
		names := map[string]bool{}
		for _, target := range targets {
			names[program.Functions[target].Name] = true
		}
		if names["Reader_size"] {
			checkedEscape = true
			if !regions.callEscapes(call, 1) {
				t.Error("storing override lost from virtual escape summary")
			}
		}
		if names["Reader_replace"] {
			checkedConvention = true
			for _, target := range targets {
				parameter := program.Functions[target].Parameters[1]
				if !reuse.consumed[parameter] {
					t.Errorf("%s did not join consumed argument", program.Functions[target].Name)
				}
			}
		}
	})
	if !checkedEscape || !checkedConvention {
		t.Fatalf("fixture missed virtual summaries: escape=%v, consumption=%v", checkedEscape, checkedConvention)
	}
}
