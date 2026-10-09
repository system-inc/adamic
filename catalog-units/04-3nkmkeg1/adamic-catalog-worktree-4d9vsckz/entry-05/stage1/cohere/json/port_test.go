package json

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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

type run struct {
	stdout, stderr []byte
	exitCode       int
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

func execute(t *testing.T, environment []string, name string, arguments ...string) run {
	t.Helper()
	result, err := executeResult(environment, name, arguments...)
	if err != nil {
		t.Fatalf("running %s: %v", name, err)
	}
	return result
}

func executeResult(environment []string, name string, arguments ...string) (run, error) {
	command := exec.Command(name, arguments...)
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := childguard.Run(command, jsonGuard)
	var exitError *exec.ExitError
	if err != nil && !errors.As(err, &exitError) {
		return run{}, err
	}
	return run{stdout: stdout.Bytes(), stderr: stderr.Bytes(), exitCode: command.ProcessState.ExitCode()}, nil
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

var portFiles = []string{"parser.ts", "doc.ts", "formatter.ts", "main.ts", "width.ts", "widthTables.ts", "identifierTables.ts"}

func portDirectory(t *testing.T, mutation *printerMutation) string {
	t.Helper()
	directory := t.TempDir()
	for _, name := range portFiles {
		source, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		text := string(source)
		if mutation != nil && mutation.file == name {
			if strings.Count(text, mutation.from) != 1 {
				t.Fatalf("mutant %s must change one place", mutation.name)
			}
			text = strings.Replace(text, mutation.from, mutation.to, 1)
		}
		if err := os.WriteFile(filepath.Join(directory, name), []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return directory
}
func escape(text string) string {
	return strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t").Replace(text)
}
func protocol(cases []textCase, answers []answer) (string, string) {
	var inputs, outputs strings.Builder
	for index, item := range cases {
		fmt.Fprintf(&inputs, "%s\t%s\n", item.Name, escape(item.Text))
		value := answers[index]
		if value.Error != "" {
			fmt.Fprintf(&outputs, "error\t%s\n", escape(value.Error))
		} else {
			fmt.Fprintf(&outputs, "ok\t%s\n", escape(value.Output))
		}
	}
	return inputs.String(), outputs.String()
}
func compare(t *testing.T, name string, result run, expected string, cases []textCase) {
	t.Helper()
	if err := comparisonError(name, result, expected, cases); err != nil {
		t.Fatal(err)
	}
}
func comparisonError(name string, result run, expected string, cases []textCase) error {
	if result.exitCode != 0 || len(result.stderr) != 0 {
		return fmt.Errorf("%s exits %d: %s", name, result.exitCode, result.stderr)
	}
	if string(result.stdout) == expected {
		return nil
	}
	got, want := strings.Split(string(result.stdout), "\n"), strings.Split(expected, "\n")
	for index := 0; index < len(got) && index < len(want); index++ {
		if got[index] != want[index] {
			label := "end"
			if index < len(cases) {
				label = cases[index].Name
			}
			a, b := got[index], want[index]
			first := 0
			for first < len(a) && first < len(b) && a[first] == b[first] {
				first++
			}
			start := max(0, first-80)
			endA := min(len(a), first+200)
			endB := min(len(b), first+200)
			return fmt.Errorf("%s case %d %s byte %d: got %q; Go %q", name, index, label, first, a[start:endA], b[start:endB])
		}
	}
	return fmt.Errorf("%s output length %d, Go %d", name, len(result.stdout), len(expected))
}
func TestPortMatchesGoCohere(t *testing.T) {
	t.Parallel()
	cases := sampledCorpusCases(t, 32)
	// The Go oracle is mandatory even without an optional Prettier installation.
	goAnswers, _ := cohereAnswers(t, cases, false)
	input, expected := protocol(cases, goAnswers)
	path := filepath.Join(t.TempDir(), "cases.txt")
	if artifacts := os.Getenv("ADAMIC_JSON_ARTIFACTS"); artifacts != "" {
		path = filepath.Join(artifacts, "cases.txt")
		writeJSON(t, filepath.Join(artifacts, "cases.json"), cases)
		writeJSON(t, filepath.Join(artifacts, "go.json"), goAnswers)
	}
	if err := os.WriteFile(path, []byte(input), 0644); err != nil {
		t.Fatal(err)
	}
	source := portDirectory(t, nil)
	entry := filepath.Join(source, "main.ts")
	started := time.Now()
	nodeRun := onNode(t, entry, "--cases", path)
	nodeTime := time.Since(started)
	if artifacts := os.Getenv("ADAMIC_JSON_ARTIFACTS"); artifacts != "" {
		os.WriteFile(filepath.Join(artifacts, "node.txt"), nodeRun.stdout, 0644)
	}
	compare(t, "Node", nodeRun, expected, cases)
	if os.Getenv("ADAMIC_JSON_NODE_ONLY") != "" {
		t.Skip("debug run requested Node only; no native parity claim")
	}
	program := lowered(t, entry)
	release := filepath.Join(t.TempDir(), "release")
	if err := native.Build(native.C(program), release, native.Options{}); err != nil {
		t.Fatal(err)
	}
	single := execute(t, nil, release, "--cases", path)
	compare(t, "release", single, expected, cases)
	binary := filepath.Join(t.TempDir(), "sanitized")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	chunks := nativeChunks(t, cases, goAnswers)
	sanitized := runNativeChunks(t, binary, []string{"ASAN_OPTIONS=detect_leaks=0"}, chunks, single, cases, "native ASan/UBSan")
	compare(t, "native ASan/UBSan", sanitized, expected, cases)
	if runtime.GOOS == "linux" {
		leaked := runNativeChunks(t, binary, []string{"ASAN_OPTIONS=detect_leaks=1"}, chunks, single, cases, "LeakSanitizer")
		compare(t, "LeakSanitizer", leaked, expected, cases)
	} else {
		// Preserve macOS's allocation-counting/leaks tool check for every chunk.
		for index, chunk := range chunks {
			if report := leaks(t, program, binary, "--cases", chunk.path); report != "" {
				t.Fatalf("leaks chunk %d: %s", index, report)
			}
		}
	}
	compare(t, "JavaScript backend", onJavaScriptBackend(t, program, "--cases", path), expected, cases)
	if os.Getenv("ADAMIC_JSON_GUARD_CALIBRATE") == "1" {
		calibrateNativeCases(t, binary, cases, goAnswers)
	}
	if os.Getenv("ADAMIC_JSON_BENCH") == "1" {
		goDriver := buildGoDriver(t)
		for round := 0; round < 3; round++ {
			started = time.Now()
			result := execute(t, nil, release, "--cases", path)
			elapsed := time.Since(started)
			compare(t, "release", result, expected, cases)
			t.Logf("round %d native %d texts in %.3fs: %.2f texts/s", round, len(cases), elapsed.Seconds(), float64(len(cases))/elapsed.Seconds())
			started = time.Now()
			result = onNode(t, entry, "--cases", path)
			elapsed = time.Since(started)
			compare(t, "Node timed", result, expected, cases)
			t.Logf("round %d Node %d texts in %.3fs: %.2f texts/s", round, len(cases), elapsed.Seconds(), float64(len(cases))/elapsed.Seconds())
			started = time.Now()
			result = execute(t, nil, goDriver, "--cases", path)
			elapsed = time.Since(started)
			compare(t, "Go timed", result, expected, cases)
			t.Logf("round %d Go %d texts in %.3fs: %.2f texts/s", round, len(cases), elapsed.Seconds(), float64(len(cases))/elapsed.Seconds())
		}
	} else {
		t.Log("timing skipped; set ADAMIC_JSON_BENCH=1 for three timing rounds")
	}
	t.Logf("initial Node %.3fs; %d texts identical on Go, Node, native and JS backend", nodeTime.Seconds(), len(cases))
}

func TestThreePortMutantsAreCaught(t *testing.T) {
	t.Parallel()
	cases := []textCase{
		{"probe.json", `{"a":1,"b":[2,3]}`},
		{"package.json", `{"a":1,"b":[2,3]}`},
		{"package.json", `{"text":"é😀","empty":[]}`},
		{"probe.json", `[1__0]`},
	}
	answers, _ := cohereAnswers(t, cases, false)
	input, expected := protocol(cases, answers)
	path := filepath.Join(t.TempDir(), "cases.txt")
	if err := os.WriteFile(path, []byte(input), 0644); err != nil {
		t.Fatal(err)
	}
	for _, mutation := range []printerMutation{
		{"missing final newline", "formatter.ts", "printer.docs.concat([document, printer.docs.line(false, true)])", "document"},
		{"missing colon space", "formatter.ts", "documents.text(': ')", "documents.text(':')"},
		{"wrong filename parser", "formatter.ts", "base === 'package.json'", "base === 'other-package.json'"},
	} {
		t.Run(mutation.name, func(t *testing.T) {
			t.Parallel()
			directory := portDirectory(t, &mutation)
			entry := filepath.Join(directory, "main.ts")
			program := lowered(t, entry)
			for _, side := range []struct {
				name   string
				result run
			}{
				{"Node", onNode(t, entry, "--cases", path)},
				{"native", nativelyRun(t, program, "--cases", path)},
			} {
				if side.result.exitCode != 0 || len(side.result.stderr) != 0 {
					t.Fatalf("%s mutant must succeed: exit %d stderr %s", side.name, side.result.exitCode, side.result.stderr)
				}
				if string(side.result.stdout) == expected {
					t.Fatalf("%s failed to catch %s", side.name, mutation.name)
				}
				got, want := strings.Split(string(side.result.stdout), "\n"), strings.Split(expected, "\n")
				count := 0
				for index := range cases {
					if got[index] != want[index] {
						count++
						t.Logf("%s caught %s on %s: %q; Go %q", side.name, mutation.name, cases[index].Name, got[index], want[index])
					}
				}
				if count == 0 {
					t.Fatal("difference must be in formatter answers")
				}
			}
		})
	}
}

func TestAdditionalJSONBoundaries(t *testing.T) {
	t.Parallel()
	texts := []string{
		"[\n// dangling\n]",
		`[1.0,1.000,.000,1e0_0,1_0e+001,1.0_0,1.e0,1.000e10]`,
		`[1e]`, `[1e+]`, `[0x]`, `[0o8]`, `[.]`, `[01]`, `[1n]`,
		`[,]`, `[1,,]`, `[,,2]`, `{1:2,1.50:3,0x10:4}`, `{'a':'can\'t',b:'"hi"'}`,
		"`raw`", "`\\u0041\\n\\x42`", "`line\nline`", "{é:1,𐌘:2}",
		"// header\n\n{\n// before\n a:1, // after\n b:[1,2]\n}\n// footer",
		"{a:1,/* before */b:2}", "{/* dangling */}", "[1,// after\n2]",
		`"\xG0"`, `"\u{}"`, `"\u{110000}"`, `"\uD800"`, `"\uD83D\uDE00"`,
	}
	for _, literal := range []string{"中文", "ｅ́", "🇺🇸", "👨‍👩‍👧", "1️⃣", "❤️", "©", "👍🏽", "𐌘"} {
		texts = append(texts, `{"text":"`+strings.Repeat(literal, 25)+`","items":[1,2,3]}`)
	}
	var cases []textCase
	for index, text := range texts {
		for _, name := range []string{"probe.json", "package.json"} {
			cases = append(cases, textCase{fmt.Sprintf("boundary/%d/%s", index, name), text})
		}
	}
	answers, _ := cohereAnswers(t, cases, false)
	input, expected := protocol(cases, answers)
	path := filepath.Join(t.TempDir(), "cases.txt")
	if artifacts := os.Getenv("ADAMIC_JSON_ARTIFACTS"); artifacts != "" {
		path = filepath.Join(artifacts, "boundary.txt")
		writeJSON(t, filepath.Join(artifacts, "boundary.json"), cases)
		writeJSON(t, filepath.Join(artifacts, "boundary-go.json"), answers)
	}
	if err := os.WriteFile(path, []byte(input), 0644); err != nil {
		t.Fatal(err)
	}
	source := portDirectory(t, nil)
	entry := filepath.Join(source, "main.ts")
	result := onNode(t, entry, "--cases", path)
	if artifacts := os.Getenv("ADAMIC_JSON_ARTIFACTS"); artifacts != "" {
		os.WriteFile(filepath.Join(artifacts, "boundary-node.txt"), result.stdout, 0644)
	}
	compare(t, "Node boundaries", result, expected, cases)
	if os.Getenv("ADAMIC_JSON_NODE_ONLY") != "" {
		t.Skip("debug Node only")
	}
	program := lowered(t, entry)
	result, binary := natively(t, program, "--cases", path)
	compare(t, "native boundaries", result, expected, cases)
	if report := leaks(t, program, binary, "--cases", path); report != "" {
		t.Fatal(report)
	}
}

func buildGoDriver(t *testing.T) string {
	t.Helper()
	directory := t.TempDir()
	cohere, err := filepath.Abs(filepath.Join(repository, "cohere"))
	if err != nil {
		t.Fatal(err)
	}
	source, err := filepath.Abs("testdata/cohere_driver.go")
	if err != nil {
		t.Fatal(err)
	}
	overlay := filepath.Join(directory, "overlay.json")
	writeJSON(t, overlay, map[string]any{"Replace": map[string]string{filepath.Join(cohere, "command/formatter_comparison/main.go"): source}})
	binary := filepath.Join(directory, "go-cohere")
	command := bounded(t, "go", "build", "-overlay="+overlay, "-o", binary, "./command/formatter_comparison")
	command.Dir = cohere
	if output, err := childguard.CombinedOutput(command, jsonGuard); err != nil {
		t.Fatalf("Go driver: %v\n%s", err, output)
	}
	return binary
}

func TestSingleFileStdoutDriver(t *testing.T) {
	t.Parallel()
	cases := []textCase{{"probe.json", `{"text":"é😀","items":[1,2,3]}`}, {"package.json", `{"text":"é😀","items":[1,2,3]}`}}
	answers, _ := cohereAnswers(t, cases, false)
	directory := portDirectory(t, nil)
	entry := filepath.Join(directory, "main.ts")
	program := lowered(t, entry)
	binary := filepath.Join(t.TempDir(), "port")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	for index, item := range cases {
		path := filepath.Join(t.TempDir(), item.Name)
		if err := os.WriteFile(path, []byte(item.Text), 0644); err != nil {
			t.Fatal(err)
		}
		compare(t, "Node file driver", onNode(t, entry, path), answers[index].Output, cases)
		compare(t, "native file driver", execute(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary, path), answers[index].Output, cases)
	}
}
