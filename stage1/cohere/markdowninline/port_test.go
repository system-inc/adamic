package markdowninline

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/corpusfiles"
	"github.com/system-inc/adamic/internal/gatesample"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

const repository = "../../.."

// The cited Cloud sweep took 936s; ceil(936/30)=32. Generated probes stay full.
const markdownInlineCorpusStride = 32

// run is one execution's observable behavior.
type run struct {
	stdout   []byte
	stderr   []byte
	exitCode int
}

// lowered checks and lowers a program, failing the test with stage 0's refusal if it can't.
func lowered(t *testing.T, path string) *ir.Program {
	t.Helper()
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	result, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatalf("Lower: %v", err)
	}
	return result
}

// bounded prepares a child for the shared output-based hang guard.
// Silent builds and buffered children use its 30-minute first-output window.
// With ten CPU burners, the longest output gap was 26.3s; the default
// two-minute Stall leaves more than three times that gap as headroom.
func bounded(t *testing.T, name string, arguments ...string) *exec.Cmd {
	t.Helper()
	return exec.Command(name, arguments...)
}

func combinedOutput(command *exec.Cmd) ([]byte, error) {
	return childguard.CombinedOutput(command, childguard.Options{})
}

func execute(t *testing.T, environment []string, name string, arguments ...string) run {
	t.Helper()
	command := bounded(t, name, arguments...)
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := childguard.Run(command, childguard.Options{})
	var exitError *exec.ExitError
	if err != nil && !errors.As(err, &exitError) {
		t.Fatalf("running %s: %v", name, err)
	}
	return run{stdout: stdout.Bytes(), stderr: stderr.Bytes(), exitCode: command.ProcessState.ExitCode()}
}

// onNode runs a program's source on Node, through the oracle's runner, with its arguments.
func onNode(t *testing.T, path string, arguments ...string) run {
	t.Helper()
	runner, err := filepath.Abs(filepath.Join(repository, "oracle", "node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	return execute(t, nil, "node", append([]string{"--disable-warning=ExperimentalWarning", runner, path}, arguments...)...)
}

// onJavaScriptBackend runs the lowered port through the JavaScript backend, on Node.
func onJavaScriptBackend(t *testing.T, program *ir.Program, arguments ...string) run {
	t.Helper()
	path := filepath.Join(t.TempDir(), "program.mjs")
	if err := os.WriteFile(path, []byte(javascript.JavaScript(program)), 0o644); err != nil {
		t.Fatal(err)
	}
	return onNode(t, path, arguments...)
}

// nativelyRun is natively's run alone.
func nativelyRun(t *testing.T, program *ir.Program, arguments ...string) run {
	t.Helper()
	result, _ := natively(t, program, arguments...)
	return result
}

// natively builds the lowered port under the address and undefined-behavior sanitizers and runs it,
// returning the binary too, for the leak check. Leak detection is off here, as in the oracle; leaks is
// its own run.
func natively(t *testing.T, program *ir.Program, arguments ...string) (run, string) {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "port")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	var environment []string
	if runtime.GOOS == "linux" {
		environment = []string{"ASAN_OPTIONS=detect_leaks=0"}
	}
	return execute(t, environment, binary, arguments...), binary
}

// leaks returns a report of everything the finished port never let go of, or "": the leak check the
// oracle runs on every fixture (internal/leakcheck), on the sanitized binary and the port's C.
func leaks(t *testing.T, program *ir.Program, sanitized string, arguments ...string) string {
	t.Helper()
	return leakcheck.Report(t, native.C(program), sanitized, arguments...)
}

type inlineCorpusData struct {
	paths, texts        []string
	files, generatedEnd int
	selected            map[string]bool
}

func collectInlineCorpus(t *testing.T) inlineCorpusData {
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	var texts []string
	patterns := []string{"*.md", "*.markdown", "*.mdown", "*.mkd"}
	paths := corpusfiles.Repository(t, root, []string{"."}, patterns)
	coherePaths := corpusfiles.Upstream(t, filepath.Join(root, "cohere"), corpusfiles.CohereCommit, []string{"CHANGELOG.md", "CONTRIBUTING.md", "README.md", "THIRD_PARTY_NOTICES.md", "TypeScript-shim", "editors", "internal", "schema", "swift"}, patterns)
	typeScriptPaths := corpusfiles.Upstream(t, filepath.Join(root, "cohere/TypeScript"), corpusfiles.TypeScriptGoCommit, []string{".github", "CODE_OF_CONDUCT.md", "CONTRIBUTING.md", "README.md", "SECURITY.md", "SUPPORT.md", "packages", "tsc"}, patterns)
	if len(paths) == 0 {
		t.Fatal("repository Markdown corpus is empty")
	}
	if len(coherePaths) != 786 {
		t.Fatalf("cohere Markdown corpus at %s: %d files, want 786", corpusfiles.CohereCommit, len(coherePaths))
	}
	if len(typeScriptPaths) != 67 {
		t.Fatalf("TypeScript-Go Markdown corpus at %s: %d files, want 67", corpusfiles.TypeScriptGoCommit, len(typeScriptPaths))
	}
	paths = append(paths, coherePaths...)
	paths = append(paths, typeScriptPaths...)
	sort.Strings(paths)
	names := make([]string, len(paths))
	for i, path := range paths {
		name, err := filepath.Rel(root, path)
		if err != nil {
			t.Fatal(err)
		}
		names[i] = filepath.ToSlash(name)
	}
	selection, err := gatesample.Select(root, names, markdownInlineCorpusStride)
	if err != nil {
		t.Fatalf("%s: %v", t.Name(), err)
	}
	selected := map[string]bool{}
	if selection.Sample {
		t.Log(selection.Log(t.Name()))
		for _, name := range selection.Paths {
			selected[filepath.Join(root, filepath.FromSlash(name))] = true
		}
	} else {
		for _, path := range paths {
			selected[path] = true
		}
	}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		texts = append(texts, string(data))
	}
	files := len(texts)
	alphabet := []string{"a", "*", "_", "\\", "`", " ", "\n", "😀"}
	var generate func(string, int)
	generate = func(s string, n int) {
		texts = append(texts, s)
		if n > 0 {
			for _, c := range alphabet {
				generate(s+c, n-1)
			}
		}
	}
	generate("", 4)
	texts = append(texts, "===", "---", "\r\t\u2028\u2029", "\x00", " \n ", " ` ``, ``` ", "a|b\n", "'\"< >)[]\\", "\t\n\t x", strings.Repeat("\\\\**a _b_ `c` |d| 😀 ", 10000))
	generatedEnd := len(texts)
	// Every BMP scalar is examined on both sides of both delimiter kinds. Surrogate units are
	// exercised by astral scalars; isolated surrogates cannot be represented by the UTF-8 file API.
	var unicode strings.Builder
	for code := rune(0); code <= 0xffff; code++ {
		if code >= 0xd800 && code <= 0xdfff {
			continue
		}
		for _, marker := range []string{"*", "_"} {
			unicode.WriteRune(code)
			unicode.WriteString(marker + "a a" + marker)
			unicode.WriteRune(code)
			unicode.WriteString("; ")
		}
	}
	texts = append(texts, unicode.String())
	for _, code := range []rune{0x10100, 0x1039f, 0x1f600, 0x1e95f, 0x10ffff} {
		texts = append(texts, string(code)+"_a a_"+string(code), string(code)+"*a a*"+string(code))
	}
	return inlineCorpusData{paths, texts, files, generatedEnd, selected}
}
func write(t *testing.T, path string, b []byte) {
	t.Helper()
	if err := os.WriteFile(path, b, 0644); err != nil {
		t.Fatal(err)
	}
}
func clean(t *testing.T, name string, r run) {
	t.Helper()
	if r.exitCode != 0 || len(r.stderr) > 0 {
		t.Fatalf("%s exit %d stderr %s", name, r.exitCode, r.stderr)
	}
}
func equal(t *testing.T, name string, a, b []byte) {
	t.Helper()
	if !bytes.Equal(a, b) {
		i := 0
		for i < len(a) && i < len(b) && a[i] == b[i] {
			i++
		}
		t.Fatalf("%s first byte difference at %d (lengths %d/%d)", name, i, len(a), len(b))
	}
}

// A batch row is one mode of one text. Name its physical file or generated probe
// even when the throughput pass finds that Go disagrees with its own baseline.
func equalInlineBatch(t *testing.T, name string, got, want []byte, paths []string, modes int) {
	t.Helper()
	if bytes.Equal(got, want) {
		return
	}
	offset := 0
	for offset < len(got) && offset < len(want) && got[offset] == want[offset] {
		offset++
	}
	row := bytes.Count(want[:min(offset, len(want))], []byte("\n"))
	text := row / modes
	input := fmt.Sprintf("generated text %d", text-len(paths))
	if text < len(paths) {
		input = paths[text]
	}
	t.Fatalf("%s first byte difference at %d (lengths %d/%d), case %d, file %s", name, offset, len(got), len(want), row, input)
}
