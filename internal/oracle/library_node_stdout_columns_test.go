package oracle

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/compilerinputs"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/toolchain"
)

func stdoutTTYColumns(t *testing.T, transport string) {
	t.Helper()
	stdoutIsTTYAgreement(t, transport)
	stdoutColumnsAgreement(t, transport)
}

func stdoutColumnsAgreement(t *testing.T, transport string) {
	t.Helper()
	path, binary, js := sanitized(t, "internal/oracle/testdata/stdout_columns/columns.a")
	loader := filepath.Join(repository, "oracle/node.mjs")
	node, err := toolchain.Path("node")
	if err != nil {
		t.Fatal(err)
	}
	variants := [][]string{{node, "--disable-warning=ExperimentalWarning", loader, path}, {binary}, {node, "--disable-warning=ExperimentalWarning", loader, js}}
	var truth run
	for i, v := range variants {
		got := stdoutIsTTYRun(t, transport, v[0], v[1:]...)
		if i == 0 {
			truth = got
			if got.exitCode != 0 {
				t.Fatalf("Node: %+v", got)
			}
			if transport != "pty" && string(got.stderr) != "undefined\n0\n" {
				t.Fatalf("off-terminal Node: %+v", got)
			}
		} else if d := disagreement(truth, got); d != "" {
			t.Fatalf("%s columns backend %d: %s; Node %+v, got %+v", transport, i, d, truth, got)
		}
	}
}

type columnsMutant struct{ name, before, after, red string }

var columnsMutants = []columnsMutant{
	{"premature-turn", "", "", "ResolvedAwaitResize"},
	{"every-read", "    adamic_maybe_number result = {stdout_width_known, (double)stdout_width};", "    struct winsize fresh;\n    if (ioctl(adamic_stdout,TIOCGWINSZ,&fresh)==0) {stdout_width=fresh.ws_col;stdout_width_known=true;}\n    adamic_maybe_number result = {stdout_width_known, (double)stdout_width};", "SynchronousResize"},
	{"never-refresh", "    if (!stdout_winch_pending) {", "    if (true) {", "AwaitResize"},
	{"uncleared-flag", "    stdout_winch_pending = 0;", "    /* mutant: keep the signal flag */", "OneRefreshPerSignal"},
	{"clear-broken", "    broken[adamic_stdout] = true;", "    /* mutant: leave stdout clear */", "RefreshFailure"},
	{"skip-unknown", "    struct winsize size;\n    if (ioctl(adamic_stdout, TIOCGWINSZ, &size) == 0) {\n        stdout_width = size.ws_col;\n        stdout_width_known = true;\n        pthread_mutex_unlock(&output_lock);", "    if (!stdout_width_known) {pthread_mutex_unlock(&output_lock);return;}\n    struct winsize size;\n    if (ioctl(adamic_stdout, TIOCGWINSZ, &size) == 0) {\n        stdout_width = size.ws_col;\n        stdout_width_known = true;\n        pthread_mutex_unlock(&output_lock);", "InitialFailureRecovery"},
}

// Probes and mutants are inputs, prepared once before the manual pty unit. The
// plain destinations use the emitted fixture, not this runtime host-loop probe.
func stdoutColumnsProducts(t *testing.T) string {
	t.Helper()
	inputs := buildcache.Inputs{Name: "stdout-columns-probes", Files: append(compilerinputs.Lowering(), "internal/oracle/library_node_stdout_columns_test.go", "internal/oracle/testdata/stdout_columns"), Flags: []string{"ADAMIC_PROGRAM_REGION=" + os.Getenv("ADAMIC_PROGRAM_REGION"), "ADAMIC_NATIVE_SPLIT=" + os.Getenv("ADAMIC_NATIVE_SPLIT")}, Toolchain: []string{buildcache.Tool("clang", "--version")}}
	return buildcache.Product(t, inputs, func(directory string) error {
		read := func(path string) (string, error) {
			b, e := os.ReadFile(filepath.Join(repository, path))
			return string(b), e
		}
		probe, err := read("internal/oracle/testdata/stdout_columns/loop.c")
		if err != nil {
			return err
		}
		if err := native.Build(probe, filepath.Join(directory, "control"), native.Options{Sanitize: true}); err != nil {
			return err
		}
		shim, err := read("internal/oracle/testdata/stdout_columns/ioctl.c")
		if err != nil {
			return err
		}
		if err := native.BuildSharedProbe(shim, filepath.Join(directory, "ioctl.so")); err != nil {
			return err
		}
		source, err := read("internal/native/runtime/adamic.c")
		if err != nil {
			return err
		}
		header, err := read("internal/native/runtime/stdout_tty_impl.h")
		if err != nil {
			return err
		}
		for _, m := range columnsMutants {
			if m.name == "premature-turn" {
				continue
			}
			if strings.Count(header, m.before) != 1 {
				return fmt.Errorf("%s mutant seam changed", m.name)
			}
			mutated := strings.Replace(header, m.before, m.after, 1)
			code := strings.Replace(source, `#include "stdout_tty_impl.h"`, mutated, 1) + "\n" + probe
			// Whole-archive linking retains the control runtime too. Rename
			// this one translation unit's public symbols and wrap incoming
			// runtime calls, so async.c reaches this mutant's refresh/cache.
			pattern := regexp.MustCompile(`(?m)^(?:_Noreturn )?[a-z_]+ +(adamic_[a-z_]+)[(][^;{}]*[)] [{]`)
			matches := pattern.FindAllStringSubmatch(source+"\n"+header, -1)
			flags := native.Flags(native.Options{Sanitize: true})
			prefix := ""
			for _, match := range matches {
				prefix += "#define " + match[1] + " __wrap_" + match[1] + "\n"
				flags = append(flags, "-Wl,--wrap="+match[1])
			}
			if len(matches) < 10 {
				return fmt.Errorf("stdout mutant public-symbol inventory changed")
			}
			if err := native.BuildFlags(prefix+code, filepath.Join(directory, m.name), native.Options{Sanitize: true}, flags); err != nil {
				return err
			}
		}
		resolved := filepath.Join(repository, "internal/oracle/testdata/stdout_columns/resolved_await.a")
		program, err := lowered(t, resolved)
		if err != nil {
			return err
		}
		code := native.C(program)
		if err := native.Build(code, filepath.Join(directory, "resolved-await"), native.Options{Sanitize: true}); err != nil {
			return err
		}
		wrapper := `
#include "async.h"
void __real_adamic_async_run(void);
void __wrap_adamic_async_run(void) {
    adamic_host_process_completions();
    __real_adamic_async_run();
}
`
		flags := append(native.SourceFlags(code, native.Options{Sanitize: true}), "-Wl,--wrap=adamic_async_run")
		if err := native.BuildFlags(code+wrapper, filepath.Join(directory, "premature-turn"), native.Options{Sanitize: true}, flags); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(directory, "resolved-await.mjs"), []byte(javascript.JavaScript(program)), 0600)
	})
}

// Called below the existing census-addressed pty guard; there is no new skip.
func stdoutColumnsLifecycle(t *testing.T) {
	t.Helper()
	if runtime.GOOS != "linux" {
		t.Fatal("manual stdout columns lifecycle probes require Linux")
	}
	products := stdoutColumnsProducts(t)
	node, err := toolchain.Path("node")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"SynchronousResize", "ResolvedAwaitResize", "AwaitResize", "OneRefreshPerSignal", "RefreshFailure", "InitialFailureNoSignal", "InitialFailureRecovery", "InitialFailureRefreshFailure"} {
		t.Run("Columns"+name, func(t *testing.T) {
			result := columnsLifecycleRun(t, products, node, name, "")
			if result.Difference != "" {
				t.Fatal(result.Difference)
			}
			// Reuse the house's exact stdout/stderr/status comparison, not a stack
			// normalizer. The protocol already pins Node's lifecycle control.
			if result.NodeCount < 1 || result.NativeCount < 1 {
				t.Fatal("ioctl observation missing")
			}
		})
	}
	for _, m := range columnsMutants {
		t.Run("ColumnsMutant_"+m.name, func(t *testing.T) {
			result := columnsLifecycleRun(t, products, node, m.red, m.name)
			if result.Difference == "" {
				t.Fatalf("%s escaped Columns%s", m.name, m.red)
			}
			t.Logf("caught %s: Columns%s: %s", m.name, m.red, result.Difference)
		})
	}
}

type columnsLifecycleObservation struct {
	Stdout   string
	Stderr   string
	ExitCode int
}
type columnsLifecycleResult struct {
	JavaScript             *columnsLifecycleObservation
	Difference             string
	Node, Native           columnsLifecycleObservation
	NodeCount, NativeCount int
}

func columnsLifecycleRun(t *testing.T, products, node, fixture, mutant string) columnsLifecycleResult {
	t.Helper()
	script := filepath.Join(repository, "internal/oracle/testdata/stdout_columns/check.py")
	binary := filepath.Join(products, "control")
	if mutant != "" {
		binary = filepath.Join(products, mutant)
	}
	args := []string{script, "--node", node, "--native", binary, "--preload", filepath.Join(products, "ioctl.so"), "--fixture", fixture, "--root", repository}
	if fixture == "ResolvedAwaitResize" {
		if mutant == "" {
			binary = filepath.Join(products, "resolved-await")
			args[4] = binary
		}
		args = append(args, "--source", filepath.Join(repository, "internal/oracle/testdata/stdout_columns/resolved_await.a"), "--javascript", filepath.Join(products, "resolved-await.mjs"))
	}
	_, cmd := bounded(t, "python3", args...)
	var output strings.Builder
	cmd.Stdout = &output
	cmd.Stderr = &output
	if err := childguard.Run(cmd, childguard.Options{}); err != nil {
		t.Fatalf("pty protocol: %v\n%s", err, output.String())
	}
	var result columnsLifecycleResult
	if err := json.Unmarshal([]byte(output.String()), &result); err != nil {
		t.Fatalf("pty result: %v\n%s", err, output.String())
	}
	result.Difference = disagreement(
		run{stdout: []byte(result.Node.Stdout), stderr: []byte(result.Node.Stderr), exitCode: result.Node.ExitCode},
		run{stdout: []byte(result.Native.Stdout), stderr: []byte(result.Native.Stderr), exitCode: result.Native.ExitCode},
	)
	if result.Difference == "" && result.JavaScript != nil {
		result.Difference = disagreement(run{stdout: []byte(result.Node.Stdout), stderr: []byte(result.Node.Stderr), exitCode: result.Node.ExitCode}, run{stdout: []byte(result.JavaScript.Stdout), stderr: []byte(result.JavaScript.Stderr), exitCode: result.JavaScript.ExitCode})
	}
	if result.Difference == "" && fixture == "OneRefreshPerSignal" && result.NativeCount != 2 {
		result.Difference = "expected first ioctl plus exactly one refresh per signal"
	}
	t.Log(output.String())
	return result
}

func TestStdoutColumnsRefusesWASI(t *testing.T) {
	t.Parallel()
	path := filepath.Join(repository, "internal/oracle/testdata/stdout_columns/columns.a")
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	err = native.Build(native.C(program), filepath.Join(t.TempDir(), "columns.wasm"), native.Options{Target: "wasm32-wasi"})
	if err == nil || !strings.Contains(err.Error(), "wasi preview1 reports no terminal size") {
		t.Fatalf("want named WASI refusal, got %v", err)
	}
}

// No type is widened for the crossing: CheckPure rejects NodeFSFile before a
// Wasm Worker build, and standalone WASI emission has the named target refusal.
