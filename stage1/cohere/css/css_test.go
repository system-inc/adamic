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
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

const repository = "../../.."

const testThePortParsesAsGoCohereDoesShards = 388

// Loaded formatter output gaps reached 66.19 seconds; four minutes gives over 3x headroom.
const childStall = 4 * time.Minute

type run struct {
	stdout, stderr []byte
	exitCode       int
}
type mutant struct{ name, file, from, to string }

var mutants = []mutant{
	{"custom properties lose their block values", "parser.ts", "const custom = start.value.startsWith('--');", "const custom = false;"},
	{"comments always disappear from values", "parser.ts", "else {\n                        value += current.value;\n                    }", "else {\n                        value += '';\n                    }"},
	{"closing braces no longer include the closing byte", "parser.ts", "this.finish(node, token.start);\n            this.current = node.parent;", "this.finish(node, token.start, 0);\n            this.current = node.parent;"},
}

func portDirectory(t *testing.T, applied *mutant) string {
	t.Helper()
	directory := t.TempDir()
	for _, slice := range []string{"css", "selector", "values", "mediaquery", "cssstrings", "cssnumbers"} {
		entries, err := os.ReadDir(filepath.Join("..", slice))
		if err != nil {
			t.Fatal(err)
		}
		target := filepath.Join(directory, slice)
		if err := os.MkdirAll(target, 0755); err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".ts") {
				continue
			}
			data, err := os.ReadFile(filepath.Join("..", slice, entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			source := string(data)
			if applied != nil && slice == "css" && entry.Name() == applied.file {
				if strings.Count(source, applied.from) != 1 {
					t.Fatalf("mutant %s matches %d places", applied.name, strings.Count(source, applied.from))
				}
				source = strings.Replace(source, applied.from, applied.to, 1)
			}
			if err := os.WriteFile(filepath.Join(target, entry.Name()), []byte(source), 0644); err != nil {
				t.Fatal(err)
			}
		}
	}
	return filepath.Join(directory, "css")
}
func askedCases(t *testing.T) (string, string) {
	t.Helper()
	return askedCasesIn(t, t.TempDir(), "")
}

func askedCasesIn(t *testing.T, directory, oracle string) (string, string) {
	t.Helper()
	cases := filepath.Join(directory, "cases.txt")
	answers := filepath.Join(directory, "answers.txt")
	repo, _ := filepath.Abs(repository)
	side, _ := filepath.Abs("testdata/cohere_side_test.go")
	packageDirectory := filepath.Join(repo, "cohere", "internal", "format", "css", "postcss")
	request := map[string]any{"cases": cases, "answers": answers, "repository": repo, "fixtures": os.Getenv("ADAMIC_CSS_FIXTURES")}
	encoded, _ := json.Marshal(request)
	requestPath := filepath.Join(directory, "request.json")
	if err := os.WriteFile(requestPath, encoded, 0644); err != nil {
		t.Fatal(err)
	}
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{filepath.Join(packageDirectory, "adamic_port_side_test.go"): side}})
	overlayPath := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	cmd := bounded(t, "go", "test", "-timeout=0", "-v", "-count=1", "-overlay="+overlayPath, "-run=^TestAdamicPortCases$", "./internal/format/css/postcss")
	cmd.Dir = filepath.Join(repo, "cohere")
	if oracle != "" {
		cmd = bounded(t, oracle, "-test.timeout=0", "-test.v", "-test.count=1", "-test.run=^TestAdamicPortCases$")
		cmd.Dir = packageDirectory
	}
	cmd.Env = append(os.Environ(), "ADAMIC_PORT_REQUEST="+requestPath)
	output, err := childguard.CombinedOutput(cmd, childguard.Options{Stall: childStall})
	if err != nil {
		t.Fatalf("Go oracle: %v\n%s", err, output)
	}
	t.Logf("Go oracle: %s", output)
	data, err := os.ReadFile(answers)
	if err != nil {
		t.Fatal(err)
	}
	if keep := os.Getenv("ADAMIC_CSS_KEEP_RAW"); keep != "" {
		if err := os.WriteFile(keep, data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	if keep := os.Getenv("ADAMIC_CSS_KEEP"); keep != "" {
		contents, err := os.ReadFile(cases)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(keep, contents, 0644); err != nil {
			t.Fatal(err)
		}
	}
	return cases, string(data)
}

// TestThePortParsesAsGoCohereDoes enumerates the full corpus before starting
// deterministic shard-NNN subtests. ADAMIC_TEST_SHARD=i/n (zero-based i) selects
// ordinals modulo n as a local convenience; the gate can instead use
// -run '^TestThePortParsesAsGoCohereDoes$/^shard-NNN$'. Every agreement shard
// compares all original sides and runs ASan/UBSan and the separate leak check.
// TestThePortParsesAsGoCohereDoes_NNN are top-level parallel units. ADAMIC_TEST_SHARD=i/n
// selects stable ordinals modulo n locally; unset runs every unit. The gate uses
// -run '^TestThePortParsesAsGoCohereDoes_NNN$'. Every original side and sanitizer/leak check stays.

func prepareThePortParsesAsGoCohereDoesSetup(t *testing.T) {
	t.Helper()
	cssParserTopSetup.once.Do(func() {
		setupStarted := time.Now()
		prepareCSSParserTopCorpus(t)
		corpus := cssParserTopCorpus(t)
		units := cssParserShards(corpus.keys)
		if len(units) != testThePortParsesAsGoCohereDoesShards {
			t.Fatalf("enumerated %d shards, declared %d; provide the complete pinned ADAMIC_CSS_FIXTURES corpus", len(units), testThePortParsesAsGoCohereDoesShards)
		}
		verifyCSSParserUnion(t, corpus.keys, units)
		// C emission annotates the IR. Finish both backends sequentially before the
		// independent native builds and parallel unit readers use immutable products.
		var programs []cssParserProgram
		var binaries []string
		for index := -1; index < len(mutants); index++ {
			programs = append(programs, cssParserPreparedProgram(t, index))
			binaries = append(binaries, cssParserPreparedNative(t, index))
		}
		leakBinary := binaries[0]
		if runtime.GOOS == "darwin" {
			leakBinary = cssParserPreparedLeaks(t)
		}
		library := os.Getenv("ADAMIC_CSS_LIBRARY")
		script, _ := filepath.Abs("testdata/library.mjs")
		libraryIdentity := ""
		if library != "" {
			libraryIdentity = cssParserOracleIdentity(t, library)
		}
		setupElapsed := time.Since(setupStarted)
		t.Logf("setup before shards %.6fs, including cached build products", setupElapsed.Seconds())
		cssParserTopSetup.units = units
		cssParserTopSetup.run = func(t *testing.T, root cssParserShard) {
			for _, unit := range root.groups {
				started := time.Now()
				defer func() {
					if elapsed := time.Since(started); elapsed > 60*time.Second {
						t.Errorf("invalid test unit %s: wall %s exceeds 60s", unit.name, elapsed)
					}
				}()
				t.Logf("%s: %s, %d stable case IDs, mutant %d", unit.name, unit.kind, len(unit.cases), unit.mutation)
				if unit.kind == "agreement" {
					shardCases, want := writeCSSParserSlice(t, corpus, unit.cases)
					for _, side := range []struct {
						name   string
						result run
					}{
						{"native", cssParserExecute(t, cssParserASAN(false), binaries[0], shardCases)},
						{"Node", cssParserOnNode(t, programs[0].main, shardCases)},
						{"JS backend", cssParserOnNode(t, programs[0].javascript, shardCases)},
					} {
						if side.result.exitCode != 0 || len(side.result.stderr) != 0 {
							t.Fatalf("%s: exit %d, %s", side.name, side.result.exitCode, side.result.stderr)
						}
						if difference := cssParserDifference(unit, string(side.result.stdout), want); difference != "" {
							t.Errorf("%s: %s", side.name, difference)
						}
					}
					checkCSSParserLeaks(t, binaries[0], leakBinary, shardCases)
					if library == "" {
						t.Skip("set ADAMIC_CSS_LIBRARY to the pinned npm scratch directory")
					}
					postCSS := cachedCSSParserPostCSSAnswers(t, script, library, shardCases, libraryIdentity)
					slice := make([]string, len(unit.cases))
					for i, index := range unit.cases {
						slice[i] = corpus.inputs[index]
					}
					checkCSSParserPostCSS(t, unit, slice, want, postCSS)
				} else {
					// Fetch this mutant product once and share the immutable binary
					// between its two original side checks for this complete case range.
					mutantBinary := binaries[unit.mutation+1]
					shardCases, want := writeCSSParserSlice(t, corpus, unit.cases)
					witness := cssParserMutantWitness(unit.mutation)
					for _, side := range []struct {
						name   string
						result run
					}{
						{"native", cssParserExecute(t, cssParserASAN(false), mutantBinary, shardCases)},
						{"Node", cssParserOnNode(t, programs[unit.mutation+1].main, shardCases)},
					} {
						if side.result.exitCode != 0 {
							t.Fatalf("%s: mutant must terminate: %s", side.name, side.result.stderr)
						}
						if difference := cssParserDifference(unit, string(side.result.stdout), want); difference == "" {
							if cssParserContainsCase(unit.cases, witness) {
								t.Errorf("%s: mutant survived witness case %d in %s", side.name, witness, unit.name)
							}
						} else {
							t.Logf("%s mutant %s caught by %s: %s", side.name, mutants[unit.mutation].name, unit.name, difference)
						}
					}
				}
			}
		}
	})
}

func runThePortParsesAsGoCohereDoesShard(t *testing.T, ordinal int) {
	t.Helper()
	// Fetch shared products once per process before starting this shard's clock.
	prepareThePortParsesAsGoCohereDoesSetup(t)
	if cssParserTopSetup.run == nil {
		t.Fatal("shared setup did not complete")
	}
	timer := time.AfterFunc(90*time.Second, func() {
		panic(fmt.Sprintf("cooked TestThePortParsesAsGoCohereDoes_%03d: case deadline exceeded 90s", ordinal))
	})
	defer timer.Stop()
	started := time.Now()
	defer func() {
		if elapsed := time.Since(started); elapsed > 60*time.Second {
			t.Errorf("invalid top-level unit: %s exceeds 60s", elapsed)
		}
	}()
	if os.Getenv("ADAMIC_CSS_PARSER_DISAGREEMENT_PROBE") == "1" {
		cssParserTopProbe(t, ordinal)
		return
	}
	if cssParserTopSetup.run == nil {
		t.Fatal("shared setup did not complete")
	}
	selected := selectedCSSParserShards(t, cssParserTopSetup.units)
	for _, unit := range selected {
		if unit.name == cssParserTopSetup.units[ordinal].name {
			cssParserTopSetup.run(t, unit)
			return
		}
	}
	t.Skip("excluded by ADAMIC_TEST_SHARD")
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

// bounded configures a child command; Run and CombinedOutput below guard its output progress.
// Silent builds and buffered children use childguard's 30-minute FirstOutput window;
// after output starts, childStall allows over 3x the measured loaded gaps.
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
	err := childguard.Run(command, childguard.Options{Stall: childStall})
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

// Hold the complete composed implementation against the external Go tree.
// Not parallel: initializes the composition worker's shared setup and publishes shared build products and native runtime cache entries.
func TestCompositionMatchesGo(t *testing.T) {
	compositionSetup(t)
}

// Not parallel: parser throughput runs alone, after agreement has been checked.
func TestCSSThroughput(t *testing.T) {
	if os.Getenv("ADAMIC_CSS_BENCH") == "" {
		t.Skip("set ADAMIC_CSS_BENCH=1 for throughput")
	}
	cases, answers := askedCases(t)
	directory := portDirectory(t, nil)
	program := lowered(t, filepath.Join(directory, "main.ts"))
	binary := filepath.Join(t.TempDir(), "css")
	if err := native.Build(native.C(program), binary, native.Options{}); err != nil {
		t.Fatal(err)
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
			{"Node source", "node", []string{"--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "main.ts"), cases, "count", "repeat"}},
		}
		if library := os.Getenv("ADAMIC_CSS_LIBRARY"); library != "" {
			sides = append(sides, struct {
				name, command string
				args          []string
			}{"PostCSS Node", "node", []string{script, library, cases, "count"}})
		}
		for _, side := range sides {
			started := time.Now()
			result := execute(t, nil, side.command, side.args...)
			elapsed := time.Since(started)
			if result.exitCode != 0 || len(result.stderr) != 0 {
				t.Fatalf("%s: exit %d, %s", side.name, result.exitCode, result.stderr)
			}
			if expected == "" {
				expected = string(result.stdout)
			}
			if string(result.stdout) != expected {
				t.Fatalf("%s checksum %q differs from %q", side.name, result.stdout, expected)
			}
			t.Logf("%s round %d: %d stylesheets in %s, %.0f stylesheets/s; %s", side.name, round+1, count, elapsed, float64(count)/elapsed.Seconds(), result.stdout)
		}
	}
}

// Only a corrupt Range catches these probes; source properties remain correct.
// Not parallel: canonical failure probes use the shared native toolchain and write adamic/runtime (and adamic/units for split builds).
func TestTheCanonicalRangeChecksCanFail(t *testing.T) {
	raw, err := filepath.Abs("testdata/range_guard.ts")
	if err != nil {
		t.Fatal(err)
	}
	program := lowered(t, raw)
	nativeRun, sanitized := natively(t, program)
	for _, side := range []struct {
		name   string
		result run
	}{{"raw native", nativeRun}, {"raw Node", onNode(t, raw)}} {
		if side.result.exitCode != 0 || len(side.result.stderr) != 0 || string(side.result.stdout) != "caught\n" {
			t.Fatalf("%s: exit %d, stdout %q, stderr %q", side.name, side.result.exitCode, side.result.stdout, side.result.stderr)
		}
	}
	if report := leaks(t, program, sanitized); report != "" {
		t.Fatal(report)
	}
	composed, err := filepath.Abs("testdata/composition_range_guard.ts")
	if err != nil {
		t.Fatal(err)
	}
	composedProgram := lowered(t, composed)
	composedNative, composedBinary := natively(t, composedProgram)
	for _, side := range []struct {
		name   string
		result run
	}{
		{"composition native", composedNative}, {"composition Node", onNode(t, composed)}, {"composition JavaScript backend", onJavaScriptBackend(t, composedProgram)},
	} {
		if side.result.exitCode != 0 || len(side.result.stderr) != 0 || string(side.result.stdout) != "caught\n" {
			t.Fatalf("%s: %d %q %s", side.name, side.result.exitCode, side.result.stdout, side.result.stderr)
		}
	}
	if report := leaks(t, composedProgram, composedBinary); report != "" {
		t.Fatal(report)
	}
	t.Log("public Range corruption caught on raw and composed backends; leak clean")
}
