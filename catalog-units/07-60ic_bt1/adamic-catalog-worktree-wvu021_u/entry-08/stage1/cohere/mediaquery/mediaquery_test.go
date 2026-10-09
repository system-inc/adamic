// Package mediaquery holds the test of stage 1's second slice: cohere's port of
// postcss-media-query-parser, ported again to Adamic (the .ts files beside this one), compiled natively
// by stage 0 and run on Node, parsing the same params cohere's own tests parse into the same trees, byte
// for byte, that Go cohere builds.
package mediaquery

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
var portFiles = []string{"whitespace.ts", "nodes.ts", "parsers.ts", "index.ts", "main.ts"}

// generatedSeed fixes the generated params, so a failure names params anyone can make again. Setting
// COHERE_MEDIAQUERY_SEED asks for others, and COHERE_MEDIAQUERY_GENERATED for more or fewer of them.
const generatedSeed = 20261005

// run is one execution's observable behavior.
type run struct {
	stdout   []byte
	stderr   []byte
	exitCode int
}

// The port parses every params string cohere's tests parse, and thousands generated from the pieces
// media queries are made of, into the trees Go cohere builds, byte for byte, natively, on Node and
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

		// Cases that parse the same whatever the parser does would agree without testing it.
		parsed, refused, undecided := strings.Count(goAnswers, "\nmedia-query-list "), strings.Count(goAnswers, "\nerror "), strings.Count(goAnswers, "<undefined>")
		if parsed == 0 || refused == 0 || undecided == 0 {
			t.Errorf("the cases ask too little: %d parsed, %d refused, %d nodes of undecided type", parsed, refused, undecided)
		}
		if agreed {
			t.Logf("%d cases (%d parsed, %d refused; %d nodes left of undecided type): every tree the same from Go cohere, the port natively, on Node and through the JavaScript backend",
				strings.Count(goAnswers, "case "), parsed, refused, undecided)
		}
	})

	// The library both were written from, when there's a copy of it to run: the port must answer as
	// postcss-media-query-parser 0.2.3 itself does, which holds the fix to Go cohere too.
	t.Run("as postcss-media-query-parser", func(t *testing.T) {
		t.Parallel()
		library := os.Getenv("ADAMIC_MEDIA_QUERY_LIBRARY")
		if library == "" {
			t.Skip("set ADAMIC_MEDIA_QUERY_LIBRARY to a directory where `npm install postcss-media-query-parser@0.2.3` ran, to compare with the library itself")
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
	// The colon's value starts one unit early: every value node's sourceIndex is off by one.
	{
		name: "a feature value's sourceIndex not counting the colon",
		file: "parsers.ts",
		from: "mediaFeatureValueBefore.length + i + 1 + indexLocal,",
		to:   "mediaFeatureValueBefore.length + i + indexLocal,",
	},
	// A quote escaped with a backslash closes the string anyway, so the colon inside it splits the
	// feature.
	{
		name: "an escaped quote closing a string",
		file: "parsers.ts",
		from: "(i === 0 || stringNormalized[i - 1] !== '\\\\')",
		to:   "(i === 0 || stringNormalized[i - 1] !== '\\\\' || true)",
	},
	// Reviewer R's round six: a closing quote ends a string whatever quote opened it, so in
	// ({'a"b'}:x) the double quote closes the single-quoted string and the colon after it splits nothing.
	{
		name: "any closing quote ending a string",
		file: "parsers.ts",
		from: "current.mode === 'string' && current.character === character &&",
		to:   "current.mode === 'string' &&",
	},
	// What follows an expression is typed a media type, where the library says only `and` can.
	{
		name: "the word after an expression typed as a media type",
		file: "parsers.ts",
		from: "if (previous.type === 'media-feature-expression') {\n\t\t\t\t\tcurrent.type = 'keyword';",
		to:   "if (previous.type === 'media-feature-expression') {\n\t\t\t\t\tcurrent.type = 'media-type';",
	},
	// A byte order mark is no longer whitespace, which JavaScript's \s says it is (and Go's
	// unicode.IsSpace says it isn't).
	{
		name: "U+FEFF not whitespace",
		file: "whitespace.ts",
		from: "\t\tcase 0xfeff:\n",
		to:   "",
	},
	// U+2029, the paragraph separator, is no longer whitespace. Reviewer R2 found it survived while the
	// generated pieces held U+2028 but not U+2029.
	{
		name: "U+2029 not whitespace",
		file: "whitespace.ts",
		from: "\t\tcase 0x2029:\n",
		to:   "",
	},
	// U+2000, the en quad, is no longer whitespace: the range starts one late. It survived while the
	// pieces held only U+2003 and U+200A of the range.
	{
		name: "U+2000 not whitespace",
		file: "whitespace.ts",
		from: "return unit >= 0x2000 && unit <= 0x200a;",
		to:   "return unit >= 0x2001 && unit <= 0x200a;",
	},
	// The driver writes DEL as itself, where Go cohere's side escapes it. It survived while no piece
	// held DEL.
	{
		name: "DEL written unescaped",
		file: "main.ts",
		from: "if (unit < 0x20 || unit === 0x7f) {",
		to:   "if (unit < 0x20) {",
	},
	// The driver writes U+001F as itself, the last control character, where Go cohere's side escapes
	// it. It survived while no piece held U+001F.
	{
		name: "U+001F written unescaped",
		file: "main.ts",
		from: "if (unit < 0x20 || unit === 0x7f) {",
		to:   "if (unit < 0x1f || unit === 0x7f) {",
	},
	// trim keeps trailing whitespace.
	{
		name: "trim as trimStart",
		file: "whitespace.ts",
		from: "return text.trim();",
		to:   "return text.trimStart();",
	},
}

// askedCases has cohere's side write every case and Go cohere's answer to each.
// cohere 61ed9f4a fixed non-ASCII feature names, so no source fix is overlaid.
func askedCases(t *testing.T) (string, string) {
	t.Helper()
	seed := int64(generatedSeed)
	var err error
	if value := os.Getenv("COHERE_MEDIAQUERY_SEED"); value != "" {
		if seed, err = strconv.ParseInt(value, 10, 64); err != nil {
			t.Fatal(err)
		}
	}
	generated := 4000
	if value := os.Getenv("COHERE_MEDIAQUERY_GENERATED"); value != "" {
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
	cases, err := os.ReadFile(casesPath)
	if err != nil {
		t.Fatal(err)
	}
	if keep := os.Getenv("ADAMIC_MEDIAQUERY_KEEP"); keep != "" {
		if err := os.WriteFile(keep, cases, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("seed %d, %d generated params", seed, generated)
	return casesPath, string(answers)
}

// cohereSide runs testdata/cohere_side_test.go inside cohere's mediaquery package,
// by overlay, with a request. Cohere's parser source is used unchanged.
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
	packageDirectory := filepath.Join(cohere, "internal", "format", "css", "mediaquery")
	replace := map[string]string{filepath.Join(packageDirectory, "adamic_port_side_test.go"): side}
	overlay, err := json.Marshal(map[string]any{"Replace": replace})
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0o644); err != nil {
		t.Fatal(err)
	}
	command := bounded(t, "go", "test", "-timeout=0", "-count=1", "-overlay="+overlayPath, "-run=^TestAdamicPortCases$", "./internal/format/css/mediaquery")
	command.Dir = cohere
	command.Env = append(os.Environ(), "ADAMIC_PORT_REQUEST="+requestPath)
	if output, err := childguard.CombinedOutput(command, childguard.Options{}); err != nil {
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

// bounded configures a child command; Run and CombinedOutput below guard its output progress.
// Silent builds and buffered children use childguard's 30-minute FirstOutput window;
// after output starts, the default two-minute Stall allows over 3x the measured loaded gaps.
func bounded(t *testing.T, name string, arguments ...string) *exec.Cmd {
	t.Helper()
	command := exec.Command(name, arguments...)
	t.Cleanup(func() {
		if command.Process != nil && command.ProcessState == nil {
			_ = syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		}
	})
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
