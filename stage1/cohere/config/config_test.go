// The command discovery port is held to the actual Go command through an overlay, without editing it.
package config

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"
)

const repository = "../../.."

func portDirectory(t *testing.T, file, from, to string) string {
	t.Helper()
	root := t.TempDir()
	for directory, names := range map[string][]string{
		"config":      {"glob.ts", "discovery.ts", "main.ts", "json.ts", "settings.ts", "sets.ts", "tsglob.ts", "tsconfig.ts", "loader_main.ts", "version.ts", "house.ts", "quote_table.ts"},
		"formatfiles": {"disk.ts", "golang.ts"},
		"gitignore":   {"gitignore.ts", "glob.ts", "path.ts"},
	} {
		if err := os.Mkdir(filepath.Join(root, directory), 0755); err != nil {
			t.Fatal(err)
		}
		for _, name := range names {
			original := filepath.Join("..", directory, name)
			content, err := os.ReadFile(original)
			if err != nil {
				t.Fatal(err)
			}
			text := string(content)
			if directory == "config" && name == file {
				if strings.Count(text, from) != 1 {
					t.Fatalf("mutant must replace one occurrence of %q", from)
				}
				text = strings.Replace(text, from, to, 1)
			}
			if err := os.WriteFile(filepath.Join(root, directory, name), []byte(text), 0644); err != nil {
				t.Fatal(err)
			}
		}
	}
	return filepath.Join(root, "config", "main.ts")
}

type run struct {
	stdout, stderr []byte
	exitCode       int
}

func goCohere(t *testing.T, input string) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join(repository, "cohere"))
	if err != nil {
		t.Fatal(err)
	}
	side, err := filepath.Abs("testdata/cohere_side_test.go.txt")
	if err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(t.TempDir(), "answers")
	overlay := filepath.Join(t.TempDir(), "overlay.json")
	contents, err := json.Marshal(map[string]any{"Replace": map[string]string{filepath.Join(root, "command/cohere/adamic_config_test.go"): side}})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(overlay, contents, 0644); err != nil {
		t.Fatal(err)
	}
	command := bounded(t, "go", "test", "-count=1", "-overlay="+overlay, "-v", "-run=^TestAdamicConfigCases$", "./command/cohere")
	command.Dir = root
	command.Env = append(os.Environ(), "ADAMIC_CONFIG_INPUT="+input, "ADAMIC_CONFIG_OUTPUT="+output)
	if got, err := command.CombinedOutput(); err != nil {
		t.Fatalf("Go oracle: %v\n%s", err, got)
	} else {
		t.Logf("%s", got)
	}
	result, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	return string(result)
}

func cases(t *testing.T) string {
	t.Helper()
	var input strings.Builder
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	for _, project := range []string{root, filepath.Join(root, "cohere"), filepath.Join(root, "cohere/TypeScript")} {
		fmt.Fprintln(&input, project)
		fmt.Fprintf(&input, "configured\t%s\n", project)
	}
	write := func(root, name, text string) {
		t.Helper()
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
	}
	link := func(root, name, target string) {
		t.Helper()
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(target, path); err != nil {
			t.Fatal(err)
		}
	}
	fixed := t.TempDir()
	for name, text := range map[string]string{
		".git/HEAD": "", ".gitignore": "ignored/\n*.tmp/\n!keep.tmp/\n", "tsconfig.json": "{}", "Package.swift": "",
		"ignored/tsconfig.json": "{}", "keep.tmp/tsconfig.json": "{}", "drop.tmp/tsconfig.json": "{}",
		"apps/.gitignore": "legacy/\n!live/\n", "apps/legacy/tsconfig.json": "{}", "apps/live/tsconfig.json": "{}",
		"apps/live/Package.swift": "", "apps/live/child/tsconfig.json": "{}", "vendor/.git/HEAD": "", "vendor/tsconfig.json": "{}",
		"linked/.git": "gitdir: elsewhere", "linked/tsconfig.json": "{}", "node_modules/tsconfig.json": "{}", "testdata/tsconfig.json": "{}",
		".cache/tsconfig.json": "{}", ".build/Package.swift": "", "a-b/tsconfig.json": "{}", "a/b/tsconfig.json": "{}", "a/tsconfig.json": "{}",
		"é/tsconfig.json": "{}", "😀/tsconfig.json": "{}", "ａ/tsconfig.json": "{}",
	} {
		write(fixed, name, text)
	}
	link(fixed, "alias", "apps/live")
	link(fixed, "broken/tsconfig.json", "missing")
	link(fixed, "loop", "loop")
	link(fixed, "git-link/.git", "missing")
	write(fixed, "git-link/tsconfig.json", "{}")
	fmt.Fprintln(&input, fixed)
	fmt.Fprintf(&input, "%s\tapps/**/{tsconfig.json,Package.swift}\tkeep.tmp/*\n", fixed)
	fmt.Fprintln(&input, filepath.Join(fixed, "apps"))
	fmt.Fprintln(&input, filepath.Join(fixed, "missing"))
	refused := t.TempDir()
	link(refused, ".gitignore", "missing")
	fmt.Fprintln(&input, refused)
	configured := t.TempDir()
	write(configured, "tsconfig.json", "{}")
	write(configured, "base.json", `{"ignorePatterns":["base/**"]}`)
	write(configured, "CohereSettings.json", `{"extends":"./base.json","ignorePatterns":["child/**"],"rules":{"bad":9}}`)
	write(configured, "base/tsconfig.json", "{}")
	write(configured, "child/tsconfig.json", "{}")
	write(configured, "keep/tsconfig.json", "{}")
	fmt.Fprintf(&input, "configured\t%s\n", configured)
	random := rand.New(rand.NewSource(20261006))
	names := []string{"a", "a-b", "src", "ignored", "keep", "node_modules", ".cache", "testdata", "é", "😀", "ａ"}
	for number := 0; number < 120; number++ {
		tree := t.TempDir()
		write(tree, ".git/HEAD", "")
		write(tree, ".gitignore", "ignored/\n*drop*\n!keep/\n")
		for index := 0; index < 12; index++ {
			directory := names[random.Intn(len(names))]
			if random.Intn(2) == 0 {
				directory += "/" + names[random.Intn(len(names))]
			}
			marker := "tsconfig.json"
			if random.Intn(3) == 0 {
				marker = "Package.swift"
			}
			write(tree, filepath.Join(directory, marker), "{}")
			if random.Intn(9) == 0 {
				write(tree, filepath.Join(directory, ".git/HEAD"), "")
			}
			if random.Intn(7) == 0 {
				write(tree, filepath.Join(directory, ".gitignore"), "src/\n!keep/\n")
			}
		}
		fmt.Fprintf(&input, "%s\t**/ignored/**\n", tree)
	}
	patterns := []string{"", "*", "?", "??", "**", "**/*.ts", "a/*", "a/**", "a/**/b", "{a,{b,c}}/**", "*.{ts,tsx}", "{a,b", "a*?b", "**//x", "\\*", "[a]", "/a/", "??.ts"}
	paths := []string{"", "a", "b", "c", "a/b", "a/x/b", "a/x/y/b", "x.ts", "é.ts", "😀", "a/b.ts", "/a/", "a//x", "[a]", "ab", "axxb"}
	for _, pattern := range patterns {
		for _, path := range paths {
			fmt.Fprintf(&input, "glob\t%s\t%s\n", pattern, path)
		}
	}
	path := filepath.Join(t.TempDir(), "cases")
	if err := os.WriteFile(path, []byte(input.String()), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestDiscoveryMatchesGoCohere(t *testing.T) {
	t.Parallel()
	input := cases(t)
	want := goCohere(t, input)
	path := portDirectory(t, "", "", "")
	program := lowered(t, path)
	result, binary := natively(t, program, input)
	for _, side := range []struct {
		name   string
		result run
	}{{"native", result}, {"Node", onNode(t, path, input)}, {"JavaScript backend", onJavaScriptBackend(t, program, input)}} {
		if side.result.exitCode != 0 || len(side.result.stderr) > 0 {
			t.Fatalf("%s: exit %d, stderr %s", side.name, side.result.exitCode, side.result.stderr)
		}
		if difference := firstDifference(string(side.result.stdout), want); difference != "" {
			t.Errorf("%s: %s", side.name, difference)
		}
	}
	if report := leaks(t, program, binary, input); report != "" {
		t.Error(report)
	}
	if t.Failed() {
		t.Fatal("baseline must agree before mutants can prove the comparisons")
	}
	t.Logf("%d roots, %d projects, %d glob queries", strings.Count(want, "root "), strings.Count(want, "project "), strings.Count(want, "glob "))
	for _, mutant := range []struct{ name, file, from, to string }{
		{"directory pruning", "discovery.ts", "if (scope.ignored(fromRepository(repository, root, child), true)[0])", "if (false)"},
		{"zero segments in double star", "glob.ts", "let index = pathIndex; index <= path.length", "let index = pathIndex + 1; index <= path.length"},
		{"nested repository boundary", "discovery.ts", "if (name === '.git' && relative !== '.')", "if (name === '.git' && relative === 'unreachable')"},
	} {
		t.Run("catches "+mutant.name, func(t *testing.T) {
			source := portDirectory(t, mutant.file, mutant.from, mutant.to)
			modified := lowered(t, source)
			for _, side := range []struct {
				name   string
				result run
			}{{"native", nativelyRun(t, modified, input)}, {"Node", onNode(t, source, input)}} {
				if side.result.exitCode != 0 {
					t.Fatalf("%s mutant failed to execute: %s", side.name, side.result.stderr)
				}
				difference := firstDifference(string(side.result.stdout), want)
				if difference == "" {
					t.Fatal("mutant survived")
				}
				t.Logf("%s caught: %s", side.name, difference)
			}
		})
	}
	if os.Getenv("ADAMIC_CONFIG_TIMING") != "" {
		binary := filepath.Join(t.TempDir(), "port")
		if err := native.Build(native.C(program), binary, native.Options{}); err != nil {
			t.Fatal(err)
		}
		for round := 0; round < 3; round++ {
			start := time.Now()
			got := execute(t, nil, binary, input)
			elapsed := time.Since(start)
			if string(got.stdout) != want || got.exitCode != 0 {
				t.Fatal("timed native differs")
			}
			t.Logf("native round %d: %.6fs, %.0f projects/s", round, elapsed.Seconds(), float64(strings.Count(want, "project "))/elapsed.Seconds())
		}
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
