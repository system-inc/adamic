// Package values holds the test of stage 1's third slice: cohere's port of postcss-values-parser,
// ported again to Adamic (the .ts files beside this one), compiled natively by stage 0 and run on Node,
// parsing the same values cohere's own tests parse, with Prettier's { loose: true } and the library's
// default { loose: false }, into the same trees, byte for byte, that Go cohere builds.
package values

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
var portFiles = []string{"tokenize.ts", "nodes.ts", "parser.ts", "values.ts", "main.ts"}

// generatedSeed fixes the generated values, so a failure names values anyone can make again. Setting
// COHERE_VALUES_SEED asks for others, and COHERE_VALUES_GENERATED for more or fewer of them.
const generatedSeed = 20261005

// run is one execution's observable behavior.
type run struct {
	stdout   []byte
	stderr   []byte
	exitCode int
}

// The port parses every value cohere's tests parse, and thousands generated from the pieces CSS values
// are made of, in both modes, into the trees Go cohere builds, byte for byte, natively, on Node and
// through the JavaScript backend; and the native port leaks nothing.
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

		// Cases that parse the same whatever the parser does would agree without testing it: each kind of
		// node and of refusal has to be among them.
		var missing []string
		for _, each := range []string{"\n    number ", "\n    func ", "\n    string ", "\n    comment ", "\n    atword ", "\n    unicode-range ", "\n    operator ", "\n    paren ", "\n    colon ", "\n    comma ", "column:NaN", "inline=true", "isColor=true",
			"\nerror ParserError: ", "\nerror TokenizeError: ", "\nerror TypeError: "} {
			if !strings.Contains(goAnswers, each) {
				missing = append(missing, each)
			}
		}
		if len(missing) > 0 {
			t.Errorf("the cases never reach %q", missing)
		}
		parses, refused := strings.Count(goAnswers, "\nroot "), strings.Count(goAnswers, "\nerror ")
		if agreed {
			t.Logf("%d parses (%d trees, %d refusals): every answer the same from Go cohere, the port natively, on Node and through the JavaScript backend",
				strings.Count(goAnswers, "\ncase ")+1, parses, refused)
		}
	})

	// The library both were written from, when there's a copy of it to run: the port must answer as
	// postcss-values-parser 2.0.1 itself does.
	t.Run("as postcss-values-parser", func(t *testing.T) {
		t.Parallel()
		library := os.Getenv("ADAMIC_VALUES_LIBRARY")
		if library == "" {
			t.Skip("set ADAMIC_VALUES_LIBRARY to a directory where `npm install postcss-values-parser@2.0.1` ran, to compare with the library itself")
		}
		script, err := filepath.Abs(filepath.Join("testdata", "library.mjs"))
		if err != nil {
			t.Fatal(err)
		}
		libraryRun := execute(t, nil, "node", script, library, casesPath)
		if libraryRun.exitCode != 0 || len(libraryRun.stderr) > 0 {
			t.Fatalf("the library: exit %d, stderr %q", libraryRun.exitCode, libraryRun.stderr)
		}
		portRun := onNode(t, filepath.Join(portSource, "main.ts"), casesPath)
		if difference := firstDifference(string(portRun.stdout), string(libraryRun.stdout)); difference != "" {
			t.Errorf("the port and the library differ: %s", strings.Replace(difference, "Go cohere", "the library", 1))
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
	// upstream's `next` starts undefined, so the first bracket or paren ends at column NaN; starting it
	// at 0 gives a number there.
	{
		name: "next starting at 0, not NaN",
		file: "tokenize.ts",
		from: "let next = NaN;",
		to:   "let next = 0;",
	},
	// An operator's sourceIndex is its token's end line, upstream's slip, not its index.
	{
		name: "an operator's sourceIndex its index",
		file: "parser.ts",
		from: "new Position(token.startLine, token.startColumn)), token.endLine);",
		to:   "new Position(token.startLine, token.startColumn)), token.index);",
	},
	// A space before a comma is compared with ',', which no token's kind is; comparing with 'comma'
	// puts the space after the word before it rather than before the comma.
	{
		name: "a space compared with the comma token's kind",
		file: "parser.ts",
		from: "nextKind === ',' ||",
		to:   "nextKind === 'comma' ||",
	},
	// A number's unit comes off its first occurrence, not its end: 1e1e is 11e and e, not 1e1 and e.
	{
		name: "a number's unit taken off its end",
		file: "parser.ts",
		from: "node = newNumber(replaceFirst(value, unit), nodeSource, nodeSourceIndex, unit);",
		to:   "node = newNumber(value.slice(0, value.length - unit.length), nodeSource, nodeSourceIndex, unit);",
	},
	// The driver writes DEL as itself, where Go cohere's side escapes it.
	{
		name: "DEL written unescaped",
		file: "main.ts",
		from: "if (unit < 0x20 || unit === 0x7f) {",
		to:   "if (unit < 0x20) {",
	},
	// The driver writes U+001F as itself, the last control character, where Go cohere's side escapes it.
	{
		name: "U+001F written unescaped",
		file: "main.ts",
		from: "if (unit < 0x20 || unit === 0x7f) {",
		to:   "if (unit < 0x1f || unit === 0x7f) {",
	},
	// Reviewer R's round six: a `#` before a 9 tokenized alone, so #999 is `#` and the number 999.
	{
		name: "alphaNum's digits stopping at 8",
		file: "tokenize.ts",
		from: "return (code >= 0x61 && code <= 0x7a) || (code >= 0x41 && code <= 0x5a) || (code >= 0x30 && code <= 0x39);",
		to:   "return (code >= 0x61 && code <= 0x7a) || (code >= 0x41 && code <= 0x5a) || (code >= 0x30 && code <= 0x38);",
	},
	// Reviewer R's round six: #999 not a color.
	{
		name: "isColor's digits stopping at 8",
		file: "parser.ts",
		from: "if (!((unit >= 0x30 && unit <= 0x39) || (unit >= 0x61 && unit <= 0x66) || (unit >= 0x41 && unit <= 0x46))) {",
		to:   "if (!((unit >= 0x30 && unit <= 0x38) || (unit >= 0x61 && unit <= 0x66) || (unit >= 0x41 && unit <= 0x46))) {",
	},
	// Reviewer R's round six: #z1 split into `#` and z1.
	{
		name: "alphaNum's lowercase stopping at y",
		file: "tokenize.ts",
		from: "return (code >= 0x61 && code <= 0x7a) || (code >= 0x41 && code <= 0x5a) || (code >= 0x30 && code <= 0x39);",
		to:   "return (code >= 0x61 && code <= 0x79) || (code >= 0x41 && code <= 0x5a) || (code >= 0x30 && code <= 0x39);",
	},
	// Reviewer R's round six: #Z1 split into `#` and Z1.
	{
		name: "alphaNum's uppercase stopping at Y",
		file: "tokenize.ts",
		from: "return (code >= 0x61 && code <= 0x7a) || (code >= 0x41 && code <= 0x5a) || (code >= 0x30 && code <= 0x39);",
		to:   "return (code >= 0x61 && code <= 0x7a) || (code >= 0x41 && code <= 0x59) || (code >= 0x30 && code <= 0x39);",
	},
	// Reviewer R's round six: #_a one word, where upstream makes `#` and _a.
	{
		name: "alphaNum taking _",
		file: "tokenize.ts",
		from: "return (code >= 0x61 && code <= 0x7a) || (code >= 0x41 && code <= 0x5a) || (code >= 0x30 && code <= 0x39);",
		to:   "return (code >= 0x61 && code <= 0x7a) || (code >= 0x41 && code <= 0x5a) || (code >= 0x30 && code <= 0x39) || code === 0x5f;",
	},
	// Reviewer R's round six: a=b two words.
	{
		name: "a word ending at =",
		file: "tokenize.ts",
		from: "function wordEnd(css: string, index: number): boolean {\n\tswitch (css.charCodeAt(index)) {\n",
		to:   "function wordEnd(css: string, index: number): boolean {\n\tswitch (css.charCodeAt(index)) {\n\t\tcase 0x3d:\n",
	},
	// Reviewer R's round six: CALC(1 -2) parsed, where upstream refuses it.
	{
		name: "the strict calc check ignoring case",
		file: "parser.ts",
		from: "this.#current.value === 'calc'",
		to:   "this.#current.value.toLowerCase() === 'calc'",
	},
	// Reviewer R's round six: URL(a b) one word in strict mode, where upstream makes two.
	{
		name: "the url argument check ignoring case",
		file: "parser.ts",
		from: "\t\t\tthis.#current.value === 'url' &&",
		to:   "\t\t\tthis.#current.value.toLowerCase() === 'url' &&",
	},
	// Reviewer R's round six: Url(//x) in loose mode parsed, where upstream reads // as a comment and refuses it.
	{
		name: "the tokenizer's url argument ignoring case",
		file: "tokenize.ts",
		from: "previous.value === 'url';",
		to:   "previous.value.toLowerCase() === 'url';",
	},
	// Reviewer R's round six: u+99 split into u+9 and 9.
	{
		name: "unicodeRange's digits stopping at 8",
		file: "tokenize.ts",
		from: "return (code >= 0x61 && code <= 0x66) || (code >= 0x41 && code <= 0x46) || (code >= 0x30 && code <= 0x39) || code === 0x3f || code === minus;",
		to:   "return (code >= 0x61 && code <= 0x66) || (code >= 0x41 && code <= 0x46) || (code >= 0x30 && code <= 0x38) || code === 0x3f || code === minus;",
	},
	// Reviewer R's round six: -- alone an operator, where upstream makes it a word.
	{
		name: "-- a word only before more text",
		file: "tokenize.ts",
		from: "if (code === minus && nextChar === minus) {",
		to:   "if (code === minus && nextChar === minus && pos + 2 < length) {",
	},
	// '?' is no longer part of a unicode range.
	{
		name: "? not in a unicode range",
		file: "tokenize.ts",
		from: "code === 0x3f ||",
		to:   "",
	},
}

// askedCases has cohere's side write every case, and Go cohere's answer to each, returning the cases
// file's path and the answers.
func askedCases(t *testing.T) (string, string) {
	t.Helper()
	seed := int64(generatedSeed)
	var err error
	if value := os.Getenv("COHERE_VALUES_SEED"); value != "" {
		if seed, err = strconv.ParseInt(value, 10, 64); err != nil {
			t.Fatal(err)
		}
	}
	generated := 4000
	if value := os.Getenv("COHERE_VALUES_GENERATED"); value != "" {
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
	if keep := os.Getenv("ADAMIC_VALUES_KEEP"); keep != "" {
		cases, err := os.ReadFile(casesPath)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(keep, cases, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("seed %d, %d generated values", seed, generated)
	return casesPath, string(answers)
}

// cohereSide runs testdata/cohere_side_test.go inside cohere's values package, by overlay, with a
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
	packageDirectory := filepath.Join(cohere, "internal", "format", "css", "values")
	replace := map[string]string{filepath.Join(packageDirectory, "adamic_port_side_test.go"): side}
	overlay, err := json.Marshal(map[string]any{"Replace": replace})
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0o644); err != nil {
		t.Fatal(err)
	}
	command := bounded(t, "go", "test", "-count=1", "-overlay="+overlayPath, "-run=^TestAdamicPortCases$", "./internal/format/css/values")
	command.Dir = cohere
	command.Env = append(os.Environ(), "ADAMIC_PORT_REQUEST="+requestPath)
	if output, err := combinedOutput(command); err != nil {
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

// bounded prepares a child; execute and combinedOutput run it with progress-based guards.
func bounded(t *testing.T, name string, arguments ...string) *exec.Cmd {
	t.Helper()
	// The child guard owns hang detection, including for Go test children.
	if name == "go" && len(arguments) > 0 && arguments[0] == "test" {
		arguments = append([]string{"test", "-timeout=0"}, arguments[1:]...)
	}
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
