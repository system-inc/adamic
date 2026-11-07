package native_test

import (
	"bytes"
	"context"
	"fmt"
	"github.com/system-inc/adamic/internal/leakcheck"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/compilerinputs"
	"github.com/system-inc/adamic/internal/native"
)

func watchHooksProduct(test *testing.T, mutant bool) string {
	variant := "control"
	if mutant {
		variant = "admitted"
	}
	return watchProviderProduct(test, variant)
}
func watchProviderProduct(test *testing.T, variant string) string {
	test.Helper()
	options := native.Options{Release: true}
	if variant == "sanitized" {
		options = native.Options{Sanitize: true}
	}
	inputs := watchProviderInputs(variant, options)

	directory := buildcache.Product(test, inputs, func(directory string) error {
		code, err := os.ReadFile("testdata/watchfile/hooks.c")
		if err != nil {
			return err
		}
		output := filepath.Join(directory, "probe")
		if variant == "control" || variant == "sanitized" {
			return native.Build(string(code), output, options)
		}
		var changed timerRuntime
		if variant == "interval-replaced" {
			changed = timerMutant(
				test,
				"node_watch_file_impl.h",
				"} else free(normal);",
				"} else { free(normal); poller->interval = interval < 1 ? 1 : (unsigned)interval; }",
			)
		} else if variant == "batch-early-jobs" {
			changed = timerMutant(
				test,
				"node_watch_file_impl.h",
				"(void)adamic_closure_call(batch[index], arguments, 2);",
				"(void)adamic_closure_call(batch[index], arguments, 2); adamic_async_run();",
			)
		} else {
			changed = timerMutant(test, "timers_impl.h", "if (host_wait_hook != host_default_wait)", "if (false)")
		}
		library, err := native.RuntimeLibraryForSource(changed.directory, string(code), options)
		if err != nil {
			return err
		}
		return native.BuildWithRuntime(string(code), output, options, library)
	})
	buildcache.RequireArtifacts(test, directory, buildcache.Artifact{Name: "probe"})
	return filepath.Join(directory, "probe")
}
func TestWatchFileCustomHooksRefused(test *testing.T) {
	test.Parallel()
	got := timerRun(nil, watchHooksProduct(test, false))
	const want = "adamic: panic: NotYet: timers with custom host loop hooks\n"
	if got.ExitCode != 70 || string(got.Stderr) != want || len(got.Stdout) != 0 {
		test.Fatalf("watchFile must reach the provider's exact refusal: %+v", got)
	}
}
func TestWatchFileMutantAdmitsCustomHooks(test *testing.T) {
	test.Parallel()
	got := timerRun(nil, watchHooksProduct(test, true))
	if got.ExitCode != 0 || !strings.Contains(string(got.Stdout), "admitted\n") || len(got.Stderr) != 0 {
		test.Fatalf("mutant must admit watchFile cleanly: %+v", got)
	}
	test.Log("TestWatchFileCustomHooksRefused goes red with the provider hook guard removed")
}
func TestProduct_WatchFileCustomHooks(test *testing.T) {
	test.Parallel()
	watchHooksProduct(test, false)
}
func TestProduct_WatchFileCustomHooksMutant(test *testing.T) {
	test.Parallel()
	watchHooksProduct(test, true)
}

func watchProviderRun(test *testing.T, name string, arguments ...string) leakcheck.Run {
	test.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, name, arguments...)
	command.Dir = test.TempDir()
	command.Env = []string{
		"TMPDIR=" + command.Dir,
		"ASAN_OPTIONS=detect_leaks=1:halt_on_error=1",
		"UBSAN_OPTIONS=halt_on_error=1",
	}
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	if command.ProcessState == nil {
		test.Fatal(err)
	}
	if ctx.Err() != nil {
		test.Fatal("watchFile provider own deadline")
	}
	return leakcheck.Run{Stdout: stdout.Bytes(), Stderr: stderr.Bytes(), ExitCode: command.ProcessState.ExitCode()}
}
func watchBatchNode(test *testing.T) leakcheck.Run { return watchProviderNode(test, "batch") }
func watchProviderNode(test *testing.T, mode string) leakcheck.Run {
	test.Helper()
	node, err := exec.LookPath("node")
	if err != nil {
		test.Fatal(err)
	}
	script, err := filepath.Abs("testdata/watchfile/oracle.a")
	if err != nil {
		test.Fatal(err)
	}
	return watchProviderRun(
		test,
		node,
		"--input-type=commonjs",
		"-e",
		`require('node:vm').runInThisContext(require('node:fs').readFileSync(process.argv[1],'utf8'))`,
		script,
		mode,
	)
}
func TestWatchFileListenerBatchBeforeJobsMatchesNode(test *testing.T) {
	test.Parallel()
	truth := watchBatchNode(test)
	if truth.ExitCode != 0 || string(truth.Stdout) != "first\nsecond\njob\n" || len(truth.Stderr) != 0 {
		test.Fatalf("Node: %+v", truth)
	}
	for _, variant := range []string{"control", "sanitized"} {
		got := watchProviderRun(test, watchProviderProduct(test, variant), "batch")
		if got.ExitCode != truth.ExitCode || !bytes.Equal(got.Stdout, truth.Stdout) || len(got.Stderr) != 0 {
			test.Fatalf("%s: %+v", variant, got)
		}
	}
}
func TestWatchFileMutantDrainsJobsInsideBatch(test *testing.T) {
	test.Parallel()
	got := watchProviderRun(test, watchProviderProduct(test, "batch-early-jobs"), "batch")
	if got.ExitCode != 0 || string(got.Stdout) != "first\njob\nsecond\n" || len(got.Stderr) != 0 {
		test.Fatalf("early-jobs mutant: %+v", got)
	}
	test.Log("TestWatchFileListenerBatchBeforeJobsMatchesNode goes red with an inner microtask checkpoint")
}
func TestProduct_WatchFileProviderSanitized(test *testing.T) {
	test.Parallel()
	watchProviderProduct(test, "sanitized")
}
func TestProduct_WatchFileBatchCheckpointMutant(test *testing.T) {
	test.Parallel()
	watchProviderProduct(test, "batch-early-jobs")
}

func watchProviderInputs(variant string, options native.Options) buildcache.Inputs {
	flags := append(native.Flags(options), "ADAMIC_NODE_HOST")
	for _, name := range []string{
		"ADAMIC_PROGRAM_REGION",
		"ADAMIC_NATIVE_SPLIT",
		"ADAMIC_NATIVE_JOBS",
		"CPATH",
		"C_INCLUDE_PATH",
		"SDKROOT",
		"MACOSX_DEPLOYMENT_TARGET",
	} {
		value, set := os.LookupEnv(name)
		flags = append(flags, fmt.Sprintf("env:%s:set=%t:value=%s", name, set, value))
	}
	return buildcache.Inputs{Name: "native-watch-file-hooks-" + variant,
		Files: append(
			compilerinputs.Lowering(),
			"internal/native/watchfile_test.go",
			"internal/native/testdata/watchfile/hooks.c",
		),
		Flags: flags,
		Toolchain: []string{
			goruntime.Version(),
			goruntime.GOOS + "/" + goruntime.GOARCH,
			buildcache.Tool("clang", "--version"),
			buildcache.Tool("ar", "--version"),
		}}
}

// Not parallel: changing the build environment must change the product key.
func TestWatchFileProductKeyIncludesBuildEnvironment(test *testing.T) {
	root := test.TempDir()
	test.Setenv("CPATH", "/first/include")
	first := watchProviderInputs("control", native.Options{Release: true})
	first.Files = nil
	before, err := buildcache.Key(root, first)
	if err != nil {
		test.Fatal(err)
	}
	test.Setenv("CPATH", "/second/include")
	second := watchProviderInputs("control", native.Options{Release: true})
	second.Files = nil
	after, err := buildcache.Key(root, second)
	if err != nil {
		test.Fatal(err)
	}
	if before == after {
		test.Fatal("changed compiler include environment reused the old product key")
	}
}

func TestWatchFileProductKeyIncludesItsHarness(test *testing.T) {
	test.Parallel()
	inputs := watchProviderInputs("control", native.Options{Release: true})
	const harness = "internal/native/testdata/watchfile/hooks.c"
	found := false
	for _, path := range inputs.Files {
		if path == harness {
			found = true
		}
	}
	if !found {
		test.Fatal("a harness edit must invalidate its compiled probe")
	}
	inputs.Files = []string{harness}
	root := test.TempDir()
	path := filepath.Join(root, harness)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		test.Fatal(err)
	}
	code, err := os.ReadFile("testdata/watchfile/hooks.c")
	if err != nil {
		test.Fatal(err)
	}
	if err := os.WriteFile(path, code, 0600); err != nil {
		test.Fatal(err)
	}
	before, err := buildcache.Key(root, inputs)
	if err != nil {
		test.Fatal(err)
	}
	if err := os.WriteFile(path, append(code, []byte("\n/* changed harness */\n")...), 0600); err != nil {
		test.Fatal(err)
	}
	after, err := buildcache.Key(root, inputs)
	if err != nil {
		test.Fatal(err)
	}
	if before == after {
		test.Fatal("changed harness reused the old product key")
	}
}

func TestWatchFileFirstIntervalWinsMatchesNode(test *testing.T) {
	test.Parallel()
	truth := watchProviderNode(test, "interval")
	if truth.ExitCode != 0 || string(truth.Stdout) != "interval true\n" || len(truth.Stderr) != 0 {
		test.Fatalf("Node interval: %+v", truth)
	}
	got := watchProviderRun(test, watchProviderProduct(test, "sanitized"), "interval")
	if got.ExitCode != truth.ExitCode || !bytes.Equal(got.Stdout, truth.Stdout) || len(got.Stderr) != 0 {
		test.Fatalf("native interval: %+v", got)
	}
}
func TestWatchFileMutantReplacesTheFirstInterval(test *testing.T) {
	test.Parallel()
	got := watchProviderRun(test, watchProviderProduct(test, "interval-replaced"), "interval")
	if got.ExitCode != 0 || string(got.Stdout) != "interval false\n" || len(got.Stderr) != 0 {
		test.Fatalf("interval mutant: %+v", got)
	}
	test.Log("TestWatchFileFirstIntervalWinsMatchesNode goes red if another listener replaces the first interval")
}
func TestWatchFileThrowEndsWithLiveRegistration(test *testing.T) {
	test.Parallel()
	truth := watchProviderNode(test, "throw")
	if truth.ExitCode != 1 || !strings.Contains(string(truth.Stderr), "Error: watch listener failed") {
		test.Fatalf("raw Node throw: %+v", truth)
	}
	for _, variant := range []string{"control", "sanitized"} {
		got := watchProviderRun(test, watchProviderProduct(test, variant), "throw")
		if got.ExitCode != 70 ||
			string(got.Stderr) != "adamic: panic: Error: watch listener failed\n" ||
			len(got.Stdout) != 0 {
			test.Fatalf("%s: %+v", variant, got)
		}
	}
}
func TestProduct_WatchFileSharedIntervalMutant(test *testing.T) {
	test.Parallel()
	watchProviderProduct(test, "interval-replaced")
}
