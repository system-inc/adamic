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

func hostBuild(t *testing.T, source, output, directory string, options native.Options) {
	t.Helper()
	if directory == "" {
		if err := native.Build(source, output, options); err != nil {
			t.Fatal(err)
		}
		return
	}
	library, err := native.RuntimeLibrary(directory, options)
	if err != nil {
		t.Fatal(err)
	}
	main := filepath.Join(t.TempDir(), "main.c")
	if err := os.WriteFile(main, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	args := append(native.LinkFlags(options), "-I", filepath.Dir(library), "-o", output, main)
	args = append(args, native.RuntimeLinkFlags(library)...)
	args = append(args, "-lm")
	if out, err := exec.Command("clang", args...).CombinedOutput(); err != nil {
		t.Fatalf("clang: %v\n%s", err, out)
	}
}
func hostRun(t *testing.T, env []string, program string, args ...string) leakcheck.Run {
	t.Helper()
	cmd := exec.Command(program, args...)
	cmd.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=0", "UBSAN_OPTIONS=halt_on_error=1", "HOST_EMPTY=", "HOST_VALUE=héllo 🌍")
	cmd.Env = append(cmd.Env, env...)
	// Test that executable identity doesn't trust invocation argv[0].
	if filepath.Base(program) != "node" {
		cmd.Args[0] = "forged-argv-zero"
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	for _, value := range env {
		if value == "HOST_MERGE=1" {
			cmd.Stderr = &stdout
		}
	}
	err := cmd.Run()
	if err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			t.Fatal(err)
		}
	}
	return leakcheck.Run{Stdout: stdout.Bytes(), Stderr: stderr.Bytes(), ExitCode: cmd.ProcessState.ExitCode()}
}
func hostMutant(t *testing.T, file, before, after string) string {
	t.Helper()
	directory := t.TempDir()
	entries, err := os.ReadDir("runtime")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join("runtime", e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if e.Name() == file {
			if strings.Count(string(data), before) != 1 {
				t.Fatalf("mutation %q must hit exactly once", before)
			}
			data = []byte(strings.Replace(string(data), before, after, 1))
		}
		if err := os.WriteFile(filepath.Join(directory, e.Name()), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	return directory
}

func TestHostRuntimeContractsAndMutants(t *testing.T) {
	data, err := os.ReadFile("testdata/host-runtime/harness.c")
	if err != nil {
		t.Fatal(err)
	}
	source := string(data)
	oracleData, err := os.ReadFile("testdata/host-runtime/oracle.mjs")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	oracle := filepath.Join(dir, "oracle.mjs")
	if err := os.WriteFile(oracle, oracleData, 0600); err != nil {
		t.Fatal(err)
	}
	// Actual default-library entry, alongside both executable and Node bundle.
	lib, err := os.ReadFile("../../cohere/TypeScript/tsc/internal/bundled/libs/lib.d.ts")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "lib.d.ts"), lib, 0600); err != nil {
		t.Fatal(err)
	}
	child := filepath.Join(dir, "child")
	if err := os.Mkdir(child, 0700); err != nil {
		t.Fatal(err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	env := []string{"HOST_EXPECT_CWD=" + cwd, "HOST_CHILD=" + child}
	modes := []string{"output", "exit", "status", "clocks", "memory", "identity", "cwd", "environment", "eol"}
	for _, options := range []native.Options{{Sanitize: true}, {Release: true}} {
		binary := filepath.Join(dir, fmt.Sprintf("host-%t", options.Sanitize))
		hostBuild(t, source, binary, "", options)
		for _, mode := range modes {
			t.Run(fmt.Sprintf("sanitized=%t/%s", options.Sanitize, mode), func(t *testing.T) {
				args := []string{mode}
				if mode == "exit" {
					args = append(args, "2")
				}
				if mode == "identity" {
					args = append(args, "--prof", "héllo 🌍", "")
				}
				truth := hostRun(t, env, "node", append([]string{oracle}, args...)...)
				got := hostRun(t, env, binary, args...)
				if got.ExitCode != truth.ExitCode || !bytes.Equal(got.Stdout, truth.Stdout) || !bytes.Equal(got.Stderr, truth.Stderr) {
					t.Fatalf("native %+v; Node %+v", got, truth)
				}
				if options.Sanitize {
					report, err := leakcheck.Check(leakcheck.Program{C: source, Sanitized: binary, Counted: filepath.Join(t.TempDir(), "counted"), Arguments: func() []string { return args }, Execute: func(extra []string, name string, args ...string) leakcheck.Run {
						return hostRun(t, append(env, extra...), name, args...)
					}})
					if err != nil || report != "" {
						t.Fatalf("leak check: %v %s", err, report)
					}
				}
			})
		}
		for _, code := range []string{"0", "1", "2", "-1", "258"} {
			truth := hostRun(t, env, "node", oracle, "exit", code)
			got := hostRun(t, env, binary, "exit", code)
			if got.ExitCode != truth.ExitCode || !bytes.Equal(got.Stdout, truth.Stdout) || !bytes.Equal(got.Stderr, truth.Stderr) {
				t.Fatalf("exit(%s): native %+v Node %+v", code, got, truth)
			}
		}
		for _, mode := range []string{"output", "exit"} {
			args := []string{mode}
			if mode == "exit" {
				args = append(args, "2")
			}
			mergedEnv := append(append([]string{}, env...), "HOST_MERGE=1")
			truth := hostRun(t, mergedEnv, "node", append([]string{oracle}, args...)...)
			got := hostRun(t, mergedEnv, binary, args...)
			if got.ExitCode != truth.ExitCode || !bytes.Equal(got.Stdout, truth.Stdout) || len(got.Stderr) != 0 {
				t.Fatalf("merged streams %s: native %+v Node %+v", mode, got, truth)
			}
		}
		naturalEnv := append(append([]string{}, env...), "ASAN_OPTIONS=detect_leaks=1")
		natural := hostRun(t, naturalEnv, binary, "natural")
		truth := hostRun(t, env, "node", oracle, "natural")
		// A normal nonzero completion still runs LSan; compare its expected Node status directly.
		// This does not broaden the shared helper's intentional-teardown exception.
		if natural.ExitCode != truth.ExitCode || !bytes.Equal(natural.Stdout, truth.Stdout) || !bytes.Equal(natural.Stderr, truth.Stderr) {
			t.Fatalf("exitCode natural completion: native %+v Node %+v", natural, truth)
		}
	}
	counted := filepath.Join(dir, "counted-host")
	hostBuild(t, source, counted, "", native.Options{Sanitize: true, Count: true})
	for _, code := range []string{"0", "1", "2"} {
		run := hostRun(t, env, counted, "exit", code)
		if !bytes.Contains(run.Stderr, []byte("adamic: intentional exit:")) || !bytes.Contains(run.Stderr, []byte("frees 0")) || !bytes.Contains(run.Stderr, []byte("releases 0")) {
			t.Fatalf("teardown must report unreleased allocations: %+v", run)
		}
		if report := leakcheck.Unbalanced(run); report != "" {
			t.Fatal(report)
		}
		unmarked := run
		unmarked.Stderr = leakcheck.IntentionalExitLine.ReplaceAll(run.Stderr, nil)
		if report := leakcheck.Unbalanced(unmarked); report == "" {
			t.Fatal("missing teardown marker mutant escaped leak check")
		}
	}
	mutations := []struct{ mode, file, before, after string }{
		{"output", "adamic.c", "write_text(stream, text->bytes, text->length);", "write_text(stream, text->bytes, stream == adamic_stdout ? 0 : text->length);"},
		{"exit", "adamic.c", "void adamic_process_exit_now(int code) {\n\tadamic_output_flush();", "void adamic_process_exit_now(int code) {\n\t/* mutant skips pending output */"},
		{"exit", "adamic.c", "\t_exit(code);", "\texit(code);"},
		{"status", "host_runtime.c", "exit_present = present; exit_value = code;", "exit_present = present; exit_value = 0;"},
		{"clocks", "host_runtime.c", "return ((double)now.tv_sec - (double)monotonic_origin.tv_sec) * 1000.0 +\n\t\t(double)(now.tv_nsec - monotonic_origin.tv_nsec) / 1000000.0;", "return 0;"},
		{"clocks", "host_runtime.c", "return epoch_origin;", "return 0;"},
		{"clocks", "host_runtime.c", "(double)(now.tv_nsec / 1000000)", "(double)now.tv_nsec / 1000000.0"},
		{"memory", "host_runtime.c", "result->external = result->arrayBuffers = (double)atomic_load_explicit(&backing_bytes, memory_order_relaxed);", "result->external = result->arrayBuffers = 0;"},
		{"identity", "host_runtime.c", "index = 1; index < saved_count", "index = 2; index < saved_count"},
		{"identity", "host_runtime.c", "adamic_string *adamic_host_exec_path(void) {", "adamic_string *adamic_host_exec_path(void) { free(identity_bytes); identity_bytes = NULL; capture_executable_identity();"},
		{"cwd", "host_runtime.c", "if (getcwd(buffer, capacity) != NULL)", "if (strcpy(buffer, \"/\"))"},
		{"environment", "host_runtime.c", "const char *value = getenv(bytes);", "const char *value = \"\";"},
		{"eol", "host_runtime.c", "ADAMIC_STRING(\"\\n\")", "ADAMIC_STRING(\"\\r\\n\")"},
	}
	for index, m := range mutations {
		t.Run(fmt.Sprintf("mutant-%d-%s", index, m.mode), func(t *testing.T) {
			runtime := hostMutant(t, m.file, m.before, m.after)
			binary := filepath.Join(dir, fmt.Sprintf("mutant-%d", index))
			hostBuild(t, source, binary, runtime, native.Options{Sanitize: true})
			args := []string{m.mode}
			if m.mode == "exit" {
				args = append(args, "2")
			}
			if m.mode == "identity" {
				args = append(args, "--prof", "héllo 🌍", "")
			}
			truth := hostRun(t, env, "node", append([]string{oracle}, args...)...)
			got := hostRun(t, env, binary, args...)
			if got.ExitCode != truth.ExitCode || !bytes.Equal(got.Stderr, truth.Stderr) {
				t.Fatalf("mutant failed outside stdout comparison: %+v, Node %+v", got, truth)
			}
			if bytes.Equal(got.Stdout, truth.Stdout) {
				t.Fatal("mutant escaped source-shape oracle")
			}
			t.Log("runtime mutant caught by Node stdout")
		})
	}
}
