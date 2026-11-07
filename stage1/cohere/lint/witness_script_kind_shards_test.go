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
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
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
var witnessScriptKindOnce sync.Once

func witnessScriptKindShard(key string) int {
	h := sha256.Sum256([]byte(key))
	return int(binary.BigEndian.Uint64(h[:8]) % testWitnessScriptKindShards)
}

// Products depend on code, not the renamed raw-text witness. The copied port
// changes only that witness; its rules and generated registry are identical.
func witnessScriptKindSetup(test *testing.T) {
	test.Helper()
	witnessScriptKindOnce.Do(func() {
		products := &witnessScriptKindState
		directory, temporaryDirectoryError := os.MkdirTemp(sharedDirectory, "witness-kind-")
		if temporaryDirectoryError != nil {
			test.Fatal(temporaryDirectoryError)
		}
		products.Directory = copyPort(test, directory, "", "")
		witness := filepath.Join(directory, "rules/no-debugger/testdata/witness.ts.txt")
		renamed := strings.TrimSuffix(witness, ".ts.txt") + ".tsx.txt"
		if renameError := os.Rename(witness, renamed); renameError != nil {
			test.Fatal(renameError)
		}
		if writeError := os.WriteFile(renamed, []byte("const node = 1; debugger;\n"), 0644); writeError != nil {
			test.Fatal(writeError)
		}
		paths, witnessesError := registry.Witnesses(filepath.Join(directory, "rules/no-debugger"))
		if witnessesError != nil {
			test.Fatal(witnessesError)
		}
		for index, path := range paths {
			key, relativePathError := filepath.Rel(directory, path)
			if relativePathError != nil {
				test.Fatal(relativePathError)
			}
			data, readError := os.ReadFile(path)
			if readError != nil {
				test.Fatal(readError)
			}
			source := filepath.Join(
				directory,
				fmt.Sprintf("source-%d%s", index, filepath.Ext(strings.TrimSuffix(path, ".txt"))),
			)
			if writeError := os.WriteFile(source, data, 0644); writeError != nil {
				test.Fatal(writeError)
			}
			if path == renamed && filepath.Ext(source) != ".tsx" {
				test.Fatal("witness script kind lost")
			}
			products.Keys = append(products.Keys, filepath.ToSlash(key))
			products.Sources = append(products.Sources, source)
		}
		lowered := witnessScriptKindLoweredProduct(test)
		products.Module = filepath.Join(lowered, "lint.mjs")
		products.Binary = filepath.Join(witnessScriptKindNativeProduct(test), "scanner")
		products.Oracle = lintGoOracleProduct(test)
		products.ready = true
	})
	products := &witnessScriptKindState
	if !products.ready ||
		products.Oracle == "" ||
		products.Binary == "" ||
		products.Module == "" ||
		len(products.Sources) == 0 ||
		len(products.Keys) != len(products.Sources) {
		test.Fatal("witness setup incomplete")
	}
}

func witnessScriptKindUnion(test *testing.T) {
	test.Helper()
	parsed, parseFileError := parser.ParseFile(token.NewFileSet(), "witness_script_kind_shards_test.go", nil, 0)
	if parseFileError != nil {
		test.Fatal(parseFileError)
	}
	found := map[string]bool{}
	for _, declaration := range parsed.Decls {
		if decl, ok := declaration.(*ast.FuncDecl); ok &&
			strings.HasPrefix(decl.Name.Name, "TestWitnessScriptKind_") &&
			decl.Name.Name != "TestWitnessScriptKind_Setup" {
			found[decl.Name.Name] = true
		}
	}
	if len(found) != testWitnessScriptKindShards {
		test.Fatalf("shards: %d, declared %d", len(found), testWitnessScriptKindShards)
	}
	for index := 0; index < testWitnessScriptKindShards; index++ {
		if !found[fmt.Sprintf("TestWitnessScriptKind_%03d", index)] {
			test.Fatalf("missing shard %d", index)
		}
	}
	products := &witnessScriptKindState
	live, witnessesError := registry.Witnesses(filepath.Join(products.Directory, "rules/no-debugger"))
	if witnessesError != nil {
		test.Fatal(witnessesError)
	}
	counts := map[string]int{}
	for shard := 0; shard < testWitnessScriptKindShards; shard++ {
		for _, key := range products.Keys {
			if witnessScriptKindShard(key) == shard {
				counts[key]++
			}
		}
	}
	if len(counts) != len(live) {
		test.Fatalf("union %d != live %d", len(counts), len(live))
	}
	for _, path := range live {
		key, relativePathError := filepath.Rel(products.Directory, path)
		if relativePathError != nil {
			test.Fatal(relativePathError)
		}
		if counts[filepath.ToSlash(key)] != 1 {
			test.Fatalf("case %s not covered exactly once", key)
		}
	}
	test.Logf("union: %d live cases, exactly once", len(live))
}

func TestWitnessScriptKind_000(test *testing.T) {
	test.Parallel()
	witnessScriptKindRun(test, 0)
}

func TestWitnessScriptKind_001(test *testing.T) {
	test.Parallel()
	witnessScriptKindRun(test, 1)
}

func witnessScriptKindRun(test *testing.T, shard int) {
	test.Helper()
	witnessScriptKindSetup(test)
	stop := witnessScriptKindDeadline(test)
	defer stop()
	products := &witnessScriptKindState
	for index, key := range products.Keys {
		if witnessScriptKindShard(key) != shard {
			continue
		}
		path := manifest(test, []string{products.Sources[index] + "\tno-debugger"})
		if os.Getenv("ADAMIC_WITNESS_KIND_PLANT") == key {
			// Only this case's emitted module disagrees; all other cases stay ordinary.
			data, readError := os.ReadFile(products.Module)
			if readError != nil {
				test.Fatal(readError)
			}
			module := filepath.Join(test.TempDir(), "planted.mjs")
			data = append(data, []byte("\nconsole.log('planted witness script kind mismatch');\n")...)
			if writeError := os.WriteFile(module, data, 0644); writeError != nil {
				test.Fatal(writeError)
			}
			compareWithJavaScript(test, products.Oracle, products.Binary, products.Directory, path, module)
			test.Fatal("planted mismatch survived")
		}
		compareWithJavaScript(test, products.Oracle, products.Binary, products.Directory, path, products.Module)
	}
}

func TestWitnessScriptKindPlantedFailure(test *testing.T) {
	test.Parallel()
	witnessScriptKindSetup(test)
	key := witnessScriptKindState.Keys[0]
	timeoutContext, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()
	command := witnessScriptKindCommand(
		timeoutContext,
		os.Args[0],
		"-test.run=^TestWitnessScriptKind_[0-9]+$",
		"-test.timeout=600s",
		"-test.v",
	)
	command.Env = append(os.Environ(), "ADAMIC_WITNESS_KIND_PLANT="+key)
	stdout, stderr, captureError := childguard.Capture(timeoutContext, command)
	output := append(stdout, stderr...)
	if timeoutContext.Err() != nil {
		test.Fatalf("P0 cooked planted-failure shard: %v", timeoutContext.Err())
	}
	caught := fmt.Sprintf("--- FAIL: TestWitnessScriptKind_%03d", witnessScriptKindShard(key))
	if captureError == nil ||
		bytes.Count(output, []byte("--- FAIL: TestWitnessScriptKind_")) != 1 ||
		!bytes.Contains(output, []byte(caught)) ||
		!bytes.Contains(output, []byte("emitted JavaScript:")) ||
		!bytes.Contains(output, []byte("planted witness script kind mismatch")) {
		test.Fatalf("wrong planted failure: %v\n%s", captureError, output)
	}
	test.Logf("case %s caught by exactly one shard: %s\n%s", key, caught, output)
}

func witnessScriptKindInputs() buildcache.Inputs {
	return buildcache.Inputs{
		Name: "witness-script-kind-lowered-no-debugger-v1",
		Files: []string{
			"internal",
			"stage1",
			"cohere/internal",
			"cohere/TypeScript/tsc",
			"cohere/TypeScript-shim",
			"cohere/static_single_assignment",
			"cohere/mutation_aliasing",
			"cohere/go.mod",
			"cohere/go.sum",
			"cohere/rule_runner",
			"cohere/policy",
			"cohere/schema",
			"go.mod",
			"go.work",
		},
		Toolchain: []string{runtime.Version()},
	}
}

// Start only after setup is ready: a leaf's budget belongs to its own cases.
func witnessScriptKindDeadline(test *testing.T) func() {
	test.Helper()
	timer := time.AfterFunc(90*time.Second, func() { panic("P0 cooked: " + test.Name() + " exceeded 90s") })
	return func() { timer.Stop() }
}

// Optional entry point; every leaf prepares these same products independently.
func TestWitnessScriptKind_Setup(test *testing.T) {
	test.Parallel()
	witnessScriptKindSetup(test)
}

// POSIX process groups keep compiler descendants within the context cancellation. Its callers read the child's output
// through childguard.Capture, to its end with no timer deciding when: exec's own copying under a two-second WaitDelay
// reported exec.ErrWaitDelay whenever the reads hadn't reached EOF two seconds after the child exited, so a leaf that
// had passed could fail its probe under load (#n3gk22h's shape). Each caller's 600 s deadline ends a hung child.
func witnessScriptKindCommand(callerContext context.Context, name string, args ...string) *exec.Cmd {
	command := exec.CommandContext(callerContext, name, args...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		killError := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if killError == syscall.ESRCH {
			return os.ErrProcessDone
		}
		return killError
	}
	return command
}

func TestWitnessScriptKindStandalone(test *testing.T) {
	test.Parallel()
	witnessScriptKindSetup(test)
	timeoutContext, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()
	// A new process selects only one leaf and fetches its own shared products.
	command := witnessScriptKindCommand(
		timeoutContext,
		os.Args[0],
		"-test.run=^TestWitnessScriptKind_000$",
		"-test.timeout=600s",
		"-test.v",
	)
	stdout, stderr, captureError := childguard.Capture(timeoutContext, command)
	output := append(stdout, stderr...)
	if timeoutContext.Err() != nil {
		test.Fatalf("P0 cooked standalone probe: %v", timeoutContext.Err())
	}
	if captureError != nil || !bytes.Contains(output, []byte("--- PASS: TestWitnessScriptKind_000")) {
		test.Fatalf("standalone leaf failed: %v\n%s", captureError, output)
	}
}
