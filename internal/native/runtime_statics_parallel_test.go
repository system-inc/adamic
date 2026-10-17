package native

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"syscall"
	"testing"
	"time"
)

// Each fixture has its own four-worker process and uses the same immutable runtime snapshot.
func runRuntimeStaticsParallel(t *testing.T, unit int) {
	t.Helper()
	if goruntime.GOOS != "linux" {
		t.Skip("runtime static race proofs require Linux ThreadSanitizer")
	}
	if _, err := os.Stat(filepath.Join(*runtimeStaticsDirectory, "parallel.c")); os.IsNotExist(err) {
		t.Skip("concurrency-area has not been merged; parallel.c is absent")
	}
	library, root := staticsRaceLibrary(t, "", "", "")
	for index, fixture := range []string{"regex", "string_bmp", "string_supplementary", "normalization", "numbers", "map", "heap_weak", "field_cache", "json", "profiling", "output", "shape_types"} {
		if index != unit {
			continue
		}
		t.Run(fixture, func(t *testing.T) {
			binary := staticsRaceFixture(t, root, library, fixture)
			staticsRaceRun(t, binary, false)
		})
	}
}

// Each protection mutant compiles only its changed runtime unit.
func runRuntimeStaticsProtectionMutant(t *testing.T, unit int) {
	t.Helper()
	if goruntime.GOOS != "linux" {
		t.Skip("runtime static race proofs require Linux ThreadSanitizer")
	}
	if _, err := os.Stat(filepath.Join(*runtimeStaticsDirectory, "parallel.c")); os.IsNotExist(err) {
		t.Skip("concurrency-area has not been merged; parallel.c is absent")
	}
	mutants := []struct{ name, file, old, changed, fixture string }{
		{"regex_budget", "regexp.c", "static _Atomic uint64_t regex_step_limit;", "static uint64_t regex_step_limit;", "regex"},
		{"checker_profiling", "tsgo.c", "static _Atomic", "static", "profiling"},
		{"packed_field_cache", "object.c", "__atomic_store_n(&cache->packed, packed, __ATOMIC_RELAXED);", "cache->packed = packed;", "field_cache"},
		{"weak_table_lock", "weak.c", "pthread_mutex_lock(&table_lock);", "(void)&table_lock;", "heap_weak"},
		{"allocator_lists", "heap.c", "static _Thread_local chunk *giving", "static chunk *giving", "heap_weak"},
		{"count_allocation", "count.c", "atomic_fetch_add_explicit(&adamic_counted.allocations, 1, memory_order_relaxed);", "(*(size_t *)&adamic_counted.allocations)++;", "heap_weak"},
		{"output_buffer", "adamic.c", "pthread_mutex_lock(&output_lock);", "(void)&output_lock;", "output"},
		{"shape_types_registry", "union.c", "pthread_mutex_lock(&shape_types_lock);", "(void)&shape_types_lock;", "shape_types"},
		{"normalization_classes", "normalize.c", "static _Thread_local uint8_t cached_classes", "static uint8_t cached_classes", "normalization"},
		{"normalization_mappings", "normalize.c", "static _Thread_local const normalize_mapping *cached_mappings", "static const normalize_mapping *cached_mappings", "normalization"},
		{"normalization_pairs", "normalize.c", "static _Thread_local normalize_pair cached_pairs", "static normalize_pair cached_pairs", "normalization"},
	}
	for index, mutant := range mutants {
		if index != unit {
			continue
		}
		t.Run(mutant.name, func(t *testing.T) {
			if strings.HasPrefix(mutant.name, "normalization_") {
				for _, unsafe := range []bool{false, true} {
					name := "guarded"
					if unsafe {
						name = "mutant"
					}
					t.Run(name, func(t *testing.T) {
						library, root := staticsNormalizationRaceLibrary(t, mutant.name, mutant.old, mutant.changed, unsafe)
						binary := staticsRaceFixture(t, root, library, "normalization_race")
						staticsSteadyRun(t, binary, mutant.name, unsafe)
					})
				}
				return
			}
			library, root := staticsRaceLibrary(t, mutant.file, mutant.old, mutant.changed)
			binary := staticsRaceFixture(t, root, library, mutant.fixture)
			staticsRaceRun(t, binary, true)
		})
	}
}

// Not parallel: these processes intentionally terminate while four pool callbacks print.
func TestRuntimeStaticsSignalAndExit(t *testing.T) {
	if goruntime.GOOS != "linux" {
		t.Skip("requires Linux ThreadSanitizer")
	}
	started := time.Now()
	library, root := staticsRaceLibrary(t, "", "", "")
	t.Logf("signal runtime archive: %s", time.Since(started))
	started = time.Now()
	binary := staticsRaceFixture(t, root, library, "signal_output")
	t.Logf("signal fixture link: %s", time.Since(started))
	for _, worker := range []bool{false, true} {
		for _, signal := range []syscall.Signal{0, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGUSR2, syscall.SIGALRM, syscall.SIGXCPU, syscall.SIGVTALRM, syscall.SIGPROF, syscall.SIGIO, 30, 34, 64} {
			if worker && signal == 0 {
				continue
			} // Normal exit belongs to the loop thread.
			t.Run(fmt.Sprintf("worker_%t_signal_%d", worker, signal), func(t *testing.T) {
				staticsSignalRun(t, binary, worker, signal, false)
			})
		}
	}
	// Hold a real writer inside output_lock with bytes buffered before delivering the signal.
	// Relaxed fixture gates provide timing without ordering the unsafe handler access.
	for _, mutant := range []bool{false, true} {
		name := "handler_buffer_guarded"
		if mutant {
			name = "handler_buffer_mutant"
		}
		t.Run(name, func(t *testing.T) {
			library, root := staticsSignalRaceLibrary(t, mutant)
			binary := staticsRaceFixture(t, root, library, "signal_race")
			staticsSignalRun(t, binary, true, syscall.SIGTERM, mutant)
		})
	}
	// A relaxed readiness gate exposes the exit race without a long print workload.
	for _, mutant := range []bool{false, true} {
		name := "exit_flush_guarded"
		if mutant {
			name = "exit_flush_mutant"
		}
		t.Run(name, func(t *testing.T) {
			library, root := staticsExitRaceLibrary(t, mutant)
			binary := staticsRaceFixture(t, root, library, "exit_race")
			staticsSteadyRun(t, binary, "exit_flush", mutant)
		})
	}
}

// Not parallel: each build occupies the same runtime/TSan budget as the pool proofs.
func TestRuntimeStaticsStopWholeLines(t *testing.T) {
	if goruntime.GOOS != "linux" {
		t.Skip("requires Linux signals and ThreadSanitizer")
	}
	node := exec.Command("node", "-e", "console.log('first whole line'); console.log('second whole line'); process.kill(process.pid, 'SIGUSR2');")
	expected, err := node.Output()
	if err == nil || string(expected) != "first whole line\nsecond whole line\n" || node.ProcessState.Sys().(syscall.WaitStatus).Signal() != syscall.SIGUSR2 {
		t.Fatalf("unexpected Node observation: %v %q", err, expected)
	}
	for _, mutant := range []bool{false, true} {
		t.Run(fmt.Sprintf("drop_flush_%t", mutant), func(t *testing.T) {
			file, old, changed := "", "", ""
			if mutant {
				file, old, changed = "adamic.c", "\t\tflush();\n\t\t// Keep writers out", "\t\t/* mutant: drop whole-line stop flush */\n\t\t// Keep writers out"
			}
			library, root := staticsRaceLibrary(t, file, old, changed)
			binary := staticsRaceFixture(t, root, library, "stop_whole_lines")
			for attempt := 0; attempt < 3; attempt++ {
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				command := exec.CommandContext(ctx, binary)
				var errors strings.Builder
				command.Stderr = &errors
				command.Env = append(os.Environ(), "TSAN_OPTIONS=halt_on_error=1")
				actual, err := command.Output()
				timedOut := ctx.Err() != nil
				cancel()
				if err == nil || timedOut || command.ProcessState.Sys().(syscall.WaitStatus).Signal() != syscall.SIGUSR2 || errors.Len() != 0 {
					t.Fatalf("unexpected stop: %v %s", err, errors.String())
				}
				if mutant {
					if len(actual) != 0 {
						t.Fatalf("drop-flush mutant survived: %q", actual)
					}
					t.Logf("caught drop-flush mutant by stdout mismatch: Node %q, mutant %q", expected, actual)
				} else if string(actual) != string(expected) {
					t.Fatalf("whole-line output differs: Node %q, native %q", expected, actual)
				}
			}
		})
	}
}

// Instrument only the test snapshot. The protected and mutant handlers use the same gates;
// the mutant's sole unsafe change is one direct flush in the actual signal handler.
func staticsSignalRaceLibrary(t *testing.T, mutant bool) (string, string) {
	t.Helper()
	source, err := os.ReadFile(filepath.Join(*runtimeStaticsDirectory, "adamic.c"))
	if err != nil {
		t.Fatal(err)
	}
	changed := string(source)
	patch := func(old, replacement string) {
		t.Helper()
		if strings.Count(changed, old) != 1 {
			t.Fatalf("missing or ambiguous signal race seam: %q", old)
		}
		changed = strings.Replace(changed, old, replacement, 1)
	}
	patch("static void stopped(int signal_number) {", "extern void adamic_test_signal_release(void);\n\nstatic void stopped(int signal_number) {")
	// Declare the writer hook before buffer(), not just before stopped().
	patch("static void buffer(const char *bytes, size_t length) {", "extern void adamic_test_signal_writer(void);\n\nstatic void buffer(const char *bytes, size_t length) {")
	patch("output_used += length;", "output_used += length;\n\tadamic_test_signal_writer();")
	flush := ""
	if mutant {
		flush = "\n\tflush();"
	}
	patch("int saved_errno = errno;", "int saved_errno = errno;"+flush+"\n\tadamic_test_signal_release();")
	return staticsRaceLibrary(t, "adamic.c", string(source), changed)
}

func staticsSignalRun(t *testing.T, binary string, worker bool, signal syscall.Signal, mutant bool) {
	t.Helper()
	for attempt := 0; attempt < 3; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		command := exec.CommandContext(ctx, binary)
		command.Env = append(os.Environ(), "ADAMIC_THREADS=4", "TSAN_OPTIONS=halt_on_error=1", fmt.Sprintf("ADAMIC_STOP_SIGNAL=%d", signal))
		if worker {
			command.Env = append(command.Env, "ADAMIC_STOP_WORKER=1")
		}
		started := time.Now()
		output, err := command.CombinedOutput()
		t.Logf("signal process attempt %d: %s (%d bytes)", attempt+1, time.Since(started), len(output))
		timedOut := ctx.Err() != nil
		cancel()
		race := strings.Contains(string(output), "WARNING: ThreadSanitizer: data race")
		if mutant {
			if err == nil || timedOut || !race {
				t.Fatalf("signal mutant not caught (attempt %d): %v\n%s", attempt+1, err, output)
			}
			for _, line := range strings.Split(string(output), "\n") {
				if strings.Contains(line, "SUMMARY: ThreadSanitizer:") {
					t.Log(line)
				}
			}
			continue
		}
		if timedOut || race || !strings.Contains(string(output), "signal worker line\n") {
			t.Fatalf("signal fixture failed: %v\n%s", err, output)
		}
		if signal == 0 {
			if err != nil {
				t.Fatalf("exit fixture failed: %v\n%s", err, output)
			}
		} else {
			status, ok := command.ProcessState.Sys().(syscall.WaitStatus)
			if !ok || !status.Signaled() || status.Signal() != signal {
				t.Fatalf("wanted signal %d, got %v\n%s", signal, err, output)
			}
		}
	}
}

func staticsRaceLibrary(t *testing.T, mutantFile, old, changed string) (string, string) {
	t.Helper()
	root := t.TempDir()
	files, err := filepath.Glob(filepath.Join(*runtimeStaticsDirectory, "*"))
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if filepath.Ext(file) != ".c" && filepath.Ext(file) != ".h" {
			continue
		}
		source, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if filepath.Base(file) == mutantFile {
			if !strings.Contains(string(source), old) {
				t.Fatalf("missing mutant seam in %s: %q", mutantFile, old)
			}
			source = []byte(strings.ReplaceAll(string(source), old, changed))
			if mutantFile == "adamic.c" && old == "pthread_mutex_lock(&output_lock);" {
				source = []byte(strings.ReplaceAll(string(source), "pthread_mutex_unlock(&output_lock);", "/* mutant: no output unlock */"))
			}
			if mutantFile == "union.c" {
				source = []byte(strings.ReplaceAll(string(source), "pthread_mutex_unlock(&shape_types_lock);", "/* mutant: no registry unlock */"))
			}
			if mutantFile == "weak.c" {
				source = []byte(strings.ReplaceAll(string(source), "pthread_mutex_unlock(&table_lock);", "/* mutant: no table unlock */"))
			}
		}
		if err := os.WriteFile(filepath.Join(root, filepath.Base(file)), source, 0600); err != nil {
			t.Fatal(err)
		}
	}
	header, err := os.ReadFile("../../bridge/tsgo/tsgo.h")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "tsgo.h"), header, 0600); err != nil {
		t.Fatal(err)
	}
	filesSnapshot, err := readRuntime(os.DirFS(root), ".")
	if err != nil {
		t.Fatal(err)
	}
	// The baseline archive is shared; only the changed C unit is recompiled for a mutant.
	baseline, err := readRuntime(runtime, "runtime")
	if err != nil {
		t.Fatal(err)
	}
	baseline = append(baseline, runtimeFile{"tsgo.h", header})
	compiler, err := exec.LookPath("clang")
	if err != nil {
		t.Fatal(err)
	}
	version, err := exec.Command(compiler, "--version").CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	cache, err := cacheDirectory("runtime-statics")
	if err != nil {
		t.Fatal(err)
	}
	original, err := cachedRuntime(baseline, staticsRaceFlags(), compiler, string(version), cache)
	if err != nil {
		t.Fatal(err)
	}
	if mutantFile == "" {
		return original, filepath.Dir(original)
	}
	var objects []string
	for _, file := range filesSnapshot {
		if !strings.HasSuffix(file.name, ".c") {
			continue
		}
		object := filepath.Join(filepath.Dir(original), strings.TrimSuffix(file.name, ".c")+".o")
		if file.name == mutantFile {
			object = filepath.Join(root, strings.TrimSuffix(file.name, ".c")+".o")
			arguments := append(staticsRaceFlags(), "-I", root, "-c", filepath.Join(root, file.name), "-o", object)
			staticsCompile(t, "clang", arguments...)
		}
		objects = append(objects, object)
	}
	library := filepath.Join(root, "runtime.a")
	staticsCompile(t, "ar", append([]string{"rcs", library}, objects...)...)
	return library, root
}

func staticsRaceFlags() []string {
	return []string{"-std=c11", "-O1", "-g", "-pthread", "-fsanitize=thread", "-DADAMIC_COUNT", "-DADAMIC_TSGO", "-Wall", "-Wextra", "-Werror", "-pedantic"}
}

func staticsCompile(t *testing.T, tool string, arguments ...string) {
	t.Helper()
	started := time.Now()
	defer func() { t.Logf("%s compile/link: %s", tool, time.Since(started)) }()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, tool, arguments...)
	if strings.HasPrefix(filepath.Base(tool), "clang") {
		command = productCompilerCommand(t, ctx, tool, arguments...)
	}
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("fixture and mutant must compile: %v\n%s", err, output)
	}
}

func staticsRaceFixture(t *testing.T, root, library, fixture string) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), fixture)
	source := "testdata/runtime-statics/" + fixture + ".c"
	arguments := append(staticsRaceFlags(), "-I", root, source, "testdata/runtime-statics/tsgo_stub.c", "-Wl,--whole-archive", library, "-Wl,--no-whole-archive", "-lm", "-o", binary)
	staticsCompile(t, "clang", arguments...)
	return binary
}

func staticsRaceRun(t *testing.T, binary string, mutant bool) {
	t.Helper()
	// Repeat both positive and negative observations. A crash, timeout or compiler
	// error is never accepted as evidence of catching a race mutant.
	for attempt := 0; attempt < 3; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		command := exec.CommandContext(ctx, binary)
		var env []string
		for _, entry := range os.Environ() {
			if !strings.HasPrefix(entry, "ADAMIC_THREADS=") && !strings.HasPrefix(entry, "TSAN_OPTIONS=") && !strings.HasPrefix(entry, "ADAMIC_TSGO_TIMING=") && !strings.HasPrefix(entry, "ADAMIC_TSGO_PROFILE=") {
				env = append(env, entry)
			}
		}
		command.Env = append(env, "ADAMIC_THREADS=4", "TSAN_OPTIONS=halt_on_error=1", "ADAMIC_TSGO_TIMING=1", "ADAMIC_TSGO_PROFILE=1")
		output, err := command.CombinedOutput()
		timedOut := ctx.Err() != nil
		cancel()
		race := strings.Contains(string(output), "WARNING: ThreadSanitizer: data race")
		if mutant {
			if err == nil || timedOut || !race {
				t.Fatalf("race mutant was not caught (attempt %d): %v\n%s", attempt+1, err, output)
			}
			for _, line := range strings.Split(string(output), "\n") {
				if strings.Contains(line, "SUMMARY: ThreadSanitizer:") {
					t.Log(line)
				}
			}
		} else if err != nil || race || !strings.Contains(string(output), "runtime statics clean\n") {
			t.Fatalf("fixture failed (attempt %d): %v\n%s", attempt+1, err, output)
		}
	}
}

func TestRuntimeStaticsParallelUnit00(t *testing.T) {
	t.Parallel()
	runRuntimeStaticsParallel(t, 0)
}

func TestRuntimeStaticsParallelUnit01(t *testing.T) {
	t.Parallel()
	runRuntimeStaticsParallel(t, 1)
}

func TestRuntimeStaticsParallelUnit02(t *testing.T) {
	t.Parallel()
	runRuntimeStaticsParallel(t, 2)
}

func TestRuntimeStaticsParallelUnit03(t *testing.T) {
	t.Parallel()
	runRuntimeStaticsParallel(t, 3)
}

func TestRuntimeStaticsParallelUnit04(t *testing.T) {
	t.Parallel()
	runRuntimeStaticsParallel(t, 4)
}

func TestRuntimeStaticsParallelUnit05(t *testing.T) {
	t.Parallel()
	runRuntimeStaticsParallel(t, 5)
}

func TestRuntimeStaticsParallelUnit06(t *testing.T) {
	t.Parallel()
	runRuntimeStaticsParallel(t, 6)
}

func TestRuntimeStaticsParallelUnit07(t *testing.T) {
	t.Parallel()
	runRuntimeStaticsParallel(t, 7)
}

func TestRuntimeStaticsParallelUnit08(t *testing.T) {
	t.Parallel()
	runRuntimeStaticsParallel(t, 8)
}

func TestRuntimeStaticsParallelUnit09(t *testing.T) {
	t.Parallel()
	runRuntimeStaticsParallel(t, 9)
}

func TestRuntimeStaticsParallelUnit10(t *testing.T) {
	t.Parallel()
	runRuntimeStaticsParallel(t, 10)
}

func TestRuntimeStaticsParallelUnit11(t *testing.T) {
	t.Parallel()
	runRuntimeStaticsParallel(t, 11)
}

func TestRuntimeStaticsProtectionMutantsUnit00(t *testing.T) {
	t.Parallel()
	runRuntimeStaticsProtectionMutant(t, 0)
}

func TestRuntimeStaticsProtectionMutantsUnit01(t *testing.T) {
	t.Parallel()
	runRuntimeStaticsProtectionMutant(t, 1)
}

func TestRuntimeStaticsProtectionMutantsUnit02(t *testing.T) {
	t.Parallel()
	runRuntimeStaticsProtectionMutant(t, 2)
}

func TestRuntimeStaticsProtectionMutantsUnit03(t *testing.T) {
	t.Parallel()
	runRuntimeStaticsProtectionMutant(t, 3)
}

func TestRuntimeStaticsProtectionMutantsUnit04(t *testing.T) {
	t.Parallel()
	runRuntimeStaticsProtectionMutant(t, 4)
}

func TestRuntimeStaticsProtectionMutantsUnit05(t *testing.T) {
	t.Parallel()
	runRuntimeStaticsProtectionMutant(t, 5)
}

func TestRuntimeStaticsProtectionMutantsUnit06(t *testing.T) {
	t.Parallel()
	runRuntimeStaticsProtectionMutant(t, 6)
}

func TestRuntimeStaticsProtectionMutantsUnit07(t *testing.T) {
	t.Parallel()
	runRuntimeStaticsProtectionMutant(t, 7)
}

func TestRuntimeStaticsProtectionMutantsUnit08(t *testing.T) {
	t.Parallel()
	runRuntimeStaticsProtectionMutant(t, 8)
}

func TestRuntimeStaticsProtectionMutantsUnit09(t *testing.T) {
	t.Parallel()
	runRuntimeStaticsProtectionMutant(t, 9)
}

func TestRuntimeStaticsProtectionMutantsUnit10(t *testing.T) {
	t.Parallel()
	runRuntimeStaticsProtectionMutant(t, 10)
}
