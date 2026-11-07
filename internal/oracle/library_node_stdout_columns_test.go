package oracle

import (
	"encoding/json"
	"errors"
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

func stdoutTTYColumns(test *testing.T, transport string) {
	test.Helper()
	stdoutIsTTYAgreement(test, transport)
	stdoutColumnsAgreement(test, transport)
}

func stdoutColumnsAgreement(test *testing.T, transport string) {
	test.Helper()
	path, binary, js := sanitized(test, "internal/oracle/testdata/stdout_columns/columns.a")
	loader := filepath.Join(repository, "oracle/node.mjs")
	node, pathError := toolchain.Path("node")
	if pathError != nil {
		test.Fatal(pathError)
	}
	variants := [][]string{
		{node, "--disable-warning=ExperimentalWarning", loader, path},
		{binary},
		{node, "--disable-warning=ExperimentalWarning", loader, js},
	}
	var truth run
	for index, variant := range variants {
		got := stdoutIsTTYRun(test, transport, variant[0], variant[1:]...)
		if index == 0 {
			truth = got
			if got.exitCode != 0 {
				test.Fatalf("Node: %+v", got)
			}
			if transport != "pty" && string(got.stderr) != "undefined\n0\n" {
				test.Fatalf("off-terminal Node: %+v", got)
			}
		} else if d := disagreement(truth, got); d != "" {
			test.Fatalf("%s columns backend %d: %s; Node %+v, got %+v", transport, index, d, truth, got)
		}
	}
}

type columnsMutant struct{ name, before, after, red string }

var columnsMutants = []columnsMutant{
	{"premature-turn", "", "", "ResolvedAwaitResize"},
	{
		"every-read",
		"    adamic_maybe_number result = {stdout_width_known, (double)stdout_width};",
		"    struct winsize fresh;\n    if (ioctl(adamic_stdout,TIOCGWINSZ,&fresh)==0) {stdout_width=fresh.ws_col;stdout_width_known=true;}\n    adamic_maybe_number result = {stdout_width_known, (double)stdout_width};",
		"SynchronousResize",
	},
	{"never-refresh", "    if (!stdout_winch_pending) {", "    if (true) {", "AwaitResize"},
	{
		"uncleared-flag",
		"    stdout_winch_pending = 0;",
		"    /* mutant: keep the signal flag */",
		"OneRefreshPerSignal",
	},
	{"clear-broken", "    broken[adamic_stdout] = true;", "    /* mutant: leave stdout clear */", "RefreshFailure"},
	{
		"skip-unknown",
		"    struct winsize size;\n    if (ioctl(adamic_stdout, TIOCGWINSZ, &size) == 0) {\n        stdout_width = size.ws_col;\n        stdout_width_known = true;\n        pthread_mutex_unlock(&output_lock);",
		"    if (!stdout_width_known) {pthread_mutex_unlock(&output_lock);return;}\n    struct winsize size;\n    if (ioctl(adamic_stdout, TIOCGWINSZ, &size) == 0) {\n        stdout_width = size.ws_col;\n        stdout_width_known = true;\n        pthread_mutex_unlock(&output_lock);",
		"InitialFailureRecovery",
	},
}

// Probes and mutants are inputs, prepared once before the manual pty unit. The
// plain destinations use the emitted fixture, not this runtime host-loop probe.
func stdoutColumnsProducts(test *testing.T) string {
	test.Helper()
	inputs := buildcache.Inputs{
		Name: "stdout-columns-probes",
		Files: append(
			compilerinputs.Lowering(),
			"internal/oracle/library_node_stdout_columns_test.go",
			"internal/oracle/testdata/stdout_columns",
		),
		Flags: []string{
			"ADAMIC_PROGRAM_REGION=" + os.Getenv("ADAMIC_PROGRAM_REGION"),
			"ADAMIC_NATIVE_SPLIT=" + os.Getenv("ADAMIC_NATIVE_SPLIT"),
		},
		Toolchain: []string{buildcache.Tool("clang", "--version")},
	}
	return buildcache.Product(test, inputs, func(directory string) error {
		read := func(path string) (string, error) {
			content, readError := os.ReadFile(filepath.Join(repository, path))
			return string(content), readError
		}
		probe, readError := read("internal/oracle/testdata/stdout_columns/loop.c")
		if readError != nil {
			return readError
		}
		if buildError := native.Build(
			probe,
			filepath.Join(directory, "control"),
			native.Options{Sanitize: true},
		); buildError != nil {
			return buildError
		}
		shim, readError := read("internal/oracle/testdata/stdout_columns/ioctl.c")
		if readError != nil {
			return readError
		}
		if buildSharedProbeError := native.BuildSharedProbe(
			shim,
			filepath.Join(directory, "ioctl.so"),
		); buildSharedProbeError != nil {
			return buildSharedProbeError
		}
		source, readError := read("internal/native/runtime/adamic.c")
		if readError != nil {
			return readError
		}
		header, readError := read("internal/native/runtime/stdout_tty_impl.h")
		if readError != nil {
			return readError
		}
		for _, columnsMutant := range columnsMutants {
			if columnsMutant.name == "premature-turn" {
				continue
			}
			if strings.Count(header, columnsMutant.before) != 1 {
				return fmt.Errorf("%s mutant seam changed", columnsMutant.name)
			}
			mutated := strings.Replace(header, columnsMutant.before, columnsMutant.after, 1)
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
			if buildFlagsError := native.BuildFlags(
				prefix+code,
				filepath.Join(directory, columnsMutant.name),
				native.Options{Sanitize: true},
				flags,
			); buildFlagsError != nil {
				return buildFlagsError
			}
		}
		resolved := filepath.Join(repository, "internal/oracle/testdata/stdout_columns/resolved_await.a")
		program, loweredError := lowered(test, resolved)
		if loweredError != nil {
			return loweredError
		}
		code := native.C(program)
		if buildError := native.Build(
			code,
			filepath.Join(directory, "resolved-await"),
			native.Options{Sanitize: true},
		); buildError != nil {
			return buildError
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
		if buildFlagsError := native.BuildFlags(
			code+wrapper,
			filepath.Join(directory, "premature-turn"),
			native.Options{Sanitize: true},
			flags,
		); buildFlagsError != nil {
			return buildFlagsError
		}
		return os.WriteFile(
			filepath.Join(directory, "resolved-await.mjs"),
			[]byte(javascript.JavaScript(program)),
			0600,
		)
	})
}

// Called below the existing census-addressed pty guard; there is no new skip.
func stdoutColumnsLifecycle(test *testing.T) {
	test.Helper()
	if runtime.GOOS != "linux" {
		test.Fatal("manual stdout columns lifecycle probes require Linux")
	}
	products := stdoutColumnsProducts(test)
	node, pathError := toolchain.Path("node")
	if pathError != nil {
		test.Fatal(pathError)
	}
	for _, name := range []string{
		"SynchronousResize",
		"ResolvedAwaitResize",
		"AwaitResize",
		"OneRefreshPerSignal",
		"RefreshFailure",
		"InitialFailureNoSignal",
		"InitialFailureRecovery",
		"InitialFailureRefreshFailure",
	} {
		test.Run("Columns"+name, func(test *testing.T) {
			result := columnsLifecycleRun(test, products, node, name, "")
			if result.Difference != "" {
				test.Fatal(result.Difference)
			}
			// Reuse the house's exact stdout/stderr/status comparison, not a stack
			// normalizer. The protocol already pins Node's lifecycle control.
			if result.NodeCount < 1 || result.NativeCount < 1 {
				test.Fatal("ioctl observation missing")
			}
		})
	}
	for _, columnsMutant := range columnsMutants {
		test.Run("ColumnsMutant_"+columnsMutant.name, func(test *testing.T) {
			result := columnsLifecycleRun(test, products, node, columnsMutant.red, columnsMutant.name)
			if result.Difference == "" {
				test.Fatalf("%s escaped Columns%s", columnsMutant.name, columnsMutant.red)
			}
			test.Logf("caught %s: Columns%s: %s", columnsMutant.name, columnsMutant.red, result.Difference)
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

func columnsLifecycleRun(test *testing.T, products, node, fixture, mutant string) columnsLifecycleResult {
	test.Helper()
	script := filepath.Join(repository, "internal/oracle/testdata/stdout_columns/check.py")
	binary := filepath.Join(products, "control")
	if mutant != "" {
		binary = filepath.Join(products, mutant)
	}
	args := []string{
		script,
		"--node",
		node,
		"--native",
		binary,
		"--preload",
		filepath.Join(products, "ioctl.so"),
		"--fixture",
		fixture,
		"--root",
		repository,
	}
	if fixture == "ResolvedAwaitResize" {
		if mutant == "" {
			binary = filepath.Join(products, "resolved-await")
			args[4] = binary
		}
		args = append(
			args,
			"--source",
			filepath.Join(repository, "internal/oracle/testdata/stdout_columns/resolved_await.a"),
			"--javascript",
			filepath.Join(products, "resolved-await.mjs"),
		)
	}
	_, command := bounded(test, "python3", args...)
	var output strings.Builder
	command.Stdout = &output
	command.Stderr = &output
	if runError := childguard.Run(command, childguard.Options{}); runError != nil {
		test.Fatalf("pty protocol: %v\n%s", runError, output.String())
	}
	var result columnsLifecycleResult
	if decodeError := json.Unmarshal([]byte(output.String()), &result); decodeError != nil {
		test.Fatalf("pty result: %v\n%s", decodeError, output.String())
	}
	result.Difference = disagreement(
		run{stdout: []byte(result.Node.Stdout), stderr: []byte(result.Node.Stderr), exitCode: result.Node.ExitCode},
		run{
			stdout:   []byte(result.Native.Stdout),
			stderr:   []byte(result.Native.Stderr),
			exitCode: result.Native.ExitCode,
		},
	)
	if result.Difference == "" && result.JavaScript != nil {
		result.Difference = disagreement(run{
			stdout:   []byte(result.Node.Stdout),
			stderr:   []byte(result.Node.Stderr),
			exitCode: result.Node.ExitCode,
		}, run{
			stdout:   []byte(result.JavaScript.Stdout),
			stderr:   []byte(result.JavaScript.Stderr),
			exitCode: result.JavaScript.ExitCode,
		})
	}
	if result.Difference == "" && fixture == "OneRefreshPerSignal" && result.NativeCount != 2 {
		result.Difference = "expected first ioctl plus exactly one refresh per signal"
	}
	test.Log(output.String())
	return result
}

func TestStdoutColumnsRefusesWASI(test *testing.T) {
	test.Parallel()
	path := filepath.Join(repository, "internal/oracle/testdata/stdout_columns/columns.a")
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	buildError := native.Build(
		native.C(program),
		filepath.Join(test.TempDir(), "columns.wasm"),
		native.Options{Target: "wasm32-wasi"},
	)
	var refused *native.TargetRefused
	if !errors.As(buildError, &refused) ||
		refused.Member != "process.stdout.columns" ||
		refused.Reason != "wasi preview1 reports no terminal size" {
		test.Fatalf("want typed, named WASI refusal, got %v", buildError)
	}
	if reason := buildWASI(
		test,
		native.C(program),
		filepath.Join(test.TempDir(), "oracle-columns.wasm"),
	); reason != refused.Error() {
		test.Fatalf("oracle lost named WASI refusal: %q", reason)
	}
}

// No type is widened for the crossing: CheckPure rejects NodeFSFile before a
// Wasm Worker build, and standalone WASI emission has the named target refusal.
