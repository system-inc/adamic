package countproducts

import (
	"os"
	"slices"
	"strings"
	"testing"
)

// Not parallel: it sets the process's environment (t.Setenv), which a parallel test may not.
//
// Nothing of the process's environment reaches a counted run, TMPDIR included: a variable set here, an ADAMIC_THREADS
// a box might carry and the process's own TMPDIR are absent from what a Plain, an Input and a ParallelMap run see. A
// Plain run sees exactly runEnvironment: PATH, and a TMPDIR the harness made (makeTemporary) of a fixed spelling length
// and /tmp's mode, so a count is a function of its row and never of its host (adamic-loom's must and ruling on
// #a4frknq). Mutants: execute's, executeInput's or executeParallel's environment left inherited, and the probe shows;
// runEnvironment's TMPDIR taken from the process again, and the Plain run sees the process's.
func TestACountedRunSeesOnlyRunEnvironment(t *testing.T) {
	t.Setenv("COUNTPRODUCTS_ENVIRONMENT_PROBE", "leaked")
	t.Setenv("ADAMIC_THREADS", "7")
	processTemporary := t.TempDir()
	t.Setenv("TMPDIR", processTemporary)
	temporary, err := makeTemporary()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(temporary) })
	if len(temporary) != len(temporaryParent)+temporaryHexDigits || !strings.HasPrefix(temporary, temporaryParent) {
		t.Errorf("a counted run's temporary directory is %s, want %s and %d hex digits", temporary, temporaryParent, temporaryHexDigits)
	}
	info, err := os.Stat(temporary)
	if err != nil {
		t.Fatal(err)
	}
	if mode := info.Mode() & (os.ModePerm | os.ModeSticky); mode != temporaryMode {
		t.Errorf("a counted run's temporary directory has mode %v, want %v", mode, temporaryMode)
	}
	environment := runEnvironment(temporary)
	const env = "/usr/bin/env"
	plain, err := execute(environment, env)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := lines(plain.stdout), []string{"PATH=/usr/bin:/bin", "TMPDIR=" + temporary}; !slices.Equal(got, want) {
		t.Errorf("a Plain run's environment is %q, want exactly %q", got, want)
	}
	input, err := executeInput(inputRun{directory: t.TempDir()}, environment, env)
	if err != nil {
		t.Fatal(err)
	}
	parallel, err := executeParallel("1", environment, env)
	if err != nil {
		t.Fatal(err)
	}
	for name, seen := range map[string][]string{"Plain": lines(plain.stdout), "Input": lines(input.stdout), "ParallelMap": lines(parallel.stdout)} {
		for _, variable := range seen {
			if strings.HasPrefix(variable, "COUNTPRODUCTS_ENVIRONMENT_PROBE=") || variable == "ADAMIC_THREADS=7" || variable == "TMPDIR="+processTemporary {
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
