package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestDiffCoverage(test *testing.T) {
	test.Parallel()
	dir, base, _, compiler := commandFixture(test)
	write := func(path, text string) {
		test.Helper()
		if createDirectoryError := os.MkdirAll(
			filepath.Dir(filepath.Join(dir, path)),
			0755,
		); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
		if writeError := os.WriteFile(filepath.Join(dir, path), []byte(text), 0600); writeError != nil {
			test.Fatal(writeError)
		}
	}
	write("deleted.a", "console.log(1);\n")
	if _, gitError := git(dir, "add", "deleted.a"); gitError != nil {
		test.Fatal(gitError)
	}
	if _, gitError := git(dir, "commit", "-qm", "baseline deletion"); gitError != nil {
		test.Fatal(gitError)
	}
	base, _ = git(dir, "rev-parse", "HEAD")
	if removeError := os.Remove(filepath.Join(dir, "deleted.a")); removeError != nil {
		test.Fatal(removeError)
	}
	write("corpus/input.a", "console.log('43');\n")
	write("outside/deep/added.a", "console.log('44');\n")
	write("outside/ignored.ts", "console.log('45');\n")
	if _, gitError := git(dir, "add", "corpus/input.a", "outside", "deleted.a"); gitError != nil {
		test.Fatal(gitError)
	}
	if _, gitError := git(dir, "commit", "-qm", "changed inputs"); gitError != nil {
		test.Fatal(gitError)
	}
	head, gitError := git(dir, "rev-parse", "HEAD")
	if gitError != nil {
		test.Fatal(gitError)
	}
	write("outside/deep/added.a", "dirty working tree\n")
	programs, unmatched, changedProgramsError := changedPrograms(dir, base, head, map[string]bool{})
	if changedProgramsError != nil {
		test.Fatal(changedProgramsError)
	}
	if !reflect.DeepEqual(unmatched, []string{"outside/ignored.ts"}) {
		test.Fatalf("a .ts no manifest lists: unmatched %q, want outside/ignored.ts", unmatched)
	}
	paths := []string{}
	for _, program := range programs {
		paths = append(paths, program.Path)
		blob, gitError := git(dir, "rev-parse", head+":"+program.Path)
		if gitError != nil || blob != program.Blob {
			test.Fatal("wrong head blob", program)
		}
	}
	if !reflect.DeepEqual(paths, []string{"corpus/input.a", "outside/deep/added.a"}) {
		test.Fatal(paths)
	}
	observation := invokeCommand(
		test,
		dir,
		"--base",
		base,
		"--head",
		head,
		"--base-binary",
		compiler,
		"--head-binary",
		compiler,
		"--corpus",
		"corpus",
		"--json",
	)
	var report report
	if decodeError := json.Unmarshal([]byte(observation.Stdout), &report); decodeError != nil {
		test.Fatal(decodeError, observation.Stderr)
	}
	if observation.Exit != 0 ||
		report.Admitted != 0 ||
		report.DiffCount != 2 ||
		len(report.Diff) != 2 ||
		len(report.Programs) != 2 ||
		report.Corpora[0].Name != "diff" {
		test.Fatalf("missing or duplicated diff coverage: %+v %s", report, observation.Stderr)
	}
	if !strings.Contains("\n"+observation.Stderr, "\nchanged programs: 2\n") {
		test.Fatalf("stderr does not count the changed programs: %s", observation.Stderr)
	}
	// An ad hoc --corpus run has no manifest to list a .ts, so the changed one is reported, never run.
	if !reflect.DeepEqual(report.DiffTSUnmatched, []string{"outside/ignored.ts"}) {
		test.Fatalf("diff_ts_unmatched %q, want outside/ignored.ts", report.DiffTSUnmatched)
	}
	var raw map[string]json.RawMessage
	if decodeError := json.Unmarshal([]byte(observation.Stdout), &raw); decodeError != nil {
		test.Fatal(decodeError)
	}
	if string(raw["admitted"]) != "0" {
		test.Fatal("zero admission omitted")
	}
}

func TestDiffSamplingIsMandatory(test *testing.T) {
	test.Parallel()
	programs := []entry{
		{Corpus: "witnesses", Class: "newly-accepted"},
		{Corpus: "fixtures", Class: "newly-accepted"},
		{Corpus: "diff", Class: "newly-accepted"},
		{Corpus: "diff", Class: "newly-accepted"},
	}
	selected, omitted := sample(programs, headSeed("head"), 1, time.Second)
	if !reflect.DeepEqual(selected, []int{2, 3, 0}) || omitted != 1 {
		test.Fatalf("diff or witness sampled away: %v, omitted %d", selected, omitted)
	}
}

// A changed .ts is in the diff exactly when the manifest lists it: a head that adds only a .ts program the manifest
// holds runs it, and a head that adds only a .ts module no corpus lists reports it as unmatched. Mutant:
// changedPrograms takes .a only again (the program is missing from the diff, and from diff_ts_unmatched).
func TestDiffCoverageTakesListedTSPrograms(test *testing.T) {
	test.Parallel()
	dir, base, compiler, _ := commandFixture(test)
	for _, step := range []struct{ path, text string }{
		{"programs/added.ts", "console.log('46');\n"},
		{"library/module.ts", "export const value = 1;\n"},
		{"library/types.d.ts", "declare const typed: number;\n"},
	} {
		if createDirectoryError := os.MkdirAll(
			filepath.Dir(filepath.Join(dir, step.path)),
			0755,
		); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
		if writeError := os.WriteFile(filepath.Join(dir, step.path), []byte(step.text), 0600); writeError != nil {
			test.Fatal(writeError)
		}
	}
	for _, args := range [][]string{{"add", "programs", "library"}, {"commit", "-qm", "a .ts program and a module"}} {
		if _, gitError := git(dir, args...); gitError != nil {
			test.Fatal(gitError)
		}
	}
	head, gitError := git(dir, "rev-parse", "HEAD")
	if gitError != nil {
		test.Fatal(gitError)
	}
	blob, gitError := git(dir, "rev-parse", head+":programs/added.ts")
	if gitError != nil {
		test.Fatal(gitError)
	}
	data, encodeError := json.Marshal(manifest{
		SHA:     head,
		Corpora: []corpus{{Name: "fixtures", Programs: []program{{Path: "programs/added.ts", Blob: blob}}}},
	})
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	listing := filepath.Join(test.TempDir(), "manifest.json")
	if writeError := os.WriteFile(listing, data, 0600); writeError != nil {
		test.Fatal(writeError)
	}
	observation := invokeCommand(
		test,
		dir,
		"--base",
		base,
		"--head",
		head,
		"--base-binary",
		compiler,
		"--head-binary",
		compiler,
		"--manifest",
		listing,
		"--json",
	)
	var report report
	if decodeError := json.Unmarshal([]byte(observation.Stdout), &report); decodeError != nil {
		test.Fatal(decodeError, observation.Stderr)
	}
	diff := []string{}
	for _, program := range report.Diff {
		diff = append(diff, program.Path)
	}
	if !reflect.DeepEqual(diff, []string{"programs/added.ts"}) ||
		!reflect.DeepEqual(report.DiffTSUnmatched, []string{"library/module.ts"}) {
		test.Fatalf(
			"diff %q and diff_ts_unmatched %q, want the listed program in the diff and only the module unmatched: %s",
			diff,
			report.DiffTSUnmatched,
			observation.Stderr,
		)
	}
}
