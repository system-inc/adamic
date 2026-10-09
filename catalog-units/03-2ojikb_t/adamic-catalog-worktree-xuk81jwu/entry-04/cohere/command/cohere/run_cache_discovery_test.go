package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// A file created after discovery lists the tree, and before the run cache starts its clock, is never replayed
// over. The build's tsconfig enumeration reads discovery's listings (#bjv0tg4), so a run like that checks a tree
// without the file. If its record watched the directory from after the file appeared, the directory's time
// would match on the next run and the verdict would replay, clean, over a file with a type error in it. The
// record's clock starts before discovery reads anything, so that run is never recorded. Found in format's
// review of the listed format walk (#g3046x5).
func TestAFileCreatedAfterDiscoveryIsNeverReplayedOver(t *testing.T) {
	t.Parallel()
	fixture := newRunCacheFixture(t, buildCohere(t))
	fixture.establishHit()

	late := filepath.Join(fixture.root, "source", "late.ts")
	command := exec.Command(fixture.binary, verboseArguments([]string{"--no-format"})...)
	command.Dir = fixture.root
	command.Env = append([]string{"HOME=" + fixture.home, "COHERE_TEST_CREATE_AFTER_DISCOVERY=" + late}, withoutHome(os.Environ())...)
	if output, err := command.CombinedOutput(); err != nil && !strings.Contains(string(output), "late.ts") {
		// A run that saw the file fails on it, which is fine; one that failed for any other reason is not.
		if _, isExit := err.(*exec.ExitError); !isExit {
			t.Fatalf("the run that created the file failed: %v\n%s", err, output)
		}
	}
	if _, err := os.Stat(late); err != nil {
		t.Fatalf("the instrument created no file, so nothing below is about it: %v", err)
	}

	output, _ := fixture.run(true)
	if isRunCacheReplay(output) {
		t.Fatalf("the run after a file appeared beside discovery's listing replayed a verdict that never saw it:\n%s", output)
	}
	if !strings.Contains(output, "late.ts") {
		t.Errorf("the file with a type error in it was not reported:\n%s", output)
	}
}

// The same moment, seen by the format walk: a bare run's walk ahead reads discovery's listings (#g3046x5), so a
// file created after discovery listed its directory is not in that run's format universe. The run is never
// recorded, since the clock started before the listing, so the next run walks again, finds the file and formats
// it, rather than replaying a verdict that never saw it.
func TestAFileCreatedAfterDiscoveryIsFormattedNextRun(t *testing.T) {
	t.Parallel()
	fixture := newRunCacheFixture(t, buildCohere(t))
	formatting := func(environment ...string) string {
		t.Helper()
		command := exec.Command(fixture.binary, verboseArguments(nil)...)
		command.Dir = fixture.root
		command.Env = append(append([]string{"HOME=" + fixture.home}, environment...), withoutHome(os.Environ())...)
		output, err := command.CombinedOutput()
		if _, isExit := err.(*exec.ExitError); err != nil && !isExit {
			t.Fatalf("running cohere: %v\n%s", err, output)
		}
		return string(output)
	}
	// The fixture's JSON is written unformatted, and a run that writes is never recorded: the first run formats
	// it, the second records, the third replays.
	formatting()
	formatting()
	if output := formatting(); !isRunCacheReplay(output) {
		t.Fatalf("an unchanged, formatted tree did not replay, so nothing below can show a change was noticed:\n%s", output)
	}

	late := filepath.Join(fixture.root, "docs", "late.md")
	unformatted := "#   Late\n\n\n\nA paragraph.\n"
	formatting("COHERE_TEST_CREATE_AFTER_DISCOVERY="+late, "COHERE_TEST_CREATE_AFTER_DISCOVERY_TEXT="+unformatted)
	if contents, err := os.ReadFile(late); err != nil || string(contents) == "" {
		t.Fatalf("the instrument created no file, so nothing below is about it: %v", err)
	}

	if output := formatting(); isRunCacheReplay(output) {
		t.Fatalf("the run after a file appeared beside discovery's listing replayed a verdict whose walk never saw it:\n%s", output)
	}
	if contents, _ := os.ReadFile(late); string(contents) == unformatted {
		t.Errorf("the file created after discovery's listing was never formatted:\n%s", contents)
	}
}

// withoutHome is the environment without its HOME, which a fixture sets for itself.
func withoutHome(environment []string) []string {
	kept := make([]string, 0, len(environment))
	for _, variable := range environment {
		if !strings.HasPrefix(variable, "HOME=") {
			kept = append(kept, variable)
		}
	}
	return kept
}
