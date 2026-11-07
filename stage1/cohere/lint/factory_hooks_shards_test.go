package lint

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
)

const testFactoryHooksShards = 5

type factoryHooksCase struct {
	name    string
	mutated bool
	backend string
}

// This pinned corpus is the original test's three positive and two mutant checks.
func factoryHooksCases() []factoryHooksCase {
	return []factoryHooksCase{
		{"Node", false, "node"}, {"emitted JavaScript", false, "javascript"},
		{"native", false, "native"}, {"Node mutant", true, "node"},
		{"emitted JavaScript mutant", true, "javascript"},
	}
}

func factoryHooksSlice(index int) []factoryHooksCase {
	cases := factoryHooksCases()
	return cases[index*len(cases)/testFactoryHooksShards : (index+1)*len(cases)/testFactoryHooksShards]
}

func factoryHooksMatches(output []byte, mutated bool) bool {
	return bytes.HasPrefix(output, []byte("case 0\nfactory\nprepare\nvisit\nfinish\n")) != mutated
}

func TestFactoryHooksUnion(test *testing.T) {
	test.Parallel()
	registered := []func(*testing.T){
		TestFactoryHooks_000,
		TestFactoryHooks_001,
		TestFactoryHooks_002,
		TestFactoryHooks_003,
		TestFactoryHooks_004,
	}
	if len(registered) != testFactoryHooksShards {
		test.Fatalf("registered %d shards, want %d", len(registered), testFactoryHooksShards)
	}
	seen := map[string]int{}
	for i := 0; i < testFactoryHooksShards; i++ {
		for _, c := range factoryHooksSlice(i) {
			seen[c.name]++
		}
	}
	for _, c := range factoryHooksCases() {
		if seen[c.name] != 1 {
			test.Fatalf("%s covered %d times", c.name, seen[c.name])
		}
	}
	if len(seen) != len(factoryHooksCases()) {
		test.Fatal("unexpected case")
	}
	test.Logf("union: %d cases, each exactly once across %d shards", len(seen), testFactoryHooksShards)
}

func TestFactoryHooksPlantedFailure(test *testing.T) {
	test.Parallel()
	caught := []int{}
	for i := 0; i < testFactoryHooksShards; i++ {
		for _, c := range factoryHooksSlice(i) {
			output := []byte("case 0\nfactory\nprepare\nvisit\nfinish\n")
			if c.mutated {
				output = []byte("case 0\nfactory\nprepare\nvisit\n")
			}
			if c.name == "native" {
				output = []byte("case 0\nfactory\nvisit\nprepare\nfinish\n")
			}
			if !factoryHooksMatches(output, c.mutated) {
				caught = append(caught, i)
			}
		}
	}
	if len(caught) != 1 || caught[0] != 2 {
		test.Fatalf("planted disagreement caught by %v", caught)
	}
	test.Log("planted native hook-order disagreement caught only by shard 002")
}

func factoryHooksDirectory(test *testing.T) string {
	test.Helper()
	directory, temporaryDirectoryError := os.MkdirTemp(sharedDirectory, "factory-hooks-")
	if temporaryDirectoryError != nil {
		test.Fatal(temporaryDirectoryError)
	}
	return directory
}

func factoryHooksManifest(test *testing.T, directory string, rows []string) string {
	test.Helper()
	path := filepath.Join(directory, "manifest")
	if writeError := os.WriteFile(path, []byte(strings.Join(rows, "\n")+"\n"), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	return path
}

func factoryHooksPrepare(test *testing.T, mutated bool) (string, string) {
	directory := copyPort(test, factoryHooksDirectory(test), "", "")
	// This test selects only no-debugger. Keep the same linter and rule source,
	// but omit unrelated rule factories from this private fixture's module graph.
	entries, readDirectoryError := os.ReadDir(filepath.Join(directory, "rules"))
	if readDirectoryError != nil {
		test.Fatal(readDirectoryError)
	}
	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != "no-debugger" {
			if removeError := os.RemoveAll(filepath.Join(directory, "rules", entry.Name())); removeError != nil {
				test.Fatal(removeError)
			}
		}
	}
	module := filepath.Join(directory, "rules/no-debugger/rule.ts")
	data, readError := os.ReadFile(module)
	if readError != nil {
		test.Fatal(readError)
	}
	source := strings.Replace(string(data), "    visit(index: number, parent: number): void {", `    ready = false;
    prepare(root: number): void { this.ready = true; console.log('prepare'); }
    finish(root: number): void { console.log('finish'); }
    visit(index: number, parent: number): void {
        if(!this.ready) { panic('visit before prepare'); }
        console.log('visit');`, 1)
	source = strings.Replace(
		source,
		"return new Rule(context);",
		`if(context.node(context.parents.length - 1).kind !== 'SourceFile') { panic('factory before ancestry'); }
    if(context.settings.read('number', '') !== '-2' || !context.settings.read('payload', '').includes('enabled')) { panic('structured option lost'); }
    console.log('factory');
    return new Rule(context);`,
		1,
	)
	source = "import { panic } from 'adamic';\n" + source
	if writeError := os.WriteFile(module, []byte(source), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	descriptor := filepath.Join(directory, "rules/no-debugger/rule.json")
	data, readError = os.ReadFile(descriptor)
	if readError != nil {
		test.Fatal(readError)
	}
	var options map[string]any
	if decodeError := json.Unmarshal(data, &options); decodeError != nil {
		test.Fatal(decodeError)
	}
	options["prepare"] = "prepare"
	options["finish"] = "finish"
	write := func() {
		data, encodeError := json.Marshal(options)
		if encodeError != nil {
			test.Fatal(encodeError)
		}
		if writeError := os.WriteFile(descriptor, data, 0644); writeError != nil {
			test.Fatal(writeError)
		}
	}
	if mutated {
		delete(options, "finish")
	}
	write()
	fixture := filepath.Join(directory, "hooks.ts")
	if writeError := os.WriteFile(fixture, []byte("debugger;\n"), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	path := factoryHooksManifest(
		test,
		directory,
		[]string{fixture + "\tno-debugger\t\t\tfalse\t{\"Number\":-2,\"Payload\":{\"enabled\":true}}"},
	)

	prepareRegistry(test, directory)
	return directory, path
}

// Each shard prepares these immutable products once before its own deadline.
var factoryHooksSetupOnce sync.Once
var factoryHooksReady *factoryHooksPrepared

type factoryHooksPrepared struct {
	products       [2]factoryHooksProduct
	javascript     [2]string
	native, runner string
}

func factoryHooksSetup(test *testing.T) {
	test.Helper()
	factoryHooksSetupOnce.Do(func() {
		started := time.Now()
		// No setup deadline; cancellation still kills the child's process group.
		cancelContext, cancel := context.WithCancel(context.Background())
		defer cancel()
		prepared := &factoryHooksPrepared{products: factoryHooksProducts(test)}
		for index := range prepared.javascript {
			prepared.javascript[index] = factoryHooksLowered(test, cancelContext, index)
		}
		prepared.native = factoryHooksNative(test, cancelContext, prepared.javascript[0])
		runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		prepared.runner = runner
		factoryHooksReady = prepared
		test.Logf("TestFactoryHooks_Setup: %.3fs", time.Since(started).Seconds())
	})
	if factoryHooksReady == nil {
		test.Fatal("factory hook setup did not complete")
	}
}

type factoryHooksProduct struct{ directory, manifest string }

func factoryHooksProducts(test *testing.T) [2]factoryHooksProduct {
	test.Helper()
	started := time.Now()
	value := shared("factory-hooks-fixtures", func(value *sharedValue) {
		for index := 0; index < 2; index++ {
			directory, manifest := factoryHooksPrepare(test, index == 1)
			value.rows = append(value.rows, directory, manifest)
		}
	})
	if len(value.rows) != 4 {
		test.Fatal("incomplete factory hook fixture setup")
	}
	products := [2]factoryHooksProduct{}
	for i := range products {
		products[i] = factoryHooksProduct{directory: value.rows[i*2], manifest: value.rows[i*2+1]}
	}
	test.Logf("TestFactoryHooks (setup): %.3fs", time.Since(started).Seconds())
	return products
}

func factoryHooksLowered(test *testing.T, callerContext context.Context, index int) string {
	test.Helper()
	value := shared(fmt.Sprintf("factory-hooks-lowered-%d", index), func(value *sharedValue) {
		p := factoryHooksProducts(test)[index]
		module, readError := os.ReadFile(filepath.Join(p.directory, "rules/no-debugger/rule.ts"))
		if readError != nil {
			test.Fatal(readError)
		}
		inputs := buildcache.Inputs{
			Name:      fmt.Sprintf("factory-hooks-lowered-%d", index),
			Files:     []string{"stage1/cohere/lint", "stage1/typescript", "internal", "cohere", "go.mod"},
			Flags:     []string{"rules=no-debugger", fmt.Sprintf("mutated=%t", index == 1), string(module)},
			Toolchain: []string{runtime.Version()},
		}
		product := buildcache.Product(test, inputs, func(destination string) error {
			if _, generateError := registry.Generate(p.directory); generateError != nil {
				return generateError
			}
			program, loadError := load.Load([]string{filepath.Join(p.directory, "main.ts")})
			if loadError != nil {
				return loadError
			}
			lowered, lowerError := lower.Lower(callerContext, program)
			if lowerError != nil {
				return lowerError
			}
			if writeError := os.WriteFile(
				filepath.Join(destination, "lint.mjs"),
				[]byte(javascript.JavaScript(lowered)),
				0644,
			); writeError != nil {
				return writeError
			}
			return os.WriteFile(filepath.Join(destination, "lint.c"), []byte(native.C(lowered)), 0644)
		})
		value.path = filepath.Join(product, "lint.mjs")
	})
	if value.path == "" {
		test.Fatal("incomplete lowered factory hooks product")
	}
	return value.path
}

func factoryHooksNative(test *testing.T, callerContext context.Context, javascript string) string {
	test.Helper()
	value := shared("factory-hooks-native", func(value *sharedValue) {
		source, readError := os.ReadFile(filepath.Join(filepath.Dir(javascript), "lint.c"))
		if readError != nil {
			test.Fatal(readError)
		}
		options := native.Options{Sanitize: true}
		inputs := buildcache.Inputs{
			Name:  "factory-hooks-native",
			Files: []string{"internal/native", "cohere", "internal/load"},
			Flags: append(
				native.Flags(options),
				fmt.Sprintf("source=%x", sha256.Sum256(source)),
				"ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"),
			),
			Toolchain: []string{runtime.Version(), buildcache.Tool("clang", "--version")},
		}
		binary := buildcache.Product(test, inputs, func(destination string) error {
			return factoryHooksBuildNative(
				callerContext,
				filepath.Join(filepath.Dir(javascript), "lint.c"),
				filepath.Join(destination, "scanner"),
			)
		})
		value.path = filepath.Join(binary, "scanner")
	})
	if value.path == "" {
		test.Fatal("incomplete native factory hooks product")
	}
	return value.path
}

// Test-only build entry: setup has no deadline and retains process-group cancellation.
func TestFactoryHooks_NativeBuildWorker(test *testing.T) {
	test.Parallel()
	sourcePath := os.Getenv("ADAMIC_FACTORY_HOOKS_NATIVE_SOURCE")
	if sourcePath == "" {
		test.Skip("invoked only by factory hook product builds")
	}
	source, readError := os.ReadFile(sourcePath)
	if readError != nil {
		test.Fatal(readError)
	}
	if buildError := native.Build(
		string(source),
		os.Getenv("ADAMIC_FACTORY_HOOKS_NATIVE_OUTPUT"),
		native.Options{Sanitize: true},
	); buildError != nil {
		test.Fatal(buildError)
	}
}

func factoryHooksBuildNative(callerContext context.Context, source, output string) error {
	executable, executableError := os.Executable()
	if executableError != nil {
		return executableError
	}
	command := lintCommand(
		callerContext,
		"",
		executable,
		"-test.run=^TestFactoryHooks_NativeBuildWorker$",
		"-test.timeout=0",
	)
	command.Env = append(
		os.Environ(),
		"ADAMIC_FACTORY_HOOKS_NATIVE_SOURCE="+source,
		"ADAMIC_FACTORY_HOOKS_NATIVE_OUTPUT="+output,
	)
	// Read through childguard.Capture for lintCommand's reason: a WaitDelay could fail a build that had finished.
	stdout, stderr, captureError := childguard.Capture(callerContext, command)
	if captureError != nil {
		return fmt.Errorf("sanitized native build: %w\n%s%s", captureError, stdout, stderr)
	}
	return nil
}

func factoryHooksExecute(test *testing.T, callerContext context.Context, name string, args ...string) execution {
	test.Helper()
	command := lintCommand(callerContext, "", name, args...)
	output, temporaryFileError := os.CreateTemp(test.TempDir(), "stdout-")
	if temporaryFileError != nil {
		test.Fatal(temporaryFileError)
	}
	defer output.Close()
	command.Stdout = output
	var stderr bytes.Buffer
	command.Stderr = &stderr
	started := time.Now()
	runError := command.Run()
	elapsed := time.Since(started)
	if callerContext.Err() != nil {
		test.Fatalf("cooked: shard exceeded 90s: %v", callerContext.Err())
	}
	if runError != nil || len(commandDiagnostics(name, stderr.Bytes())) != 0 {
		test.Fatalf("%s %v: %v\n%s", name, args, runError, &stderr)
	}
	data, readError := os.ReadFile(output.Name())
	if readError != nil {
		test.Fatal(readError)
	}
	return execution{output: data, duration: elapsed}
}

func factoryHooksVariant(mutated bool) int {
	if mutated {
		return 1
	}
	return 0
}

func factoryHooksShard(test *testing.T, index int) {
	test.Helper()
	factoryHooksSetup(test)
	prepared := factoryHooksReady
	// The shard deadline starts only after setup has published every immutable product.
	started := time.Now()
	timeoutContext, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	for _, c := range factoryHooksSlice(index) {
		p := prepared.products[0]
		if c.mutated {
			p = prepared.products[1]
		}
		var result execution
		switch c.backend {
		case "node":
			result = factoryHooksExecute(
				test,
				timeoutContext,
				"node",
				"--disable-warning=ExperimentalWarning",
				prepared.runner,
				filepath.Join(p.directory, "main.ts"),
				"--manifest",
				p.manifest,
			)
		case "javascript":
			result = factoryHooksExecute(
				test,
				timeoutContext,
				"node",
				"--disable-warning=ExperimentalWarning",
				prepared.runner,
				prepared.javascript[factoryHooksVariant(c.mutated)],
				"--manifest",
				p.manifest,
			)
		case "native":
			result = factoryHooksExecute(test, timeoutContext, prepared.native, "--manifest", p.manifest)
		default:
			test.Fatalf("unknown backend %s", c.backend)
		}
		if !factoryHooksMatches(result.output, c.mutated) {
			test.Fatalf("hook sequence on %s (mutant=%t): %s", c.name, c.mutated, result.output)
		}
		test.Logf("%s: %.3fs", c.name, result.duration.Seconds())
	}
	elapsed := time.Since(started)
	test.Logf("shard-%03d: %.3fs cooked=%t", index, elapsed.Seconds(), elapsed >= 60*time.Second)
	if elapsed >= 60*time.Second {
		test.Fatal("cooked: shard exceeds 60s budget")
	}
}

func TestFactoryHooks_000(test *testing.T) {
	test.Parallel()
	factoryHooksShard(test, 0)
}
func TestFactoryHooks_001(test *testing.T) {
	test.Parallel()
	factoryHooksShard(test, 1)
}
func TestFactoryHooks_002(test *testing.T) {
	test.Parallel()
	factoryHooksShard(test, 2)
}
func TestFactoryHooks_003(test *testing.T) {
	test.Parallel()
	factoryHooksShard(test, 3)
}
func TestFactoryHooks_004(test *testing.T) {
	test.Parallel()
	factoryHooksShard(test, 4)
}
