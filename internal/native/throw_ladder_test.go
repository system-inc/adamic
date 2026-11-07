package native

import (
	"reflect"
	"strconv"
	"testing"
)

// A check joins a chain of a statement's ladder only where the chain's list starts with its own,
// name for name, and only with the same way out. A temporary moved out between checks (kept) makes
// the next check's list no prefix of the chain's, so it starts a chain of its own, and the ladder
// never lets go of the moved temporary on its way. Only the rungs some check enters are labeled, and
// a way out more than one chain falls into is labeled for the others' jumps.
func TestLadderChainsFollowEachCheckList(test *testing.T) {
	test.Parallel()
	const way = "goto landing;\n"
	sites := []*throwSite{
		{owned: []string{"a"}, way: way},
		{owned: []string{"a", "b"}, way: way},
		{owned: []string{"a", "b", "c"}, way: way},
		{owned: []string{"a", "c"}, way: way},
		{owned: []string{"a"}, way: way},
		{way: way},
		{owned: []string{"a", "b"}, way: "adamic_release(held);\n" + way},
	}
	ways, chains := ladderPlan(sites)
	if !reflect.DeepEqual(ways, []string{way, "adamic_release(held);\n" + way}) {
		test.Fatalf("ways out: %q", ways)
	}
	want := []ladderChain{
		{owned: []string{"a", "b", "c"}, way: 0, entries: map[int]bool{1: true, 2: true, 3: true}},
		{owned: []string{"a", "c"}, way: 0, entries: map[int]bool{2: true}},
		{owned: []string{"a", "b"}, way: 1, entries: map[int]bool{2: true}},
	}
	if len(chains) != len(want) {
		test.Fatalf("%d chains, want %d", len(chains), len(want))
	}
	for index := range want {
		if !reflect.DeepEqual(*chains[index], want[index]) {
			test.Errorf("chain %d: %+v, want %+v", index, *chains[index], want[index])
		}
	}
	entered := [][2]int{{0, 1}, {0, 2}, {0, 3}, {1, 2}, {0, 1}, {-1, 0}, {2, 2}}
	for index, site := range sites {
		if got := [2]int{site.chain, site.rung}; got != entered[index] {
			test.Errorf("check %d enters chain %d at %d, want %v", index, got[0], got[1], entered[index])
		}
	}
	got := (&emitter{}).ladderText("L", ways, chains, sites)
	const text = "goto L_done;\n" +
		"L_rung_0_3:;\nadamic_release(c);\nL_rung_0_2:;\nadamic_release(b);\nL_rung_0_1:;\nadamic_release(a);\ngoto L_out_0;\n" +
		"L_rung_1_2:;\nadamic_release(c);\nadamic_release(a);\n" +
		"L_out_0:;\ngoto landing;\n" +
		"L_rung_2_2:;\nadamic_release(b);\nadamic_release(a);\nadamic_release(held);\ngoto landing;\n" +
		"L_done:;\n"
	if got != text {
		test.Errorf("ladder:\n%s\nwant:\n%s", got, text)
	}
}

// A marker names its ladder as well as its check, so a statement's end replaces only its own: an
// enclosing statement's marker that reaches a nested statement's text passes through untouched, to
// be replaced by the enclosing statement, even where the nested ladder has a check of the same index.
// Mutant: closeLadder reading every marker as its own, and the outer check's text becomes the inner
// one's.
func TestLadderMarkersReplaceOnlyTheirOwnStatements(test *testing.T) {
	test.Parallel()
	e := &emitter{}
	outer := e.openLadder()
	e.ladder.sites = []*throwSite{{text: "outer check\n"}}
	outerSerial := e.ladder.serial
	inner := e.openLadder()
	e.ladder.sites = []*throwSite{{text: "inner check\n"}}
	e.out.WriteString("\x00" + strconv.Itoa(e.ladder.serial) + ":0\x00")
	e.out.WriteString("\x00" + strconv.Itoa(outerSerial) + ":0\x00")
	e.closeLadder(inner)
	e.closeLadder(outer)
	if got := e.out.String(); got != "inner check\nouter check\n" {
		test.Fatalf("markers replaced as %q, want each by its own statement's check", got)
	}
}

// No throw marker may reach a program's or a unit's finished C: cProgram and splitC both hand
// their text through unmarked, which says so out loud. A marker left in the text, here put there
// directly, is the panic; clean text passes through as it is.
func TestThrowMarkerReachingTheCPanics(test *testing.T) {
	test.Parallel()
	const clean = "int adamic_x;\n"
	if got := unmarked(clean); got != clean {
		test.Fatalf("clean C came back as %q", got)
	}
	defer func() {
		if recovered := recover(); recovered != "native: compiler bug: a throw marker reached the C" {
			test.Fatalf("want the marker's panic, got %v", recovered)
		}
	}()
	unmarked(clean + "\x001:0\x00" + clean)
	test.Fatal("a marker reached the C without a panic")
}
