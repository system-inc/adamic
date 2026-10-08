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
	directory := t.TempDir()
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
func TestThePortParsesAsGoCohereDoes(t *testing.T) {
	cases, answers := askedCases(t)
	directory := portDirectory(t, nil)
	program := lowered(t, filepath.Join(directory, "main.ts"))
	t.Run("Go agreement", func(t *testing.T) {
		nativeRun, sanitized := natively(t, program, cases)
		for _, side := range []struct {
			name   string
			result run
		}{{"native", nativeRun}, {"Node", onNode(t, filepath.Join(directory, "main.ts"), cases)}, {"JS backend", onJavaScriptBackend(t, program, cases)}} {
			if side.result.exitCode != 0 || len(side.result.stderr) != 0 {
				t.Fatalf("%s: exit %d, %s", side.name, side.result.exitCode, side.result.stderr)
			}
			if difference := firstDifference(string(side.result.stdout), answers); difference != "" {
				t.Errorf("%s: %s", side.name, difference)
			}
		}
		if report := leaks(t, program, sanitized, cases); report != "" {
			t.Errorf("leaks: %s", report)
		}
		if !t.Failed() {
			t.Logf("%d cases agree with Go, leak clean", strings.Count(answers, "\n")/2)
		}
	})
	t.Run("PostCSS", func(t *testing.T) {
		library := os.Getenv("ADAMIC_CSS_LIBRARY")
		if library == "" {
			t.Skip("set ADAMIC_CSS_LIBRARY to the pinned npm scratch directory")
		}
		script, _ := filepath.Abs("testdata/library.mjs")
		result := execute(t, nil, "node", script, library, cases)
		if result.exitCode != 0 {
			t.Fatalf("library: %s", result.stderr)
		}
		inputs, err := os.ReadFile(cases)
		if err != nil {
			t.Fatal(err)
		}
		lines := strings.Split(strings.TrimSuffix(string(inputs), "\n"), "\n")
		goLines := strings.Split(strings.TrimSuffix(answers, "\n"), "\n")
		jsLines := strings.Split(strings.TrimSuffix(string(result.stdout), "\n"), "\n")
		if len(goLines) != len(jsLines) {
			t.Fatalf("library printed %d lines, Go %d", len(jsLines), len(goLines))
		}
		gaps := 0
		for index, line := range lines {
			if goLines[index*2] != jsLines[index*2] {
				t.Fatalf("case marker differs at %d", index)
			}
			goAnswer, jsAnswer := goLines[index*2+1], jsLines[index*2+1]
			if goAnswer == jsAnswer {
				continue
			}
			// The original library splits the astral character after its high
			// surrogate; cohere's input.slice keeps the whole character on the left.
			if line[2:] == `\\😀|a` &&
				goAnswer == `error {"column":1,"endColumn":4,"endLine":1,"endOffset":5,"line":1,"name":"CssSyntaxError","offset":0,"reason":"Unknown word \\😀"}` &&
				jsAnswer == `error {"column":1,"endColumn":3,"endLine":1,"endOffset":4,"line":1,"name":"CssSyntaxError","offset":0,"reason":"Unknown word \\\ud83d"}` {
				gaps++
				continue
			}
			t.Errorf("case %d %q: unrecorded library difference: %s", index, line, firstDifference(jsAnswer, goAnswer))
		}
		t.Logf("PostCSS: %d exact agreements, %d occurrences of the proved surrogate gap", len(lines)-gaps, gaps)

	})
	for _, mutation := range mutants {
		t.Run("catches "+mutation.name, func(t *testing.T) {
			directory := portDirectory(t, &mutation)
			program := lowered(t, filepath.Join(directory, "main.ts"))
			for _, side := range []struct {
				name   string
				result run
			}{{"native", nativelyRun(t, program, cases)}, {"Node", onNode(t, filepath.Join(directory, "main.ts"), cases)}} {
				if side.result.exitCode != 0 {
					t.Fatalf("%s: mutant must terminate: %s", side.name, side.result.stderr)
				}
				difference := firstDifference(string(side.result.stdout), answers)
				if difference == "" {
					t.Errorf("%s: mutant survived", side.name)
				} else {
					t.Logf("%s caught: %s", side.name, difference)
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

// leaks returns a report of everything the finished port never let go of, or "": the leak check the
// oracle runs on every fixture (internal/leakcheck), on the sanitized binary and the port's C.
func leaks(t *testing.T, program *ir.Program, sanitized string, arguments ...string) string {
	t.Helper()
	return leakcheck.Report(t, native.C(program), sanitized, arguments...)
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
