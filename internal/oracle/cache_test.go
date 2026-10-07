package oracle

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/nodepin"
)

// A result is evidence, not a saved verdict: ordinary fixture comparisons and counts still run.
// The specialized stream/permission probes also retain a successful check, keyed by their whole
// harness, source, generated C/JavaScript and toolchains, to avoid repeating their timed protocols.
type recordedRun struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
}

func record(result run) recordedRun {
	return recordedRun{result.stdout, result.stderr, result.exitCode}
}
func (result recordedRun) run() run { return run{result.Stdout, result.Stderr, result.ExitCode} }

type nativeResult struct {
	Run        recordedRun
	LeakReport []byte
}

type resultEnvelope struct {
	Key    string
	Digest string
	Value  json.RawMessage
}

type resultCache struct {
	directory string
	locks     sync.Map
	hits      [3]atomic.Int64
	misses    [3]atomic.Int64
}

const (
	nativeResults = iota
	nodeResults
	checkedProbes
)

var resultKinds = [3]string{"native", "node", "probe"}

// cacheKey length-prefixes every input, preserving arbitrary bytes, boundaries and ordering.
func cacheKey(parts ...string) string {
	hash := sha256.New()
	for _, part := range parts {
		fmt.Fprintf(hash, "%d:", len(part))
		hash.Write([]byte(part))
	}
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func nativeResultKey(code, runtimeKey, nodeVersion, context string) string {
	return cacheKey("adamic-native-result-v1", code, runtimeKey, nodeVersion, context)
}

func nodeResultKey(source, javascript, nodeVersion, context string) string {
	return cacheKey("adamic-node-result-v1", source, javascript, nodeVersion, context)
}

func (cache *resultCache) read(key string, value any) bool {
	contents, err := os.ReadFile(filepath.Join(cache.directory, key+".json"))
	if err != nil {
		return false
	}
	var envelope resultEnvelope
	if json.Unmarshal(contents, &envelope) != nil || envelope.Key != key || envelope.Digest != cacheKey(string(envelope.Value)) {
		return false
	}
	return json.Unmarshal(envelope.Value, value) == nil
}

func (cache *resultCache) write(t *testing.T, key string, value any) {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	contents, err := json.Marshal(resultEnvelope{key, cacheKey(string(encoded)), encoded})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(cache.directory, 0o700); err != nil {
		t.Fatal(err)
	}
	temporary, err := os.CreateTemp(cache.directory, ".result-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(temporary.Name())
	if _, err := temporary.Write(contents); err != nil {
		temporary.Close()
		t.Fatal(err)
	}
	if err := temporary.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(temporary.Name(), filepath.Join(cache.directory, key+".json")); err != nil {
		t.Fatal(err)
	}
}

func cachedResult[T any](t *testing.T, cache *resultCache, kind int, key string, execute func() T) T {
	t.Helper()
	// Check dynamically so a forced integration run bypasses both reads and writes, including probes.
	if os.Getenv("ADAMIC_GATE_UNCACHED") == "1" {
		cache.misses[kind].Add(1)
		return execute()
	}
	return reusableResult(t, cache, kind, key, execute)
}

// The primitive is tested with isolated directories even during an uncached integration gate.
// Fixture callers always go through cachedResult, which owns the environment bypass.
func reusableResult[T any](t *testing.T, cache *resultCache, kind int, key string, execute func() T) T {
	t.Helper()
	value, _ := cache.locks.LoadOrStore(key, &sync.Mutex{})
	lock := value.(*sync.Mutex)
	lock.Lock()
	defer lock.Unlock()
	var result T
	if cache.read(key, &result) {
		cache.hits[kind].Add(1)
		t.Logf("gate cache %s hit", resultKinds[kind])
		return result
	}
	cache.misses[kind].Add(1)
	t.Logf("gate cache %s miss", resultKinds[kind])
	result = execute()
	// Never publish a result after a harness failure. Failed observations may still be compared by
	// the caller; they remain evidence, and every cached observation is compared again on a hit.
	if !t.Failed() {
		cache.write(t, key, result)
	}
	return result
}

type gateIdentity struct {
	cache       *resultCache
	context     string
	nodeVersion string
	libraries   [4]string // by nativeVariant: sanitized, release, counted, slabs
}

var gateOnce sync.Once
var gate gateIdentity
var gateError error

func identity(t *testing.T) gateIdentity {
	t.Helper()
	gateOnce.Do(func() {
		directory, err := os.UserCacheDir()
		if err != nil {
			gateError = err
			return
		}
		gate.cache = &resultCache{directory: filepath.Join(directory, "adamic", "gate")}
		version, err := exec.Command("node", "--version").CombinedOutput()
		if err != nil {
			gateError = fmt.Errorf("node --version: %w: %s", err, version)
			return
		}
		gate.nodeVersion = string(version)
		for index, options := range []native.Options{{Sanitize: true}, {}, {Count: true}, {Sanitize: true, Slabs: true}} {
			library, err := native.RuntimeLibrary("", options)
			if err != nil {
				gateError = err
				return
			}
			gate.libraries[index] = filepath.Base(filepath.Dir(library))
		}
		// The cache holds what these exact checks observed. Editing a harness or oracle runner must
		// invalidate it even when a fixture's generated code happens to remain identical.
		files, err := filepath.Glob("*.go")
		if err != nil {
			gateError = err
			return
		}
		runners, err := filepath.Glob(filepath.Join(repository, "oracle", "*.mjs"))
		if err != nil {
			gateError = err
			return
		}
		files = append(files, runners...)
		files = append(files, repository+"/internal/native/native.go", repository+"/internal/native/library.go")
		parts := []string{"gate-context-v1", runtime.GOOS, runtime.GOARCH, fmt.Sprint(os.Geteuid())}
		root, err := filepath.Abs(repository)
		if err != nil {
			gateError = err
			return
		}
		parts = append(parts, root)
		var stack syscall.Rlimit
		if err := syscall.Getrlimit(syscall.RLIMIT_STACK, &stack); err != nil {
			gateError = err
			return
		}
		parts = append(parts, fmt.Sprint(stack.Cur, stack.Max))
		environment := os.Environ()
		sort.Strings(environment)
		for _, variable := range environment {
			if !strings.HasPrefix(variable, "ADAMIC_GATE_UNCACHED=") && !strings.HasPrefix(variable, "ADAMIC_ORACLE_JSON_TYPES=") {
				parts = append(parts, variable)
			}
		}
		for _, file := range files {
			contents, err := os.ReadFile(file)
			if err != nil {
				gateError = err
				return
			}
			parts = append(parts, file, string(contents))
		}
		gate.context = jsonTypesContext(cacheKey(parts...), jsonTypesDigest)
	})
	if gateError != nil {
		t.Fatal(gateError)
	}
	return gate
}

// Reuse import declarations already parsed by lowered, keyed by the current source bytes. This
// saves a second type-check solely to discover dependencies; edits still read and hash fresh bytes.
var sourceImports sync.Map

func rememberImports(path string, text string, program *load.Program) []string {
	var imports []string
	for _, file := range program.Files() {
		for _, statement := range file.Statements.Nodes {
			if statement.Kind != ast.KindImportDeclaration && statement.Kind != ast.KindExportDeclaration {
				continue
			}
			specifier := statement.ModuleSpecifier()
			if specifier != nil && strings.HasPrefix(specifier.Text(), ".") {
				imports = append(imports, specifier.Text())
			}
		}
	}
	sourceImports.Store(cacheKey(path, text), imports)
	return imports
}

// sourceIdentity follows the checker's actual import declarations, not a regex over comments or
// strings. All module bytes and relative names participate. Temporary generated entry paths do
// not, so the counter sweep can reuse its observations across test processes.
func sourceIdentity(t *testing.T, path string) string {
	t.Helper()
	root := filepath.Dir(path)
	visited := map[string]bool{}
	var parts []string
	var visit func(string)
	visit = func(path string) {
		absolute, err := filepath.Abs(path)
		if err != nil {
			t.Fatal(err)
		}
		if visited[absolute] {
			return
		}
		visited[absolute] = true
		contents, err := os.ReadFile(absolute)
		if err != nil {
			t.Fatal(err)
		}
		relative, err := filepath.Rel(root, absolute)
		if err != nil {
			t.Fatal(err)
		}
		parts = append(parts, relative, string(contents))
		if filepath.Ext(absolute) == ".mjs" {
			return
		}
		value, ok := sourceImports.Load(cacheKey(absolute, string(contents)))
		var imports []string
		if ok {
			imports = value.([]string)
		} else {
			program, err := load.Load([]string{absolute})
			if err != nil {
				t.Fatal(err)
			}
			imports = rememberImports(absolute, string(contents), program)
		}
		for _, specifier := range imports {
			visit(filepath.Join(filepath.Dir(absolute), specifier))
		}

	}
	visit(path)
	return cacheKey(parts...)
}

func cachedNode(t *testing.T, path string, execute func() run) run {
	t.Helper()
	given := identity(t)
	source, javascript := "", ""
	if filepath.Ext(path) == ".mjs" {
		contents, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		javascript = string(contents)
	} else {
		source = sourceIdentity(t, path)
	}
	key := nodeResultKey(source, javascript, given.nodeVersion, given.context)
	return cachedResult(t, given.cache, nodeResults, key, func() recordedRun { return record(execute()) }).run()
}

func cachedNative(t *testing.T, program *ir.Program, variant nativeVariant) nativeResult {
	t.Helper()
	given := identity(t)
	code := native.C(program)
	key := nativeResultKey(code, given.libraries[variant], given.nodeVersion, cacheKey(given.context, fmt.Sprint(int(variant))))
	return cachedResult(t, given.cache, nativeResults, key, func() nativeResult {
		switch variant {
		case releaseBuild:
			return nativeResult{Run: record(releasedUncached(t, program))}
		case slabsBuild:
			return nativeResult{Run: record(slabbedUncached(t, program))}
		}
		result, binary := nativelyUncached(t, program)
		leak := ""
		if result.exitCode == 0 {
			leak = leaksUncached(t, program, binary)
		}
		return nativeResult{record(result), []byte(leak)}
	})
}

// A probe is replayed only after all of its fixture inputs and generated outputs are checked. Its
// callback runs the original protocol, and a failure is never saved as a successful check.
type recordedFile struct {
	Path  []byte
	Value []byte
}

type probeEvidence struct {
	Passed bool
	Runs   []recordedRun
	Leaks  [][]byte
	Files  [][]recordedFile
}

var probeRuns sync.Map // test name -> evidence for input/permission protocols

func rememberRun(t *testing.T, result run) {
	if value, ok := probeRuns.Load(t.Name()); ok {
		evidence := value.(*probeEvidence)
		evidence.Runs = append(evidence.Runs, record(result))
	}
}
func rememberLeak(t *testing.T, leak string) {
	if value, ok := probeRuns.Load(t.Name()); ok {
		evidence := value.(*probeEvidence)
		evidence.Leaks = append(evidence.Leaks, []byte(leak))
	}
}

// Input fixtures read these directories, as well as their sources. File modes, symlink targets and
// directory names matter as much as contents. Do not hash neighboring .a fixtures: changing one
// unrelated program must not invalidate another fixture's file inputs.
func inputIdentity(t *testing.T) string {
	t.Helper()
	var parts []string
	for _, name := range []string{"reading", "walking"} {
		root := filepath.Join(repository, "internal", "oracle", "testdata", name)
		err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			info, err := os.Lstat(path)
			if err != nil {
				return err
			}
			parts = append(parts, path, fmt.Sprint(info.Mode()))
			if info.Mode()&os.ModeSymlink != 0 {
				target, err := os.Readlink(path)
				if err != nil {
					return err
				}
				parts = append(parts, target)
			} else if !entry.IsDir() {
				contents, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				parts = append(parts, string(contents))
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	return cacheKey(parts...)
}

func cacheProbe(t *testing.T, path string, program *ir.Program, external string, execute func()) {
	t.Helper()
	if os.Getenv("ADAMIC_GATE_UNCACHED") == "1" {
		identity(t).cache.misses[checkedProbes].Add(1)
		execute()
		return
	}
	if program == nil {
		absolute, err := filepath.Abs(filepath.Join(repository, path))
		if err != nil {
			t.Fatal(err)
		}
		path = absolute
		program, err = lowered(t, path)
		if err != nil {
			t.Fatal(err)
		}
	}
	given := identity(t)
	context := cacheKey(given.context, t.Name(), sourceIdentity(t, path), javascript.JavaScript(program), external)
	key := nativeResultKey(native.C(program), cacheKey(given.libraries[:]...), given.nodeVersion, context)
	evidence := cachedResult(t, given.cache, checkedProbes, key, func() probeEvidence {
		evidence := &probeEvidence{}
		probeRuns.Store(t.Name(), evidence)
		defer probeRuns.Delete(t.Name())
		execute()
		evidence.Passed = !t.Failed() && !t.Skipped()
		return *evidence
	})
	if !evidence.Passed && !t.Failed() {
		t.Fatal("cached probe did not complete successfully")
	}
}

func TestMain(main *testing.M) {
	path, err := nodepin.Check()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "oracle: node %s at %s\n", nodepin.Version, path)
	directory, err := prepareJSONTypes()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	status := main.Run()
	if directory != "" {
		os.RemoveAll(directory)
	}
	if gate.cache != nil {
		for index, kind := range resultKinds {
			fmt.Printf("gate cache: %s hits=%d misses=%d\n", kind, gate.cache.hits[index].Load(), gate.cache.misses[index].Load())
		}
	}
	os.Exit(status)
}

// Changing generated C is the lowering-change case: the source and both toolchains stay put.
func TestGateCacheGeneratedC(t *testing.T) {
	t.Parallel()
	cache := &resultCache{directory: t.TempDir()}
	for _, word := range []string{"before", "after"} {
		code := fmt.Sprintf("#include <stdio.h>\nint main(void) { puts(%q); return 0; }\n", word)
		key := nativeResultKey(code, "runtime", "node", "harness")
		result := reusableResult(t, cache, nativeResults, key, func() nativeResult {
			binary := filepath.Join(t.TempDir(), "program")
			if err := native.Build(code, binary, native.Options{}); err != nil {
				t.Fatal(err)
			}
			return nativeResult{Run: record(execute(t, binary))}
		})
		if string(result.Run.Stdout) != word+"\n" {
			t.Fatalf("changed generated C served stale stdout %q, want %q", result.Run.Stdout, word+"\n")
		}
	}
}

func TestGateCacheToolchains(t *testing.T) {
	t.Parallel()
	for _, dimension := range []string{"runtime", "node", "node runner"} {
		t.Run(dimension, func(t *testing.T) {
			t.Parallel()
			cache := &resultCache{directory: t.TempDir()}
			for _, version := range []string{"old", "new"} {
				library, node := "runtime", "node"
				if dimension == "runtime" {
					library = version
				} else {
					node = version
				}
				key := nativeResultKey("code", library, node, "harness")
				if dimension == "node runner" {
					key = nodeResultKey("source", "javascript", node, "harness")
				}
				result := reusableResult(t, cache, nativeResults, key, func() string { return version })
				if result != version {
					t.Fatalf("%s changed but served %q, want %q", dimension, result, version)
				}
			}
		})
	}
}

func TestGateCacheNodeInputs(t *testing.T) {
	t.Parallel()
	original := nodeResultKey("source", "javascript", "node", "context")
	for _, changed := range []string{nodeResultKey("edited", "javascript", "node", "context"), nodeResultKey("source", "edited", "node", "context")} {
		if changed == original {
			t.Fatal("Node input change did not invalidate observation")
		}
	}
	directory := t.TempDir()
	entry := filepath.Join(directory, "main.ts")
	dependency := filepath.Join(directory, "part.ts")
	if err := os.WriteFile(entry, []byte("import { value } from './part.ts';\nconsole.log(value);\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dependency, []byte("export const value = 'first';\n"), 0600); err != nil {
		t.Fatal(err)
	}
	before := sourceIdentity(t, entry)
	if err := os.WriteFile(dependency, []byte("export const value = 'other';\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if sourceIdentity(t, entry) == before {
		t.Fatal("an imported module changed without invalidating Node's source key")
	}
}

func TestGateCacheAtomicEvidence(t *testing.T) {
	t.Parallel()
	cache := &resultCache{directory: t.TempDir()}
	key := cacheKey("same key")
	want := nativeResult{Run: recordedRun{Stdout: []byte{0, 255, 'x'}, Stderr: []byte{254, 0, 'y'}, ExitCode: 70}, LeakReport: []byte{255, 0, 'l'}}
	var builds atomic.Int64
	var workers sync.WaitGroup
	for range 32 {
		workers.Go(func() {
			got := reusableResult(t, cache, nativeResults, key, func() nativeResult { builds.Add(1); return want })
			if !bytes.Equal(got.Run.Stdout, want.Run.Stdout) || !bytes.Equal(got.Run.Stderr, want.Run.Stderr) || got.Run.ExitCode != 70 || !bytes.Equal(got.LeakReport, want.LeakReport) {
				t.Error("cached evidence lost bytes, exit code or leak report")
			}
		})
	}
	workers.Wait()
	if builds.Load() != 1 {
		t.Fatalf("same key executed %d times, want one", builds.Load())
	}
	// A partial or corrupted entry is a miss, never trusted evidence.
	if err := os.WriteFile(filepath.Join(cache.directory, key+".json"), []byte(`{"Key":"wrong","Value":null}`), 0600); err != nil {
		t.Fatal(err)
	}
	reusableResult(t, cache, nativeResults, key, func() nativeResult { builds.Add(1); return want })
	if builds.Load() != 2 {
		t.Fatal("corrupt observation was reused")
	}
	if disagreement(run{stdout: []byte("different")}, want.Run.run()) == "" {
		t.Fatal("replaying evidence suppressed the output comparison")
	}
}

func TestGateCacheUncached(t *testing.T) {
	// Not parallel: Setenv changes the process-wide gate bypass.
	t.Setenv("ADAMIC_GATE_UNCACHED", "0")
	cache := &resultCache{directory: t.TempDir()}
	key := cacheKey("forced run")
	cachedResult(t, cache, nativeResults, key, func() string { return "cached" })
	t.Setenv("ADAMIC_GATE_UNCACHED", "1")
	if got := cachedResult(t, cache, nativeResults, key, func() string { return "fresh" }); got != "fresh" {
		t.Fatalf("uncached run reused %q", got)
	}
	t.Setenv("ADAMIC_GATE_UNCACHED", "0")
	if got := cachedResult(t, cache, nativeResults, key, func() string { t.Fatal("old evidence disappeared"); return "" }); got != "cached" {
		t.Fatal("uncached run wrote its results into the cache")
	}
}
