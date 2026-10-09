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
	shared := formatfilesPrepareShard(t)
	ctx := formatfilesDeadline(t, "shard")
	goAnswers := shared.answers
	portSource, program, binary := shared.port.source, shared.port.program, shared.port.binary

	part := shared.parts[selected]
	partCases := formatfilesCaseFile(t, part)
	for _, side := range []struct {
		name string
		run  func(*testing.T) run
	}{
		{"Node", func(t *testing.T) run {
			return formatfilesNode(t, ctx, filepath.Join(portSource, "main.ts"), partCases)
		}},
		{"native", func(t *testing.T) run {
			return formatfilesExecute(t, ctx, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary, partCases)
		}},
		{"JavaScript backend", func(t *testing.T) run { return formatfilesNode(t, ctx, shared.port.javascript, partCases) }},
	} {
		t.Run(side.name, func(t *testing.T) {
			t.Parallel()
			result := side.run(t)
			if result.exitCode != 0 || len(result.stderr) > 0 {
				t.Fatalf("%s: exit %d, stderr %q", side.name, result.exitCode, result.stderr)
			}
			if difference := firstDifference(string(result.stdout), part.answers); difference != "" {
				t.Fatalf("%s and Go cohere differ: %s", side.name, difference)
			}
			t.Logf("%d trees: every answer the same from %s and Go cohere", strings.Count(part.answers, "tree "), side.name)
		})
	}
	t.Run("leaks", func(t *testing.T) {
		t.Parallel()
		if leaked := formatfilesLeaks(t, ctx, program, binary, partCases); leaked != "" {
			t.Errorf("leaks:\n%s", leaked)
		}
	})

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

}

// Each mutant keeps the entire live corpus, in its own independently bounded unit.
func formatfilesCheckMutant(t *testing.T, selected int) {
	t.Helper()
	mutant := mutants[selected]
	if requested := formatfilesSelectedShard(t); requested >= 0 && requested != formatfilesShard("mutant/"+mutant.name) {
		t.Skip("another shard selected")
	}
	shared := formatfilesReady(t)
	prepared := formatfilesPreparedMutant(t, selected)
	ctx := formatfilesDeadline(t, "mutant")
	casesPath, goAnswers := shared.casesPath, shared.answers
	mutated, mutatedBinary := prepared.source, prepared.binary
	for _, side := range []struct {
		name string
		run  func(*testing.T) run
	}{
		{"natively", func(t *testing.T) run {
			return formatfilesExecute(t, ctx, []string{"ASAN_OPTIONS=detect_leaks=0"}, mutatedBinary, casesPath)
		}},
		{"on Node", func(t *testing.T) run {
			return formatfilesNode(t, ctx, filepath.Join(mutated, "main.ts"), casesPath)
		}},
	} {
		t.Run(mutant.name+"/"+side.name, func(t *testing.T) {
			t.Parallel()
			result := side.run(t)
			if result.exitCode != 0 {
				t.Fatalf("%s the mutant exits %d (stderr %q); it must be caught by its answers, not by failing", side.name, result.exitCode, result.stderr)
			}
			difference := firstDifference(string(result.stdout), goAnswers)
			if difference == "" {
				t.Fatal("the mutant agrees with Go cohere: the comparison cannot see it")
			}
			t.Logf("%s, caught: %s", side.name, difference)
		})
	}
}
