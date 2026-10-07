package native

import (
	"context"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func TestVisitorParameterConventions(t *testing.T) {
	t.Parallel()
	loaded, err := load.Load([]string{"../oracle/testdata/borrow_visitor.a"})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), loaded)
	if err != nil {
		t.Fatal(err)
	}
	inferParameterBorrows(program)
	checked := 0
	for _, function := range program.Functions {
		if function.Name == "visit" || strings.HasSuffix(function.Name, "_inspect") || strings.HasSuffix(function.Name, "_childNodes") {
			for _, parameter := range function.Parameters {
				if !program.Locals[parameter].Borrowed {
					t.Errorf("%s.%s is owned", function.Name, program.Locals[parameter].Name)
				}
			}
			checked++
		}
	}
	if checked != 4 {
		t.Fatalf("want visitor and three methods, got %d", checked)
	}
}

func TestBorrowConventionJoinsEveryTarget(t *testing.T) {
	t.Parallel()
	program := &ir.Program{
		Locals:        []ir.Local{{Type: ir.Object, Borrowed: true}, {Type: ir.Object, Borrowed: false}},
		Functions:     []ir.Function{{Parameters: []int{0}}, {Parameters: []int{1}}},
		MethodTargets: map[int][]int{0: {0, 1}},
	}
	if program.CallBorrows(ir.Call{Function: 0, Virtual: 1}, 0) {
		t.Fatal("storing override borrowed")
	}
	if !program.CallBorrows(ir.Call{Function: 0}, 0) {
		t.Fatal("direct read-only method inherited override's convention")
	}
	if program.ClosureBorrows(ir.CallClosure{Closure: ir.Read{Local: 0, Of: ir.Closure}}, 0) {
		t.Fatal("unknown function value borrowed")
	}
}

func TestVisitorLoopsBorrow(t *testing.T) {
	t.Parallel()
	loaded, err := load.Load([]string{"../oracle/testdata/borrow_visitor.a"})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), loaded)
	if err != nil {
		t.Fatal(err)
	}
	generated := C(program)
	found := false
	for local, declared := range program.Locals {
		if declared.Name == "child" {
			found = true
			if !declared.Borrowed {
				t.Error("visitor element is owned")
			}
			name := (&emitter{program: program}).localName(local)
			if strings.Contains(generated, "adamic_release("+name+");") || strings.Contains(generated, "adamic_retain("+name+")") {
				t.Error("visitor element is counted")
			}
		}
	}
	if !found {
		t.Fatal("visitor has no child binding")
	}
}

func TestFieldReturnBorrowPlans(t *testing.T) {
	t.Parallel()
	loaded, err := load.Load([]string{"../oracle/testdata/borrow_return.a"})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), loaded)
	if err != nil {
		t.Fatal(err)
	}
	C(program)
	found := map[string]bool{}
	for _, local := range program.Locals {
		if local.Name == "node" && local.Function >= 0 {
			name := program.Functions[local.Function].Name
			found[name] = true
			if local.Borrowed != (name == "onlyRead" || name == "readsView") {
				t.Errorf("%s: borrowed=%v", name, local.Borrowed)
			}
		}
	}
	for _, name := range []string{"onlyRead", "writesField", "keepsResult", "readsView"} {
		if !found[name] {
			t.Errorf("missing %s", name)
		}
	}
}
