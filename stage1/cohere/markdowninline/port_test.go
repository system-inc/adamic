package markdowninline

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

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
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

// bounded is a command that can't outlive its test: it has a deadline, it runs in a process group of
// its own, and when the deadline passes or the test ends, the whole group is killed.
func bounded(t *testing.T, name string, arguments ...string) *exec.Cmd {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	t.Cleanup(cancel)
	command := exec.CommandContext(ctx, name, arguments...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		return syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
	}
	command.WaitDelay = 5 * time.Second
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
	err := command.Run()
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

func TestMarkdownInline(t *testing.T) {
	t.Parallel()
	library := os.Getenv("ADAMIC_MARKDOWNINLINE_LIBRARY")
	if library == "" {
		t.Skip("ADAMIC_MARKDOWNINLINE_LIBRARY unset; #xq2ecw6 (setup --gate-inputs) provisions this oracle")
	}
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	var texts []string
	var paths []string
	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if entry.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		switch strings.ToLower(filepath.Ext(path)) {
		case ".md", ".markdown", ".mdown", ".mkd":
			b, e := os.ReadFile(path)
			if e != nil {
				return e
			}
			texts = append(texts, string(b))
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	files := len(texts)
	alphabet := []string{"a", "*", "_", "\\", "`", " ", "\n", "😀"}
	var generate func(string, int)
	generate = func(s string, n int) {
		texts = append(texts, s)
		if n > 0 {
			for _, c := range alphabet {
				generate(s+c, n-1)
			}
		}
	}
	generate("", 4)
	texts = append(texts, "===", "---", "\r\t\u2028\u2029", "\x00", " \n ", " ` ``, ``` ", "a|b\n", "'\"< >)[]\\", "\t\n\t x", strings.Repeat("\\\\**a _b_ `c` |d| 😀 ", 10000))
	// Every BMP scalar is examined on both sides of both delimiter kinds. Surrogate units are
	// exercised by astral scalars; isolated surrogates cannot be represented by the UTF-8 file API.
	var unicode strings.Builder
	for code := rune(0); code <= 0xffff; code++ {
		if code >= 0xd800 && code <= 0xdfff {
			continue
		}
		for _, marker := range []string{"*", "_"} {
			unicode.WriteRune(code)
			unicode.WriteString(marker + "a a" + marker)
			unicode.WriteRune(code)
			unicode.WriteString("; ")
		}
	}
	texts = append(texts, unicode.String())
	for _, code := range []rune{0x10100, 0x1039f, 0x1f600, 0x1e95f, 0x10ffff} {
		texts = append(texts, string(code)+"_a a_"+string(code), string(code)+"*a a*"+string(code))
	}
	modes := "wefnspctrukvhijlboq"
	var input strings.Builder
	encode := strings.NewReplacer(`\`, `\\`, "\n", `\n`, "\r", `\r`, "\t", `\t`)
	for _, text := range texts {
		for _, mode := range modes {
			input.WriteString(string(mode) + encode.Replace(text) + "\n")
		}
	}
	cases := filepath.Join(dir, "cases.txt")
	write(t, cases, []byte(input.String()))
	manifest, _ := json.MarshalIndent(paths, "", "  ")
	if keep := os.Getenv("ADAMIC_MARKDOWNINLINE_KEEP"); keep != "" {
		write(t, keep, []byte(input.String()))
		write(t, keep+".files.json", manifest)
	}
	bridge, _ := filepath.Abs("testdata/bridge.go")
	driver, _ := filepath.Abs("testdata/go_driver.go")
	cohere := filepath.Join(root, "cohere")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{
		filepath.Join(cohere, "internal/format/markdown/adamic_stage_one.go"): bridge,
		filepath.Join(cohere, "cmd/adamic_stage_one/main.go"):                 driver,
	}})
	overlayPath := filepath.Join(dir, "overlay.json")
	write(t, overlayPath, overlay)
	goBinary := filepath.Join(dir, "go-printer")
	command := bounded(t, "go", "build", "-overlay="+overlayPath, "-o", goBinary, filepath.Join(cohere, "cmd/adamic_stage_one/main.go"))
	command.Dir = cohere
	if output, err := command.CombinedOutput(); err != nil {
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
	script, _ := filepath.Abs("testdata/library.mjs")
	if library != "" {
		answer := execute(t, nil, "node", script, library, cases)
		clean(t, "Prettier", answer)
		equal(t, "Prettier", answer.stdout, want.stdout)
	}
	t.Logf("Go/native/Node/JavaScript backend/Prettier batch byte parity passed (%d cases)", len(texts)*len(modes))
	// Independently verify raw stdout preserves every repository text and terminal newlines.
	rawTexts := append(append([]string{}, texts[:files]...), "", "a", "a\n", "a\r\n\n", "😀*x*", "\\_")
	for _, text := range rawTexts {
		path := filepath.Join(dir, "raw.md")
		write(t, path, []byte(text))
		singleCase := filepath.Join(dir, "single.txt")
		write(t, singleCase, []byte("w"+encode.Replace(text)+"\n"))
		expected := execute(t, nil, goBinary, singleCase)
		clean(t, "Go raw", expected)
		decoded := strings.NewReplacer(`\n`, "\n", `\r`, "\r", `\t`, "\t", `\\`, `\`).Replace(strings.TrimSuffix(string(expected.stdout), "\n"))
		for _, r := range []run{execute(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary, path, "w")} {
			clean(t, "raw", r)
			equal(t, "raw", r.stdout, []byte(decoded))
		}
	}
	for _, mutation := range []struct{ name, from, to string }{
		{"escaped delimiter parity", "(found.preceding - position) % 2 === 1", "(found.preceding - position) % 2 === 0"},
		{"table pipe escaping", "if(table)", "if(!table)"},
		{"minimum absent fence", "while(runs.includes(count))", "while(runs.includes(count) && count < 1)"},
	} {
		t.Run(mutation.name, func(t *testing.T) {
			parent := t.TempDir()
			scratch := filepath.Join(parent, "markdowninline")
			if err := os.Mkdir(scratch, 0755); err != nil {
				t.Fatal(err)
			}
			dependency, err := os.ReadFile("classes.ts")
			if err != nil {
				t.Fatal(err)
			}
			write(t, filepath.Join(scratch, "classes.ts"), dependency)
			source, err := os.ReadFile("inline.ts")
			if err != nil {
				t.Fatal(err)
			}
			if strings.Count(string(source), mutation.from) != 1 {
				t.Fatal("mutant must change one site")
			}
			write(t, filepath.Join(scratch, "inline.ts"), []byte(strings.Replace(string(source), mutation.from, mutation.to, 1)))
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
		t.Logf("throughput %s %.1f texts/s, 3 process runs %.6fs (startup, input, escaping, output included; build excluded)", side.name, float64(len(texts)*len(modes)*3)/elapsed.Seconds(), elapsed.Seconds())
	}
	t.Logf("corpus: %d repository/submodule files, %d generated texts, %d leaf/context cases; all byte-identical", files, len(texts)-files, len(texts)*len(modes))
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
