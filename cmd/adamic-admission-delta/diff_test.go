package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestDiffCoverage(t *testing.T) {
	t.Parallel()
	dir, base, _, compiler := commandFixture(t)
	write := func(path, text string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, path)), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, path), []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write("deleted.a", "console.log(1);\n")
	if _, err := git(dir, "add", "deleted.a"); err != nil {
		t.Fatal(err)
	}
	if _, err := git(dir, "commit", "-qm", "baseline deletion"); err != nil {
		t.Fatal(err)
	}
	base, _ = git(dir, "rev-parse", "HEAD")
	if err := os.Remove(filepath.Join(dir, "deleted.a")); err != nil {
		t.Fatal(err)
	}
	write("corpus/input.a", "console.log('43');\n")
	write("outside/deep/added.a", "console.log('44');\n")
	write("outside/ignored.ts", "console.log('45');\n")
	if _, err := git(dir, "add", "corpus/input.a", "outside", "deleted.a"); err != nil {
		t.Fatal(err)
	}
	if _, err := git(dir, "commit", "-qm", "changed inputs"); err != nil {
		t.Fatal(err)
	}
	head, err := git(dir, "rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	write("outside/deep/added.a", "dirty working tree\n")
	programs, unmatched, err := changedPrograms(dir, base, head, map[string]bool{})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(unmatched, []string{"outside/ignored.ts"}) {
		t.Fatalf("a .ts no manifest lists: unmatched %q, want outside/ignored.ts", unmatched)
	}
	paths := []string{}
	for _, p := range programs {
		paths = append(paths, p.Path)
		blob, err := git(dir, "rev-parse", head+":"+p.Path)
		if err != nil || blob != p.Blob {
			t.Fatal("wrong head blob", p)
		}
	}
	if !reflect.DeepEqual(paths, []string{"corpus/input.a", "outside/deep/added.a"}) {
		t.Fatal(paths)
	}
	o := invokeCommand(t, dir, "--base", base, "--head", head, "--base-binary", compiler, "--head-binary", compiler, "--corpus", "corpus", "--json")
	var r report
	if err := json.Unmarshal([]byte(o.Stdout), &r); err != nil {
		t.Fatal(err, o.Stderr)
	}
	if o.Exit != 0 || r.Admitted != 0 || r.DiffCount != 2 || len(r.Diff) != 2 || len(r.Programs) != 2 || r.Corpora[0].Name != "diff" {
		t.Fatalf("missing or duplicated diff coverage: %+v %s", r, o.Stderr)
	}
	// An ad hoc --corpus run has no manifest to list a .ts, so the changed one is reported, never run.
	if !reflect.DeepEqual(r.DiffTSUnmatched, []string{"outside/ignored.ts"}) {
		t.Fatalf("diff_ts_unmatched %q, want outside/ignored.ts", r.DiffTSUnmatched)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(o.Stdout), &raw); err != nil {
		t.Fatal(err)
	}
	if string(raw["admitted"]) != "0" {
		t.Fatal("zero admission omitted")
	}
}

func TestDiffSamplingIsMandatory(t *testing.T) {
	t.Parallel()
	programs := []entry{
		{Corpus: "witnesses", Class: "newly-accepted"},
		{Corpus: "fixtures", Class: "newly-accepted"},
		{Corpus: "diff", Class: "newly-accepted"},
		{Corpus: "diff", Class: "newly-accepted"},
	}
	selected, omitted := sample(programs, headSeed("head"), 1, time.Second)
	if !reflect.DeepEqual(selected, []int{2, 3, 0}) || omitted != 1 {
		t.Fatalf("diff or witness sampled away: %v, omitted %d", selected, omitted)
	}
}

// A changed .ts is in the diff exactly when the manifest lists it: a head that adds only a .ts program the manifest
// holds runs it, and a head that adds only a .ts module no corpus lists reports it as unmatched. Mutant: changedPrograms
// takes .a only again (the program is missing from the diff, and from diff_ts_unmatched).
func TestDiffCoverageTakesListedTSPrograms(t *testing.T) {
	t.Parallel()
	dir, base, compiler, _ := commandFixture(t)
	for _, step := range []struct{ path, text string }{
		{"programs/added.ts", "console.log('46');\n"},
		{"library/module.ts", "export const value = 1;\n"},
		{"library/types.d.ts", "declare const typed: number;\n"},
	} {
		if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, step.path)), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, step.path), []byte(step.text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, args := range [][]string{{"add", "programs", "library"}, {"commit", "-qm", "a .ts program and a module"}} {
		if _, err := git(dir, args...); err != nil {
			t.Fatal(err)
		}
	}
	head, err := git(dir, "rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	blob, err := git(dir, "rev-parse", head+":programs/added.ts")
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(manifest{SHA: head, Corpora: []corpus{{Name: "fixtures", Programs: []program{{Path: "programs/added.ts", Blob: blob}}}}})
	if err != nil {
		t.Fatal(err)
	}
	listing := filepath.Join(t.TempDir(), "manifest.json")
	if err := os.WriteFile(listing, data, 0600); err != nil {
		t.Fatal(err)
	}
	o := invokeCommand(t, dir, "--base", base, "--head", head, "--base-binary", compiler, "--head-binary", compiler, "--manifest", listing, "--json")
	var r report
	if err := json.Unmarshal([]byte(o.Stdout), &r); err != nil {
		t.Fatal(err, o.Stderr)
	}
	diff := []string{}
	for _, p := range r.Diff {
		diff = append(diff, p.Path)
	}
	if !reflect.DeepEqual(diff, []string{"programs/added.ts"}) || !reflect.DeepEqual(r.DiffTSUnmatched, []string{"library/module.ts"}) {
		t.Fatalf("diff %q and diff_ts_unmatched %q, want the listed program in the diff and only the module unmatched: %s", diff, r.DiffTSUnmatched, o.Stderr)
	}
}
