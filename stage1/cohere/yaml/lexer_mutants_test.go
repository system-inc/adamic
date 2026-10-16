package yaml

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/toolchain"

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
		if err := os.WriteFile(filepath.Join(directory, "go-lexer"), data, 0755); err != nil {
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

func TestLexerMutants(t *testing.T) {
	t.Parallel()
	cases, _, _ := lexCases(t)
	expected := goLexer(t, cases)
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	for index, mutant := range lexerMutants.mutations {
		// Not parallel: one mutant at a time keeps this test to one child at a time, its share.
		t.Run(mutant.name, func(t *testing.T) {
			product := lexerMutants.product(t, index)
			nativeOut := run(t, "", []string{"ASAN_OPTIONS=detect_leaks=1"}, filepath.Join(product, "mutant"), cases)
			// The mutant's emitted JavaScript must run cleanly, as it always had to; native and Node are compared with Go.
			run(t, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(product, "mutant.mjs"), cases)
			node := run(t, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(product, "sources", "lex_main.ts"), cases)
			for _, side := range []struct {
				name   string
				output []byte
			}{{"native", nativeOut}, {"Node", node}} {
				if bytes.Equal(side.output, expected) {
					t.Fatalf("%s missed mutant", side.name)
				}
				t.Logf("%s successful execution, wrong bytes caught: %s", side.name, firstDifference(side.output, expected))
			}
		})
	}
}
