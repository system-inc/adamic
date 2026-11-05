package native_test

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	goruntime "runtime"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/compilerinputs"
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
)

// The readiness source has no Adamic-program caller yet, so a C harness pins it, as the timer provider was pinned:
// a non-blocking pipe is each source, written before the loop or left empty. There is no Node oracle for a raw
// descriptor, so each finishing mode's whole stdout is written here.
func sourceHarness(t *testing.T) string {
	t.Helper()
	source, err := os.ReadFile("testdata/sources/harness.c")
	if err != nil {
		t.Fatal(err)
	}
	return string(source)
}

// sourceOutputs is each mode that finishes, with its whole stdout.
var sourceOutputs = map[string]string{
	"self-close":  "prefix\nsource\n",
	"unref":       "prefix\n",
	"unref-live":  "prefix\nsource\ntimer\n",
	"fairness":    "prefix\ninterval 1\ninterval 2\ninterval 3\nsource repeated\n",
	"close-other": "prefix\nfirst\n",
	"checkpoint":  "prefix\nfirst\njob\nsecond\n",
	"error":       "prefix\nerror\n",
}

// sourceDeadline is each pin's own limit. A loop that spins or waits on nothing it counts never finishes, so the
// pin reads the expiry as the failure rather than leaving it to the test binary's timeout.
const sourceDeadline = 3 * time.Second

var sourceSanitizers = []string{"ASAN_OPTIONS=detect_leaks=0:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"}

// sourceFinishes runs mode on binary, a release build, and says how it failed to finish with its whole stdout, or
// "" when it did. A pin and the mutant it catches read the same answer.
func sourceFinishes(binary, mode string) string {
	run, expired := timerRunWithin(sourceDeadline, nil, binary, mode)
	if expired {
		return fmt.Sprintf("deadline: unfinished after %s with stdout %q", sourceDeadline, run.Stdout)
	}
	if run.ExitCode != 0 || len(run.Stderr) != 0 {
		return fmt.Sprintf("exit: %+v", run)
	}
	if string(run.Stdout) != sourceOutputs[mode] {
		return fmt.Sprintf("stdout: %q, want %q", run.Stdout, sourceOutputs[mode])
	}
	return ""
}

func sourcePin(t *testing.T, mode string) {
	t.Helper()
	if failure := sourceFinishes(sourceProduct(t, "release", native.Options{Release: true}), mode); failure != "" {
		t.Fatal(failure)
	}
}

// A lone source that closes itself from its dispatch ends the loop. The wait's old predicate saw no request and no
// timer, returned without polling, and the loop spun on a source it never read.
func TestSourceSelfCloseExits(t *testing.T) {
	t.Parallel()
	sourcePin(t, "self-close")
}

// A lone unreferenced source keeps nothing alive: the loop ends before any wait.
func TestSourceUnrefExits(t *testing.T) {
	t.Parallel()
	sourcePin(t, "unref")
}

// A source its dispatch never drains is dispatched once a turn, from the wait's record, so the interval beside it
// still fires.
func TestSourceNeverDrainedYieldsToInterval(t *testing.T) {
	t.Parallel()
	sourcePin(t, "fairness")
}

// Each dispatch is a macrotask: the first source's job runs before the second source's dispatch.
func TestSourceDispatchIsAMacrotask(t *testing.T) {
	t.Parallel()
	sourcePin(t, "checkpoint")
}

// Every finishing mode under the sanitizers, its counts balanced and its leaks checked, then again in release.
func TestSources(t *testing.T) {
	t.Parallel()
	code := sourceHarness(t)
	binary := sourceProduct(t, "sanitized", native.Options{Sanitize: true, Count: true})
	released := sourceProduct(t, "release", native.Options{Release: true})
	for _, mode := range slices.Sorted(maps.Keys(sourceOutputs)) {
		t.Run(mode, func(t *testing.T) {
			run := timerRun(sourceSanitizers, binary, mode)
			if run.ExitCode != 0 || string(run.Stdout) != sourceOutputs[mode] {
				t.Fatalf("native: %+v", run)
			}
			if report := leakcheck.Unbalanced(run); report != "" {
				t.Fatal(report)
			}
			if failure := sourceFinishes(released, mode); failure != "" {
				t.Fatalf("release: %s", failure)
			}
			report, err := leakcheck.Check(leakcheck.Program{C: code, Sanitized: binary, Counted: filepath.Join(t.TempDir(), "counted"),
				Arguments: func() []string { return []string{mode} }, Execute: timerRun,
				BuildCounted: func(_ string, output string) error {
					return copyTimerProduct(sourceProduct(t, "counted", native.Options{Count: true}), output)
				}})
			if err != nil || report != "" {
				t.Fatalf("leakcheck: %v %s", err, report)
			}
		})
	}
}

const sourceThrowMessage = "adamic: panic: Error: source dispatch failed\n"

// A dispatch that leaves adamic_thrown reaches adamic_uncaught, as a timer callback's does.
func TestSourceThrowIsUncaught(t *testing.T) {
	t.Parallel()
	released := timerRun(nil, sourceProduct(t, "release", native.Options{Release: true}), "throw")
	if released.ExitCode != 70 || string(released.Stdout) != "prefix\n" || string(released.Stderr) != sourceThrowMessage {
		t.Fatalf("uncaught dispatch: %+v", released)
	}
	checked := timerRun(sourceSanitizers, sourceProduct(t, "sanitized", native.Options{Sanitize: true, Count: true}), "throw")
	if checked.ExitCode != 70 || !strings.HasPrefix(string(checked.Stderr), sourceThrowMessage) || strings.Contains(string(checked.Stderr), "Sanitizer") || strings.Contains(string(checked.Stderr), "runtime error:") {
		t.Fatalf("sanitized uncaught dispatch: %+v", checked)
	}
}

const sourceCustomHooksMessage = "adamic: panic: NotYet: readiness sources with custom host loop hooks\n"

func TestSourceCustomHooksRefused(t *testing.T) {
	t.Parallel()
	run := timerRun(nil, sourceProduct(t, "release", native.Options{Release: true}), "custom-hooks")
	if run.ExitCode != 70 || string(run.Stderr) != sourceCustomHooksMessage {
		t.Fatalf("custom hooks must refuse with the exact provider NotYet: %+v", run)
	}
}

// A blocking descriptor would stall the loop in its own dispatch, so registration stops loudly.
func TestSourceBlockingDescriptorRefused(t *testing.T) {
	t.Parallel()
	run := timerRun(nil, sourceProduct(t, "release", native.Options{Release: true}), "blocking")
	if run.ExitCode == 0 || string(run.Stdout) != "prefix\n" || !strings.HasPrefix(string(run.Stderr), "adamic: readiness source descriptor must be non-blocking\n") {
		t.Fatalf("a blocking descriptor must stop registration: %+v", run)
	}
}

// sourceMutation is a real production edit, the pin that must catch it, and how sourceFinishes reports the catch.
type sourceMutation struct{ name, pin, mode, file, before, after, caught string }

var sourceMutations = []sourceMutation{
	{"predicate-ignores-sources", "TestSourceSelfCloseExits", "self-close", "async_host_impl.h",
		"timeout < 0 && source_total == 0);", "timeout < 0);", "deadline"},
	{"liveness-counts-unref", "TestSourceUnrefExits", "unref", "sources_impl.h",
		"if (entry->referenced) return true;", "if (entry->active) return true;", "deadline"},
	{"readiness-read-live", "TestSourceNeverDrainedYieldsToInterval", "fairness", "sources_impl.h",
		"    chosen->ready = 0; /* At most once a turn: only the next wait records it again. */",
		"    struct pollfd live = {.fd = chosen->descriptor, .events = POLLIN};\n    chosen->ready = poll(&live, 1, 0) == 1 ? ready : 0; /* mutant: readiness read live, never a snapshot */", "deadline"},
	{"dispatch-checkpoint-omitted", "TestSourceDispatchIsAMacrotask", "checkpoint", "async.c",
		"while (source_dispatch()) drain_jobs();", "while (source_dispatch()) {}", "stdout"},
}

func TestSourceMutants(t *testing.T) {
	t.Parallel()
	for _, mutant := range sourceMutations {
		t.Run(mutant.name, func(t *testing.T) {
			failure := sourceFinishes(sourceProduct(t, "mutant-"+mutant.name, native.Options{Release: true}), mutant.mode)
			if failure == "" {
				t.Fatalf("%s did not catch the mutant", mutant.pin)
			}
			if !strings.HasPrefix(failure, mutant.caught+":") {
				t.Fatalf("%s must catch it by %s: %s", mutant.pin, mutant.caught, failure)
			}
			t.Logf("caught by %s's %s check: %s", mutant.pin, mutant.caught, failure)
		})
	}
}

// Products contain compiled probes only. Every run, comparison and mutant assertion still executes in the ordinary
// test on the runner. A mutant runtime is the timers' (timerMutant), linked through clangIn.
func sourceProduct(t *testing.T, variant string, options native.Options) string {
	t.Helper()
	flags := append([]string{fmt.Sprintf("options=%+v", options)}, native.Flags(options)...)
	for _, variable := range []string{"ADAMIC_PROGRAM_REGION", "ADAMIC_NATIVE_SPLIT", "ADAMIC_NATIVE_JOBS", "CPATH", "C_INCLUDE_PATH", "SDKROOT", "MACOSX_DEPLOYMENT_TARGET"} {
		value, set := os.LookupEnv(variable)
		flags = append(flags, fmt.Sprintf("env:%s:set=%t:value=%s", variable, set, value))
	}
	inputs := buildcache.Inputs{Name: "native-sources-" + variant, Files: compilerinputs.Lowering(), Flags: flags,
		Toolchain: []string{goruntime.Version(), goruntime.GOOS + "/" + goruntime.GOARCH, buildcache.Tool("clang", "--version"), buildcache.Tool("ar", "--version")}}
	directory := buildcache.Product(t, inputs, func(directory string) error {
		code := sourceHarness(t)
		output := filepath.Join(directory, "probe")
		for _, mutation := range sourceMutations {
			if variant == "mutant-"+mutation.name {
				runtime := timerMutant(t, mutation.file, mutation.before, mutation.after)
				return runtime.build(code, output, options)
			}
		}
		if variant != "sanitized" && variant != "release" && variant != "counted" {
			return fmt.Errorf("unknown source product %s", variant)
		}
		return native.Build(code, output, options)
	})
	buildcache.RequireArtifacts(t, directory, buildcache.Artifact{Name: "probe"})
	return filepath.Join(directory, "probe")
}
func TestProduct_SourcesSanitized(t *testing.T) {
	t.Parallel()
	sourceProduct(t, "sanitized", native.Options{Sanitize: true, Count: true})
}
func TestProduct_SourcesRelease(t *testing.T) {
	t.Parallel()
	sourceProduct(t, "release", native.Options{Release: true})
}
func TestProduct_SourcesCounted(t *testing.T) {
	t.Parallel()
	sourceProduct(t, "counted", native.Options{Count: true})
}
func TestProduct_SourceMutantPredicate(t *testing.T) {
	t.Parallel()
	sourceProduct(t, "mutant-predicate-ignores-sources", native.Options{Release: true})
}
func TestProduct_SourceMutantLiveness(t *testing.T) {
	t.Parallel()
	sourceProduct(t, "mutant-liveness-counts-unref", native.Options{Release: true})
}
func TestProduct_SourceMutantReadinessLive(t *testing.T) {
	t.Parallel()
	sourceProduct(t, "mutant-readiness-read-live", native.Options{Release: true})
}
func TestProduct_SourceMutantCheckpoint(t *testing.T) {
	t.Parallel()
	sourceProduct(t, "mutant-dispatch-checkpoint-omitted", native.Options{Release: true})
}
