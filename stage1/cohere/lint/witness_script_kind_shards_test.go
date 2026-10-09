package lint

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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

// Hash the live witness keys so new witnesses join the union automatically.
const testWitnessScriptKindShards = 2

type witnessScriptKindProducts struct {
	Directory, Oracle, Binary, Module string
	Keys, Sources                     []string
	ready                             bool
}

var witnessScriptKindState witnessScriptKindProducts

func witnessScriptKindShard(key string) int {
	h := sha256.Sum256([]byte(key))
	return int(binary.BigEndian.Uint64(h[:8]) % testWitnessScriptKindShards)
}

// Products depend on code, not the renamed raw-text witness. The copied port
// changes only that witness; its rules and generated registry are identical.
func witnessScriptKindSetup(t *testing.T, ctx context.Context) {
	t.Helper()
	witnessScriptKindState = witnessScriptKindProducts{}
	s := &witnessScriptKindState
	directory, err := os.MkdirTemp(sharedDirectory, "witness-kind-")
	if err != nil {
		t.Fatal(err)
	}
	s.Directory = copyPort(t, directory, "", "")
	// Build the same full Go oracle while lowering/clang use the other CPUs.
	type oracleResult struct {
		path string
		err  error
	}
	oracleDone := make(chan oracleResult, 1)
	go func() {
		inputs := witnessScriptKindInputs()
		inputs.Name = "witness-script-kind-go-oracle-v1"
		inputs.Flags = []string{"go build", "overlay=full-rule-registry", "GOTOOLCHAIN=" + os.Getenv("GOTOOLCHAIN"), "GOFLAGS=" + os.Getenv("GOFLAGS"), "CGO_ENABLED=" + os.Getenv("CGO_ENABLED")}
		inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("go", "version"))
		product, err := buildcache.Get(inputs, func(out string) error {
			_, err := witnessScriptKindGoOracleIn(ctx, packageDirectory, out)
			return err
		})
		oracleDone <- oracleResult{filepath.Join(product, "oracle"), err}
	}()
	witness := filepath.Join(directory, "rules/no-debugger/testdata/witness.ts.txt")
	renamed := strings.TrimSuffix(witness, ".ts.txt") + ".tsx.txt"
	if err := os.Rename(witness, renamed); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(renamed, []byte("const node = 1; debugger;\n"), 0644); err != nil {
		t.Fatal(err)
	}
	paths, err := registry.Witnesses(filepath.Join(directory, "rules/no-debugger"))
	if err != nil {
		t.Fatal(err)
	}
	for i, path := range paths {
		key, err := filepath.Rel(directory, path)
		if err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		source := filepath.Join(directory, fmt.Sprintf("source-%d%s", i, filepath.Ext(strings.TrimSuffix(path, ".txt"))))
		if err := os.WriteFile(source, data, 0644); err != nil {
			t.Fatal(err)
		}
		if path == renamed && filepath.Ext(source) != ".tsx" {
			t.Fatal("witness script kind lost")
		}
		s.Keys = append(s.Keys, filepath.ToSlash(key))
		s.Sources = append(s.Sources, source)
	}
	// Include the checker and every lowering/runtime input. Test edits may cause
	// conservative misses, but cannot reuse a stale native or lowered product.
	inputs := witnessScriptKindInputs()
	lowered := buildcache.Product(t, inputs, func(out string) error {
		// The manifest selects only no-debugger. Compile that unchanged rule
		// with the unchanged scanner and serializers, avoiding every unrelated
		// rule's lowering. Node and Go still use the complete registry.
		source := copyPort(t, filepath.Join(out, "source"), "", "")
		var selected []registry.Descriptor
		for _, descriptor := range prepareRegistry(t, ".") {
			if descriptor.Slug == "no-debugger" {
				selected = append(selected, descriptor)
			}
		}
		if len(selected) != 1 {
			return fmt.Errorf("no-debugger descriptor count: %d", len(selected))
		}
		ts, _ := registry.Render(selected)
		if err := os.WriteFile(filepath.Join(source, ".generated/registry.ts"), ts, 0644); err != nil {
			return err
		}
		program, err := load.Load([]string{filepath.Join(source, "main.ts")})
		if err != nil {
			return err
		}
		result, err := lower.Lower(ctx, program)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(out, "main.c"), []byte(native.C(result)), 0644); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(out, "lint.mjs"), []byte(javascript.JavaScript(result)), 0644)
	})
	s.Module = filepath.Join(lowered, "lint.mjs")
	options := native.Options{Sanitize: true, Split: true}
	inputs.Name = "witness-script-kind-native-no-debugger-v1"
	inputs.Flags = append(native.Flags(options), "Split=true", "ADAMIC_NATIVE_JOBS="+os.Getenv("ADAMIC_NATIVE_JOBS"), "ADAMIC_GATE_UNCACHED="+os.Getenv("ADAMIC_GATE_UNCACHED"))
	inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("clang", "--version"))
	product := buildcache.Product(t, inputs, func(out string) error {
		data, err := os.ReadFile(filepath.Join(lowered, "main.c"))
		if err != nil {
			return err
		}
		return native.Build(string(data), filepath.Join(out, "scanner"), options)
	})
	s.Binary = filepath.Join(product, "scanner")
	// The full Go oracle is fetched once by this explicit setup.
	oracle := <-oracleDone
	if oracle.err != nil {
		t.Fatal(oracle.err)
	}
	s.Oracle = oracle.path
	s.ready = true
	if s.Oracle == "" || s.Binary == "" || len(s.Sources) == 0 {
		t.Fatal("witness setup incomplete")
	}
}

func witnessScriptKindUnion(t *testing.T) {
	t.Helper()
	parsed, err := parser.ParseFile(token.NewFileSet(), "witness_script_kind_shards_test.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	found := map[string]bool{}
	for _, declaration := range parsed.Decls {
		if f, ok := declaration.(*ast.FuncDecl); ok && strings.HasPrefix(f.Name.Name, "TestWitnessScriptKind_") && f.Name.Name != "TestWitnessScriptKind_Setup" {
			found[f.Name.Name] = true
		}
	}
	if len(found) != testWitnessScriptKindShards {
		t.Fatalf("shards: %d, declared %d", len(found), testWitnessScriptKindShards)
	}
	for i := 0; i < testWitnessScriptKindShards; i++ {
		if !found[fmt.Sprintf("TestWitnessScriptKind_%03d", i)] {
			t.Fatalf("missing shard %d", i)
		}
	}
	s := &witnessScriptKindState
	live, err := registry.Witnesses(filepath.Join(s.Directory, "rules/no-debugger"))
	if err != nil {
		t.Fatal(err)
	}
	counts := map[string]int{}
	for shard := 0; shard < testWitnessScriptKindShards; shard++ {
		for _, key := range s.Keys {
			if witnessScriptKindShard(key) == shard {
				counts[key]++
			}
		}
	}
	if len(counts) != len(live) {
		t.Fatalf("union %d != live %d", len(counts), len(live))
	}
	for _, path := range live {
		key, err := filepath.Rel(s.Directory, path)
		if err != nil {
			t.Fatal(err)
		}
		if counts[filepath.ToSlash(key)] != 1 {
			t.Fatalf("case %s not covered exactly once", key)
		}
	}
	t.Logf("union: %d live cases, exactly once", len(live))
}

func TestWitnessScriptKind_000(t *testing.T) {
	t.Parallel()
	witnessScriptKindRun(t, 0)
}

func TestWitnessScriptKind_001(t *testing.T) {
	t.Parallel()
	witnessScriptKindRun(t, 1)
}

func witnessScriptKindRun(t *testing.T, shard int) {
	t.Helper()
	witnessScriptKindRequireReady(t)
	stop := witnessScriptKindDeadline(t)
	defer stop()
	s := &witnessScriptKindState
	for i, key := range s.Keys {
		if witnessScriptKindShard(key) != shard {
			continue
		}
		path := manifest(t, []string{s.Sources[i] + "\tno-debugger"})
		if os.Getenv("ADAMIC_WITNESS_KIND_PLANT") == key {
			// Only this case's emitted module disagrees; all other cases stay ordinary.
			data, err := os.ReadFile(s.Module)
			if err != nil {
				t.Fatal(err)
			}
			module := filepath.Join(t.TempDir(), "planted.mjs")
			data = append(data, []byte("\nconsole.log('planted witness script kind mismatch');\n")...)
			if err := os.WriteFile(module, data, 0644); err != nil {
				t.Fatal(err)
			}
			compareWithJavaScript(t, s.Oracle, s.Binary, s.Directory, path, module)
			t.Fatal("planted mismatch survived")
		}
		compareWithJavaScript(t, s.Oracle, s.Binary, s.Directory, path, s.Module)
	}
}

func TestWitnessScriptKindPlantedFailure(t *testing.T) {
	t.Parallel()
	witnessScriptKindRequireReady(t)
	key := witnessScriptKindState.Keys[0]
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	command := witnessScriptKindCommand(ctx, os.Args[0], "-test.run=^TestWitnessScriptKind_(Setup|[0-9]+)$", "-test.timeout=90s", "-test.v")
	snapshot, err := json.Marshal(witnessScriptKindState)
	if err != nil {
		t.Fatal(err)
	}
	command.Env = append(os.Environ(), "ADAMIC_WITNESS_KIND_PLANT="+key, "ADAMIC_WITNESS_KIND_READY="+string(snapshot))
	output, err := command.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("P0 cooked planted-failure shard: %v", ctx.Err())
	}
	caught := fmt.Sprintf("--- FAIL: TestWitnessScriptKind_%03d", witnessScriptKindShard(key))
	if err == nil || bytes.Count(output, []byte("--- FAIL: TestWitnessScriptKind_")) != 1 || !bytes.Contains(output, []byte(caught)) || !bytes.Contains(output, []byte("emitted JavaScript:")) || !bytes.Contains(output, []byte("planted witness script kind mismatch")) {
		t.Fatalf("wrong planted failure: %v\n%s", err, output)
	}
	t.Logf("case %s caught by exactly one shard: %s\n%s", key, caught, output)
}

func witnessScriptKindInputs() buildcache.Inputs {
	return buildcache.Inputs{
		Name:      "witness-script-kind-lowered-no-debugger-v1",
		Files:     []string{"internal", "stage1", "cohere/internal", "cohere/TypeScript/tsc", "cohere/TypeScript-shim", "cohere/static_single_assignment", "cohere/mutation_aliasing", "cohere/go.mod", "cohere/go.sum", "cohere/rule_runner", "cohere/policy", "cohere/schema", "go.mod", "go.work"},
		Toolchain: []string{runtime.Version()},
	}
}

func witnessScriptKindRequireReady(t *testing.T) {
	t.Helper()
	s := &witnessScriptKindState
	if !s.ready || s.Oracle == "" || s.Binary == "" || s.Module == "" || len(s.Sources) == 0 || len(s.Keys) != len(s.Sources) {
		t.Fatal("shared witness setup is not ready; select TestWitnessScriptKind_Setup together with the leaves")
	}
}

// Start only after setup is ready: a leaf's budget belongs to its own cases.
func witnessScriptKindDeadline(t *testing.T) func() {
	t.Helper()
	timer := time.AfterFunc(90*time.Second, func() { panic("P0 cooked: " + t.Name() + " exceeded 90s") })
	return func() { timer.Stop() }
}

// Not parallel: publishes shared witness products before the parallel leaves run.
func TestWitnessScriptKind_Setup(t *testing.T) {
	stop := witnessScriptKindDeadline(t)
	defer stop()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	if snapshot := os.Getenv("ADAMIC_WITNESS_KIND_READY"); snapshot != "" {
		// The planted-failure child receives its parent's already-built products.
		// No child shard prepares state or rebuilds an oracle.
		if err := json.Unmarshal([]byte(snapshot), &witnessScriptKindState); err != nil {
			t.Fatal(err)
		}
		witnessScriptKindState.ready = true
	} else {
		witnessScriptKindSetup(t, ctx)
	}
	witnessScriptKindRequireReady(t)
	witnessScriptKindUnion(t)
}

// POSIX process groups keep compiler descendants within the context cancellation.
func witnessScriptKindCommand(ctx context.Context, name string, args ...string) *exec.Cmd {
	command := exec.CommandContext(ctx, name, args...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if err == syscall.ESRCH {
			return os.ErrProcessDone
		}
		return err
	}
	command.WaitDelay = 2 * time.Second
	return command
}

func witnessScriptKindGoOracleIn(ctx context.Context, sourceRoot, directory string) (string, error) {
	root, err := filepath.Abs(filepath.Join(repository, "cohere"))
	if err != nil {
		return "", err
	}
	side, err := filepath.Abs(filepath.Join(packageDirectory, "testdata/oracle.go"))
	if err != nil {
		return "", err
	}
	descriptors, err := registry.Generate(sourceRoot)
	if err != nil {
		return "", err
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
		return "", failure
	}
	overlay, err := json.Marshal(map[string]any{"Replace": replacements})
	if err != nil {
		return "", err
	}
	path := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(path, overlay, 0644); err != nil {
		return "", err
	}
	binary := filepath.Join(directory, "oracle")
	args := append([]string{"build", "-overlay=" + path, "-o", binary}, virtualFiles...)
	command := witnessScriptKindCommand(ctx, "go", args...)
	command.Dir = root
	command.Env = append(os.Environ(), "PWD="+root)
	output, err := command.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("Go oracle build: %w\n%s", err, output)
	}
	return binary, nil
}

func TestWitnessScriptKindRequiresSetup(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	// Selecting a leaf alone must reject missing setup, never build it lazily.
	command := witnessScriptKindCommand(ctx, os.Args[0], "-test.run=^TestWitnessScriptKind_000$", "-test.timeout=90s", "-test.v")
	output, err := command.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("P0 cooked missing-setup probe: %v", ctx.Err())
	}
	if err == nil || !bytes.Contains(output, []byte("shared witness setup is not ready")) || bytes.Contains(output, []byte("build witness-script-kind-")) {
		t.Fatalf("leaf did not reject missing setup before building: %v\n%s", err, output)
	}
}
