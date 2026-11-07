package selector

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/system-inc/adamic/stage1/cohere/selector/oracles"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
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
var portFiles = []string{"source.ts", "nodes.ts", "tokenize.ts", "parser.ts", "main.ts"}

// generatedSeed fixes the generated selectors, so a failure names selectors anyone can make again.
// Setting COHERE_SELECTOR_SEED asks for others, and COHERE_SELECTOR_GENERATED for more or fewer of them.
const generatedSeed = oracles.DefaultSeed

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
//
// The comparison with the library and each mutant are tests of their own, beside this one, so no test
// carries another's lowering and builds; the cases and Go cohere's answers are a product they all read, and
// the sanitized port is portSanitized's, built ahead. The lowered program is for the JavaScript backend,
// and for leaks off Linux.
func TestThePortParsesAsGoCohereDoes(test *testing.T) {
	test.Parallel()
	casesPath, goAnswers := askedCases(test)
	portSource := portDirectory(test, nil)
	program := lowered(test, filepath.Join(portSource, "main.ts"))
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

	if agreed {
		test.Logf(
			"%d selectors: every answer agrees on Node, native, and the JavaScript backend",
			strings.Count(goAnswers, "case "),
		)
	}
}

// postcss-selector-parser 2.2.3 itself parses the port's cases as Go cohere does, except its explicitly
// reported nonterminating inputs.
func TestPostcssSelectorParserParsesAsGoCohereDoes(test *testing.T) {
	test.Parallel()
	library := os.Getenv("ADAMIC_SELECTOR_LIBRARY")
	if library == "" {
		test.Skip(
			"set ADAMIC_SELECTOR_LIBRARY to a directory where `npm install postcss-selector-parser@2.2.3` ran, to compare with the library itself",
		)
	}
	casesPath, goAnswers := askedCases(test)
	script, absolutePathError := filepath.Abs(filepath.Join("testdata", "library.mjs"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	answersPath := filepath.Join(test.TempDir(), "go-answers.txt")
	if writeError := os.WriteFile(answersPath, []byte(goAnswers), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	libraryRun := execute(test, nil, "node", script, library, casesPath, answersPath)
	if libraryRun.exitCode != 0 {
		test.Fatalf("the library: exit %d, stderr %q", libraryRun.exitCode, libraryRun.stderr)
	}
	test.Logf("%s", libraryRun.stderr)
	if difference := firstDifference(string(libraryRun.stdout), goAnswers); difference != "" {
		test.Errorf(
			"postcss-selector-parser and Go cohere differ: %s",
			strings.Replace(difference, "line", "postcss-selector-parser's line", 1),
		)
	}
}

func TestTheComparisonCatchesTheSuffixOperatorLosingItsBar(test *testing.T) {
	test.Parallel()
	catches(test, suffixOperatorLosesBar)
}

func TestTheComparisonCatchesAQuoteIgnoringEscapeParity(test *testing.T) {
	test.Parallel()
	catches(test, quoteIgnoresEscapeParity)
}

func TestTheComparisonCatchesAMergedWordUsingItsFirstTokenPosition(test *testing.T) {
	test.Parallel()
	catches(test, mergedWordUsesFirstTokenPosition)
}

// mutant is one change to one port file.
type mutant struct {
	name, file, from, to string
}

// The mutants the comparison must catch, each by the one test named for it; comparisonMutants lists them for
// TestProduct_SelectorPortMutantsSanitized, which builds each ahead.
var (
	suffixOperatorLosesBar = mutant{
		name: "the suffix operator loses |=",
		file: "parser.ts",
		from: "/([*~^$|]?=)([\\s\\S]*)/",
		to:   "/([*~^$]?=)([\\s\\S]*)/",
	}
	quoteIgnoresEscapeParity = mutant{
		name: "a quote ignores escape parity",
		file: "tokenize.ts",
		from: "if(!escaped) {",
		to:   "if(true) {",
	}
	mergedWordUsesFirstTokenPosition = mutant{
		name: "a merged word uses its first token position",
		file: "parser.ts",
		from: "const token = this.curr();\n            if(partIndex === 0",
		to:   "const token = starting;\n            if(partIndex === 0",
	}
	comparisonMutants = []mutant{
		suffixOperatorLosesBar,
		quoteIgnoresEscapeParity,
		mergedWordUsesFirstTokenPosition,
	}
)

// catches shows the comparison can fail. Each mutant changes the port where only its answers can show it,
// and the port must then disagree with Go cohere: natively, since that is the program the agreement test
// holds, built ahead as portSanitized's, and on Node, which shows the fault is the port's and not stage 0's.
// Each mutant is caught by the one test named for it, so none can go unchecked or be checked twice.
func catches(test *testing.T, applied mutant) {
	test.Helper()
	casesPath, goAnswers := askedCases(test)
	mutated := portDirectory(test, &applied)
	binary := filepath.Join(portSanitized(test, &applied), "port")
	for _, side := range []struct {
		name string
		run  run
	}{
		{"natively", sanitizedRun(test, binary, casesPath)},
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
}

// askedCases returns the cases file's path and Go cohere's answer to each, from the product
// selectorAnswers builds, after checking the cohere checkout they come from.
func askedCases(test *testing.T) (string, string) { return oracles.Cases(test) }

// selectorAnswers is the directory holding cases.txt, every case cohere's side writes, and answers.txt, Go
// cohere's answer to each: a product built ahead by TestProduct_SelectorFindsGoAnswers, since asking takes two
// oracle runs and a corpus extraction every test reading them would otherwise pay. It is keyed by the
// request, the oracle and the library by content, the cohere directories the oracle and the extraction read,
// and this file and testdata, which hold the recipe.

// selectorAnswersInputs keys the answers product: the request, the oracle's content digest, and, when the
// library extracts the corpus, its content digest and node's report. libraryDigest is "" without the library.
func selectorAnswersInputs(seed int64, generated int, oracleDigest, libraryDigest, node string) buildcache.Inputs {
	return oracles.AnswerInputs(seed, generated, oracleDigest, libraryDigest, node)
}

// The answers product's name key (what a runner finds it by) and declared key (what this machine alone builds it by)
// each move with the library, the seed and the oracle, so no change to any serves stale answers. The digests are
// placeholders: keying builds nothing.
func TestSelectorAnswersRekey(test *testing.T) {
	test.Parallel()
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	keys := func(inputs buildcache.Inputs) (string, string) {
		declared, keyError := buildcache.Key(root, inputs)
		if keyError != nil {
			test.Fatal(keyError)
		}
		return buildcache.NameKey(root, inputs), declared
	}
	nameKey, declared := keys(selectorAnswersInputs(generatedSeed, 4000, "0", "1", "node"))
	for _, change := range []struct {
		name   string
		inputs buildcache.Inputs
	}{
		{"no library", selectorAnswersInputs(generatedSeed, 4000, "0", "", "")},
		{"another seed", selectorAnswersInputs(generatedSeed+1, 4000, "0", "1", "node")},
		{"another oracle", selectorAnswersInputs(generatedSeed, 4000, "2", "1", "node")},
	} {
		changedNameKey, changedDeclared := keys(change.inputs)
		if changedNameKey == nameKey || changedDeclared == declared {
			test.Errorf("%s kept name key %s or key %s", change.name, nameKey[:12], declared[:12])
			continue
		}
		test.Logf(
			"%s: name key %s to %s, key %s to %s",
			change.name,
			nameKey[:12],
			changedNameKey[:12],
			declared[:12],
			changedDeclared[:12],
		)
	}
}

// contentDigest hashes a file, or every file under a directory by its relative name, so a key names what a path
// holds and never where it is.

// cohereSide runs testdata/cohere_side_test.go inside cohere's selector package, by overlay, with a
// request.

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

// bounded configures a child command; Run and CombinedOutput below guard its output progress.
// Silent builds and buffered children use childguard's 30-minute FirstOutput window;
// after output starts, the default two-minute Stall allows over 3x the measured loaded gaps.
func bounded(test *testing.T, name string, arguments ...string) *exec.Cmd {
	test.Helper()
	command := exec.Command(name, arguments...)
	test.Cleanup(func() {
		if command.Process != nil && command.ProcessState == nil {
			_ = syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		}
	})
	return command
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

// Not parallel: throughput is measured alone, not while other ports compile.
func TestSelectorThroughput(test *testing.T) {
	if os.Getenv("ADAMIC_SELECTOR_BENCH") == "" {
		test.Skip("set ADAMIC_SELECTOR_BENCH=1 to time parser throughput")
	}
	casesPath, answers := askedCases(test)
	program := lowered(test, filepath.Join(portDirectory(test, nil), "main.ts"))
	binary := filepath.Join(test.TempDir(), "selector")
	if buildError := native.Build(native.C(program), binary, native.Options{}); buildError != nil {
		test.Fatal(buildError)
	}
	count := strings.Count(answers, "case ") * 10
	answersPath := filepath.Join(test.TempDir(), "answers.txt")
	if writeError := os.WriteFile(answersPath, []byte(answers), 0644); writeError != nil {
		test.Fatal(writeError)
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
			{"Node source", "node", []string{
				"--disable-warning=ExperimentalWarning",
				filepath.Join(repository, "oracle", "node.mjs"),
				"main.ts",
				casesPath,
				"count",
				"repeat",
			}},
		}
		if library != "" {
			sides = append(sides, struct {
				name, command string
				args          []string
			}{"postcss-selector-parser Node", "node", []string{script, library, "count", casesPath, answersPath}})
		}
		for _, side := range sides {
			start := time.Now()
			result := execute(test, nil, side.command, side.args...)
			elapsed := time.Since(start)
			if result.exitCode != 0 || len(result.stderr) != 0 {
				test.Fatalf("%s: exit %d, %s", side.name, result.exitCode, result.stderr)
			}
			if expected == "" {
				expected = string(result.stdout)
			}
			if string(result.stdout) != expected {
				test.Fatalf("throughput checksums differ: %q and %q", expected, result.stdout)
			}
			test.Logf(
				"%s round %d: %d selectors in %s, %.0f selectors/s; %s",
				side.name,
				round+1,
				count,
				elapsed,
				float64(count)/elapsed.Seconds(),
				result.stdout,
			)
		}
	}
}

// The vendored error fixtures may fail CSS parsing, but a parseable file must still contribute
// every selector, including one inside _errors_. Unexpected malformed files and exclusions that
// become parseable remain corpus failures.
func TestCorpusKeepsEveryParseableFile(test *testing.T) {
	test.Parallel()
	library := os.Getenv("ADAMIC_SELECTOR_LIBRARY")
	if library == "" {
		test.Skip("set ADAMIC_SELECTOR_LIBRARY to check CSS corpus loading")
	}
	root := test.TempDir()
	excluded := "internal/format/css/testdata/prettier/css/_errors_/less-syntax.css"
	for name, source := range map[string]string{
		"valid.css":          ".required { color: red; }",
		"_errors_/valid.css": ".also-required { color: blue; }",
		excluded:             "a {.bordered();}",
	} {
		path := filepath.Join(root, name)
		if createDirectoryError := os.MkdirAll(filepath.Dir(path), 0755); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
		if writeError := os.WriteFile(path, []byte(source), 0644); writeError != nil {
			test.Fatal(writeError)
		}
	}
	script, absolutePathError := filepath.Abs(filepath.Join("testdata", "library.mjs"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	corpus := filepath.Join(test.TempDir(), "corpus.txt")
	// These are explicit generated controls, not an enumerated repository corpus.
	files := []string{
		filepath.Join(root, "_errors_/valid.css"),
		filepath.Join(root, excluded),
		filepath.Join(root, "valid.css"),
	}
	fileList := filepath.Join(test.TempDir(), "css-files.json")
	writeList := func() {
		data, encodeError := json.Marshal(files)
		if encodeError != nil {
			test.Fatal(encodeError)
		}
		if writeError := os.WriteFile(fileList, data, 0644); writeError != nil {
			test.Fatal(writeError)
		}
	}
	writeList()
	result := execute(test, nil, "node", script, library, "corpus", root, corpus, "", fileList)
	if result.exitCode != 0 {
		test.Fatalf("corpus: %s", result.stderr)
	}
	selectors, readError := os.ReadFile(corpus)
	if readError != nil {
		test.Fatal(readError)
	}
	if string(selectors) != "\".also-required\"\n\".required\"\n" {
		test.Fatalf("parseable files lost selectors: %q", selectors)
	}
	if !strings.Contains(string(result.stdout), "3 CSS files, 1 expected CSS errors, 2 selectors") {
		test.Fatalf("expected CSS error was not counted: %s", result.stdout)
	}
	// A file not named by the Git manifest cannot silently join the corpus.
	if writeError := os.WriteFile(filepath.Join(root, "unlisted.css"), []byte(".stray {}"), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	unchanged := execute(test, nil, "node", script, library, "corpus", root, corpus, "", fileList)
	if unchanged.exitCode != 0 || string(unchanged.stdout) != string(result.stdout) {
		test.Fatalf("unlisted file changed corpus: exit %d, %s", unchanged.exitCode, unchanged.stderr)
	}
	if writeError := os.WriteFile(
		filepath.Join(root, excluded),
		[]byte(".newly-parseable {}"),
		0644,
	); writeError != nil {
		test.Fatal(writeError)
	}
	result = execute(test, nil, "node", script, library, "corpus", root, corpus, "", fileList)
	if result.exitCode == 0 || !strings.Contains(string(result.stderr), "expected CSS error became parseable") {
		test.Fatalf("parseable exclusion was omitted: exit %d, %s", result.exitCode, result.stderr)
	}
	if writeError := os.WriteFile(filepath.Join(root, excluded), []byte("a {.bordered();}"), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	if writeError := os.WriteFile(
		filepath.Join(root, "unexpected.css"),
		[]byte("a {.bordered();}"),
		0644,
	); writeError != nil {
		test.Fatal(writeError)
	}
	files = append(files, filepath.Join(root, "unexpected.css"))
	writeList()
	result = execute(test, nil, "node", script, library, "corpus", root, corpus, "", fileList)
	if result.exitCode == 0 || !strings.Contains(string(result.stderr), "unexpected.css") {
		test.Fatalf("unexpected CSS error was omitted: exit %d, %s", result.exitCode, result.stderr)
	}
}

// parserEntry forwards the first stdout line while retaining all output for the proof.
// Only the stdout copy goroutine writes it; the test reads the buffer after Run finishes.
type parserEntry struct {
	bytes.Buffer
	entered chan string
	sent    bool
}

func (output *parserEntry) Write(data []byte) (int, error) {
	n, writeError := output.Buffer.Write(data)
	if !output.sent {
		if end := bytes.IndexByte(output.Bytes(), '\n'); end >= 0 {
			output.sent = true
			output.entered <- string(output.Bytes()[:end+1])
		}
	}
	return n, writeError
}

// The explicit exception to JS agreement is observable, not an inference from Go.
func TestTheLibraryDoesNotReturnOnUnconsumedNamespaceBars(test *testing.T) {
	test.Parallel()
	library := os.Getenv("ADAMIC_SELECTOR_LIBRARY")
	if library == "" {
		test.Skip("set ADAMIC_SELECTOR_LIBRARY to prove upstream nontermination")
	}
	for _, text := range []string{"a| b", ":is(a|)", "a|@x", "*|)"} {
		test.Run(text, func(test *testing.T) {
			test.Parallel()
			// Startup uses the ordinary progress guard. The one-second observation after
			// "entering parser" deliberately proves nontermination, rather than speed.
			command := bounded(test, "node", "testdata/nontermination.mjs", library, text)
			stdout := &parserEntry{entered: make(chan string, 1)}
			var stderr bytes.Buffer
			command.Stdout, command.Stderr = stdout, &stderr
			ended := make(chan error, 1)
			go func() { ended <- childguard.Run(command, childguard.Options{}) }()
			stop := func() error {
				_ = syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
				return <-ended
			}
			select {
			case line := <-stdout.entered:
				if line != "entering parser\n" {
					_ = stop()
					test.Fatalf("expected %q first, got %q, stderr %q", "entering parser\n", line, stderr.String())
				}
			case err := <-ended:
				test.Fatalf(
					"node stopped before entering the parser: %v, stdout %q, stderr %q",
					err,
					stdout.String(),
					stderr.String(),
				)
			}
			select {
			case err := <-ended:
				test.Fatalf(
					"the parser returned within 1s of entering: exit %v, output after entering %q, stderr %q",
					err,
					strings.TrimPrefix(stdout.String(), "entering parser\n"),
					stderr.String(),
				)
			case <-time.After(time.Second):
			}
			_ = stop()
			if rest := strings.TrimPrefix(stdout.String(), "entering parser\n"); len(rest) != 0 {
				test.Fatalf("output after entering the parser: %q", rest)
			}
			test.Log("entered parser, no return within 1s of entering, killed")
		})
	}
}

// cohereSideOracle is the test binary go test would build with testdata/cohere_side_test.go laid over cohere's code, a
// product (buildcache.GoTest) built ahead, and the directory go test would run it in.
func cohereSideOracle(test testing.TB) (string, string) { return oracles.Parser(test) }

func TestProduct_SelectorFindsGoOracle(test *testing.T) {
	test.Parallel()
	cohereSideOracle(test)
}

func TestProduct_SelectorFindsGoAnswers(test *testing.T) {
	test.Parallel()
	askedCases(test)
}
