package yaml

import (
	"bytes"
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
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

var testFormatterMutantsOracle struct {
	once            sync.Once
	cases, expected []byte
	count           int
	ready           bool
}

// Only the immutable corpus and Go answer are shared in memory. Products are
// built once per content key, including when shards run in separate processes.
func formatterMutantsSetup(t *testing.T) {
	t.Helper()
	testFormatterMutantsOracle.once.Do(func() {
		started := time.Now()
		defer func() { t.Logf("TestFormatterMutants (setup): %.3fs", time.Since(started).Seconds()) }()
		cases, _, count := formatCases(t)
		data, err := os.ReadFile(cases)
		if err != nil {
			t.Fatal(err)
		}
		if len(strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")) != count {
			t.Fatal("corpus count disagrees with live enumeration")
		}
		testFormatterMutantsOracle.cases = data
		testFormatterMutantsOracle.expected = goFormat(t, cases)
		testFormatterMutantsOracle.count = count
		testFormatterMutantsOracle.ready = true
	})
	if !testFormatterMutantsOracle.ready {
		t.Fatal("shared oracle setup failed")
	}
	plan := formatterMutantsEnumeration(t)
	if len(testFormatterMutants) != testFormatterMutantsShards {
		t.Fatal("shard enumeration disagrees with declared count")
	}
	// The pinned mutant table partitions the complete live corpus by mutant.
	seen := map[string]bool{}
	for _, index := range plan {
		mutant := testFormatterMutants[index]
		for i := 0; i < testFormatterMutantsOracle.count; i++ {
			id := fmt.Sprintf("%s/%d", mutant.name, i)
			if seen[id] {
				t.Fatalf("repeated case %s", id)
			}
			seen[id] = true
		}
	}
	if len(seen) != len(testFormatterMutants)*testFormatterMutantsOracle.count {
		t.Fatal("incomplete union")
	}
	t.Logf("union: %d mutant/case pairs (%d complete corpus cases)", len(seen), testFormatterMutantsOracle.count)
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
		if !ok || !strings.HasPrefix(function.Name.Name, "TestFormatterMutants_") {
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
	formatterMutantsSetup(t)
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
		lowered, err := lower.Lower(context.Background(), program)
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
	if err := os.WriteFile(cases, testFormatterMutantsOracle.cases, 0644); err != nil {
		t.Fatal(err)
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	for _, side := range []struct {
		name string
		out  []byte
	}{
		{"native", run(t, "", nil, filepath.Join(product, "mutant"), "--cases", cases)},
		{"Node", run(t, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, entry, "--cases", cases)},
	} {
		if err := formatterMutantSurvived(side.out, testFormatterMutantsOracle.expected); err != nil {
			t.Fatalf("%s: %v", side.name, err)
		}
		t.Logf("%s successful execution, wrong bytes caught: %s", side.name, firstDifference(side.out, testFormatterMutantsOracle.expected))
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
