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
	if goruntime.GOOS == "darwin" && !options.Sanitize && !options.ThreadSanitize && !options.Count {
		return buildWithFlags(t, string(source), append(Flags(options), "-DADAMIC_MALLOC"))
	}
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
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
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
		for _, line := range strings.Split(stderr.String(), "\n") {
			if !strings.HasPrefix(line, "adamic: counts:") {
				continue
			}
			var allocations, frees, retains, releases, peak, regions uint64
			parsed, err := fmt.Sscanf(line, "adamic: counts: allocations %d frees %d retains %d releases %d peak %d regions %d", &allocations, &frees, &retains, &releases, &peak, &regions)
			if err != nil || parsed != 6 || allocations != frees+regions {
				t.Fatalf("unbalanced counted parallel harness: %s", line)
			}
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
	}{{"asan", Options{Sanitize: true}}, {"asan_slabs", Options{Sanitize: true, slabs: true}}, {"count", Options{Count: true}}, {"release", Options{}}}
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

// Each build keeps both worker counts together for the output equality assertion.
func parallelMapMode(t *testing.T, mode string, options Options) {
	t.Helper()
	binary := parallelHarness(t, "map.c", options)
	one, _ := parallelRun(t, binary, "1", mode)
	many, _ := parallelRun(t, binary, "4", mode)
	parallelLeaks(t, binary, options, "1", mode)
	parallelLeaks(t, binary, options, "4", mode)
	if one != many {
		t.Fatalf("one worker %q, four workers %q", one, many)
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

func parallelCheckMutant(t *testing.T, name string) {
	mutants := []struct {
		name, file, old, changed, harness, mode, threads, want string
		race                                                   bool
	}{
		{"view_parent_units", "string_slice_impl.h", "if (slice != string) { slice->units = last - first + 1; }", "slice->units = last - first + 1;", "map.c", "strings", "4", "WARNING: ThreadSanitizer: data race", true},
		{"view_owner_count", "string_share.c", "size_t references = adamic_reference_count(&owner->heap);", "size_t references = owner->heap.references;", "map.c", "strings", "4", "WARNING: ThreadSanitizer: data race", true},
		{"ascii_character_write", "string_slice_impl.h", "return &characters[value];", "characters[value].units = 2; return &characters[value];", "string_views.c", "", "4", "WARNING: ThreadSanitizer: data race", true},
		{"skip_items_share", "parallel.c", "adamic_share(items);", "/* mutant: publication without preparation */", "map.c", "strings", "4", "WARNING: ThreadSanitizer: data race", true},
		{"lazy_cache", "string_index.c", "if (__atomic_compare_exchange_n(&((adamic_string *)string)->index, &expected, candidate,\n\t\tfalse, __ATOMIC_RELEASE, __ATOMIC_ACQUIRE)) { return candidate; }", "((adamic_string *)string)->index = candidate; return candidate;", "cache_race.c", "", "4", "WARNING: ThreadSanitizer: data race", true},
		{"fixture_shared_counter", "", "for (size_t i = 0; i < 256; i++) { adamic_retain(shared_text); }", "for (size_t i = 0; i < 256; i++) { shared_text->heap.references++; }", "memory.c", "", "4", "WARNING: ThreadSanitizer: data race", true},
		{"plain_shared_count", "heap.c", "__atomic_fetch_add(&heap->references, 1, __ATOMIC_RELAXED);", "heap->references++;", "memory.c", "", "4", "WARNING: ThreadSanitizer: data race", true},
		{"field_cache", "adamic.h", "static inline adamic_value *adamic_object_data_field(const adamic_object *object, const char *name, adamic_slot_cache *cache) {\n\tuint64_t packed = __atomic_load_n(&cache->packed, __ATOMIC_RELAXED)", "static inline adamic_value *adamic_object_data_field(const adamic_object *object, const char *name, adamic_slot_cache *cache) {\n\tuint64_t packed = cache->packed", "cache_pairs.c", "", "4", "WARNING: ThreadSanitizer: data race", true},
		{"remote_free", "heap.c", "remote_slot *node = malloc(sizeof *node);", "give_local(slot, each); return; remote_slot *node = malloc(sizeof *node);", "memory.c", "", "4", "WARNING: ThreadSanitizer: data race", true},
		{"result_order", "parallel.c", "scope->results->elements[index] = result;", "scope->results->elements[scope->items->length - 1 - index] = result;", "map.c", "numbers", "4", "index 0:", false},
		{"reused_graph", "share.c", "if (visited(&pending, heap)) { continue; }", "if (shared || visited(&pending, heap)) { continue; }", "lifecycle.c", "", "4", "", false},
		{"oversized_slot", "object.c", "uint64_t packed = index <= UINT16_MAX ? pointer | ((uint64_t)index << 48) : 0;", "uint64_t packed = pointer | ((uint64_t)index << 48);", "scaling.c", "", "4", "", false},
		{"fixed_grain", "parallel.c", "return grain == 0 ? 1 : grain > 256 ? 256 : grain;", "return 256;", "scaling.c", "", "4", "", false},
		{"eager_strings", "share.c", "if (!shared) {", "if (!shared && heap->kind == adamic_kind_string) { adamic_string_prepare_shared((adamic_string *)heap); } if (!shared) {", "scaling.c", "", "4", "", false},
		{"one_worker", "parallel.c", "if (thread_count == 1) { return; }", "if (thread_count == 1) { thread_count = 2; }", "map.c", "numbers", "1", "", false},
	}
	for _, mutant := range mutants {
		if mutant.name != name {
			continue
		}
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
				ctx, cancel := context.WithTimeout(context.Background(), time.Minute)

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

// Linux uses LeakSanitizer; Darwin also checks the counted build with leaks.
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
	if goruntime.GOOS != "darwin" || (options.Sanitize || options.ThreadSanitize) {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
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

func TestParallelMutant_view_parent_units(t *testing.T) {
	t.Parallel()
	parallelCheckMutant(t, "view_parent_units")
}

func TestParallelMutant_view_owner_count(t *testing.T) {
	t.Parallel()
	parallelCheckMutant(t, "view_owner_count")
}

func TestParallelMutant_ascii_character_write(t *testing.T) {
	t.Parallel()
	parallelCheckMutant(t, "ascii_character_write")
}

func TestParallelMutant_skip_items_share(t *testing.T) {
	t.Parallel()
	parallelCheckMutant(t, "skip_items_share")
}

func TestParallelMutant_lazy_cache(t *testing.T) {
	t.Parallel()
	parallelCheckMutant(t, "lazy_cache")
}

func TestParallelMutant_plain_shared_count(t *testing.T) {
	t.Parallel()
	parallelCheckMutant(t, "plain_shared_count")
}

func TestParallelMutant_field_cache(t *testing.T) {
	t.Parallel()
	parallelCheckMutant(t, "field_cache")
}

func TestParallelMutant_remote_free(t *testing.T) {
	t.Parallel()
	parallelCheckMutant(t, "remote_free")
}

func TestParallelMutant_result_order(t *testing.T) {
	t.Parallel()
	parallelCheckMutant(t, "result_order")
}

func TestParallelMutant_reused_graph(t *testing.T) {
	t.Parallel()
	parallelCheckMutant(t, "reused_graph")
}

func TestParallelMutant_oversized_slot(t *testing.T) {
	t.Parallel()
	parallelCheckMutant(t, "oversized_slot")
}

func TestParallelMutant_fixed_grain(t *testing.T) {
	t.Parallel()
	parallelCheckMutant(t, "fixed_grain")
}

func TestParallelMutant_eager_strings(t *testing.T) {
	t.Parallel()
	parallelCheckMutant(t, "eager_strings")
}

func TestParallelMutant_one_worker(t *testing.T) {
	t.Parallel()
	parallelCheckMutant(t, "one_worker")
}

func TestParallelMap_numbers_asan(t *testing.T) {
	t.Parallel()
	parallelMapMode(t, "numbers", Options{Sanitize: true})
}

func TestParallelMap_numbers_asan_slabs(t *testing.T) {
	t.Parallel()
	parallelMapMode(t, "numbers", Options{Sanitize: true, slabs: true})
}

func TestParallelMap_numbers_count(t *testing.T) {
	t.Parallel()
	parallelMapMode(t, "numbers", Options{Count: true})
}

func TestParallelMap_numbers_release(t *testing.T) {
	t.Parallel()
	parallelMapMode(t, "numbers", Options{})
}

func TestParallelMap_numbers_tsan(t *testing.T) {
	t.Parallel()
	if !parallelTSan(t) {
		t.Skip("TSan is unavailable on this platform; see capability probe")
	}
	parallelMapMode(t, "numbers", Options{ThreadSanitize: true})
}

func TestParallelMap_strings_asan(t *testing.T) {
	t.Parallel()
	parallelMapMode(t, "strings", Options{Sanitize: true})
}

func TestParallelMap_strings_asan_slabs(t *testing.T) {
	t.Parallel()
	parallelMapMode(t, "strings", Options{Sanitize: true, slabs: true})
}

func TestParallelMap_strings_count(t *testing.T) {
	t.Parallel()
	parallelMapMode(t, "strings", Options{Count: true})
}

func TestParallelMap_strings_release(t *testing.T) {
	t.Parallel()
	parallelMapMode(t, "strings", Options{})
}

func TestParallelMap_strings_tsan(t *testing.T) {
	t.Parallel()
	if !parallelTSan(t) {
		t.Skip("TSan is unavailable on this platform; see capability probe")
	}
	parallelMapMode(t, "strings", Options{ThreadSanitize: true})
}

func TestParallelMap_objects_asan(t *testing.T) {
	t.Parallel()
	parallelMapMode(t, "objects", Options{Sanitize: true})
}

func TestParallelMap_objects_asan_slabs(t *testing.T) {
	t.Parallel()
	parallelMapMode(t, "objects", Options{Sanitize: true, slabs: true})
}

func TestParallelMap_objects_count(t *testing.T) {
	t.Parallel()
	parallelMapMode(t, "objects", Options{Count: true})
}

func TestParallelMap_objects_release(t *testing.T) {
	t.Parallel()
	parallelMapMode(t, "objects", Options{})
}

func TestParallelMap_objects_tsan(t *testing.T) {
	t.Parallel()
	if !parallelTSan(t) {
		t.Skip("TSan is unavailable on this platform; see capability probe")
	}
	parallelMapMode(t, "objects", Options{ThreadSanitize: true})
}

func TestParallelMap_map_asan(t *testing.T) {
	t.Parallel()
	parallelMapMode(t, "map", Options{Sanitize: true})
}

func TestParallelMap_map_asan_slabs(t *testing.T) {
	t.Parallel()
	parallelMapMode(t, "map", Options{Sanitize: true, slabs: true})
}

func TestParallelMap_map_count(t *testing.T) {
	t.Parallel()
	parallelMapMode(t, "map", Options{Count: true})
}

func TestParallelMap_map_release(t *testing.T) {
	t.Parallel()
	parallelMapMode(t, "map", Options{})
}

func TestParallelMap_map_tsan(t *testing.T) {
	t.Parallel()
	if !parallelTSan(t) {
		t.Skip("TSan is unavailable on this platform; see capability probe")
	}
	parallelMapMode(t, "map", Options{ThreadSanitize: true})
}

func TestParallelMap_fresh_asan(t *testing.T) {
	t.Parallel()
	parallelMapMode(t, "fresh", Options{Sanitize: true})
}

func TestParallelMap_fresh_asan_slabs(t *testing.T) {
	t.Parallel()
	parallelMapMode(t, "fresh", Options{Sanitize: true, slabs: true})
}

func TestParallelMap_fresh_count(t *testing.T) {
	t.Parallel()
	parallelMapMode(t, "fresh", Options{Count: true})
}

func TestParallelMap_fresh_release(t *testing.T) {
	t.Parallel()
	parallelMapMode(t, "fresh", Options{})
}

func TestParallelMap_fresh_tsan(t *testing.T) {
	t.Parallel()
	if !parallelTSan(t) {
		t.Skip("TSan is unavailable on this platform; see capability probe")
	}
	parallelMapMode(t, "fresh", Options{ThreadSanitize: true})
}

func TestParallelMap_nested_asan(t *testing.T) {
	t.Parallel()
	parallelMapMode(t, "nested", Options{Sanitize: true})
}

func TestParallelMap_nested_asan_slabs(t *testing.T) {
	t.Parallel()
	parallelMapMode(t, "nested", Options{Sanitize: true, slabs: true})
}

func TestParallelMap_nested_count(t *testing.T) {
	t.Parallel()
	parallelMapMode(t, "nested", Options{Count: true})
}

func TestParallelMap_nested_release(t *testing.T) {
	t.Parallel()
	parallelMapMode(t, "nested", Options{})
}

func TestParallelMap_nested_tsan(t *testing.T) {
	t.Parallel()
	if !parallelTSan(t) {
		t.Skip("TSan is unavailable on this platform; see capability probe")
	}
	parallelMapMode(t, "nested", Options{ThreadSanitize: true})
}

func TestParallelMap_exception_asan(t *testing.T) {
	t.Parallel()
	parallelMapMode(t, "exception", Options{Sanitize: true})
}

func TestParallelMap_exception_asan_slabs(t *testing.T) {
	t.Parallel()
	parallelMapMode(t, "exception", Options{Sanitize: true, slabs: true})
}

func TestParallelMap_exception_count(t *testing.T) {
	t.Parallel()
	parallelMapMode(t, "exception", Options{Count: true})
}

func TestParallelMap_exception_release(t *testing.T) {
	t.Parallel()
	parallelMapMode(t, "exception", Options{})
}

func TestParallelMap_exception_tsan(t *testing.T) {
	t.Parallel()
	if !parallelTSan(t) {
		t.Skip("TSan is unavailable on this platform; see capability probe")
	}
	parallelMapMode(t, "exception", Options{ThreadSanitize: true})
}

func TestParallelMap_nested_exception_asan(t *testing.T) {
	t.Parallel()
	parallelMapMode(t, "nested_exception", Options{Sanitize: true})
}

func TestParallelMap_nested_exception_asan_slabs(t *testing.T) {
	t.Parallel()
	parallelMapMode(t, "nested_exception", Options{Sanitize: true, slabs: true})
}

func TestParallelMap_nested_exception_count(t *testing.T) {
	t.Parallel()
	parallelMapMode(t, "nested_exception", Options{Count: true})
}

func TestParallelMap_nested_exception_release(t *testing.T) {
	t.Parallel()
	parallelMapMode(t, "nested_exception", Options{})
}

func TestParallelMap_nested_exception_tsan(t *testing.T) {
	t.Parallel()
	if !parallelTSan(t) {
		t.Skip("TSan is unavailable on this platform; see capability probe")
	}
	parallelMapMode(t, "nested_exception", Options{ThreadSanitize: true})
}

func TestParallelMap_million_asan(t *testing.T) {
	t.Parallel()
	parallelMapMode(t, "million", Options{Sanitize: true})
}

func TestParallelMap_million_asan_slabs(t *testing.T) {
	t.Parallel()
	parallelMapMode(t, "million", Options{Sanitize: true, slabs: true})
}

func TestParallelMap_million_count(t *testing.T) {
	t.Parallel()
	parallelMapMode(t, "million", Options{Count: true})
}

func TestParallelMap_million_release(t *testing.T) {
	t.Parallel()
	parallelMapMode(t, "million", Options{})
}

func TestParallelMap_million_tsan(t *testing.T) {
	t.Parallel()
	if !parallelTSan(t) {
		t.Skip("TSan is unavailable on this platform; see capability probe")
	}
	parallelMapMode(t, "million", Options{ThreadSanitize: true})
}

func TestParallelMutant_fixture_shared_counter(t *testing.T) {
	t.Parallel()
	parallelCheckMutant(t, "fixture_shared_counter")
}
