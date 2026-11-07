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
	run, _ := timerRunWithin(15*time.Second, environment, name, args...)
	return run
}
func timerRunWithin(limit time.Duration, environment []string, name string, args ...string) (leakcheck.Run, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), limit)
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
	return leakcheck.Run{Stdout: stdout.Bytes(), Stderr: stderr.Bytes(), ExitCode: exit}, ctx.Err() == context.DeadlineExceeded
}
func timerNode(t *testing.T, mode string) []byte {
	t.Helper()
	// .a is deliberate: no new TypeScript source files in an Adamic unit.
	run := timerNodeRun(t, mode, false)
	if run.ExitCode != 0 || len(run.Stderr) != 0 {
		t.Fatalf("Node: %+v", run)
	}
	return run.Stdout
}

func timerNodeRun(t *testing.T, mode string, formatted bool) leakcheck.Run {
	t.Helper()
	args := []string{}
	if formatted {
		// The unchanged house oracle reports uncaught errors exactly as
		// adamic_uncaught: one panic line and exit 70 (raw Node exits 1).
		runtime, err := filepath.Abs("../../oracle/adamic.mjs")
		if err != nil {
			t.Fatal(err)
		}
		args = append(args, "--import", runtime)
	}
	args = append(args, "--input-type=commonjs", "-e",
		`require('node:vm').runInThisContext(require('node:fs').readFileSync(process.argv[1], 'utf8'))`,
		"testdata/timers/oracle.a", mode)
	return timerRun(nil, "node", args...)
}

const timerPhaseDeadline = 3 * time.Second

func TestTimerTurnPhaseYieldsToHostCompletion(t *testing.T) {
	t.Parallel()
	run, expired := timerRunWithin(timerPhaseDeadline, nil, timerProduct(t, "release", native.Options{Release: true}), "turn-phase")
	if expired {
		t.Fatal("timer phase starved the pending host completion: own 3 s deadline")
	}
	if run.ExitCode != 0 || len(run.Stderr) != 0 || !bytes.Equal(run.Stdout, timerNode(t, "turn-phase")) {
		t.Fatalf("timer phase must yield to Node's pending completion: %+v", run)
	}
}

func timerThrowOracle(t *testing.T) leakcheck.Run {
	t.Helper()
	run := timerNodeRun(t, "throw", true)
	if run.ExitCode != 70 || string(run.Stderr) != "adamic: panic: Error: timer callback failed\n" {
		t.Fatalf("house Node uncaught format: %+v", run)
	}
	return run
}
func TestTimerCallbackThrowMatchesNode(t *testing.T) {
	t.Parallel()
	raw := timerNodeRun(t, "throw", false)
	found := false
	for _, line := range strings.Split(string(raw.Stderr), "\n") {
		found = found || line == "Error: timer callback failed"
	}
	if raw.ExitCode != 1 || !found {
		t.Fatalf("raw Node must throw: %+v", raw)
	}
	oracle := timerThrowOracle(t)
	released := timerRun(nil, timerProduct(t, "release", native.Options{Release: true}), "throw")
	if released.ExitCode != oracle.ExitCode || !bytes.Equal(released.Stdout, oracle.Stdout) || !bytes.Equal(released.Stderr, oracle.Stderr) {
		t.Fatalf("uncaught callback differs from Node's house format: native %+v Node %+v", released, oracle)
	}
	checked := timerRun([]string{"ASAN_OPTIONS=detect_leaks=0:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"}, timerProduct(t, "sanitized", native.Options{Sanitize: true, Count: true}), "throw")
	if checked.ExitCode != oracle.ExitCode || !bytes.Equal(checked.Stdout, oracle.Stdout) || !bytes.HasPrefix(checked.Stderr, oracle.Stderr) || bytes.Contains(checked.Stderr, []byte("Sanitizer")) || bytes.Contains(checked.Stderr, []byte("runtime error:")) {
		t.Fatalf("sanitized uncaught callback: %+v", checked)
	}
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
	for _, mode := range []string{"turn-phase", "order", "interval", "cancel", "cancel-self", "foreign", "unref", "unref-live", "deadlines", "graph", "arity", "host"} {
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

// build links code against this mutant runtime the way native.Build links a product: through clangIn, with the build
// directory and the runtime's directory mapped out of everything the compile records of a path. A bare clang here put
// the product's random staging directory into __FILE__ and the debug info, so no two builds of one mutant matched, even
// on one box (Loom's Builder, #landing90, Nov 2). TestTimerMutantProductIsReproducible pins it, and
// TestTimerMutantProductPinCatchesBareClang shows the pin goes red with the old bare clang.
func (runtime timerRuntime) build(code, output string, options native.Options) error {
	library, err := native.RuntimeLibrary(runtime.directory, options)
	if err != nil {
		return err
	}
	if err := native.BuildWithRuntime(code, output, options, library); err != nil {
		return fmt.Errorf("mutant must compile: %w", err)
	}
	return nil
}

// timerBuildBare is the probe build this file used before #landing90: the source written beside the output and compiled
// by a bare clang. It stays only as the pin's mutant.
func timerBuildBare(code, output string, options native.Options, library string) error {
	source := output + ".c"
	if err := os.WriteFile(source, []byte(code), 0600); err != nil {
		return err
	}
	args := append(native.LinkFlags(options), "-I", filepath.Dir(library), "-o", output, source)
	args = append(args, native.RuntimeLinkFlags(library)...)
	args = append(args, "-lm")
	result, err := exec.Command("clang", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("mutant must compile: %w\n%s", err, result)
	}
	return nil
}

// timerMutantTwice builds one mutant runtime, then links its sanitized probe into two output directories.
// RuntimeLibrary already maps the runtime directory; the pin varies the probe's build and output directories.
func timerMutantTwice(t *testing.T, bare bool) ([]byte, []byte) {
	t.Helper()
	mutation := timerMutations[5]
	if mutation.name != "unref-ignored" {
		t.Fatalf("timerMutations[5] is %s, want unref-ignored", mutation.name)
	}
	options := native.Options{Sanitize: true, Count: true}
	runtime := timerMutant(t, mutation.file, mutation.before, mutation.after)
	library, err := native.RuntimeLibrary(runtime.directory, options)
	if err != nil {
		t.Fatal(err)
	}
	build := native.BuildWithRuntime
	if bare {
		build = timerBuildBare
	}
	code := timerHarness(t)
	var built [2][]byte
	for index := range built {
		output := filepath.Join(t.TempDir(), "probe")
		if err := build(code, output, options, library); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(output)
		if err != nil {
			t.Fatal(err)
		}
		built[index] = data
	}
	return built[0], built[1]
}

func TestTimerMutantProductIsReproducible(t *testing.T) {
	t.Parallel()
	first, second := timerMutantTwice(t, false)
	if !bytes.Equal(first, second) {
		t.Fatalf("one mutant built in two directories differs (%d and %d bytes): a path reached the binary", len(first), len(second))
	}
}

func TestTimerMutantProductPinCatchesBareClang(t *testing.T) {
	t.Parallel()
	first, second := timerMutantTwice(t, true)
	if bytes.Equal(first, second) {
		t.Fatal("the bare-clang build matched across directories, so the reproducibility pin proves nothing")
	}
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
	{"turn-phase-never-yields", "turn-phase", "async.c", "entry->deadline <= turn_time", "entry->deadline <= timer_now()", "deadline"},
	{"callback-uncaught-omitted", "throw", "timers_impl.h", "    if (adamic_thrown != NULL) adamic_uncaught();", "    /* mutant: ignore the callback's pending exception */", "throw"},
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
			if mutant.caught == "deadline" {
				run, expired := timerRunWithin(timerPhaseDeadline, []string{"ASAN_OPTIONS=detect_leaks=0:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary, mutant.mode)
				if !expired || bytes.Count(run.Stdout, []byte("interval\n")) < 2 || bytes.Contains(run.Stdout, []byte("host\n")) || len(run.Stderr) != 0 {
					t.Fatalf("phase mutant must loop on the interval until its own deadline: expired=%t %+v", expired, run)
				}
				t.Log("caught by TestTimerTurnPhaseYieldsToHostCompletion's own 3 s deadline")
				return
			}
			if mutant.caught == "throw" {
				run := timerRun([]string{"ASAN_OPTIONS=detect_leaks=0:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary, mutant.mode)
				oracle := timerThrowOracle(t)
				if run.ExitCode == oracle.ExitCode && bytes.HasPrefix(run.Stderr, oracle.Stderr) {
					t.Fatal("uncaught callback mutant survived Node comparison")
				}
				if run.ExitCode != 0 {
					t.Fatalf("dropped uncaught must finish silently: %+v", run)
				}
				t.Log("caught by TestTimerCallbackThrowMatchesNode's exit and error-line comparison")
				return
			}
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
		Toolchain: []string{goruntime.Version(), goruntime.GOOS + "/" + goruntime.GOARCH, buildcache.CompilerReport("clang"), buildcache.Tool("ar", "--version")}}
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

func TestProduct_TimerMutantTurnPhase(t *testing.T) {
	t.Parallel()
	timerProduct(t, "mutant-turn-phase-never-yields", native.Options{Sanitize: true, Count: true})
}
func TestProduct_TimerMutantCallbackUncaught(t *testing.T) {
	t.Parallel()
	timerProduct(t, "mutant-callback-uncaught-omitted", native.Options{Sanitize: true, Count: true})
}
