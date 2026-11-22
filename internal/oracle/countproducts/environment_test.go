package countproducts

import (
	"slices"
	"strings"
	"testing"
)

// Not parallel: it sets the process's environment (t.Setenv), which a parallel test may not.
//
// Nothing of the process's environment reaches a counted run but TMPDIR, which its shard's key names: a variable set
// here, and an ADAMIC_THREADS a box might carry, are absent from what a Plain, an Input and a ParallelMap run see, and a
// Plain run sees runEnvironment exactly, TMPDIR's value among it. A count is reused on every box and across landings
// (adamic-loom's must on #a4frknq). Mutant: execute's, executeInput's or executeParallel's environment left inherited,
// and the probe shows.
func TestACountedRunSeesOnlyRunEnvironment(t *testing.T) {
	t.Setenv("COUNTPRODUCTS_ENVIRONMENT_PROBE", "leaked")
	t.Setenv("ADAMIC_THREADS", "7")
	temporary := t.TempDir()
	t.Setenv("TMPDIR", temporary)
	const env = "/usr/bin/env"
	plain, err := execute(env)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := lines(plain.stdout), []string{"PATH=/usr/bin:/bin", "TMPDIR=" + temporary}; !slices.Equal(got, want) {
		t.Errorf("a Plain run's environment is %q, want exactly %q", got, want)
	}
	input, err := executeInput(inputRun{directory: t.TempDir()}, env)
	if err != nil {
		t.Fatal(err)
	}
	parallel, err := executeParallel("1", env)
	if err != nil {
		t.Fatal(err)
	}
	for name, seen := range map[string][]string{"Plain": lines(plain.stdout), "Input": lines(input.stdout), "ParallelMap": lines(parallel.stdout)} {
		for _, variable := range seen {
			if strings.HasPrefix(variable, "COUNTPRODUCTS_ENVIRONMENT_PROBE=") || variable == "ADAMIC_THREADS=7" {
				t.Errorf("a %s run sees the process's %s", name, variable)
			}
		}
	}
	if !slices.Contains(lines(parallel.stdout), "ADAMIC_THREADS=1") {
		t.Errorf("a ParallelMap run's environment %q doesn't hold its one worker", lines(parallel.stdout))
	}
}

// lines are output's lines, without the last newline.
func lines(output []byte) []string {
	text := strings.TrimSuffix(string(output), "\n")
	if text == "" {
		return nil
	}
	return strings.Split(text, "\n")
}
