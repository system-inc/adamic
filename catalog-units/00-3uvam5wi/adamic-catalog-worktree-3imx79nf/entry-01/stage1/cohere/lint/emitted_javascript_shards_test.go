package lint

import (
	"bytes"
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"hash/fnv"
	"io/fs"
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

const testEmittedJavaScriptMismatchShards = 1

var emittedMismatchProducts struct {
	sync.Once
	oracle, binary, module string
}

func emittedMismatchSetup(t *testing.T) {
	t.Helper()
	emittedMismatchProducts.Do(func() {
		// The subprocess receives immutable products; it never builds them again.
		if os.Getenv("ADAMIC_LINT_MISMATCH_PROBE") == "1" {
			emittedMismatchProducts.oracle = os.Getenv("ADAMIC_MISMATCH_ORACLE")
			emittedMismatchProducts.binary = os.Getenv("ADAMIC_MISMATCH_NATIVE")
			emittedMismatchProducts.module = os.Getenv("ADAMIC_MISMATCH_MODULE")
			for _, path := range []string{emittedMismatchProducts.oracle, emittedMismatchProducts.binary, emittedMismatchProducts.module} {
				if _, err := os.Stat(path); err != nil {
					t.Fatal(err)
				}
			}
			return
		}
		oracleDirectory, err := os.MkdirTemp(sharedDirectory, "emitted-mismatch-oracle-")
		if err != nil {
			t.Fatal(err)
		}
		type oracleResult struct {
			path string
			err  error
		}
		oracleDone := make(chan oracleResult, 1)
		go func() {
			path, err := goOracleIn(packageDirectory, oracleDirectory)
			oracleDone <- oracleResult{path, err}
		}()
		files := []string{"stage1/typescript", "stage1/cohere/lint/registry", "internal", "go.mod", "cohere/TypeScript/tsc/internal", "cohere/TypeScript-shim", "cohere/TypeScript/tsc/go.mod", "cohere/TypeScript/tsc/go.sum", "cohere/static_single_assignment", "cohere/mutation_aliasing"}
		for _, path := range portFiles(t) {
			files = append(files, filepath.ToSlash(filepath.Join("stage1/cohere/lint", path)))
		}

		// Checker baselines are not build inputs. Hash production sources and embeds,
		// skipping corpus directories before visiting their tens of thousands of files.
		var production []string
		root, err := filepath.Abs(repository)
		if err != nil {
			t.Fatal(err)
		}
		for _, input := range files {
			err := filepath.WalkDir(filepath.Join(root, input), func(path string, entry fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if entry.IsDir() {
					if entry.Name() == "testdata" || entry.Name() == "performance" {
						return filepath.SkipDir
					}
					return nil
				}
				if strings.HasSuffix(path, "_test.go") {
					return nil
				}
				relative, err := filepath.Rel(root, path)
				if err != nil {
					return err
				}
				production = append(production, filepath.ToSlash(relative))
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
		}
		files = production
		lowered := buildcache.Product(t, buildcache.Inputs{Name: "emitted-mismatch-lowered", Files: files, Toolchain: []string{runtime.Version(), runtime.GOOS, runtime.GOARCH}}, func(out string) error {
			if _, err := registry.Generate(packageDirectory); err != nil {
				return err
			}
			program, err := load.Load([]string{filepath.Join(packageDirectory, "main.ts")})
			if err != nil {
				return err
			}
			lowered, err := lower.Lower(context.Background(), program)
			if err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(out, "lint.mjs"), []byte(javascript.JavaScript(lowered)), 0644); err != nil {
				return err
			}
			return os.WriteFile(filepath.Join(out, "lint.c"), []byte(native.C(lowered)), 0644)
		})
		options := native.Options{Sanitize: true, Split: true, Jobs: 4}
		flags := append(native.Flags(options), "Split=true", "Jobs=4", "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"))
		product := buildcache.Product(t, buildcache.Inputs{Name: "emitted-mismatch-native", Files: files, Flags: flags, Toolchain: []string{runtime.Version(), runtime.GOOS, runtime.GOARCH, buildcache.Tool("clang", "--version")}}, func(out string) error {
			source, err := os.ReadFile(filepath.Join(lowered, "lint.c"))
			if err != nil {
				return err
			}
			return native.Build(string(source), filepath.Join(out, "scanner"), options)
		})
		emittedMismatchProducts.binary = filepath.Join(product, "scanner")
		emittedMismatchProducts.module = filepath.Join(lowered, "lint.mjs")
		// main has no buildcache.GoBuild yet: preserve the existing Go overlay build.
		oracle := <-oracleDone
		if oracle.err != nil {
			t.Fatal(oracle.err)
		}
		emittedMismatchProducts.oracle = oracle.path
	})
	if emittedMismatchProducts.oracle == "" {
		t.Fatal("emitted mismatch setup did not complete")
	}
}

// The original test selects the first live no-var witness. Its stable repository
// path, rather than the temporary materialization path, is the assignment key.
func emittedMismatchCases(t *testing.T) []string {
	t.Helper()
	paths, err := registry.Witnesses(filepath.Join(packageDirectory, "rules/no-var"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatal("no-var has no witnesses")
	}
	return paths[:1]
}

func emittedMismatchAssignment(key string) int {
	hash := fnv.New32a()
	relative, err := filepath.Rel(packageDirectory, key)
	if err != nil {
		panic(err)
	}
	hash.Write([]byte(filepath.ToSlash(relative)))
	return int(hash.Sum32() % testEmittedJavaScriptMismatchShards)
}

func emittedMismatchUnion(t *testing.T) {
	t.Helper()
	// Check the actual top-level declaration census, including accidental extras.
	tree, err := parser.ParseFile(token.NewFileSet(), "emitted_javascript_shards_test.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, declaration := range tree.Decls {
		if function, ok := declaration.(*ast.FuncDecl); ok && strings.HasPrefix(function.Name.Name, "TestEmittedJavaScriptMismatch_") {
			if function.Name.Name != fmt.Sprintf("TestEmittedJavaScriptMismatch_%03d", count) {
				t.Fatal("noncontiguous shard enumeration")
			}
			count++
		}
	}
	if count != testEmittedJavaScriptMismatchShards {
		t.Fatalf("shard declarations: got %d want %d", count, testEmittedJavaScriptMismatchShards)
	}
	cases := emittedMismatchCases(t)
	seen := make(map[string]int)
	for shard := 0; shard < testEmittedJavaScriptMismatchShards; shard++ {
		for _, key := range cases {
			if emittedMismatchAssignment(key) == shard {
				seen[key]++
			}
		}
	}
	for _, key := range cases {
		if seen[key] != 1 {
			t.Fatalf("case %s assigned %d times", key, seen[key])
		}
	}
	if len(seen) != len(cases) {
		t.Fatal("shard union differs from live enumeration")
	}
	t.Logf("union: %d live cases, each exactly once", len(cases))
}

func TestEmittedJavaScriptMismatch_000(t *testing.T) {
	t.Parallel()
	started := time.Now()
	// A hard deadline applies even when this leaf is selected alone.
	deadline := time.AfterFunc(90*time.Second, func() { panic("P0: emitted JavaScript shard exceeded 90s") })
	defer deadline.Stop()
	emittedMismatchSetup(t)
	emittedMismatchUnion(t)
	for _, key := range emittedMismatchCases(t) {
		if emittedMismatchAssignment(key) != 0 {
			continue
		}
		source := ownedWitnesses(t, packageDirectory, "no-var")[0]
		path := manifest(t, []string{source + "\tno-var"})
		module := emittedMismatchProducts.module
		if os.Getenv("ADAMIC_LINT_MISMATCH_PROBE") == "1" {
			data, err := os.ReadFile(module)
			if err != nil {
				t.Fatal(err)
			}
			module = filepath.Join(t.TempDir(), "lint.mjs")
			data = append(data, []byte("\nconsole.log('planted emitted JavaScript mismatch');\n")...)
			if err := os.WriteFile(module, data, 0644); err != nil {
				t.Fatal(err)
			}
			compareWithJavaScript(t, emittedMismatchProducts.oracle, emittedMismatchProducts.binary, packageDirectory, path, module)
			t.Fatal("emitted JavaScript mutant survived")
		}
		compareWithJavaScript(t, emittedMismatchProducts.oracle, emittedMismatchProducts.binary, packageDirectory, path, module)
		command := exec.Command(os.Args[0], "-test.run=^TestEmittedJavaScriptMismatch_000$", "-test.timeout=90s", "-test.v")
		command.Env = append(os.Environ(), "ADAMIC_LINT_MISMATCH_PROBE=1", "ADAMIC_MISMATCH_ORACLE="+emittedMismatchProducts.oracle, "ADAMIC_MISMATCH_NATIVE="+emittedMismatchProducts.binary, "ADAMIC_MISMATCH_MODULE="+module)
		log, err := os.Create(filepath.Join(t.TempDir(), "mismatch.log"))
		if err != nil {
			t.Fatal(err)
		}
		command.Stdout, command.Stderr = log, log
		runError := command.Run()
		if err := log.Close(); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(log.Name())
		if err != nil {
			t.Fatal(err)
		}
		if runError == nil || !bytes.Contains(data, []byte("emitted JavaScript:")) || !bytes.Contains(data, []byte("planted emitted JavaScript mismatch")) || bytes.Count(data, []byte("--- FAIL: TestEmittedJavaScriptMismatch_")) != 1 {
			t.Fatalf("wrong mutant failure: %v\n%s", runError, data)
		}
		t.Logf("one planted case caught by shard 000: %s", data)
	}
	elapsed := time.Since(started)
	t.Logf("shard 000: %.3fs cooked=%t", elapsed.Seconds(), elapsed >= 90*time.Second)
	if elapsed >= 60*time.Second {
		t.Fatal(fmt.Sprintf("shard exceeds 60s budget: %s", elapsed))
	}
}
