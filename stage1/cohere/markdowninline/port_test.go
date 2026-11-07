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
func lowered(test *testing.T, path string) *ir.Program {
	test.Helper()
	program, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatalf("Load: %v", loadError)
	}
	result, lowerError := lower.Lower(context.Background(), program)
	if lowerError != nil {
		test.Fatalf("Lower: %v", lowerError)
	}
	return result
}

// bounded prepares a child without a flat deadline; childguard owns its process group. It returns the command's
// context with it, for childguard.Capture to end the reads by.
func bounded(test *testing.T, name string, arguments ...string) (*exec.Cmd, context.Context) {
	test.Helper()
	command, ctx, cancel := inlineCommand(name, arguments...)
	test.Cleanup(cancel)
	return command, ctx
}

// execute reads a run's output through childguard.Capture, to its end with no timer deciding when: exec's own copying
// under a five-second WaitDelay reported exec.ErrWaitDelay whenever the reads hadn't reached EOF five seconds after the
// child exited, so a Go, native, leaks or Node run that had finished could fail its shard under load (#n3gk22h's
// shape). Capture has no stall guard, and the runs lose nothing by it: childguard.Run's windows (two minutes of
// silence, thirty before the first byte) are longer than the 90 s Loom gives a test unit, which ends a hung run sooner.
func execute(test *testing.T, environment []string, name string, arguments ...string) run {
	test.Helper()
	command, ctx := bounded(test, name, arguments...)
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	stdout, stderr, captureError := childguard.Capture(ctx, command)
	var exitError *exec.ExitError
	if captureError != nil && !errors.As(captureError, &exitError) {
		test.Fatalf("running %s: %v", name, captureError)
	}
	return run{stdout: stdout, stderr: stderr, exitCode: command.ProcessState.ExitCode()}
}

// onNode runs a program's source on Node, through the oracle's runner, with its arguments.
func onNode(test *testing.T, path string, arguments ...string) run {
	test.Helper()
	runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle", "node.mjs"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	return execute(
		test,
		nil,
		"node",
		append([]string{"--disable-warning=ExperimentalWarning", runner, path}, arguments...)...,
	)
}

// onJavaScriptBackend runs the lowered port through the JavaScript backend, on Node.
func onJavaScriptBackend(test *testing.T, program *ir.Program, arguments ...string) run {
	test.Helper()
	path := filepath.Join(test.TempDir(), "program.mjs")
	if writeError := os.WriteFile(path, []byte(javascript.JavaScript(program)), 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	return onNode(test, path, arguments...)
}

// sanitizedRun runs a sanitized build with leak detection off, as the oracle does; leaks is its own run.
func sanitizedRun(test *testing.T, binary string, arguments ...string) run {
	test.Helper()
	var environment []string
	if runtime.GOOS == "linux" {
		environment = []string{"ASAN_OPTIONS=detect_leaks=0"}
	}
	return execute(test, environment, binary, arguments...)
}

// leaks returns a report of everything the finished port never let go of, or "": macOS's leaks tool on
// an unsanitized build, or LeakSanitizer on Linux running the sanitized binary again, as the oracle
// checks every fixture.
func leaks(test *testing.T, program *ir.Program, sanitized string, arguments ...string) string {
	test.Helper()
	switch runtime.GOOS {
	case "darwin":
		binary := filepath.Join(test.TempDir(), "port")
		if buildError := native.Build(native.C(program), binary, native.Options{}); buildError != nil {
			test.Fatal(buildError)
		}
		report := execute(test, nil, "leaks", append([]string{"--atExit", "--", binary}, arguments...)...)
		if report.exitCode == 0 {
			return ""
		}
		return string(report.stdout)
	case "linux":
		report := execute(test, []string{"ASAN_OPTIONS=detect_leaks=1"}, sanitized, arguments...)
		if report.exitCode == 0 {
			return ""
		}
		return fmt.Sprintf("exit %d\n%s", report.exitCode, report.stderr)
	}
	test.Fatalf("no leak check for %s", runtime.GOOS)
	return ""
}

type inlineCorpusData struct {
	paths, texts        []string
	files, generatedEnd int
	selected            map[string]bool
}

func collectInlineCorpus(test *testing.T) inlineCorpusData {
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	var texts []string
	patterns := []string{"*.md", "*.markdown", "*.mdown", "*.mkd"}
	paths := corpusfiles.Repository(test, root, []string{"."}, patterns)
	coherePaths := corpusfiles.Upstream(test, filepath.Join(root, "cohere"), corpusfiles.CohereCommit, []string{
		"CHANGELOG.md",
		"CONTRIBUTING.md",
		"README.md",
		"THIRD_PARTY_NOTICES.md",
		"TypeScript-shim",
		"editors",
		"internal",
		"schema",
		"swift",
	}, patterns)
	typeScriptPaths := corpusfiles.Upstream(
		test,
		filepath.Join(root, "cohere/TypeScript"),
		corpusfiles.TypeScriptGoCommit,
		[]string{
			".github",
			"CODE_OF_CONDUCT.md",
			"CONTRIBUTING.md",
			"README.md",
			"SECURITY.md",
			"SUPPORT.md",
			"packages",
			"tsc",
		},
		patterns,
	)
	if len(paths) == 0 {
		test.Fatal("repository Markdown corpus is empty")
	}
	if len(coherePaths) != 786 {
		test.Fatalf("cohere Markdown corpus at %s: %d files, want 786", corpusfiles.CohereCommit, len(coherePaths))
	}
	if len(typeScriptPaths) != 67 {
		test.Fatalf(
			"TypeScript-Go Markdown corpus at %s: %d files, want 67",
			corpusfiles.TypeScriptGoCommit,
			len(typeScriptPaths),
		)
	}
	paths = append(paths, coherePaths...)
	paths = append(paths, typeScriptPaths...)
	sort.Strings(paths)
	names := make([]string, len(paths))
	for i, path := range paths {
		name, relativePathError := filepath.Rel(root, path)
		if relativePathError != nil {
			test.Fatal(relativePathError)
		}
		names[i] = filepath.ToSlash(name)
	}
	selection, selectError := gatesample.Select(root, names, markdownInlineCorpusStride)
	if selectError != nil {
		test.Fatalf("%s: %v", test.Name(), selectError)
	}
	selected := map[string]bool{}
	if selection.Sample {
		test.Log(selection.Log(test.Name()))
		for _, name := range selection.Paths {
			selected[filepath.Join(root, filepath.FromSlash(name))] = true
		}
	} else {
		for _, path := range paths {
			selected[path] = true
		}
	}
	for _, path := range paths {
		data, readError := os.ReadFile(path)
		if readError != nil {
			test.Fatal(readError)
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
	texts = append(
		texts,
		"===",
		"---",
		"\r\t\u2028\u2029",
		"\x00",
		" \n ",
		" ` ``, ``` ",
		"a|b\n",
		"'\"< >)[]\\",
		"\t\n\t x",
		strings.Repeat("\\\\**a _b_ `c` |d| 😀 ", 10000),
	)
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
func write(test *testing.T, path string, b []byte) {
	test.Helper()
	if writeError := os.WriteFile(path, b, 0644); writeError != nil {
		test.Fatal(writeError)
	}
}
func clean(test *testing.T, name string, r run) {
	test.Helper()
	if r.exitCode != 0 || len(r.stderr) > 0 {
		test.Fatalf("%s exit %d stderr %s", name, r.exitCode, r.stderr)
	}
}
func equal(test *testing.T, name string, a, b []byte) {
	test.Helper()
	if !bytes.Equal(a, b) {
		i := 0
		for i < len(a) && i < len(b) && a[i] == b[i] {
			i++
		}
		test.Fatalf("%s first byte difference at %d (lengths %d/%d)", name, i, len(a), len(b))
	}
}

// A batch row is one mode of one text. Name its physical file or generated probe
// even when the throughput pass finds that Go disagrees with its own baseline.
func equalInlineBatch(test *testing.T, name string, got, want []byte, paths []string, modes int) {
	test.Helper()
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
	test.Fatalf(
		"%s first byte difference at %d (lengths %d/%d), case %d, file %s",
		name,
		offset,
		len(got),
		len(want),
		row,
		input,
	)
}
