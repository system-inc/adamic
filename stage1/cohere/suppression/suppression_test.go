// Package suppression holds the test of stage 1's fourth slice: cohere's suppression index, the
// eslint-disable comments and what they withheld, ported to Adamic (the .ts files beside this one),
// compiled natively by stage 0 and run on Node, building the index of the same sources cohere's own
// tests build and answering the same questions, in the same order, byte for byte as Go cohere does.
package suppression

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
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// repository is the repository's root, from this package's directory.
const repository = "../../.."

// portFiles are the port and its driver.
var portFiles = []string{"directives.ts", "scan.ts", "suppression.ts", "main.ts"}

// generatedSeed fixes the generated sources, so a failure names sources anyone can make again. Setting
// COHERE_SUPPRESSION_SEED asks for others, and COHERE_SUPPRESSION_GENERATED for more or fewer of them.
const generatedSeed = 20261005

// run is one execution's observable behavior.
type run struct {
	stdout   []byte
	stderr   []byte
	exitCode int
}

// The port finds the comments and directives Go cohere finds in every source cohere's tests build an
// index of, and in thousands generated, and answers every question about them as Go cohere does, in
// the same order, so the counts of what each directive withheld agree too: natively, on Node and
// through the JavaScript backend, byte for byte; and the native port leaks nothing.
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

		// Cases that answer the same whatever the index does would agree without testing it: each kind
		// of directive and answer has to be among them.
		var missing []string
		for _, each := range []string{" next-line rules=", " same-line rules=", " file rules=", " eslint-enable\n", " -\n", " 1\n", " 0\n",
			"\nunused 0", "\nwithout-reason 0", "query probe/reports-on-directives "} {
			if !strings.Contains(goAnswers, each) {
				missing = append(missing, each)
			}
		}
		if len(missing) > 0 {
			t.Errorf("the cases never reach %q", missing)
		}
		if agreed {
			t.Logf("%d sources, %d directives, %d questions (%d suppressed): every answer the same from Go cohere, the port natively, on Node and through the JavaScript backend",
				strings.Count(goAnswers, "case "), strings.Count(goAnswers, "\ndirective "), strings.Count(goAnswers, "\nquery ")+strings.Count(goAnswers, "\nlineof "),
				strings.Count(goAnswers, " 1\nquery")+strings.Count(goAnswers, " 1\nlineof"))
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
	// An edge of a class or a keyword's case: after `)` a slash divides; read as a pattern, `(a) / b // c` swallows the comment.
	{
		name: "a ) opening a regular expression",
		file: "scan.ts",
		from: "\t\tcase '%':\n\t\t\treturn true;",
		to:   "\t\tcase '%':\n\t\tcase ')':\n\t\t\treturn true;",
	},
	// An edge of a class or a keyword's case: the scanner skips only space, tab, newline and carriage return; skipping \v too makes `(\v/x/` a pattern.
	{
		name: "a vertical tab skipped as space",
		file: "scan.ts",
		from: "return character === ' ' || character === '\\t' || character === '\\n' || character === '\\r';",
		to:   "return character === ' ' || character === '\\t' || character === '\\n' || character === '\\r' || character === '\\v';",
	},
	// Restore the old boundary that wrongly rejected JavaScript whitespace after a directive word.
	{
		name: "only space or tab ending a directive word",
		file: "directives.ts",
		from: "return rest === '' || isJavaScriptSpace(rest.charCodeAt(0));",
		to:   "return rest === '' || rest.startsWith(' ') || rest.startsWith('\\t');",
	},
	// An edge of a class or a keyword's case: `ESLINT-DISABLE-LINE` is no directive.
	{
		name: "a directive word in any case",
		file: "directives.ts",
		from: "if (body.startsWith(candidate)) {",
		to:   "if (body.toLowerCase().startsWith(candidate)) {",
	},
	// An edge of a class or a keyword's case: an unterminated string ends at its line's end, so the comments after it are still found.
	{
		name: "a string running past its line",
		file: "scan.ts",
		from: "\t\tif (character === '\\n') {\n\t\t\treturn index + 1;\n\t\t}\n\t\tindex++;",
		to:   "\t\tindex++;",
	},
	// An edge of a class or a keyword's case: a `/` inside `[...]` doesn't end the pattern.
	{
		name: "a regular expression's class ignored",
		file: "scan.ts",
		from: "\t\t\tcase '[':\n\t\t\t\tinClass = true;",
		to:   "\t\t\tcase '[':\n\t\t\t\tinClass = false;",
	},
	// A file-scope disable in a line comment honored, which ESLint doesn't and cohere stopped doing on
	// Kirk's ruling of 2026-10-01: prose such as `// eslint-disable + generated banner` disabled a file.
	{
		name: "a // file-scope disable honored",
		file: "directives.ts",
		from: "if (scopeRead.scope === 'file' && isLineComment(commentText)) {",
		to:   "if (scopeRead.scope === 'file' && isLineComment(commentText) && false) {",
	},
	// Restore Go's whitespace set, which wrongly trims U+0085 and leaves U+FEFF.
	{
		name: "Go whitespace instead of JavaScript whitespace",
		file: "directives.ts",
		from: "\t\tcase 0xfeff:\n",
		to:   "\t\tcase 0x85:\n",
	},
	// A rule name matched by any suffix, not one on a `/`, so `no-enum` names `consistency-no-enum`.
	{
		name: "a suffix match off a / boundary",
		file: "suppression.ts",
		from: "return prefix.endsWith('/');",
		to:   "return true;",
	},
	// A directive that silences the finding about itself: a same-line directive covers a rule that
	// reports on directives.
	{
		name: "a directive covering the finding about itself",
		file: "suppression.ts",
		from: "return candidate.kind === 'file' && candidate.line < line;",
		to:   "return candidate.kind === 'file' ? candidate.line < line : true;",
	},
	// An enable on the block's own line closes it, so `/* eslint-disable */ /* eslint-enable */` covers
	// that line only.
	{
		name: "an enable on the block's own line closing it",
		file: "suppression.ts",
		from: "if (enable.line <= block.line) {",
		to:   "if (enable.line < block.line) {",
	},
	// The braces an interpolation opens are no longer counted, so `${ {a: 1}.a }` closes on the object's
	// brace and the rest of the template is read as code.
	{
		name: "an interpolation closing on an object's brace",
		file: "scan.ts",
		from: "\t\t\tif (character === '{') {\n\t\t\t\tdepth++;",
		to:   "\t\t\tif (character === '{') {\n\t\t\t\tdepth += 0;",
	},
}

// askedCases has cohere's side write every case, and Go cohere's answer to each, returning the cases
// file's path and the answers.
func askedCases(t *testing.T) (string, string) {
	t.Helper()
	seed := int64(generatedSeed)
	var err error
	if value := os.Getenv("COHERE_SUPPRESSION_SEED"); value != "" {
		if seed, err = strconv.ParseInt(value, 10, 64); err != nil {
			t.Fatal(err)
		}
	}
	generated := 2500
	if value := os.Getenv("COHERE_SUPPRESSION_GENERATED"); value != "" {
		if generated, err = strconv.Atoi(value); err != nil {
			t.Fatal(err)
		}
	}
	directory := t.TempDir()
	casesPath := filepath.Join(directory, "cases.txt")
	answersPath := filepath.Join(directory, "answers.txt")
	cohereSide(t, map[string]any{"seed": seed, "generated": generated, "cases": casesPath, "answers": answersPath})
	answers, err := os.ReadFile(answersPath)
	if err != nil {
		t.Fatal(err)
	}
	if keep := os.Getenv("ADAMIC_SUPPRESSION_KEEP"); keep != "" {
		cases, err := os.ReadFile(casesPath)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(keep, cases, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("seed %d, %d generated sources", seed, generated)
	return casesPath, string(answers)
}

// cohereSide runs testdata/cohere_side_test.go inside cohere's suppression package, by overlay, with a
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
	packageDirectory := filepath.Join(cohere, "internal", "lint", "suppression")
	replace := map[string]string{filepath.Join(packageDirectory, "adamic_port_side_test.go"): side}
	overlay, err := json.Marshal(map[string]any{"Replace": replace})
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0o644); err != nil {
		t.Fatal(err)
	}
	command := bounded(t, "go", "test", "-count=1", "-overlay="+overlayPath, "-run=^TestAdamicPortCases$", "./internal/lint/suppression")
	command.Dir = cohere
	command.Env = append(os.Environ(), "ADAMIC_PORT_REQUEST="+requestPath)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("cohere's side: %v\n%s", err, output)
	}
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
			return fmt.Sprintf("line %d: %q, Go cohere %q", index+1, gotLine, wantLine)
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

// leaks returns a report of everything the finished port never let go of, or "": the leak check the
// oracle runs on every fixture (internal/leakcheck), on the sanitized binary and the port's C.
func leaks(t *testing.T, program *ir.Program, sanitized string, arguments ...string) string {
	t.Helper()
	return leakcheck.Report(t, native.C(program), sanitized, arguments...)
}
