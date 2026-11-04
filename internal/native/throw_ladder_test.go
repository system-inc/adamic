package native

import (
	"reflect"
	"testing"
)

// A check joins a chain of a statement's ladder only where the chain's list starts with its own,
// name for name, and only with the same way out. A temporary moved out between checks (kept) makes
// the next check's list no prefix of the chain's, so it starts a chain of its own, and the ladder
// never lets go of the moved temporary on its way. Only the rungs some check enters are labeled, and
// a way out more than one chain falls into is labeled for the others' jumps.
func TestLadderChainsFollowEachCheckList(t *testing.T) {
	t.Parallel()
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
		t.Fatalf("ways out: %q", ways)
	}
	want := []ladderChain{
		{owned: []string{"a", "b", "c"}, way: 0, entries: map[int]bool{1: true, 2: true, 3: true}},
		{owned: []string{"a", "c"}, way: 0, entries: map[int]bool{2: true}},
		{owned: []string{"a", "b"}, way: 1, entries: map[int]bool{2: true}},
	}
	if len(chains) != len(want) {
		t.Fatalf("%d chains, want %d", len(chains), len(want))
	}
	for index := range want {
		if !reflect.DeepEqual(*chains[index], want[index]) {
			t.Errorf("chain %d: %+v, want %+v", index, *chains[index], want[index])
		}
	}
	entered := [][2]int{{0, 1}, {0, 2}, {0, 3}, {1, 2}, {0, 1}, {-1, 0}, {2, 2}}
	for index, site := range sites {
		if got := [2]int{site.chain, site.rung}; got != entered[index] {
			t.Errorf("check %d enters chain %d at %d, want %v", index, got[0], got[1], entered[index])
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
		t.Errorf("ladder:\n%s\nwant:\n%s", got, text)
	}
}
