package native

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// The hooks exist only in a cloned runtime; the guarded and mutant snapshots
// have identical gates. No production code or production synchronization changes.
func staticsSteadySource(test *testing.T, file string) string {
	test.Helper()
	source, readError := os.ReadFile(filepath.Join(*runtimeStaticsDirectory, file))
	if readError != nil {
		test.Fatal(readError)
	}
	return string(source)
}

func staticsSteadyPatch(test *testing.T, source *string, old, changed string) {
	test.Helper()
	if strings.Count(*source, old) != 1 {
		test.Fatalf("missing or ambiguous steady race seam: %q", old)
	}
	*source = strings.Replace(*source, old, changed, 1)
}

func staticsNormalizationRaceLibrary(test *testing.T, kind, old, mutant string, unsafe bool) (string, string) {
	test.Helper()
	source := staticsSteadySource(test, "normalize.c")
	changed := source
	// Gate the real write and every corresponding cache-hit read. Identical inputs
	// avoid corrupting normalization output before TSan can diagnose the mutant.
	staticsSteadyPatch(
		test,
		&changed,
		"#include \"adamic.h\"",
		"#include \"adamic.h\"\nextern void adamic_test_normalization_before(void);\nextern void adamic_test_normalization_after(void);",
	)
	var declaration, write, hit string
	switch kind {
	case "normalization_classes":
		declaration = "static _Thread_local uint8_t cached_classes[64];"
		write = "cached_classes[slot] = class;"
		hit = "return cached_classes[slot];"
	case "normalization_mappings":
		declaration = "static _Thread_local const normalize_mapping *cached_mappings[64];"
		write = "cached_mappings[slot] = mapping;"
		hit = "return cached_mappings[slot];"
	case "normalization_pairs":
		declaration = "static _Thread_local normalize_pair cached_pairs[64];"
		write = "cached_pairs[slot] = (normalize_pair){first, second, joined};"
		hit = "return cached_pairs[slot].composite;"
	default:
		test.Fatalf("unknown normalization gate: %s", kind)
	}
	staticsSteadyPatch(test, &changed, declaration, declaration+"\n\tadamic_test_normalization_before();")
	staticsSteadyPatch(test, &changed, write, write+"\n\tadamic_test_normalization_after();")
	// Keep the read before the release hook, also on the global pair-cache hit.
	expression := strings.TrimSuffix(strings.TrimPrefix(hit, "return "), ";")
	resultType := "uint32_t"
	if kind == "normalization_classes" {
		resultType = "uint8_t"
	}
	if kind == "normalization_mappings" {
		resultType = "const normalize_mapping *"
	}
	staticsSteadyPatch(
		test,
		&changed,
		hit,
		resultType+" result = "+expression+";\n\t\tadamic_test_normalization_after();\n\t\treturn result;",
	)
	if unsafe {
		staticsSteadyPatch(test, &changed, old, mutant)
	}
	return staticsRaceLibrary(test, "normalize.c", source, changed)
}

func staticsExitRaceLibrary(test *testing.T, mutant bool) (string, string) {
	test.Helper()
	source := staticsSteadySource(test, "adamic.c")
	changed := source
	staticsSteadyPatch(
		test,
		&changed,
		"static void buffer(const char *bytes, size_t length) {",
		"extern void adamic_test_exit_writer(void);\n\nstatic void buffer(const char *bytes, size_t length) {",
	)
	staticsSteadyPatch(test, &changed, "output_used += length;", "output_used += length;\n\tadamic_test_exit_writer();")
	staticsSteadyPatch(
		test,
		&changed,
		"static void finish(void) {",
		"extern void adamic_test_exit_release(void);\n\nstatic void finish(void) {",
	)
	seam := "adamic_parallel_shutdown();\n\tpthread_mutex_lock(&output_lock);\n\tflush();\n\tbool failed = broken[adamic_stdout] || broken[adamic_stderr];\n\tpthread_mutex_unlock(&output_lock);"
	replacement := "adamic_test_exit_release();\n\t" + seam
	if mutant {
		replacement = "flush();\n\tadamic_test_exit_release();\n\tbool failed = broken[adamic_stdout] || broken[adamic_stderr];"
	}
	staticsSteadyPatch(test, &changed, seam, replacement)
	return staticsRaceLibrary(test, "adamic.c", source, changed)
}

// Each ordinary gate invocation checks both pool sizes, three fresh processes
// per size. Crashes and timeouts without explicit data-race reports never pass.
func staticsSteadyRun(test *testing.T, binary, kind string, mutant bool) {
	test.Helper()
	for _, threads := range []string{"4", "16"} {
		test.Run("threads_"+threads, func(test *testing.T) {
			for attempt := 1; attempt <= 3; attempt++ {
				ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
				command := exec.CommandContext(ctx, binary)
				env := os.Environ()
				for _, key := range []string{"ADAMIC_THREADS", "TSAN_OPTIONS", "ADAMIC_NORMALIZATION_KIND"} {
					filtered := env[:0]
					for _, entry := range env {
						if !strings.HasPrefix(entry, key+"=") {
							filtered = append(filtered, entry)
						}
					}
					env = filtered
				}
				command.Env = append(
					env,
					"ADAMIC_THREADS="+threads,
					"TSAN_OPTIONS=halt_on_error=1:exitcode=66",
					"ADAMIC_NORMALIZATION_KIND="+kind,
				)
				output, combinedOutputError := command.CombinedOutput()
				timedOut := ctx.Err() != nil
				cancel()
				function := map[string]string{
					"normalization_classes":  "combining_class",
					"normalization_mappings": "mapping_of",
					"normalization_pairs":    "composite",
					"exit_flush":             "flush",
				}[kind]
				report := strings.Contains(string(output), "WARNING: ThreadSanitizer: data race") &&
					strings.Contains(string(output), " in "+function+"\n")
				if mutant {
					if combinedOutputError == nil || timedOut || !report {
						test.Fatalf(
							"%s mutant not caught (attempt %d): %v\n%s",
							kind,
							attempt,
							combinedOutputError,
							output,
						)
					}
					for _, line := range strings.Split(string(output), "\n") {
						if strings.Contains(line, "SUMMARY: ThreadSanitizer:") {
							test.Log(line)
						}
					}
				} else {
					expected := "runtime statics clean\n"
					if kind == "exit_flush" {
						expected = "signal worker line\n"
					}
					if combinedOutputError != nil ||
						timedOut ||
						strings.Contains(string(output), "ThreadSanitizer:") ||
						!strings.Contains(string(output), expected) {
						test.Fatalf(
							"%s guarded fixture failed (attempt %d): %v\n%s",
							kind,
							attempt,
							combinedOutputError,
							output,
						)
					}
				}
			}
		})
	}
}
