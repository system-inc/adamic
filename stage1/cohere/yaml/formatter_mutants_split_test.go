package yaml

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
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

var testFormatterMutants = []struct{ name, file, from, to string }{
	{"root final newline lost", "printer.ts", "const hard = !(", "const hard = false && !("},
	{"colon separation lost", "printer.ts", "this.layout.text(': '), printedValue", "this.layout.text(':'), printedValue"},
	{"flow trailing comma lost", "printer.ts", "this.layout.ifBreak(this.layout.text(','), this.empty, -1)", "this.layout.ifBreak(this.layout.text(''), this.empty, -1)"},
	{"batch backslash scan misses escapes", "main.ts", "text.charCodeAt(index) !== 92", "text.charCodeAt(index) !== 13"},
	{"emoji first unit range loses endpoint", "width.ts", "code <= last", "code < last"},
	{"emoji surrogate slot shifted", "width.ts", "code - 0x100000 + 0xd800", "code - 0x100000 + 0xd900"},
}

const testFormatterMutantsShards = 6

// Written only by the serial setup test, before parallel leaves are released.
// This also shares the fresh product when ADAMIC_BUILD_CACHE=off.
var formatterMutantsSetupProduct string

// The setup product persists across one-shard invocations. Leaves never build
// this product: a cold leaf-only invocation must run the setup test first.
func formatterMutantsSetupInputs() buildcache.Inputs {
	return buildcache.Inputs{
		Name:      "yaml-formatter-mutants-oracle-and-corpus-v2",
		Files:     []string{"cohere", "internal/corpusfiles", "stage1/cohere/yaml", "go.mod", "go.work"},
		Flags:     []string{"go build -overlay ./command/formatter_comparison", "GOFLAGS=" + os.Getenv("GOFLAGS"), "CGO_ENABLED=" + os.Getenv("CGO_ENABLED")},
		Toolchain: []string{buildcache.Tool("go", "version"), buildcache.Tool("go", "env", "GOOS", "GOARCH", "CGO_ENABLED")},
	}
}

func formatterMutantsDeadline(t *testing.T) func() {
	t.Helper()
	timer := time.AfterFunc(90*time.Second, func() {
		fmt.Fprintf(os.Stderr, "%s exceeded 90s deadline\n", t.Name())
		os.Exit(124)
	})
	return func() { timer.Stop() }
}

// Commands share their unit's deadline. Kill the entire process group, including
// compilers spawned by go build, rather than leaving children behind.
func formatterMutantsCommand(ctx context.Context, directory, name string, args ...string) ([]byte, error) {
	command := exec.CommandContext(ctx, name, args...)
	command.Dir = directory
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if err == syscall.ESRCH {
			return os.ErrProcessDone
		}
		return err
	}
	var out, errOut bytes.Buffer
	command.Stdout, command.Stderr = &out, &errOut
	if err := command.Run(); err != nil {
		return nil, fmt.Errorf("%s: %w\n%s", name, err, errOut.Bytes())
	}
	if errOut.Len() != 0 {
		return nil, fmt.Errorf("%s stderr: %s", name, errOut.Bytes())
	}
	return out.Bytes(), nil
}

// Not parallel: prepares the shared formatter-mutant corpus and Go oracle before parallel leaves are released.
func TestFormatterMutants_Setup(t *testing.T) {
	defer formatterMutantsDeadline(t)()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	started := time.Now()
	product := buildcache.Product(t, formatterMutantsSetupInputs(), func(directory string) error {
		cases, _, count := formatCases(t)
		data, err := os.ReadFile(cases)
		if err != nil {
			return err
		}
		if len(strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")) != count {
			return fmt.Errorf("corpus count disagrees with live enumeration")
		}
		root, err := filepath.Abs(repository)
		if err != nil {
			return err
		}
		source, err := filepath.Abs("testdata/format_go.go")
		if err != nil {
			return err
		}
		overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{filepath.Join(root, "cohere/command/formatter_comparison/main.go"): source}})
		if err != nil {
			return err
		}
		path := filepath.Join(directory, "overlay.json")
		if err := os.WriteFile(path, overlay, 0644); err != nil {
			return err
		}
		binary := filepath.Join(directory, "go-format")
		if _, err := formatterMutantsCommand(ctx, filepath.Join(root, "cohere"), "go", "build", "-overlay", path, "-o", binary, "./command/formatter_comparison"); err != nil {
			return err
		}
		expected, err := formatterMutantsCommand(ctx, "", binary, "--cases", cases)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(directory, "cases.txt"), data, 0644); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(directory, "expected.txt"), expected, 0644)
	})
	data, _ := formatterMutantsReadCorpus(t, product)
	formatterMutantsUnion(t, data)
	formatterMutantsSetupProduct = product
	t.Logf("TestFormatterMutants (setup): %.3fs", time.Since(started).Seconds())
}

func formatterMutantsReadCorpus(t *testing.T, product string) ([]byte, []byte) {
	t.Helper()
	cases, err := os.ReadFile(filepath.Join(product, "cases.txt"))
	if err != nil {
		t.Fatal(err)
	}
	expected, err := os.ReadFile(filepath.Join(product, "expected.txt"))
	if err != nil {
		t.Fatal(err)
	}
	return cases, expected
}

func formatterMutantsUnion(t *testing.T, cases []byte) {
	t.Helper()
	count := len(strings.Split(strings.TrimSuffix(string(cases), "\n"), "\n"))
	plan := formatterMutantsEnumeration(t)
	if len(testFormatterMutants) != testFormatterMutantsShards {
		t.Fatal("shard enumeration disagrees with declared count")
	}
	seen := map[string]bool{}
	for _, index := range plan {
		mutant := testFormatterMutants[index]
		for i := 0; i < count; i++ {
			id := fmt.Sprintf("%s/%d", mutant.name, i)
			if seen[id] {
				t.Fatalf("repeated case %s", id)
			}
			seen[id] = true
		}
	}
	if len(seen) != len(testFormatterMutants)*count {
		t.Fatal("incomplete union")
	}
	t.Logf("union: %d mutant/case pairs (%d complete corpus cases)", len(seen), count)
}

// Read the actual top-level wrappers, so a missing, duplicated or miswired
// shard cannot silently make the claimed union complete.
func formatterMutantsEnumeration(t *testing.T) []int {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), "formatter_mutants_split_test.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	plan := map[int]int{}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Name.Name == "TestFormatterMutants_Setup" || !strings.HasPrefix(function.Name.Name, "TestFormatterMutants_") {
			continue
		}
		shard, err := strconv.Atoi(strings.TrimPrefix(function.Name.Name, "TestFormatterMutants_"))
		if err != nil {
			t.Fatal(err)
		}
		calls := 0
		ast.Inspect(function.Body, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			name, ok := call.Fun.(*ast.Ident)
			if !ok || name.Name != "formatterMutantsShard" {
				return true
			}
			if len(call.Args) != 2 {
				t.Fatal("invalid shard call")
			}
			literal, ok := call.Args[1].(*ast.BasicLit)
			if !ok {
				t.Fatal("shard index must be literal")
			}
			index, err := strconv.Atoi(literal.Value)
			if err != nil || index != shard {
				t.Fatal("shard index disagrees with its name")
			}
			plan[shard] = index
			calls++
			return true
		})
		if calls != 1 {
			t.Fatal("shard must execute exactly one mutant")
		}
	}
	if len(plan) != testFormatterMutantsShards {
		t.Fatalf("enumerated %d top-level shards, want %d", len(plan), testFormatterMutantsShards)
	}
	indices := make([]int, testFormatterMutantsShards)
	for i := range indices {
		index, ok := plan[i]
		if !ok {
			t.Fatalf("missing shard %03d", i)
		}
		indices[i] = index
	}
	return indices
}

func formatterMutantSurvived(actual, expected []byte) error {
	if bytes.Equal(actual, expected) {
		return fmt.Errorf("missed mutant")
	}
	return nil
}

func formatterMutantsShard(t *testing.T, index int) {
	t.Helper()
	// Fetch shared state before starting this leaf's budget. The callback cannot
	// prepare it lazily; only TestFormatterMutants_Setup is allowed to build it.
	oracle := formatterMutantsSetupProduct
	if oracle == "" {
		oracle = buildcache.Product(t, formatterMutantsSetupInputs(), func(string) error {
			return fmt.Errorf("shared setup missing: run TestFormatterMutants_Setup before this shard")
		})
	}
	corpus, expected := formatterMutantsReadCorpus(t, oracle)
	defer formatterMutantsDeadline(t)()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	formatterMutantsUnion(t, corpus)
	mutant := testFormatterMutants[index]
	entries, err := filepath.Glob("*.ts")
	if err != nil {
		t.Fatal(err)
	}
	files := []string{"internal", "cohere", "go.mod"}
	for _, file := range entries {
		files = append(files, "stage1/cohere/yaml/"+file)
	}
	flags := []string{mutant.file, mutant.from, mutant.to, "ADAMIC_NATIVE_SPLIT=" + os.Getenv("ADAMIC_NATIVE_SPLIT")}
	tools := []string{runtime.Version(), runtime.GOOS, runtime.GOARCH}
	sources := buildcache.Product(t, buildcache.Inputs{Name: "yaml-formatter-mutant-sources", Files: files, Flags: flags, Toolchain: tools}, func(directory string) error {
		for _, file := range entries {
			source, err := os.ReadFile(file)
			if err != nil {
				return err
			}
			if file == mutant.file {
				if strings.Count(string(source), mutant.from) != 1 {
					return fmt.Errorf("mutation site must occur once")
				}
				source = []byte(strings.Replace(string(source), mutant.from, mutant.to, 1))
			}
			if err := os.WriteFile(filepath.Join(directory, file), source, 0644); err != nil {
				return err
			}
		}
		return nil
	})
	entry := filepath.Join(sources, "main.ts")
	lowered := buildcache.Product(t, buildcache.Inputs{Name: "yaml-formatter-mutant-lowered", Files: files, Flags: flags, Toolchain: tools}, func(directory string) error {
		program, err := load.Load([]string{entry})
		if err != nil {
			return err
		}
		lowered, err := lower.Lower(ctx, program)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(directory, "mutant.c"), []byte(native.C(lowered)), 0644)
	})
	options := native.Options{}
	product := buildcache.Product(t, buildcache.Inputs{Name: "yaml-formatter-mutant-native", Files: files, Flags: append(flags, native.Flags(options)...), Toolchain: append(tools, buildcache.Tool("clang", "--version"))}, func(directory string) error {
		code, err := os.ReadFile(filepath.Join(lowered, "mutant.c"))
		if err != nil {
			return err
		}
		return native.Build(string(code), filepath.Join(directory, "mutant"), options)
	})
	cases := filepath.Join(t.TempDir(), "cases.txt")
	if err := os.WriteFile(cases, corpus, 0644); err != nil {
		t.Fatal(err)
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	for _, side := range []struct {
		name, command string
		args          []string
	}{
		{"native", filepath.Join(product, "mutant"), []string{"--cases", cases}},
		{"Node", "node", []string{"--disable-warning=ExperimentalWarning", runner, entry, "--cases", cases}},
	} {
		out, err := formatterMutantsCommand(ctx, "", side.command, side.args...)
		if err != nil {
			t.Fatal(err)
		}
		if err := formatterMutantSurvived(out, expected); err != nil {
			t.Fatalf("%s: %v", side.name, err)
		}
		t.Logf("%s successful execution, wrong bytes caught: %s", side.name, firstDifference(out, expected))
	}
}

func TestFormatterMutants_000(t *testing.T) { t.Parallel(); formatterMutantsShard(t, 0) }
func TestFormatterMutants_001(t *testing.T) { t.Parallel(); formatterMutantsShard(t, 1) }
func TestFormatterMutants_002(t *testing.T) { t.Parallel(); formatterMutantsShard(t, 2) }
func TestFormatterMutants_003(t *testing.T) { t.Parallel(); formatterMutantsShard(t, 3) }
func TestFormatterMutants_004(t *testing.T) { t.Parallel(); formatterMutantsShard(t, 4) }
func TestFormatterMutants_005(t *testing.T) { t.Parallel(); formatterMutantsShard(t, 5) }

func TestFormatterMutantsPlantedFailure(t *testing.T) {
	t.Parallel()
	// Plant a survivor in one mutant's complete output. Exactly its owner must
	// reject it; all other mutants differ from the oracle and remain accepted.
	caught := []int{}
	for _, shard := range formatterMutantsEnumeration(t) {
		actual := []byte("wrong bytes")
		if shard == 3 {
			actual = []byte("oracle bytes")
		}
		if formatterMutantSurvived(actual, []byte("oracle bytes")) != nil {
			caught = append(caught, shard)
		}
	}
	if len(caught) != 1 || caught[0] != 3 {
		t.Fatalf("planted survivor caught by %v, want [3]", caught)
	}
	t.Log("planted survivor caught exactly by TestFormatterMutants_003")
}
