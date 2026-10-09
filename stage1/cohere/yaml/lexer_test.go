package yaml

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/corpusfiles"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

const repository = "../../.."

func run(t *testing.T, directory string, environment []string, name string, args ...string) []byte {
	t.Helper()
	// Silent builds and buffered children use the shared first-output window.
	// With ten CPU burners, the longest output gap was 3.93s; the default
	// two-minute Stall leaves more than three times that gap as headroom.
	command := exec.Command(name, args...)
	command.Dir = directory
	command.Env = append(os.Environ(), environment...)
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

func lexCases(t *testing.T) (string, int, int) {
	t.Helper()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	var texts []string
	files := 0
	patterns := []string{"*.yaml", "*.yml"}
	paths := corpusfiles.Upstream(t, filepath.Join(root, "cohere"), corpusfiles.CohereCommit, []string{".github/workflows"}, patterns)
	paths = append(paths, corpusfiles.Upstream(t, filepath.Join(root, "cohere/TypeScript"), corpusfiles.TypeScriptGoCommit, []string{".custom-gcl.yml", ".github", ".golangci.yml", "tools/pipelines"}, patterns)...)
	sort.Strings(paths)
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !utf8.Valid(data) {
			t.Fatalf("invalid UTF-8: %s", path)
		}
		texts = append(texts, string(data))
		files++
	}
	// Each lexical state, separators, directives, tags, markers, chomping and Unicode boundaries.
	texts = append(texts, "", "\ufeff", "\ufeffa: b\n", "%YAML 1.1 # c\n---\na: b\n...\n", "a: |+\n  b\n\n\n", "a: |-\n  b\n\n", "a: >2-\n  b\n", "{a: [1, 2], b: \"q\\\"r\"}\n", "a: 'q''r'\n", "a: !<tag:example.org,%E2%82%AC> &ref b\nc: *ref\n", "a: x\n y\nz: x\n", "a: [x\nq: y\n", "a: |\n\tb\n", "a: \"x\ny\"\n", "---x\n...\n", "😀: café\n中: x\n", "a:\tx\r\n# b\r\n", "\x00\n")
	values := []string{"x", "true", "null", "1", "'q''r'", `"a\nb"`, "[x,y]", "{a:b, c: d}", "&a x", "*a", "!str x", "😀", "中文"}
	for _, value := range values {
		for _, indicator := range []string{"", "- ", "a: ", "? ", "[", "{a: "} {
			texts = append(texts, indicator+value+"\n", indicator+value+" # c\r\n")
		}
	}
	random := rand.New(rand.NewSource(20261006))
	alphabet := []rune("abc012:?!&*[]{}|>+-.,%'\"# \\\t\r\n😀中\ufeff\x00")
	for count := 0; count < 2000; count++ {
		var text strings.Builder
		for size := random.Intn(180); size > 0; size-- {
			text.WriteRune(alphabet[random.Intn(len(alphabet))])
		}
		texts = append(texts, text.String())
	}
	escape := strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t")
	var source strings.Builder
	cases := 0
	for index, text := range texts {
		chunks := []int{0}
		if index >= files {
			chunks = append(chunks, 1, 2, 7)
		}
		for _, chunk := range chunks {
			fmt.Fprintf(&source, "%d\t%s\n", chunk, escape.Replace(text))
			cases++
		}
	}
	path := filepath.Join(t.TempDir(), "cases.txt")
	if err := os.WriteFile(path, []byte(source.String()), 0644); err != nil {
		t.Fatal(err)
	}
	return path, files, cases
}

func goLexer(t *testing.T, cases string) []byte {
	t.Helper()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	source, err := filepath.Abs("testdata/lexer_go.go")
	if err != nil {
		t.Fatal(err)
	}
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{filepath.Join(root, "cohere/command/formatter_comparison/main.go"): source}})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "overlay.json")
	if err := os.WriteFile(path, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "go-lexer")
	run(t, filepath.Join(root, "cohere"), nil, "go", "build", "-overlay", path, "-o", binary, "./command/formatter_comparison")
	if directory := os.Getenv("ADAMIC_YAML_ARTIFACTS"); directory != "" {
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

func nativeLexer(t *testing.T, directory, cases string, sanitizer bool) ([]byte, []byte) {
	t.Helper()
	entry := filepath.Join(directory, "lex_main.ts")
	program, err := load.Load([]string{entry})
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "lexer")
	if err := native.Build(native.C(lowered), binary, native.Options{Sanitize: sanitizer}); err != nil {
		t.Fatal(err)
	}
	environment := []string{"ASAN_OPTIONS=detect_leaks=1"}
	out := run(t, "", environment, binary, cases)
	emitted := filepath.Join(t.TempDir(), "lexer.mjs")
	if err := os.WriteFile(emitted, []byte(javascript.JavaScript(lowered)), 0644); err != nil {
		t.Fatal(err)
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	return out, run(t, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, emitted, cases)
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

// Not parallel: native.Build writes the shared user cache (adamic/runtime or adamic/units); fixed filenames in ADAMIC_YAML_ARTIFACTS.
func TestLexerMatchesGo(t *testing.T) {
	cases, files, count := lexCases(t)
	expected := goLexer(t, cases)
	if artifacts := os.Getenv("ADAMIC_YAML_ARTIFACTS"); artifacts != "" {
		data, err := os.ReadFile(cases)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(artifacts, "lexer-cases.txt"), data, 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(artifacts, "lexer-expected.txt"), expected, 0644); err != nil {
			t.Fatal(err)
		}
	}
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	nativeOut, emitted := nativeLexer(t, directory, cases, true)
	node := run(t, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "lex_main.ts"), cases)
	for _, side := range []struct {
		name   string
		output []byte
	}{{"native ASan/UBSan/LSan", nativeOut}, {"Node source", node}, {"emitted JavaScript", emitted}} {
		if !bytes.Equal(side.output, expected) {
			t.Fatalf("%s: %s", side.name, firstDifference(side.output, expected))
		}
	}
	library := os.Getenv("ADAMIC_YAML_LIBRARY")
	if library == "" {
		t.Skip("set ADAMIC_YAML_LIBRARY to an npm install of yaml@2.9.0 and prettier@3.9.6; the gate skips this oracle until #xq2ecw6 (setup --gate-inputs) installs it")
	}
	external := run(t, "", nil, "node", "testdata/lexer_library.mjs", library, cases)
	if !bytes.Equal(external, expected) {
		t.Fatal(firstDifference(external, expected))
	}
	t.Logf("%d repository files; %d complete/chunked cases; %d answer bytes: Go, native, Node source, emitted JavaScript, yaml 2.9.0 identical", files, count, len(expected))
}

// Not parallel: native.Build writes the shared user cache (adamic/runtime or adamic/units); fixed filenames in ADAMIC_YAML_ARTIFACTS.
func TestLexerMutants(t *testing.T) {
	cases, _, _ := lexCases(t)
	expected := goLexer(t, cases)
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	for _, mutant := range []struct{ name, from, to string }{
		{"empty character waits at end", "if(!nextUnit && !this.atEnd)", "if(!nextUnit)"},
		{"keep chomping lost", "this.blockScalarKeep = true;", "this.blockScalarKeep = false;"},
		{"tab not whitespace", "|| unit === 9;", "|| unit === 8;"},
		{"BOM not split", "character(line, 0) === '\\ufeff'", "character(line, 0) === 'X'"},
		{"cached NUL becomes space", "units.push(String.fromCodePoint(code))", "units.push(String.fromCodePoint(code === 0 ? 32 : code))"},
		{"numeric whitespace excludes space", "return unit === -1 || unit === 32 || unit === 10 || unit === 13 || unit === 9;", "return unit === -1 || unit === 31 || unit === 10 || unit === 13 || unit === 9;"},
	} {
		// Not parallel: native.Build writes the shared adamic/runtime or adamic/units cache.
		t.Run(mutant.name, func(t *testing.T) {
			directory := t.TempDir()
			for _, file := range []string{"lexer.ts", "lex_main.ts"} {
				source, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				if file == "lexer.ts" {
					if strings.Count(string(source), mutant.from) != 1 {
						t.Fatal("mutation site must occur once")
					}
					source = []byte(strings.Replace(string(source), mutant.from, mutant.to, 1))
				}
				if err := os.WriteFile(filepath.Join(directory, file), source, 0644); err != nil {
					t.Fatal(err)
				}
			}
			nativeOut, _ := nativeLexer(t, directory, cases, false)
			node := run(t, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "lex_main.ts"), cases)
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
