package ir_test

import (
	"context"
	"reflect"
	"slices"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func TestCallTargetsIncludeEveryDescendant(t *testing.T) {
	t.Parallel()
	source, err := load.Load([]string{"testdata/call_targets.a"})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), source)
	if err != nil {
		t.Fatal(err)
	}
	var base ir.Class
	for _, class := range program.Classes {
		if class.Name == "Base" {
			base = class
		}
	}
	if len(base.Methods) != 1 {
		t.Fatalf("Base methods: %v", base.Methods)
	}
	targets := program.CallTargets(ir.Call{Function: base.Methods[0], Virtual: 1})
	for _, class := range program.Classes {
		if len(class.Methods) == 1 && !slices.Contains(targets, class.Methods[0]) {
			t.Errorf("missing %s override %d in %v", class.Name, class.Methods[0], targets)
		}
	}
	if len(targets) != 4 {
		t.Fatalf("want four implementations, got %v", targets)
	}
	if got := program.CallTargets(ir.Call{Function: base.Methods[0]}); !reflect.DeepEqual(got, []int{base.Methods[0]}) {
		t.Fatal(got)
	}
	for _, local := range program.Locals {
		if local.Name == "fixed" && local.ConstantClosure == 0 {
			t.Error("const literal lost target proof")
		}
		if (local.Name == "mutable" || local.Name == "namedBinding") && local.ConstantClosure != 0 {
			t.Error("nonliteral or let binding incorrectly bounded")
		}
	}
}

func TestClosureTargetsBoundOnlyProvenValues(t *testing.T) {
	t.Parallel()
	program := &ir.Program{Locals: []ir.Local{{ConstantClosure: 2}, {}}, Functions: []ir.Function{{MayThrow: true}, {}}}
	literal := ir.MakeClosure{Function: 1}
	calls := []ir.Expression{
		ir.CallClosure{Closure: literal}, ir.ArrayMap{Callback: literal}, ir.ArrayVisit{Callback: literal}, ir.ArrayReduce{Callback: literal}, ir.ArrayFrom{Callback: literal}, ir.MapForEach{Callback: literal}, ir.ArraySort{Callback: literal}, ir.ArraySort{Comparator: 1}, ir.CallClosure{Closure: ir.Read{Local: 0}},
	}
	for _, call := range calls {
		got := program.ClosureTargets(call)
		if got.Unknown || !reflect.DeepEqual(got.Functions, []int{1}) {
			t.Errorf("%T: %v", call, got)
		}
		if program.ClosureMayThrow(call) {
			t.Errorf("%T: unrelated throw spoiled bounded target", call)
		}
	}
	unknown := []ir.Expression{ir.Read{Local: 1}, ir.Property{Of: ir.Closure}, ir.Call{Returns: ir.Closure}, ir.Conditional{WhenTrue: literal, WhenNot: literal}}
	for _, value := range unknown {
		call := ir.CallClosure{Closure: value}
		if !program.ClosureTargets(call).Unknown {
			t.Errorf("%T bounded without proof", value)
		}
		if !program.ClosureMayThrow(call) {
			t.Errorf("%T: Unknown failed to include throwing function", value)
		}
	}
}
