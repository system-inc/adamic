// Package formatfiles holds the test of stage 1's fifth slice: cohere's format walk, which files a
// tree offers the formatter and why each other one was left, ported to Adamic (the .ts files beside this
// one, with the gitignore slice's matcher), compiled natively by stage 0 and run on Node, walking the
// same trees on disk that Go cohere walks and answering byte for byte as it does.
package formatfiles

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
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// repository is the repository's root, from this package's directory.
const repository = "../../.."

// portFiles are the port and its driver.
var portFiles = []string{"golang.ts", "disk.ts", "enumerate.ts", "main.ts"}

// gitignoreFiles are the gitignore slice's files the port imports, copied beside it as ../gitignore.
var gitignoreFiles = []string{"path.ts", "glob.ts", "gitignore.ts"}

// generatedSeed fixes the generated trees, so a failure names trees anyone can make again. Setting
// COHERE_FORMATFILES_SEED asks for others, and COHERE_FORMATFILES_GENERATED for more or fewer of them.
const generatedSeed = 20261005

// run is one execution's observable behavior.
type run struct {
	stdout   []byte
	stderr   []byte
	exitCode int
}

// The port walks every tree cohere's tests build, and hundreds generated at the edges of the walk's
// classes, and finds what Go cohere finds, file for file, count for count and error for error, and
// answers every question about repositories in them as Go cohere does: natively, on Node and through
// the JavaScript backend, byte for byte; and the native port leaks nothing.
func formatfilesCheckShard(t *testing.T, selected int) {
	t.Helper()
	if requested := formatfilesSelectedShard(t); requested >= 0 && requested != selected {
		t.Skip("another shard selected")
	}
	casesPath, goAnswers := formatfilesAskedCases(t)
	shards := formatfilesPartition(t, casesPath, goAnswers)
	portSource := portDirectory(t, nil)
	program := lowered(t, filepath.Join(portSource, "main.ts"))
	binary := formatfilesBinary(t, program, nil)
	checks := make([][]func(*testing.T), testThePortParsesAsGoCohereDoesShards)

	for index, part := range shards {
		checks[index] = append(checks[index], func(t *testing.T) {
			casesPath, goAnswers := formatfilesCaseFile(t, part), part.answers
			nodeRun := onNode(t, filepath.Join(portSource, "main.ts"), casesPath)
			nativeRun := execute(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary, casesPath)
			sanitized := binary
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
				t.Logf("%d trees, %d files offered, %d directories entered, %d refused walks: every answer the same from Go cohere, the port natively, on Node and through the JavaScript backend",
					strings.Count(goAnswers, "tree "), strings.Count(goAnswers, "\nfile "), strings.Count(goAnswers, "\ndirectory "), strings.Count(goAnswers, "\nenumerate error "))
			}
		})
	}

	// Cases that answer the same whatever the walk does would agree without testing it: each kind of
	// answer has to be among them.
	var missing []string
	for _, each := range []string{"\nenumerate error ", "\nnested-below error ", "\nnested \"", "\nnested-below \"", "\nsymbolic-links 1",
		"layer \"format.ignore\" 1", "layer \".gitignore\" 1", "\nlayer \"ignorePatterns\" 0", "\ndeclined \".i\"", "\ndeclined \".\u03b1\u03c3\"",
		"\ndeclined \".\"", "\ndeclined \"(none)\"", "\" 1\nhas", "\nfile \"", "\nadamic \""} {
		if !strings.Contains(goAnswers, each) {
			missing = append(missing, each)
		}
	}
	if len(missing) > 0 {
		t.Errorf("the cases never reach %q", missing)
	}

	// The comparison has to be able to fail. Each mutant changes the port where only its answers can
	// show it, and the port must then disagree with Go cohere: natively, since that is the program this
	// test holds, and on Node, which shows the fault is the port's and not stage 0's.
	for _, mutant := range mutants {
		index := formatfilesShard("mutant/" + mutant.name)
		if selected >= 0 && index != selected {
			continue
		}
		mutated := portDirectory(t, &mutant)
		mutatedProgram := lowered(t, filepath.Join(mutated, "main.ts"))
		mutatedBinary := formatfilesBinary(t, mutatedProgram, &mutant)
		checks[index] = append(checks[index], func(t *testing.T) {
			t.Logf("mutant %s", mutant.name)
			for _, side := range []struct {
				name string
				run  run
			}{{"natively", execute(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, mutatedBinary, casesPath)}, {"on Node", onNode(t, filepath.Join(mutated, "main.ts"), casesPath)}} {
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

	if len(checks) != testThePortParsesAsGoCohereDoesShards {
		t.Fatal("shard count differs")
	}
	for _, check := range checks[selected] {
		check(t)
	}

}

// mutant is one change to one port file.
type mutant struct {
	name, file, from, to string
}

var mutants = []mutant{
	{
		name: "Adamic files declined instead of held back",
		file: "enumerate.ts",
		from: "if (path.endsWith('.a') && base(path) !== '.a') {",
		to:   "if (path.endsWith('.a') && base(path) !== '.a' && false) {",
	},
	{
		name: "a bare .a name held back as Adamic",
		file: "enumerate.ts",
		from: "if (path.endsWith('.a') && base(path) !== '.a') {",
		to:   "if (path.endsWith('.a')) {",
	},
	{
		name: "uppercase .A held back as Adamic",
		file: "enumerate.ts",
		from: "if (path.endsWith('.a') && base(path) !== '.a') {",
		to:   "if (path.toLowerCase().endsWith('.a') && base(path) !== '.a') {",
	},
	// JavaScript's lowercasing for Go's: U+0130 becomes two characters, and Σ ending a word ς.
	{
		name: "an extension lowercased as JavaScript does",
		file: "enumerate.ts",
		from: "let extension = goToLower(ext(path));",
		to:   "let extension = ext(path).toLowerCase();",
	},
	// The declined extensions in UTF-16 order, where Go's map is printed in UTF-8's: U+FF41 and U+1F600
	// trade places.
	{
		name: "the declined extensions sorted by UTF-16 unit",
		file: "main.ts",
		from: "return [...counts.keys()].sort(compareCodePoints);",
		to:   "return [...counts.keys()].sort((left, right) => (left < right ? -1 : left > right ? 1 : 0));",
	},
	// A symbolic link round a loop taken for nothing, as fileStatus's failure says, where Lstat finds the
	// link: the port's own first mistake, which the cases found.
	{
		name: "a link round a loop taken for nothing",
		file: "disk.ts",
		from: "if (!status.message.endsWith(': permission denied')) {",
		to:   "if (status.message.endsWith(': no such file')) {",
	},
	// The quoting's fast path missing DEL, or a quote, which a file's name can hold.
	{
		name: "DEL passed by the quoting's fast path",
		file: "main.ts",
		from: "if (unit < 0x20 || unit === 0x7f || unit === 0x22 || unit === 0x5c) {",
		to:   "if (unit < 0x20 || unit === 0x22 || unit === 0x5c) {",
	},
	{
		name: "a quote passed by the quoting's fast path",
		file: "main.ts",
		from: "if (unit < 0x20 || unit === 0x7f || unit === 0x22 || unit === 0x5c) {",
		to:   "if (unit < 0x20 || unit === 0x7f || unit === 0x5c) {",
	},
	// A symbolic link to nothing taken for nothing, as fileStatus says, where Lstat finds the link.
	{
		name: "a link to nothing taken for nothing",
		file: "disk.ts",
		from: "if (listing.kind === 'Ok' && listing.names.includes(base(path))) {",
		to:   "if (listing.kind === 'Ok' && listing.names.includes(base(path)) && listing.names.length < 0) {",
	},
	// filepath.Ext of a name ending in a dot is the dot; dropping it declines o. as (none).
	{
		name: "a trailing dot no extension",
		file: "golang.ts",
		from: "\t\t\treturn path.slice(index);",
		to:   "\t\t\treturn index === path.length - 1 ? '' : path.slice(index);",
	},
	// HasOwnRepository follows a .git link, as os.Stat does; a link to nothing is no repository.
	{
		name: "a .git link to nothing taken for a repository",
		file: "enumerate.ts",
		from: "return statExists(join(directory, '.git'));",
		to:   "return lstat(join(directory, '.git')).exists;",
	},
	// A .prettierignore that is a link to nothing is refused, as os.Lstat finds it.
	{
		name: "a .prettierignore link to nothing let through",
		file: "enumerate.ts",
		from: "if (lstat(leftover).exists) {",
		to:   "if (statExists(leftover)) {",
	},
}

// askedCases has cohere's side write every case, and Go cohere's answer to each, returning the cases
// file's path and the answers.
func askedCases(t *testing.T) (string, string) {
	t.Helper()
	seed := int64(generatedSeed)
	var err error
	if value := os.Getenv("COHERE_FORMATFILES_SEED"); value != "" {
		if seed, err = strconv.ParseInt(value, 10, 64); err != nil {
			t.Fatal(err)
		}
	}
	generated := 400
	if value := os.Getenv("COHERE_FORMATFILES_GENERATED"); value != "" {
		if generated, err = strconv.Atoi(value); err != nil {
			t.Fatal(err)
		}
	}
	directory := t.TempDir()
	casesPath := filepath.Join(directory, "cases.txt")
	answersPath := filepath.Join(directory, "answers.txt")
	// The trees live here, where the port walks them after cohere's side has laid them out; the
	// directory is the test's own, so it outlives cohere's side and goes when the test ends.
	scratch := filepath.Join(directory, "trees")
	if err := os.Mkdir(scratch, 0o755); err != nil {
		t.Fatal(err)
	}
	cohereSide(t, map[string]any{"scratch": scratch, "seed": seed, "generated": generated, "cases": casesPath, "answers": answersPath})
	answers, err := os.ReadFile(answersPath)
	if err != nil {
		t.Fatal(err)
	}
	if keep := os.Getenv("ADAMIC_FORMATFILES_KEEP"); keep != "" {
		cases, err := os.ReadFile(casesPath)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(keep, cases, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("seed %d, %d generated trees", seed, generated)
	return casesPath, string(answers)
}

// cohereSide runs testdata/cohere_side_test.go inside cohere's formatfiles package, by overlay, with a
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
	packageDirectory := filepath.Join(cohere, "internal", "format", "formatfiles")
	replace := map[string]string{filepath.Join(packageDirectory, "adamic_port_side_test.go"): side}
	overlay, err := json.Marshal(map[string]any{"Replace": replace})
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0o644); err != nil {
		t.Fatal(err)
	}
	command := bounded(t, "go", "test", "-count=1", "-overlay="+overlayPath, "-run=^TestAdamicPortCases$", "./internal/format/formatfiles")
	command.Dir = cohere
	command.Env = append(os.Environ(), "ADAMIC_PORT_REQUEST="+requestPath)
	if output, err := combinedOutput(command); err != nil {
		t.Fatalf("cohere's side: %v\n%s", err, output)
	}
}

// portDirectory copies the port into a directory of its own, with a mutant applied when there is one,
// and the gitignore slice it imports beside it, as it sits in the repository.
func portDirectory(t *testing.T, applied *mutant) string {
	t.Helper()
	directory := t.TempDir()
	port := filepath.Join(directory, "formatfiles")
	gitignore := filepath.Join(directory, "gitignore")
	for _, made := range []string{port, gitignore} {
		if err := os.Mkdir(made, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range gitignoreFiles {
		contents, err := os.ReadFile(filepath.Join("..", "gitignore", name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(gitignore, name), contents, 0o644); err != nil {
			t.Fatal(err)
		}
	}
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
		if err := os.WriteFile(filepath.Join(port, name), []byte(source), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return port
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

// leaks returns a report of everything the finished port never let go of, or "": the leak check the
// oracle runs on every fixture (internal/leakcheck), on the sanitized binary and the port's C.
func leaks(t *testing.T, program *ir.Program, sanitized string, arguments ...string) string {
	t.Helper()
	return leakcheck.Report(t, native.C(program), sanitized, arguments...)
}
