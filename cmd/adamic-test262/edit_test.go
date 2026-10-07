package main

import (
	"fmt"
	"io"
	"os"

	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/system-inc/adamic/internal/boundedrun"
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

func TestLoweringSourceEdit(t *testing.T) {
	// Not parallel: explicitly enable cached observations for successive source edits.
	t.Setenv("ADAMIC_GATE_UNCACHED", "0")
	root := t.TempDir()
	runtime, err := filepath.Abs("../../internal/native/runtime")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "internal", "native"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(runtime, filepath.Join(root, "internal", "native", "runtime")); err != nil {
		t.Fatal(err)
	}
	lower := filepath.Join(root, "internal", "lower")
	if err := os.MkdirAll(lower, 0700); err != nil {
		t.Fatal(err)
	}
	cache := &resultCache{directory: t.TempDir()}
	var previousContext string
	for _, text := range []string{"// A\n", "// B\n"} {
		if err := os.WriteFile(filepath.Join(lower, "probe.go"), []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
		profile := newRunProfile()
		e, err := prepareMode(root, "testdata/mini", t.TempDir(), profile, true)
		if err != nil {
			t.Fatal(err)
		}
		e.cache = cache
		e.compiler = &compilerWorker{}
		result := e.attempt(classified{Path: "source-edit.js", Program: program("assertSameValue(1,1);")})
		e.compiler.close()
		if result.Kind != outcomePass {
			t.Fatal(result)
		}
		if e.context == previousContext {
			t.Fatal("one-byte lowering edit did not change native identity")
		}
		for _, phase := range profile.Tests[0].Phases {
			if phase.Stage == "native-observation" && phase.Hit {
				t.Fatal("one-byte lowering edit reused native result")
			}
			if phase.Stage == "node" && phase.Hit != (previousContext != "") {
				t.Fatal("lowering edit invalidated Node")
			}
		}
		previousContext = e.context
	}
}

func TestRunnerLocationHelper(t *testing.T) {
	if os.Getenv("ADAMIC_TEST262_CONTEXT_HELPER") != "1" {
		return
	}
	cache, _, context, err := prepareCache()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%s %s\n", context, cache.nodeContext)
}

func TestRunnerLocationIdentity(t *testing.T) {
	t.Parallel()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(executable)
	if err != nil {
		t.Fatal(err)
	}
	relocated := filepath.Join(t.TempDir(), "runner")
	if err := os.WriteFile(relocated, contents, 0700); err != nil {
		t.Fatal(err)
	}
	var before []byte
	for _, path := range []string{executable, relocated} {
		command, release := boundedrun.Command(boundedrun.Build, path, "-test.run=^TestRunnerLocationHelper$")
		command.Env = append(os.Environ(), "ADAMIC_TEST262_CONTEXT_HELPER=1")
		output, err := command.CombinedOutput()
		release()
		if err != nil {
			t.Fatalf("helper: %v: %s", err, output)
		}
		if before != nil && string(before) != string(output) {
			t.Fatalf("relocating identical runner bytes invalidates observations: %s vs %s", before, output)
		}
		before = output
	}
}

func TestNodeHelperIdentityCache(t *testing.T) {
	t.Parallel()
	sources := fstest.MapFS{"run.go": {Data: []byte("unchanged runner")}}
	cache := &resultCache{directory: t.TempDir()}
	for _, helper := range []string{"old execution helper", "changed execution helper"} {
		context := nodeHarnessWithHelperIdentity(sources, helper)
		key := nodeResultKey("unchanged program", "unchanged Node", "same adaptation", "same command", context)
		observed := cache.reuse(key, func() (execution, bool) { return execution{Stdout: helper}, true })
		if observed.Stdout != helper {
			t.Fatal("changed child execution helper reused Node observation")
		}
	}
}
