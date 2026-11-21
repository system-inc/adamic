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
// it; a different machine, tool or round count is never compared at all, only named.
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
			direction := ""
			switch {
			case measure.after.Max < measure.before.Min:
				direction = "faster"
			case measure.after.Min > measure.before.Max:
				direction = "slower"
			default:
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
	if len(moved) == 0 {
		moved = []string{"level against " + shortSha(previous.MeasuredCommit)}
	}
	return strings.Join(append(moved, named...), "; ")
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
	} {
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
