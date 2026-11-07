package oracle

import (
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
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/native/runtimeproducts"
)

func init() {
	if runtime.GOOS == "linux" {
		runtimeproducts.Variants = append(
			runtimeproducts.Variants,
			runtimeproducts.Variant{Name: "parallel_tsan", Options: native.Options{ThreadSanitize: true}},
		)
	}
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
	return emitsParallelMap(native.C(program))
}

// emitsParallelMap is whether emitted C calls the pool: the native emitter declares this ABI only when it emits a
// parallelMap expression.
func emitsParallelMap(code string) bool {
	return strings.Contains(code, "adamic_parallel_map(") || strings.Contains(code, "adamic_parallel_map_move(")
}

// Each variant runs sequentially and with the pool's default size. Do not cache race proofs.
func checkParallelVariants(test *testing.T, program *ir.Program, oracle run) {
	checkParallelVariantUnit(test, program, oracle, "", "", -1)
}

func checkParallelVariantUnit(
	test *testing.T,
	program *ir.Program,
	oracle run,
	selectedBuild, selectedThreads string,
	selectedAttempt int,
) {
	checkParallelCodeUnit(test, native.C(program), oracle, selectedBuild, selectedThreads, selectedAttempt)
}

// Move units can consume their independently generated C product.
func checkParallelCodeUnit(
	test *testing.T,
	code string,
	oracle run,
	selectedBuild, selectedThreads string,
	selectedAttempt int,
) {
	test.Helper()
	builds := []struct {
		name    string
		options native.Options
	}{
		{"asan", native.Options{Sanitize: true}},
		{"release", native.Options{}},
		{"asan_slabs", native.Options{Sanitize: true}},
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
		}{"malloc", native.Options{}})
	}
	for _, build := range builds {
		if selectedBuild != "" && build.name != selectedBuild {
			continue
		}
		test.Run(build.name, func(test *testing.T) {
			binary := filepath.Join(test.TempDir(), "program")
			allocator := ""
			if build.name == "asan_slabs" {
				allocator = "ADAMIC_SLABS"
			} else if build.name == "malloc" {
				allocator = "ADAMIC_MALLOC"
			}
			buildParallelProgram(test, code, binary, build.options, allocator)
			for _, threads := range []string{"1", ""} {
				if selectedBuild != "" && selectedThreads != "all" && threads != selectedThreads {
					continue
				}
				name := threads
				if name == "" {
					name = "default"
				}
				test.Run(name, func(test *testing.T) {
					attempts := 1
					if build.options.ThreadSanitize {
						attempts = 3
					}
					started := time.Now()
					for attempt := 0; attempt < attempts; attempt++ {
						if selectedAttempt >= 0 && attempt != selectedAttempt {
							continue
						}
						deadline := time.Minute
						observed := executeParallelDeadline(test, threads, false, deadline, binary)
						if difference := disagreement(oracle, observed); difference != "" {
							test.Fatalf(
								"%s: Node exit %d stdout %q stderr %q; native exit %d stdout %q stderr %q",
								difference,
								oracle.exitCode,
								oracle.stdout,
								oracle.stderr,
								observed.exitCode,
								observed.stdout,
								observed.stderr,
							)
						}
					}
					executed := attempts
					if selectedAttempt >= 0 {
						executed = 1
					}
					test.Logf("race variant runs=%d elapsed=%s", executed, time.Since(started))
					if oracle.exitCode != 0 {
						return
					}
					if runtime.GOOS == "linux" && build.options.Sanitize {
						report := executeParallel(test, threads, true, binary)
						if difference := disagreement(oracle, report); difference != "" {
							test.Fatalf("leak check: %s: exit %d stderr %s", difference, report.exitCode, report.stderr)
						}
					}
					if runtime.GOOS == "darwin" && build.name == "malloc" {
						report := executeParallel(test, threads, false, "leaks", "--atExit", "--", binary)
						if report.exitCode != 0 {
							test.Fatalf("leaks: exit %d\n%s\n%s", report.exitCode, report.stdout, report.stderr)
						}
					}
				})
			}
		})
	}
}

// An unset override really is unset, even when the test's shell set ADAMIC_THREADS.
func executeParallel(test *testing.T, threads string, leaks bool, name string, arguments ...string) run {
	test.Helper()
	return executeParallelDeadline(test, threads, leaks, time.Minute, name, arguments...)
}

// Every independent execution must finish within one minute and match Node.
func executeParallelDeadline(
	test *testing.T,
	threads string,
	leaks bool,
	deadline time.Duration,
	name string,
	arguments ...string,
) run {
	test.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), deadline)
	test.Cleanup(cancel)
	command := exec.CommandContext(ctx, name, arguments...)
	command.Env = []string{}
	for _, entry := range os.Environ() {
		if strings.HasPrefix(entry, "ADAMIC_THREADS=") ||
			strings.HasPrefix(entry, "ASAN_OPTIONS=") ||
			strings.HasPrefix(entry, "TSAN_OPTIONS=") ||
			strings.HasPrefix(entry, "ADAMIC_TSAN_PERTURB=") {
			continue
		}
		command.Env = append(command.Env, entry)
	}
	if threads != "" {
		command.Env = append(command.Env, "ADAMIC_THREADS="+threads)
	}
	command.Env = append(
		command.Env,
		"TSAN_OPTIONS=halt_on_error=1:history_size=4:report_atomic_races=1",
		"ADAMIC_TSAN_PERTURB=1",
	)
	if runtime.GOOS == "linux" {
		flag := 0
		if leaks {
			flag = 1
		}
		command.Env = append(command.Env, fmt.Sprintf("ASAN_OPTIONS=detect_leaks=%d", flag))
	}
	stdout, stderr, err := childguard.Capture(ctx, command)
	var failure *exec.ExitError
	if err != nil && !errors.As(err, &failure) {
		test.Fatalf("%s: %v", name, err)
	}
	result := run{stdout: stdout, stderr: stderr, exitCode: command.ProcessState.ExitCode()}
	rememberRun(test, result)
	return result
}

// Reordering still compiles, finishes and leaks nothing. Only Node's ordered output catches it.
func TestParallelOracleCatchesResultOrder(test *testing.T) {
	test.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/concurrency/accepted/numbers.a"))
	if err != nil {
		test.Fatal(err)
	}
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	source := native.C(program)
	if !strings.Contains(source, "adamic_parallel_map(") {
		test.Fatal("mutant changed no call")
	}
	source = strings.ReplaceAll(source, "adamic_parallel_map(", "oracle_reordered_map(")
	source = `#include "adamic.h"
adamic_array *oracle_reordered_map(adamic_array *items, adamic_closure *work, enum adamic_ir_type element) {
 adamic_array *results = adamic_parallel_map(items, work, element);
 if (results != NULL) { adamic_array_reverse(results); }
 return results;
}
` + source
	binary := filepath.Join(test.TempDir(), "mutant")
	if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
		test.Fatal(err)
	}
	for _, threads := range []string{"1", ""} {
		observed := executeParallel(test, threads, runtime.GOOS == "linux", binary)
		if observed.exitCode != 0 || len(observed.stderr) != 0 {
			test.Fatalf("mutant must finish sanitizer and leak clean: %+v", observed)
		}
		if difference := disagreement(onNode(test, path), observed); difference != "stdout differs" {
			test.Fatalf("mutant caught by %q, want only stdout differs", difference)
		}
	}
	if runtime.GOOS == "darwin" {
		mallocBinary := filepath.Join(test.TempDir(), "malloc-mutant")
		buildParallelProgram(test, source, mallocBinary, native.Options{}, "ADAMIC_MALLOC")
		report := executeParallel(test, "1", false, "leaks", "--atExit", "--", mallocBinary)
		if report.exitCode != 0 {
			test.Fatalf("mutant leaks: %s %s", report.stdout, report.stderr)
		}
	}
	test.Log(
		"result-order mutant compiled, finished sanitizer and leak clean; Node caught stdout differs at one thread and default",
	)
}

func parallelFixtureUnit(test *testing.T, path, build, threads string, attempt int) {
	test.Helper()
	absolute, err := filepath.Abs(filepath.Join(repository, path))
	if err != nil {
		test.Fatal(err)
	}
	program, err := lowered(test, absolute)
	if err != nil {
		test.Fatal(err)
	}
	oracle := onNode(test, absolute)
	if diff := disagreement(oracle, onJavaScriptBackend(test, program)); diff != "" {
		test.Fatal(diff)
	}
	checkParallelVariantUnit(test, program, oracle, build, threads, attempt)
}

func TestParallelNative_fresh(test *testing.T) {
	test.Parallel()
	parallelFixtureUnit(test, "internal/oracle/testdata/concurrency/accepted/fresh.a", "", "", -1)
}

func TestParallelNative_global_capture(test *testing.T) {
	test.Parallel()
	parallelFixtureUnit(test, "internal/oracle/testdata/concurrency/accepted/global_capture.a", "", "", -1)
}

func TestParallelNative_identity(test *testing.T) {
	test.Parallel()
	parallelFixtureUnit(test, "internal/oracle/testdata/concurrency/accepted/identity.a", "", "", -1)
}

func TestParallelNative_large(test *testing.T) {
	test.Parallel()
	parallelFixtureUnit(test, "internal/oracle/testdata/concurrency/accepted/large.a", "", "", -1)
}

func TestParallelNative_map_capture(test *testing.T) {
	test.Parallel()
	parallelFixtureUnit(test, "internal/oracle/testdata/concurrency/accepted/map_capture.a", "", "", -1)
}

func TestParallelNative_nested(test *testing.T) {
	test.Parallel()
	parallelFixtureUnit(test, "internal/oracle/testdata/concurrency/accepted/nested.a", "", "", -1)
}

func TestParallelNative_numbers(test *testing.T) {
	test.Parallel()
	parallelFixtureUnit(test, "internal/oracle/testdata/concurrency/accepted/numbers.a", "", "", -1)
}

func TestParallelNative_readonly_forms(test *testing.T) {
	test.Parallel()
	parallelFixtureUnit(test, "internal/oracle/testdata/concurrency/accepted/readonly_forms.a", "", "", -1)
}

func TestParallelNative_readonly_function(test *testing.T) {
	test.Parallel()
	parallelFixtureUnit(test, "internal/oracle/testdata/concurrency/accepted/readonly_function.a", "", "", -1)
}

func TestParallelNative_records(test *testing.T) {
	test.Parallel()
	parallelFixtureUnit(test, "internal/oracle/testdata/concurrency/accepted/records.a", "", "", -1)
}

func TestParallelNative_recursive_tree(test *testing.T) {
	test.Parallel()
	parallelFixtureUnit(test, "internal/oracle/testdata/concurrency/accepted/recursive_tree.a", "", "", -1)
}

func TestParallelNative_strings(test *testing.T) {
	test.Parallel()
	parallelFixtureUnit(test, "internal/oracle/testdata/concurrency/accepted/strings.a", "", "", -1)
}

func TestParallelNative_throw_middle(test *testing.T) {
	test.Parallel()
	parallelFixtureUnit(test, "internal/oracle/testdata/concurrency/accepted/throw_middle.a", "", "", -1)
}

func TestParallelBenchmark_asan_One_0(test *testing.T) {
	test.Parallel()
	parallelFixtureUnit(test, "bench/parallel_files.a", "asan", "1", 0)
}

func TestParallelBenchmark_asan_Default_0(test *testing.T) {
	test.Parallel()
	parallelFixtureUnit(test, "bench/parallel_files.a", "asan", "", 0)
}

func TestParallelBenchmark_release_One_0(test *testing.T) {
	test.Parallel()
	parallelFixtureUnit(test, "bench/parallel_files.a", "release", "1", 0)
}

func TestParallelBenchmark_release_Default_0(test *testing.T) {
	test.Parallel()
	parallelFixtureUnit(test, "bench/parallel_files.a", "release", "", 0)
}

func TestParallelBenchmark_asan_slabs_One_0(test *testing.T) {
	test.Parallel()
	parallelFixtureUnit(test, "bench/parallel_files.a", "asan_slabs", "1", 0)
}

func TestParallelBenchmark_asan_slabs_Default_0(test *testing.T) {
	test.Parallel()
	parallelFixtureUnit(test, "bench/parallel_files.a", "asan_slabs", "", 0)
}

func TestParallelBenchmark_tsan_One_0(test *testing.T) {
	test.Parallel()
	parallelFixtureUnit(test, "bench/parallel_files.a", "tsan", "1", 0)
}

func TestParallelBenchmark_tsan_One_1(test *testing.T) {
	test.Parallel()
	parallelFixtureUnit(test, "bench/parallel_files.a", "tsan", "1", 1)
}

func TestParallelBenchmark_tsan_One_2(test *testing.T) {
	test.Parallel()
	parallelFixtureUnit(test, "bench/parallel_files.a", "tsan", "1", 2)
}

func TestParallelBenchmark_tsan_Default_0(test *testing.T) {
	test.Parallel()
	parallelFixtureUnit(test, "bench/parallel_files.a", "tsan", "", 0)
}

func TestParallelBenchmark_tsan_Default_1(test *testing.T) {
	test.Parallel()
	parallelFixtureUnit(test, "bench/parallel_files.a", "tsan", "", 1)
}

func TestParallelBenchmark_tsan_Default_2(test *testing.T) {
	test.Parallel()
	parallelFixtureUnit(test, "bench/parallel_files.a", "tsan", "", 2)
}

// The base keeps slabs private to native tests. A snapshot header selects that
// allocator without widening the compiler's public options for fixture tooling.
func buildParallelProgram(test *testing.T, code, binary string, options native.Options, allocator string) {
	test.Helper()
	if allocator == "" {
		if err := native.Build(code, binary, options); err != nil {
			test.Fatal(err)
		}
		return
	}
	directory := test.TempDir()
	// Reads: internal/native/runtime/regexp_replace.c
	root := filepath.Join(repository, "internal/native/runtime")
	files, err := os.ReadDir(root)
	if err != nil {
		test.Fatal(err)
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, file.Name()))
		if err != nil {
			test.Fatal(err)
		}
		if file.Name() == "adamic.h" || filepath.Ext(file.Name()) == ".c" {
			data = append([]byte("#define "+allocator+" 1\n"), data...)
		}
		if err := os.WriteFile(filepath.Join(directory, file.Name()), data, 0600); err != nil {
			test.Fatal(err)
		}
	}
	library, err := native.RuntimeLibraryForSource(directory, code, options)
	if err != nil {
		test.Fatal(err)
	}
	if err := native.BuildWithRuntime(code, binary, options, library); err != nil {
		test.Fatalf("slab fixture build: %v", err)
	}
}

func TestProduct_ParallelTSanRuntime0(test *testing.T) {
	test.Parallel()
	if runtime.GOOS != "linux" {
		test.Skip("Linux TSan fixture runtime")
	}
	runtimeproducts.Build(test, "parallel_tsan", 0)
}
func TestProduct_ParallelTSanRuntime1(test *testing.T) {
	test.Parallel()
	if runtime.GOOS != "linux" {
		test.Skip("Linux TSan fixture runtime")
	}
	runtimeproducts.Build(test, "parallel_tsan", 1)
}
func TestProduct_ParallelTSanRuntime2(test *testing.T) {
	test.Parallel()
	if runtime.GOOS != "linux" {
		test.Skip("Linux TSan fixture runtime")
	}
	runtimeproducts.Build(test, "parallel_tsan", 2)
}
