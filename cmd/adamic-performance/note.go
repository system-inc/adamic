package main

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// shortSha is a commit as the note names it.
func shortSha(commit string) string {
	return commit[:min(7, len(commit))]
}

// compareNote is what a measured snapshot says about the one before it. A number moved only when its new range of
// rounds doesn't overlap the old one, since two medians of a noisy machine differ nearly always and mean nothing by
// it; a different machine, tool or round count is never compared, only named (incomparable). What both halves
// measure with, the machine, adamic and clang, stops the whole snapshot; what only one half measures with stops that
// half alone, named after the other half's comparison (#fjs0yd9). The execute half's moves follow the compile half's,
// and its new and gone programs the compile half's (compareExecute). It reads level only when a half was compared.
func compareNote(previous *snapshot, current snapshot) string {
	if previous == nil {
		return "first snapshot"
	}
	apart := incomparable(*previous, current)
	if len(apart.shared) > 0 {
		return "not compared: " + strings.Join(slices.Concat(apart.shared, apart.compile, apart.execute), "; ")
	}
	var moved, named, notCompared []string
	if len(apart.compile) > 0 {
		notCompared = append(notCompared, "compile not compared: "+strings.Join(apart.compile, "; "))
	} else {
		compileMoved, compileNamed := compareCompile(previous.Compile, current.Compile)
		moved = append(moved, compileMoved...)
		named = append(named, compileNamed...)
	}
	if len(apart.execute) > 0 {
		notCompared = append(notCompared, "execute not compared: "+strings.Join(apart.execute, "; "))
	} else {
		executeMoved, executeNamed := compareExecute(previous.Execute, current.Execute)
		moved = append(moved, executeMoved...)
		named = append(named, executeNamed...)
	}
	// A half was compared when its own fields matched and it had numbers before: the compile half always does, and the
	// execute half once a previous snapshot measured it.
	compared := len(apart.compile) == 0 || (len(apart.execute) == 0 && previous.Execute.Rounds > 0)
	if len(moved) == 0 && compared {
		moved = []string{"level against " + shortSha(previous.MeasuredCommit)}
	}
	return strings.Join(slices.Concat(moved, named, notCompared), "; ")
}

// compareCompile is what moved in the compile half, by compareNote's rule, and which of its programs are new or gone.
func compareCompile(previous, current compile) (moved, named []string) {
	for _, entry := range current.Programs {
		before, found := previous.Programs.lookup(entry.Name)
		if !found {
			named = append(named, entry.Name+" new")
			continue
		}
		for _, measure := range []struct {
			name          string
			before, after spread
		}{
			{name: "warmBuild", before: before.WarmBuild, after: entry.Program.WarmBuild},
			{name: "frontEnd", before: before.FrontEnd, after: entry.Program.FrontEnd},
		} {
			direction := movedDirection(measure.before, measure.after)
			if direction == "" {
				continue
			}
			moved = append(moved, fmt.Sprintf("%s %s %d to %d ms (%s)", entry.Name, measure.name, measure.before.Median, measure.after.Median, direction))
		}
	}
	for _, entry := range previous.Programs {
		if _, found := current.Programs.lookup(entry.Name); !found {
			named = append(named, entry.Name+" gone")
		}
	}
	return moved, named
}

// movedDirection is how a spread moved: faster when its new range is wholly below the old one and slower when it is
// wholly above, and only when its median moved more than the floor (movedFloor) as well. Otherwise it is empty, and the
// number reads level, an end shared between the ranges included, and a median that moved exactly the floor too: on a
// loaded box sort's 163 to 168 ms, 5 ms at a 5 ms floor, read slower for a commit that changed nothing (J-n5vnrtq). Ranges alone fire on noise at few rounds: for a
// number that didn't change, the chance its new rounds land wholly on one side is 2 in C(2n, n), 2 in 20 at 3
// rounds, so nine execute numbers named a false move on about 61% of commits that changed nothing (#hed1mkb, Adamic's
// ruling, from J-45rkvr9's sort 239 to 236 ms).
func movedDirection(before, after spread) string {
	if delta := after.Median - before.Median; max(delta, -delta) <= movedFloor(before.Median) {
		return ""
	}
	switch {
	case after.Max < before.Min:
		return "faster"
	case after.Min > before.Max:
		return "slower"
	}
	return ""
}

// movedFloor is what a median must move more than, in milliseconds, for its number to read moved: 2% of the old
// median, rounded up, and never under 5 ms.
func movedFloor(median int64) int64 {
	return max((median*2+99)/100, 5)
}

// compareExecute is what moved in the execute half, by compareNote's rule, and which of its programs are new, gone or
// judged on another measure. A previous snapshot with no execute half, from before it was measured, is named once as
// execute first measured, never as every program new. Only native's judged spread is named as moved, since it is the
// one the commit makes: Node's numbers move with the machine and Node's own version, and the ratio with both sides, so
// they are context beside native's and never named. A program judged on another measure than before is named and not
// compared, since the two spreads are of different times.
func compareExecute(previous, current execute) (moved, named []string) {
	if previous.Rounds == 0 {
		if current.Rounds == 0 {
			return nil, nil
		}
		return nil, []string{"execute first measured"}
	}
	for _, entry := range current.Programs {
		before, found := previous.Programs.lookup(entry.Name)
		if !found {
			named = append(named, entry.Name+" new")
			continue
		}
		if before.Judge != entry.Execution.Judge {
			named = append(named, fmt.Sprintf("%s judge %s to %s", entry.Name, before.Judge, entry.Execution.Judge))
			continue
		}
		old, now := before.Native.judged(before.Judge), entry.Execution.Native.judged(entry.Execution.Judge)
		if direction := movedDirection(old, now); direction != "" {
			moved = append(moved, fmt.Sprintf("%s native %s %d to %d ms (%s)", entry.Name, entry.Execution.Judge, old.Median, now.Median, direction))
		}
	}
	for _, entry := range previous.Programs {
		if _, found := current.Programs.lookup(entry.Name); !found {
			named = append(named, entry.Name+" gone")
		}
	}
	return moved, named
}

// differences are the ways two snapshots were measured apart, each as what differs, old to new, by the halves it
// keeps from being compared: shared by both, or the compile half's or the execute half's alone.
type differences struct {
	shared, compile, execute []string
}

// incomparable is every way two snapshots were measured apart: a number from another machine, another tool or another
// count of rounds says nothing about the commit. A machine is its processor too, since a box keeps its name across a
// hardware swap. The machine, adamic and clang are both halves': each half's programs are built by adamic and clang on
// that machine. Node and the execute rounds are the execute half's alone, since a compile build never runs Node, and
// the compile rounds the compile half's.
func incomparable(previous, current snapshot) differences {
	var found differences
	for _, field := range []struct {
		half          *[]string
		name          string
		before, after string
	}{
		{half: &found.shared, name: "machine.name", before: strconv.Quote(previous.Machine.Name), after: strconv.Quote(current.Machine.Name)},
		{half: &found.shared, name: "machine.cpu", before: strconv.Quote(previous.Machine.CPU), after: strconv.Quote(current.Machine.CPU)},
		{half: &found.shared, name: "machine.cores", before: strconv.Itoa(previous.Machine.Cores), after: strconv.Itoa(current.Machine.Cores)},
		{half: &found.shared, name: "tools.adamic", before: strconv.Quote(previous.Tools.Adamic), after: strconv.Quote(current.Tools.Adamic)},
		{half: &found.shared, name: "tools.clang", before: strconv.Quote(previous.Tools.Clang), after: strconv.Quote(current.Tools.Clang)},
		{half: &found.execute, name: "tools.node", before: strconv.Quote(previous.Tools.Node), after: strconv.Quote(current.Tools.Node)},
		{half: &found.compile, name: "compile.rounds", before: strconv.Itoa(previous.Compile.Rounds), after: strconv.Itoa(current.Compile.Rounds)},
		{half: &found.execute, name: "execute.rounds", before: strconv.Itoa(previous.Execute.Rounds), after: strconv.Itoa(current.Execute.Rounds)},
	} {
		// A previous snapshot with no execute half has nothing of it to differ from: compareExecute names it first
		// measured.
		if field.half == &found.execute && previous.Execute.Rounds == 0 {
			continue
		}
		if field.before != field.after {
			*field.half = append(*field.half, fmt.Sprintf("%s %s to %s", field.name, field.before, field.after))
		}
	}
	return found
}

// unmeasuredSnapshot is the snapshot of a commit the line didn't measure: the previous one's numbers, machine, tools
// and measuredCommit kept as they were, under the new commit, with why and whose numbers they are.
func unmeasuredSnapshot(previous snapshot, commit string, reason string) snapshot {
	kept := previous
	kept.Commit = commit
	kept.Unmeasured = reason
	kept.Note = fmt.Sprintf("unmeasured at %s: %s; numbers are %s's", shortSha(commit), reason, shortSha(previous.MeasuredCommit))
	return kept
}
