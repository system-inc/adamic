package formatfiles

import (
	"path/filepath"
	"strings"
	"testing"
)

// The port walks every tree cohere's tests build, and hundreds generated at the edges of the walk's
// classes, and finds what Go cohere finds, file for file, count for count and error for error, and
// answers every question about repositories in them as Go cohere does: natively, on Node and through
// the JavaScript backend, byte for byte; and the native port leaks nothing.
func formatfilesCheckShard(t *testing.T, selected int) {
	t.Helper()
	if requested := formatfilesSelectedShard(t); requested >= 0 && requested != selected {
		t.Skip("another shard selected")
	}
	shared := formatfilesPrepareShard(t, selected)
	ctx := formatfilesDeadline(t, "shard")
	casesPath, goAnswers := shared.casesPath, shared.answers
	shards := shared.parts
	portSource, program, binary := shared.port.source, shared.port.program, shared.port.binary

	checks := make([][]func(*testing.T), testThePortParsesAsGoCohereDoesShards)

	for index, part := range shards {
		checks[index] = append(checks[index], func(t *testing.T) {
			casesPath, goAnswers := formatfilesCaseFile(t, part), part.answers
			nodeRun := formatfilesNode(t, ctx, filepath.Join(portSource, "main.ts"), casesPath)
			nativeRun := formatfilesExecute(t, ctx, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary, casesPath)
			sanitized := binary
			backendRun := formatfilesNode(t, ctx, shared.port.javascript, casesPath)
			for _, side := range []struct {
				name string
				run  run
			}{{"Node", nodeRun}, {"native", nativeRun}, {"the JavaScript backend", backendRun}} {
				if side.run.exitCode != 0 || len(side.run.stderr) > 0 {
					t.Fatalf("%s: exit %d, stderr %q", side.name, side.run.exitCode, side.run.stderr)
				}
				if difference := firstDifference(string(side.run.stdout), goAnswers); difference != "" {
					t.Errorf("%s and Go cohere differ: %s", side.name, difference)
				}
			}
			agreed := !t.Failed()
			if leaked := formatfilesLeaks(t, ctx, program, sanitized, casesPath); leaked != "" {
				t.Errorf("leaks:\n%s", leaked)
			}

			if agreed {
				t.Logf("%d trees, %d files offered, %d directories entered, %d refused walks: every answer the same from Go cohere, the port natively, on Node and through the JavaScript backend",
					strings.Count(goAnswers, "tree "), strings.Count(goAnswers, "\nfile "), strings.Count(goAnswers, "\ndirectory "), strings.Count(goAnswers, "\nenumerate error "))
			}
		})
	}

	// Cases that answer the same whatever the walk does would agree without testing it: each kind of
	// answer has to be among them.
	var missing []string
	for _, each := range []string{"\nenumerate error ", "\nnested-below error ", "\nnested \"", "\nnested-below \"", "\nsymbolic-links 1",
		"layer \"format.ignore\" 1", "layer \".gitignore\" 1", "\nlayer \"ignorePatterns\" 0", "\ndeclined \".i\"", "\ndeclined \".\u03b1\u03c3\"",
		"\ndeclined \".\"", "\ndeclined \"(none)\"", "\" 1\nhas", "\nfile \"", "\nadamic \""} {
		if !strings.Contains(goAnswers, each) {
			missing = append(missing, each)
		}
	}
	if len(missing) > 0 {
		t.Errorf("the cases never reach %q", missing)
	}

	// The comparison has to be able to fail. Each mutant changes the port where only its answers can
	// show it, and the port must then disagree with Go cohere: natively, since that is the program this
	// test holds, and on Node, which shows the fault is the port's and not stage 0's.
	for _, mutant := range mutants {
		index := formatfilesShard("mutant/" + mutant.name)
		if selected >= 0 && index != selected {
			continue
		}
		prepared := shared.mutated[mutant.name]
		mutated, mutatedBinary := prepared.source, prepared.binary

		checks[index] = append(checks[index], func(t *testing.T) {
			t.Logf("mutant %s", mutant.name)
			for _, side := range []struct {
				name string
				run  run
			}{{"natively", formatfilesExecute(t, ctx, []string{"ASAN_OPTIONS=detect_leaks=0"}, mutatedBinary, casesPath)}, {"on Node", formatfilesNode(t, ctx, filepath.Join(mutated, "main.ts"), casesPath)}} {
				if side.run.exitCode != 0 {
					t.Errorf("%s the mutant exits %d (stderr %q); it must be caught by its answers, not by failing", side.name, side.run.exitCode, side.run.stderr)
					continue
				}
				difference := firstDifference(string(side.run.stdout), goAnswers)
				if difference == "" {
					t.Errorf("%s the mutant agrees with Go cohere: the comparison cannot see it", side.name)
					continue
				}
				t.Logf("%s, caught: %s", side.name, difference)
			}
		})
	}

	if len(checks) != testThePortParsesAsGoCohereDoesShards {
		t.Fatal("shard count differs")
	}
	for _, check := range checks[selected] {
		check(t)
	}

}
