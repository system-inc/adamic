package native_test

import (
	"bytes"
	"context"
	"fmt"
	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/compilerinputs"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
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

const timerCustomHooksMessage = "adamic: panic: NotYet: timers with custom host loop hooks\n"

func timerCustomHooksRefused(run leakcheck.Run) bool {
	return run.ExitCode == 70 && bytes.HasPrefix(run.Stderr, []byte(timerCustomHooksMessage))
}

func TestTimerCustomHooksRefused(t *testing.T) {
	t.Parallel()
	run := timerRun(nil, timerProduct(t, "release", native.Options{Release: true}), "custom-hooks")
	if !timerCustomHooksRefused(run) || string(run.Stderr) != timerCustomHooksMessage {
		t.Fatalf("custom hooks must refuse with the exact provider NotYet: %+v", run)
	}
}

func TestTimersAgainstNode(t *testing.T) {
	t.Parallel()
	code := timerHarness(t)
	binary := timerProduct(t, "sanitized", native.Options{Sanitize: true, Count: true})
	released := timerProduct(t, "release", native.Options{Release: true})
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
			report, err := leakcheck.Check(leakcheck.Program{C: code, Sanitized: binary, Counted: filepath.Join(t.TempDir(), "counted"),
				Arguments: func() []string { return []string{mode} }, Execute: timerRun,
				BuildCounted: func(_ string, output string) error {
					return copyTimerProduct(timerProduct(t, "counted", native.Options{Count: true}), output)
				}})
			if err != nil || report != "" {
				t.Fatalf("leakcheck: %v %s", err, report)
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
	args := append(native.Flags(options), "-I", filepath.Dir(library), "-o", output, source)
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

type timerMutation struct{ name, mode, file, before, after, caught string }

var timerMutations = []timerMutation{
	{"custom-hooks-admitted", "custom-hooks", "timers_impl.h", "if (host_wait_hook != host_default_wait) {", "if (false) {", "custom-hooks"},
	{"released-before-last-fire", "interval", "timers_impl.h", "chosen->firing = false;", "if (chosen->active) adamic_release(chosen->callback);\n    chosen->firing = false;", "heap-use-after-free"},
	{"foreign-cancellation", "foreign", "timers_impl.h", "if (handle.provider != &timer_provider) return NULL;", "(void)handle.provider;", "Assertion"},
	{"unref-ignored", "unref", "timers_impl.h", "entry->referenced = false;", "entry->referenced = true;", "Node"},
	{"cancel-leaks-registration", "cancel", "timers_impl.h", "if (!entry->firing) timer_drop(entry);", "if (!entry->firing) { /* deliberately leak */ }", "leak"},
	{"cancel-frees-running-callback", "cancel-self", "timers_impl.h", "if (!entry->firing) timer_drop(entry);", "timer_drop(entry);", "heap-use-after-free"},
	{"promise-checkpoint-omitted", "order", "async.c", "            if (!due || !timer_dispatch()) break;\n            drain_jobs();", "            if (!due || !timer_dispatch()) break;", "Node"},
}

func TestTimerMutants(t *testing.T) {
	t.Parallel()
	for _, mutant := range timerMutations {
		t.Run(mutant.name, func(t *testing.T) {
			binary := timerProduct(t, "mutant-"+mutant.name, native.Options{Sanitize: true, Count: true})
			code := timerHarness(t)
			if mutant.caught == "leak" {
				report, err := leakcheck.Check(leakcheck.Program{C: code, Sanitized: binary, Counted: filepath.Join(t.TempDir(), "counted"),
					BuildCounted: func(code, output string) error {
						return copyTimerProduct(timerProduct(t, "counted-"+mutant.name, native.Options{Count: true}), output)
					},
					Arguments: func() []string { return []string{mutant.mode} }, Execute: timerRun})
				if err != nil || (!strings.Contains(report, "LeakSanitizer: detected memory leaks") && !strings.Contains(report, "heap values leaked:")) {
					t.Fatalf("leak mutant must fail specifically the leak check: %v %q", err, report)
				}
				t.Log("caught by shared leakcheck")
				return
			}
			run := timerRun([]string{"ASAN_OPTIONS=detect_leaks=0:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary, mutant.mode)
			if mutant.caught == "custom-hooks" {
				if timerCustomHooksRefused(run) {
					t.Fatal("custom-hooks refusal pin did not catch mutant")
				}
				if run.ExitCode != 0 {
					t.Fatalf("admitted mutant must reach the harness return: %+v", run)
				}
			} else if mutant.caught == "Node" {
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

// Products contain compiled probes only. Every comparison, sanitizer run and
// mutant assertion still executes in the ordinary test on the runner.
func timerProduct(t *testing.T, variant string, options native.Options) string {
	t.Helper()
	flags := append([]string{fmt.Sprintf("options=%+v", options)}, native.Flags(options)...)
	for _, variable := range []string{"ADAMIC_PROGRAM_REGION", "ADAMIC_NATIVE_SPLIT", "ADAMIC_NATIVE_JOBS", "CPATH", "C_INCLUDE_PATH", "SDKROOT", "MACOSX_DEPLOYMENT_TARGET"} {
		value, set := os.LookupEnv(variable)
		flags = append(flags, fmt.Sprintf("env:%s:set=%t:value=%s", variable, set, value))
	}
	inputs := buildcache.Inputs{Name: "native-timers-" + variant, Files: compilerinputs.Lowering(), Flags: flags,
		Toolchain: []string{goruntime.Version(), goruntime.GOOS + "/" + goruntime.GOARCH, buildcache.Tool("clang", "--version"), buildcache.Tool("ar", "--version")}}
	directory := buildcache.Product(t, inputs, func(directory string) error {
		code := timerHarness(t)
		output := filepath.Join(directory, "probe")
		for _, mutation := range timerMutations {
			if variant == "mutant-"+mutation.name || variant == "counted-"+mutation.name {
				runtime := timerMutant(t, mutation.file, mutation.before, mutation.after)
				return runtime.build(code, output, options)
			}
		}
		if variant != "sanitized" && variant != "release" && variant != "counted" {
			return fmt.Errorf("unknown timer product %s", variant)
		}
		return native.Build(code, output, options)
	})
	buildcache.RequireArtifacts(t, directory, buildcache.Artifact{Name: "probe"})
	return filepath.Join(directory, "probe")
}
func copyTimerProduct(source, destination string) error {
	data, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	return os.WriteFile(destination, data, 0700)
}
func TestProduct_TimersSanitized(t *testing.T) {
	t.Parallel()
	timerProduct(t, "sanitized", native.Options{Sanitize: true, Count: true})
}
func TestProduct_TimersRelease(t *testing.T) {
	t.Parallel()
	timerProduct(t, "release", native.Options{Release: true})
}
func TestProduct_TimersCounted(t *testing.T) {
	t.Parallel()
	timerProduct(t, "counted", native.Options{Count: true})
}
func TestProduct_TimerMutantReleasedBeforeLastFire(t *testing.T) {
	t.Parallel()
	timerProduct(t, "mutant-released-before-last-fire", native.Options{Sanitize: true, Count: true})
}
func TestProduct_TimerMutantForeignCancellation(t *testing.T) {
	t.Parallel()
	timerProduct(t, "mutant-foreign-cancellation", native.Options{Sanitize: true, Count: true})
}
func TestProduct_TimerMutantUnrefIgnored(t *testing.T) {
	t.Parallel()
	timerProduct(t, "mutant-unref-ignored", native.Options{Sanitize: true, Count: true})
}
func TestProduct_TimerMutantCancelLeaksRegistration(t *testing.T) {
	t.Parallel()
	timerProduct(t, "mutant-cancel-leaks-registration", native.Options{Sanitize: true, Count: true})
}
func TestProduct_TimerMutantCancelLeaksCounted(t *testing.T) {
	t.Parallel()
	timerProduct(t, "counted-cancel-leaks-registration", native.Options{Count: true})
}
func TestProduct_TimerMutantCancelFreesCallback(t *testing.T) {
	t.Parallel()
	timerProduct(t, "mutant-cancel-frees-running-callback", native.Options{Sanitize: true, Count: true})
}
func TestProduct_TimerMutantPromiseCheckpoint(t *testing.T) {
	t.Parallel()
	timerProduct(t, "mutant-promise-checkpoint-omitted", native.Options{Sanitize: true, Count: true})
}

func TestProduct_TimerMutantCustomHooks(t *testing.T) {
	t.Parallel()
	timerProduct(t, "mutant-custom-hooks-admitted", native.Options{Sanitize: true, Count: true})
}
