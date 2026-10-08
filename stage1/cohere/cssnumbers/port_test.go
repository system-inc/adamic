package cssnumbers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/corpusfiles"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

const repository = "../../.."

// run is one execution's observable behavior.
type run struct {
	stdout   []byte
	stderr   []byte
	exitCode int
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

// leaks returns a report of everything the finished port never let go of, or "": the leak check the
// oracle runs on every fixture (internal/leakcheck), on the sanitized binary and the port's C.
func leaks(t *testing.T, program *ir.Program, sanitized string, arguments ...string) string {
	t.Helper()
	return leakcheck.Report(t, native.C(program), sanitized, arguments...)
}

func TestCSSNumbers(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	var texts []string
	paths := corpusfiles.Upstream(t, filepath.Join(root, "cohere"), corpusfiles.CohereCommit, []string{"internal/format/css/testdata/prettier", "internal/lint/rules/tailwind"}, []string{"*.css", "*.scss", "*.less"})
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		texts = append(texts, string(data))
	}
	files := len(texts)
	alphabet := []string{"a", "0", "1", ".", "e", "-", "\"", "\\"}
	var generate func(string, int)
	generate = func(s string, n int) {
		texts = append(texts, s)
		if n > 0 {
			for _, c := range alphabet {
				generate(s+c, n-1)
			}
		}
	}
	generate("", 5)
	texts = append(texts, "\r\t\u2028\u2029", strings.Repeat("'😀\\\"x' ", 10000), "'never closed\\", `"\'"`, "\x00'null'")
	for _, mantissa := range []string{"0", "01", ".0", ".00100", "1.", "1.000", "12.34000"} {
		for _, exponent := range []string{"", "e0", "E+000", "e-000", "e+001", "e-002", "E00020", "e+", "e-"} {
			for _, unit := range strings.Split("|n|EM|Q|HZ|kHZ|px|cQmAX|unknown|foo|é|Σ|İ", "|") {
				for _, prefix := range []string{"", "a", "$x", "@foo", "-", "+", "😀", "'", "\""} {
					texts = append(texts, prefix+mantissa+exponent+unit)
				}
			}
		}
	}
	for _, unit := range strings.Split("em|rem|ex|rex|cap|rcap|ch|rch|ic|ric|lh|rlh|vw|svw|lvw|dvw|vh|svh|lvh|dvh|vi|svi|lvi|dvi|vb|svb|lvb|dvb|vmin|svmin|lvmin|dvmin|vmax|svmax|lvmax|dvmax|cm|mm|Q|in|pt|pc|px|deg|grad|rad|turn|s|ms|Hz|kHz|dpi|dpcm|dppx|x|cqw|cqh|cqi|cqb|cqmin|cqmax|fr", "|") {
		texts = append(texts, ".1000E+002"+strings.ToUpper(unit))
	}
	texts = append(texts, "'1.000px'", "1.000unknown", "a1.000px", "😀1.000px", strings.Repeat(".000100E-002KHZ ", 10000))
	var input strings.Builder
	encode := strings.NewReplacer(`\`, `\\`, "\n", `\n`, "\r", `\r`, "\t", `\t`)
	for _, text := range texts {
		for _, preference := range []string{"d", "s"} {
			input.WriteString(preference + encode.Replace(text) + "\n")
		}
	}
	cases := filepath.Join(dir, "cases.txt")
	write(t, cases, []byte(input.String()))
	manifest, _ := json.MarshalIndent(paths, "", "  ")
	if keep := os.Getenv("ADAMIC_CSSNUMBERS_KEEP"); keep != "" {
		write(t, keep, []byte(input.String()))
		write(t, keep+".files.json", manifest)
	}
	bridge, _ := filepath.Abs("testdata/bridge.go")
	driver, _ := filepath.Abs("testdata/go_driver.go")
	cohere := filepath.Join(root, "cohere")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{
		filepath.Join(cohere, "internal/format/css/adamic_stage_one.go"): bridge,
		filepath.Join(cohere, "cmd/adamic_stage_one/main.go"):            driver,
	}})
	overlayPath := filepath.Join(dir, "overlay.json")
	write(t, overlayPath, overlay)
	goBinary := filepath.Join(dir, "go-printer")
	command := bounded(t, "go", "build", "-overlay="+overlayPath, "-o", goBinary, filepath.Join(cohere, "cmd/adamic_stage_one/main.go"))
	command.Dir = cohere
	if output, err := childguard.CombinedOutput(command, childguard.Options{}); err != nil {
		t.Fatalf("Go bridge: %v\n%s", err, output)
	}
	want := execute(t, nil, goBinary, cases)
	clean(t, "Go", want)
	main, _ := filepath.Abs("main.ts")
	program := lowered(t, main)
	got, binary := natively(t, program, "--batch", cases)
	for _, side := range []struct {
		name   string
		result run
	}{{"native", got}, {"Node", onNode(t, main, "--batch", cases)}, {"JavaScript backend", onJavaScriptBackend(t, program, "--batch", cases)}} {
		clean(t, side.name, side.result)
		equal(t, side.name, side.result.stdout, want.stdout)
	}
	if report := leaks(t, program, binary, "--batch", cases); report != "" {
		t.Fatal(report)
	}
	library := os.Getenv("ADAMIC_CSSNUMBERS_LIBRARY")
	script, _ := filepath.Abs("testdata/library.mjs")
	if library != "" {
		answer := execute(t, nil, "node", script, library, cases)
		clean(t, "Prettier", answer)
		equal(t, "Prettier", answer.stdout, want.stdout)
	} else {
		t.Log("external library not checked: set ADAMIC_CSSNUMBERS_LIBRARY")
	}
	// Raw stdout must preserve empty input and terminal newlines, independently of batch escaping.
	rawTexts := append(append([]string{}, texts[:files]...), "", "a", "a\n", "a\r\n\n", "'😀'", "\"\\'\"")
	for _, text := range rawTexts {
		path := filepath.Join(dir, "raw.css")
		write(t, path, []byte(text))
		for _, single := range []bool{false, true} {
			args := []string{path}
			if single {
				args = append(args, "--single")
			}
			singleCase := filepath.Join(dir, "single.txt")
			pref := "d"
			if single {
				pref = "s"
			}
			write(t, singleCase, []byte(pref+encode.Replace(text)+"\n"))
			expected := execute(t, nil, goBinary, singleCase)
			clean(t, "Go raw", expected)
			decoded := strings.NewReplacer(`\n`, "\n", `\r`, "\r", `\t`, "\t", `\\`, `\`).Replace(strings.TrimSuffix(string(expected.stdout), "\n"))
			for _, r := range []run{execute(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary, args...), onNode(t, main, args...)} {
				clean(t, "raw", r)
				equal(t, "raw", r.stdout, []byte(decoded))
			}
		}
	}
	for _, mutation := range []struct{ name, from, to string }{
		{"quoted numbers", "quotedEnd >= 0", "quotedEnd < -1"},
		{"unknown unit", "unit !== ''", "unit === ''"},
		{"word prefix", "!match.hasWordPart", "match.hasWordPart || !match.hasWordPart"},
	} {
		t.Run(mutation.name, func(t *testing.T) {
			parent := t.TempDir()
			scratch := filepath.Join(parent, "cssnumbers")
			if err := os.Mkdir(scratch, 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.Mkdir(filepath.Join(parent, "cssstrings"), 0755); err != nil {
				t.Fatal(err)
			}
			dependency, err := os.ReadFile("../cssstrings/strings.ts")
			if err != nil {
				t.Fatal(err)
			}
			write(t, filepath.Join(parent, "cssstrings/strings.ts"), dependency)
			source, err := os.ReadFile("numbers.ts")
			if err != nil {
				t.Fatal(err)
			}
			if strings.Count(string(source), mutation.from) != 1 {
				t.Fatal("mutant must change one site")
			}
			write(t, filepath.Join(scratch, "numbers.ts"), []byte(strings.Replace(string(source), mutation.from, mutation.to, 1)))
			source, err = os.ReadFile("main.ts")
			if err != nil {
				t.Fatal(err)
			}
			mutatedMain := filepath.Join(scratch, "main.ts")
			write(t, mutatedMain, source)
			mutated := lowered(t, mutatedMain)
			for _, r := range []run{nativelyRun(t, mutated, "--batch", cases), onNode(t, mutatedMain, "--batch", cases)} {
				clean(t, "mutant", r)
				if bytes.Equal(r.stdout, want.stdout) {
					t.Fatal("mutant survived")
				}
				a, b := strings.Split(string(r.stdout), "\n"), strings.Split(string(want.stdout), "\n")
				for i := range b {
					if i >= len(a) {
						t.Logf("caught by missing output case %d", i)
						break
					}
					if a[i] != b[i] {
						offset := 0
						for offset < len(a[i]) && offset < len(b[i]) && a[i][offset] == b[i][offset] {
							offset++
						}
						start := max(0, offset-20)
						t.Logf("caught by output case %d, byte %d: got %q want %q", i, offset,
							a[i][start:min(len(a[i]), offset+40)], b[i][start:min(len(b[i]), offset+40)])
						break
					}
				}
			}
		})
	}
	fast := filepath.Join(dir, "native-fast")
	if err := native.Build(native.C(program), fast, native.Options{}); err != nil {
		t.Fatal(err)
	}
	runner, _ := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	measurements := []struct {
		name    string
		command string
		args    []string
	}{{"Go", goBinary, []string{cases}}, {"native", fast, []string{"--batch", cases}}, {"Node", "node", []string{"--disable-warning=ExperimentalWarning", runner, main, "--batch", cases}}}
	if library != "" {
		measurements = append(measurements, struct {
			name    string
			command string
			args    []string
		}{"Prettier", "node", []string{script, library, cases}})
	}
	for _, side := range measurements {
		var elapsed time.Duration
		for round := 0; round < 3; round++ {
			start := time.Now()
			answer := execute(t, nil, side.command, side.args...)
			elapsed += time.Since(start)
			clean(t, side.name, answer)
			equal(t, side.name, answer.stdout, want.stdout)
		}
		t.Logf("throughput %s %.1f texts/s, 3 process runs %.6fs (startup, input, escaping, output included; build excluded)", side.name, float64(len(texts)*2*3)/elapsed.Seconds(), elapsed.Seconds())
	}
	t.Logf("corpus: %d repository/submodule files, %d generated texts, %d preference cases; all byte-identical", files, len(texts)-files, len(texts)*2)
}
func write(t *testing.T, path string, b []byte) {
	t.Helper()
	if err := os.WriteFile(path, b, 0644); err != nil {
		t.Fatal(err)
	}
}
func clean(t *testing.T, name string, r run) {
	t.Helper()
	if r.exitCode != 0 || len(r.stderr) > 0 {
		t.Fatalf("%s exit %d stderr %s", name, r.exitCode, r.stderr)
	}
}
func equal(t *testing.T, name string, a, b []byte) {
	t.Helper()
	if !bytes.Equal(a, b) {
		i := 0
		for i < len(a) && i < len(b) && a[i] == b[i] {
			i++
		}
		t.Fatalf("%s first byte difference at %d (lengths %d/%d)", name, i, len(a), len(b))
	}
}
