package lint

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
)

// The pinned case is compared once on each of the original three runtimes.
const testSuggestionAlongsideAutomaticFixShards = 3

var suggestionAlongsideCases = []string{"Node", "emitted JavaScript", "native"}
var suggestionAlongside struct {
	sync.Mutex
	ready                                       bool
	directory, manifest, oracle, binary, module string
}

func suggestionAlongsideSetup(t *testing.T) {
	t.Helper()
	suggestionAlongside.Lock()
	defer suggestionAlongside.Unlock()
	if suggestionAlongside.ready {
		return
	}
	directory, err := os.MkdirTemp(sharedDirectory, "suggestion-alongside-")
	if err != nil {
		t.Fatal(err)
	}
	copyPort(t, directory, "", "")
	// This pinned case selects only no-debugger. Exclude unrelated rules from
	// the generated dispatch so setup does not lower the entire live corpus.
	rules, err := os.ReadDir(filepath.Join(directory, "rules"))
	if err != nil {
		t.Fatal(err)
	}
	for _, rule := range rules {
		if rule.IsDir() && rule.Name() != "no-debugger" {
			if err := os.RemoveAll(filepath.Join(directory, "rules", rule.Name())); err != nil {
				t.Fatal(err)
			}
		}
	}
	for _, name := range []string{"rule.a", "oracle.go"} {
		data, err := os.ReadFile(filepath.Join("testdata/serialization", name))
		if err != nil {
			t.Fatal(err)
		}
		from, to := "debuggerMessage, '', '', ''", "debuggerMessage, 'fix', ';', ''"
		if name == "oracle.go" {
			from, to = "d.Fixes = nil", "d.Fixes[0].Text = \";\""
		}
		if bytes.Count(data, []byte(from)) != 1 {
			t.Fatal("automatic fix anchor changed")
		}
		if err := os.WriteFile(filepath.Join(directory, "rules/no-debugger", name), bytes.Replace(data, []byte(from), []byte(to), 1), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Remove(filepath.Join(directory, "rules/no-debugger/rule.ts")); err != nil {
		t.Fatal(err)
	}
	prepareRegistry(t, directory)
	source := filepath.Join(directory, "mixed.ts")
	if err := os.WriteFile(source, []byte("/*😀*/debugger;\n"), 0644); err != nil {
		t.Fatal(err)
	}
	// Keep the manifest alive after the setup test returns.
	path := filepath.Join(directory, "manifest.txt")
	if err := os.WriteFile(path, []byte(source+"\tno-debugger\n"), 0644); err != nil {
		t.Fatal(err)
	}
	inputs := buildcache.Inputs{Name: "suggestion-alongside-lowered-v2", Files: []string{"stage1/cohere/lint", "internal", "cohere/TypeScript/tsc", "cohere/TypeScript-shim", "cohere/go.mod", "cohere/go.sum", "go.mod", "go.work", "cohere/go.work", "cohere/go.work.sum"}, Flags: []string{"automatic-fix=semicolon", "serialization-fixture=v1"}, Toolchain: []string{runtime.Version()}}
	lowered := buildcache.Product(t, inputs, func(out string) error {
		program, err := load.Load([]string{filepath.Join(directory, "main.ts")})
		if err != nil {
			return err
		}
		ir, err := lower.Lower(context.Background(), program)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(out, "lint.c"), []byte(native.C(ir)), 0644); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(out, "lint.mjs"), []byte(javascript.JavaScript(ir)), 0644)
	})
	inputs.Name = "suggestion-alongside-native-v2"
	inputs.Flags = append(inputs.Flags, "split=true", "jobs=4")
	inputs.Flags = append(inputs.Flags, native.Flags(native.Options{Sanitize: true, Split: true, Jobs: 4})...)
	inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("clang", "--version"))
	built := buildcache.Product(t, inputs, func(out string) error {
		source, err := os.ReadFile(filepath.Join(lowered, "lint.c"))
		if err != nil {
			return err
		}
		return native.Build(string(source), filepath.Join(out, "scanner"), native.Options{Sanitize: true, Split: true, Jobs: 4})
	})
	oracle, err := suggestionAlongsideGoOracle(t, directory, directory)
	if err != nil {
		t.Fatal(err)
	}
	suggestionAlongside.directory, suggestionAlongside.manifest, suggestionAlongside.oracle = directory, path, oracle
	suggestionAlongside.binary, suggestionAlongside.module = filepath.Join(built, "scanner"), filepath.Join(lowered, "lint.mjs")
	suggestionAlongside.ready = true
}

func suggestionAlongsideCheck(got, want []byte) string { return difference(got, want) }
func suggestionAlongsideShard(t *testing.T, index int) {
	t.Helper()
	suggestionAlongsideSetup(t)
	s := &suggestionAlongside
	want := execute(t, "", s.oracle, "--manifest", s.manifest).output
	if !bytes.Contains(want, []byte("fixed\t/*\\ud83d\\ude00*/;\\u000a")) {
		t.Fatalf("automatic fix lost: %s", want)
	}
	var got []byte
	switch index {
	case 0:
		got = node(t, s.directory, s.manifest, false).output
	case 1:
		got = runJavaScript(t, s.module, s.manifest, false).output
	case 2:
		got = execute(t, "", s.binary, "--manifest", s.manifest).output
	default:
		t.Fatalf("unknown shard %d", index)
	}
	if os.Getenv("ADAMIC_SUGGESTION_ALONGSIDE_PLANT") == suggestionAlongsideCases[index] {
		got = append(got, []byte("planted automatic fix mismatch\n")...)
	}
	if diff := suggestionAlongsideCheck(got, want); diff != "" {
		t.Fatalf("%s: %s", suggestionAlongsideCases[index], diff)
	}
	t.Logf("%s identical to Go: %d bytes", suggestionAlongsideCases[index], len(want))
}
func TestSuggestionAlongsideAutomaticFix_000(t *testing.T) {
	t.Parallel()
	suggestionAlongsideShard(t, 0)
}
func TestSuggestionAlongsideAutomaticFix_001(t *testing.T) {
	t.Parallel()
	suggestionAlongsideShard(t, 1)
}
func TestSuggestionAlongsideAutomaticFix_002(t *testing.T) {
	t.Parallel()
	suggestionAlongsideShard(t, 2)
}

func suggestionAlongsideUnion(t *testing.T) {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), "suggestion_alongside_shards_test.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]int{}
	assignments := make([]int, len(suggestionAlongsideCases))
	for _, declaration := range file.Decls {
		if f, ok := declaration.(*ast.FuncDecl); ok && strings.HasPrefix(f.Name.Name, "TestSuggestionAlongsideAutomaticFix_") {
			seen[f.Name.Name]++
			parallel, calls := 0, 0
			ast.Inspect(f.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				if selector, ok := call.Fun.(*ast.SelectorExpr); ok && selector.Sel.Name == "Parallel" {
					parallel++
				}
				if name, ok := call.Fun.(*ast.Ident); ok && name.Name == "suggestionAlongsideShard" {
					calls++
					if len(call.Args) != 2 {
						t.Fatal("wrong shard arguments")
					}
					literal, ok := call.Args[1].(*ast.BasicLit)
					if !ok {
						t.Fatal("shard index must be literal")
					}
					index, err := strconv.Atoi(literal.Value)
					if err != nil || index < 0 || index >= len(assignments) {
						t.Fatal("invalid shard index")
					}
					assignments[index]++
				}
				return true
			})
			if parallel != 1 || calls != 1 {
				t.Fatal("shard must be parallel and run exactly one case")
			}
		}
	}
	if len(seen) != testSuggestionAlongsideAutomaticFixShards || len(suggestionAlongsideCases) != testSuggestionAlongsideAutomaticFixShards {
		t.Fatal("shard enumeration changed")
	}
	for i := range suggestionAlongsideCases {
		if assignments[i] != 1 {
			t.Fatal("case missing or duplicated", i)
		}
		if seen[fmt.Sprintf("TestSuggestionAlongsideAutomaticFix_%03d", i)] != 1 {
			t.Fatal("missing or duplicated shard", i)
		}
	}
	t.Logf("union: %d runtime comparisons, one pinned source/rule case, each exactly once", len(suggestionAlongsideCases))
}

func TestSuggestionAlongsideAutomaticFixPlantedFailure(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestSuggestionAlongsideAutomaticFix_[0-9]+$", "-test.timeout=90s", "-test.v")
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if err == syscall.ESRCH {
			return os.ErrProcessDone
		}
		return err
	}
	command.WaitDelay = time.Second
	command.Env = append(os.Environ(), "ADAMIC_SUGGESTION_ALONGSIDE_PLANT=native")
	output, err := command.CombinedOutput()
	if err == nil || bytes.Count(output, []byte("--- FAIL: TestSuggestionAlongsideAutomaticFix_")) != 1 || !bytes.Contains(output, []byte("--- FAIL: TestSuggestionAlongsideAutomaticFix_002")) || !bytes.Contains(output, []byte("planted automatic fix mismatch")) {
		t.Fatalf("wrong planted failure: %v\n%s", err, output)
	}
	t.Log("planted mismatch caught only by shard 002")
}

// Use the same overlay and oracle, accepting only the Go download diagnostics
// that execute already permits on a fresh instance.
func suggestionAlongsideGoOracle(t *testing.T, sourceRoot, directory string) (string, error) {
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
	execute(t, root, "go", args...)
	return binary, nil
}
