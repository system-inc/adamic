package native_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
)

func timerHarness(t *testing.T) string {
	t.Helper()
	source, err := os.ReadFile("testdata/timers/harness.c")
	if err != nil {
		t.Fatal(err)
	}
	return string(source)
}
func timerRun(environment []string, name string, args ...string) leakcheck.Run {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, name, args...)
	command.Env = append(os.Environ(), environment...)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	exit := 0
	if err != nil {
		exit = -1
		if command.ProcessState != nil {
			exit = command.ProcessState.ExitCode()
		}
	}
	return leakcheck.Run{Stdout: stdout.Bytes(), Stderr: stderr.Bytes(), ExitCode: exit}
}
func timerNode(t *testing.T, mode string) []byte {
	t.Helper()
	// .a is deliberate: no new TypeScript source files in an Adamic unit.
	run := timerRun(nil, "node", "--input-type=commonjs", "-e",
		`require('node:vm').runInThisContext(require('node:fs').readFileSync(process.argv[1], 'utf8'))`,
		"testdata/timers/oracle.a", mode)
	if run.ExitCode != 0 || len(run.Stderr) != 0 {
		t.Fatalf("Node: %+v", run)
	}
	return run.Stdout
}
func TestTimersAgainstNode(t *testing.T) {
	code := timerHarness(t)
	binary := filepath.Join(t.TempDir(), "timers")
	if err := native.Build(code, binary, native.Options{Sanitize: true, Count: true}); err != nil {
		t.Fatal(err)
	}
	released := filepath.Join(t.TempDir(), "timers-release")
	// The stack's optimized build; slice 4 adds the shipped release (Options.Release, ThinLTO).
	if err := native.Build(code, released, native.Options{}); err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"order", "interval", "cancel", "cancel-self", "foreign", "unref", "unref-live", "deadlines", "graph", "arity", "host"} {
		t.Run(mode, func(t *testing.T) {
			run := timerRun([]string{"ASAN_OPTIONS=detect_leaks=0:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary, mode)
			if run.ExitCode != 0 {
				t.Fatalf("native: %+v", run)
			}
			if report := leakcheck.Unbalanced(run); report != "" {
				t.Fatal(report)
			}
			release := timerRun(nil, released, mode)
			if release.ExitCode != 0 || len(release.Stderr) != 0 || !bytes.Equal(release.Stdout, run.Stdout) {
				t.Fatalf("release differs: %+v", release)
			}
			if !bytes.Equal(run.Stdout, timerNode(t, mode)) {
				t.Fatalf("Node mismatch: native %q Node %q", run.Stdout, timerNode(t, mode))
			}
			if report := leakcheck.Report(t, code, binary, mode); report != "" {
				t.Fatal(report)
			}
			t.Log(strings.TrimSpace(string(run.Stderr)))
		})
	}
}

type timerRuntime struct{ directory string }

func (runtime timerRuntime) build(code, output string, options native.Options) error {
	library, err := native.RuntimeLibrary(runtime.directory, options)
	if err != nil {
		return err
	}
	source := output + ".c"
	if err := os.WriteFile(source, []byte(code), 0600); err != nil {
		return err
	}
	args := append(native.LinkFlags(options), "-I", filepath.Dir(library), "-o", output, source)
	args = append(args, native.RuntimeLinkFlags(library)...)
	args = append(args, "-lm")
	result, err := exec.Command("clang", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("mutant must compile: %w\n%s", err, result)
	}
	return nil
}
func timerMutant(t *testing.T, file, before, after string) timerRuntime {
	t.Helper()
	runtime := timerRuntime{t.TempDir()}
	entries, err := os.ReadDir("runtime")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, entry := range entries {
		if entry.IsDir() || (filepath.Ext(entry.Name()) != ".h" && filepath.Ext(entry.Name()) != ".c") {
			continue
		}
		data, err := os.ReadFile(filepath.Join("runtime", entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if entry.Name() == file {
			if strings.Count(string(data), before) != 1 {
				t.Fatalf("mutation must name one site: %s %q", file, before)
			}
			data = []byte(strings.Replace(string(data), before, after, 1))
			found = true
		}
		if err := os.WriteFile(filepath.Join(runtime.directory, entry.Name()), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	if !found {
		t.Fatalf("missing mutation file %s", file)
	}
	return runtime
}
func TestTimerMutants(t *testing.T) {
	for _, mutant := range []struct{ name, mode, file, before, after, caught string }{
		{"released-before-last-fire", "interval", "timers_impl.h", "chosen->firing = false;", "if (chosen->active) adamic_release(chosen->callback);\n    chosen->firing = false;", "heap-use-after-free"},
		{"foreign-cancellation", "foreign", "timers_impl.h", "if (handle.provider != &timer_provider) return NULL;", "(void)handle.provider;", "Assertion"},
		{"unref-ignored", "unref", "timers_impl.h", "entry->referenced = false;", "entry->referenced = true;", "Node"},
		{"cancel-leaks-registration", "cancel", "timers_impl.h", "if (!entry->firing) timer_drop(entry);", "if (!entry->firing) { /* deliberately leak */ }", "leak"},
		{"cancel-frees-running-callback", "cancel-self", "timers_impl.h", "if (!entry->firing) timer_drop(entry);", "timer_drop(entry);", "heap-use-after-free"},
		{"promise-checkpoint-omitted", "order", "async.c", "        adamic_host_process_completions();", "        if (timers == NULL) adamic_host_process_completions();", "Node"},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			runtime := timerMutant(t, mutant.file, mutant.before, mutant.after)
			binary := filepath.Join(t.TempDir(), "mutant")
			code := timerHarness(t)
			if err := runtime.build(code, binary, native.Options{Sanitize: true, Count: true}); err != nil {
				t.Fatal(err)
			}
			if mutant.caught == "leak" {
				report, err := leakcheck.Check(leakcheck.Program{C: code, Sanitized: binary, Counted: filepath.Join(t.TempDir(), "counted"),
					BuildCounted: func(code, output string) error { return runtime.build(code, output, native.Options{Count: true}) },
					Arguments:    func() []string { return []string{mutant.mode} }, Execute: timerRun})
				if err != nil || (!strings.Contains(report, "LeakSanitizer: detected memory leaks") && !strings.Contains(report, "heap values leaked:")) {
					t.Fatalf("leak mutant must fail specifically the leak check: %v %q", err, report)
				}
				t.Log("caught by shared leakcheck")
				return
			}
			run := timerRun([]string{"ASAN_OPTIONS=detect_leaks=0:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary, mutant.mode)
			if mutant.caught == "Node" {
				if run.ExitCode != 0 {
					t.Fatalf("behavior mutant must finish: %+v", run)
				}
				if bytes.Equal(run.Stdout, timerNode(t, mutant.mode)) {
					t.Fatal("Node did not catch mutant")
				}
			} else if run.ExitCode == 0 || !strings.Contains(string(run.Stderr), mutant.caught) {
				t.Fatalf("want %s: %+v", mutant.caught, run)
			}
			t.Logf("caught by %s", mutant.caught)
		})
	}
}
