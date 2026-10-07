package selector

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// repository is the repository's root, from this package's directory.
const repository = "../../.."

// portFiles are the port and its driver.
var portFiles = []string{"source.ts", "nodes.ts", "tokenize.ts", "parser.ts", "main.ts"}

// generatedSeed fixes the generated selectors, so a failure names selectors anyone can make again.
// Setting COHERE_SELECTOR_SEED asks for others, and COHERE_SELECTOR_GENERATED for more or fewer of them.
const generatedSeed = 20261005

// run is one execution's observable behavior.
type run struct {
	stdout   []byte
	stderr   []byte
	exitCode int
}

// The port parses every text Go cohere's tests hold and thousands generated as Go cohere does: the same
// tree, field for field and offset for offset, every own property, or the same syntax error, natively,
// on Node and through the JavaScript backend, byte for byte; and the native port leaks nothing, on the
// texts it parses and on those it throws out of.
func TestThePortParsesAsGoCohereDoes(t *testing.T) {
	t.Parallel()
	casesPath, goAnswers := askedCases(t)
	portSource := portDirectory(t, nil)
	program := lowered(t, filepath.Join(portSource, "main.ts"))

	t.Run("natively and on Node, as Go cohere", func(t *testing.T) {
		t.Parallel()
		nodeRun := onNode(t, filepath.Join(portSource, "main.ts"), casesPath)
		nativeRun, sanitized := natively(t, program, casesPath)
		backendRun := onJavaScriptBackend(t, program, casesPath)
		for _, side := range []struct {
			name string
			run  run
		}{{"Node", nodeRun}, {"native", nativeRun}, {"the JavaScript backend", backendRun}} {
			if side.run.exitCode != 0 || len(side.run.stderr) > 0 {
				t.Fatalf("%s: exit %d, stderr %q", side.name, side.run.exitCode, side.run.stderr)
			}
			if difference := firstDifference(string(side.run.stdout), goAnswers); difference != "" {
				t.Errorf("%s and Go cohere differ: %s", side.name, difference)
			}
		}
		agreed := !t.Failed()
		if leaked := leaks(t, program, sanitized, casesPath); leaked != "" {
			t.Errorf("leaks:\n%s", leaked)
		}

		if agreed {
			t.Logf("%d selectors: every answer agrees on Node, native, and the JavaScript backend", strings.Count(goAnswers, "case "))
		}

	})

	// postcss-selector-parser 2.2.3 itself, except its explicitly reported nonterminating inputs.
	t.Run("as postcss-selector-parser", func(t *testing.T) {
		t.Parallel()
		library := os.Getenv("ADAMIC_SELECTOR_LIBRARY")
		if library == "" {
			t.Skip("set ADAMIC_SELECTOR_LIBRARY to a directory where `npm install postcss-selector-parser@2.2.3` ran, to compare with the library itself")
		}
		script, err := filepath.Abs(filepath.Join("testdata", "library.mjs"))
		if err != nil {
			t.Fatal(err)
		}
		answersPath := filepath.Join(t.TempDir(), "go-answers.txt")
		if err := os.WriteFile(answersPath, []byte(goAnswers), 0644); err != nil {
			t.Fatal(err)
		}
		libraryRun := execute(t, nil, "node", script, library, casesPath, answersPath)
		if libraryRun.exitCode != 0 {
			t.Fatalf("the library: exit %d, stderr %q", libraryRun.exitCode, libraryRun.stderr)
		}
		t.Logf("%s", libraryRun.stderr)
		if difference := firstDifference(string(libraryRun.stdout), goAnswers); difference != "" {
			t.Errorf("postcss-selector-parser and Go cohere differ: %s", strings.Replace(difference, "line", "postcss-selector-parser's line", 1))
		}
	})

	// The comparison has to be able to fail. Each mutant changes the port where only its answers can
	// show it, and the port must then disagree with Go cohere: natively, since that is the program this
	// test holds, and on Node, which shows the fault is the port's and not stage 0's.
	for _, mutant := range mutants {
		t.Run("catches "+mutant.name, func(t *testing.T) {
			t.Parallel()
			mutated := portDirectory(t, &mutant)
			mutatedProgram := lowered(t, filepath.Join(mutated, "main.ts"))
			for _, side := range []struct {
				name string
				run  run
			}{{"natively", nativelyRun(t, mutatedProgram, casesPath)}, {"on Node", onNode(t, filepath.Join(mutated, "main.ts"), casesPath)}} {
				if side.run.exitCode != 0 {
					t.Errorf("%s the mutant exits %d (stderr %q); it must be caught by its answers, not by failing", side.name, side.run.exitCode, side.run.stderr)
					continue
				}
				difference := firstDifference(string(side.run.stdout), goAnswers)
				if difference == "" {
					t.Errorf("%s the mutant agrees with Go cohere: the comparison cannot see it", side.name)
					continue
				}
				t.Logf("%s, caught: %s", side.name, difference)
			}
		})
	}
}

// mutant is one change to one port file.
type mutant struct {
	name, file, from, to string
}

var mutants = []mutant{
	{name: "the suffix operator loses |=", file: "parser.ts", from: "/([*~^$|]?=)([\\s\\S]*)/", to: "/([*~^$]?=)([\\s\\S]*)/"},
	{name: "a quote ignores escape parity", file: "tokenize.ts", from: "if(!escaped) {", to: "if(true) {"},
	{name: "a merged word uses its first token position", file: "parser.ts", from: "const token = this.curr();\n            if(partIndex === 0", to: "const token = starting;\n            if(partIndex === 0"},
}

// askedCases has cohere's side write every case, and Go cohere's answer to each, returning the cases
// file's path and the answers.
func askedCases(t *testing.T) (string, string) {
	t.Helper()
	seed := int64(generatedSeed)
	var err error
	if value := os.Getenv("COHERE_SELECTOR_SEED"); value != "" {
		if seed, err = strconv.ParseInt(value, 10, 64); err != nil {
			t.Fatal(err)
		}
	}
	generated := 4000
	if value := os.Getenv("COHERE_SELECTOR_GENERATED"); value != "" {
		if generated, err = strconv.Atoi(value); err != nil {
			t.Fatal(err)
		}
	}
	directory := t.TempDir()
	casesPath := filepath.Join(directory, "cases.txt")
	answersPath := filepath.Join(directory, "answers.txt")
	corpus := filepath.Join(directory, "corpus.txt")
	testTexts := filepath.Join(directory, "test-texts.json")
	request := map[string]any{"seed": seed, "generated": generated, "cases": casesPath, "answers": answersPath, "corpus": corpus, "testTexts": testTexts}
	cohereSide(t, request)
	library := os.Getenv("ADAMIC_SELECTOR_LIBRARY")
	if library != "" {
		script, _ := filepath.Abs(filepath.Join("testdata", "library.mjs"))
		root, _ := filepath.Abs(filepath.Join(repository, "cohere"))
		result := execute(t, nil, "node", script, library, "corpus", root, corpus, testTexts)
		if result.exitCode != 0 {
			t.Fatalf("corpus: %s", result.stderr)
		}
		t.Logf("%s", result.stdout)
		cohereSide(t, request)
	} else {
		t.Log("CSS corpus extraction skipped: set ADAMIC_SELECTOR_LIBRARY")
	}
	answers, err := os.ReadFile(answersPath)
	if err != nil {
		t.Fatal(err)
	}
	if keep := os.Getenv("ADAMIC_SELECTOR_KEEP"); keep != "" {
		cases, err := os.ReadFile(casesPath)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(keep, cases, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("seed %d, %d generated selectors", seed, generated)
	return casesPath, string(answers)
}

// cohereSide runs testdata/cohere_side_test.go inside cohere's selector package, by overlay, with a
// request.
func cohereSide(t *testing.T, request map[string]any) {
	t.Helper()
	directory := t.TempDir()
	requestPath := filepath.Join(directory, "request.json")
	encoded, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(requestPath, encoded, 0o644); err != nil {
		t.Fatal(err)
	}
	cohere, err := filepath.Abs(filepath.Join(repository, "cohere"))
	if err != nil {
		t.Fatal(err)
	}
	side, err := filepath.Abs(filepath.Join("testdata", "cohere_side_test.go"))
	if err != nil {
		t.Fatal(err)
	}
	packageDirectory := filepath.Join(cohere, "internal", "format", "css", "selector")
	replace := map[string]string{filepath.Join(packageDirectory, "adamic_port_side_test.go"): side}
	overlay, err := json.Marshal(map[string]any{"Replace": replace})
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0o644); err != nil {
		t.Fatal(err)
	}
	command := bounded(t, "go", "test", "-v", "-count=1", "-overlay="+overlayPath, "-run=^TestAdamicPortCases$", "./internal/format/css/selector")
	command.Dir = cohere
	command.Env = append(os.Environ(), "ADAMIC_PORT_REQUEST="+requestPath)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("cohere's side: %v\n%s", err, output)
	}
	t.Logf("Go oracle: %s", output)
}

// portDirectory copies the port into a directory of its own, with a mutant applied when there is one.
func portDirectory(t *testing.T, applied *mutant) string {
	t.Helper()
	directory := t.TempDir()
	for _, name := range portFiles {
		contents, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		source := string(contents)
		if applied != nil && applied.file == name {
			if strings.Count(source, applied.from) != 1 {
				t.Fatalf("the mutant %q must change exactly one place in %s", applied.name, name)
			}
			source = strings.Replace(source, applied.from, applied.to, 1)
		}
		if err := os.WriteFile(filepath.Join(directory, name), []byte(source), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return directory
}

// firstDifference says where two outputs first differ, line by line, or "" when they don't.
func firstDifference(got string, want string) string {
	if got == want {
		return ""
	}
	gotLines, wantLines := strings.Split(got, "\n"), strings.Split(want, "\n")
	for index := 0; index < len(gotLines) || index < len(wantLines); index++ {
		var gotLine, wantLine string
		if index < len(gotLines) {
			gotLine = gotLines[index]
		}
		if index < len(wantLines) {
			wantLine = wantLines[index]
		}
		if gotLine != wantLine {
			prefix := 0
			for prefix < len(gotLine) && prefix < len(wantLine) && gotLine[prefix] == wantLine[prefix] {
				prefix++
			}
			start := max(0, prefix-40)
			return fmt.Sprintf("line %d, byte %d: %q, Go cohere %q", index+1, prefix+1,
				gotLine[start:min(len(gotLine), prefix+120)], wantLine[start:min(len(wantLine), prefix+120)])
		}
	}
	return "the same lines, not the same bytes"
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

// bounded is a command that can't outlive its test: it has a deadline, it runs in a process group of
// its own, and when the deadline passes or the test ends, the whole group is killed.
func bounded(t *testing.T, name string, arguments ...string) *exec.Cmd {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	t.Cleanup(cancel)
	command := exec.CommandContext(ctx, name, arguments...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		return syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
	}
	command.WaitDelay = 5 * time.Second
	return command
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
	err := command.Run()
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

// leaks returns a report of everything the finished port never let go of, or "": macOS's leaks tool on
// an unsanitized build, or LeakSanitizer on Linux running the sanitized binary again, as the oracle
// checks every fixture.
func leaks(t *testing.T, program *ir.Program, sanitized string, arguments ...string) string {
	t.Helper()
	switch runtime.GOOS {
	case "darwin":
		binary := filepath.Join(t.TempDir(), "port")
		if err := native.Build(native.C(program), binary, native.Options{}); err != nil {
			t.Fatal(err)
		}
		report := execute(t, nil, "leaks", append([]string{"--atExit", "--", binary}, arguments...)...)
		if report.exitCode == 0 {
			return ""
		}
		return string(report.stdout)
	case "linux":
		report := execute(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, sanitized, arguments...)
		if report.exitCode == 0 {
			return ""
		}
		return fmt.Sprintf("exit %d\n%s", report.exitCode, report.stderr)
	}
	t.Fatalf("no leak check for %s", runtime.GOOS)
	return ""
}

// Not parallel: throughput is measured alone, not while other ports compile.
func TestSelectorThroughput(t *testing.T) {
	if os.Getenv("ADAMIC_SELECTOR_BENCH") == "" {
		t.Skip("set ADAMIC_SELECTOR_BENCH=1 to time parser throughput")
	}
	casesPath, answers := askedCases(t)
	program := lowered(t, filepath.Join(portDirectory(t, nil), "main.ts"))
	binary := filepath.Join(t.TempDir(), "selector")
	if err := native.Build(native.C(program), binary, native.Options{}); err != nil {
		t.Fatal(err)
	}
	count := strings.Count(answers, "case ") * 10
	answersPath := filepath.Join(t.TempDir(), "answers.txt")
	if err := os.WriteFile(answersPath, []byte(answers), 0644); err != nil {
		t.Fatal(err)
	}
	script, _ := filepath.Abs(filepath.Join("testdata", "library.mjs"))
	library := os.Getenv("ADAMIC_SELECTOR_LIBRARY")
	var expected string
	for round := 0; round < 3; round++ {
		sides := []struct {
			name, command string
			args          []string
		}{
			{"native", binary, []string{casesPath, "count", "repeat"}},
			{"Node source", "node", []string{"--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle", "node.mjs"), "main.ts", casesPath, "count", "repeat"}},
		}
		if library != "" {
			sides = append(sides, struct {
				name, command string
				args          []string
			}{"postcss-selector-parser Node", "node", []string{script, library, "count", casesPath, answersPath}})
		}
		for _, side := range sides {
			start := time.Now()
			result := execute(t, nil, side.command, side.args...)
			elapsed := time.Since(start)
			if result.exitCode != 0 || len(result.stderr) != 0 {
				t.Fatalf("%s: exit %d, %s", side.name, result.exitCode, result.stderr)
			}
			if expected == "" {
				expected = string(result.stdout)
			}
			if string(result.stdout) != expected {
				t.Fatalf("throughput checksums differ: %q and %q", expected, result.stdout)
			}
			t.Logf("%s round %d: %d selectors in %s, %.0f selectors/s; %s", side.name, round+1, count, elapsed, float64(count)/elapsed.Seconds(), result.stdout)
		}
	}
}

// The vendored error fixtures may fail CSS parsing, but a parseable file must still contribute
// every selector, including one inside _errors_. Unexpected malformed files and exclusions that
// become parseable remain corpus failures.
func TestCorpusKeepsEveryParseableFile(t *testing.T) {
	t.Parallel()
	library := os.Getenv("ADAMIC_SELECTOR_LIBRARY")
	if library == "" {
		t.Skip("set ADAMIC_SELECTOR_LIBRARY to check CSS corpus loading")
	}
	root := t.TempDir()
	excluded := "internal/format/css/testdata/prettier/css/_errors_/less-syntax.css"
	for name, source := range map[string]string{
		"valid.css":          ".required { color: red; }",
		"_errors_/valid.css": ".also-required { color: blue; }",
		excluded:             "a {.bordered();}",
	} {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
	}
	script, err := filepath.Abs(filepath.Join("testdata", "library.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	corpus := filepath.Join(t.TempDir(), "corpus.txt")
	result := execute(t, nil, "node", script, library, "corpus", root, corpus)
	if result.exitCode != 0 {
		t.Fatalf("corpus: %s", result.stderr)
	}
	selectors, err := os.ReadFile(corpus)
	if err != nil {
		t.Fatal(err)
	}
	if string(selectors) != "\".also-required\"\n\".required\"\n" {
		t.Fatalf("parseable files lost selectors: %q", selectors)
	}
	if !strings.Contains(string(result.stdout), "3 CSS files, 1 expected CSS errors, 2 selectors") {
		t.Fatalf("expected CSS error was not counted: %s", result.stdout)
	}
	if err := os.WriteFile(filepath.Join(root, excluded), []byte(".newly-parseable {}"), 0644); err != nil {
		t.Fatal(err)
	}
	result = execute(t, nil, "node", script, library, "corpus", root, corpus)
	if result.exitCode == 0 || !strings.Contains(string(result.stderr), "expected CSS error became parseable") {
		t.Fatalf("parseable exclusion was omitted: exit %d, %s", result.exitCode, result.stderr)
	}
	if err := os.WriteFile(filepath.Join(root, excluded), []byte("a {.bordered();}"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "unexpected.css"), []byte("a {.bordered();}"), 0644); err != nil {
		t.Fatal(err)
	}
	result = execute(t, nil, "node", script, library, "corpus", root, corpus)
	if result.exitCode == 0 || !strings.Contains(string(result.stderr), "unexpected.css") {
		t.Fatalf("unexpected CSS error was omitted: exit %d, %s", result.exitCode, result.stderr)
	}
}

// The explicit exception to JS agreement is observable, not an inference from Go.
func TestTheLibraryDoesNotReturnOnUnconsumedNamespaceBars(t *testing.T) {
	t.Parallel()
	library := os.Getenv("ADAMIC_SELECTOR_LIBRARY")
	if library == "" {
		t.Skip("set ADAMIC_SELECTOR_LIBRARY to prove upstream nontermination")
	}
	for _, text := range []string{"a| b", ":is(a|)", "a|@x", "*|)"} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, "node", "testdata/nontermination.mjs", library, text)
			output, err := command.CombinedOutput()
			if ctx.Err() != context.DeadlineExceeded || err == nil || string(output) != "entering parser\n" {
				t.Fatalf("expected the entered parser still running at deadline, got %v, context %v, output %q", err, ctx.Err(), output)
			}
			t.Log("entered parser, no return within 1s, killed at deadline")
		})
	}
}
