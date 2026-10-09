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
)

const repository = "../../.."

const testThePortParsesAsGoCohereDoesShards = 100

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
func TestThePortParsesAsGoCohereDoes(t *testing.T) {
	t.Parallel()
	setupStarted := time.Now()
	oracle := cssParserOracle(t)
	cases, answers := askedCasesIn(t, t.TempDir(), oracle)
	corpus := readCSSParserCorpus(t, cases, answers)
	units := cssParserShards(len(corpus.inputs))
	if len(units) != testThePortParsesAsGoCohereDoesShards {
		t.Fatalf("enumerated %d shards, declared %d; provide the complete pinned ADAMIC_CSS_FIXTURES corpus", len(units), testThePortParsesAsGoCohereDoesShards)
	}
	verifyCSSParserUnion(t, len(corpus.inputs), units)
	selected := selectedCSSParserShards(t, units)
	// C emission annotates the IR. Finish both backends sequentially before the
	// independent native builds and parallel unit readers use immutable products.
	main, err := filepath.Abs("main.ts")
	if err != nil {
		t.Fatal(err)
	}
	programs := []cssParserProgram{buildCSSParserProgram(t, "parser", filepath.Dir(main))}
	for i := range mutants {
		programs = append(programs, buildCSSParserProgram(t, fmt.Sprintf("mutant-%d", i), portDirectory(t, &mutants[i])))
	}
	binaries := buildCSSParserBinaries(t, programs[:1])
	leakBinary := binaries[0]
	if runtime.GOOS == "darwin" {
		leakBinary = buildCSSParserUnsanitized(t, programs[0])
	}
	library := os.Getenv("ADAMIC_CSS_LIBRARY")
	script, _ := filepath.Abs("testdata/library.mjs")
	setupElapsed := time.Since(setupStarted)
	t.Logf("setup before shards %.6fs, including cached normal products and private overlay builds", setupElapsed.Seconds())
	if setupElapsed > 30*time.Second {
		t.Fatalf("invalid test setup: wall %s exceeds 30s", setupElapsed)
	}
	for _, unit := range selected {
		t.Run(unit.name, func(t *testing.T) {
			t.Parallel()
			started := time.Now()
			defer func() {
				if elapsed := time.Since(started); elapsed > 30*time.Second {
					t.Errorf("invalid test unit %s: wall %s exceeds 30s", unit.name, elapsed)
				}
			}()
			t.Logf("%s: %s range [%d,%d), mutant %d", unit.name, unit.kind, unit.lo, unit.hi, unit.mutation)
			if unit.kind == "agreement" {
				shardCases, want := writeCSSParserRange(t, corpus, unit.lo, unit.hi)
				for _, side := range []struct {
					name   string
					result run
				}{
					{"native", execute(t, cssParserASAN(false), binaries[0], shardCases)},
					{"Node", onNode(t, programs[0].main, shardCases)},
					{"JS backend", onNode(t, programs[0].javascript, shardCases)},
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
				result := execute(t, nil, "node", script, library, shardCases)
				if result.exitCode != 0 {
					t.Fatalf("library: %s", result.stderr)
				}
				checkCSSParserPostCSS(t, unit, corpus.inputs[unit.lo:unit.hi], want, string(result.stdout))
			} else {
				// Private overlay products belong only to this mutant unit. Build once
				// and share the binary between its two original full-corpus checks.
				mutantBinary := buildCSSParserBinaries(t, programs[unit.mutation+1:unit.mutation+2])[0]
				for _, side := range []struct {
					name   string
					result run
				}{
					{"native", execute(t, cssParserASAN(false), mutantBinary, cases)},
					{"Node", onNode(t, programs[unit.mutation+1].main, cases)},
				} {
					if side.result.exitCode != 0 {
						t.Fatalf("%s: mutant must terminate: %s", side.name, side.result.stderr)
					}
					if difference := firstDifference(string(side.result.stdout), answers); difference == "" {
						t.Errorf("%s: mutant survived", side.name)
					} else {
						t.Logf("%s mutant %s caught by %s: %s", side.name, mutants[unit.mutation].name, unit.name, difference)
					}
				}
			}
		})
	}
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

// Hold the complete composed implementation against the external Go tree.
func TestCompositionMatchesGo(t *testing.T) {
	cases, _ := askedCases(t)
	directory := t.TempDir()
	answers := filepath.Join(directory, "answers.txt")
	repo, _ := filepath.Abs(repository)
	side, _ := filepath.Abs("testdata/compose_side_test.go")
	request, _ := json.Marshal(map[string]string{"Cases": cases, "Answers": answers})
	requestPath := filepath.Join(directory, "request.json")
	if err := os.WriteFile(requestPath, request, 0644); err != nil {
		t.Fatal(err)
	}
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{filepath.Join(repo, "cohere", "internal", "format", "css", "adamic_compose_side_test.go"): side}})
	overlayPath := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	command := bounded(t, "go", "test", "-timeout=0", "-v", "-count=1", "-overlay="+overlayPath, "-run=^TestAdamicCompositionCases$", "./internal/format/css")
	command.Dir = filepath.Join(repo, "cohere")
	command.Env = append(os.Environ(), "ADAMIC_PORT_REQUEST="+requestPath)
	output, err := childguard.CombinedOutput(command, childguard.Options{Stall: childStall})
	if err != nil {
		t.Fatalf("Go composition oracle: %v\n%s", err, output)
	}
	t.Logf("%s", output)
	data, err := os.ReadFile(answers)
	if err != nil {
		t.Fatal(err)
	}
	source, _ := filepath.Abs("compose_main.ts")
	program := lowered(t, source)
	nativeRun, sanitized := natively(t, program, cases)
	for _, side := range []struct {
		name   string
		result run
	}{
		{"native ASan/UBSan", nativeRun}, {"Node", onNode(t, source, cases)}, {"JavaScript backend", onJavaScriptBackend(t, program, cases)},
	} {
		if side.result.exitCode != 0 || len(side.result.stderr) != 0 {
			t.Fatalf("%s: %d %s", side.name, side.result.exitCode, side.result.stderr)
		}
		if difference := firstDifference(string(side.result.stdout), string(data)); difference != "" {
			t.Fatalf("%s: %s", side.name, difference)
		}
	}
	if report := leaks(t, program, sanitized, cases); report != "" {
		t.Fatal(report)
	}
	t.Logf("%d composed trees/error positions agree with Go on native ASan/UBSan, Node and JavaScript backend; LeakSanitizer clean", strings.Count(string(data), "\n")/2)
	if keep := os.Getenv("ADAMIC_CSS_KEEP_COMPOSED"); keep != "" {
		if err := os.WriteFile(keep, data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	for _, mutation := range mutants {
		t.Run("catches "+mutation.name, func(t *testing.T) {
			mutated := portDirectory(t, &mutation)
			result := onNode(t, filepath.Join(mutated, "compose_main.ts"), cases)
			if result.exitCode != 0 || len(result.stderr) != 0 {
				t.Fatalf("composed mutant must terminate: %s", result.stderr)
			}
			difference := firstDifference(string(result.stdout), string(data))
			if difference == "" {
				t.Fatal("composition comparison missed the mutant")
			}
			t.Logf("Node composition caught: %s", difference)
		})
	}
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
