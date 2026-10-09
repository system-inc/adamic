package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestCheckpointKeysIncludeAllExecutionInputs(t *testing.T) {
	t.Parallel()
	base := resumeState{PlanDigest: "commit-and-plan", Context: "environment-tools-external-inputs", Index: 0}
	pattern := []string{"^TestPass$"}
	key := checkpointKey(base, "p", pattern)
	cases := []struct {
		state    resumeState
		pkg      string
		patterns []string
	}{
		{resumeState{PlanDigest: "different-commit", Context: base.Context, Index: 0}, "p", pattern},
		{resumeState{PlanDigest: base.PlanDigest, Context: "different-flags-or-inputs", Index: 0}, "p", pattern},
		{resumeState{PlanDigest: base.PlanDigest, Context: base.Context, Index: 1}, "p", pattern},
		{base, "different-package", pattern}, {base, "p", []string{"^TestPass.*$"}},
	}
	for _, c := range cases {
		if checkpointKey(c.state, c.pkg, c.patterns) == key {
			t.Fatal("checkpoint key lost an execution input", c)
		}
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "input")
	os.WriteFile(path, []byte("first"), 0600)
	first, err := pathDigest(dir)
	if err != nil {
		t.Fatal(err)
	}
	os.WriteFile(path, []byte("second"), 0600)
	second, err := pathDigest(dir)
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatal("external input bytes absent from key")
	}
	os.Chmod(path, 0644)
	third, err := pathDigest(dir)
	if err != nil {
		t.Fatal(err)
	}
	if second == third {
		t.Fatal("external input permissions absent from key")
	}
	env := []string{"FLAG=one"}
	tools := map[string]string{"go": "binary-one"}
	external := map[string]string{"corpus": "bytes-one"}
	baseline := executionKey(env, tools, external)
	if executionKey([]string{"FLAG=two"}, tools, external) == baseline {
		t.Fatal("environment absent from execution key")
	}
	if executionKey(env, map[string]string{"go": "binary-two"}, external) == baseline {
		t.Fatal("tools absent from execution key")
	}
	if executionKey(env, tools, map[string]string{"corpus": "bytes-two"}) == baseline {
		t.Fatal("external files absent from execution key")
	}

	a := map[string]string{"GOGCCFLAGS": "-m64 -ffile-prefix-map=/scratch/go-build123=/tmp/go-build"}
	b := map[string]string{"GOGCCFLAGS": "-m64 -ffile-prefix-map=/scratch/go-build456=/tmp/go-build"}
	if stableGoEnvironment(a) != stableGoEnvironment(b) {
		t.Fatal("random go-env work directory changed key")
	}
	b["GOGCCFLAGS"] = "-m32 -ffile-prefix-map=/scratch/go-build456=/tmp/go-build"
	if stableGoEnvironment(a) == stableGoEnvironment(b) {
		t.Fatal("compiler flags absent from key")
	}

}

func TestPackageCheckpointRequiresIntactCompleteEvidence(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := plan{Count: 1, Units: []unit{{Package: "p", Test: "TestPass", Shard: 0}}}
	log := `{"Action":"run","Package":"p","Test":"TestPass"}
{"Action":"pass","Package":"p","Test":"TestPass"}
{"Action":"pass","Package":"p"}
`
	path := filepath.Join(dir, "test.jsonl")
	stderr := filepath.Join(dir, "test.stderr")
	e := packageEvidence{Key: "key", Package: "p", Invocations: []invocation{{Package: "p", Args: testArgs("p", patterns(p.Units)[0]), Uncached: true}}}
	save := func(text string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(stderr, nil, 0600); err != nil {
			t.Fatal(err)
		}
		e.LogDigest, _ = fileDigest(path)
		e.StderrDigest, _ = fileDigest(stderr)
		if err := atomicJSON(filepath.Join(dir, "complete.json"), e); err != nil {
			t.Fatal(err)
		}
	}
	save(log)
	if _, found, err := loadPackage(dir, "key", p, 0, "p"); err != nil || !found {
		t.Fatalf("complete control rejected: %v", err)
	}
	if _, _, err := loadPackage(dir, "different-plan", p, 0, "p"); err == nil {
		t.Fatal("stale key accepted")
	}
	os.WriteFile(path, []byte(log+"{\"Action\":\"output\",\"Package\":\"p\",\"Output\":\"changed bytes\"}\n"), 0600)
	if _, _, err := loadPackage(dir, "key", p, 0, "p"); err == nil || !strings.Contains(err.Error(), "checksum") {
		t.Fatal("changed log accepted", err)
	}
	save(log)
	os.WriteFile(stderr, []byte("changed stderr"), 0600)
	if _, _, err := loadPackage(dir, "key", p, 0, "p"); err == nil || !strings.Contains(err.Error(), "checksum") {
		t.Fatal("changed stderr accepted", err)
	}
	save(strings.Replace(log, "{\"Action\":\"pass\",\"Package\":\"p\"}\n", "", 1))
	if _, _, err := loadPackage(dir, "key", p, 0, "p"); err == nil || !strings.Contains(err.Error(), "terminal") {
		t.Fatal("package without terminal accepted", err)
	}
	save(strings.Replace(log, "{\"Action\":\"pass\",\"Package\":\"p\",\"Test\":\"TestPass\"}\n", "", 1))
	if _, _, err := loadPackage(dir, "key", p, 0, "p"); err == nil {
		t.Fatal("test without terminal accepted")
	}
	save(strings.Replace(log, "\"Action\":\"pass\",\"Package\":\"p\",\"Test\":\"TestPass\"", "\"Action\":\"fail\",\"Package\":\"p\",\"Test\":\"TestPass\"", 1))
	if _, found, err := loadPackage(dir, "key", p, 0, "p"); err != nil || !found {
		t.Fatalf("complete failed evidence should stay reusable and red: %v", err)
	}
	os.Remove(filepath.Join(dir, "complete.json"))
	if _, found, err := loadPackage(dir, "key", p, 0, "p"); err != nil || found {
		t.Fatal("partial package considered complete")
	}
}

func TestLiteralChildrenAndParentCost(t *testing.T) {
	t.Parallel()
	// A table of the test's own, not another package's: typeaware split TestVolumeAgreementAndMutants into
	// independent units (113707a8, Oct 9), and reading its file made this test red on main for every gate after.
	file := filepath.Join(t.TempDir(), "volume_test.go")
	source := `package p

func TestVolume(t *testing.T) {
	changes := []struct {
		name string
		edit func(string) string
	}{
		{"assignable-types", nil},
		{"base shapes", nil},
	}
	for _, change := range changes {
		_ = change
	}
}
`
	if err := os.WriteFile(file, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	names, err := literalChildren(file, "TestVolume", "changes")
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(names, []string{"assignable-types", "base_shapes"}) {
		t.Fatalf("wrong rows from a named table: %v", names)
	}
	if _, err := literalChildren(file, "TestAbsent", "changes"); err == nil {
		t.Fatal("an absent parent enumerated")
	}
	p := plan{Count: 2, Units: []unit{{Package: "p", Test: "TestParent/one", Shard: 0, Seconds: 3}, {Package: "p", Test: "TestParent/two", Shard: 1, Seconds: 5}}}
	w := map[string]float64{"p::TestParent": 10, "p::TestParent/one": 3, "p::TestParent/two": 5}
	prediction := predictions(p, w)
	if prediction[0].Seconds != 5 || prediction[1].Seconds != 7 || prediction[1].Largest.Test != "TestParent/two" {
		t.Fatal("lost repeated parent cost", prediction)
	}
	if planDigest(p) == planDigest(plan{Count: 3, Units: p.Units}) {
		t.Fatal("plan digest lost shard count")
	}
}

func TestResumeCleanupPreservesForeignPaths(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	checkpoint := filepath.Join(root, "checkpoint")
	os.Mkdir(checkpoint, 0700)
	foreign := filepath.Join(root, "foreign")
	os.Mkdir(foreign, 0700)
	if err := atomicJSON(filepath.Join(checkpoint, "scratch.json"), scratchRecord{Key: "key", Path: foreign}); err != nil {
		t.Fatal(err)
	}
	if err := cleanupPackageScratch(checkpoint, "key", root); err == nil {
		t.Fatal("foreign scratch path accepted")
	}
	if _, err := os.Stat(foreign); err != nil {
		t.Fatal("foreign path deleted", err)
	}
	owned, err := os.MkdirTemp(root, "package-")
	if err != nil {
		t.Fatal(err)
	}
	atomicJSON(filepath.Join(checkpoint, "scratch.json"), scratchRecord{Key: "key", Path: owned})
	if err := cleanupPackageScratch(checkpoint, "key", root); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(owned); !os.IsNotExist(err) {
		t.Fatal("orphan package scratch retained")
	}
}
