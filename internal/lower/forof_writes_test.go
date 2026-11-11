package lower

import (
	"slices"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// A for-of writes its binding, or each of its pattern's, on every step, and a catch writes its binding
// with what was thrown, with no Declare or Assign the flow graph collects. Where the body also
// assigns the binding, the graph saw only that assignment, so a reassigned loop variable was proven
// the literal it's reassigned and class_accessors read a getter as a plain field (Language's must on
// #p2n17c4: on main 9ac2dc52 the JavaScript backend printed undefined then 1, where Node prints 42
// then 1, and native panicked). Each such binding is now an unknown write. forof_writes.mutants.md
// names the edit that turns each back.

const forOfWritesBox = `class Box {
	get weight(): number {
		return 42;
	}
}
`

// forOfWritesLower lowers source as main.ts and fails unless it lowers.
func forOfWritesLower(t *testing.T, source string) *ir.Program {
	t.Helper()
	program, err := lowerSourceAs(t, source, ".ts")
	if err != nil {
		t.Fatalf("lowering: %v", err)
	}
	return program
}

// forOfWritesReach is what the flow graph says the one local named name can hold.
func forOfWritesReach(t *testing.T, program *ir.Program, name string) ir.AllocationSet {
	t.Helper()
	local := -1
	for index, declared := range program.Locals {
		if declared.Name != name {
			continue
		}
		if local >= 0 {
			t.Fatalf("two locals are named %s", name)
		}
		local = index
	}
	if local < 0 {
		t.Fatalf("no local is named %s", name)
	}
	assignAllocationSites(program)
	return newAllocationFlowGraph(program).ReachingAllocations(ir.Read{Local: local, Of: program.Locals[local].Type})
}

// forOfWritesDispatched fails unless every read of weight in Main goes through the getter's
// dispatcher, never a plain field load, and at least one does.
func forOfWritesDispatched(t *testing.T, program *ir.Program) {
	t.Helper()
	dispatched := 0
	walk(program.Main, func(node any) bool {
		switch value := node.(type) {
		case ir.Property:
			if value.Name == "weight" {
				t.Errorf("Main reads weight as a plain field: %#v", value)
			}
		case ir.Call:
			if value.Function >= 0 && value.Function < len(program.Functions) && program.Functions[value.Function].Name == "property_get:weight" {
				dispatched++
			}
		}
		return true
	})
	if dispatched == 0 {
		t.Error("Main reads weight through no getter dispatcher")
	}
}

// r1: a let binding reassigned in the loop's body is Unknown, so both reads of weight dispatch the
// getter. Mutant 1.
func TestForOfReassignedBindingIsUnknown(t *testing.T) {
	t.Parallel()
	program := forOfWritesLower(t, forOfWritesBox+`const items: Box[] = [new Box()];
for (let item of items) {
	console.log(String(item.weight));
	item = { weight: 1 };
	console.log(String(item.weight));
}
`)
	reach := forOfWritesReach(t, program, "item")
	if !reach.Unknown || !slices.Contains(reach.Reasons, "a for-of binding takes each element") {
		t.Fatalf("reach is %+v, want Unknown for a for-of binding", reach)
	}
	forOfWritesDispatched(t, program)
}

// r2: the same through a pattern, each of whose locals the step writes. Mutant 1.
func TestForOfReassignedPatternBindingIsUnknown(t *testing.T) {
	t.Parallel()
	program := forOfWritesLower(t, forOfWritesBox+`const pairs: [Box, number][] = [[new Box(), 0]];
for (let [item, index] of pairs) {
	console.log(String(item.weight + index));
	item = { weight: 1 };
	console.log(String(item.weight + index));
}
`)
	reach := forOfWritesReach(t, program, "item")
	if !reach.Unknown || !slices.Contains(reach.Reasons, "a for-of binding takes each element") {
		t.Fatalf("reach is %+v, want Unknown for a for-of pattern's binding", reach)
	}
	forOfWritesDispatched(t, program)
}

// r0: a const binding nothing reassigns was Unknown before, with no producer at all, and still
// lowers to the getter's dispatcher.
func TestForOfConstBindingStillDispatches(t *testing.T) {
	t.Parallel()
	program := forOfWritesLower(t, forOfWritesBox+`const items: Box[] = [new Box()];
for (const item of items) {
	console.log(String(item.weight));
}
`)
	if reach := forOfWritesReach(t, program, "item"); !reach.Unknown {
		t.Fatalf("reach is %+v, want Unknown", reach)
	}
	forOfWritesDispatched(t, program)
}

// A catch binding assigned in its block reaches what was thrown too, which the graph can't name. No
// source writes one today (assignments.go stops "assigning to what a catch caught"), so the pin is
// built as IR: without the catch's write, the graph would prove it the literal assigned. Mutant 2.
func TestCatchBindingAssignedIsUnknown(t *testing.T) {
	t.Parallel()
	program := &ir.Program{
		Locals: []ir.Local{{Name: "caught", Type: ir.Object, Function: -1}},
		Main: []ir.Statement{ir.Try{
			HasCatch:   true,
			CatchLocal: 0,
			Catch:      []ir.Statement{ir.Assign{Local: 0, Value: ir.ObjectLiteral{GraphTypes: []int{-1}}}},
		}},
	}
	reach := newAllocationFlowGraph(program).ReachingAllocations(ir.Read{Local: 0, Of: ir.Object})
	if !reach.Unknown || !slices.Contains(reach.Reasons, "a catch binding takes what was thrown") {
		t.Fatalf("reach is %+v, want Unknown for a catch binding", reach)
	}
}

// collect's table of local-writing fields, localWrites, has exactly the rows local_uses.go's
// localFields classes as writes, and TestLocalFieldsClassifyEveryIntField holds that to internal/ir,
// so an IR node that comes to write a local reds here until collect says how it takes it. Only
// Declare and Assign are sources, read by collect's own cases; every other write is unknown, with why.
func TestLocalWritesAreLocalFieldsWrites(t *testing.T) {
	t.Parallel()
	for field, classed := range localFields {
		if _, listed := localWrites[field]; classed.role == localWrite && !listed {
			t.Errorf("localFields says %s writes a local, and localWrites doesn't say how collect takes it", field)
		}
	}
	for field, why := range localWrites {
		if localFields[field].role != localWrite {
			t.Errorf("localWrites lists %s, which localFields doesn't class as a write", field)
		}
		source := field == "Declare.Local" || field == "Assign.Local"
		if source != (why == "") {
			t.Errorf("localWrites takes %s as %q: only Declare and Assign are sources, every other write needs its reason", field, why)
		}
	}
}
