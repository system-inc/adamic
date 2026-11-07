package css

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
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/stage1/cohere/css/oracles"
)

const repository = "../../.."

type mutant struct{ name, file, from, to string }

var mutants = []mutant{
	{
		"custom properties lose their block values",
		"parser.ts",
		"const custom = start.value.startsWith('--');",
		"const custom = false;",
	},
	{
		"comments always disappear from values",
		"parser.ts",
		"else {\n                        value += current.value;\n                    }",
		"else {\n                        value += '';\n                    }",
	},
	{
		"closing braces no longer include the closing byte",
		"parser.ts",
		"this.finish(node, token.start);\n            this.current = node.parent;",
		"this.finish(node, token.start, 0);\n            this.current = node.parent;",
	},
}

type run struct {
	stdout, stderr []byte
	exitCode       int
}

func portDirectory(test *testing.T, applied *mutant) string {
	test.Helper()
	directory := test.TempDir()
	for _, slice := range []string{"css", "selector", "values", "mediaquery", "cssstrings", "cssnumbers"} {
		// Reads: stage1/cohere/cssnumbers/*.ts
		// Reads: stage1/cohere/cssnumbers/numbers.ts
		// Reads: stage1/cohere/cssstrings/*.ts
		// Reads: stage1/cohere/cssstrings/strings.ts
		// Reads: stage1/cohere/mediaquery/*.ts
		// Reads: stage1/cohere/mediaquery/index.ts
		// Reads: stage1/cohere/mediaquery/nodes.ts
		// Reads: stage1/cohere/mediaquery/parsers.ts
		// Reads: stage1/cohere/mediaquery/whitespace.ts
		// Reads: stage1/cohere/selector/*.ts
		// Reads: stage1/cohere/selector/nodes.ts
		// Reads: stage1/cohere/selector/parser.ts
		// Reads: stage1/cohere/selector/source.ts
		// Reads: stage1/cohere/selector/tokenize.ts
		// Reads: stage1/cohere/values/*.ts
		// Reads: stage1/cohere/values/nodes.ts
		// Reads: stage1/cohere/values/parser.ts
		// Reads: stage1/cohere/values/tokenize.ts
		// Reads: stage1/cohere/values/values.ts
		entries, readDirectoryError := os.ReadDir(filepath.Join("..", slice))
		if readDirectoryError != nil {
			test.Fatal(readDirectoryError)
		}
		target := filepath.Join(directory, slice)
		if createDirectoryError := os.MkdirAll(target, 0755); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".ts") {
				continue
			}
			data, readError := os.ReadFile(filepath.Join("..", slice, entry.Name()))
			if readError != nil {
				test.Fatal(readError)
			}
			source := string(data)
			if applied != nil && slice == "css" && entry.Name() == applied.file {
				if strings.Count(source, applied.from) != 1 {
					test.Fatalf("mutant %s matches %d places", applied.name, strings.Count(source, applied.from))
				}
				source = strings.Replace(source, applied.from, applied.to, 1)
			}
			if writeError := os.WriteFile(
				filepath.Join(target, entry.Name()),
				[]byte(source),
				0644,
			); writeError != nil {
				test.Fatal(writeError)
			}
		}
	}
	return filepath.Join(directory, "css")
}
func askedCases(test *testing.T) (string, string) {
	test.Helper()
	return askedCasesIn(test, test.TempDir(), "")
}

func askedCasesIn(test *testing.T, directory, oracle string) (string, string) {
	test.Helper()
	cases := filepath.Join(directory, "cases.txt")
	answers := filepath.Join(directory, "answers.txt")
	repo, _ := filepath.Abs(repository)
	packageDirectory := filepath.Join(repo, "cohere", "internal", "format", "css", "postcss")
	request := map[string]any{
		"cases":      cases,
		"answers":    answers,
		"repository": repo,
		"roots":      oracles.CorpusRoots(),
		"fixtures":   os.Getenv("ADAMIC_CSS_FIXTURES"),
	}
	encoded, _ := json.Marshal(request)
	requestPath := filepath.Join(directory, "request.json")
	if writeError := os.WriteFile(requestPath, encoded, 0644); writeError != nil {
		test.Fatal(writeError)
	}
	// The Go oracle is css-printer-oracle-0, cohere's postcss with testdata/cohere_side_test.go laid over it, built
	// ahead; the unit only runs it, in the package's directory (#5qykzj5).
	if oracle == "" {
		oracle = oracles.ComparisonOracle(test, false)
	}
	cmd := bounded(test, oracle, "-test.timeout=0", "-test.v", "-test.count=1", "-test.run=^TestAdamicPortCases$")
	cmd.Dir = packageDirectory
	cmd.Env = append(os.Environ(), "ADAMIC_PORT_REQUEST="+requestPath)
	output, combinedOutputError := childguard.CombinedOutput(cmd, childguard.Options{Stall: oracles.ChildStallLimit})
	if combinedOutputError != nil {
		test.Fatalf("Go oracle: %v\n%s", combinedOutputError, output)
	}
	test.Logf("Go oracle: %s", output)
	data, readError := os.ReadFile(answers)
	if readError != nil {
		test.Fatal(readError)
	}
	if keep := os.Getenv("ADAMIC_CSS_KEEP_RAW"); keep != "" {
		if writeError := os.WriteFile(keep, data, 0644); writeError != nil {
			test.Fatal(writeError)
		}
	}
	if keep := os.Getenv("ADAMIC_CSS_KEEP"); keep != "" {
		contents, readError := os.ReadFile(cases)
		if readError != nil {
			test.Fatal(readError)
		}
		if writeError := os.WriteFile(keep, contents, 0644); writeError != nil {
			test.Fatal(writeError)
		}
	}
	return cases, string(data)
}

// TestThePortParsesAsGoCohereDoes enumerates the full corpus before starting
// deterministic shard-NNN subtests. ADAMIC_TEST_SHARD=i/n (zero-based i) selects
// ordinals modulo n as a local convenience; the gate can instead use
// -run '^TestThePortParsesAsGoCohereDoes$/^shard-NNN$'. Every agreement shard
// compares all original sides and runs ASan/UBSan, with leak detection in the same
// run on Linux. Each mutant runs over exactly its witnesses (cssParserWitnesses), in
// the shard that owns the first of them.
// TestThePortParsesAsGoCohereDoes_NNN are top-level parallel units. ADAMIC_TEST_SHARD=i/n
// selects stable ordinals modulo n locally; unset runs every unit. The gate uses
// -run '^TestThePortParsesAsGoCohereDoes_NNN$'. Every original side and sanitizer/leak check stays.

func prepareThePortParsesAsGoCohereDoesSetup(test *testing.T) {
	test.Helper()
	cssParserTopSetup.once.Do(func() {
		setupStarted := time.Now()
		corpus := oracles.LoadParserCorpus(test)
		witnesses := cssParserAllWitnesses(test)
		units := oracles.ParserWithMutants(oracles.ParserAgreementShards(corpus.CaseKeys), corpus.CaseKeys, witnesses)
		if len(units) != oracles.ParserShardCount {
			test.Fatalf(
				"enumerated %d shards, declared %d; provide the complete pinned ADAMIC_CSS_FIXTURES corpus",
				len(units),
				oracles.ParserShardCount,
			)
		}
		verifyCSSParserUnion(test, corpus.CaseKeys, units, witnesses)
		// C emission annotates the IR. Finish both backends sequentially before the
		// independent native builds and parallel unit readers use immutable products.
		var programs []cssParserProgram
		var binaries []string
		for index := -1; index < len(mutants); index++ {
			programs = append(programs, cssParserPreparedProgram(test, index))
			binaries = append(binaries, cssParserPreparedNative(test, index))
		}
		leakBinary := binaries[0]
		if runtime.GOOS == "darwin" {
			leakBinary = cssParserPreparedLeaks(test)
		}
		library := os.Getenv("ADAMIC_CSS_LIBRARY")
		script, _ := filepath.Abs("testdata/library.mjs")
		libraryIdentity := ""
		if library != "" {
			libraryIdentity = oracles.ContentIdentity(test, library)
		}
		setupElapsed := time.Since(setupStarted)
		test.Logf("setup before shards %.6fs, including cached build products", setupElapsed.Seconds())
		cssParserTopSetup.units = units
		cssParserTopSetup.run = func(test *testing.T, root oracles.ParserShard) {
			for _, unit := range root.Groups {
				started := time.Now()
				defer func() {
					if elapsed := time.Since(started); elapsed > 60*time.Second {
						test.Errorf("invalid test unit %s: wall %s exceeds 60s", unit.Name, elapsed)
					}
				}()
				test.Logf("%s: %s, %d stable case IDs, mutant %d", unit.Name, unit.Kind, len(unit.Cases), unit.Mutation)
				if unit.Kind == "agreement" {
					shardCases, want := oracles.WriteParserSlice(test, corpus, unit.Cases)
					for _, side := range []struct {
						name   string
						result run
					}{
						{"native", cssParserSanitizedAgreement(test, binaries[0], shardCases)},
						{"Node", cssParserOnNode(test, programs[0].main, shardCases)},
						{"JS backend", cssParserOnNode(test, programs[0].javascript, shardCases)},
					} {
						if side.result.exitCode != 0 || len(side.result.stderr) != 0 {
							test.Fatalf("%s: exit %d, %s", side.name, side.result.exitCode, side.result.stderr)
						}
						if difference := cssParserDifference(unit, string(side.result.stdout), want); difference != "" {
							test.Errorf("%s: %s", side.name, difference)
						}
					}
					checkCSSParserLeaks(test, binaries[0], leakBinary, shardCases)
					if library == "" {
						test.Skip("set ADAMIC_CSS_LIBRARY to the pinned npm scratch directory")
					}
					postCSS := oracles.CachedPostCSSAnswers(test, script, library, shardCases, libraryIdentity)
					slice := make([]string, len(unit.Cases))
					for i, index := range unit.Cases {
						slice[i] = corpus.Inputs[index]
					}
					checkCSSParserPostCSS(test, unit, slice, want, postCSS)
				} else {
					// Fetch this mutant product once and share the immutable binary
					// between its two original side checks over exactly its witnesses.
					mutantBinary := binaries[unit.Mutation+1]
					shardCases, want := oracles.WriteParserSlice(test, corpus, unit.Cases)
					for _, side := range []struct {
						name   string
						result run
					}{
						{"native", cssParserExecute(test, cssParserASAN(false), mutantBinary, shardCases)},
						{"Node", cssParserOnNode(test, programs[unit.Mutation+1].main, shardCases)},
					} {
						if side.result.exitCode != 0 {
							test.Fatalf("%s: mutant must terminate: %s", side.name, side.result.stderr)
						}
						differing, cssParserDifferingCasesError := cssParserDifferingCases(
							string(side.result.stdout),
							want,
						)
						if cssParserDifferingCasesError != nil {
							test.Fatalf(
								"%s: mutant %s in %s: %v",
								side.name,
								mutants[unit.Mutation].name,
								unit.Name,
								cssParserDifferingCasesError,
							)
						}
						if len(differing) != len(unit.Cases) {
							test.Errorf(
								"%s: mutant %s differs from Go on %d of its %d witnesses in %s",
								side.name,
								mutants[unit.Mutation].name,
								len(differing),
								len(unit.Cases),
								unit.Name,
							)
						} else {
							test.Logf(
								"%s mutant %s caught by %s on all %d witnesses: %s",
								side.name,
								mutants[unit.Mutation].name,
								unit.Name,
								len(unit.Cases),
								cssParserDifference(unit, string(side.result.stdout), want),
							)
						}
					}
				}
			}
		}
	})
}

func runThePortParsesAsGoCohereDoesShard(test *testing.T, ordinal int) {
	test.Helper()
	// Fetch shared products once per process before starting this shard's clock.
	prepareThePortParsesAsGoCohereDoesSetup(test)
	if cssParserTopSetup.run == nil {
		test.Fatal("shared setup did not complete")
	}
	timer := time.AfterFunc(90*time.Second, func() {
		panic(fmt.Sprintf("cooked TestThePortParsesAsGoCohereDoes_%03d: case deadline exceeded 90s", ordinal))
	})
	defer timer.Stop()
	started := time.Now()
	defer func() {
		if elapsed := time.Since(started); elapsed > 60*time.Second {
			test.Errorf("invalid top-level unit: %s exceeds 60s", elapsed)
		}
	}()
	if os.Getenv("ADAMIC_CSS_PARSER_DISAGREEMENT_PROBE") == "1" {
		cssParserTopProbe(test, ordinal)
		return
	}
	if cssParserTopSetup.run == nil {
		test.Fatal("shared setup did not complete")
	}
	selected := selectedCSSParserShards(test, cssParserTopSetup.units)
	for _, unit := range selected {
		if unit.Name == cssParserTopSetup.units[ordinal].Name {
			cssParserTopSetup.run(test, unit)
			return
		}
	}
	test.Skip("excluded by ADAMIC_TEST_SHARD")
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
// after output starts, childStall allows over 3x the measured loaded gaps.
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
	runError := childguard.Run(command, childguard.Options{Stall: oracles.ChildStallLimit})
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

// Hold the complete composed implementation against the external Go tree.
// Not parallel: initializes the composition worker's shared setup and publishes shared build products and native runtime cache entries.
func TestCompositionMatchesGo(test *testing.T) {
	compositionSetup(test)
}

// Not parallel: parser throughput runs alone, after agreement has been checked.
func TestCSSThroughput(test *testing.T) {
	if os.Getenv("ADAMIC_CSS_BENCH") == "" {
		test.Skip("set ADAMIC_CSS_BENCH=1 for throughput")
	}
	cases, answers := askedCases(test)
	directory := portDirectory(test, nil)
	program := lowered(test, filepath.Join(directory, "main.ts"))
	binary := filepath.Join(test.TempDir(), "css")
	if buildError := native.Build(native.C(program), binary, native.Options{}); buildError != nil {
		test.Fatal(buildError)
	}
	count := strings.Count(answers, "\n") / 2 * 10
	runner, _ := filepath.Abs(filepath.Join(repository, "oracle", "node.mjs"))
	script, _ := filepath.Abs("testdata/library.mjs")
	var expected string
	for round := 0; round < 3; round++ {
		sides := []struct {
			name, command string
			args          []string
		}{
			{"native", binary, []string{cases, "count", "repeat"}},
			{"Node source", "node", []string{
				"--disable-warning=ExperimentalWarning",
				runner,
				filepath.Join(directory, "main.ts"),
				cases,
				"count",
				"repeat",
			}},
		}
		if library := os.Getenv("ADAMIC_CSS_LIBRARY"); library != "" {
			sides = append(sides, struct {
				name, command string
				args          []string
			}{"PostCSS Node", "node", []string{script, library, cases, "count"}})
		}
		for _, side := range sides {
			started := time.Now()
			result := execute(test, nil, side.command, side.args...)
			elapsed := time.Since(started)
			if result.exitCode != 0 || len(result.stderr) != 0 {
				test.Fatalf("%s: exit %d, %s", side.name, result.exitCode, result.stderr)
			}
			if expected == "" {
				expected = string(result.stdout)
			}
			if string(result.stdout) != expected {
				test.Fatalf("%s checksum %q differs from %q", side.name, result.stdout, expected)
			}
			test.Logf(
				"%s round %d: %d stylesheets in %s, %.0f stylesheets/s; %s",
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

// Only a corrupt Range catches these probes; source properties remain correct.
// Not parallel: canonical failure probes use the shared native toolchain and write adamic/runtime (and adamic/units for split builds).
func TestTheCanonicalRangeChecksCanFail(test *testing.T) {
	raw, absolutePathError := filepath.Abs("testdata/range_guard.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program := lowered(test, raw)
	sanitized := filepath.Join(programSanitized(test, "testdata/range_guard.ts"), "port")
	nativeRun := sanitizedRun(test, sanitized)
	for _, side := range []struct {
		name   string
		result run
	}{{"raw native", nativeRun}, {"raw Node", onNode(test, raw)}} {
		if side.result.exitCode != 0 || len(side.result.stderr) != 0 || string(side.result.stdout) != "caught\n" {
			test.Fatalf(
				"%s: exit %d, stdout %q, stderr %q",
				side.name,
				side.result.exitCode,
				side.result.stdout,
				side.result.stderr,
			)
		}
	}
	if report := leaks(test, program, sanitized); report != "" {
		test.Fatal(report)
	}
	composed, absolutePathError := filepath.Abs("testdata/composition_range_guard.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	composedProgram := lowered(test, composed)
	composedBinary := filepath.Join(programSanitized(test, "testdata/composition_range_guard.ts"), "port")
	composedNative := sanitizedRun(test, composedBinary)
	for _, side := range []struct {
		name   string
		result run
	}{
		{"composition native", composedNative},
		{"composition Node", onNode(test, composed)},
		{"composition JavaScript backend", onJavaScriptBackend(test, composedProgram)},
	} {
		if side.result.exitCode != 0 || len(side.result.stderr) != 0 || string(side.result.stdout) != "caught\n" {
			test.Fatalf("%s: %d %q %s", side.name, side.result.exitCode, side.result.stdout, side.result.stderr)
		}
	}
	if report := leaks(test, composedProgram, composedBinary); report != "" {
		test.Fatal(report)
	}
	test.Log("public Range corruption caught on raw and composed backends; leak clean")
}
