package oracle

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// TestWASIAgreesWithNode runs every ordinary fixture as a WASI command, as a family (families_test.go):
// TestWASIAgreesWithNode_000 and on each run the fixtures their shard owns.
// Opt in with ADAMIC_ORACLE_WASI=1; these observations always run uncached.
func runWASIAgreesShard(test *testing.T, shard int) {
	test.Helper()
	planted, probing := familyProbe("TestWASIAgreesWithNode")
	var mutate func(string, *ir.Program)
	if probing {
		// TestFamiliesCatchAPlantedCase's mutant: TestTheOracleCatchesOneByte's byte, after lowering.
		mutate = func(path string, program *ir.Program) {
			if path == planted {
				program.Strings[0] += "!"
			}
		}
	}
	runWASIFixtures(test, familyCases(test, "TestWASIAgreesWithNode", shard, fixturePaths()), mutate)
}

func wasiFixtureDifference(name string, expected, actual run) error {
	if difference := disagreement(expected, actual); difference != "" {
		return fmt.Errorf(
			"%s: %s\nnode: exit %d, stdout %q, stderr %q\nwasi: exit %d, stdout %q, stderr %q",
			name,
			difference,
			expected.exitCode,
			expected.stdout,
			expected.stderr,
			actual.exitCode,
			actual.stdout,
			actual.stderr,
		)
	}
	return nil
}

// Setup belongs inside the selected shard, including the Node helper's native
// runtime identity. No result-cache bypass or runtime-cache policy changes here.
func runWASIFixtures(test *testing.T, rows []int, mutate func(string, *ir.Program)) {
	test.Helper()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		test.Skip("set ADAMIC_ORACLE_WASI=1 to run the WASI oracle")
	}
	started := time.Now()
	if err := native.ValidateOptions(native.Options{Target: "wasm32-wasi"}); err != nil {
		test.Fatal(err)
	}
	identity(test)
	test.Logf("WASI shard setup (Node identity, three native runtime archives): %.3fs", time.Since(started).Seconds())
	for _, fixtureIndex := range rows {
		fixture := fixtures[fixtureIndex]
		test.Run(fixture.path, func(test *testing.T) {
			test.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, fixture.path))
			if err != nil {
				test.Fatal(err)
			}
			program, err := lowered(test, path)
			if runtimeBoundaryViewFixture(fixture.path) {
				assertRuntimeBoundaryViewRefusal(test, err)
				return
			}
			if assertAsyncWitnessRefusal(test, fixture.path, path, err) {
				return
			}
			if !fixture.lowers {
				var notYet *lower.NotYet
				if !errors.As(err, &notYet) {
					test.Fatalf("want NotYet, got %v", err)
				}
				return
			}
			if err != nil {
				test.Fatal(err)
			}
			platform, readsPlatform := wasiHostPlatform(test, fixture.path, program)
			expected := onNode(test, path)
			// A ruled backend divergence differs from JavaScript by ruling, and wasm32 is a native build
			// (backend_divergences_test.go), as TestWASIShimAgreesWithNode's witness has it.
			if outcome, ruled := ruledNativeOutcome(fixture.path); ruled {
				expected = outcome
			} else if fixture.checked {
				expected = onJavaScriptBackend(test, program)
			}
			if readsPlatform {
				expected = platform
			}
			if mutate != nil {
				mutate(fixture.path, program)
			}
			actual, reason := observeWASI(test, native.C(program))
			if assertWASIRefusal(test, fixture.path, reason, wasiRefusedFixtures) {
				return
			}
			if fixture.path == "internal/oracle/testdata/host_stdout/closed.a" {
				stdoutWASIClosedPipe(test, native.C(program))
			}
			if err := wasiFixtureDifference(test.Name(), expected, actual); err != nil {
				test.Error(err)
			}
		})
	}
}

func observeWASI(test *testing.T, source string) (run, string) {
	test.Helper()
	binary := filepath.Join(test.TempDir(), "program.wasm")
	if reason := buildWASI(test, source, binary); reason != "" {
		return run{}, reason
	}
	actual := execute(
		test,
		"node",
		"--disable-warning=ExperimentalWarning",
		filepath.Join(repository, "oracle", "wasi.mjs"),
		binary,
	)
	return actual, wasiRefusalReason(actual)
}

// Callers without a named refusal must execute; no host regression becomes a skip.
func onWASI(test *testing.T, source string) run {
	test.Helper()
	actual, reason := observeWASI(test, source)
	if reason != "" {
		test.Fatalf("unlisted wasm32 refusal: %s", reason)
	}
	return actual
}

func wasiOracleCatchesMutant(test *testing.T, exit bool) {
	test.Helper()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		test.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	path, err := filepath.Abs(filepath.Join(repository, "dedication", "dedication.a"))
	if err != nil {
		test.Fatal(err)
	}
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	// A WASI oracle mutant needs a fresh Node witness, not native runtime cache identity.
	expected := onNodeWith(test, inputRun{}, path)
	if exit {
		if difference := disagreement(
			expected,
			onWASI(test, "int main(void) { return 23; }\n"),
		); difference != "exit codes differ" {
			test.Fatalf("exit mutant: got %q", difference)
		}
	} else {
		program.Strings[0] += "!"
		if difference := disagreement(expected, onWASI(test, native.C(program))); difference != "stdout differs" {
			test.Fatalf("byte mutant: got %q", difference)
		}
	}
}

func TestWASIOracleCatchesMutants_000(test *testing.T) {
	test.Parallel()
	wasiOracleCatchesMutant(test, false)
}
func TestWASIOracleCatchesMutants_001(test *testing.T) {
	test.Parallel()
	wasiOracleCatchesMutant(test, true)
}

// This probe holds the runner to real Wasm artifacts even before the runtime port is available.
func TestWASIRunnerCatchesMutants(test *testing.T) {
	test.Parallel()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		test.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	sysroot := os.Getenv("WASI_SYSROOT")
	compiler := filepath.Join(filepath.Dir(filepath.Dir(sysroot)), "bin", "clang")
	probe := func(text string, exit int) run {
		directory := test.TempDir()
		source, binary := filepath.Join(directory, "probe.c"), filepath.Join(directory, "probe.wasm")
		if err := os.WriteFile(
			source,
			[]byte(fmt.Sprintf("#include <stdio.h>\nint main(void) { fputs(%q, stdout); return %d; }\n", text, exit)),
			0644,
		); err != nil {
			test.Fatal(err)
		}
		ctx, command := bounded(
			test,
			compiler,
			"--target=wasm32-wasi",
			"--sysroot="+sysroot,
			"-O2",
			"-mexec-model=command",
			source,
			"-o",
			binary,
		)
		if stdout, stderr, err := childguard.Capture(ctx, command); err != nil {
			test.Fatalf("probe build: %v\n%s%s", err, stdout, stderr)
		}
		return execute(
			test,
			"node",
			"--disable-warning=ExperimentalWarning",
			filepath.Join(repository, "oracle", "wasi.mjs"),
			binary,
		)
	}
	expected := execute(test, "node", "-e", "process.stdout.write('probe\\n')")
	if difference := disagreement(expected, probe("probe\n", 0)); difference != "" {
		test.Fatalf("control: %s", difference)
	}
	if difference := disagreement(expected, probe("probe!\n", 0)); difference != "stdout differs" {
		test.Fatalf("byte mutant: %q", difference)
	}
	if difference := disagreement(expected, probe("probe\n", 23)); difference != "exit codes differ" {
		test.Fatalf("exit mutant: %q", difference)
	}
}
