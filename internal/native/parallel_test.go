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
	"sync"
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
	name := "harness"
	if options.ThreadSanitize {
		name += "-tsan"
	}
	binary := filepath.Join(t.TempDir(), name)
	if err := Build(string(source), binary, options); err != nil {
		t.Fatal(err)
	}
	return binary
}

func parallelRun(t *testing.T, binary string, threads string, arguments ...string) (string, string) {
	t.Helper()
	runs := 1
	if strings.HasSuffix(binary, "-tsan") {
		runs = 3
	}
	var firstOut, firstErr string
	started := time.Now()
	for attempt := 0; attempt < runs; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		command := exec.CommandContext(ctx, binary, arguments...)
		command.Env = parallelEnvironment(threads, true)
		var stdout, stderr bytes.Buffer
		command.Stdout, command.Stderr = &stdout, &stderr
		err := command.Run()
		cancel()
		if err != nil {
			t.Fatalf("threads=%s attempt=%d: %v\n%s\n%s", threads, attempt+1, err, stdout.String(), stderr.String())
		}
		if strings.Contains(stderr.String(), "WARNING: ThreadSanitizer") || strings.Contains(stderr.String(), "ERROR:") {
			t.Fatal(stderr.String())
		}
		if attempt == 0 {
			firstOut, firstErr = stdout.String(), stderr.String()
		} else if firstOut != stdout.String() || firstErr != stderr.String() {
			t.Fatal("repeated race fixture changed its output")
		}
	}
	t.Logf("harness threads=%s arguments=%v runs=%d elapsed=%s", threads, arguments, runs, time.Since(started))
	return firstOut, firstErr
}

func parallelBuilds(t *testing.T) []struct {
	name    string
	options Options
} {
	builds := []struct {
		name    string
		options Options
	}{{"asan", Options{Sanitize: true}}, {"asan_slabs", Options{Sanitize: true, Slabs: true}}, {"count", Options{Count: true}}, {"malloc", Options{Malloc: true}}}
	if parallelTSan(t) {
		builds = append(builds, struct {
			name    string
			options Options
		}{"tsan", Options{ThreadSanitize: true}})
	}
	return builds
}

func TestParallelMemory(t *testing.T) {
	t.Parallel()
	for _, build := range parallelBuilds(t) {
		t.Run(build.name, func(t *testing.T) {
			binary := parallelHarness(t, "memory.c", build.options)
			stdout, _ := parallelRun(t, binary, "4")
			parallelLeaks(t, binary, build.options, "4")
			if stdout != "memory clean\n" {
				t.Fatalf("got %q", stdout)
			}
		})
	}
}

func TestParallelMap(t *testing.T) {
	t.Parallel()
	for _, build := range parallelBuilds(t) {
		t.Run(build.name, func(t *testing.T) {
			binary := parallelHarness(t, "map.c", build.options)
			for _, mode := range []string{"numbers", "strings", "objects", "map", "fresh", "nested", "exception", "nested_exception", "million"} {
				t.Run(mode, func(t *testing.T) {
					one, _ := parallelRun(t, binary, "1", mode)
					many, _ := parallelRun(t, binary, "4", mode)
					parallelLeaks(t, binary, build.options, "1", mode)
					parallelLeaks(t, binary, build.options, "4", mode)
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
	for _, build := range parallelBuilds(t) {
		t.Run(build.name, func(t *testing.T) {
			binary := parallelHarness(t, "map.c", build.options)
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, binary, "panic")
			command.Env = parallelEnvironment("4", false)
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
		{"lazy_cache", "string_index.c", "if (__atomic_compare_exchange_n(&((adamic_string *)string)->index, &expected, candidate,\n\t\tfalse, __ATOMIC_RELEASE, __ATOMIC_ACQUIRE)) { return candidate; }", "((adamic_string *)string)->index = candidate; return candidate;", "cache_race.c", "", "4", "WARNING: ThreadSanitizer: data race", true},
		{"plain_shared_count", "heap.c", "__atomic_fetch_add(&heap->references, 1, __ATOMIC_RELAXED);", "heap->references++;", "memory.c", "", "4", "WARNING: ThreadSanitizer: data race", true},
		{"field_cache", "adamic.h", "static inline adamic_value *adamic_object_data_field(const adamic_object *object, const char *name, adamic_slot_cache *cache) {\n\tif (object->has_captured_stack && strcmp(name, \"stack\") == 0) { return &((adamic_object *)object)->captured_stack; }\n\tuint64_t packed = __atomic_load_n(&cache->packed, __ATOMIC_RELAXED)", "static inline adamic_value *adamic_object_data_field(const adamic_object *object, const char *name, adamic_slot_cache *cache) {\n\tif (object->has_captured_stack && strcmp(name, \"stack\") == 0) { return &((adamic_object *)object)->captured_stack; }\n\tuint64_t packed = cache->packed", "cache_pairs.c", "", "4", "WARNING: ThreadSanitizer: data race", true},
		{"remote_free", "heap.c", "remote_slot *node = malloc(sizeof *node);", "give_local(slot, each); return; remote_slot *node = malloc(sizeof *node);", "memory.c", "", "4", "WARNING: ThreadSanitizer: data race", true},
		{"result_order", "parallel.c", "scope->results->elements[index] = result;", "scope->results->elements[scope->items->length - 1 - index] = result;", "map.c", "numbers", "4", "index 0:", false},
		{"reused_graph", "share.c", "if (visited(&pending, heap)) { continue; }", "if (shared || visited(&pending, heap)) { continue; }", "lifecycle.c", "", "4", "", false},
		{"oversized_slot", "object.c", "uint64_t packed = index <= UINT16_MAX ? pointer | ((uint64_t)index << 48) : 0;", "uint64_t packed = pointer | ((uint64_t)index << 48);", "scaling.c", "", "4", "", false},
		{"fixed_grain", "parallel.c", "return grain == 0 ? 1 : grain > 256 ? 256 : grain;", "return 256;", "scaling.c", "", "4", "", false},
		{"eager_strings", "share.c", "if (!shared) {", "if (!shared && heap->kind == adamic_kind_string) { adamic_string_prepare_shared((adamic_string *)heap); } if (!shared) {", "scaling.c", "", "4", "", false},
		{"one_worker", "parallel.c", "if (thread_count == 1) { return; }", "if (thread_count == 1) { thread_count = 2; }", "map.c", "numbers", "1", "", false},
	}
	for _, mutant := range mutants {
		t.Run(mutant.name, func(t *testing.T) {
			if mutant.race && !parallelTSan(t) {
				t.Skip("TSan is unavailable on this platform; see capability probe")
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
				if mutant.harness == "cache_race.c" && file.Name() == "string_index.c" {
					contents = []byte(cacheBuilderGate(string(contents)))
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
			attempts := 1
			if mutant.race {
				attempts = 3
			}
			for attempt := 0; attempt < attempts; attempt++ {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)

				args := []string{}
				if mutant.mode != "" {
					args = append(args, mutant.mode)
				}
				command := exec.CommandContext(ctx, binary, args...)
				command.Env = parallelEnvironment(mutant.threads, false)
				started := time.Now()
				output, err := command.CombinedOutput()
				timedOut := ctx.Err() != nil
				cancel()
				t.Logf("mutant attempt=%d elapsed=%s", attempt+1, time.Since(started))
				if err == nil || timedOut || !strings.Contains(string(output), mutant.want) {
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
			}
		})
	}
}

func TestParallelLifecycle(t *testing.T) {
	t.Parallel()
	for _, build := range parallelBuilds(t) {
		t.Run(build.name, func(t *testing.T) {
			binary := parallelHarness(t, "lifecycle.c", build.options)
			for _, threads := range []string{"1", "4"} {
				stdout, _ := parallelRun(t, binary, threads)
				parallelLeaks(t, binary, build.options, threads)
				if stdout != "lifecycle clean\n" {
					t.Fatalf("got %q", stdout)
				}
			}
		})
	}
}

// LeakSanitizer is Linux-only. Darwin's leaks tool checks an unsanitized malloc build.
func parallelEnvironment(threads string, leaks bool) []string {
	env := []string{}
	for _, entry := range os.Environ() {
		if strings.HasPrefix(entry, "ADAMIC_THREADS=") || strings.HasPrefix(entry, "ASAN_OPTIONS=") || strings.HasPrefix(entry, "TSAN_OPTIONS=") || strings.HasPrefix(entry, "ADAMIC_TSAN_PERTURB=") {
			continue
		}
		env = append(env, entry)
	}
	env = append(env, "ADAMIC_THREADS="+threads, "TSAN_OPTIONS=halt_on_error=1:history_size=4:report_atomic_races=1", "ADAMIC_TSAN_PERTURB=1")
	if goruntime.GOOS == "linux" {
		value := "0"
		if leaks {
			value = "1"
		}
		env = append(env, "ASAN_OPTIONS=detect_leaks="+value)
	}
	return env
}

func parallelLeaks(t *testing.T, binary string, options Options, threads string, arguments ...string) {
	t.Helper()
	if goruntime.GOOS != "darwin" || !options.Malloc {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	args := append([]string{"--atExit", "--", binary}, arguments...)
	command := exec.CommandContext(ctx, "leaks", args...)
	command.Env = parallelEnvironment(threads, false)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("Darwin leaks, threads=%s: %v\n%s", threads, err, output)
	}
}

var parallelTSanProbe struct {
	sync.Once
	available bool
	reason    string
}

func parallelTSan(t *testing.T) bool {
	t.Helper()
	if goruntime.GOOS == "linux" {
		return true
	}
	if goruntime.GOOS != "darwin" {
		return false
	}
	parallelTSanProbe.Do(func() {
		directory, err := os.MkdirTemp("", "adamic-tsan-probe-")
		if err != nil {
			parallelTSanProbe.reason = err.Error()
			return
		}
		defer os.RemoveAll(directory)
		source := filepath.Join(directory, "probe.c")
		binary := filepath.Join(directory, "probe")
		if err := os.WriteFile(source, []byte("int main(void) { return 0; }\n"), 0600); err != nil {
			parallelTSanProbe.reason = err.Error()
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if output, err := exec.CommandContext(ctx, "clang", "-fsanitize=thread", source, "-o", binary).CombinedOutput(); err != nil {
			parallelTSanProbe.reason = fmt.Sprintf("clang -fsanitize=thread: %v: %s", err, output)
			return
		}
		if output, err := exec.CommandContext(ctx, binary).CombinedOutput(); err != nil {
			parallelTSanProbe.reason = fmt.Sprintf("TSan probe runtime: %v: %s", err, output)
			return
		}
		parallelTSanProbe.available = true
	})
	if !parallelTSanProbe.available {
		t.Logf("Darwin TSan unavailable: %s", parallelTSanProbe.reason)
	}
	return parallelTSanProbe.available
}
