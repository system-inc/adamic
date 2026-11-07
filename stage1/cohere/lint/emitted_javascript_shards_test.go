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

const testEmittedJavaScriptMismatchShards = 1

var emittedMismatchProducts struct {
	sync.Once
	oracle, binary, module string
}

func emittedMismatchSetup(test *testing.T) {
	test.Helper()
	emittedMismatchProducts.Do(func() {
		// Shared preparation has no deadline; Loom covers the whole unit.
		// The subprocess receives immutable products; it never builds them again.
		if os.Getenv("ADAMIC_LINT_MISMATCH_PROBE") == "1" {
			emittedMismatchProducts.oracle = os.Getenv("ADAMIC_MISMATCH_ORACLE")
			emittedMismatchProducts.binary = os.Getenv("ADAMIC_MISMATCH_NATIVE")
			emittedMismatchProducts.module = os.Getenv("ADAMIC_MISMATCH_MODULE")
			for _, path := range []string{
				emittedMismatchProducts.oracle,
				emittedMismatchProducts.binary,
				emittedMismatchProducts.module,
			} {
				if _, statError := os.Stat(path); statError != nil {
					test.Fatal(statError)
				}
			}
			return
		}
		emittedMismatchProducts.oracle = lintGoOracleProduct(test)
		lowered := emittedMismatchLowered(test)
		emittedMismatchProducts.binary = emittedMismatchNative(test)
		emittedMismatchProducts.module = filepath.Join(lowered, "lint.mjs")
	})
	if emittedMismatchProducts.oracle == "" {
		test.Fatal("emitted mismatch setup did not complete")
	}
}

// The original test selects the first live no-var witness. Its stable repository
// path, rather than the temporary materialization path, is the assignment key.
func emittedMismatchCases(test *testing.T) []string {
	test.Helper()
	paths, witnessesError := registry.Witnesses(filepath.Join(packageDirectory, "rules/no-var"))
	if witnessesError != nil {
		test.Fatal(witnessesError)
	}
	if len(paths) == 0 {
		test.Fatal("no-var has no witnesses")
	}
	return paths[:1]
}

func emittedMismatchAssignment(key string) int {
	hash := fnv.New32a()
	relative, relativePathError := filepath.Rel(packageDirectory, key)
	if relativePathError != nil {
		panic(relativePathError)
	}
	hash.Write([]byte(filepath.ToSlash(relative)))
	return int(hash.Sum32() % testEmittedJavaScriptMismatchShards)
}

func emittedMismatchUnion(test *testing.T) {
	test.Helper()
	// Check the actual top-level declaration census, including accidental extras.
	tree, parseFileError := parser.ParseFile(token.NewFileSet(), "emitted_javascript_shards_test.go", nil, 0)
	if parseFileError != nil {
		test.Fatal(parseFileError)
	}
	count := 0
	for _, declaration := range tree.Decls {
		if function, ok := declaration.(*ast.FuncDecl); ok &&
			strings.HasPrefix(function.Name.Name, "TestEmittedJavaScriptMismatch_") {
			if function.Name.Name != fmt.Sprintf("TestEmittedJavaScriptMismatch_%03d", count) {
				test.Fatal("noncontiguous shard enumeration")
			}
			count++
		}
	}
	if count != testEmittedJavaScriptMismatchShards {
		test.Fatalf("shard declarations: got %d want %d", count, testEmittedJavaScriptMismatchShards)
	}
	cases := emittedMismatchCases(test)
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
			test.Fatalf("case %s assigned %d times", key, seen[key])
		}
	}
	if len(seen) != len(cases) {
		test.Fatal("shard union differs from live enumeration")
	}
	test.Logf("union: %d live cases, each exactly once", len(cases))
}

func TestEmittedJavaScriptMismatch_000(test *testing.T) {
	test.Parallel()
	// Prepare once per process, including when selected without _Setup.
	emittedMismatchSetup(test)
	started := time.Now()
	// A hard deadline applies even when this leaf is selected alone.
	deadline := time.AfterFunc(90*time.Second, func() { panic("P0: emitted JavaScript shard exceeded 90s") })
	defer deadline.Stop()
	emittedMismatchUnion(test)
	for _, key := range emittedMismatchCases(test) {
		if emittedMismatchAssignment(key) != 0 {
			continue
		}
		source := ownedWitnesses(test, packageDirectory, "no-var")[0]
		path := manifest(test, []string{source + "\tno-var"})
		module := emittedMismatchProducts.module
		if os.Getenv("ADAMIC_LINT_MISMATCH_PROBE") == "1" {
			data, readError := os.ReadFile(module)
			if readError != nil {
				test.Fatal(readError)
			}
			module = filepath.Join(test.TempDir(), "lint.mjs")
			data = append(data, []byte("\nconsole.log('planted emitted JavaScript mismatch');\n")...)
			if writeError := os.WriteFile(module, data, 0644); writeError != nil {
				test.Fatal(writeError)
			}
			compareWithJavaScript(
				test,
				emittedMismatchProducts.oracle,
				emittedMismatchProducts.binary,
				packageDirectory,
				path,
				module,
			)
			test.Fatal("emitted JavaScript mutant survived")
		}
		compareWithJavaScript(
			test,
			emittedMismatchProducts.oracle,
			emittedMismatchProducts.binary,
			packageDirectory,
			path,
			module,
		)
		timeoutContext, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()
		command := exec.CommandContext(
			timeoutContext,
			os.Args[0],
			"-test.run=^TestEmittedJavaScriptMismatch_000$",
			"-test.timeout=90s",
			"-test.v",
		)
		command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		command.Cancel = func() error {
			killError := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
			if killError == syscall.ESRCH {
				return os.ErrProcessDone
			}
			return killError
		}
		command.Env = append(
			os.Environ(),
			"ADAMIC_LINT_MISMATCH_PROBE=1",
			"ADAMIC_MISMATCH_ORACLE="+emittedMismatchProducts.oracle,
			"ADAMIC_MISMATCH_NATIVE="+emittedMismatchProducts.binary,
			"ADAMIC_MISMATCH_MODULE="+module,
		)
		log, createError := os.Create(filepath.Join(test.TempDir(), "mismatch.log"))
		if createError != nil {
			test.Fatal(createError)
		}
		command.Stdout, command.Stderr = log, log
		runError := command.Run()
		contextError := timeoutContext.Err()
		cancel()
		if contextError != nil {
			test.Fatalf("P0: mutant child cooked at 90s: %v", contextError)
		}
		if closeError := log.Close(); closeError != nil {
			test.Fatal(closeError)
		}
		data, readError := os.ReadFile(log.Name())
		if readError != nil {
			test.Fatal(readError)
		}
		if runError == nil ||
			!bytes.Contains(data, []byte("emitted JavaScript:")) ||
			!bytes.Contains(data, []byte("planted emitted JavaScript mismatch")) ||
			bytes.Count(data, []byte("--- FAIL: TestEmittedJavaScriptMismatch_")) != 1 {
			test.Fatalf("wrong mutant failure: %v\n%s", runError, data)
		}
		test.Logf("one planted case caught by shard 000: %s", data)
	}
	elapsed := time.Since(started)
	test.Logf("shard 000: %.3fs cooked=%t", elapsed.Seconds(), elapsed >= 90*time.Second)
	if elapsed >= 60*time.Second {
		test.Fatal(fmt.Sprintf("shard exceeds 60s budget: %s", elapsed))
	}
}

func emittedMismatchInputs(test *testing.T) []string {
	test.Helper()
	files := []string{
		"stage1/typescript",
		"stage1/cohere/lint/registry",
		"internal",
		"go.mod",
		"cohere/TypeScript/tsc/internal",
		"cohere/TypeScript-shim",
		"cohere/TypeScript/tsc/go.mod",
		"cohere/TypeScript/tsc/go.sum",
		"cohere/static_single_assignment",
		"cohere/mutation_aliasing",
	}
	for _, path := range portFiles(test) {
		files = append(files, filepath.ToSlash(filepath.Join("stage1/cohere/lint", path)))
	}

	// Checker baselines are not build inputs. Hash production sources and embeds,
	// skipping corpus directories before visiting their tens of thousands of files.
	var production []string
	// Reads: cohere
	// Reads: stage1/*.a
	// Reads: stage1/*.ts
	// Reads: stage1/typescript/parser/grammar.ts
	// Reads: stage1/typescript/parser/jsx.ts
	// Reads: stage1/typescript/parser/lookahead.ts
	// Reads: stage1/typescript/parser/nodes.ts
	// Reads: stage1/typescript/parser/parser.ts
	// Reads: stage1/typescript/parser/statements.ts
	// Reads: stage1/typescript/scanner/characters.ts
	// Reads: stage1/typescript/scanner/scanner.ts
	// Reads: stage1/typescript/scanner/tokens.ts
	// Reads: internal/load/prelude.d.ts
	// Reads: internal/native/runtime/*.[ch]
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	for _, input := range files {
		walkError := filepath.WalkDir(filepath.Join(
			root,
			input,
		), func(path string, entry fs.DirEntry, pathError error) error {
			if pathError != nil {
				return pathError
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
			relative, relativePathError := filepath.Rel(root, path)
			if relativePathError != nil {
				return relativePathError
			}
			production = append(production, filepath.ToSlash(relative))
			return nil
		})
		if walkError != nil {
			test.Fatal(walkError)
		}
	}
	files = production
	return files
}

func emittedMismatchLowered(test *testing.T) string {
	test.Helper()
	value := shared("emitted-mismatch-lowered", func(value *sharedValue) {
		files := emittedMismatchInputs(test)
		lowered := buildcache.Product(test, buildcache.Inputs{
			Name:      "emitted-mismatch-lowered",
			Files:     files,
			Toolchain: []string{runtime.Version(), runtime.GOOS, runtime.GOARCH},
		}, func(out string) error {
			if _, generateError := registry.Generate(packageDirectory); generateError != nil {
				return generateError
			}
			program, loadError := load.Load([]string{filepath.Join(packageDirectory, "main.ts")})
			if loadError != nil {
				return loadError
			}
			lowered, lowerError := lower.Lower(context.Background(), program)
			if lowerError != nil {
				return lowerError
			}
			if writeError := os.WriteFile(
				filepath.Join(out, "lint.mjs"),
				[]byte(javascript.JavaScript(lowered)),
				0644,
			); writeError != nil {
				return writeError
			}
			return os.WriteFile(filepath.Join(out, "lint.c"), []byte(native.C(lowered)), 0644)
		})
		value.path = lowered
	})
	if value.path == "" {
		test.Fatal("emitted mismatch lowered preparation failed")
	}
	return value.path
}

func emittedMismatchNative(test *testing.T) string {
	test.Helper()
	value := shared("emitted-mismatch-native", func(value *sharedValue) {
		files := emittedMismatchInputs(test)
		lowered := emittedMismatchLowered(test)
		options := native.Options{Sanitize: true, Split: true, Jobs: 4}
		flags := append(
			native.Flags(options),
			"Split=true",
			"Jobs=4",
			"ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"),
		)
		product := buildcache.Product(test, buildcache.Inputs{
			Name:      "emitted-mismatch-native",
			Files:     files,
			Flags:     flags,
			Toolchain: []string{runtime.Version(), runtime.GOOS, runtime.GOARCH, buildcache.Tool("clang", "--version")},
		}, func(out string) error {
			source, readError := os.ReadFile(filepath.Join(lowered, "lint.c"))
			if readError != nil {
				return readError
			}
			return native.Build(string(source), filepath.Join(out, "scanner"), options)
		})
		value.path = filepath.Join(product, "scanner")
	})
	if value.path == "" {
		test.Fatal("emitted mismatch native preparation failed")
	}
	return value.path
}
