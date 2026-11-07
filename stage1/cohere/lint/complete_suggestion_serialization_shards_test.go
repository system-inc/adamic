package lint

import (
	"bytes"
	"context"

	"fmt"
	"github.com/system-inc/adamic/stage1/cohere/lint/oracles"
	"os"
	"path/filepath"
	"runtime"

	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

const testCompleteSuggestionSerializationShards = 6

// Pinned checks, rather than a growing file corpus: each is owned by one top-level leaf.
var completeSuggestionCases = []string{
	"oracle-fields",
	"Node",
	"emitted JavaScript",
	"sanitized native",
	"mutant Node",
	"mutant emitted JavaScript",
}
var completeSuggestionLeaves = []func(*testing.T){
	TestCompleteSuggestionSerialization_000,
	TestCompleteSuggestionSerialization_001,
	TestCompleteSuggestionSerialization_002,
	TestCompleteSuggestionSerialization_003,
	TestCompleteSuggestionSerialization_004,
	TestCompleteSuggestionSerialization_005,
}

func completeSuggestionUnion(test *testing.T) {
	test.Helper()
	completeSuggestionReady(test)
	if len(completeSuggestionLeaves) != testCompleteSuggestionSerializationShards ||
		len(completeSuggestionCases) != testCompleteSuggestionSerializationShards {
		test.Fatal("shard enumeration changed")
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
			test.Fatalf("%s assigned %d times", key, seen[key])
		}
	}
	// A planted disagreement must be rejected by precisely its owning shard.
	caught := []int{}
	for shard := range completeSuggestionLeaves {
		for i := range completeSuggestionCases {
			if i%testCompleteSuggestionSerializationShards == shard &&
				!completeSuggestionEqual(i, 1, []byte("planted"), []byte("ordinary")) {
				caught = append(caught, shard)
			}
		}
	}
	if len(caught) != 1 || caught[0] != 1 {
		test.Fatalf("planted disagreement caught by %v", caught)
	}
	test.Logf("union: %d checks exactly once; planted disagreement caught by shard %03d", len(seen), caught[0])
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
func TestCompleteSuggestionSerialization_Setup(test *testing.T) {
	test.Parallel()
	completeSuggestionReady(test)
}

func completeSuggestionPrepare(test *testing.T) {
	test.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	started := time.Now()
	defer func() {
		elapsed := time.Since(started)
		test.Logf("complete suggestion preparation: %.3fs", elapsed.Seconds())
	}()
	p := &completeSuggestionProductsValue
	directory := completeSuggestionFixture(test, false)
	p.directory = directory
	source := filepath.Join(directory, "suggestions.ts")
	text, err := os.ReadFile("testdata/serialization/audit-columns.ts.txt")
	if err != nil {
		test.Fatal(err)
	}
	if err := os.WriteFile(source, text, 0644); err != nil {
		test.Fatal(err)
	}
	p.path = filepath.Join(directory, "manifest.txt")
	if err := os.WriteFile(p.path, []byte(source+"\tno-debugger\n"), 0644); err != nil {
		test.Fatal(err)
	}
	p.oracle = filepath.Join(completeSuggestionOracleProduct(ctx, test, directory), "oracle")
	p.want = completeSuggestionExecute(ctx, test, "", p.oracle, "--manifest", p.path).output
	p.mutant = completeSuggestionFixture(test, true)
	product := completeSuggestionLoweredProduct(ctx, test, directory, false)
	p.module = filepath.Join(product, "lint.mjs")
	p.binary = filepath.Join(completeSuggestionNativeProduct(ctx, test, directory), "scanner")
	p.mutantModule = filepath.Join(completeSuggestionLoweredProduct(ctx, test, p.mutant, true), "lint.mjs")
	completeSuggestionProductsReady = true
}

// Fixtures live until TestMain exits, independent of the calling test's lifetime.
func completeSuggestionFixture(test *testing.T, mutant bool) string {
	test.Helper()
	directory, err := os.MkdirTemp(sharedDirectory, "complete-suggestion-")
	if err != nil {
		test.Fatal(err)
	}
	copyPort(test, directory, "", "")
	for _, name := range []string{"rule.a", "oracle.go"} {
		data, err := os.ReadFile(filepath.Join("testdata/serialization", name))
		if err != nil {
			test.Fatal(err)
		}
		if mutant && name == "rule.a" {
			from := []byte("start + 1, start + 2, ''")
			if bytes.Count(data, from) != 1 {
				test.Fatal("suggestion mutant anchor changed")
			}
			data = bytes.Replace(data, from, []byte("start + 1, start + 3, ''"), 1)
		}
		if err = os.WriteFile(filepath.Join(directory, "rules/no-debugger", name), data, 0644); err != nil {
			test.Fatal(err)
		}
	}
	if err = os.Remove(filepath.Join(directory, "rules/no-debugger/rule.ts")); err != nil {
		test.Fatal(err)
	}
	completeSuggestionOnlyRule(test, directory)
	return directory
}

func completeSuggestionProductFiles(test *testing.T) []string {
	test.Helper()
	files := []string{"internal", "oracle", "bridge", "go.mod", "cohere", "stage1/cohere/lint/registry"}
	for _, file := range portFiles(test) {
		files = append(files, filepath.ToSlash(filepath.Join("stage1/cohere/lint", file)))
	}
	files = append(files, "stage1/cohere/lint/testdata/serialization", "stage1/cohere/lint/testdata/oracle.go")
	return files
}

// The build phase and shard preparation call these same recipes and keys.
func completeSuggestionOracleProduct(ctx context.Context, test *testing.T, directory string) string {
	test.Helper()
	return oracles.CompleteSuggestionOracle(ctx, test, directory)
}

func completeSuggestionLoweredProduct(ctx context.Context, test *testing.T, directory string, mutant bool) string {
	test.Helper()
	return buildcache.Product(test, buildcache.Inputs{
		Name: fmt.Sprintf("complete-suggestion-lowered-%t", mutant), Files: completeSuggestionProductFiles(test),
		Flags: []string{
			packageDirectory,
			"serialization-only-no-debugger-v1",
			"second-edit-end+3=" + fmt.Sprint(mutant),
		},
		Toolchain: []string{runtime.Version(), buildcache.Tool("clang", "--version")},
	}, func(out string) error {
		prepareRegistry(test, directory)
		program, err := load.Load([]string{filepath.Join(directory, "main.ts")})
		if err != nil {
			return err
		}
		lowered, err := lower.Lower(ctx, program)
		if err != nil {
			return err
		}
		if err = os.WriteFile(
			filepath.Join(out, "lint.mjs"),
			[]byte(javascript.JavaScript(lowered)),
			0644,
		); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(out, "lint.c"), []byte(native.C(lowered)), 0644)
	})
}

func completeSuggestionNativeProduct(ctx context.Context, test *testing.T, directory string) string {
	test.Helper()
	return buildcache.Product(test, buildcache.Inputs{
		Name: "complete-suggestion-native", Files: completeSuggestionProductFiles(test),
		Flags: []string{
			packageDirectory,
			"serialization-only-no-debugger-v1",
			"sanitize=true",
			"ADAMIC_NATIVE_SPLIT=" + os.Getenv("ADAMIC_NATIVE_SPLIT"),
		},
		Toolchain: []string{runtime.Version(), buildcache.Tool("clang", "--version")},
	}, func(out string) error {
		product := completeSuggestionLoweredProduct(ctx, test, directory, false)
		data, err := os.ReadFile(filepath.Join(product, "lint.c"))
		if err != nil {
			return err
		}
		return native.Build(string(data), filepath.Join(out, "scanner"), native.Options{Sanitize: true})
	})
}

func completeSuggestionShard(test *testing.T, shard int) {
	test.Helper()
	p := completeSuggestionReady(test)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	deadline := time.AfterFunc(
		90*time.Second,
		func() { panic(fmt.Sprintf("P0: serialization shard %03d exceeded 90s", shard)) },
	)
	defer deadline.Stop()
	started := time.Now()
	defer func() {
		elapsed := time.Since(started)
		test.Logf("shard %03d: %.3fs cooked=%t", shard, elapsed.Seconds(), elapsed >= 60*time.Second)
	}()
	if shard < 0 || shard >= len(completeSuggestionCases) {
		test.Fatal("unknown shard")
	}
	if shard == 0 {
		for _, field := range []string{
			"suggestion\tfirst",
			"suggestion\tsecond",
			"suggestion\tempty",
			"suggestion-edit\t8 9",
			"fixed\t/*",
		} {
			if !bytes.Contains(p.want, []byte(field)) {
				test.Fatalf("missing field %q: %s", field, p.want)
			}
		}
		return
	}
	var got []byte
	switch shard {
	case 1:
		got = completeSuggestionNode(ctx, test, filepath.Join(p.directory, "main.ts"), p.path).output
	case 2:
		got = completeSuggestionNode(ctx, test, p.module, p.path).output
	case 3:
		got = completeSuggestionExecute(ctx, test, "", p.binary, "--manifest", p.path).output
	case 4:
		got = completeSuggestionNode(ctx, test, filepath.Join(p.mutant, "main.ts"), p.path).output
	case 5:
		got = completeSuggestionNode(ctx, test, p.mutantModule, p.path).output
	}
	if shard >= 4 {
		if bytes.Equal(got, p.want) {
			test.Fatalf("second suggestion edit mutant survived on %s", completeSuggestionCases[shard])
		}
		test.Logf(
			"second suggestion edit mutant caught on %s: %s",
			completeSuggestionCases[shard],
			difference(got, p.want),
		)
		return
	}
	if shard == 1 && os.Getenv("ADAMIC_COMPLETE_SUGGESTION_PLANT") == "1" {
		got = append(got, []byte("planted suggestion disagreement")...)
	}
	if !completeSuggestionEqual(shard, shard, got, p.want) {
		test.Fatalf("%s: %s", completeSuggestionCases[shard], difference(got, p.want))
	}
}
func TestCompleteSuggestionSerialization_000(test *testing.T) {
	test.Parallel()
	completeSuggestionShard(test, 0)
}
func TestCompleteSuggestionSerialization_001(test *testing.T) {
	test.Parallel()
	completeSuggestionShard(test, 1)
}
func TestCompleteSuggestionSerialization_002(test *testing.T) {
	test.Parallel()
	completeSuggestionShard(test, 2)
}
func TestCompleteSuggestionSerialization_003(test *testing.T) {
	test.Parallel()
	completeSuggestionShard(test, 3)
}
func TestCompleteSuggestionSerialization_004(test *testing.T) {
	test.Parallel()
	completeSuggestionShard(test, 4)
}
func TestCompleteSuggestionSerialization_005(test *testing.T) {
	test.Parallel()
	completeSuggestionShard(test, 5)
}

// This pinned case explicitly selects no-debugger. Other rule implementations are
// unreachable for it; avoid lowering them while retaining the rule, driver, and
// Go oracle used by the original comparison.
func completeSuggestionOnlyRule(test *testing.T, directory string) {
	test.Helper()
	entries, err := os.ReadDir(filepath.Join(directory, "rules"))
	if err != nil {
		test.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != "no-debugger" {
			if err := os.RemoveAll(filepath.Join(directory, "rules", entry.Name())); err != nil {
				test.Fatal(err)
			}
		}
	}
	prepareRegistry(test, directory)
}

// Each process prepares its own products, with real builders on cache misses.
// Call this before starting a case deadline, including time spent waiting for Once.
func completeSuggestionReady(test *testing.T) *completeSuggestionProducts {
	test.Helper()
	completeSuggestionProductsOnce.Do(func() { completeSuggestionPrepare(test) })
	// Fatal inside preparation completes Once too; other callers must also fail.
	if !completeSuggestionProductsReady {
		test.Fatal("complete suggestion shared preparation failed")
	}
	return &completeSuggestionProductsValue
}
func completeSuggestionExecute(ctx context.Context, test *testing.T, directory, name string, args ...string) execution {
	test.Helper()
	command := lintCommand(ctx, directory, name, args...)
	output, err := os.CreateTemp(test.TempDir(), "complete-stdout-")
	if err != nil {
		test.Fatal(err)
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
		test.Fatalf("%s %v: %v\n%s", name, args, err, &stderr)
	}
	data, err := os.ReadFile(output.Name())
	if err != nil {
		test.Fatal(err)
	}
	return execution{data, time.Since(started)}
}
func completeSuggestionNode(ctx context.Context, test *testing.T, module, path string) execution {
	test.Helper()
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		test.Fatal(err)
	}
	return completeSuggestionExecute(
		ctx,
		test,
		"",
		"node",
		"--disable-warning=ExperimentalWarning",
		runner,
		module,
		"--manifest",
		path,
	)
}
