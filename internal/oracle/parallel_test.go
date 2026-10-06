package oracle

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

func init() {
	paths, err := filepath.Glob(filepath.Join(repository, "internal/oracle/testdata/concurrency/accepted/*.a"))
	if err != nil {
		panic(err)
	}
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
					observed := executeParallel(t, threads, false, binary)
					if difference := disagreement(oracle, observed); difference != "" {
						t.Fatalf("%s: Node exit %d stdout %q stderr %q; native exit %d stdout %q stderr %q", difference, oracle.exitCode, oracle.stdout, oracle.stderr, observed.exitCode, observed.stdout, observed.stderr)
					}
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
	command := bounded(t, name, arguments...)
	command.Env = []string{}
	for _, entry := range os.Environ() {
		if strings.HasPrefix(entry, "ADAMIC_THREADS=") || strings.HasPrefix(entry, "ASAN_OPTIONS=") || strings.HasPrefix(entry, "TSAN_OPTIONS=") {
			continue
		}
		command.Env = append(command.Env, entry)
	}
	if threads != "" {
		command.Env = append(command.Env, "ADAMIC_THREADS="+threads)
	}
	command.Env = append(command.Env, "TSAN_OPTIONS=halt_on_error=1")
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
