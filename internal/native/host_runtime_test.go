package native_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
)

func hostBuild(test *testing.T, source, output, directory string, options native.Options) {
	test.Helper()
	if directory == "" {
		if buildError := native.Build(source, output, options); buildError != nil {
			test.Fatal(buildError)
		}
		return
	}
	library, runtimeLibraryError := native.RuntimeLibrary(directory, options)
	if runtimeLibraryError != nil {
		test.Fatal(runtimeLibraryError)
	}
	main := filepath.Join(test.TempDir(), "main.c")
	if writeError := os.WriteFile(main, []byte(source), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	args := append(native.Flags(options), "-I", filepath.Dir(library), "-o", output, main)
	args = append(args, native.RuntimeLinkFlags(library)...)
	args = append(args, "-lm")
	if out, combinedOutputError := exec.Command("clang", args...).CombinedOutput(); combinedOutputError != nil {
		test.Fatalf("clang: %v\n%s", combinedOutputError, out)
	}
}
func hostRun(test *testing.T, env []string, program string, args ...string) leakcheck.Run {
	test.Helper()
	command := exec.Command(program, args...)
	command.Env = append(
		os.Environ(),
		"ASAN_OPTIONS=detect_leaks=0",
		"UBSAN_OPTIONS=halt_on_error=1",
		"HOST_EMPTY=",
		"HOST_VALUE=héllo 🌍",
	)
	command.Env = append(command.Env, env...)
	// Test that executable identity doesn't trust invocation argv[0].
	if filepath.Base(program) != "node" {
		command.Args[0] = "forged-argv-zero"
	}
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	for _, value := range env {
		if value == "HOST_MERGE=1" {
			command.Stderr = &stdout
		}
	}
	runError := command.Run()
	if runError != nil {
		if _, ok := runError.(*exec.ExitError); !ok {
			test.Fatal(runError)
		}
	}
	return leakcheck.Run{Stdout: stdout.Bytes(), Stderr: stderr.Bytes(), ExitCode: command.ProcessState.ExitCode()}
}
func hostMutant(test *testing.T, file, before, after string) string {
	test.Helper()
	directory := test.TempDir()
	entries, readDirectoryError := os.ReadDir("runtime")
	if readDirectoryError != nil {
		test.Fatal(readDirectoryError)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, readError := os.ReadFile(filepath.Join("runtime", entry.Name()))
		if readError != nil {
			test.Fatal(readError)
		}
		if entry.Name() == file {
			if strings.Count(string(data), before) != 1 {
				test.Fatalf("mutation %q must hit exactly once", before)
			}
			data = []byte(strings.Replace(string(data), before, after, 1))
		}
		if writeError := os.WriteFile(filepath.Join(directory, entry.Name()), data, 0600); writeError != nil {
			test.Fatal(writeError)
		}
	}
	return directory
}

func hostFixture(test *testing.T) (source, dir, oracle string, env []string) {
	test.Helper()
	data, readError := os.ReadFile("testdata/host-runtime/harness.c")
	if readError != nil {
		test.Fatal(readError)
	}
	source = string(data)
	oracleData, readError := os.ReadFile("testdata/host-runtime/oracle.mjs")
	if readError != nil {
		test.Fatal(readError)
	}
	dir = test.TempDir()
	oracle = filepath.Join(dir, "oracle.mjs")
	if writeError := os.WriteFile(oracle, oracleData, 0600); writeError != nil {
		test.Fatal(writeError)
	}
	// Actual default-library entry, alongside both executable and Node bundle.
	lib, readError := os.ReadFile("../../cohere/TypeScript/tsc/internal/bundled/libs/lib.d.ts")
	if readError != nil {
		test.Fatal(readError)
	}
	if writeError := os.WriteFile(filepath.Join(dir, "lib.d.ts"), lib, 0600); writeError != nil {
		test.Fatal(writeError)
	}
	child := filepath.Join(dir, "child")
	if createDirectoryError := os.Mkdir(child, 0700); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	cwd, workingDirectoryError := os.Getwd()
	if workingDirectoryError != nil {
		test.Fatal(workingDirectoryError)
	}
	if writeError := os.WriteFile(filepath.Join(child, "file"), []byte("data"), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	for name, target := range map[string]string{
		"link":     "file",
		"dangling": "missing",
		"loop-a":   "loop-b",
		"loop-b":   "loop-a",
	} {
		if symlinkError := os.Symlink(target, filepath.Join(child, name)); symlinkError != nil {
			test.Fatal(symlinkError)
		}
	}
	env = []string{"HOST_EXPECT_CWD=" + cwd, "HOST_CHILD=" + child}
	return source, dir, oracle, env
}

func runHostRuntimeContracts(test *testing.T, options native.Options) {
	test.Helper()
	source, dir, oracle, env := hostFixture(test)
	modes := []string{
		"output",
		"exit",
		"status",
		"clocks",
		"memory",
		"identity",
		"cwd",
		"environment",
		"realpath",
		"realpath-error",
		"realpath-walk-error",
		"eol",
	}
	binary := filepath.Join(dir, fmt.Sprintf("host-%t", options.Sanitize))
	hostBuild(test, source, binary, "", options)
	for _, mode := range modes {
		// Not parallel: modes share filesystem fixtures, including directory mutations.
		test.Run(fmt.Sprintf("sanitized=%t/%s", options.Sanitize, mode), func(test *testing.T) {
			args := []string{mode}
			if mode == "exit" {
				args = append(args, "2")
			}
			if mode == "identity" {
				args = append(args, "--prof", "héllo 🌍", "")
			}
			truth := hostRun(test, env, "node", append([]string{oracle}, args...)...)
			got := hostRun(test, env, binary, args...)
			if got.ExitCode != truth.ExitCode ||
				!bytes.Equal(got.Stdout, truth.Stdout) ||
				!bytes.Equal(got.Stderr, truth.Stderr) {
				test.Fatalf("native %+v; Node %+v", got, truth)
			}
			if options.Sanitize {
				report, checkError := leakcheck.Check(leakcheck.Program{
					C:         source,
					Sanitized: binary,
					Counted:   filepath.Join(test.TempDir(), "counted"),
					Arguments: func() []string { return args },
					Execute: func(extra []string, name string, args ...string) leakcheck.Run {
						return hostRun(test, append(env, extra...), name, args...)
					},
				})
				if checkError != nil || report != "" {
					test.Fatalf("leak check: %v %s", checkError, report)
				}
			}
		})
	}
	for _, code := range []string{"0", "1", "2", "-1", "258"} {
		truth := hostRun(test, env, "node", oracle, "exit", code)
		got := hostRun(test, env, binary, "exit", code)
		if got.ExitCode != truth.ExitCode ||
			!bytes.Equal(got.Stdout, truth.Stdout) ||
			!bytes.Equal(got.Stderr, truth.Stderr) {
			test.Fatalf("exit(%s): native %+v Node %+v", code, got, truth)
		}
	}
	for _, mode := range []string{"output", "exit"} {
		args := []string{mode}
		if mode == "exit" {
			args = append(args, "2")
		}
		mergedEnvironment := append(append([]string{}, env...), "HOST_MERGE=1")
		truth := hostRun(test, mergedEnvironment, "node", append([]string{oracle}, args...)...)
		got := hostRun(test, mergedEnvironment, binary, args...)
		if got.ExitCode != truth.ExitCode || !bytes.Equal(got.Stdout, truth.Stdout) || len(got.Stderr) != 0 {
			test.Fatalf("merged streams %s: native %+v Node %+v", mode, got, truth)
		}
	}
	naturalEnvironment := append(append([]string{}, env...), "ASAN_OPTIONS=detect_leaks=1")
	natural := hostRun(test, naturalEnvironment, binary, "natural")
	truth := hostRun(test, env, "node", oracle, "natural")
	// A normal nonzero completion still runs LSan; compare its expected Node status directly.
	// This does not broaden the shared helper's intentional-teardown exception.
	if natural.ExitCode != truth.ExitCode ||
		!bytes.Equal(natural.Stdout, truth.Stdout) ||
		!bytes.Equal(natural.Stderr, truth.Stderr) {
		test.Fatalf("exitCode natural completion: native %+v Node %+v", natural, truth)
	}
}

func TestHostRuntimeContractsSanitized(test *testing.T) {
	test.Parallel()
	runHostRuntimeContracts(test, native.Options{Sanitize: true})
}

func TestHostRuntimeContractsRelease(test *testing.T) {
	test.Parallel()
	runHostRuntimeContracts(test, native.Options{})
}

func TestHostRuntimeIntentionalTeardown(test *testing.T) {
	test.Parallel()
	source, dir, _, env := hostFixture(test)
	counted := filepath.Join(dir, "counted-host")
	hostBuild(test, source, counted, "", native.Options{Sanitize: true, Count: true})
	for _, code := range []string{"0", "1", "2"} {
		run := hostRun(test, env, counted, "exit", code)
		if !bytes.Contains(run.Stderr, []byte("adamic: intentional exit:")) ||
			!bytes.Contains(run.Stderr, []byte("frees 0")) ||
			!bytes.Contains(run.Stderr, []byte("releases 0")) {
			test.Fatalf("teardown must report unreleased allocations: %+v", run)
		}
		if report := leakcheck.Unbalanced(run); report != "" {
			test.Fatal(report)
		}
		unmarked := run
		unmarked.Stderr = leakcheck.IntentionalExitLine.ReplaceAll(run.Stderr, nil)
		if report := leakcheck.Unbalanced(unmarked); report == "" {
			test.Fatal("missing teardown marker mutant escaped leak check")
		}
	}
}

func runHostRuntimeMutant(test *testing.T, index int) {
	test.Helper()
	source, dir, oracle, env := hostFixture(test)
	mutations := []struct{ mode, file, before, after string }{
		{
			"output",
			"adamic.c",
			"write_text(stream, text->bytes, text->length);",
			"write_text(stream, text->bytes, stream == adamic_stdout ? 0 : text->length);",
		},
		{
			"exit",
			"adamic.c",
			"void adamic_process_exit_now(int code) {\n\tadamic_output_flush();",
			"void adamic_process_exit_now(int code) {\n\t/* mutant skips pending output */",
		},
		{"exit", "adamic.c", "\t_exit(code);", "\texit(code);"},
		{
			"status",
			"host_runtime.c",
			"exit_present = present; exit_value = code;",
			"exit_present = present; exit_value = 0;",
		},
		{
			"clocks",
			"host_runtime.c",
			"return ((double)now.tv_sec - (double)monotonic_origin.tv_sec) * 1000.0 +\n\t\t(double)(now.tv_nsec - monotonic_origin.tv_nsec) / 1000000.0;",
			"return 0;",
		},
		{"clocks", "host_runtime.c", "return epoch_origin;", "return 0;"},
		{"clocks", "host_runtime.c", "(double)(now.tv_nsec / 1000000)", "(double)now.tv_nsec / 1000000.0"},
		{
			"memory",
			"host_runtime.c",
			"result->external = result->arrayBuffers = (double)atomic_load_explicit(&backing_bytes, memory_order_relaxed);",
			"result->external = result->arrayBuffers = 0;",
		},
		{"identity", "host_runtime.c", "index = 1; index < saved_count", "index = 2; index < saved_count"},
		{
			"identity",
			"host_runtime.c",
			"adamic_string *adamic_host_exec_path(void) {",
			"adamic_string *adamic_host_exec_path(void) { free(identity_bytes); identity_bytes = NULL; capture_executable_identity();",
		},
		{"cwd", "host_runtime.c", "if (getcwd(buffer, capacity) != NULL)", "if (strcpy(buffer, \"/\"))"},
		{"environment", "host_runtime.c", "const char *value = getenv(bytes);", "const char *value = \"\";"},
		{"eol", "host_runtime.c", "ADAMIC_STRING(\"\\n\")", "ADAMIC_STRING(\"\\r\\n\")"},
	}
	mutation := mutations[index]
	runtime := hostMutant(test, mutation.file, mutation.before, mutation.after)
	binary := filepath.Join(dir, fmt.Sprintf("mutant-%d", index))
	hostBuild(test, source, binary, runtime, native.Options{Sanitize: true})
	args := []string{mutation.mode}
	if mutation.mode == "exit" {
		args = append(args, "2")
	}
	if mutation.mode == "identity" {
		args = append(args, "--prof", "héllo 🌍", "")
	}
	truth := hostRun(test, env, "node", append([]string{oracle}, args...)...)
	got := hostRun(test, env, binary, args...)
	if got.ExitCode != truth.ExitCode || !bytes.Equal(got.Stderr, truth.Stderr) {
		test.Fatalf("mutant failed outside stdout comparison: %+v, Node %+v", got, truth)
	}
	if bytes.Equal(got.Stdout, truth.Stdout) {
		test.Fatal("mutant escaped source-shape oracle")
	}
	test.Logf("runtime mutant caught by Node stdout: mode %s, got %q, want %q", mutation.mode, got.Stdout, truth.Stdout)
}

func TestHostRuntimeMutantOutput(test *testing.T) {
	test.Parallel()
	runHostRuntimeMutant(test, 0)
}

func TestHostRuntimeMutantExitFlush(test *testing.T) {
	test.Parallel()
	runHostRuntimeMutant(test, 1)
}

func TestHostRuntimeMutantExitHooks(test *testing.T) {
	test.Parallel()
	runHostRuntimeMutant(test, 2)
}

func TestHostRuntimeMutantExitStatus(test *testing.T) {
	test.Parallel()
	runHostRuntimeMutant(test, 3)
}

func TestHostRuntimeMutantPerformanceNow(test *testing.T) {
	test.Parallel()
	runHostRuntimeMutant(test, 4)
}

func TestHostRuntimeMutantTimeOrigin(test *testing.T) {
	test.Parallel()
	runHostRuntimeMutant(test, 5)
}

func TestHostRuntimeMutantDateNow(test *testing.T) {
	test.Parallel()
	runHostRuntimeMutant(test, 6)
}

func TestHostRuntimeMutantMemoryUsage(test *testing.T) {
	test.Parallel()
	runHostRuntimeMutant(test, 7)
}

func TestHostRuntimeMutantArgv(test *testing.T) {
	test.Parallel()
	runHostRuntimeMutant(test, 8)
}

func TestHostRuntimeMutantExecutableIdentity(test *testing.T) {
	test.Parallel()
	runHostRuntimeMutant(test, 9)
}

func TestHostRuntimeMutantCwd(test *testing.T) {
	test.Parallel()
	runHostRuntimeMutant(test, 10)
}

func TestHostRuntimeMutantEnvironment(test *testing.T) {
	test.Parallel()
	runHostRuntimeMutant(test, 11)
}

func TestHostRuntimeMutantEol(test *testing.T) {
	test.Parallel()
	runHostRuntimeMutant(test, 12)
}
