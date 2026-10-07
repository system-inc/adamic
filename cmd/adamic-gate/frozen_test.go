package main

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestFrozenPlanRefusalAndEquivalentEvents(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	t.Setenv("GOWORK", "off")
	t.Setenv("ADAMIC_TOOLS", "")
	t.Setenv("ADAMIC_MARKDOWNWIDTH_DEPS", "")
	t.Setenv("WASI_SYSROOT", "")
	previous := os.Getenv("TMPDIR")
	t.Cleanup(func() { os.Setenv("TMPDIR", previous) })
	for _, name := range []string{"cmd/adamic-gate", "internal", "probe", "another"} {
		if err := os.MkdirAll(name, 0755); err != nil {
			t.Fatal(err)
		}
	}
	source := "package probe\nimport \"testing\"\nfunc TestFirst(t *testing.T) {}\nfunc TestSecond(t *testing.T) {t.Run(\"child\",func(t *testing.T){})}\n"
	for name, data := range map[string]string{"go.mod": "module github.com/system-inc/adamic\n\ngo 1.27\n", "probe/probe_test.go": source, "another/another_test.go": "package another\nimport \"testing\"\nfunc TestThird(t *testing.T) {}\n", timingPath: `{"github.com/system-inc/adamic/probe::TestFirst":4,"github.com/system-inc/adamic/probe::TestSecond":8,"github.com/system-inc/adamic/another::TestThird":1}`} {
		if err := os.WriteFile(name, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	declareFixtureCensus(t, root)
	gitFixture(t, root, "init", "-q")
	gitFixture(t, root, "config", "user.name", "Gate fixture")
	gitFixture(t, root, "config", "user.email", "gate@example.invalid")
	gitFixture(t, root, "add", ".")
	gitFixture(t, root, "commit", "-qm", "Frozen plan fixture")
	external := filepath.Join(t.TempDir(), "input.a")
	if err := os.WriteFile(external, []byte("original"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ADAMIC_FROZEN_PROBE", external)
	p, err := makeFrozenPlan(2)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateFrozenPlan(p, 2, 1); err != nil {
		t.Fatal("matching frozen plan", err)
	}
	if err := os.WriteFile("probe/probe_test.go", []byte(source+"// source mutant\n"), 0600); err != nil {
		t.Fatal(err)
	}
	gitFixture(t, root, "update-index", "--assume-unchanged", "probe/probe_test.go")
	if err := validateFrozenPlan(p, 2, 1); err == nil || !strings.Contains(err.Error(), "probe/probe_test.go") {
		t.Fatal("source change not refused by name", err)
	}
	gitFixture(t, root, "update-index", "--no-assume-unchanged", "probe/probe_test.go")
	if err := os.WriteFile("probe/probe_test.go", []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(external, []byte("changed"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := validateFrozenPlan(p, 2, 1); err == nil || !strings.Contains(err.Error(), "ADAMIC_FROZEN_PROBE") {
		t.Fatal("changed input accepted", err)
	}
	if err := os.WriteFile(external, []byte("original"), 0600); err != nil {
		t.Fatal(err)
	}
	node := p.Frozen.Tools["node"]
	nodeVersion := p.NodeVersion
	p.Frozen.Tools["node"] = "v0.0.0"
	p.NodeVersion = "v0.0.0"
	p.Digest = planDigest(p)
	if err := validateFrozenPlan(p, 2, 1); err == nil || !strings.Contains(err.Error(), "node") {
		t.Fatal("different Node accepted", err)
	}
	p.Frozen.Tools["node"] = node
	p.NodeVersion = nodeVersion
	p.Digest = planDigest(p)
	flags := p.Frozen.GoEnvironment["GOFLAGS"]
	p.Frozen.GoEnvironment["GOFLAGS"] = "-tags changed"
	p.Digest = planDigest(p)
	if err := validateFrozenPlan(p, 2, 1); err == nil || !strings.Contains(err.Error(), "GOFLAGS") {
		t.Fatal("different Go flags accepted", err)
	}
	p.Frozen.GoEnvironment["GOFLAGS"] = flags
	p.Digest = planDigest(p)
	digest := p.Digest
	p.Digest = "changed"
	if err := validateFrozenPlan(p, 2, 1); err == nil || !strings.Contains(err.Error(), "digest") {
		t.Fatal("changed digest accepted", err)
	}
	p.Digest = digest
	evidence := t.TempDir()
	path := filepath.Join(evidence, "plan.json")
	if err := saveJSON(path, p); err != nil {
		t.Fatal(err)
	}
	cache, err := os.UserCacheDir()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(cache, 0755); err != nil {
		t.Fatal(err)
	}
	scratch, err := os.MkdirTemp(cache, "gate-frozen-scratch-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(scratch) })
	var frozenDirs []string
	for index := 0; index < 2; index++ {
		self := filepath.Join(evidence, "self-"+string(rune('0'+index)))
		frozen := filepath.Join(evidence, "frozen-"+string(rune('0'+index)))
		if err := shard(index, 2, self, scratch, false); err != nil {
			t.Fatal(err)
		}
		if err := shardWithPlan(index, 2, frozen, scratch, false, path); err != nil {
			t.Fatal(err)
		}
		projection := func(dir string) []string {
			rows, _, _, err := readLog(filepath.Join(dir, "test.jsonl"))
			if err != nil {
				t.Fatal(err)
			}
			var events []string
			for _, row := range rows {
				events = append(events, row.key()+"::"+row.Action)
			}
			sort.Strings(events)
			return events
		}
		if !reflect.DeepEqual(projection(self), projection(frozen)) {
			t.Fatal("frozen and self-planned test events differ", index)
		}
		frozenDirs = append(frozenDirs, frozen)
	}
	if err := mergeWithPlan(frozenDirs, filepath.Join(evidence, "merged"), path); err != nil {
		t.Fatal("matching frozen merge", err)
	}
	if err := os.WriteFile("probe/probe_test.go", []byte(source+"// merge mutant\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := mergeWithPlan(frozenDirs, filepath.Join(evidence, "refused"), path); err == nil || !strings.Contains(err.Error(), "probe/probe_test.go") {
		t.Fatal("merge source change not refused by name", err)
	}
}

func TestFrozenInputGitMetadataAndBytes(t *testing.T) {
	root := t.TempDir()
	gitFixture(t, root, "init", "-q")
	gitFixture(t, root, "config", "user.name", "Gate fixture")
	gitFixture(t, root, "config", "user.email", "gate@example.invalid")
	path := filepath.Join(root, "input.a")
	if err := os.WriteFile(path, []byte("original"), 0600); err != nil {
		t.Fatal(err)
	}
	gitFixture(t, root, "add", ".")
	gitFixture(t, root, "commit", "-qm", "Input fixture")
	a, err := inputDigest(root)
	if err != nil {
		t.Fatal(err)
	}
	gitFixture(t, root, "config", "measurement.local", "different")
	b, err := inputDigest(root)
	if err != nil {
		t.Fatal(err)
	}
	if a != b {
		t.Fatal("clone-only Git config changes answer identity")
	}
	if err := os.WriteFile(path, []byte("mutant"), 0600); err != nil {
		t.Fatal(err)
	}
	c, err := inputDigest(root)
	if err != nil {
		t.Fatal(err)
	}
	if a == c {
		t.Fatal("dirty input bytes not hashed")
	}
}
