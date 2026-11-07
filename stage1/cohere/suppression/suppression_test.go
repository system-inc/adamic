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
	"github.com/system-inc/adamic/stage1/cohere/suppression/oracles"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
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
func TestThePortParsesAsGoCohereDoes(test *testing.T) {
	test.Parallel()
	casesPath, goAnswers := askedCases(test)
	portSource := portDirectory(test, nil)
	program := lowered(test, filepath.Join(portSource, "main.ts"))

	test.Run("natively and on Node, as Go cohere", func(test *testing.T) {
		test.Parallel()
		nodeRun := onNode(test, filepath.Join(portSource, "main.ts"), casesPath)
		sanitized := filepath.Join(portSanitized(test, nil), "port")
		nativeRun := sanitizedRun(test, sanitized, casesPath)
		backendRun := onJavaScriptBackend(test, program, casesPath)
		for _, side := range []struct {
			name string
			run  run
		}{{"Node", nodeRun}, {"native", nativeRun}, {"the JavaScript backend", backendRun}} {
			if side.run.exitCode != 0 || len(side.run.stderr) > 0 {
				test.Fatalf("%s: exit %d, stderr %q", side.name, side.run.exitCode, side.run.stderr)
			}
			if difference := firstDifference(string(side.run.stdout), goAnswers); difference != "" {
				test.Errorf("%s and Go cohere differ: %s", side.name, difference)
			}
		}
		agreed := !test.Failed()
		if leaked := leaks(test, program, sanitized, casesPath); leaked != "" {
			test.Errorf("leaks:\n%s", leaked)
		}

		// Cases that answer the same whatever the index does would agree without testing it: each kind
		// of directive and answer has to be among them.
		var missing []string
		for _, each := range []string{
			" next-line rules=",
			" same-line rules=",
			" file rules=",
			" eslint-enable\n",
			" -\n",
			" 1\n",
			" 0\n",
			"\nunused 0",
			"\nwithout-reason 0",
			"query probe/reports-on-directives ",
		} {
			if !strings.Contains(goAnswers, each) {
				missing = append(missing, each)
			}
		}
		if len(missing) > 0 {
			test.Errorf("the cases never reach %q", missing)
		}
		if agreed {
			test.Logf(
				"%d sources, %d directives, %d questions (%d suppressed): every answer the same from Go cohere, the port natively, on Node and through the JavaScript backend",
				strings.Count(goAnswers, "case "),
				strings.Count(goAnswers, "\ndirective "),
				strings.Count(goAnswers, "\nquery ")+strings.Count(goAnswers, "\nlineof "),
				strings.Count(goAnswers, " 1\nquery")+strings.Count(goAnswers, " 1\nlineof"),
			)
		}
	})

	// The comparison has to be able to fail. Each mutant changes the port where only its answers can
	// show it, and the port must then disagree with Go cohere: natively, since that is the program this
	// test is about (the mutant's sanitized product, built ahead), and on Node, which shows the fault is the
	// port's and not stage 0's.
	for _, mutant := range mutants {
		test.Run("catches "+mutant.name, func(test *testing.T) {
			test.Parallel()
			mutated := portDirectory(test, &mutant)
			sanitized := filepath.Join(portSanitized(test, &mutant), "port")
			for _, side := range []struct {
				name string
				run  run
			}{
				{"natively", sanitizedRun(test, sanitized, casesPath)},
				{"on Node", onNode(test, filepath.Join(mutated, "main.ts"), casesPath)},
			} {
				if side.run.exitCode != 0 {
					test.Errorf(
						"%s the mutant exits %d (stderr %q); it must be caught by its answers, not by failing",
						side.name,
						side.run.exitCode,
						side.run.stderr,
					)
					continue
				}
				difference := firstDifference(string(side.run.stdout), goAnswers)
				if difference == "" {
					test.Errorf("%s the mutant agrees with Go cohere: the comparison cannot see it", side.name)
					continue
				}
				test.Logf("%s, caught: %s", side.name, difference)
			}
		})
	}
}

// mutant is one change to one port file.
type mutant struct {
	name, file, from, to string
}

var mutants = []mutant{
	// An edge of a class or a keyword's case: after `)` a slash divides; read as a pattern, `(a) / b // c` swallows the
	// comment.
	{
		name: "a ) opening a regular expression",
		file: "scan.ts",
		from: "        case '%':\n            return true;",
		to:   "        case '%':\n\t\tcase ')':\n            return true;",
	},
	// An edge of a class or a keyword's case: the scanner skips only space, tab, newline and carriage return; skipping
	// \v too makes `(\v/x/` a pattern.
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
		from: "if(body.startsWith(candidate)) {",
		to:   "if(body.toLowerCase().startsWith(candidate)) {",
	},
	// An edge of a class or a keyword's case: an unterminated string ends at its line's end, so the comments after it
	// are still found.
	{
		name: "a string running past its line",
		file: "scan.ts",
		from: "        if(character === '\\n') {\n            return index + 1;\n        }\n        index++;",
		to:   "        index++;",
	},
	// An edge of a class or a keyword's case: a `/` inside `[...]` doesn't end the pattern.
	{
		name: "a regular expression's class ignored",
		file: "scan.ts",
		from: "            case '[':\n                inClass = true;",
		to:   "            case '[':\n                inClass = false;",
	},
	// A file-scope disable in a line comment honored, which ESLint doesn't and cohere stopped doing on
	// Kirk's ruling of 2026-10-01: prose such as `// eslint-disable + generated banner` disabled a file.
	{
		name: "a // file-scope disable honored",
		file: "directives.ts",
		from: "if(scopeRead.scope === 'file' && isLineComment(commentText)) {",
		to:   "if(scopeRead.scope === 'file' && isLineComment(commentText) && false) {",
	},
	// Restore Go's whitespace set, which wrongly trims U+0085 and leaves U+FEFF.
	{
		name: "Go whitespace instead of JavaScript whitespace",
		file: "directives.ts",
		from: "        case 0xfeff:\n",
		to:   "        case 0x85:\n",
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
		from: "if(enable.line <= block.line) {",
		to:   "if(enable.line < block.line) {",
	},
	// The braces an interpolation opens are no longer counted, so `${ {a: 1}.a }` closes on the object's
	// brace and the rest of the template is read as code.
	{
		name: "an interpolation closing on an object's brace",
		file: "scan.ts",
		from: "            if(character === '{') {\n                depth++;",
		to:   "            if(character === '{') {\n                depth += 0;",
	},
}

// askedCases has cohere's side write every case, and Go cohere's answer to each, returning the cases
// file's path and the answers.
func askedCases(test *testing.T) (string, string) {
	test.Helper()
	seed := int64(generatedSeed)
	var parseError error
	if value := os.Getenv("COHERE_SUPPRESSION_SEED"); value != "" {
		if seed, parseError = strconv.ParseInt(value, 10, 64); parseError != nil {
			test.Fatal(parseError)
		}
	}
	generated := 2500
	if value := os.Getenv("COHERE_SUPPRESSION_GENERATED"); value != "" {
		if generated, parseError = strconv.Atoi(value); parseError != nil {
			test.Fatal(parseError)
		}
	}
	directory := test.TempDir()
	casesPath := filepath.Join(directory, "cases.txt")
	answersPath := filepath.Join(directory, "answers.txt")
	cohereSide(test, map[string]any{"seed": seed, "generated": generated, "cases": casesPath, "answers": answersPath})
	answers, readError := os.ReadFile(answersPath)
	if readError != nil {
		test.Fatal(readError)
	}
	if keep := os.Getenv("ADAMIC_SUPPRESSION_KEEP"); keep != "" {
		cases, readError := os.ReadFile(casesPath)
		if readError != nil {
			test.Fatal(readError)
		}
		if writeError := os.WriteFile(keep, cases, 0o644); writeError != nil {
			test.Fatal(writeError)
		}
	}
	test.Logf("seed %d, %d generated sources", seed, generated)
	return casesPath, string(answers)
}

// cohereSide runs testdata/cohere_side_test.go inside cohere's suppression package, by overlay, with a
// request. The package's test binary with the harness laid over it is a product (buildcache.GoTest), built ahead, and
// run here in the package's directory, where go test would run it.
func cohereSide(test *testing.T, request map[string]any) {
	test.Helper()
	directory := test.TempDir()
	requestPath := filepath.Join(directory, "request.json")
	encoded, encodeError := json.Marshal(request)
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	if writeError := os.WriteFile(requestPath, encoded, 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	binary, packageDirectory := cohereSideOracle(test)
	command := bounded(test, binary, "-test.run=^TestAdamicPortCases$")
	command.Dir = packageDirectory
	command.Env = append(os.Environ(), "ADAMIC_PORT_REQUEST="+requestPath)
	if output, combinedOutputError := childguard.CombinedOutput(
		command,
		childguard.Options{},
	); combinedOutputError != nil {
		test.Fatalf("cohere's side: %v\n%s", combinedOutputError, output)
	}
}

// portDirectory copies the port into a directory of its own, with a mutant applied when there is one.
func portDirectory(test *testing.T, applied *mutant) string {
	test.Helper()
	directory := test.TempDir()
	for _, name := range portFiles {
		contents, readError := os.ReadFile(name)
		if readError != nil {
			test.Fatal(readError)
		}
		source := string(contents)
		if applied != nil && applied.file == name {
			if strings.Count(source, applied.from) != 1 {
				test.Fatalf("the mutant %q must change exactly one place in %s", applied.name, name)
			}
			source = strings.Replace(source, applied.from, applied.to, 1)
		}
		if writeError := os.WriteFile(filepath.Join(directory, name), []byte(source), 0o644); writeError != nil {
			test.Fatal(writeError)
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

// bounded constructs an unstarted command; its callers use childguard to bound silence.
func bounded(test *testing.T, name string, arguments ...string) *exec.Cmd {
	test.Helper()
	return exec.Command(name, arguments...)
}

func execute(test *testing.T, environment []string, name string, arguments ...string) run {
	test.Helper()
	command := bounded(test, name, arguments...)
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	runError := childguard.Run(command, childguard.Options{})
	var exitError *exec.ExitError
	if runError != nil && !errors.As(runError, &exitError) {
		test.Fatalf("running %s: %v", name, runError)
	}
	return run{stdout: stdout.Bytes(), stderr: stderr.Bytes(), exitCode: command.ProcessState.ExitCode()}
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

// cohereSideOracle is the test binary go test would build with testdata/cohere_side_test.go laid over cohere's code, a
// product (buildcache.GoTest) built ahead, and the directory go test would run it in.
func cohereSideOracle(test testing.TB) (string, string) {
	test.Helper()
	return oracles.Suppression(test)
}

func TestProduct_SuppressionFindsGoOracle(test *testing.T) {
	test.Parallel()
	cohereSideOracle(test)
}
