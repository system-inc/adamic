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

// executed is a program's execution judged on judge, with native's and Node's judged spreads as given and every other
// number fixed, so a test moves only what it names.
func executed(name, judge string, native, node spread) namedExecution {
	numbers := func(judged spread) runtimeNumbers {
		other := spread{Median: 500, Min: 490, Max: 510}
		if judge == "wall" {
			return runtimeNumbers{Wall: judged, User: other, PeakKiB: 2048}
		}
		return runtimeNumbers{Wall: other, User: judged, PeakKiB: 2048}
	}
	return namedExecution{Name: name, Execution: execution{Judge: judge, Native: numbers(native), Node: numbers(node), Ratio: ratioOf(native.Median, node.Median)}}
}

// executeNoteSnapshot is noteSnapshot with hello level in the compile half and the given execute programs over 3 rounds.
func executeNoteSnapshot(commit string, entries ...namedExecution) snapshot {
	value := noteSnapshot(commit, timed("hello", spread{Median: 349, Min: 340, Max: 362}, spread{Median: 45, Min: 44, Max: 47}))
	value.Execute = execute{Rounds: 3, Programs: executions(entries)}
	return value
}

// In the execute half only native's judged spread can move, by the compile half's rule: trees' native user time clear
// of its old range reads slower, after a compile move. Node's numbers moving, native's measure that isn't judged
// moving, and the ratio with them, are context and read level, as does parallel_files' native wall within its old
// range however its user time moved. Mutants: Node's judged spread compared too, and trees' Node faster is named;
// native's user time compared for a wall-judged program, and parallel_files reads slower.
func TestPerformanceNoteNamesOnlyNativesJudgedMove(t *testing.T) {
	t.Parallel()
	previous := executeNoteSnapshot(earlierCommit,
		executed("trees", "user", spread{Median: 1301, Min: 1290, Max: 1320}, spread{Median: 2032, Min: 2000, Max: 2050}),
		executed("parallel_files", "wall", spread{Median: 311, Min: 300, Max: 320}, spread{Median: 701, Min: 690, Max: 710}),
	)
	current := executeNoteSnapshot(laterCommit,
		executed("trees", "user", spread{Median: 1402, Min: 1380, Max: 1420}, spread{Median: 2032, Min: 2000, Max: 2050}),
		executed("parallel_files", "wall", spread{Median: 318, Min: 309, Max: 330}, spread{Median: 701, Min: 690, Max: 710}),
	)
	current.Execute.Programs[1].Execution.Native.User = spread{Median: 3200, Min: 3100, Max: 3300}
	current.Compile.Programs[0].Program.WarmBuild = spread{Median: 500, Min: 480, Max: 520}
	if got, want := compareNote(&previous, current), "hello warmBuild 349 to 500 ms (slower); trees native user 1301 to 1402 ms (slower)"; got != want {
		t.Errorf("note %q, want %q", got, want)
	}

	level := executeNoteSnapshot(laterCommit,
		executed("trees", "user", spread{Median: 1310, Min: 1300, Max: 1330}, spread{Median: 1500, Min: 1450, Max: 1550}),
		executed("parallel_files", "wall", spread{Median: 311, Min: 300, Max: 320}, spread{Median: 401, Min: 390, Max: 410}),
	)
	level.Execute.Programs[0].Execution.Native.Wall = spread{Median: 9000, Min: 8900, Max: 9100}
	level.Execute.Programs[1].Execution.Native.User = spread{Median: 3200, Min: 3100, Max: 3300}
	if got, want := compareNote(&previous, level), "level against 1ed745b"; got != want {
		t.Errorf("note %q, want %q", got, want)
	}
}

// An execute program in only one of the two is named new or gone beside the compile half's, and one judged on another
// measure is named and not compared; another count of execute rounds is never compared at all. Mutants: a new program
// compared against zeros, and it reads slower; execute.rounds left out of incomparable, and the note reads level.
func TestPerformanceNoteNamesNewGoneAndRejudgedExecutePrograms(t *testing.T) {
	t.Parallel()
	trees := spread{Median: 1301, Min: 1290, Max: 1320}
	onNode := spread{Median: 2032, Min: 2000, Max: 2050}
	previous := executeNoteSnapshot(earlierCommit, executed("trees", "user", trees, onNode), executed("graph", "user", trees, onNode), executed("sort", "user", trees, onNode))
	current := executeNoteSnapshot(laterCommit, executed("trees", "user", trees, onNode), executed("primes_parallel", "wall", trees, onNode), executed("sort", "wall", spread{Median: 9000, Min: 8900, Max: 9100}, onNode))
	if got, want := compareNote(&previous, current), "level against 1ed745b; primes_parallel new; sort judge user to wall; graph gone"; got != want {
		t.Errorf("note %q, want %q", got, want)
	}
	current.Execute.Rounds = 5
	if got, want := compareNote(&previous, current), "not compared: execute.rounds 3 to 5"; got != want {
		t.Errorf("note %q, want %q", got, want)
	}
}

// Against a snapshot from before the execute half was measured, its {} with no rounds, the compile half compares as
// ever and the execute half is named once, as first measured: its round count isn't another one, and its programs
// aren't each new. Mutants: execute.rounds compared against the 0, and the note reads not compared; every program
// named new.
func TestPerformanceNoteNamesAnExecuteHalfFirstMeasured(t *testing.T) {
	t.Parallel()
	trees := spread{Median: 1301, Min: 1290, Max: 1320}
	onNode := spread{Median: 2032, Min: 2000, Max: 2050}
	previous := noteSnapshot(earlierCommit, timed("hello", spread{Median: 349, Min: 340, Max: 362}, spread{Median: 45, Min: 44, Max: 47}))
	current := executeNoteSnapshot(laterCommit, executed("trees", "user", trees, onNode), executed("sort", "user", trees, onNode))
	if got, want := compareNote(&previous, current), "level against 1ed745b; execute first measured"; got != want {
		t.Errorf("note %q, want %q", got, want)
	}
	current.Compile.Programs[0].Program.WarmBuild = spread{Median: 500, Min: 480, Max: 520}
	if got, want := compareNote(&previous, current), "hello warmBuild 349 to 500 ms (slower); execute first measured"; got != want {
		t.Errorf("note %q, want %q", got, want)
	}
}
