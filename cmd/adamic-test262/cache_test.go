package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/native"
)

func TestNodeCacheProgram(t *testing.T) {
	t.Parallel()
	cache := &resultCache{directory: t.TempDir()}
	module := filepath.Join(t.TempDir(), "program.mts")
	for _, word := range []string{"before", "after"} {
		source := fmt.Sprintf("console.log(%q);\n", word)
		if err := os.WriteFile(module, []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
		key := nodeResultKey(source, "node", "adapt", module, "context")
		result := cache.reuse(key, func() (execution, bool) {
			return runCommand(time.Second*15, nil, "node", module), true
		})
		if result.Stdout != word+"\n" {
			t.Fatalf("changed Node program served %q, want %q", result.Stdout, word+"\n")
		}
	}
}

func TestNativeCacheGeneratedC(t *testing.T) {
	t.Parallel()
	cache := &resultCache{directory: t.TempDir()}
	binary := filepath.Join(t.TempDir(), "program")
	for _, word := range []string{"before", "after"} {
		code := fmt.Sprintf("#include <stdio.h>\nint main(void) { puts(%q); return 0; }\n", word)
		key := nativeResultKey(code, "runtime", binary, "context")
		result := cache.reuse(key, func() (execution, bool) {
			if err := native.Build(code, binary, native.Options{}); err != nil {
				t.Fatal(err)
			}
			return runCommand(15*time.Second, nil, binary), true
		})
		if result.Stdout != word+"\n" {
			t.Fatalf("changed emitted C served %q, want %q", result.Stdout, word+"\n")
		}
	}
}

func TestCacheKeyDimensions(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"node", "native"} {
		count := 4
		if kind == "node" {
			count = 5
		}
		for dimension := 0; dimension < count; dimension++ {
			cache := &resultCache{directory: t.TempDir()}
			for _, version := range []string{"old", "new"} {
				parts := []string{"code", "tool", "command", "context", "adaptation"}
				parts[dimension] = version
				key := nativeResultKey(parts[0], parts[1], parts[2], parts[3])
				if kind == "node" {
					key = nodeResultKey(parts[0], parts[1], parts[4], parts[2], parts[3])
				}
				result := cache.reuse(key, func() (execution, bool) { return execution{Stdout: version}, true })
				if result.Stdout != version {
					t.Fatalf("%s dimension %d changed but served %q", kind, dimension, result.Stdout)
				}
			}
		}
	}
	if cacheKey("ab", "c") == cacheKey("a", "bc") {
		t.Fatal("key boundaries lost")
	}
}

func TestCacheAtomicBytes(t *testing.T) {
	t.Parallel()
	cache := &resultCache{directory: t.TempDir()}
	want := execution{Stdout: string([]byte{0, 255, 'x'}), Stderr: string([]byte{254, 0, 'y'}), Exit: -1, Signal: "aborted"}
	var calls atomic.Int64
	var workers sync.WaitGroup
	for range 16 {
		workers.Go(func() {
			got := cache.reuse("same", func() (execution, bool) { calls.Add(1); return want, true })
			if !reflect.DeepEqual(got, want) {
				t.Error("cache lost execution bytes or status")
			}
		})
	}
	workers.Wait()
	if calls.Load() != 1 {
		t.Fatalf("executed %d times", calls.Load())
	}
	if err := os.WriteFile(filepath.Join(cache.directory, "same.json"), []byte(`{"Key":"same","Digest":"wrong","Value":{}}`), 0600); err != nil {
		t.Fatal(err)
	}
	cache.reuse("same", func() (execution, bool) { calls.Add(1); return want, true })
	if calls.Load() != 2 {
		t.Fatal("corrupt evidence reused")
	}
}

func TestCacheBypass(t *testing.T) {
	// Not parallel: the bypass changes process environment.
	t.Setenv("ADAMIC_GATE_UNCACHED", "0")
	cache := &resultCache{directory: t.TempDir()}
	cache.observe("same", func() (execution, bool) { return execution{Stdout: "cached"}, true })
	t.Setenv("ADAMIC_GATE_UNCACHED", "1")
	got := cache.observe("same", func() (execution, bool) { return execution{Stdout: "fresh"}, true })
	if got.Stdout != "fresh" {
		t.Fatal("bypass reused evidence")
	}
	t.Setenv("ADAMIC_GATE_UNCACHED", "0")
	got = cache.observe("same", func() (execution, bool) { t.Fatal("cache missing"); return execution{}, true })
	if got.Stdout != "cached" {
		t.Fatal("bypass wrote evidence")
	}
	for _, result := range []execution{{TimedOut: true, Exit: -1}, {Exit: -1}} {
		for range 2 {
			cache.observe("transient", func() (execution, bool) { return result, true })
		}
		if _, err := os.Stat(filepath.Join(cache.directory, "transient.json")); !os.IsNotExist(err) {
			t.Fatal("transient failure cached")
		}
	}
}

func TestParallelCachedMatchesSerial(t *testing.T) {
	// Not parallel: compare both modes while changing the bypass environment.
	t.Setenv("ADAMIC_GATE_UNCACHED", "1")
	e, err := prepare("../..", "testdata/mini", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	e.cache = &resultCache{directory: t.TempDir()}
	var serialLog bytes.Buffer
	e.log = &serialLog
	e.jobs = 1
	serial, err := e.runFilter("", 0, false)
	if err != nil {
		t.Fatal(err)
	}
	if serial.Pass == 0 || serial.Refused == 0 || serial.Skipped == 0 || serial.Fail == 0 || serial.Crashed == 0 {
		t.Fatalf("fixture lacks verdict coverage: %+v", serial)
	}
	document := reportDocument{Filters: []filterReport{serial}}
	var serialJSON, serialTable bytes.Buffer
	if err := writeJSON(&serialJSON, document); err != nil {
		t.Fatal(err)
	}
	printTables(&serialTable, document)
	t.Setenv("ADAMIC_GATE_UNCACHED", "0")
	e.jobs = 4
	e.inProcess = true
	for round := 0; round < 2; round++ {
		var log, encoded, table bytes.Buffer
		e.log = &log
		parallel, err := e.runFilter("", 0, false)
		if err != nil {
			t.Fatal(err)
		}
		document.Filters = []filterReport{parallel}
		if err := writeJSON(&encoded, document); err != nil {
			t.Fatal(err)
		}
		printTables(&table, document)
		if !bytes.Equal(serialJSON.Bytes(), encoded.Bytes()) || !bytes.Equal(serialTable.Bytes(), table.Bytes()) || !bytes.Equal(serialLog.Bytes(), log.Bytes()) {
			t.Fatalf("round %d differs:\nserial=%s\nparallel=%s", round, serialJSON.Bytes(), encoded.Bytes())
		}
	}
	// The limit selects the same attempted files before workers are scheduled.
	e.jobs = 1
	var log bytes.Buffer
	e.log = &log
	limited, err := e.runFilter("", 1, false)
	if err != nil {
		t.Fatal(err)
	}
	e.jobs = 4
	parallel, err := e.runFilter("", 1, false)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(limited, parallel) {
		t.Fatal("parallel limit selected different tests")
	}
}

func TestOrderedProgress(t *testing.T) {
	t.Parallel()
	corpus := t.TempDir()
	directory := filepath.Join(corpus, "test")
	if err := os.MkdirAll(directory, 0700); err != nil {
		t.Fatal(err)
	}
	skipped, err := os.ReadFile("testdata/mini/test/skip/negative.js")
	if err != nil {
		t.Fatal(err)
	}
	for index := 0; index < 50; index++ {
		if err := os.WriteFile(filepath.Join(directory, fmt.Sprintf("%02d.js", index)), skipped, 0600); err != nil {
			t.Fatal(err)
		}
	}
	passed, err := os.ReadFile("testdata/mini/test/pass/pad.js")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "50.js"), passed, 0600); err != nil {
		t.Fatal(err)
	}
	e, err := prepare("../..", corpus, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	e.inProcess = true
	e.jobs = 4
	var log bytes.Buffer
	e.log = &log
	report, err := e.runFilter("", 0, false)
	if err != nil {
		t.Fatal(err)
	}
	want := " 50/51 pass=0 fail=0 refused=0 not-typescript=0 crashed=0 skipped=50\n 51/51 pass=1 fail=0 refused=0 not-typescript=0 crashed=0 skipped=50\n"
	if log.String() != want || report.Pass != 1 || report.Skipped != 50 {
		t.Fatalf("ordered progress: %q, want %q", log.String(), want)
	}
}

func TestUnavailableCacheStillRuns(t *testing.T) {
	t.Parallel()
	e, err := prepare("../..", "testdata/mini", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(path, []byte("file"), 0600); err != nil {
		t.Fatal(err)
	}
	e.cache = &resultCache{directory: path}
	source, err := os.ReadFile("testdata/mini/test/pass/pad.js")
	if err != nil {
		t.Fatal(err)
	}
	result := e.attempt(classify("pass/pad.js", string(source), false))
	if result.Kind != outcomePass {
		t.Fatalf("unavailable result cache changed verdict: %+v", result)
	}
}

func TestImportedInputsRunFresh(t *testing.T) {
	t.Parallel()
	e, err := prepare("../..", "testdata/mini", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	e.cache = &resultCache{directory: t.TempDir()}
	input := filepath.Join(t.TempDir(), "input")
	source := fmt.Sprintf("import { readTextFile } from 'adamic';\nconst value = readTextFile(%q);\nif (value.kind === 'Ok' && value.text === 'after') { throw new Error('changed'); }\n", input)
	test := classify("input.js", source, false)
	var reasons []string
	for _, word := range []string{"before", "after"} {
		if err := os.WriteFile(input, []byte(word), 0600); err != nil {
			t.Fatal(err)
		}
		result := e.attempt(test)
		if result.Kind != outcomeFail {
			t.Fatalf("Node cannot import the native runtime; want fail, got %+v", result)
		}
		reasons = append(reasons, result.Reason)
	}
	if reasons[0] == reasons[1] {
		t.Fatalf("mutable imported input served stale native observation: %v", reasons)
	}
}

func TestCompilerCacheProgram(t *testing.T) {
	t.Parallel()
	cache := &resultCache{directory: t.TempDir()}
	path := filepath.Join(t.TempDir(), "program.a")
	for _, word := range []string{"before", "after"} {
		source := program(fmt.Sprintf("console.log(%q);", word))
		if err := os.WriteFile(path, []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
		key := compilerResultKey(source, "compiler", path, "context")
		compiled := cache.reuse(key, func() (execution, bool) {
			value := compileInProcess(path)
			if value.Exit != 0 {
				t.Fatal(value.Stderr)
			}
			return value, true
		})
		binary := filepath.Join(t.TempDir(), "program")
		if err := native.Build(compiled.Stdout, binary, native.Options{}); err != nil {
			t.Fatal(err)
		}
		result := runCommand(15*time.Second, nil, binary)
		if result.Stdout != word+"\n" {
			t.Fatalf("changed compiler program served %q, want %q", result.Stdout, word+"\n")
		}
	}
	for dimension := 0; dimension < 4; dimension++ {
		cache := &resultCache{directory: t.TempDir()}
		for _, version := range []string{"old", "new"} {
			parts := []string{"source", "compiler", "command", "context"}
			parts[dimension] = version
			result := cache.reuse(compilerResultKey(parts[0], parts[1], parts[2], parts[3]), func() (execution, bool) { return execution{Stdout: version}, true })
			if result.Stdout != version {
				t.Fatalf("compiler key dimension %d served stale output", dimension)
			}
		}
	}
	if !dependentProgram("/// <reference path='external.d.ts' />\nconst x = 1;") {
		t.Fatal("reference dependency treated as closed")
	}
}

func TestCompilerCacheAcrossScratchDirectories(t *testing.T) {
	// Not parallel: explicitly enable reuse even in the uncached integration gate.
	t.Setenv("ADAMIC_GATE_UNCACHED", "0")
	e, err := prepare("../..", "testdata/mini", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	e.cache = &resultCache{directory: t.TempDir()}
	e.compiler = &compilerWorker{}
	defer e.compiler.close()
	source, err := os.ReadFile("testdata/mini/test/pass/pad.js")
	if err != nil {
		t.Fatal(err)
	}
	test := classify("pass/pad.js", string(source), false)
	if result := e.attempt(test); result.Kind != outcomePass {
		t.Fatalf("priming C cache: %+v", result)
	}
	other := *e
	other.work = t.TempDir()
	other.adamic = filepath.Join(other.work, "not-used-by-worker")
	other.compiler = &compilerWorker{}
	defer other.compiler.close()
	if result := other.attempt(test); result.Kind != outcomePass || other.compiler.command != nil {
		t.Fatalf("worker C cache missed when only scratch compiler path changed: %+v", result)
	}
}

// Not parallel: changes the process-wide ADAMIC_GATE_UNCACHED environment.
func TestCacheBypassPreservesIdentity(t *testing.T) {
	t.Setenv("ADAMIC_GATE_UNCACHED", "0")
	before, versionBefore, contextBefore, err := prepareCache()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("ADAMIC_GATE_UNCACHED", "1")
	after, versionAfter, contextAfter, err := prepareCache()
	if err != nil {
		t.Fatal(err)
	}
	if contextBefore != contextAfter || before.nodeContext != after.nodeContext {
		t.Fatalf("cache bypass changed identity: context_equal=%t node_equal=%t",
			contextBefore == contextAfter, before.nodeContext == after.nodeContext)
	}
	if before.directory != after.directory || versionBefore != versionAfter {
		t.Fatal("cache bypass changed directory or Node version")
	}
}
