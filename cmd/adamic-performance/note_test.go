package main

import (
	"testing"
)

// The commits the note tests compare, the earlier one's short sha being what a level note names.
const (
	earlierCommit = "1ed745b3dbc1867856adc15710fabe277b6da101"
	laterCommit   = "2fa0c4d7e1b9a8c6f5e4d3c2b1a0918273645546"
)

// noteSnapshot is a measured snapshot at commit on one machine, with the given programs, each from its warm build's
// and front end's spreads.
func noteSnapshot(commit string, entries ...namedProgram) snapshot {
	return snapshot{
		Schema:         schemaVersion,
		Commit:         commit,
		MeasuredCommit: commit,
		Machine:        machine{Name: "Home", CPU: "AMD EPYC 9V45", Cores: 16},
		Tools:          tools{Adamic: "go1.27.1", Clang: "Ubuntu clang version 20.1.8", Node: "v24.19.0"},
		Compile:        compile{Rounds: 5, Programs: programs(entries)},
	}
}

func timed(name string, warmBuild, frontEnd spread) namedProgram {
	return namedProgram{Name: name, Program: program{Path: "bench/" + name + ".ts", WarmBuild: warmBuild, FrontEnd: frontEnd}}
}

// Ranges that overlap are level, however far apart their medians, in both directions: hello's warm build moves its
// median up from 349 to 360 within ranges that share 355 to 362, and Parser's down from 436 to 425 within ranges that
// share 425 to 440. Each boundary is overlap too: trees' front end has the old max as its new min, and Parser's front
// end the old min as its new max. Mutant: moved written as the medians differing, in either the faster or the slower
// branch, which reads the move in that direction as moved.
func TestPerformanceNoteIsLevelWhenRangesOverlap(t *testing.T) {
	t.Parallel()
	previous := noteSnapshot(earlierCommit,
		timed("hello", spread{Median: 349, Min: 340, Max: 362}, spread{Median: 45, Min: 44, Max: 47}),
		timed("Parser", spread{Median: 436, Min: 425, Max: 450}, spread{Median: 105, Min: 100, Max: 110}),
		timed("trees", spread{Median: 328, Min: 320, Max: 335}, spread{Median: 32, Min: 30, Max: 34}),
	)
	current := noteSnapshot(laterCommit,
		timed("hello", spread{Median: 360, Min: 355, Max: 370}, spread{Median: 45, Min: 44, Max: 47}),
		timed("Parser", spread{Median: 425, Min: 410, Max: 440}, spread{Median: 96, Min: 92, Max: 100}),
		timed("trees", spread{Median: 328, Min: 320, Max: 335}, spread{Median: 36, Min: 34, Max: 39}),
	)
	if got, want := compareNote(&previous, current), "level against 1ed745b"; got != want {
		t.Fatalf("note %q, want %q", got, want)
	}
}

// A range clear of the old one moved, faster when it's wholly below and slower when it's wholly above, each read as
// old median to new, in the programs' order and warm build before front end. Mutant: faster and slower swapped, or the
// boundary taken as clear (old max equal to new min reads moved, so the level test reds instead).
func TestPerformanceNoteNamesWhatMovedFasterOrSlower(t *testing.T) {
	t.Parallel()
	previous := noteSnapshot(earlierCommit,
		timed("hello", spread{Median: 1668, Min: 1600, Max: 1700}, spread{Median: 45, Min: 44, Max: 47}),
		timed("trees", spread{Median: 328, Min: 320, Max: 335}, spread{Median: 32, Min: 30, Max: 34}),
	)
	current := noteSnapshot(laterCommit,
		timed("hello", spread{Median: 349, Min: 340, Max: 362}, spread{Median: 45, Min: 44, Max: 47}),
		timed("trees", spread{Median: 328, Min: 320, Max: 335}, spread{Median: 40, Min: 35, Max: 41}),
	)
	want := "hello warmBuild 1668 to 349 ms (faster); trees frontEnd 32 to 40 ms (slower)"
	if got := compareNote(&previous, current); got != want {
		t.Fatalf("note %q, want %q", got, want)
	}
}

// Numbers from another machine, another tool or another count of rounds are never compared, even where every range
// moved: the note names what differs, old to new, and nothing else. Mutant: the comparison done across machines (the
// incomparable check dropped), and each case reads a moved number instead.
func TestPerformanceNoteRefusesToCompareAcrossMachinesToolsOrRounds(t *testing.T) {
	t.Parallel()
	previous := noteSnapshot(earlierCommit, timed("hello", spread{Median: 1668, Min: 1600, Max: 1700}, spread{Median: 45, Min: 44, Max: 47}))
	cases := []struct {
		change func(current *snapshot)
		want   string
	}{
		{change: func(current *snapshot) { current.Machine.Name = "Lunatic" }, want: `not compared: machine.name "Home" to "Lunatic"`},
		{change: func(current *snapshot) { current.Machine.CPU = "AMD Ryzen Threadripper 3970X 32-Core Processor" }, want: `not compared: machine.cpu "AMD EPYC 9V45" to "AMD Ryzen Threadripper 3970X 32-Core Processor"`},
		{change: func(current *snapshot) { current.Machine.Cores = 64 }, want: "not compared: machine.cores 16 to 64"},
		{change: func(current *snapshot) { current.Tools.Adamic = "go1.28.0" }, want: `not compared: tools.adamic "go1.27.1" to "go1.28.0"`},
		{change: func(current *snapshot) { current.Tools.Clang = "Ubuntu clang version 21.1.0" }, want: `not compared: tools.clang "Ubuntu clang version 20.1.8" to "Ubuntu clang version 21.1.0"`},
		{change: func(current *snapshot) { current.Tools.Node = "" }, want: `not compared: tools.node "v24.19.0" to ""`},
		{change: func(current *snapshot) { current.Compile.Rounds = 3 }, want: "not compared: compile.rounds 5 to 3"},
		{change: func(current *snapshot) { current.Machine.Name, current.Compile.Rounds = "Lunatic", 3 }, want: `not compared: machine.name "Home" to "Lunatic"; compile.rounds 5 to 3`},
	}
	for _, each := range cases {
		current := noteSnapshot(laterCommit, timed("hello", spread{Median: 349, Min: 340, Max: 362}, spread{Median: 45, Min: 44, Max: 47}))
		each.change(&current)
		if got := compareNote(&previous, current); got != each.want {
			t.Errorf("note %q, want %q", got, each.want)
		}
	}
	// A machine's load, or whether it was held quiet, doesn't make it another machine.
	current := noteSnapshot(laterCommit, timed("hello", spread{Median: 1668, Min: 1600, Max: 1700}, spread{Median: 45, Min: 44, Max: 47}))
	current.Machine.LoadBefore, current.Machine.HeldQuiet = [3]float64{17.5, 9, 4}, true
	if got, want := compareNote(&previous, current), "level against 1ed745b"; got != want {
		t.Errorf("note %q, want %q", got, want)
	}
}

// With nothing before it, a snapshot is the first. Mutant: a nil previous compared as an empty one, which reads every
// program as new.
func TestPerformanceNoteOfTheFirstSnapshot(t *testing.T) {
	t.Parallel()
	current := noteSnapshot(laterCommit, timed("hello", spread{Median: 349, Min: 340, Max: 362}, spread{Median: 45, Min: 44, Max: 47}))
	if got, want := compareNote(nil, current), "first snapshot"; got != want {
		t.Fatalf("note %q, want %q", got, want)
	}
}

// A program in only one of the two is named new or gone and never compared, beside the comparison of the rest: level
// when the rest are, and after what moved when something did. Mutant: a program missing from the previous one
// compared against zeros, which reads it as moved slower.
func TestPerformanceNoteNamesNewAndGonePrograms(t *testing.T) {
	t.Parallel()
	hello := spread{Median: 349, Min: 340, Max: 362}
	frontEnd := spread{Median: 45, Min: 44, Max: 47}
	previous := noteSnapshot(earlierCommit, timed("hello", hello, frontEnd), timed("graph", hello, frontEnd))
	current := noteSnapshot(laterCommit, timed("hello", hello, frontEnd), timed("trees", hello, frontEnd))
	if got, want := compareNote(&previous, current), "level against 1ed745b; trees new; graph gone"; got != want {
		t.Errorf("note %q, want %q", got, want)
	}
	current.Compile.Programs[0].Program.WarmBuild = spread{Median: 500, Min: 480, Max: 520}
	if got, want := compareNote(&previous, current), "hello warmBuild 349 to 500 ms (slower); trees new; graph gone"; got != want {
		t.Errorf("note %q, want %q", got, want)
	}
}
