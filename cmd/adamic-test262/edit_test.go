package main

import (
	"io"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestEditCacheSeparation(t *testing.T) {
	// Not parallel: enable observations even when the surrounding gate is uncached.
	t.Setenv("ADAMIC_GATE_UNCACHED", "0")
	e, err := prepareMode("../..", "testdata/mini", t.TempDir(), nil, true)
	if err != nil {
		t.Fatal(err)
	}
	e.cache = &resultCache{directory: t.TempDir()}
	e.compiler = &compilerWorker{}
	defer e.compiler.close()
	e.profile = newRunProfile()
	test := classified{Path: "edit.js", Program: program("assertSameValue(Math.abs(-4), 4);")}
	for _, edit := range []string{"prime", "runtime", "lowering"} {
		if edit == "runtime" {
			e.runtimeKey += "edited"
		}
		if edit == "lowering" {
			e.context += "edited"
			e.compilerIdentity += "edited"
		}
		if got := e.attempt(test); got.Kind != outcomePass {
			t.Fatalf("%s: %+v", edit, got)
		}
		phases := e.profile.Tests[len(e.profile.Tests)-1].Phases
		wantNodeHit := edit != "prime"
		for _, phase := range phases {
			if phase.Stage == "node" && phase.Hit != wantNodeHit {
				t.Fatalf("%s Node hit=%v want %v", edit, phase.Hit, wantNodeHit)
			}
			if phase.Stage == "native-observation" && phase.Hit {
				t.Fatalf("%s native result hit after changed identity", edit)
			}
		}
	}
}

func TestNodeHarnessIdentity(t *testing.T) {
	t.Parallel()
	source := fstest.MapFS{"run.go": {Data: []byte("runner")}, "compiler.go": {Data: []byte("compiler")}, "edit_test.go": {Data: []byte("test")}}
	before := nodeHarnessSourceIdentity(source)
	source["compiler.go"].Data = []byte("changed compiler")
	if before != nodeHarnessSourceIdentity(source) {
		t.Fatal("compiler source invalidates Node harness")
	}
	source["run.go"].Data = []byte("changed execution")
	if before == nodeHarnessSourceIdentity(source) {
		t.Fatal("changed runner execution reused Node harness")
	}
	if nodeHarnessIdentity() == "" {
		t.Fatal("built harness missing")
	}
}

func TestWorkerLazyFallback(t *testing.T) {
	t.Parallel()
	e, err := prepareMode("../..", "testdata/mini", t.TempDir(), nil, true)
	if err != nil {
		t.Fatal(err)
	}
	e.log = io.Discard
	e.jobs = 2
	e.cache = &resultCache{directory: t.TempDir()}
	report, err := e.runFilter("pass", 0, false)
	if err != nil || report.Pass != 1 {
		t.Fatalf("worker report %+v: %v", report, err)
	}
	if _, err := os.Stat(e.adamic); !os.IsNotExist(err) {
		t.Fatal("unused backup compiler was built")
	}
	source := filepath.Join(e.work, "fallback.a")
	if err := os.WriteFile(source, []byte(program("assertSameValue(1, 1);")), 0600); err != nil {
		t.Fatal(err)
	}
	result := e.fallbackCompile(source)
	if result.Exit != 0 || result.Stdout == "" {
		t.Fatalf("lazy fallback lost compiler behavior: %+v", result)
	}
	if _, err := os.Stat(e.adamic); err != nil {
		t.Fatal("fallback compiler missing", err)
	}
}
