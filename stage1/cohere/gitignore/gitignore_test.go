// Package gitignore holds the test of stage 1's first slice: cohere's gitignore matcher, ported to
// Adamic (the .ts files beside this one), compiled natively by stage 0 and run on Node, answering the
// same questions cohere's own tests ask with the same bytes Go cohere and git answer with.
package gitignore

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
	"unicode/utf8"

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
var portFiles = []string{"path.ts", "glob.ts", "gitignore.ts", "case.ts", "main.ts"}

// generatedSeed fixes the generated trees, so a failure names trees anyone can make again. Setting
// COHERE_GITIGNORE_SEED, as for cohere's own differential, asks for others.
const generatedSeed = 20261005

// cases is everything the port is asked, as cohere_side_test.go describes it.
type cases struct {
	Trees    []tree     `json:"trees"`
	Patterns []patterns `json:"patterns"`
	Globs    []glob     `json:"globs"`
}

type patterns struct {
	Name    string   `json:"name"`
	Lines   []string `json:"lines"`
	Queries []query  `json:"queries"`
}

type tree struct {
	Name     string   `json:"name"`
	Disk     string   `json:"disk"`
	Root     string   `json:"root"`
	Entries  []entry  `json:"entries"`
	Queries  []query  `json:"queries"`
	Git      string   `json:"git"`
	Expected []string `json:"expected"`
}

type entry struct {
	Path     string `json:"path"`
	Kind     string `json:"kind"`
	Contents string `json:"contents"`
	Target   string `json:"target"`
}

type query struct {
	Path        string `json:"path"`
	IsDirectory bool   `json:"isDirectory"`
}

type glob struct {
	Pattern  string `json:"pattern"`
	Text     string `json:"text"`
	Path     bool   `json:"path"`
	Expected string `json:"expected"`
}

// run is one execution's observable behavior.
type run struct {
	stdout   []byte
	stderr   []byte
	exitCode int
}

// The port answers every question cohere's tests ask, byte for byte as Go cohere does, natively and on
// Node; and where git has an answer, git check-ignore run on the tree or git's own test script, the
// port's is git's.
//
// The questions: cohere's documented semantics, its refusals, generated trees from cohere's own
// generator (COHERE_GITIGNORE_GENERATED of them, 40 by default), this repository and cohere's checkout
// as real trees, each asked path by path and again as cohere's walk asks it, pattern lists, and, when
// COHERE_GIT_SOURCE names a git source checkout as it does for cohere, git's t0008 check-ignore corpus
// and t3070 wildmatch corpus.
func TestThePortAnswersAsGoCohereAndGitDoUnion(t *testing.T) {
	t.Parallel()
	checkPortAnswerUnion(t)
}

// agreement is how the port's answers stand against git's.
type agreement struct {
	compared, ignored, reincluded, scripted, globs int
	differences                                    []string
}

// againstGit holds the port's output to git's answers: git check-ignore's, run on each tree that
// gitAnswers holds, the answers git's t0008 states for its fixture, and the answers t3070 and cohere's
// documented rules state for globs.
func againstGit(asked cases, gitAnswers map[string][]string, output string) agreement {
	var result agreement
	port := answersByTree(output)
	for _, tree := range asked.Trees {
		var gits []string
		switch tree.Git {
		case "run":
			gits = gitAnswers[tree.Name]
		case "script":
			gits = tree.Expected
			result.scripted += len(gits)
		default:
			continue
		}
		ours := port[tree.Name]
		if len(ours) != len(gits) {
			result.differences = append(result.differences, fmt.Sprintf("%s: the port answered %d paths, git %d", tree.Name, len(ours), len(gits)))
			continue
		}
		for index := range ours {
			result.compared++
			if strings.HasPrefix(ours[index], "1 ") {
				result.ignored++
			} else if !strings.HasPrefix(ours[index], "0 ::") {
				result.reincluded++
			}
			if ours[index] != gits[index] {
				result.differences = append(result.differences, fmt.Sprintf("%s: the port says %q, git %q", tree.Name, ours[index], gits[index]))
			}
		}
	}
	globAnswers := port["globs"]
	for index, glob := range asked.Globs {
		if glob.Expected == "" {
			continue
		}
		result.globs++
		want := fmt.Sprintf("glob %d %s", index, glob.Expected)
		if index >= len(globAnswers) || globAnswers[index] != want {
			result.differences = append(result.differences, fmt.Sprintf("glob %q against %q (path %v): git expects %s", glob.Pattern, glob.Text, glob.Path, glob.Expected))
		}
	}
	return result
}

// mutant is one change to one port file.
type mutant struct {
	name, file, from, to string

	// seenByGit says whether git has a word on what the mutant breaks, which it does for trees and not
	// for a pattern list's refusal.
	seenByGit bool

	// needsLargest is a mutant only the 100 MiB ignore file can show, asked about only when
	// ADAMIC_GITIGNORE_LARGEST is set (largest).
	needsLargest bool
}

var mutants = []mutant{
	// `**/` stops after one directory: the token is no longer inside once it has handed on at a `/`, so
	// `a/**/b` matches a/x/b but no longer a/x/y/b. One segment too few.
	{
		name: "** matching one segment too few",
		file: "glob.ts",
		from: "\t\t\t\t\t\taddPosition(nextInside, position);\n",
		to:   "\t\t\t\t\t\tif (character !== 0x2f) {\n\t\t\t\t\t\t\taddPosition(nextInside, position);\n\t\t\t\t\t\t}\n",

		seenByGit: true,
	},
	// Entering from a matcher below the root keeps only the root's ignore file, forgetting those read
	// on the way down to it. From the root there is nothing else to forget, so only a walk sees it.
	{
		name: "entering from below the root forgetting the files above",
		file: "gitignore.ts",
		from: "const entered = new Matcher(this.#tree, this.#directory, this.#files.slice(), this.#exclude, this.#excludedBy);",
		to:   "const entered = new Matcher(this.#tree, this.#directory, this.#files.slice(0, 1), this.#exclude, this.#excludedBy);",
	},
	// The deepest ignore file no longer goes first: the root's decides before a subdirectory's.
	{
		name: "precedence read from the root down",
		file: "gitignore.ts",
		from: "for (let fileIndex = fileCount - 1; fileIndex >= 0; fileIndex--) {",
		to:   "for (let fileIndex = 0; fileIndex < fileCount; fileIndex++) {",

		seenByGit: true,
	},
	// A linked info/exclude is refused as a linked .gitignore is, though git follows it.
	{
		name: "a linked exclude file refused",
		file: "gitignore.ts",
		from: "if (!follow) {",
		to:   "if (follow || !follow) {",

		seenByGit: true,
	},
	// A refusal quotes its line with uppercase hexadecimal, where Go's %q writes lowercase.
	{
		name: "%q in uppercase hexadecimal",
		file: "gitignore.ts",
		from: "'0123456789abcdef'",
		to:   "'0123456789ABCDEF'",
	},
	// Reviewer R2's six wrong ports (review/r2/gitignore on claude/review-afternoon-merges-8la3q9), each
	// of which this test once let through.
	{
		name: "R2 [[:upper:]] one letter short",
		file: "glob.ts",
		from: "inClass = (character) => character >= 0x41 && character <= 0x5a;",
		to:   "inClass = (character) => character >= 0x41 && character <= 0x59;",
	},
	{
		name: "R2 [[:xdigit:]] taking G",
		file: "glob.ts",
		from: "(character | 0x20) <= 0x66)",
		to:   "(character | 0x20) <= 0x67)",
	},
	{
		name: "R2 an escaped range end read as the backslash",
		file: "glob.ts",
		from: "\t\t\t\tupper = utf8At(pattern, index);\n",
		to:   "\t\t\t\tupper = 0x5c;\n",
	},
	{
		name: "R2 a trailing tab trimmed as a space",
		file: "gitignore.ts",
		from: "\t\t\tcase ' ':\n\t\t\t\tif (spacesStart < 0) {",
		to:   "\t\t\tcase ' ':\n\t\t\tcase '\\t':\n\t\t\t\tif (spacesStart < 0) {",

		seenByGit: true,
	},
	{
		name: "R2 a path never cleaned",
		file: "gitignore.ts",
		from: "const cleaned = clean(relative);",
		to:   "const cleaned = relative;",
	},
	{
		name: "R2 the size limit one byte lower",
		file: "gitignore.ts",
		from: "if (utf8Length(info.contents) > maximumFileSize) {",
		to:   "if (utf8Length(info.contents) >= maximumFileSize) {",

		needsLargest: true,
	},
}

// largest says whether to ask about an ignore file of exactly 100 MiB, which makes every cases file
// over 100 MB: too much for every gate's parallel runs, so it's asked for with ADAMIC_GITIGNORE_LARGEST.
func largest() bool {
	return os.Getenv("ADAMIC_GITIGNORE_LARGEST") != ""
}

// gitSource is the git checkout git's corpora are read from, as cohere's tests read them, or "".
func gitSource() string {
	return os.Getenv("COHERE_GIT_SOURCE")
}

// askedCases has cohere's side generate every case, with the trees laid out under the test's scratch
// directory.
func askedCases(t *testing.T) cases {
	t.Helper()
	scratch, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	seed := int64(generatedSeed)
	if value := os.Getenv("COHERE_GITIGNORE_SEED"); value != "" {
		if seed, err = strconv.ParseInt(value, 10, 64); err != nil {
			t.Fatal(err)
		}
	}
	generated := 40
	if value := os.Getenv("COHERE_GITIGNORE_GENERATED"); value != "" {
		if generated, err = strconv.Atoi(value); err != nil {
			t.Fatal(err)
		}
	}
	repositoryRoot, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	realTrees := []string{repositoryRoot, filepath.Join(repositoryRoot, "cohere")}
	output := filepath.Join(scratch, "cases.json")
	cohereSide(t, map[string]any{
		"mode": "generate", "scratch": scratch, "seed": seed, "generated": generated,
		"gitSource": gitSource(), "realTrees": realTrees, "output": output, "largest": largest(),
	})
	contents, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	var asked cases
	if err := json.Unmarshal(contents, &asked); err != nil {
		t.Fatal(err)
	}
	t.Logf("seed %d, %d generated trees", seed, generated)
	return asked
}

// goCohere is Go cohere's answer to every case.
func goCohere(t *testing.T, asked cases) string {
	t.Helper()
	directory := t.TempDir()
	casesPath := filepath.Join(directory, "cases.json")
	encoded, err := json.Marshal(asked)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(casesPath, encoded, 0o644); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(directory, "answers.txt")
	cohereSide(t, map[string]any{"mode": "answer", "cases": casesPath, "output": output})
	answers, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	return string(answers)
}

// cohereSide runs testdata/cohere_side_test.go inside cohere's gitignore package, by overlay, with a
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
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{
		filepath.Join(cohere, "internal", "gitignore", "adamic_port_side_test.go"): side,
	}})
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0o644); err != nil {
		t.Fatal(err)
	}
	command := bounded(t, "go", "test", "-count=1", "-overlay="+overlayPath, "-run=^TestAdamicPortCases$", "./internal/gitignore")
	command.Dir = cohere
	command.Env = append(os.Environ(), "ADAMIC_PORT_REQUEST="+requestPath)
	if output, err := combinedOutput(command); err != nil {
		t.Fatalf("cohere's side, %s: %v\n%s", request["mode"], err, output)
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

// casesFile writes asked as the cases file main.ts reads (case.ts says the format), and returns its
// path. ADAMIC_GITIGNORE_KEEP names a path to keep a copy at, to run the port on by hand.
func casesFile(t *testing.T, asked cases) string {
	t.Helper()
	escape := strings.NewReplacer("\\", "\\\\", "\t", "\\t", "\n", "\\n", "\r", "\\r")
	var text strings.Builder
	record := func(fields ...string) {
		for index, field := range fields {
			if !utf8.ValidString(field) {
				t.Fatalf("%q is not UTF-8, which the port reads its cases as", field)
			}
			if index > 0 {
				text.WriteByte('\t')
			}
			text.WriteString(escape.Replace(field))
		}
		text.WriteByte('\n')
	}
	bit := func(value bool) string {
		if value {
			return "1"
		}
		return "0"
	}
	for _, tree := range asked.Trees {
		record("tree", tree.Name, tree.Root)
		for _, entry := range tree.Entries {
			switch entry.Kind {
			case "File":
				record("entry", entry.Path, entry.Kind, entry.Contents)
			case "SymbolicLink":
				record("entry", entry.Path, entry.Kind, entry.Target)
			case "Directory", "Other":
				record("entry", entry.Path, entry.Kind)
			default:
				t.Fatalf("an entry of kind %q", entry.Kind)
			}
		}
		for _, query := range tree.Queries {
			record("query", query.Path, bit(query.IsDirectory))
		}
	}
	for _, list := range asked.Patterns {
		record("patterns", list.Name)
		for _, line := range list.Lines {
			record("line", line)
		}
		for _, query := range list.Queries {
			record("query", query.Path, bit(query.IsDirectory))
		}
	}
	for _, glob := range asked.Globs {
		record("glob", glob.Pattern, glob.Text, bit(glob.Path))
	}
	path := filepath.Join(t.TempDir(), "cases.txt")
	if err := os.WriteFile(path, []byte(text.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	if keep := os.Getenv("ADAMIC_GITIGNORE_KEEP"); keep != "" {
		if err := os.WriteFile(keep, []byte(text.String()), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return path
}

// answersByTree splits main.ts's output into each tree's answer lines, and the glob lines under
// "globs".
func answersByTree(output string) map[string][]string {
	answers := map[string][]string{}
	current := ""
	for _, line := range strings.Split(strings.TrimSuffix(output, "\n"), "\n") {
		switch {
		case strings.HasPrefix(line, "tree "):
			current = strings.TrimPrefix(line, "tree ")
		case strings.HasPrefix(line, "patterns "), strings.HasPrefix(line, "walk "):
			current = line
		case strings.HasPrefix(line, "glob "):
			answers["globs"] = append(answers["globs"], line)
		default:
			answers[current] = append(answers[current], line)
		}
	}
	return answers
}

// gitCheckIgnore is git's answer for every query of a tree, in main.ts's words, from git check-ignore
// run on the tree as cohere's differential runs it: no index, no configuration but the repository's own,
// no global excludes file, and case sensitive. cohere runs git under macOS's sandbox-exec, since it
// points git at real trees; here git reads only trees this test made, and this repository.
func gitCheckIgnore(t *testing.T, tree tree) []string {
	t.Helper()
	scratch := t.TempDir()
	emptyExclude := filepath.Join(scratch, "empty-excludes")
	if err := os.WriteFile(emptyExclude, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	var input bytes.Buffer
	for _, query := range tree.Queries {
		input.WriteString(query.Path)
		input.WriteByte(0)
	}
	command := bounded(t, "git", "-c", "core.excludesFile="+emptyExclude, "-c", "core.ignorecase=false",
		"check-ignore", "--no-index", "--verbose", "--non-matching", "-z", "--stdin")
	command.Dir = tree.Disk
	command.Env = []string{
		"PATH=/usr/bin:/bin", "HOME=" + scratch, "XDG_CONFIG_HOME=" + scratch,
		"GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=/dev/null", "GIT_OPTIONAL_LOCKS=0",
	}
	command.Stdin = &input
	var standardOutput, standardError bytes.Buffer
	command.Stdout = &standardOutput
	command.Stderr = &standardError
	err := childguard.Run(command, childguard.Options{})
	output := standardOutput.Bytes()
	// check-ignore exits 1 when nothing is ignored, which is an answer, not a failure.
	var exitError *exec.ExitError
	if err != nil && !(errors.As(err, &exitError) && exitError.ExitCode() == 1) {
		t.Fatalf("git check-ignore in %s: %v\n%s", tree.Disk, err, standardError.String())
	}
	fields := bytes.Split(bytes.TrimSuffix(output, []byte{0}), []byte{0})
	if len(output) > 0 && len(fields)%4 != 0 {
		t.Fatalf("%s: git printed %d fields, not a multiple of four", tree.Name, len(fields))
	}
	var answers []string
	for index := 0; index+3 < len(fields); index += 4 {
		file, line, pattern, relative := string(fields[index]), string(fields[index+1]), string(fields[index+2]), string(fields[index+3])
		if file == "" {
			answers = append(answers, "0 ::\t"+relative)
			continue
		}
		ignored := "1"
		if strings.HasPrefix(pattern, "!") {
			ignored = "0"
		}
		answers = append(answers, fmt.Sprintf("%s %s:%s:%s\t%s", ignored, file, line, pattern, relative))
	}
	return answers
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
