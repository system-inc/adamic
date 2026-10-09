package lint

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
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
	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
)

// Hash the live witness keys so new witnesses join the union automatically.
const testWitnessScriptKindShards = 2

var witnessScriptKindState struct {
	sync.Once
	directory, oracle, binary, module string
	keys, sources                     []string
}

func witnessScriptKindShard(key string) int {
	h := sha256.Sum256([]byte(key))
	return int(binary.BigEndian.Uint64(h[:8]) % testWitnessScriptKindShards)
}

// Products depend on code, not the renamed raw-text witness. The copied port
// changes only that witness; its rules and generated registry are identical.
func witnessScriptKindSetup(t *testing.T) {
	t.Helper()
	s := &witnessScriptKindState
	s.Do(func() {
		directory, err := os.MkdirTemp(sharedDirectory, "witness-kind-")
		if err != nil {
			t.Fatal(err)
		}
		s.directory = copyPort(t, directory, "", "")
		// Build the same full Go oracle while lowering/clang use the other CPUs.
		type oracleResult struct {
			path string
			err  error
		}
		oracleDone := make(chan oracleResult, 1)
		go func() {
			out, err := os.MkdirTemp(sharedDirectory, "witness-oracle-")
			if err != nil {
				oracleDone <- oracleResult{err: err}
				return
			}
			path, err := goOracleIn(packageDirectory, out)
			// The shared runner treats module-download notices as errors. Retry
			// after that successful fetch, without changing the shared helper.
			if err != nil && strings.Contains(err.Error(), "go: downloading ") {
				path, err = goOracleIn(packageDirectory, out)
			}
			oracleDone <- oracleResult{path, err}
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
			s.keys = append(s.keys, filepath.ToSlash(key))
			s.sources = append(s.sources, source)
		}
		// Include the checker and every lowering/runtime input. Test edits may cause
		// conservative misses, but cannot reuse a stale native or lowered product.
		inputs := buildcache.Inputs{
			Name:      "witness-script-kind-lowered-no-debugger-v1",
			Files:     []string{"internal", "stage1", "cohere/internal", "cohere/TypeScript/tsc", "cohere/TypeScript-shim", "cohere/static_single_assignment", "cohere/mutation_aliasing", "cohere/go.mod", "cohere/go.sum", "cohere/rule_runner", "cohere/policy", "cohere/schema", "go.mod", "go.work"},
			Toolchain: []string{runtime.Version()},
		}
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
			result, err := lower.Lower(context.Background(), program)
			if err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(out, "main.c"), []byte(native.C(result)), 0644); err != nil {
				return err
			}
			return os.WriteFile(filepath.Join(out, "lint.mjs"), []byte(javascript.JavaScript(result)), 0644)
		})
		s.module = filepath.Join(lowered, "lint.mjs")
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
		s.binary = filepath.Join(product, "scanner")
		// GoBuild is not on this base; retain the existing Go build path.
		oracle := <-oracleDone
		if oracle.err != nil {
			t.Fatal(oracle.err)
		}
		s.oracle = oracle.path
	})
	if s.oracle == "" || s.binary == "" || len(s.sources) == 0 {
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
		if f, ok := declaration.(*ast.FuncDecl); ok && strings.HasPrefix(f.Name.Name, "TestWitnessScriptKind_") {
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
	live, err := registry.Witnesses(filepath.Join(s.directory, "rules/no-debugger"))
	if err != nil {
		t.Fatal(err)
	}
	counts := map[string]int{}
	for shard := 0; shard < testWitnessScriptKindShards; shard++ {
		for _, key := range s.keys {
			if witnessScriptKindShard(key) == shard {
				counts[key]++
			}
		}
	}
	if len(counts) != len(live) {
		t.Fatalf("union %d != live %d", len(counts), len(live))
	}
	for _, path := range live {
		key, err := filepath.Rel(s.directory, path)
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
	witnessScriptKindSetup(t)
	s := &witnessScriptKindState
	for i, key := range s.keys {
		if witnessScriptKindShard(key) != shard {
			continue
		}
		path := manifest(t, []string{s.sources[i] + "\tno-debugger"})
		if os.Getenv("ADAMIC_WITNESS_KIND_PLANT") == key {
			// Only this case's emitted module disagrees; all other cases stay ordinary.
			data, err := os.ReadFile(s.module)
			if err != nil {
				t.Fatal(err)
			}
			module := filepath.Join(t.TempDir(), "planted.mjs")
			data = append(data, []byte("\nconsole.log('planted witness script kind mismatch');\n")...)
			if err := os.WriteFile(module, data, 0644); err != nil {
				t.Fatal(err)
			}
			compareWithJavaScript(t, s.oracle, s.binary, s.directory, path, module)
			t.Fatal("planted mismatch survived")
		}
		compareWithJavaScript(t, s.oracle, s.binary, s.directory, path, s.module)
	}
}

func TestWitnessScriptKindPlantedFailure(t *testing.T) {
	witnessScriptKindSetup(t)
	key := witnessScriptKindState.keys[0]
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestWitnessScriptKind_[0-9]+$", "-test.timeout=90s", "-test.v")
	command.Env = append(os.Environ(), "ADAMIC_WITNESS_KIND_PLANT="+key)
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
