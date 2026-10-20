package yaml

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/toolchain"

	"github.com/system-inc/adamic/stage1/cohere/executable"
	"github.com/system-inc/adamic/stage1/cohere/yaml/oracles"
)

const repository = "../../.."

// yamlArtifacts holds the writes to ADAMIC_YAML_ARTIFACTS, whose fixed names parallel tests share (goCompose's files
// are written by TestComposeMatchGo and TestComposeMutants alike).
var yamlArtifacts sync.Mutex

func run(t *testing.T, directory string, environment []string, name string, args ...string) []byte {
	t.Helper()
	// Silent builds and buffered children use the shared first-output window.
	// With ten CPU burners, the longest output gap was 3.93s; the default
	// two-minute Stall leaves more than three times that gap as headroom.
	command := exec.Command(name, args...)
	command.Dir = directory
	command.Env = append(os.Environ(), environment...)
	if !strings.ContainsRune(name, filepath.Separator) {
		// A bare name (node) is a tool: the one internal/toolchain resolves, with a tool's environment plus the caller's
		// settings, not PATH's, run in directory with PWD to match. A program a test built (an oracle, a port) runs as
		// named, with this process's environment.
		command = toolchain.CommandIn(directory, name, args...)
		command.Env = append(command.Env, environment...)
	}
	var out, errOut bytes.Buffer
	command.Stdout, command.Stderr = &out, &errOut
	if err := childguard.Run(command, childguard.Options{}); err != nil {
		t.Fatalf("%s: %v\n%s", name, err, errOut.Bytes())
	}
	if errOut.Len() != 0 {
		t.Fatalf("%s stderr: %s", name, errOut.Bytes())
	}
	return out.Bytes()
}

func lexCases(t *testing.T) (string, int, int) { return oracles.LexicalCases(t) }

func goLexer(t *testing.T, cases string) []byte {
	t.Helper()
	binary := goLexerOracle(t)
	if directory := os.Getenv("ADAMIC_YAML_ARTIFACTS"); directory != "" {
		yamlArtifacts.Lock()
		defer yamlArtifacts.Unlock()
		if err := os.MkdirAll(directory, 0755); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(binary)
		if err != nil {
			t.Fatal(err)
		}
		// A program copied while sibling tests fork goes through the fork-safe writer, or a fork can hold its writable
		// descriptor and a later exec of it fails with ETXTBSY (#6hjazy4).
		if err := executable.Write(filepath.Join(directory, "go-lexer"), data); err != nil {
			t.Fatal(err)
		}
	}
	return run(t, "", nil, binary, cases)
}

// goLexerOracle is formatter_comparison over testdata/lexer_go.go, a product built ahead (#5qykzj5).
func goLexerOracle(t *testing.T) string { return oracles.Lexer(t) }

func TestProduct_YamlFindsLexerGo(t *testing.T) {
	t.Parallel()
	goLexerOracle(t)
}

func firstDifference(a, b []byte) string {
	for index := 0; index < min(len(a), len(b)); index++ {
		if a[index] != b[index] {
			start, end := max(0, index-80), min(min(len(a), len(b)), index+80)
			return fmt.Sprintf("byte %d: got %q, want %q", index, a[start:end], b[start:end])
		}
	}
	return fmt.Sprintf("lengths %d and %d", len(a), len(b))
}

var lexerMutants = yamlMutants{family: "lexer", entry: "lex_main.ts", files: []string{"lexer.ts", "lex_main.ts"}, mutated: "lexer.ts", mutations: []yamlMutation{
	{"keep chomping lost", "this.blockScalarKeep = true;", "this.blockScalarKeep = false;"},
	{"tab not whitespace", "|| unit === 9;", "|| unit === 8;"},
	{"BOM not split", "character(line, 0) === '\\ufeff'", "character(line, 0) === 'X'"},
	{"cached NUL becomes space", "units.push(String.fromCodePoint(code))", "units.push(String.fromCodePoint(code === 0 ? 32 : code))"},
	{"numeric whitespace excludes space", "return unit === -1 || unit === 32 || unit === 10 || unit === 13 || unit === 9;", "return unit === -1 || unit === 31 || unit === 10 || unit === 13 || unit === 9;"},
}}

func TestProduct_YAMLLexerMutant000(t *testing.T) { t.Parallel(); lexerMutants.product(t, 0) }
func TestProduct_YAMLLexerMutant001(t *testing.T) { t.Parallel(); lexerMutants.product(t, 1) }
func TestProduct_YAMLLexerMutant002(t *testing.T) { t.Parallel(); lexerMutants.product(t, 2) }
func TestProduct_YAMLLexerMutant003(t *testing.T) { t.Parallel(); lexerMutants.product(t, 3) }
func TestProduct_YAMLLexerMutant004(t *testing.T) { t.Parallel(); lexerMutants.product(t, 4) }

// lexerMutantsShards is the count of lexerMutants.mutations, each its own top-level test so Loom's planner can cut
// between mutants, which it can't between subtests.
const lexerMutantsShards = 5

// lexerMutantsAnswers holds what every mutant compares against, made once per process: the lexical cases and the Go
// lexer's answer to them.
var lexerMutantsAnswers struct {
	once            sync.Once
	cases, expected []byte
}

func lexerMutantsExpected(t *testing.T) ([]byte, []byte) {
	t.Helper()
	lexerMutantsAnswers.once.Do(func() {
		// The cases file lives in the first caller's temporary directory, gone when that test ends, so its bytes are kept.
		path, _, _ := lexCases(t)
		cases, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		lexerMutantsAnswers.expected = goLexer(t, path)
		lexerMutantsAnswers.cases = cases
	})
	if lexerMutantsAnswers.cases == nil {
		t.Fatal("the Go lexer's answers failed earlier in this process")
	}
	return lexerMutantsAnswers.cases, lexerMutantsAnswers.expected
}

// lexerMutantsEnumeration reads the top-level wrappers themselves, so a missing, duplicated or miswired shard can't
// leave a mutant unrun while the count still agrees.
func lexerMutantsEnumeration(t *testing.T) []int {
	t.Helper()
	if len(lexerMutants.mutations) != lexerMutantsShards {
		t.Fatalf("%d lexer mutants, but lexerMutantsShards is %d", len(lexerMutants.mutations), lexerMutantsShards)
	}
	file, err := parser.ParseFile(token.NewFileSet(), "lexer_mutants_test.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var plan []int
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || !strings.HasPrefix(function.Name.Name, "TestLexerMutants_") {
			continue
		}
		shard, err := strconv.Atoi(strings.TrimPrefix(function.Name.Name, "TestLexerMutants_"))
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
			if !ok || name.Name != "lexerMutantsShard" {
				return true
			}
			if len(call.Args) != 2 {
				t.Fatalf("%s: invalid shard call", function.Name.Name)
			}
			literal, ok := call.Args[1].(*ast.BasicLit)
			if !ok {
				t.Fatalf("%s: shard index must be a literal", function.Name.Name)
			}
			if index, err := strconv.Atoi(literal.Value); err != nil || index != shard {
				t.Fatalf("%s runs mutant %s", function.Name.Name, literal.Value)
			}
			calls++
			return true
		})
		if calls != 1 {
			t.Fatalf("%s runs %d mutants, want one", function.Name.Name, calls)
		}
		plan = append(plan, shard)
	}
	return plan
}

// lexerMutantSurvived is a side's verdict: a mutant whose output still equals the Go lexer's was missed.
func lexerMutantSurvived(side string, output, expected []byte) error {
	if bytes.Equal(output, expected) {
		return fmt.Errorf("%s missed mutant", side)
	}
	return nil
}

func lexerMutantsShard(t *testing.T, index int) {
	t.Helper()
	setup := time.Now()
	mutant := lexerMutants.mutations[index]
	product := lexerMutants.product(t, index)
	corpus, expected := lexerMutantsExpected(t)
	cases := filepath.Join(t.TempDir(), "cases.txt")
	if err := os.WriteFile(cases, corpus, 0644); err != nil {
		t.Fatal(err)
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s; setup %.3fs", mutant.name, time.Since(setup).Seconds())
	nativeOut := run(t, "", []string{"ASAN_OPTIONS=detect_leaks=1"}, filepath.Join(product, "mutant"), cases)
	// The mutant's emitted JavaScript must run cleanly, as it always had to; native and Node are compared with Go.
	run(t, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(product, "mutant.mjs"), cases)
	node := run(t, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(product, "sources", "lex_main.ts"), cases)
	for _, side := range []struct {
		name   string
		output []byte
	}{{"native", nativeOut}, {"Node", node}} {
		if err := lexerMutantSurvived(side.name, side.output, expected); err != nil {
			t.Fatal(err)
		}
		t.Logf("%s successful execution, wrong bytes caught: %s", side.name, firstDifference(side.output, expected))
	}
}

func TestLexerMutantsUnion(t *testing.T) {
	t.Parallel()
	visits := make([]int, len(lexerMutants.mutations))
	for _, index := range lexerMutantsEnumeration(t) {
		if index < 0 || index >= len(visits) {
			t.Fatalf("TestLexerMutants_%03d names no mutant", index)
		}
		visits[index]++
	}
	for index, count := range visits {
		if count != 1 {
			t.Fatalf("mutant %d (%s) run by %d top-level tests, want one", index, lexerMutants.mutations[index].name, count)
		}
	}
	t.Logf("union: all %d lexer mutants run, each by exactly one top-level test", len(visits))
}

func TestLexerMutantsPlantedFailure(t *testing.T) {
	t.Parallel()
	// Plant a survivor in one mutant's output. Exactly its owner must reject it; every other mutant differs from the Go
	// lexer and stays accepted.
	var caught []int
	for _, shard := range lexerMutantsEnumeration(t) {
		output := []byte("wrong bytes")
		if shard == 2 {
			output = []byte("Go lexer bytes")
		}
		if lexerMutantSurvived("native", output, []byte("Go lexer bytes")) != nil {
			caught = append(caught, shard)
		}
	}
	if len(caught) != 1 || caught[0] != 2 {
		t.Fatalf("planted survivor caught by %v, want [2]", caught)
	}
	t.Log("planted survivor caught exactly by TestLexerMutants_002")
}

func TestLexerMutants_000(t *testing.T) { t.Parallel(); lexerMutantsShard(t, 0) }
func TestLexerMutants_001(t *testing.T) { t.Parallel(); lexerMutantsShard(t, 1) }
func TestLexerMutants_002(t *testing.T) { t.Parallel(); lexerMutantsShard(t, 2) }
func TestLexerMutants_003(t *testing.T) { t.Parallel(); lexerMutantsShard(t, 3) }
func TestLexerMutants_004(t *testing.T) { t.Parallel(); lexerMutantsShard(t, 4) }
