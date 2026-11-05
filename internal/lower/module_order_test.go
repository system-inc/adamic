package lower

import (
	"context"
	"errors"
	"maps"
	"path/filepath"
	"slices"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// moduleOrderLowering is a fresh lowering of the program at path, from its entry's checker.
func moduleOrderLowering(t *testing.T, path string) *lowering {
	t.Helper()
	absolute, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{absolute})
	if err != nil {
		t.Fatal(err)
	}
	checked, release := program.Checker(context.Background(), program.Files()[0])
	t.Cleanup(release)
	return &lowering{program: program, checker: checked, result: &ir.Program{}, this: -1, functionIndex: -1}
}

// moduleOrder is computed once per lowering and entry, however many times it's asked, and a held answer is a fresh
// one: the same order, the same cyclic flag and the same proven module reads as a lowering that asks once, on a
// cyclic program (#20pb346). Mutant, red: the held answer dropped, so every ask computes again.
func TestModuleOrderIsComputedOncePerLowering(t *testing.T) {
	t.Parallel()
	const cyclic = "../oracle/testdata/import_cycles/order/main.a"
	held := moduleOrderLowering(t, cyclic)
	entry := held.program.Files()[0]
	first, err := held.moduleOrder(entry)
	if err != nil {
		t.Fatal(err)
	}
	if !held.cyclicModules {
		t.Fatal("the fixture's modules aren't cyclic, so it can't hold proveModuleReads to its answer")
	}
	for asked := 2; asked <= 5; asked++ {
		again, err := held.moduleOrder(entry)
		if err != nil || !slices.Equal(again, first) {
			t.Fatalf("ask %d: %v, %d modules against %d", asked, err, len(again), len(first))
		}
	}
	if held.moduleOrders != 1 {
		t.Fatalf("module order computed %d times for one lowering and entry", held.moduleOrders)
	}
	fresh := moduleOrderLowering(t, cyclic)
	order, err := fresh.moduleOrder(fresh.program.Files()[0])
	if err != nil {
		t.Fatal(err)
	}
	if len(order) != len(first) || fresh.cyclicModules != held.cyclicModules || len(fresh.provenModuleReads) != len(held.provenModuleReads) {
		t.Fatalf("a held answer isn't a fresh one: %d modules against %d, cyclic %t against %t, %d proven reads against %d", len(first), len(order), held.cyclicModules, fresh.cyclicModules, len(held.provenModuleReads), len(fresh.provenModuleReads))
	}
	for index := range order {
		if fresh.program.FileName(order[index]) != held.program.FileName(first[index]) {
			t.Fatalf("module %d: held %s, fresh %s", index, held.program.FileName(first[index]), fresh.program.FileName(order[index]))
		}
	}
	if !slices.Equal(provenReadPlaces(held), provenReadPlaces(fresh)) {
		t.Fatal("a held answer's proven module reads aren't a fresh answer's")
	}
	// A caller's change to the order it was handed reaches no other caller.
	first[0] = nil
	if again, _ := held.moduleOrder(entry); again[0] == nil {
		t.Fatal("a caller's edit reached the held order")
	}
}

// provenReadPlaces is every proven module read's place, sorted, so two lowerings' proofs compare by where they are.
func provenReadPlaces(l *lowering) []string {
	places := []string{}
	for node := range maps.Keys(l.provenModuleReads) {
		places = append(places, l.program.Where(node))
	}
	slices.Sort(places)
	return places
}

// A module order that refused is held as its refusal: a second ask answers the same NotYet, so a dynamic import can't
// lower past it on a later ask (#20pb346, Lower's read). Mutant, red: only a successful order held.
func TestAHeldModuleOrderKeepsItsRefusal(t *testing.T) {
	t.Parallel()
	l := moduleOrderLowering(t, "testdata/source_identity_imports/dynamic_import.a")
	entry := l.program.Files()[0]
	_, first := l.moduleOrder(entry)
	var notYet *NotYet
	if !errors.As(first, &notYet) || notYet.What != "a dynamic import()" {
		t.Fatalf("first ask: %v", first)
	}
	for asked := 2; asked <= 3; asked++ {
		if _, again := l.moduleOrder(entry); again != first {
			t.Fatalf("ask %d answered %v, not the held refusal %v", asked, again, first)
		}
	}
	if l.moduleOrders != 1 {
		t.Fatalf("a refused module order computed %d times", l.moduleOrders)
	}
}
