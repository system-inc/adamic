package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestEditCacheSeparation(t *testing.T) {
	t.Parallel()
	// Reads: internal/lower/*.go
	engine, prepareModeError := prepareMode("../..", "testdata/mini", t.TempDir(), nil, nil, true)
	if prepareModeError != nil {
		t.Fatal(prepareModeError)
	}
	engine.cache = &resultCache{directory: t.TempDir()}
	engine.compiler = &compilerWorker{}
	defer engine.compiler.close()
	engine.profile = newRunProfile()
	test := classified{Path: "edit.js", Program: program("assertSameValue(Math.abs(-4), 4);")}
	for _, edit := range []string{"prime", "runtime", "lowering"} {
		if edit == "runtime" {
			engine.runtimeKey += "edited"
		}
		if edit == "lowering" {
			engine.context += "edited"
			engine.compilerIdentity += "edited"
		}
		if got := engine.attempt(test); got.Kind != outcomePass {
			t.Fatalf("%s: %+v", edit, got)
		}
		phases := engine.profile.Tests[len(engine.profile.Tests)-1].Phases
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
	engine.cache.bypass = true
	engine.profile = newRunProfile()
	if got := engine.attempt(test); got.Kind != outcomePass {
		t.Fatal(got)
	}
	assertFreshCacheProfile(t, engine.profile)
}

func TestNodeHarnessIdentity(test *testing.T) {
	test.Parallel()
	source := fstest.MapFS{
		"run.go":       {Data: []byte("runner")},
		"compiler.go":  {Data: []byte("compiler")},
		"edit_test.go": {Data: []byte("test")},
	}
	before := nodeHarnessSourceIdentity(source)
	source["compiler.go"].Data = []byte("changed compiler")
	if before != nodeHarnessSourceIdentity(source) {
		test.Fatal("compiler source invalidates Node harness")
	}
	source["run.go"].Data = []byte("changed execution")
	if before == nodeHarnessSourceIdentity(source) {
		test.Fatal("changed runner execution reused Node harness")
	}
	if nodeHarnessIdentity() == "" {
		test.Fatal("built harness missing")
	}
}

func TestWorkerLazyFallback(test *testing.T) {
	test.Parallel()
	// Reads: internal/lower/*.go
	engine, prepareModeError := prepareMode("../..", "testdata/mini", test.TempDir(), productAdamic(test), nil, true)
	if prepareModeError != nil {
		test.Fatal(prepareModeError)
	}
	engine.log = io.Discard
	engine.jobs = 2
	engine.cache = &resultCache{directory: test.TempDir()}
	report, runFilterError := engine.runFilter("pass", 0, false)
	if runFilterError != nil || report.Pass != 1 {
		test.Fatalf("worker report %+v: %v", report, runFilterError)
	}
	if _, statError := os.Stat(engine.adamic); !os.IsNotExist(statError) {
		test.Fatal("unused backup compiler was built")
	}
	source := filepath.Join(engine.work, "fallback.a")
	if writeError := os.WriteFile(source, []byte(program("assertSameValue(1, 1);")), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	result := engine.fallbackCompile(source)
	if result.Exit != 0 || result.Stdout == "" {
		test.Fatalf("lazy fallback lost compiler behavior: %+v", result)
	}
	if _, statError := os.Stat(engine.adamic); statError != nil {
		test.Fatal("fallback compiler missing", statError)
	}
}

func TestLoweringSourceEdit(test *testing.T) {
	test.Parallel()
	root := test.TempDir()
	runtime, absolutePathError := filepath.Abs("../../internal/native/runtime")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	if createDirectoryError := os.MkdirAll(
		filepath.Join(root, "internal", "native"),
		0700,
	); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	if symlinkError := os.Symlink(runtime, filepath.Join(root, "internal", "native", "runtime")); symlinkError != nil {
		test.Fatal(symlinkError)
	}
	lower := filepath.Join(root, "internal", "lower")
	if createDirectoryError := os.MkdirAll(lower, 0700); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	cache := &resultCache{directory: test.TempDir()}
	var previousContext string
	for _, text := range []string{"// A\n", "// B\n"} {
		if writeError := os.WriteFile(filepath.Join(lower, "probe.go"), []byte(text), 0600); writeError != nil {
			test.Fatal(writeError)
		}
		profile := newRunProfile()
		engine, prepareModeError := prepareMode(root, "testdata/mini", test.TempDir(), nil, profile, true)
		if prepareModeError != nil {
			test.Fatal(prepareModeError)
		}
		engine.cache = cache
		engine.compiler = &compilerWorker{}
		result := engine.attempt(classified{Path: "source-edit.js", Program: program("assertSameValue(1,1);")})
		engine.compiler.close()
		if result.Kind != outcomePass {
			test.Fatal(result)
		}
		if engine.context == previousContext {
			test.Fatal("one-byte lowering edit did not change native identity")
		}
		for _, phase := range profile.Tests[0].Phases {
			if phase.Stage == "native-observation" && phase.Hit {
				test.Fatal("one-byte lowering edit reused native result")
			}
			if phase.Stage == "node" && phase.Hit != (previousContext != "") {
				test.Fatal("lowering edit invalidated Node")
			}
		}
		if previousContext != "" {
			engine.cache.bypass = true
			engine.profile = newRunProfile()
			engine.compiler = &compilerWorker{}
			if got := engine.attempt(
				classified{Path: "source-edit.js", Program: program("assertSameValue(1,1);")},
			); got.Kind != outcomePass {
				test.Fatal(got)
			}
			engine.compiler.close()
			assertFreshCacheProfile(test, engine.profile)
		}
		previousContext = engine.context
	}
}

func TestRunnerLocationHelper(test *testing.T) {
	test.Parallel()
	if os.Getenv("ADAMIC_TEST262_CONTEXT_HELPER") != "1" {
		return
	}
	// A caller standing in a copy of the tree names that copy's root; by default this one's.
	root := os.Getenv("ADAMIC_TEST262_CONTEXT_ROOT")
	if root == "" {
		root = "../.."
	}
	tools, resolveToolsError := resolveTools()
	if resolveToolsError != nil {
		test.Fatal(resolveToolsError)
	}
	cache, _, context, prepareCacheError := prepareCache(root, tools)
	if prepareCacheError != nil {
		test.Fatal(prepareCacheError)
	}
	fmt.Printf("%s %s\n", context, cache.nodeContext)
}

func TestRunnerLocationIdentity(test *testing.T) {
	test.Parallel()
	executable, executableError := os.Executable()
	if executableError != nil {
		test.Fatal(executableError)
	}
	contents, readError := os.ReadFile(executable)
	if readError != nil {
		test.Fatal(readError)
	}
	relocated := filepath.Join(test.TempDir(), "runner")
	if writeError := os.WriteFile(relocated, contents, 0700); writeError != nil {
		test.Fatal(writeError)
	}
	var before []byte
	for _, path := range []string{executable, relocated} {
		command := exec.Command(path, "-test.run=^TestRunnerLocationHelper$")
		command.Env = append(os.Environ(), "ADAMIC_TEST262_CONTEXT_HELPER=1")
		output, combinedOutputError := command.CombinedOutput()
		if combinedOutputError != nil {
			test.Fatalf("helper: %v: %s", combinedOutputError, output)
		}
		if before != nil && string(before) != string(output) {
			test.Fatalf("relocating identical runner bytes invalidates observations: %s vs %s", before, output)
		}
		before = output
	}
}

// A bypassed repeat of a primed program must execute every observation again.
func assertFreshCacheProfile(test testing.TB, profile *runProfile) {
	test.Helper()
	if len(profile.Tests) == 0 {
		test.Fatal("no cache observations recorded")
	}
	for _, testCase := range profile.Tests {
		for _, phase := range testCase.Phases {
			if phase.Hit {
				test.Fatalf("bypassed %s observation was reused", phase.Stage)
			}
		}
	}
}
