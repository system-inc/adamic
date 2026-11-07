package lint

import (
	"bytes"
	"context"

	"fmt"
	"github.com/system-inc/adamic/stage1/cohere/lint/oracles"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// The pinned case is compared once on each of the original three runtimes.
const testSuggestionAlongsideAutomaticFixShards = 3

var suggestionAlongsideCases = []string{"Node", "emitted JavaScript", "native"}
var suggestionAlongside struct {
	ready                                       bool
	directory, manifest, oracle, binary, module string
}

var suggestionAlongsideOnce sync.Once

func suggestionAlongsideSetup(test *testing.T) {
	suggestionAlongsideOnce.Do(func() { suggestionAlongsidePrepare(test) })
}

func suggestionAlongsidePrepare(test *testing.T) {
	test.Helper()
	if suggestionAlongside.ready {
		return
	}
	directory, temporaryDirectoryError := os.MkdirTemp(sharedDirectory, "suggestion-alongside-")
	if temporaryDirectoryError != nil {
		test.Fatal(temporaryDirectoryError)
	}
	copyPort(test, directory, "", "")
	// This pinned case selects only no-debugger. Exclude unrelated rules from
	// the generated dispatch so setup does not lower the entire live corpus.
	rules, readDirectoryError := os.ReadDir(filepath.Join(directory, "rules"))
	if readDirectoryError != nil {
		test.Fatal(readDirectoryError)
	}
	for _, rule := range rules {
		if rule.IsDir() && rule.Name() != "no-debugger" {
			if removeError := os.RemoveAll(filepath.Join(directory, "rules", rule.Name())); removeError != nil {
				test.Fatal(removeError)
			}
		}
	}
	for _, name := range []string{"rule.a", "oracle.go"} {
		data, readError := os.ReadFile(filepath.Join("testdata/serialization", name))
		if readError != nil {
			test.Fatal(readError)
		}
		from, to := "debuggerMessage, '', '', ''", "debuggerMessage, 'fix', ';', ''"
		if name == "oracle.go" {
			from, to = "d.Fixes = nil", "d.Fixes[0].Text = \";\""
		}
		if bytes.Count(data, []byte(from)) != 1 {
			test.Fatal("automatic fix anchor changed")
		}
		if writeError := os.WriteFile(
			filepath.Join(directory, "rules/no-debugger", name),
			bytes.Replace(data, []byte(from), []byte(to), 1),
			0644,
		); writeError != nil {
			test.Fatal(writeError)
		}
	}
	if removeError := os.Remove(filepath.Join(directory, "rules/no-debugger/rule.ts")); removeError != nil {
		test.Fatal(removeError)
	}
	prepareRegistry(test, directory)
	source := filepath.Join(directory, "mixed.ts")
	if writeError := os.WriteFile(source, []byte("/*😀*/debugger;\n"), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	// Keep the manifest alive after the setup test returns.
	path := filepath.Join(directory, "manifest.txt")
	if writeError := os.WriteFile(path, []byte(source+"\tno-debugger\n"), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	inputs := buildcache.Inputs{Name: "suggestion-alongside-lowered-v2", Files: []string{
		"stage1/cohere/lint",
		"internal",
		"cohere/TypeScript/tsc",
		"cohere/TypeScript-shim",
		"cohere/go.mod",
		"cohere/go.sum",
		"go.mod",
		"go.work",
		"cohere/go.work",
		"cohere/go.work.sum",
	}, Flags: []string{"automatic-fix=semicolon", "serialization-fixture=v1"}, Toolchain: []string{runtime.Version()}}
	lowered := buildcache.Product(test, inputs, func(out string) error {
		program, loadError := load.Load([]string{filepath.Join(directory, "main.ts")})
		if loadError != nil {
			return loadError
		}
		ir, lowerError := lower.Lower(context.Background(), program)
		if lowerError != nil {
			return lowerError
		}
		if writeError := os.WriteFile(filepath.Join(out, "lint.c"), []byte(native.C(ir)), 0644); writeError != nil {
			return writeError
		}
		return os.WriteFile(filepath.Join(out, "lint.mjs"), []byte(javascript.JavaScript(ir)), 0644)
	})
	inputs.Name = "suggestion-alongside-native-v2"
	inputs.Flags = append(inputs.Flags, "split=true", "jobs=4")
	inputs.Flags = append(inputs.Flags, native.Flags(native.Options{Sanitize: true, Split: true, Jobs: 4})...)
	inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("clang", "--version"))
	built := buildcache.Product(test, inputs, func(out string) error {
		source, readError := os.ReadFile(filepath.Join(lowered, "lint.c"))
		if readError != nil {
			return readError
		}
		return native.Build(
			string(source),
			filepath.Join(out, "scanner"),
			native.Options{Sanitize: true, Split: true, Jobs: 4},
		)
	})
	oracle := oracles.SuggestionAlongsideOracle(test, directory)

	suggestionAlongside.directory, suggestionAlongside.manifest, suggestionAlongside.oracle = directory, path, oracle
	suggestionAlongside.binary, suggestionAlongside.module = filepath.Join(
		built,
		"scanner",
	), filepath.Join(lowered, "lint.mjs")
	suggestionAlongside.ready = true
}

// Not parallel: prepares the shared fixture and cached runtime products before parallel leaves.
func TestSuggestionAlongsideAutomaticFix_Setup(test *testing.T) {
	suggestionAlongsideSetup(test)
}

// The setup test is serial: testing releases parallel leaves only after it returns.
// A filtered leaf prepares its shared products before starting its work deadline.
func suggestionAlongsideReady(test *testing.T) {
	test.Helper()
	suggestionAlongsideSetup(test)
}

func suggestionAlongsideDeadline(test *testing.T) func() {
	test.Helper()
	started := time.Now()
	timer := time.AfterFunc(90*time.Second, func() {
		fmt.Fprintf(os.Stderr, "P0: %s cooked at 90s\n", test.Name())
		os.Exit(124)
	})
	return func() {
		timer.Stop()
		elapsed := time.Since(started)
		test.Logf("%s: %.3fs cooked=false", test.Name(), elapsed.Seconds())
		if elapsed >= 60*time.Second {
			test.Error("exceeds 60s budget")
		}
	}
}

func suggestionAlongsideCheck(got, want []byte) string { return difference(got, want) }
func suggestionAlongsideShard(test *testing.T, index int) {
	test.Helper()
	suggestionAlongsideReady(test)
	defer suggestionAlongsideDeadline(test)()
	s := &suggestionAlongside
	want := execute(test, "", s.oracle, "--manifest", s.manifest).output
	if !bytes.Contains(want, []byte("fixed\t/*\\ud83d\\ude00*/;\\u000a")) {
		test.Fatalf("automatic fix lost: %s", want)
	}
	var got []byte
	switch index {
	case 0:
		got = node(test, s.directory, s.manifest, false).output
	case 1:
		got = runJavaScript(test, s.module, s.manifest, false).output
	case 2:
		got = execute(test, "", s.binary, "--manifest", s.manifest).output
	default:
		test.Fatalf("unknown shard %d", index)
	}
	if os.Getenv("ADAMIC_SUGGESTION_ALONGSIDE_PLANT") == suggestionAlongsideCases[index] {
		got = append(got, []byte("planted automatic fix mismatch\n")...)
	}
	if diff := suggestionAlongsideCheck(got, want); diff != "" {
		test.Fatalf("%s: %s", suggestionAlongsideCases[index], diff)
	}
	test.Logf("%s identical to Go: %d bytes", suggestionAlongsideCases[index], len(want))
}
func TestSuggestionAlongsideAutomaticFix_000(test *testing.T) {
	test.Parallel()
	suggestionAlongsideShard(test, 0)
}
func TestSuggestionAlongsideAutomaticFix_001(test *testing.T) {
	test.Parallel()
	suggestionAlongsideShard(test, 1)
}
func TestSuggestionAlongsideAutomaticFix_002(test *testing.T) {
	test.Parallel()
	suggestionAlongsideShard(test, 2)
}

func suggestionAlongsideUnion(test *testing.T) {
	test.Helper()
	file, parseFileError := parser.ParseFile(token.NewFileSet(), "suggestion_alongside_shards_test.go", nil, 0)
	if parseFileError != nil {
		test.Fatal(parseFileError)
	}
	seen := map[string]int{}
	assignments := make([]int, len(suggestionAlongsideCases))
	for _, declaration := range file.Decls {
		if decl, ok := declaration.(*ast.FuncDecl); ok &&
			strings.HasPrefix(decl.Name.Name, "TestSuggestionAlongsideAutomaticFix_") &&
			decl.Name.Name != "TestSuggestionAlongsideAutomaticFix_Setup" {
			seen[decl.Name.Name]++
			parallel, calls := 0, 0
			ast.Inspect(decl.Body, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				if selector, ok := call.Fun.(*ast.SelectorExpr); ok && selector.Sel.Name == "Parallel" {
					parallel++
				}
				if name, ok := call.Fun.(*ast.Ident); ok && name.Name == "suggestionAlongsideShard" {
					calls++
					if len(call.Args) != 2 {
						test.Fatal("wrong shard arguments")
					}
					literal, ok := call.Args[1].(*ast.BasicLit)
					if !ok {
						test.Fatal("shard index must be literal")
					}
					index, parseError := strconv.Atoi(literal.Value)
					if parseError != nil || index < 0 || index >= len(assignments) {
						test.Fatal("invalid shard index")
					}
					assignments[index]++
				}
				return true
			})
			if parallel != 1 || calls != 1 {
				test.Fatal("shard must be parallel and run exactly one case")
			}
		}
	}
	if len(seen) != testSuggestionAlongsideAutomaticFixShards ||
		len(suggestionAlongsideCases) != testSuggestionAlongsideAutomaticFixShards {
		test.Fatal("shard enumeration changed")
	}
	for index := range suggestionAlongsideCases {
		if assignments[index] != 1 {
			test.Fatal("case missing or duplicated", index)
		}
		if seen[fmt.Sprintf("TestSuggestionAlongsideAutomaticFix_%03d", index)] != 1 {
			test.Fatal("missing or duplicated shard", index)
		}
	}
	test.Logf(
		"union: %d runtime comparisons, one pinned source/rule case, each exactly once",
		len(suggestionAlongsideCases),
	)
}

func TestSuggestionAlongsideAutomaticFixPlantedFailure(test *testing.T) {
	test.Parallel()
	suggestionAlongsideReady(test)
	timeoutContext, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()
	command := exec.CommandContext(
		timeoutContext,
		os.Args[0],
		"-test.run=^TestSuggestionAlongsideAutomaticFix_(Setup|[0-9]+)$",
		"-test.timeout=600s",
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
	command.Env = append(os.Environ(), "ADAMIC_SUGGESTION_ALONGSIDE_PLANT=native")
	// Read through childguard.Capture for lintBuildPhaseChild's reason, stdout then stderr; the 600 s deadline ends a
	// hung child.
	stdout, stderr, captureError := childguard.Capture(timeoutContext, command)
	output := append(stdout, stderr...)
	if captureError == nil ||
		bytes.Count(output, []byte("--- FAIL: TestSuggestionAlongsideAutomaticFix_")) != 1 ||
		!bytes.Contains(output, []byte("--- FAIL: TestSuggestionAlongsideAutomaticFix_002")) ||
		!bytes.Contains(output, []byte("planted automatic fix mismatch")) {
		test.Fatalf("wrong planted failure: %v\n%s", captureError, output)
	}
	test.Log("planted mismatch caught only by shard 002")
}

// A filtered leaf looks up its own inputs in a fresh process, inheriting this run's products (a runner's are
// Workshop's, read and never built; TestProduct_ proves each builds alone).
func TestSuggestionAlongsideAutomaticFixSetupIsRequired(test *testing.T) {
	test.Parallel()
	command, ctx := lintBuildPhaseChild(
		test,
		os.Args[0],
		"-test.run=^TestSuggestionAlongsideAutomaticFix_002$",
		"-test.timeout=600s",
		"-test.v",
	)
	command.Env = append(os.Environ(), "ADAMIC_SUGGESTION_ALONGSIDE_PLANT=")
	stdout, stderr, captureError := childguard.Capture(ctx, command)
	output := append(stdout, stderr...)
	if captureError != nil ||
		!bytes.Contains(output, []byte("--- PASS: TestSuggestionAlongsideAutomaticFix_002")) ||
		!productLookedUp(output, `suggestion-alongside-\S+`) {
		test.Fatalf("standalone shard: %v\n%s", captureError, output)
	}
}
