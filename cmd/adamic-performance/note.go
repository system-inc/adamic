package main

import (
	"fmt"
	"strconv"
	"strings"
)

// shortSha is a commit as the note names it.
func shortSha(commit string) string {
	return commit[:min(7, len(commit))]
}

// compareNote is what a measured snapshot says about the one before it. A number moved only when its new range of
// rounds doesn't overlap the old one, since two medians of a noisy machine differ nearly always and mean nothing by
// it; a different machine, tool or round count is never compared at all, only named. The execute half's moves follow
// the compile half's, and its new and gone programs the compile half's (compareExecute).
func compareNote(previous *snapshot, current snapshot) string {
	if previous == nil {
		return "first snapshot"
	}
	if differences := incomparable(*previous, current); len(differences) > 0 {
		return "not compared: " + strings.Join(differences, "; ")
	}
	var moved, named []string
	for _, entry := range current.Compile.Programs {
		before, found := previous.Compile.Programs.lookup(entry.Name)
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
	for _, entry := range previous.Compile.Programs {
		if _, found := current.Compile.Programs.lookup(entry.Name); !found {
			named = append(named, entry.Name+" gone")
		}
	}
	executeMoved, executeNamed := compareExecute(previous.Execute, current.Execute)
	moved = append(moved, executeMoved...)
	named = append(named, executeNamed...)
	if len(moved) == 0 {
		moved = []string{"level against " + shortSha(previous.MeasuredCommit)}
	}
	return strings.Join(append(moved, named...), "; ")
}

// movedDirection is how a spread moved, faster when its new range is wholly below the old one and slower when it is
// wholly above, or empty when the two overlap, an end shared among them.
func movedDirection(before, after spread) string {
	switch {
	case after.Max < before.Min:
		return "faster"
	case after.Min > before.Max:
		return "slower"
	}
	return ""
}

// compareExecute is what moved in the execute half, by compareNote's rule, and which of its programs are new, gone or
// judged on another measure. A previous snapshot with no execute half, from before it was measured, is named once as
// execute first measured, never as every program new. Only native's judged spread is named as moved, since it is the one the commit makes:
// Node's numbers move with the machine and Node's own version, and the ratio with both sides, so they are context
// beside native's and never named. A program judged on another measure than before is named and not compared, since
// the two spreads are of different times.
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

// incomparable is every way two snapshots were measured apart, each as what differs, old to new: a number from another
// machine, another tool or another count of rounds says nothing about the commit. A machine is its processor too, since a
// box keeps its name across a hardware swap.
func incomparable(previous, current snapshot) []string {
	var differences []string
	for _, field := range []struct {
		name          string
		before, after string
	}{
		{name: "machine.name", before: strconv.Quote(previous.Machine.Name), after: strconv.Quote(current.Machine.Name)},
		{name: "machine.cpu", before: strconv.Quote(previous.Machine.CPU), after: strconv.Quote(current.Machine.CPU)},
		{name: "machine.cores", before: strconv.Itoa(previous.Machine.Cores), after: strconv.Itoa(current.Machine.Cores)},
		{name: "tools.adamic", before: strconv.Quote(previous.Tools.Adamic), after: strconv.Quote(current.Tools.Adamic)},
		{name: "tools.clang", before: strconv.Quote(previous.Tools.Clang), after: strconv.Quote(current.Tools.Clang)},
		{name: "tools.node", before: strconv.Quote(previous.Tools.Node), after: strconv.Quote(current.Tools.Node)},
		{name: "compile.rounds", before: strconv.Itoa(previous.Compile.Rounds), after: strconv.Itoa(current.Compile.Rounds)},
		{name: "execute.rounds", before: strconv.Itoa(previous.Execute.Rounds), after: strconv.Itoa(current.Execute.Rounds)},
	} {
		// A previous snapshot with no execute half has no round count to differ from: compareExecute names it.
		if field.name == "execute.rounds" && previous.Execute.Rounds == 0 {
			continue
		}
		if field.before != field.after {
			differences = append(differences, fmt.Sprintf("%s %s to %s", field.name, field.before, field.after))
		}
	}
	return differences
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
