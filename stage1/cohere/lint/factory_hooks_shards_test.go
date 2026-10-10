package lint

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
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

func TestFactoryHooksUnion(t *testing.T) {
	t.Parallel()
	registered := []func(*testing.T){TestFactoryHooks_000, TestFactoryHooks_001, TestFactoryHooks_002, TestFactoryHooks_003, TestFactoryHooks_004}
	if len(registered) != testFactoryHooksShards {
		t.Fatalf("registered %d shards, want %d", len(registered), testFactoryHooksShards)
	}
	seen := map[string]int{}
	for i := 0; i < testFactoryHooksShards; i++ {
		for _, c := range factoryHooksSlice(i) {
			seen[c.name]++
		}
	}
	for _, c := range factoryHooksCases() {
		if seen[c.name] != 1 {
			t.Fatalf("%s covered %d times", c.name, seen[c.name])
		}
	}
	if len(seen) != len(factoryHooksCases()) {
		t.Fatal("unexpected case")
	}
	t.Logf("union: %d cases, each exactly once across %d shards", len(seen), testFactoryHooksShards)
}

func TestFactoryHooksPlantedFailure(t *testing.T) {
	t.Parallel()
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
		t.Fatalf("planted disagreement caught by %v", caught)
	}
	t.Log("planted native hook-order disagreement caught only by shard 002")
}

func factoryHooksDirectory(t *testing.T) string {
	t.Helper()
	directory, err := os.MkdirTemp(sharedDirectory, "factory-hooks-")
	if err != nil {
		t.Fatal(err)
	}
	return directory
}

func factoryHooksManifest(t *testing.T, directory string, rows []string) string {
	t.Helper()
	path := filepath.Join(directory, "manifest")
	if err := os.WriteFile(path, []byte(strings.Join(rows, "\n")+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func factoryHooksPrepare(t *testing.T, mutated bool) (string, string) {
	directory := copyPort(t, factoryHooksDirectory(t), "", "")
	// This test selects only no-debugger. Keep the same linter and rule source,
	// but omit unrelated rule factories from this private fixture's module graph.
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
	module := filepath.Join(directory, "rules/no-debugger/rule.ts")
	data, err := os.ReadFile(module)
	if err != nil {
		t.Fatal(err)
	}
	source := strings.Replace(string(data), "    visit(index: number, parent: number): void {", `    ready = false;
    prepare(root: number): void { this.ready = true; console.log('prepare'); }
    finish(root: number): void { console.log('finish'); }
    visit(index: number, parent: number): void {
        if(!this.ready) { panic('visit before prepare'); }
        console.log('visit');`, 1)
	source = strings.Replace(source, "return new Rule(context);", `if(context.node(context.parents.length - 1).kind !== 'SourceFile') { panic('factory before ancestry'); }
    if(context.settings.read('number', '') !== '-2' || !context.settings.read('payload', '').includes('enabled')) { panic('structured option lost'); }
    console.log('factory');
    return new Rule(context);`, 1)
	source = "import { panic } from 'adamic';\n" + source
	if err := os.WriteFile(module, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	descriptor := filepath.Join(directory, "rules/no-debugger/rule.json")
	data, err = os.ReadFile(descriptor)
	if err != nil {
		t.Fatal(err)
	}
	var options map[string]any
	if err := json.Unmarshal(data, &options); err != nil {
		t.Fatal(err)
	}
	options["prepare"] = "prepare"
	options["finish"] = "finish"
	write := func() {
		data, err := json.Marshal(options)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(descriptor, data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	if mutated {
		delete(options, "finish")
	}
	write()
	fixture := filepath.Join(directory, "hooks.ts")
	if err := os.WriteFile(fixture, []byte("debugger;\n"), 0644); err != nil {
		t.Fatal(err)
	}
	path := factoryHooksManifest(t, directory, []string{fixture + "\tno-debugger\t\t\tfalse\t{\"Number\":-2,\"Payload\":{\"enabled\":true}}"})

	prepareRegistry(t, directory)
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

func factoryHooksSetup(t *testing.T) {
	t.Helper()
	factoryHooksSetupOnce.Do(func() {
		started := time.Now()
		// No setup deadline; cancellation still kills the child's process group.
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		prepared := &factoryHooksPrepared{products: factoryHooksProducts(t)}
		for index := range prepared.javascript {
			prepared.javascript[index] = factoryHooksLowered(t, ctx, index)
		}
		prepared.native = factoryHooksNative(t, ctx, prepared.javascript[0])
		runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
		if err != nil {
			t.Fatal(err)
		}
		prepared.runner = runner
		factoryHooksReady = prepared
		t.Logf("TestFactoryHooks_Setup: %.3fs", time.Since(started).Seconds())
	})
	if factoryHooksReady == nil {
		t.Fatal("factory hook setup did not complete")
	}
}

type factoryHooksProduct struct{ directory, manifest string }

func factoryHooksProducts(t *testing.T) [2]factoryHooksProduct {
	t.Helper()
	started := time.Now()
	value := shared("factory-hooks-fixtures", func(value *sharedValue) {
		for index := 0; index < 2; index++ {
			directory, manifest := factoryHooksPrepare(t, index == 1)
			value.rows = append(value.rows, directory, manifest)
		}
	})
	if len(value.rows) != 4 {
		t.Fatal("incomplete factory hook fixture setup")
	}
	products := [2]factoryHooksProduct{}
	for i := range products {
		products[i] = factoryHooksProduct{directory: value.rows[i*2], manifest: value.rows[i*2+1]}
	}
	t.Logf("TestFactoryHooks (setup): %.3fs", time.Since(started).Seconds())
	return products
}

func factoryHooksLowered(t *testing.T, ctx context.Context, index int) string {
	t.Helper()
	value := shared(fmt.Sprintf("factory-hooks-lowered-%d", index), func(value *sharedValue) {
		p := factoryHooksProducts(t)[index]
		module, err := os.ReadFile(filepath.Join(p.directory, "rules/no-debugger/rule.ts"))
		if err != nil {
			t.Fatal(err)
		}
		inputs := buildcache.Inputs{
			Name:      fmt.Sprintf("factory-hooks-lowered-%d", index),
			Files:     []string{"stage1/cohere/lint", "stage1/typescript", "internal", "cohere", "go.mod"},
			Flags:     []string{"rules=no-debugger", fmt.Sprintf("mutated=%t", index == 1), string(module)},
			Toolchain: []string{runtime.Version()},
		}
		product := buildcache.Product(t, inputs, func(destination string) error {
			if _, err := registry.Generate(p.directory); err != nil {
				return err
			}
			program, err := load.Load([]string{filepath.Join(p.directory, "main.ts")})
			if err != nil {
				return err
			}
			lowered, err := lower.Lower(ctx, program)
			if err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(destination, "lint.mjs"), []byte(javascript.JavaScript(lowered)), 0644); err != nil {
				return err
			}
			return os.WriteFile(filepath.Join(destination, "lint.c"), []byte(native.C(lowered)), 0644)
		})
		value.path = filepath.Join(product, "lint.mjs")
	})
	if value.path == "" {
		t.Fatal("incomplete lowered factory hooks product")
	}
	return value.path
}

func factoryHooksNative(t *testing.T, ctx context.Context, javascript string) string {
	t.Helper()
	value := shared("factory-hooks-native", func(value *sharedValue) {
		source, err := os.ReadFile(filepath.Join(filepath.Dir(javascript), "lint.c"))
		if err != nil {
			t.Fatal(err)
		}
		options := native.Options{Sanitize: true}
		inputs := buildcache.Inputs{Name: "factory-hooks-native", Files: []string{"internal/native", "cohere", "internal/load"},
			Flags:     append(native.Flags(options), fmt.Sprintf("source=%x", sha256.Sum256(source)), "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT")),
			Toolchain: []string{runtime.Version(), buildcache.Tool("clang", "--version")}}
		binary := buildcache.Product(t, inputs, func(destination string) error {
			return factoryHooksBuildNative(ctx, filepath.Join(filepath.Dir(javascript), "lint.c"), filepath.Join(destination, "scanner"))
		})
		value.path = filepath.Join(binary, "scanner")
	})
	if value.path == "" {
		t.Fatal("incomplete native factory hooks product")
	}
	return value.path
}

// Test-only build entry: setup has no deadline and retains process-group cancellation.
func TestFactoryHooks_NativeBuildWorker(t *testing.T) {
	t.Parallel()
	sourcePath := os.Getenv("ADAMIC_FACTORY_HOOKS_NATIVE_SOURCE")
	if sourcePath == "" {
		t.Skip("invoked only by factory hook product builds")
	}
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := native.Build(string(source), os.Getenv("ADAMIC_FACTORY_HOOKS_NATIVE_OUTPUT"), native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
}

func factoryHooksCommand(ctx context.Context, name string, args ...string) *exec.Cmd {
	command := exec.CommandContext(ctx, name, args...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	command.WaitDelay = time.Second
	return command
}

func factoryHooksBuildNative(ctx context.Context, source, output string) error {
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	command := factoryHooksCommand(ctx, executable, "-test.run=^TestFactoryHooks_NativeBuildWorker$", "-test.timeout=0")
	command.Env = append(os.Environ(), "ADAMIC_FACTORY_HOOKS_NATIVE_SOURCE="+source, "ADAMIC_FACTORY_HOOKS_NATIVE_OUTPUT="+output)
	data, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("sanitized native build: %w\n%s", err, data)
	}
	return nil
}

func factoryHooksExecute(t *testing.T, ctx context.Context, name string, args ...string) execution {
	t.Helper()
	command := factoryHooksCommand(ctx, name, args...)
	output, err := os.CreateTemp(t.TempDir(), "stdout-")
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	command.Stdout = output
	var stderr bytes.Buffer
	command.Stderr = &stderr
	started := time.Now()
	err = command.Run()
	elapsed := time.Since(started)
	if ctx.Err() != nil {
		t.Fatalf("cooked: shard exceeded 90s: %v", ctx.Err())
	}
	if err != nil || len(commandDiagnostics(name, stderr.Bytes())) != 0 {
		t.Fatalf("%s %v: %v\n%s", name, args, err, &stderr)
	}
	data, err := os.ReadFile(output.Name())
	if err != nil {
		t.Fatal(err)
	}
	return execution{output: data, duration: elapsed}
}

func factoryHooksVariant(mutated bool) int {
	if mutated {
		return 1
	}
	return 0
}

func factoryHooksShard(t *testing.T, index int) {
	t.Helper()
	factoryHooksSetup(t)
	prepared := factoryHooksReady
	// The shard deadline starts only after setup has published every immutable product.
	started := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	for _, c := range factoryHooksSlice(index) {
		p := prepared.products[0]
		if c.mutated {
			p = prepared.products[1]
		}
		var result execution
		switch c.backend {
		case "node":
			result = factoryHooksExecute(t, ctx, "node", "--disable-warning=ExperimentalWarning", prepared.runner, filepath.Join(p.directory, "main.ts"), "--manifest", p.manifest)
		case "javascript":
			result = factoryHooksExecute(t, ctx, "node", "--disable-warning=ExperimentalWarning", prepared.runner, prepared.javascript[factoryHooksVariant(c.mutated)], "--manifest", p.manifest)
		case "native":
			result = factoryHooksExecute(t, ctx, prepared.native, "--manifest", p.manifest)
		default:
			t.Fatalf("unknown backend %s", c.backend)
		}
		if !factoryHooksMatches(result.output, c.mutated) {
			t.Fatalf("hook sequence on %s (mutant=%t): %s", c.name, c.mutated, result.output)
		}
		t.Logf("%s: %.3fs", c.name, result.duration.Seconds())
	}
	elapsed := time.Since(started)
	t.Logf("shard-%03d: %.3fs cooked=%t", index, elapsed.Seconds(), elapsed >= 60*time.Second)
	if elapsed >= 60*time.Second {
		t.Fatal("cooked: shard exceeds 60s budget")
	}
}

func TestFactoryHooks_000(t *testing.T) {
	t.Parallel()
	factoryHooksShard(t, 0)
}
func TestFactoryHooks_001(t *testing.T) {
	t.Parallel()
	factoryHooksShard(t, 1)
}
func TestFactoryHooks_002(t *testing.T) {
	t.Parallel()
	factoryHooksShard(t, 2)
}
func TestFactoryHooks_003(t *testing.T) {
	t.Parallel()
	factoryHooksShard(t, 3)
}
func TestFactoryHooks_004(t *testing.T) {
	t.Parallel()
	factoryHooksShard(t, 4)
}
