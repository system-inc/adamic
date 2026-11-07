package yaml

import (
	"bytes"
	"context"
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
	_ "github.com/system-inc/adamic/internal/ccjobserver"
	"github.com/system-inc/adamic/internal/compilerinputs"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/toolchain"

	"github.com/system-inc/adamic/stage1/cohere/yaml/oracles"
)

var testFormatterMutants = []struct{ name, file, from, to string }{
	{"root final newline lost", "printer.ts", "const hard = !(", "const hard = false && !("},
	{
		"colon separation lost",
		"printer.ts",
		"this.layout.text(': '), printedValue",
		"this.layout.text(':'), printedValue",
	},
	{
		"flow trailing comma lost",
		"printer.ts",
		"this.layout.ifBreak(this.layout.text(','), this.empty, -1)",
		"this.layout.ifBreak(this.layout.text(''), this.empty, -1)",
	},
	{
		"batch backslash scan misses escapes",
		"main.ts",
		"text.charCodeAt(index) !== 92",
		"text.charCodeAt(index) !== 13",
	},
	{"emoji first unit range loses endpoint", "width.ts", "code <= last", "code < last"},
	{"emoji surrogate slot shifted", "width.ts", "code - 0x100000 + 0xd800", "code - 0x100000 + 0xd900"},
}

const testFormatterMutantsShards = 6

var formatterMutantsOracle struct {
	once    sync.Once
	product string
}

func formatterMutantsDeadline(test *testing.T) func() {
	test.Helper()
	timer := time.AfterFunc(90*time.Second, func() {
		fmt.Fprintf(os.Stderr, "%s exceeded 90s deadline\n", test.Name())
		os.Exit(124)
	})
	return func() { timer.Stop() }
}

// Commands share their unit's deadline. Kill the entire process group rather than
// leaving children behind.
func formatterMutantsCommand(callerContext context.Context, directory, name string, args ...string) ([]byte, error) {
	command := exec.CommandContext(callerContext, name, args...)
	command.Dir = directory
	if !strings.ContainsRune(name, filepath.Separator) {
		// A bare name (node) is a tool: the one internal/toolchain resolves, with a tool's environment, not PATH's, run
		// in directory with PWD to match, as toolchain.CommandIn runs one under a context it has no form for. A program
		// a test built (the Go oracle, a mutant) runs as named, with this process's environment. An error resolving it
		// fails the command as it starts.
		if path, pathError := toolchain.Path(name); pathError != nil {
			command.Err = pathError
		} else if absolute, absolutePathError := filepath.Abs(directory); absolutePathError != nil {
			command.Err = absolutePathError
		} else {
			command = exec.CommandContext(callerContext, path, args...)
			command.Dir = directory
			command.Env = append(toolchain.Env(), "PWD="+absolute)
		}
	}
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		killError := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if killError == syscall.ESRCH {
			return os.ErrProcessDone
		}
		return killError
	}
	var out, outError bytes.Buffer
	command.Stdout, command.Stderr = &out, &outError
	if runError := command.Run(); runError != nil {
		return nil, fmt.Errorf("%s: %w\n%s", name, runError, outError.Bytes())
	}
	if outError.Len() != 0 {
		return nil, fmt.Errorf("%s stderr: %s", name, outError.Bytes())
	}
	return out.Bytes(), nil
}

func formatterMutantsOracleProduct(test *testing.T) string {
	test.Helper()
	formatterMutantsOracle.once.Do(func() {
		product := oracles.FormatterMutantAnswers(test)
		data, _ := formatterMutantsReadCorpus(test, product)
		formatterMutantsUnion(test, data)
		formatterMutantsOracle.product = product
	})
	if formatterMutantsOracle.product == "" {
		test.Fatal("oracle preparation failed")
	}
	return formatterMutantsOracle.product
}

func TestProduct_YamlFindsFormatterMutantsOracle(test *testing.T) {
	test.Parallel()
	formatterMutantsOracleProduct(test)
}

func formatterMutantsReadCorpus(test *testing.T, product string) ([]byte, []byte) {
	test.Helper()
	cases, readError := os.ReadFile(filepath.Join(product, "cases.txt"))
	if readError != nil {
		test.Fatal(readError)
	}
	expected, readError := os.ReadFile(filepath.Join(product, "expected.txt"))
	if readError != nil {
		test.Fatal(readError)
	}
	return cases, expected
}

func formatterMutantsUnion(test *testing.T, cases []byte) {
	test.Helper()
	count := len(strings.Split(strings.TrimSuffix(string(cases), "\n"), "\n"))
	plan := formatterMutantsEnumeration(test)
	if len(testFormatterMutants) != testFormatterMutantsShards {
		test.Fatal("shard enumeration disagrees with declared count")
	}
	seen := map[string]bool{}
	for _, index := range plan {
		mutant := testFormatterMutants[index]
		for i := 0; i < count; i++ {
			id := fmt.Sprintf("%s/%d", mutant.name, i)
			if seen[id] {
				test.Fatalf("repeated case %s", id)
			}
			seen[id] = true
		}
	}
	if len(seen) != len(testFormatterMutants)*count {
		test.Fatal("incomplete union")
	}
	test.Logf("union: %d mutant/case pairs (%d complete corpus cases)", len(seen), count)
}

// Read the actual top-level wrappers, so a missing, duplicated or miswired
// shard cannot silently make the claimed union complete.
func formatterMutantsEnumeration(test *testing.T) []int {
	test.Helper()
	file, parseFileError := parser.ParseFile(token.NewFileSet(), "grain30_yaml_formatter_mutants_test.go", nil, 0)
	if parseFileError != nil {
		test.Fatal(parseFileError)
	}
	plan := map[int]int{}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok ||
			function.Name.Name == "TestFormatterMutants_Setup" ||
			!strings.HasPrefix(function.Name.Name, "TestFormatterMutants_") {
			continue
		}
		shard, parseError := strconv.Atoi(strings.TrimPrefix(function.Name.Name, "TestFormatterMutants_"))
		if parseError != nil {
			test.Fatal(parseError)
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
				test.Fatal("invalid shard call")
			}
			literal, ok := call.Args[1].(*ast.BasicLit)
			if !ok {
				test.Fatal("shard index must be literal")
			}
			index, parseError := strconv.Atoi(literal.Value)
			if parseError != nil || index != shard {
				test.Fatal("shard index disagrees with its name")
			}
			plan[shard] = index
			calls++
			return true
		})
		if calls != 1 {
			test.Fatal("shard must execute exactly one mutant")
		}
	}
	if len(plan) != testFormatterMutantsShards {
		test.Fatalf("enumerated %d top-level shards, want %d", len(plan), testFormatterMutantsShards)
	}
	indices := make([]int, testFormatterMutantsShards)
	for i := range indices {
		index, ok := plan[i]
		if !ok {
			test.Fatalf("missing shard %03d", i)
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

func formatterMutantsShard(test *testing.T, index int) {
	test.Helper()
	started := time.Now()
	oracle := formatterMutantsOracleProduct(test)
	sources, product := formatterMutantsProduct(test, index, 2)
	corpus, expected := formatterMutantsReadCorpus(test, oracle)
	test.Logf("setup: %.3fs", time.Since(started).Seconds())
	started = time.Now()
	defer func() { test.Logf("own work: %.3fs", time.Since(started).Seconds()) }()
	defer formatterMutantsDeadline(test)()
	timeoutContext, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	formatterMutantsUnion(test, corpus)
	entry := filepath.Join(sources, "main.ts")
	cases := filepath.Join(test.TempDir(), "cases.txt")
	if writeError := os.WriteFile(cases, corpus, 0644); writeError != nil {
		test.Fatal(writeError)
	}
	runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	for _, side := range []struct {
		name, command string
		args          []string
	}{
		{"native", filepath.Join(product, "mutant"), []string{"--cases", cases}},
		{"Node", "node", []string{"--disable-warning=ExperimentalWarning", runner, entry, "--cases", cases}},
	} {
		out, formatterMutantsCommandError := formatterMutantsCommand(timeoutContext, "", side.command, side.args...)
		if formatterMutantsCommandError != nil {
			test.Fatal(formatterMutantsCommandError)
		}
		if formatterMutantSurvivedError := formatterMutantSurvived(out, expected); formatterMutantSurvivedError != nil {
			test.Fatalf("%s: %v", side.name, formatterMutantSurvivedError)
		}
		test.Logf("%s successful execution, wrong bytes caught: %s", side.name, firstDifference(out, expected))
	}
}

var formatterMutantsProducts [testFormatterMutantsShards][3]struct {
	once             sync.Once
	sources, product string
}

func formatterMutantsProduct(test *testing.T, index, level int) (string, string) {
	test.Helper()
	state := &formatterMutantsProducts[index][level]
	state.once.Do(func() {
		state.sources, state.product = formatterMutantsBuildProduct(test, index, level)
	})
	if state.sources == "" {
		test.Fatal("mutant preparation failed")
	}
	return state.sources, state.product
}

func formatterMutantsBuildProduct(test *testing.T, index, level int) (string, string) {
	test.Helper()
	started := time.Now()
	defer func() { test.Logf("mutant %03d product %d: %.3fs", index, level, time.Since(started).Seconds()) }()
	// Shared setup is bounded by the product unit, before shard own work starts.
	backgroundContext := context.Background()
	mutant := testFormatterMutants[index]
	entries, globError := filepath.Glob("*.ts")
	if globError != nil {
		test.Fatal(globError)
	}
	// The sources product copies the .ts files and reads nothing else; the lowered and native products are the
	// compiler's, so they name it too.
	var copied []string
	for _, file := range entries {
		copied = append(copied, "stage1/cohere/yaml/"+file)
	}
	files := append(compilerinputs.Lowering(), copied...)
	flags := []string{mutant.file, mutant.from, mutant.to, "ADAMIC_NATIVE_SPLIT=" + os.Getenv("ADAMIC_NATIVE_SPLIT")}
	tools := []string{runtime.Version(), runtime.GOOS, runtime.GOARCH}
	var sources string
	if level > 0 {
		sources, _ = formatterMutantsProduct(test, index, 0)
	} else {
		sources = buildcache.Product(test, buildcache.Inputs{
			Name:      "yaml-formatter-mutant-sources",
			Files:     copied,
			Flags:     flags,
			Toolchain: tools,
		}, func(directory string) error {
			for _, file := range entries {
				source, readError := os.ReadFile(file)
				if readError != nil {
					return readError
				}
				if file == mutant.file {
					if strings.Count(string(source), mutant.from) != 1 {
						return fmt.Errorf("mutation site must occur once")
					}
					source = []byte(strings.Replace(string(source), mutant.from, mutant.to, 1))
				}
				if writeError := os.WriteFile(filepath.Join(directory, file), source, 0644); writeError != nil {
					return writeError
				}
			}
			return nil
		})
	}
	if level == 0 {
		return sources, ""
	}
	entry := filepath.Join(sources, "main.ts")
	var lowered string
	if level > 1 {
		_, lowered = formatterMutantsProduct(test, index, 1)
	} else {
		lowered = buildcache.Product(test, buildcache.Inputs{
			Name:      "yaml-formatter-mutant-lowered",
			Files:     files,
			Flags:     flags,
			Toolchain: tools,
		}, func(directory string) error {
			program, loadError := load.Load([]string{entry})
			if loadError != nil {
				return loadError
			}
			lowered, lowerError := lower.Lower(backgroundContext, program)
			if lowerError != nil {
				return lowerError
			}
			return os.WriteFile(filepath.Join(directory, "mutant.c"), []byte(native.C(lowered)), 0644)
		})
	}
	if level == 1 {
		return sources, lowered
	}
	options := native.Options{}
	product := buildcache.Product(test, buildcache.Inputs{
		Name:      "yaml-formatter-mutant-native",
		Files:     files,
		Flags:     append(flags, native.Flags(options)...),
		Toolchain: append(tools, buildcache.Tool("clang", "--version")),
	}, func(directory string) error {
		code, readError := os.ReadFile(filepath.Join(lowered, "mutant.c"))
		if readError != nil {
			return readError
		}
		return native.Build(string(code), filepath.Join(directory, "mutant"), options)
	})
	return sources, product
}

func TestProduct_YAMLFormatterMutant000Sources(test *testing.T) {
	test.Parallel()
	formatterMutantsProduct(test, 0, 0)
}

func TestProduct_YAMLFormatterMutant000Lowered(test *testing.T) {
	test.Parallel()
	formatterMutantsProduct(test, 0, 1)
}

func TestProduct_YAMLFormatterMutant000Native(test *testing.T) {
	test.Parallel()
	formatterMutantsProduct(test, 0, 2)
}

func TestProduct_YAMLFormatterMutant001Sources(test *testing.T) {
	test.Parallel()
	formatterMutantsProduct(test, 1, 0)
}

func TestProduct_YAMLFormatterMutant001Lowered(test *testing.T) {
	test.Parallel()
	formatterMutantsProduct(test, 1, 1)
}

func TestProduct_YAMLFormatterMutant001Native(test *testing.T) {
	test.Parallel()
	formatterMutantsProduct(test, 1, 2)
}

func TestProduct_YAMLFormatterMutant002Sources(test *testing.T) {
	test.Parallel()
	formatterMutantsProduct(test, 2, 0)
}

func TestProduct_YAMLFormatterMutant002Lowered(test *testing.T) {
	test.Parallel()
	formatterMutantsProduct(test, 2, 1)
}

func TestProduct_YAMLFormatterMutant002Native(test *testing.T) {
	test.Parallel()
	formatterMutantsProduct(test, 2, 2)
}

func TestProduct_YAMLFormatterMutant003Sources(test *testing.T) {
	test.Parallel()
	formatterMutantsProduct(test, 3, 0)
}

func TestProduct_YAMLFormatterMutant003Lowered(test *testing.T) {
	test.Parallel()
	formatterMutantsProduct(test, 3, 1)
}

func TestProduct_YAMLFormatterMutant003Native(test *testing.T) {
	test.Parallel()
	formatterMutantsProduct(test, 3, 2)
}

func TestProduct_YAMLFormatterMutant004Sources(test *testing.T) {
	test.Parallel()
	formatterMutantsProduct(test, 4, 0)
}

func TestProduct_YAMLFormatterMutant004Lowered(test *testing.T) {
	test.Parallel()
	formatterMutantsProduct(test, 4, 1)
}

func TestProduct_YAMLFormatterMutant004Native(test *testing.T) {
	test.Parallel()
	formatterMutantsProduct(test, 4, 2)
}

func TestProduct_YAMLFormatterMutant005Sources(test *testing.T) {
	test.Parallel()
	formatterMutantsProduct(test, 5, 0)
}

func TestProduct_YAMLFormatterMutant005Lowered(test *testing.T) {
	test.Parallel()
	formatterMutantsProduct(test, 5, 1)
}

func TestProduct_YAMLFormatterMutant005Native(test *testing.T) {
	test.Parallel()
	formatterMutantsProduct(test, 5, 2)
}
func TestFormatterMutants_000(test *testing.T) { test.Parallel(); formatterMutantsShard(test, 0) }
func TestFormatterMutants_001(test *testing.T) { test.Parallel(); formatterMutantsShard(test, 1) }
func TestFormatterMutants_002(test *testing.T) { test.Parallel(); formatterMutantsShard(test, 2) }
func TestFormatterMutants_003(test *testing.T) { test.Parallel(); formatterMutantsShard(test, 3) }
func TestFormatterMutants_004(test *testing.T) { test.Parallel(); formatterMutantsShard(test, 4) }
func TestFormatterMutants_005(test *testing.T) { test.Parallel(); formatterMutantsShard(test, 5) }

func TestFormatterMutantsPlantedFailure(test *testing.T) {
	test.Parallel()
	// Plant a survivor in one mutant's complete output. Exactly its owner must
	// reject it; all other mutants differ from the oracle and remain accepted.
	caught := []int{}
	for _, shard := range formatterMutantsEnumeration(test) {
		actual := []byte("wrong bytes")
		if shard == 3 {
			actual = []byte("oracle bytes")
		}
		if formatterMutantSurvived(actual, []byte("oracle bytes")) != nil {
			caught = append(caught, shard)
		}
	}
	if len(caught) != 1 || caught[0] != 3 {
		test.Fatalf("planted survivor caught by %v, want [3]", caught)
	}
	test.Log("planted survivor caught exactly by TestFormatterMutants_003")
}
