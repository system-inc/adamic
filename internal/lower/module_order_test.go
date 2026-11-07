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
func moduleOrderLowering(test *testing.T, path string) *lowering {
	test.Helper()
	absolute, absolutePathError := filepath.Abs(path)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loadError := load.Load([]string{absolute})
	if loadError != nil {
		test.Fatal(loadError)
	}
	checked, release := program.Checker(context.Background(), program.Files()[0])
	test.Cleanup(release)
	return &lowering{program: program, checker: checked, result: &ir.Program{}, this: -1, functionIndex: -1}
}

// moduleOrder is computed once per lowering and entry, however many times it's asked, and a held answer is a fresh
// one: the same order, the same cyclic flag and the same proven module reads as a lowering that asks once, on a
// cyclic program (#20pb346). Mutant, red: the held answer dropped, so every ask computes again.
func TestModuleOrderIsComputedOncePerLowering(test *testing.T) {
	test.Parallel()
	const cyclic = "../oracle/testdata/import_cycles/order/main.a"
	held := moduleOrderLowering(test, cyclic)
	entry := held.program.Files()[0]
	first, moduleOrderError := held.moduleOrder(entry)
	if moduleOrderError != nil {
		test.Fatal(moduleOrderError)
	}
	if !held.cyclicModules {
		test.Fatal("the fixture's modules aren't cyclic, so it can't hold proveModuleReads to its answer")
	}
	for asked := 2; asked <= 5; asked++ {
		again, moduleOrderError := held.moduleOrder(entry)
		if moduleOrderError != nil || !slices.Equal(again, first) {
			test.Fatalf("ask %d: %v, %d modules against %d", asked, moduleOrderError, len(again), len(first))
		}
	}
	if held.moduleOrders != 1 {
		test.Fatalf("module order computed %d times for one lowering and entry", held.moduleOrders)
	}
	fresh := moduleOrderLowering(test, cyclic)
	order, moduleOrderError := fresh.moduleOrder(fresh.program.Files()[0])
	if moduleOrderError != nil {
		test.Fatal(moduleOrderError)
	}
	if len(order) != len(first) ||
		fresh.cyclicModules != held.cyclicModules ||
		len(fresh.provenModuleReads) != len(held.provenModuleReads) {
		test.Fatalf(
			"a held answer isn't a fresh one: %d modules against %d, cyclic %t against %t, %d proven reads against %d",
			len(first),
			len(order),
			held.cyclicModules,
			fresh.cyclicModules,
			len(held.provenModuleReads),
			len(fresh.provenModuleReads),
		)
	}
	for index := range order {
		if fresh.program.FileName(order[index]) != held.program.FileName(first[index]) {
			test.Fatalf(
				"module %d: held %s, fresh %s",
				index,
				held.program.FileName(first[index]),
				fresh.program.FileName(order[index]),
			)
		}
	}
	if !slices.Equal(provenReadPlaces(held), provenReadPlaces(fresh)) {
		test.Fatal("a held answer's proven module reads aren't a fresh answer's")
	}
	// A caller's change to the order it was handed reaches no other caller.
	first[0] = nil
	if again, _ := held.moduleOrder(entry); again[0] == nil {
		test.Fatal("a caller's edit reached the held order")
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
func TestAHeldModuleOrderKeepsItsRefusal(test *testing.T) {
	test.Parallel()
	l := moduleOrderLowering(test, "testdata/source_identity_imports/dynamic_import.a")
	entry := l.program.Files()[0]
	_, first := l.moduleOrder(entry)
	var notYet *NotYet
	if !errors.As(first, &notYet) || notYet.What != "a dynamic import()" {
		test.Fatalf("first ask: %v", first)
	}
	for asked := 2; asked <= 3; asked++ {
		if _, again := l.moduleOrder(entry); again != first {
			test.Fatalf("ask %d answered %v, not the held refusal %v", asked, again, first)
		}
	}
	if l.moduleOrders != 1 {
		test.Fatalf("a refused module order computed %d times", l.moduleOrders)
	}
}
