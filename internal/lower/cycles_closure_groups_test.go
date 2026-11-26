package lower

import (
	"context"
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// closureLinksPerRecord is the closure arm as it was before closureGroups: every function value related
// to proven on its own, in closureRecords' order, and its captured cells added when it relates. The
// grouped arm must give exactly these links, in this order.
func closureLinksPerRecord(l *lowering, proven *checker.Type) []cycleNode {
	links := []cycleNode{}
	for _, closure := range l.closureRecords {
		if !l.checker.IsTypeAssignableTo(closure.proven, proven) {
			continue
		}
		for _, local := range l.closureLinks(closure) {
			links = append(links, cycleNode{cell: local + 1})
		}
	}
	return links
}

// The closure arm relates each distinct type of a function value that captured something once, and never
// relates a value that captured nothing (#c0eypws). Five values of one type (two capturing nothing), three of
// a second, and two of a third whose values capture nothing ask two queries per function type, where the
// per-value arm asked ten. The links are the per-value arm's, in its order, and the cycle through the second
// value of type one, whose first value captures nothing and whose next captures only a number, is found.
// Mutants: (a) values not grouped by type, and each query count is 5; (b) a value that captured nothing
// related too, and each count is 3; a group keeping its first value's links alone, and holder goes unreached.
func TestClosureArmRelatesEachCapturingTypeOnce(t *testing.T) {
	t.Parallel()
	f, types := cycleIndexTypes(t, `declare const one: (value: number) => number; declare const other: (text: string) => string; declare const none: () => void; declare const holder: { back: number }; declare const count: number;`)
	for _, name := range []string{"one", "other", "none", "holder", "count"} {
		if types[name] == nil {
			t.Fatalf("no type for %s", name)
		}
	}
	l := f.l
	l.result.Locals = []ir.Local{{Name: "held", Function: -1, Captured: true}, {Name: "tally", Function: -1, Captured: true}, {Name: "total", Function: -1, Captured: true}}
	l.localTypes = map[int]*checker.Type{0: types["holder"], 1: types["count"], 2: types["count"]}
	l.result.Functions = []ir.Function{{}, {}, {Environment: []int{1}}, {Environment: []int{0}}, {Environment: []int{2}}, {}}
	l.closureRecords = []closureRecord{
		{proven: types["one"], function: 0},
		{proven: types["none"], function: 1},
		{proven: types["one"], function: 2},
		{proven: types["other"], function: 4},
		{proven: types["one"], function: 3},
		{proven: types["none"], function: 5},
		{proven: types["one"], function: 2},
		{proven: types["other"], function: 4},
		{proven: types["one"], function: 0},
		{proven: types["other"], function: 4},
	}
	want := map[string][]cycleNode{
		"one":   {{cell: 2}, {cell: 1}, {cell: 2}},
		"other": {{cell: 3}, {cell: 3}, {cell: 3}},
		"none":  {},
	}
	for _, name := range []string{"one", "other", "none"} {
		asked := f.closureQueries
		links := f.closureLinksTo(types[name])
		if queries := f.closureQueries - asked; queries != 2 {
			t.Errorf("relating the values to %s asked %d queries, want 2, one for each type whose values captured something", name, queries)
		}
		if !reflect.DeepEqual(links, want[name]) {
			t.Errorf("links to %s: got %v, want %v", name, links, want[name])
		}
		if perRecord := closureLinksPerRecord(l, types[name]); !reflect.DeepEqual(links, perRecord) {
			t.Errorf("links to %s: grouped %v, per value %v", name, links, perRecord)
		}
	}
	if !f.reaches(types["one"], cycleNode{proven: types["holder"]}) {
		t.Error("a value of type one reaches no holder, though the second of its values that captured something captured one")
	}
	if f.reaches(types["none"], cycleNode{proven: types["holder"]}) {
		t.Error("a value of type none reaches a holder, though none of its values captured anything")
	}
}

// On a program the lowering made, the grouped closure arm gives each function type the per-value arm's links
// in the per-value arm's order, and asks one query for each distinct type of a value that captured something.
// The program lowers with no refusal both ways, and a cycle through a closure stays refused ("a closure in a
// cell it captures", TestWhatZeroOneRefusesIsRefusedWithAFix).
func TestClosureArmMatchesThePerValueArm(t *testing.T) {
	t.Parallel()
	source := "function make(n: number): (from: number) => number {\n" +
		"\treturn (from: number): number => from * n;\n" +
		"}\n" +
		"function run(n: number): number {\n" +
		"\tconst scale = make(n);\n" +
		"\tconst again = make(n + 1);\n" +
		"\tconst twice = (from: number): number => from * 2;\n" +
		"\tconst shift = (from: number): number => from + n;\n" +
		"\tconst other = (from: number): number => from - n;\n" +
		"\tconst label = (text: string): string => `${text}${n}`;\n" +
		"\treturn scale(2) + again(3) + twice(1) + shift(4) + other(5) + (label(\"a\") === \"a1\" ? 1 : 0);\n" +
		"}\n" +
		"console.log(`${run(1)}`);\n"
	loaded, err := load.Load([]string{writeAgreementSource(t, "main.a", source)})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	entry := loaded.Files()[0]
	check, release := loaded.Checker(context.Background(), entry)
	t.Cleanup(release)
	lowered, err := lowerChecked(loaded, check, entry, nil, false, false)
	if err != nil {
		t.Fatal(err)
	}
	capturing := map[*checker.Type]bool{}
	captureless := 0
	targets := []*checker.Type{}
	targeted := map[*checker.Type]bool{}
	for _, closure := range lowered.closureRecords {
		if len(lowered.closureLinks(closure)) == 0 {
			captureless++
		} else {
			capturing[closure.proven] = true
		}
		if !targeted[closure.proven] {
			targeted[closure.proven] = true
			targets = append(targets, closure.proven)
		}
	}
	if len(capturing) < 2 || captureless == 0 {
		t.Fatalf("the program's %d function values hold %d capturing types and %d values that capture nothing; the pin needs at least 2 and 1", len(lowered.closureRecords), len(capturing), captureless)
	}
	finder := &cycleFinder{l: lowered}
	for _, target := range targets {
		asked := finder.closureQueries
		links := finder.closureLinksTo(target)
		if queries := finder.closureQueries - asked; queries != len(capturing) {
			t.Errorf("relating the values to %s asked %d queries, want %d", check.TypeToString(target), queries, len(capturing))
		}
		if perRecord := closureLinksPerRecord(lowered, target); !reflect.DeepEqual(links, perRecord) {
			t.Errorf("links to %s: grouped %v, per value %v", check.TypeToString(target), links, perRecord)
		}
	}
	t.Logf("%d function values, %d capturing types, %d capturing nothing", len(lowered.closureRecords), len(capturing), captureless)
}
