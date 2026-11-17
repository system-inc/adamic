package lower

import (
	"errors"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// graphFlows' collect took a Declare's and an Assign's value as a local's sources, and a for-of's
// binding's elements over an array literal, and looked at no other writer: a catch binding takes what
// was thrown, and a for-of pattern over an array literal binds its elements' parts, which nothing
// seeds, so a graph value could reach a graph-typed local unplanned (#hny64km, Lower's design, Memory
// agreed on #zxxp2h4). collect now takes every local-writing field by graphLocalWrites, and a demand of
// a local some node writes with a value it doesn't follow is NotYet at the local.
// graph-flow-writers.mutants.md names the edits that turn each back.

// TestGraphLocalWritesAreLocalFieldsWrites holds graphLocalWrites to the rows local_uses.go's
// localFields classes as writes, both ways, as TestLocalWritesAreLocalFieldsWrites holds view_flow.go's
// table, and TestLocalFieldsClassifyEveryIntField holds localFields to internal/ir: an IR node that
// comes to write a local reds here until collect says how it takes it. Only Declare and Assign are
// sources; every other row says why. Mutant 2.
func TestGraphLocalWritesAreLocalFieldsWrites(t *testing.T) {
	t.Parallel()
	for _, field := range slices.Sorted(maps.Keys(localFields)) {
		if _, listed := graphLocalWrites[field]; localFields[field].role == localWrite && !listed {
			t.Errorf("localFields says %s writes a local, and graphLocalWrites doesn't say how graphFlows' collect takes it", field)
		}
	}
	for _, field := range slices.Sorted(maps.Keys(graphLocalWrites)) {
		if localFields[field].role != localWrite {
			t.Errorf("graphLocalWrites lists %s, which localFields doesn't class as a write", field)
		}
		source := field == "Declare.Local" || field == "Assign.Local"
		if why := graphLocalWrites[field]; source != (strings.TrimSpace(why) == "") {
			t.Errorf("graphLocalWrites takes %s as %q: only Declare and Assign are sources, every other write needs its reason", field, why)
		}
	}
}

// graphUnknownWriters, which collect calls at every node, records a catch's binding and each local of
// a for-of pattern over an array literal, and nothing else: a catch that binds nothing, a pattern over
// anything but a literal, and a plain for-of binding over a literal, whose elements collect seeds as
// sources. Built as IR, so it holds whatever source reaches graphFlows today. Mutant 1.
func TestGraphUnknownWriters(t *testing.T) {
	t.Parallel()
	catchWords := "a graph value a catch binding takes, which the plan doesn't follow"
	patternWords := "a graph value a for-of pattern takes from an array literal, which the plan doesn't follow"
	for _, each := range []struct {
		name string
		node any
		want map[int]string
	}{
		{"a catch's binding", ir.Try{HasCatch: true, CatchLocal: 2}, map[int]string{3: catchWords}},
		{"a catch that binds nothing", ir.Try{HasCatch: true, CatchLocal: -1}, map[int]string{}},
		{"a try with only a finally", ir.Try{HasFinally: true, CatchLocal: -1}, map[int]string{}},
		{"a pattern over an array literal", ir.ForOf{Iterable: ir.ArrayLiteral{Element: ir.Object}, Element: ir.Object, Pattern: []ir.Binding{{Local: 4, Field: "0"}, {Local: 5, Field: "1"}}}, map[int]string{5: patternWords, 6: patternWords}},
		{"a pattern over a local", ir.ForOf{Iterable: ir.Read{Local: 0, Of: ir.Array}, Element: ir.Object, Pattern: []ir.Binding{{Local: 4, Field: "0"}}}, map[int]string{}},
		{"a binding over an array literal, its elements its sources", ir.ForOf{Iterable: ir.ArrayLiteral{Element: ir.Object}, Element: ir.Object, Local: 4}, map[int]string{}},
	} {
		unknown := map[int]string{}
		graphUnknownWriters(each.node, unknown)
		if !maps.Equal(unknown, each.want) {
			t.Errorf("%s: graphUnknownWriters recorded %v, want %v", each.name, unknown, each.want)
		}
	}
}

// graphFlowWritersStop lowers source and fails unless it's NotYet with words.
func graphFlowWritersStop(t *testing.T, source, words string) {
	t.Helper()
	_, err := lowerSource(t, source)
	var notYet *NotYet
	if !errors.As(err, &notYet) || notYet.What != words {
		t.Fatalf("got %v, want NotYet: %s", err, words)
	}
}

// p3, the probe: a fresh literal destructured out of an inline array literal reaches a graph slot
// through the pattern's local, whose writer collect doesn't follow. It was planned nowhere, so it stops
// at the local now.
func TestGraphFlowPatternOverLiteralStops(t *testing.T) {
	t.Parallel()
	graphFlowWritersStop(t, `interface Cell { next: Cell | undefined }
const head: Cell = { next: undefined };
head.next = head;
for (const [n] of [[{ next: undefined }, 0] as [Cell, number]]) {
	head.next = n;
}
console.log(head.next === head ? 'head' : 'fresh');
`, "a graph value a for-of pattern takes from an array literal, which the plan doesn't follow")
}

// p2: a catch binding narrowed to a graph-typed class and stored into a graph slot. A thrown value is
// always an Error today, so this is defense: whatever stops it, it never lowers. Its words are logged,
// since a stop before graphFlows (an instanceof on what a catch took, or the narrowed read) holds as
// well as graphFlows' own; the job's log says which it is. TestGraphUnknownWriters holds the catch's
// record whatever lowers today. Mutant 1 lowers it, if it reaches graphFlows.
func TestGraphFlowCatchBindingStops(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `class Cell {
	next: Cell | undefined = undefined;
}
const head = new Cell();
head.next = head;
try {
	throw new Error('thrown');
} catch (error) {
	if (error instanceof Cell) {
		head.next = error;
	}
}
console.log(head.next === head ? 'head' : 'caught');
`)
	var notYet *NotYet
	var refused *Refused
	switch {
	case errors.As(err, &notYet):
		t.Logf("p2 is NotYet: %s", notYet.What)
	case errors.As(err, &refused):
		t.Logf("p2 is refused: %s", refused.What)
	default:
		t.Fatalf("got %v, want it stopped: a graph value a catch binding takes is planned nowhere", err)
	}
}
