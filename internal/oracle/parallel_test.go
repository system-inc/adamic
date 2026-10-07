package oracle

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
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

func init() {
	paths, err := filepath.Glob(filepath.Join(repository, "internal/oracle/testdata/concurrency/accepted/*.a"))
	if err != nil {
		panic(err)
	}
	paths = append(paths, filepath.Join(repository, "bench/parallel_files.a"))
	for _, path := range paths {
		relative, err := filepath.Rel(repository, path)
		if err != nil {
			panic(err)
		}
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{filepath.ToSlash(relative), true, false})
	}
}

func usesParallelMap(program *ir.Program) bool {
	// The native emitter declares this ABI only when it emits a parallelMap expression.
	return strings.Contains(native.C(program), "adamic_parallel_map(")
}

// Each variant runs sequentially and with the pool's default size. Do not cache race proofs.
func checkParallelVariants(t *testing.T, program *ir.Program, oracle run) {
	t.Helper()
	builds := []struct {
		name    string
		options native.Options
	}{
		{"asan", native.Options{Sanitize: true}},
		{"release", native.Options{}},
		{"asan_slabs", native.Options{Sanitize: true, Slabs: true}},
	}
	if runtime.GOOS == "linux" {
		builds = append(builds, struct {
			name    string
			options native.Options
		}{"tsan", native.Options{ThreadSanitize: true}})
	}
	if runtime.GOOS == "darwin" {
		builds = append(builds, struct {
			name    string
			options native.Options
		}{"malloc", native.Options{Malloc: true}})
	}
	for _, build := range builds {
		t.Run(build.name, func(t *testing.T) {
			binary := filepath.Join(t.TempDir(), "program")
			if err := native.Build(native.C(program), binary, build.options); err != nil {
				t.Fatal(err)
			}
			for _, threads := range []string{"1", ""} {
				name := threads
				if name == "" {
					name = "default"
				}
				t.Run(name, func(t *testing.T) {
					attempts := 1
					if build.options.ThreadSanitize {
						attempts = 3
					}
					started := time.Now()
					for attempt := 0; attempt < attempts; attempt++ {
						deadline := time.Minute
						if build.options.ThreadSanitize {
							deadline = 5 * time.Minute
						}
						observed := executeParallelDeadline(t, threads, false, deadline, binary)
						if difference := disagreement(oracle, observed); difference != "" {
							t.Fatalf("%s: Node exit %d stdout %q stderr %q; native exit %d stdout %q stderr %q", difference, oracle.exitCode, oracle.stdout, oracle.stderr, observed.exitCode, observed.stdout, observed.stderr)
						}
					}
					t.Logf("race variant runs=%d elapsed=%s", attempts, time.Since(started))
					if oracle.exitCode != 0 {
						return
					}
					if runtime.GOOS == "linux" && build.options.Sanitize {
						report := executeParallel(t, threads, true, binary)
						if difference := disagreement(oracle, report); difference != "" {
							t.Fatalf("leak check: %s: exit %d stderr %s", difference, report.exitCode, report.stderr)
						}
					}
					if runtime.GOOS == "darwin" && build.options.Malloc {
						report := executeParallel(t, threads, false, "leaks", "--atExit", "--", binary)
						if report.exitCode != 0 {
							t.Fatalf("leaks: exit %d\n%s\n%s", report.exitCode, report.stdout, report.stderr)
						}
					}
				})
			}
		})
	}
}

// An unset override really is unset, even when the test's shell set ADAMIC_THREADS.
func executeParallel(t *testing.T, threads string, leaks bool, name string, arguments ...string) run {
	t.Helper()
	return executeParallelDeadline(t, threads, leaks, time.Minute, name, arguments...)
}

// TSan instruments the full benchmark and can exceed a minute under gate load.
// Every execution remains bounded and must still match Node without a race report.
func executeParallelDeadline(t *testing.T, threads string, leaks bool, deadline time.Duration, name string, arguments ...string) run {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), deadline)
	t.Cleanup(cancel)
	command := exec.CommandContext(ctx, name, arguments...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error { return syscall.Kill(-command.Process.Pid, syscall.SIGKILL) }
	command.WaitDelay = 5 * time.Second
	command.Env = []string{}
	for _, entry := range os.Environ() {
		if strings.HasPrefix(entry, "ADAMIC_THREADS=") || strings.HasPrefix(entry, "ASAN_OPTIONS=") || strings.HasPrefix(entry, "TSAN_OPTIONS=") || strings.HasPrefix(entry, "ADAMIC_TSAN_PERTURB=") {
			continue
		}
		command.Env = append(command.Env, entry)
	}
	if threads != "" {
		command.Env = append(command.Env, "ADAMIC_THREADS="+threads)
	}
	command.Env = append(command.Env, "TSAN_OPTIONS=halt_on_error=1:history_size=4:report_atomic_races=1", "ADAMIC_TSAN_PERTURB=1")
	if runtime.GOOS == "linux" {
		flag := 0
		if leaks {
			flag = 1
		}
		command.Env = append(command.Env, fmt.Sprintf("ASAN_OPTIONS=detect_leaks=%d", flag))
	}
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	var failure *exec.ExitError
	if err != nil && !errors.As(err, &failure) {
		t.Fatalf("%s: %v", name, err)
	}
	result := run{stdout: stdout.Bytes(), stderr: stderr.Bytes(), exitCode: command.ProcessState.ExitCode()}
	rememberRun(t, result)
	return result
}

// Reordering still compiles, finishes and leaks nothing. Only Node's ordered output catches it.
func TestParallelOracleCatchesResultOrder(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/concurrency/accepted/numbers.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	if !strings.Contains(source, "adamic_parallel_map(") {
		t.Fatal("mutant changed no call")
	}
	source = strings.ReplaceAll(source, "adamic_parallel_map(", "oracle_reordered_map(")
	source = `#include "adamic.h"
adamic_array *oracle_reordered_map(adamic_array *items, adamic_closure *work, bool references) {
 adamic_array *results = adamic_parallel_map(items, work, references);
 if (results != NULL) { adamic_array_reverse(results); }
 return results;
}
` + source
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	for _, threads := range []string{"1", ""} {
		observed := executeParallel(t, threads, runtime.GOOS == "linux", binary)
		if observed.exitCode != 0 || len(observed.stderr) != 0 {
			t.Fatalf("mutant must finish sanitizer and leak clean: %+v", observed)
		}
		if difference := disagreement(onNode(t, path), observed); difference != "stdout differs" {
			t.Fatalf("mutant caught by %q, want only stdout differs", difference)
		}
	}
	if runtime.GOOS == "darwin" {
		mallocBinary := filepath.Join(t.TempDir(), "malloc-mutant")
		if err := native.Build(source, mallocBinary, native.Options{Malloc: true}); err != nil {
			t.Fatal(err)
		}
		report := executeParallel(t, "1", false, "leaks", "--atExit", "--", mallocBinary)
		if report.exitCode != 0 {
			t.Fatalf("mutant leaks: %s %s", report.stdout, report.stderr)
		}
	}
	t.Log("result-order mutant compiled, finished sanitizer and leak clean; Node caught stdout differs at one thread and default")
}
