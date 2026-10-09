package lint

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/testguard"
)

const testCompleteSuggestionSerializationShards = 6

// Pinned checks, rather than a growing file corpus: each is owned by one top-level leaf.
var completeSuggestionCases = []string{"oracle-fields", "Node", "emitted JavaScript", "sanitized native", "mutant Node", "mutant emitted JavaScript"}
var completeSuggestionLeaves = []func(*testing.T){TestCompleteSuggestionSerialization_000, TestCompleteSuggestionSerialization_001, TestCompleteSuggestionSerialization_002, TestCompleteSuggestionSerialization_003, TestCompleteSuggestionSerialization_004, TestCompleteSuggestionSerialization_005}

func completeSuggestionUnion(t *testing.T) {
	t.Helper()
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

var completeSuggestionOnce sync.Once
var completeSuggestionProductsValue completeSuggestionProducts

func completeSuggestionSetup(t *testing.T) *completeSuggestionProducts {
	t.Helper()
	completeSuggestionOnce.Do(func() {
		started := time.Now()
		defer func() { t.Logf("TestCompleteSuggestionSerialization (setup): %.3fs", time.Since(started).Seconds()) }()
		p := &completeSuggestionProductsValue
		directory, err := os.MkdirTemp(sharedDirectory, "complete-suggestion-")
		if err != nil {
			t.Fatal(err)
		}
		p.directory = copyPort(t, directory, "", "")
		for _, name := range []string{"rule.a", "oracle.go"} {
			data, err := os.ReadFile(filepath.Join("testdata/serialization", name))
			if err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(filepath.Join(directory, "rules/no-debugger", name), data, 0644); err != nil {
				t.Fatal(err)
			}
		}
		if err = os.Remove(filepath.Join(directory, "rules/no-debugger/rule.ts")); err != nil {
			t.Fatal(err)
		}
		completeSuggestionOnlyRule(t, directory)
		source := filepath.Join(directory, "suggestions.ts")
		if err = os.WriteFile(source, []byte("/*😀*/debugger;\n"), 0644); err != nil {
			t.Fatal(err)
		}
		p.path = filepath.Join(directory, "manifest.txt")
		if err = os.WriteFile(p.path, []byte(source+"\tno-debugger\n"), 0644); err != nil {
			t.Fatal(err)
		}
		oracleDirectory, err := os.MkdirTemp(sharedDirectory, "complete-oracle-")
		if err != nil {
			t.Fatal(err)
		}
		p.oracle, err = goOracleIn(directory, oracleDirectory)
		if err != nil {
			t.Fatal(err)
		}
		p.want = execute(t, "", p.oracle, "--manifest", p.path).output
		mutantDirectory, err := os.MkdirTemp(sharedDirectory, "complete-mutant-")
		if err != nil {
			t.Fatal(err)
		}
		p.mutant = copyPort(t, mutantDirectory, "", "")
		for _, name := range []string{"rule.a", "oracle.go"} {
			data, err := os.ReadFile(filepath.Join(directory, "rules/no-debugger", name))
			if err != nil {
				t.Fatal(err)
			}
			if name == "rule.a" {
				from := []byte("start + 1, start + 2, ''")
				if bytes.Count(data, from) != 1 {
					t.Fatal("suggestion mutant anchor changed")
				}
				data = bytes.Replace(data, from, []byte("start + 1, start + 3, ''"), 1)
			}
			if err = os.WriteFile(filepath.Join(mutantDirectory, "rules/no-debugger", name), data, 0644); err != nil {
				t.Fatal(err)
			}
		}
		if err = os.Remove(filepath.Join(mutantDirectory, "rules/no-debugger/rule.ts")); err != nil {
			t.Fatal(err)
		}
		completeSuggestionOnlyRule(t, mutantDirectory)
		files := []string{"internal", "oracle", "bridge", "go.mod", "cohere", "stage1/cohere/lint/registry"}
		for _, file := range portFiles(t) {
			files = append(files, filepath.ToSlash(filepath.Join("stage1/cohere/lint", file)))
		}
		files = append(files, "stage1/cohere/lint/testdata/serialization")
		toolchain := []string{runtime.Version(), buildcache.Tool("clang", "--version")}
		for _, side := range []struct {
			dir    string
			mutant bool
		}{{directory, false}, {mutantDirectory, true}} {
			product := buildcache.Product(t, buildcache.Inputs{Name: fmt.Sprintf("complete-suggestion-lowered-%t", side.mutant), Files: files, Flags: []string{packageDirectory, "serialization-only-no-debugger-v1", "second-edit-end+3=" + fmt.Sprint(side.mutant)}, Toolchain: toolchain}, func(out string) error {
				prepareRegistry(t, side.dir)
				program, err := load.Load([]string{filepath.Join(side.dir, "main.ts")})
				if err != nil {
					return err
				}
				lowered, err := lower.Lower(context.Background(), program)
				if err != nil {
					return err
				}
				if err = os.WriteFile(filepath.Join(out, "lint.mjs"), []byte(javascript.JavaScript(lowered)), 0644); err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(out, "lint.c"), []byte(native.C(lowered)), 0644)
			})
			if side.mutant {
				p.mutantModule = filepath.Join(product, "lint.mjs")
			} else {
				p.module = filepath.Join(product, "lint.mjs")
				nativeProduct := buildcache.Product(t, buildcache.Inputs{Name: "complete-suggestion-native", Files: files, Flags: []string{packageDirectory, "serialization-only-no-debugger-v1", "sanitize=true", "ADAMIC_NATIVE_SPLIT=" + os.Getenv("ADAMIC_NATIVE_SPLIT")}, Toolchain: toolchain}, func(out string) error {
					data, err := os.ReadFile(filepath.Join(product, "lint.c"))
					if err != nil {
						return err
					}
					return native.Build(string(data), filepath.Join(out, "scanner"), native.Options{Sanitize: true})
				})
				p.binary = filepath.Join(nativeProduct, "scanner")
			}
		}
	})
	return &completeSuggestionProductsValue
}

func completeSuggestionShard(t *testing.T, shard int) {
	t.Helper()
	// Setup precedes Parallel, so an individually selected leaf also has all products.
	p := completeSuggestionSetup(t)
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
		got = node(t, p.directory, p.path, false).output
	case 2:
		got = runJavaScript(t, p.module, p.path, false).output
	case 3:
		got = execute(t, "", p.binary, "--manifest", p.path).output
	case 4:
		got = node(t, p.mutant, p.path, false).output
	case 5:
		got = runJavaScript(t, p.mutantModule, p.path, false).output
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
	completeSuggestionSetup(t)
	t.Parallel()
	completeSuggestionShard(t, 0)
}
func TestCompleteSuggestionSerialization_001(t *testing.T) {
	completeSuggestionSetup(t)
	t.Parallel()
	completeSuggestionShard(t, 1)
}
func TestCompleteSuggestionSerialization_002(t *testing.T) {
	completeSuggestionSetup(t)
	t.Parallel()
	completeSuggestionShard(t, 2)
}
func TestCompleteSuggestionSerialization_003(t *testing.T) {
	completeSuggestionSetup(t)
	t.Parallel()
	completeSuggestionShard(t, 3)
}
func TestCompleteSuggestionSerialization_004(t *testing.T) {
	completeSuggestionSetup(t)
	t.Parallel()
	completeSuggestionShard(t, 4)
}
func TestCompleteSuggestionSerialization_005(t *testing.T) {
	completeSuggestionSetup(t)
	t.Parallel()
	completeSuggestionShard(t, 5)
}

// Exercise the real leaves: one altered comparison fails precisely its owner.
func TestCompleteSuggestionSerialization_PlantedFailure(t *testing.T) {
	completeSuggestionSetup(t)
	t.Parallel()
	command := exec.Command(os.Args[0], "-test.run=^TestCompleteSuggestionSerialization_[0-9]{3}$", "-test.timeout=90s", "-test.parallel=4", "-test.v")
	command.Env = append(os.Environ(), "ADAMIC_COMPLETE_SUGGESTION_PLANT=1")
	var output bytes.Buffer
	command.Stdout, command.Stderr = &output, &output
	err := testguard.Run(command, testguard.Budget, 90*time.Second)
	failures := []string{}
	for _, line := range strings.Split(output.String(), "\n") {
		if strings.HasPrefix(line, "--- FAIL:") {
			failures = append(failures, line)
		}
	}
	if err == nil || len(failures) != 1 || !strings.HasPrefix(failures[0], "--- FAIL: TestCompleteSuggestionSerialization_001 ") || !strings.Contains(output.String(), "planted suggestion disagreement") {
		t.Fatalf("wrong planted failure: %v; failures %v\n%s", err, failures, &output)
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
