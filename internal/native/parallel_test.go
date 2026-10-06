package native

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"testing"
	"time"
)

func parallelHarness(t *testing.T, file string, options Options) string {
	t.Helper()
	source, err := os.ReadFile(filepath.Join("testdata", "parallel", file))
	if err != nil {
		t.Fatal(err)
	}
	source = []byte(parallelSource(string(source)))
	binary := filepath.Join(t.TempDir(), "harness")
	if err := Build(string(source), binary, options); err != nil {
		t.Fatal(err)
	}
	return binary
}

func parallelRun(t *testing.T, binary string, threads string, arguments ...string) (string, string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, binary, arguments...)
	command.Env = append(os.Environ(), "ADAMIC_THREADS="+threads, "ASAN_OPTIONS=detect_leaks=1", "TSAN_OPTIONS=halt_on_error=1")
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Run(); err != nil {
		t.Fatalf("threads=%s: %v\n%s\n%s", threads, err, stdout.String(), stderr.String())
	}
	if strings.Contains(stderr.String(), "WARNING: ThreadSanitizer") || strings.Contains(stderr.String(), "ERROR:") {
		t.Fatal(stderr.String())
	}
	return stdout.String(), stderr.String()
}

func parallelBuilds() []struct {
	name    string
	options Options
} {
	builds := []struct {
		name    string
		options Options
	}{{"asan", Options{Sanitize: true}}, {"count", Options{Count: true}}}
	if goruntime.GOOS == "linux" {
		builds = append(builds, struct {
			name    string
			options Options
		}{"tsan", Options{ThreadSanitize: true}})
	}
	return builds
}

func TestParallelMemory(t *testing.T) {
	t.Parallel()
	for _, build := range parallelBuilds() {
		t.Run(build.name, func(t *testing.T) {
			binary := parallelHarness(t, "memory.c", build.options)
			stdout, _ := parallelRun(t, binary, "4")
			if stdout != "memory clean\n" {
				t.Fatalf("got %q", stdout)
			}
		})
	}
}

func TestParallelMap(t *testing.T) {
	t.Parallel()
	for _, build := range parallelBuilds() {
		t.Run(build.name, func(t *testing.T) {
			binary := parallelHarness(t, "map.c", build.options)
			for _, mode := range []string{"numbers", "strings", "objects", "map", "fresh", "nested", "exception", "nested_exception", "million"} {
				t.Run(mode, func(t *testing.T) {
					one, _ := parallelRun(t, binary, "1", mode)
					many, _ := parallelRun(t, binary, "4", mode)
					if one != many {
						t.Fatalf("one worker %q, four workers %q", one, many)
					}
				})
			}
		})
	}
}

func TestParallelWorkerPanic(t *testing.T) {
	t.Parallel()
	for _, build := range parallelBuilds() {
		t.Run(build.name, func(t *testing.T) {
			binary := parallelHarness(t, "map.c", build.options)
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, binary, "panic")
			command.Env = append(os.Environ(), "ADAMIC_THREADS=4", "ASAN_OPTIONS=detect_leaks=0", "TSAN_OPTIONS=halt_on_error=1")
			output, err := command.CombinedOutput()
			failure, ok := err.(*exec.ExitError)
			if !ok || failure.ExitCode() != 70 || strings.Count(string(output), "adamic: panic: worker panic\n") != 1 || strings.Contains(string(output), "Sanitizer") || !strings.HasPrefix(string(output), "before worker panic\n") {
				t.Fatalf("got %v\n%s", err, output)
			}
		})
	}
}

// Use the emitter's actual declaration in the memory harness, so its one-line hook is held by TSan.
func parallelSource(source string) string {
	emitter := &emitter{}
	name := emitter.cache()
	declaration := strings.ReplaceAll(emitter.declarations[0], name, "cache")
	return strings.ReplaceAll(source, "static _Thread_local adamic_slot_cache cache;", declaration)
}

func TestParallelChecksCatchMutants(t *testing.T) {
	t.Parallel()
	mutants := []struct {
		name, file, old, changed, harness, mode, threads, want string
		race                                                   bool
	}{
		{"skip_items_share", "parallel.c", "adamic_share(items);", "/* mutant: publication without preparation */", "map.c", "strings", "4", "WARNING: ThreadSanitizer: data race", true},
		{"plain_shared_count", "heap.c", "__atomic_fetch_add(&heap->references, 1, __ATOMIC_RELAXED);", "heap->references++;", "memory.c", "", "4", "WARNING: ThreadSanitizer: data race", true},
		{"field_cache", "", "static _Thread_local adamic_slot_cache cache;", "static adamic_slot_cache cache;", "memory.c", "", "4", "WARNING: ThreadSanitizer: data race", true},
		{"remote_free", "heap.c", "remote_slot *node = malloc(sizeof *node);", "give_local(slot, each); return; remote_slot *node = malloc(sizeof *node);", "memory.c", "", "4", "WARNING: ThreadSanitizer: data race", true},
		{"result_order", "parallel.c", "scope->results->elements[index] = result;", "scope->results->elements[scope->items->length - 1 - index] = result;", "map.c", "numbers", "4", "index 0:", false},
		{"reused_graph", "share.c", "if (visited(&pending, heap)) { continue; }", "if (shared || visited(&pending, heap)) { continue; }", "lifecycle.c", "", "4", "", false},
		{"one_worker", "parallel.c", "if (thread_count == 1) { return; }", "if (thread_count == 1) { thread_count = 2; }", "map.c", "numbers", "1", "", false},
	}
	for _, mutant := range mutants {
		t.Run(mutant.name, func(t *testing.T) {
			if mutant.race && goruntime.GOOS != "linux" {
				t.Skip("TSan mutant proof is Linux only")
			}
			directory := t.TempDir()
			files, err := fs.ReadDir(runtime, "runtime")
			if err != nil {
				t.Fatal(err)
			}
			for _, file := range files {
				contents, err := fs.ReadFile(runtime, "runtime/"+file.Name())
				if err != nil {
					t.Fatal(err)
				}
				if file.Name() == mutant.file {
					if strings.Count(string(contents), mutant.old) != 1 {
						t.Fatalf("mutant seam %q is not unique", mutant.old)
					}
					contents = []byte(strings.Replace(string(contents), mutant.old, mutant.changed, 1))
				}
				if err := os.WriteFile(filepath.Join(directory, file.Name()), contents, 0600); err != nil {
					t.Fatal(err)
				}
			}
			source, err := os.ReadFile(filepath.Join("testdata", "parallel", mutant.harness))
			if err != nil {
				t.Fatal(err)
			}
			source = []byte(parallelSource(string(source)))
			if mutant.file == "" {
				source = []byte(strings.Replace(string(source), mutant.old, mutant.changed, 1))
			}
			options := Options{ThreadSanitize: mutant.race}
			library, err := RuntimeLibrary(directory, options)
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(directory, "main.c")
			if err := os.WriteFile(path, source, 0600); err != nil {
				t.Fatal(err)
			}
			binary := filepath.Join(directory, "mutant")
			arguments := append(Flags(options), "-I", filepath.Dir(library), path, "-o", binary)
			arguments = append(arguments, RuntimeLinkFlags(library)...)
			arguments = append(arguments, "-lm")
			if output, err := exec.Command("clang", arguments...).CombinedOutput(); err != nil {
				t.Fatalf("mutant must compile: %v\n%s", err, output)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()
			args := []string{}
			if mutant.mode != "" {
				args = append(args, mutant.mode)
			}
			command := exec.CommandContext(ctx, binary, args...)
			command.Env = append(os.Environ(), "ADAMIC_THREADS="+mutant.threads, "TSAN_OPTIONS=halt_on_error=1")
			output, err := command.CombinedOutput()
			if err == nil || ctx.Err() != nil || !strings.Contains(string(output), mutant.want) {
				t.Fatalf("mutant not caught by %q: %v\n%s", mutant.want, err, output)
			}
			if mutant.name == "one_worker" {
				failure, ok := err.(*exec.ExitError)
				if !ok || failure.ExitCode() != -1 {
					t.Fatalf("want harness abort, got %v\n%s", err, output)
				}
			}
			report := "harness failure"
			if mutant.race {
				report = "TSan data race"
				for _, line := range strings.Split(string(output), "\n") {
					if strings.Contains(line, "SUMMARY: ThreadSanitizer:") {
						report = line
						break
					}
				}
			}
			t.Log(fmt.Sprintf("%s caught: %s (%v)", mutant.name, report, err))
		})
	}
}

func TestParallelLifecycle(t *testing.T) {
	t.Parallel()
	for _, build := range parallelBuilds() {
		t.Run(build.name, func(t *testing.T) {
			binary := parallelHarness(t, "lifecycle.c", build.options)
			for _, threads := range []string{"1", "4"} {
				stdout, _ := parallelRun(t, binary, threads)
				if stdout != "lifecycle clean\n" {
					t.Fatalf("got %q", stdout)
				}
			}
		})
	}
}
