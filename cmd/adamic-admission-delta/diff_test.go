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
	programs, err := changedPrograms(dir, base, head)
	if err != nil {
		t.Fatal(err)
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
	selected, omitted := sample(programs, "head", 1, time.Second)
	if !reflect.DeepEqual(selected, []int{2, 3, 0}) || omitted != 1 {
		t.Fatalf("diff or witness sampled away: %v, omitted %d", selected, omitted)
	}
}
