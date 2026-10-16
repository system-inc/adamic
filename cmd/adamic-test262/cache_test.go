package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/native"
)

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
	t.Parallel()
	cache := &resultCache{directory: t.TempDir()}
	cache.observe("same", func() (execution, bool) { return execution{Stdout: "cached"}, true })
	cache.bypass = true
	got := cache.observe("same", func() (execution, bool) { return execution{Stdout: "fresh"}, true })
	if got.Stdout != "fresh" {
		t.Fatal("bypass reused evidence")
	}
	cache.bypass = false
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
	t.Parallel()
	e, err := prepare(t, "testdata/mini", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	e.cache = &resultCache{directory: t.TempDir()}
	// Prime the passed program, then prove that the uncached serial run
	// executes it again instead of consuming that observation.
	text, err := os.ReadFile("testdata/mini/test/pass/pad.js")
	if err != nil {
		t.Fatal(err)
	}
	if got := e.attempt(classify("pass/pad.js", string(text), false)); got.Kind != outcomePass {
		t.Fatal(got)
	}
	e.cache.bypass = true
	e.profile = newRunProfile()
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
	assertFreshCacheProfile(t, e.profile)
	document := reportDocument{Filters: []filterReport{serial}}
	var serialJSON, serialTable bytes.Buffer
	if err := writeJSON(&serialJSON, document); err != nil {
		t.Fatal(err)
	}
	printTables(&serialTable, document)
	e.cache.bypass = false
	e.jobs = 4
	e.inProcess = true
	for round := 0; round < 2; round++ {
		e.profile = newRunProfile()
		var log, encoded, table bytes.Buffer
		e.log = &log
		parallel, err := e.runFilter("", 0, false)
		if err != nil {
			t.Fatal(err)
		}
		if round == 1 {
			hit := false
			for _, test := range e.profile.Tests {
				for _, phase := range test.Phases {
					if phase.Stage == "node" && phase.Hit {
						hit = true
					}
				}
			}
			if !hit {
				t.Fatal("cached parallel run did not reuse Node observations")
			}
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
	e, err := prepare(t, corpus, t.TempDir())
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
	want := " 50/51 pass=0 fail=0 refused=0 crashed=0 skipped=50\n 51/51 pass=1 fail=0 refused=0 crashed=0 skipped=50\n"
	if log.String() != want || report.Pass != 1 || report.Skipped != 50 {
		t.Fatalf("ordered progress: %q, want %q", log.String(), want)
	}
}

func TestUnavailableCacheStillRuns(t *testing.T) {
	t.Parallel()
	e, err := prepare(t, "testdata/mini", t.TempDir())
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
	e, err := prepare(t, "testdata/mini", t.TempDir())
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
	t.Parallel()
	e, err := prepare(t, "testdata/mini", t.TempDir())
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
	other.cache.bypass = true
	if result := other.attempt(test); result.Kind != outcomePass || other.compiler.command == nil {
		t.Fatalf("bypassed worker reused C cache: %+v", result)
	}
}

func TestCacheBypassPreservesIdentity(t *testing.T) {
	t.Parallel()
	tools, err := resolveTools()
	if err != nil {
		t.Fatal(err)
	}
	before, versionBefore, contextBefore, err := prepareCache("../..", tools)
	if err != nil {
		t.Fatal(err)
	}
	before.bypass = false
	after, versionAfter, contextAfter, err := prepareCache("../..", tools)
	if err != nil {
		t.Fatal(err)
	}
	after.bypass = true
	if contextBefore != contextAfter || before.nodeContext != after.nodeContext {
		t.Fatalf("cache bypass changed identity: context_equal=%t node_equal=%t",
			contextBefore == contextAfter, before.nodeContext == after.nodeContext)
	}
	if before.directory != after.directory || versionBefore != versionAfter {
		t.Fatal("cache bypass changed directory or Node version")
	}
	before.directory = t.TempDir()
	after.directory = before.directory
	before.observe("identity", func() (execution, bool) { return execution{Stdout: "cached"}, true })
	if got := after.observe("identity", func() (execution, bool) { return execution{Stdout: "fresh"}, true }); got.Stdout != "fresh" {
		t.Fatal("bypass reused identity observation")
	}
	if got := before.observe("identity", func() (execution, bool) { t.Fatal("cache identity observation missing"); return execution{}, true }); got.Stdout != "cached" {
		t.Fatal("bypass changed identity observation")
	}
}

// runnerContext runs this binary's TestRunnerLocationHelper standing in directory, as the runner of
// the checkout at root, under environment, and returns the compiler, native and Node contexts it
// derived. A child process, because the working directory and environment are the process's own.
func runnerContext(t *testing.T, directory string, root string, environment []string) string {
	t.Helper()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command(executable, "-test.run=^TestRunnerLocationHelper$")
	command.Dir = directory
	command.Env = append(append([]string{}, environment...),
		"PWD="+directory, "ADAMIC_TEST262_CONTEXT_HELPER=1", "ADAMIC_TEST262_CONTEXT_ROOT="+root)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("helper: %v: %s", err, output)
	}
	return string(output)
}

// The same tree checked out at two paths of different depth keys the same observations. A runner
// standing elsewhere in the tree, where a relative -root or -work names other files, does not.
func TestContextIgnoresCheckoutLocation(t *testing.T) {
	t.Parallel()
	scratch := t.TempDir()
	var packages []string
	for _, checkout := range []string{"a/adamic", "bb/cc/adamic"} {
		directory := filepath.Join(scratch, filepath.FromSlash(checkout), "cmd", "adamic-test262")
		if err := os.MkdirAll(directory, 0700); err != nil {
			t.Fatal(err)
		}
		packages = append(packages, directory)
	}
	environment := os.Environ()
	first := runnerContext(t, packages[0], "../..", environment)
	second := runnerContext(t, packages[1], "../..", environment)
	if first != second {
		t.Fatalf("moving the checkout changed the context: %s vs %s", first, second)
	}
	root := filepath.Join(scratch, "a", "adamic")
	if runnerContext(t, root, ".", environment) == first {
		t.Fatal("a runner standing elsewhere in the checkout reused the package's context")
	}
}

// A variable no child reads leaves the context alone; one that changes what Node prints does not.
func TestContextEnvironment(t *testing.T) {
	t.Parallel()
	directory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	var environment []string
	for _, variable := range os.Environ() {
		name, _, _ := strings.Cut(variable, "=")
		if name != "HOME" && name != "TZ" && name != "ADAMIC_JUNK" {
			environment = append(environment, variable)
		}
	}
	plain := runnerContext(t, directory, "../..", append(append([]string{}, environment...), "HOME="+t.TempDir()))
	junk := runnerContext(t, directory, "../..", append(append([]string{}, environment...), "HOME="+t.TempDir(), "ADAMIC_JUNK=1"))
	if plain != junk {
		t.Fatalf("an unrelated variable and HOME changed the context: %s vs %s", plain, junk)
	}
	// Node's context alone: a time zone changes what Node's Date prints, so Node's observations must not be reused.
	zoned := runnerContext(t, directory, "../..", append(append([]string{}, environment...), "HOME="+t.TempDir(), "TZ=Pacific/Kiritimati"))
	if strings.Fields(zoned)[1] == strings.Fields(plain)[1] {
		t.Fatal("a time zone, which changes Node's Date output, reused Node's context")
	}
}

// Every variable resultEnvironment names is keyed when it's set, even to nothing, and none when none is: the context
// and childEnvironment both read through keyedEnvironment, so a name that went missing from what it keys would be a
// variable children see and the key doesn't name.
func TestEveryResultVariableIsKeyed(t *testing.T) {
	t.Parallel()
	if keyed := keyedEnvironment(func(string) (string, bool) { return "", false }); len(keyed) != 0 {
		t.Fatalf("nothing set keyed %q", keyed)
	}
	seen := map[string]bool{}
	for _, name := range append(append([]string{}, resultEnvironment...), passedEnvironment...) {
		if seen[name] {
			t.Errorf("%s is named twice", name)
		}
		seen[name] = true
	}
	for _, name := range resultEnvironment {
		keyed := keyedEnvironment(func(wanted string) (string, bool) { return "", wanted == name })
		if !reflect.DeepEqual(keyed, []string{name + "="}) {
			t.Errorf("%s set to nothing keyed %q", name, keyed)
		}
	}
}

// A child runs with what the key names and passedEnvironment's, then the engine's own settings, and nothing else of
// the runner's environment: a variable the key doesn't name can't change an observation a hit replays. The child here
// is env itself, run as clang, the native program and Node are. Mutant: runCommandWithLimit back on os.Environ, and
// ADAMIC_JUNK reaches the child.
// Not parallel: sets the runner's environment through t.Setenv.
func TestChildrenSeeOnlyWhatTheKeyNames(t *testing.T) {
	t.Setenv("ADAMIC_JUNK", "1")
	t.Setenv("TZ", "Pacific/Kiritimati")
	allowed := map[string]bool{"ADAMIC_ENGINE_SETTING": true}
	for _, name := range append(append([]string{}, resultEnvironment...), passedEnvironment...) {
		allowed[name] = true
	}
	result := runCommand(time.Minute, []string{"ADAMIC_ENGINE_SETTING=1"}, "/usr/bin/env")
	if result.Exit != 0 {
		t.Fatalf("env: exit %d: %s", result.Exit, result.Stderr)
	}
	seen := map[string]bool{}
	for _, line := range strings.Split(strings.TrimSpace(result.Stdout), "\n") {
		name, _, _ := strings.Cut(line, "=")
		seen[name] = true
		if !allowed[name] {
			t.Errorf("a child saw %s, which the key doesn't name", name)
		}
	}
	for _, name := range []string{"TZ", "ADAMIC_ENGINE_SETTING"} {
		if !seen[name] {
			t.Errorf("a child didn't see %s", name)
		}
	}
}

func TestCacheBypassProcessDefault(t *testing.T) {
	t.Parallel()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{"0", "1"} {
		command := exec.Command(executable, "-test.run=^TestCacheBypassProcessDefaultHelper$")
		command.Env = append(os.Environ(), "ADAMIC_TEST262_CACHE_BYPASS_HELPER=1", "ADAMIC_GATE_UNCACHED="+value)
		output, err := command.CombinedOutput()
		if err != nil {
			t.Fatalf("process default %s: %v: %s", value, err, output)
		}
	}
}

func TestCacheBypassProcessDefaultHelper(t *testing.T) {
	t.Parallel()
	if os.Getenv("ADAMIC_TEST262_CACHE_BYPASS_HELPER") != "1" {
		return
	}
	tools, err := resolveTools()
	if err != nil {
		t.Fatal(err)
	}
	cache, _, _, err := prepareCache("../..", tools)
	if err != nil {
		t.Fatal(err)
	}
	want := os.Getenv("ADAMIC_GATE_UNCACHED") == "1"
	if cache.bypass != want {
		t.Fatalf("process default bypass=%t want %t", cache.bypass, want)
	}
	// This is a dedicated child process, not a change to the package's environment.
	next := "1"
	if want {
		next = "0"
	}
	if err := os.Setenv("ADAMIC_GATE_UNCACHED", next); err != nil {
		t.Fatal(err)
	}
	cache.directory = t.TempDir()
	calls := 0
	for range 2 {
		cache.observe("default", func() (execution, bool) { calls++; return execution{}, true })
	}
	expected := 1
	if want {
		expected = 2
	}
	if calls != expected {
		t.Fatalf("runner default changed with environment: %d executions want %d", calls, expected)
	}
}
