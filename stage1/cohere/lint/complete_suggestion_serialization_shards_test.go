package lint

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
)

const testCompleteSuggestionSerializationShards = 6

// Pinned checks, rather than a growing file corpus: each is owned by one top-level leaf.
var completeSuggestionCases = []string{"oracle-fields", "Node", "emitted JavaScript", "sanitized native", "mutant Node", "mutant emitted JavaScript"}
var completeSuggestionLeaves = []func(*testing.T){TestCompleteSuggestionSerialization_000, TestCompleteSuggestionSerialization_001, TestCompleteSuggestionSerialization_002, TestCompleteSuggestionSerialization_003, TestCompleteSuggestionSerialization_004, TestCompleteSuggestionSerialization_005}

func completeSuggestionUnion(t *testing.T) {
	t.Helper()
	completeSuggestionReady(t)
	if len(completeSuggestionLeaves) != testCompleteSuggestionSerializationShards || len(completeSuggestionCases) != testCompleteSuggestionSerializationShards {
		t.Fatal("shard enumeration changed")
	}
	seen := map[string]int{}
	for shard := range completeSuggestionLeaves {
		for i, key := range completeSuggestionCases {
			if i%testCompleteSuggestionSerializationShards == shard {
				seen[key]++
			}
		}
	}
	for _, key := range completeSuggestionCases {
		if seen[key] != 1 {
			t.Fatalf("%s assigned %d times", key, seen[key])
		}
	}
	// A planted disagreement must be rejected by precisely its owning shard.
	caught := []int{}
	for shard := range completeSuggestionLeaves {
		for i := range completeSuggestionCases {
			if i%testCompleteSuggestionSerializationShards == shard && !completeSuggestionEqual(i, 1, []byte("planted"), []byte("ordinary")) {
				caught = append(caught, shard)
			}
		}
	}
	if len(caught) != 1 || caught[0] != 1 {
		t.Fatalf("planted disagreement caught by %v", caught)
	}
	t.Logf("union: %d checks exactly once; planted disagreement caught by shard %03d", len(seen), caught[0])
}
func completeSuggestionEqual(caseID, plantedID int, got, want []byte) bool {
	return caseID != plantedID || bytes.Equal(got, want)
}

type completeSuggestionProducts struct {
	directory, mutant, path, oracle, binary, module, mutantModule string
	want                                                          []byte
}

var completeSuggestionProductsOnce sync.Once
var completeSuggestionProductsReady bool
var completeSuggestionProductsValue completeSuggestionProducts

// Setup remains an independently selectable preparation check.
func TestCompleteSuggestionSerialization_Setup(t *testing.T) {
	t.Parallel()
	completeSuggestionReady(t)
}

func completeSuggestionPrepare(t *testing.T) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	started := time.Now()
	defer func() {
		elapsed := time.Since(started)
		t.Logf("complete suggestion preparation: %.3fs", elapsed.Seconds())
	}()
	p := &completeSuggestionProductsValue
	directory := completeSuggestionFixture(t, false)
	p.directory = directory
	source := filepath.Join(directory, "suggestions.ts")
	if err := os.WriteFile(source, []byte("/*😀*/debugger;\n"), 0644); err != nil {
		t.Fatal(err)
	}
	p.path = filepath.Join(directory, "manifest.txt")
	if err := os.WriteFile(p.path, []byte(source+"\tno-debugger\n"), 0644); err != nil {
		t.Fatal(err)
	}
	p.oracle = filepath.Join(completeSuggestionOracleProduct(ctx, t, directory), "oracle")
	p.want = completeSuggestionExecute(ctx, t, "", p.oracle, "--manifest", p.path).output
	p.mutant = completeSuggestionFixture(t, true)
	product := completeSuggestionLoweredProduct(ctx, t, directory, false)
	p.module = filepath.Join(product, "lint.mjs")
	p.binary = filepath.Join(completeSuggestionNativeProduct(ctx, t, directory), "scanner")
	p.mutantModule = filepath.Join(completeSuggestionLoweredProduct(ctx, t, p.mutant, true), "lint.mjs")
	completeSuggestionProductsReady = true
}

// Fixtures live until TestMain exits, independent of the calling test's lifetime.
func completeSuggestionFixture(t *testing.T, mutant bool) string {
	t.Helper()
	directory, err := os.MkdirTemp(sharedDirectory, "complete-suggestion-")
	if err != nil {
		t.Fatal(err)
	}
	copyPort(t, directory, "", "")
	for _, name := range []string{"rule.a", "oracle.go"} {
		data, err := os.ReadFile(filepath.Join("testdata/serialization", name))
		if err != nil {
			t.Fatal(err)
		}
		if mutant && name == "rule.a" {
			from := []byte("start + 1, start + 2, ''")
			if bytes.Count(data, from) != 1 {
				t.Fatal("suggestion mutant anchor changed")
			}
			data = bytes.Replace(data, from, []byte("start + 1, start + 3, ''"), 1)
		}
		if err = os.WriteFile(filepath.Join(directory, "rules/no-debugger", name), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err = os.Remove(filepath.Join(directory, "rules/no-debugger/rule.ts")); err != nil {
		t.Fatal(err)
	}
	completeSuggestionOnlyRule(t, directory)
	return directory
}

func completeSuggestionProductFiles(t *testing.T) []string {
	t.Helper()
	files := []string{"internal", "oracle", "bridge", "go.mod", "cohere", "stage1/cohere/lint/registry"}
	for _, file := range portFiles(t) {
		files = append(files, filepath.ToSlash(filepath.Join("stage1/cohere/lint", file)))
	}
	files = append(files, "stage1/cohere/lint/testdata/serialization", "stage1/cohere/lint/testdata/oracle.go")
	return files
}

// The build phase and shard preparation call these same recipes and keys.
func completeSuggestionOracleProduct(ctx context.Context, t *testing.T, directory string) string {
	t.Helper()
	return buildcache.Product(t, buildcache.Inputs{
		Name: "complete-suggestion-go-oracle", Files: completeSuggestionProductFiles(t),
		Flags:     []string{packageDirectory, "serialization-only-no-debugger-v1"},
		Toolchain: []string{runtime.Version(), buildcache.Tool("go", "version")},
	}, func(out string) error { return completeSuggestionBuildOracle(ctx, directory, out) })
}

func completeSuggestionLoweredProduct(ctx context.Context, t *testing.T, directory string, mutant bool) string {
	t.Helper()
	return buildcache.Product(t, buildcache.Inputs{
		Name: fmt.Sprintf("complete-suggestion-lowered-%t", mutant), Files: completeSuggestionProductFiles(t),
		Flags:     []string{packageDirectory, "serialization-only-no-debugger-v1", "second-edit-end+3=" + fmt.Sprint(mutant)},
		Toolchain: []string{runtime.Version(), buildcache.Tool("clang", "--version")},
	}, func(out string) error {
		prepareRegistry(t, directory)
		program, err := load.Load([]string{filepath.Join(directory, "main.ts")})
		if err != nil {
			return err
		}
		lowered, err := lower.Lower(ctx, program)
		if err != nil {
			return err
		}
		if err = os.WriteFile(filepath.Join(out, "lint.mjs"), []byte(javascript.JavaScript(lowered)), 0644); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(out, "lint.c"), []byte(native.C(lowered)), 0644)
	})
}

func completeSuggestionNativeProduct(ctx context.Context, t *testing.T, directory string) string {
	t.Helper()
	return buildcache.Product(t, buildcache.Inputs{
		Name: "complete-suggestion-native", Files: completeSuggestionProductFiles(t),
		Flags:     []string{packageDirectory, "serialization-only-no-debugger-v1", "sanitize=true", "ADAMIC_NATIVE_SPLIT=" + os.Getenv("ADAMIC_NATIVE_SPLIT")},
		Toolchain: []string{runtime.Version(), buildcache.Tool("clang", "--version")},
	}, func(out string) error {
		product := completeSuggestionLoweredProduct(ctx, t, directory, false)
		data, err := os.ReadFile(filepath.Join(product, "lint.c"))
		if err != nil {
			return err
		}
		return native.Build(string(data), filepath.Join(out, "scanner"), native.Options{Sanitize: true})
	})
}

func completeSuggestionShard(t *testing.T, shard int) {
	t.Helper()
	p := completeSuggestionReady(t)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	deadline := time.AfterFunc(90*time.Second, func() { panic(fmt.Sprintf("P0: serialization shard %03d exceeded 90s", shard)) })
	defer deadline.Stop()
	started := time.Now()
	defer func() {
		elapsed := time.Since(started)
		t.Logf("shard %03d: %.3fs cooked=%t", shard, elapsed.Seconds(), elapsed >= 60*time.Second)
	}()
	if shard < 0 || shard >= len(completeSuggestionCases) {
		t.Fatal("unknown shard")
	}
	if shard == 0 {
		for _, field := range []string{"suggestion\tfirst", "suggestion\tsecond", "suggestion\tempty", "suggestion-edit\t8 9", "fixed\t/*"} {
			if !bytes.Contains(p.want, []byte(field)) {
				t.Fatalf("missing field %q: %s", field, p.want)
			}
		}
		return
	}
	var got []byte
	switch shard {
	case 1:
		got = completeSuggestionNode(ctx, t, filepath.Join(p.directory, "main.ts"), p.path).output
	case 2:
		got = completeSuggestionNode(ctx, t, p.module, p.path).output
	case 3:
		got = completeSuggestionExecute(ctx, t, "", p.binary, "--manifest", p.path).output
	case 4:
		got = completeSuggestionNode(ctx, t, filepath.Join(p.mutant, "main.ts"), p.path).output
	case 5:
		got = completeSuggestionNode(ctx, t, p.mutantModule, p.path).output
	}
	if shard >= 4 {
		if bytes.Equal(got, p.want) {
			t.Fatalf("second suggestion edit mutant survived on %s", completeSuggestionCases[shard])
		}
		t.Logf("second suggestion edit mutant caught on %s: %s", completeSuggestionCases[shard], difference(got, p.want))
		return
	}
	if shard == 1 && os.Getenv("ADAMIC_COMPLETE_SUGGESTION_PLANT") == "1" {
		got = append(got, []byte("planted suggestion disagreement")...)
	}
	if !completeSuggestionEqual(shard, shard, got, p.want) {
		t.Fatalf("%s: %s", completeSuggestionCases[shard], difference(got, p.want))
	}
}
func TestCompleteSuggestionSerialization_000(t *testing.T) {
	t.Parallel()
	completeSuggestionShard(t, 0)
}
func TestCompleteSuggestionSerialization_001(t *testing.T) {
	t.Parallel()
	completeSuggestionShard(t, 1)
}
func TestCompleteSuggestionSerialization_002(t *testing.T) {
	t.Parallel()
	completeSuggestionShard(t, 2)
}
func TestCompleteSuggestionSerialization_003(t *testing.T) {
	t.Parallel()
	completeSuggestionShard(t, 3)
}
func TestCompleteSuggestionSerialization_004(t *testing.T) {
	t.Parallel()
	completeSuggestionShard(t, 4)
}
func TestCompleteSuggestionSerialization_005(t *testing.T) {
	t.Parallel()
	completeSuggestionShard(t, 5)
}

// Exercise the real leaves: one altered comparison fails precisely its owner.
func TestCompleteSuggestionSerialization_PlantedFailure(t *testing.T) {
	t.Parallel()
	completeSuggestionReady(t)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	// Loom runs one top-level test per process. Exercise that same selection,
	// preserving the proof that every non-owner (including both mutants) passes.
	for shard := range completeSuggestionLeaves {
		name := fmt.Sprintf("TestCompleteSuggestionSerialization_%03d", shard)
		command := completeSuggestionCommand(ctx, os.Args[0], "-test.run=^"+name+"$", "-test.timeout=90s", "-test.v")
		command.Env = append(os.Environ(), "ADAMIC_COMPLETE_SUGGESTION_PLANT=1")
		output, err := command.CombinedOutput()
		if command.Process != nil {
			_ = syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		}
		failures := []string{}
		for _, line := range strings.Split(string(output), "\n") {
			if strings.HasPrefix(line, "--- FAIL:") {
				failures = append(failures, line)
			}
		}
		if shard == 1 {
			if err == nil || len(failures) != 1 || !strings.HasPrefix(failures[0], "--- FAIL: "+name+" ") || !bytes.Contains(output, []byte("planted suggestion disagreement")) {
				t.Fatalf("wrong planted failure: %v; failures %v\n%s", err, failures, output)
			}
		} else if err != nil || len(failures) != 0 {
			t.Fatalf("non-owner %s failed: %v\n%s", name, err, output)
		}
	}
	t.Log("planted disagreement rejected by exactly shard 001")
}

// This pinned case explicitly selects no-debugger. Other rule implementations are
// unreachable for it; avoid lowering them while retaining the rule, driver, and
// Go oracle used by the original comparison.
func completeSuggestionOnlyRule(t *testing.T, directory string) {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(directory, "rules"))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != "no-debugger" {
			if err := os.RemoveAll(filepath.Join(directory, "rules", entry.Name())); err != nil {
				t.Fatal(err)
			}
		}
	}
	prepareRegistry(t, directory)
}

// Each process prepares its own products, with real builders on cache misses.
// Call this before starting a case deadline, including time spent waiting for Once.
func completeSuggestionReady(t *testing.T) *completeSuggestionProducts {
	t.Helper()
	completeSuggestionProductsOnce.Do(func() { completeSuggestionPrepare(t) })
	// Fatal inside preparation completes Once too; other callers must also fail.
	if !completeSuggestionProductsReady {
		t.Fatal("complete suggestion shared preparation failed")
	}
	return &completeSuggestionProductsValue
}
func completeSuggestionCommand(ctx context.Context, name string, args ...string) *exec.Cmd {
	command := exec.CommandContext(ctx, name, args...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if err == syscall.ESRCH {
			return os.ErrProcessDone
		}
		return err
	}
	command.WaitDelay = time.Second
	return command
}
func completeSuggestionExecute(ctx context.Context, t *testing.T, directory, name string, args ...string) execution {
	t.Helper()
	command := completeSuggestionCommand(ctx, name, args...)
	command.Dir = directory
	output, err := os.CreateTemp(t.TempDir(), "complete-stdout-")
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	command.Stdout = output
	var stderr bytes.Buffer
	command.Stderr = &stderr
	started := time.Now()
	err = command.Run()
	if command.Process != nil {
		_ = syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
	}
	if err != nil || len(commandDiagnostics(name, stderr.Bytes())) != 0 {
		t.Fatalf("%s %v: %v\n%s", name, args, err, &stderr)
	}
	data, err := os.ReadFile(output.Name())
	if err != nil {
		t.Fatal(err)
	}
	return execution{data, time.Since(started)}
}
func completeSuggestionNode(ctx context.Context, t *testing.T, module, path string) execution {
	t.Helper()
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	return completeSuggestionExecute(ctx, t, "", "node", "--disable-warning=ExperimentalWarning", runner, module, "--manifest", path)
}
func completeSuggestionBuildOracle(ctx context.Context, sourceRoot, directory string) error {
	root, err := filepath.Abs(filepath.Join(repository, "cohere"))
	if err != nil {
		return err
	}
	side, err := filepath.Abs(filepath.Join(packageDirectory, "testdata/oracle.go"))
	if err != nil {
		return err
	}
	descriptors, err := registry.Generate(sourceRoot)
	if err != nil {
		return err
	}
	replacements := map[string]string{}
	var virtualFiles []string
	var failure error
	add := func(name, source string) {
		virtual := filepath.Join(root, "adamic_lint_"+name+".go")
		absolute, err := filepath.Abs(source)
		if err != nil {
			failure = err
			return
		}
		replacements[virtual] = absolute
		virtualFiles = append(virtualFiles, virtual)
	}
	add("oracle", side)
	add("registry", filepath.Join(sourceRoot, ".generated/registry.go"))
	for _, d := range descriptors {
		add(strings.ReplaceAll(d.Slug, "-", "_"), filepath.Join(sourceRoot, "rules", d.Slug, "oracle.go"))
	}
	if failure != nil {
		return failure
	}
	overlay, err := json.Marshal(map[string]any{"Replace": replacements})
	if err != nil {
		return err
	}
	path := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(path, overlay, 0644); err != nil {
		return err
	}
	binary := filepath.Join(directory, "oracle")
	args := append([]string{"build", "-overlay=" + path, "-o", binary}, virtualFiles...)
	command := completeSuggestionCommand(ctx, "go", args...)
	command.Dir = root
	var output, stderr bytes.Buffer
	command.Stdout, command.Stderr = &output, &stderr
	err = command.Run()
	if command.Process != nil {
		_ = syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
	}
	if err != nil || len(commandDiagnostics("go", stderr.Bytes())) != 0 {
		return fmt.Errorf("Go oracle build: %v\n%s\n%s", err, &output, &stderr)
	}
	return nil
}
