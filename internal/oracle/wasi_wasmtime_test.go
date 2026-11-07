package oracle

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// The wasmtime witness: the same artifacts on a second engine. It is opt-in
// (ADAMIC_ORACLE_WASMTIME=1) and every test that runs wasmtime lives here.

func requireWasmtime(test *testing.T) {
	test.Helper()
	if os.Getenv("ADAMIC_ORACLE_WASMTIME") != "1" {
		test.Skip("set ADAMIC_ORACLE_WASMTIME=1")
	}
	if _, lookPathError := exec.LookPath("wasmtime"); lookPathError != nil {
		test.Fatal(lookPathError)
	}
	if validateOptionsError := native.ValidateOptions(
		native.Options{Target: "wasm32-wasi"},
	); validateOptionsError != nil {
		test.Fatal(validateOptionsError)
	}
}

// The root mapping deliberately matches oracle/wasi.mjs. Restricting it to the
// working directory would change absolute paths and permission probes.
func onWasmtime(test *testing.T, how inputRun, binary string) run {
	test.Helper()
	// As onV8Input does: the working directory reaches the module as ADAMIC_WASI_CWD, which the
	// cwd-dependent runtime adapters enter before resolving a relative path (node_cwd_wasi.c), so
	// one root preopen serves both. A dot preopen would shadow absolute paths.
	directory := how.directory
	if directory == "" {
		var workingDirectoryError error
		if directory, workingDirectoryError = os.Getwd(); workingDirectoryError != nil {
			test.Fatal(workingDirectoryError)
		}
	}
	arguments := []string{"run", "--dir", "/::/", "--env", "ADAMIC_WASI_CWD=" + directory, binary}
	return executeInput(test, how, nil, "wasmtime", append(arguments, how.arguments...)...)
}

var wasmtimeEngine = wasmEngine{run: onWasmtime, limits: wasmtimeHostLimits}

// Exact named observations, never a blanket permission to disagree.
func wasmtimeHostLimits(test *testing.T, path string, expected run) (run, bool) {
	test.Helper()
	// Limits are pinned for fixtures directly in testdata; a same-named file in a subdirectory, such as
	// regexp_replace/arguments.a, is a different program and answers to Node unchanged.
	name := filepath.Base(path)
	if filepath.Base(filepath.Dir(path)) != "testdata" {
		name = ""
	}
	if strings.HasSuffix(filepath.ToSlash(path), "internal/oracle/testdata/stdout_edges/read_after_print.a") {
		name = "stdout_edges/read_after_print.a"
	}
	limited := false
	replace := func(old, changed string) {
		test.Helper()
		if strings.Count(string(expected.stdout), old) != 1 {
			test.Fatalf("%s witness missing exact control %q", name, old)
		}
		expected.stdout = []byte(strings.Replace(string(expected.stdout), old, changed, 1))
	}
	switch name {
	case "walk.a":
		limited = true
		replace(
			"<dir>: 4 names, bad� name closed locked unlisted\n  bad� name: not looked up\n  closed: directory, true\n    cannot read directory <dir>/closed: permission denied\n  locked: directory, true\n    <dir>/locked: 0 names, \n  unlisted: directory, true\n    cannot read directory <dir>/unlisted: permission denied\n",
			"cannot read directory <dir>: failed\n",
		)
		test.Log("HOST LIMIT walk.a: wasmtime rejects directory entries with non-UTF-8 filenames")
	case "write_stdout_order.a":
		expected = run{stdout: []byte("first\nthird, after cannot write /dev/stdout: permission denied\n")}
		limited = true
	case "write_stderr_order.a":
		expected = run{stdout: []byte("first\nthird\n")}
		limited = true
	case "prompt_then_read.a":
		expected = run{stdout: []byte("ready\ncannot read /dev/stdin: permission denied\n")}
		limited = true
	case "stdout_edges/read_after_print.a":
		// Reads stdin after printing to both streams, so it meets prompt_then_read.a's limit; the stderr after the
		// read still arrives.
		expected = run{
			stdout: []byte("out\ncannot read /dev/stdin: permission denied\n"),
			stderr: []byte("err\ndone\n"),
		}
		limited = true
	case "arguments.a":
		expected = run{exitCode: 1, stderr: []byte("Error: failed to convert \"a\\xFFb\" to utf-8\n")}
		limited = true
	}
	if limited && name != "walk.a" {
		test.Log("HOST LIMIT", name, "exact pinned wasmtime stdout/stderr/exit asserted")
	}
	return expected, limited
}

// On a disagreement, run the same artifact under V8 too. This is diagnostic
// evidence only: the source or checked JavaScript backend remains the witness.
func compareWasmtime(test *testing.T, expected, actual run, how inputRun, binary string) {
	test.Helper()
	if difference := disagreement(expected, actual); difference != "" {
		v8 := onV8Input(test, how, binary)
		test.Errorf(
			"wasmtime vs expected witness: %s\nexpected witness/known behavior: exit %d, stdout %q, stderr %q\nwasmtime: exit %d, stdout %q, stderr %q\nv8 wasm: exit %d, stdout %q, stderr %q",
			difference,
			expected.exitCode,
			expected.stdout,
			expected.stderr,
			actual.exitCode,
			actual.stdout,
			actual.stderr,
			v8.exitCode,
			v8.stdout,
			v8.stderr,
		)
	}
}

// TestWasmtimeAgreesWithNode runs every ordinary fixture's wasm32 build on wasmtime, as a family (families_test.go):
// TestWasmtimeAgreesWithNode_000 and on each run the fixtures their shard owns. Each is built in process, as
// TestWASIAgreesWithNode's are, since building cmd/adamic and then running adamic build per fixture, every fixture in
// one test, was past Loom's 60 s budget uncached. Within a shard the fixtures run in parallel, and each shard logs its
// own counts once they have all finished.
func runWasmtimeAgreesShard(test *testing.T, shard int) {
	test.Helper()
	requireWasmtime(test)
	rows := familyCases(test, "TestWasmtimeAgreesWithNode", shard, fixturePaths())
	if len(rows) == 0 {
		return
	}
	planted, probing := familyProbe("TestWasmtimeAgreesWithNode")
	var passed, failed, skipped, limitations, refused atomic.Int32
	// Cleanups run after the parallel subtests, so the counts are the shard's whole.
	test.Cleanup(func() {
		test.Logf(
			"wasmtime shard=%d fixtures=%d pass=%d limitations=%d refused=%d fail=%d skip=%d",
			shard,
			passed.Load()+failed.Load()+skipped.Load()+limitations.Load()+refused.Load(),
			passed.Load(),
			limitations.Load(),
			refused.Load(),
			failed.Load(),
			skipped.Load(),
		)
	})
	for _, index := range rows {
		fixture := fixtures[index]
		test.Run(fixture.path, func(test *testing.T) {
			test.Parallel()
			limited, boundary := false, false
			// Fatal build/run errors must count as failures too.
			defer func() {
				if test.Skipped() {
					skipped.Add(1)
					test.Log("SKIP", fixture.path)
				} else if test.Failed() {
					failed.Add(1)
					test.Log("FAIL", fixture.path)
				} else if boundary {
					refused.Add(1)
					test.Log("REFUSED", fixture.path)
				} else if limited {
					limitations.Add(1)
					test.Log("LIMIT", fixture.path)
				} else {
					passed.Add(1)
					test.Log("PASS", fixture.path)
				}
			}()
			path, absolutePathError := filepath.Abs(filepath.Join(repository, fixture.path))
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			program, loweredError := lowered(test, path)
			// Held to the refusal TestWASIAgreesWithNode holds them to (runtime_boundary_test.go).
			if runtimeBoundaryViewFixture(fixture.path) {
				assertRuntimeBoundaryViewRefusal(test, loweredError)
				boundary = true
				return
			}
			// The async witnesses' named refusals, as TestWASIAgreesWithNode holds them (async_witnesses_test.go).
			if assertAsyncWitnessRefusal(test, fixture.path, path, loweredError) {
				boundary = true
				return
			}
			if !fixture.lowers {
				var notYet *lower.NotYet
				if !errors.As(loweredError, &notYet) {
					test.Fatalf("want NotYet, got %v", loweredError)
				}
				test.Skip("fixture does not lower")
			}
			if loweredError != nil {
				test.Fatal(loweredError)
			}
			// A fixture that reads process.platform stops at the named panic here as under node:wasi
			// (wasiHostPlatformFixtures, asserted both ways for every fixture).
			platform, readsPlatform := wasiHostPlatform(test, fixture.path, program)
			// The witness comes from the program as lowered, before the probe's mutant can reach it.
			var expected run
			if readsPlatform {
				expected = platform
			} else if outcome, ruled := ruledNativeOutcome(fixture.path); ruled {
				// A ruled backend divergence's wasm32 build is native, so its ruled native outcome is its witness.
				expected = outcome
			} else if fixture.checked {
				expected = onJavaScriptBackend(test, program)
			} else {
				// onNode, as TestWASIAgreesWithNode's witness: it hands json fixtures the prebuilt json_types, where
				// a bare run of oracle/node.mjs builds it with go run, which Loom's runners refuse.
				expected = onNode(test, path)
			}
			if probing && fixture.path == planted {
				// TestFamiliesCatchAPlantedCase's mutant: one byte more in the program wasmtime runs.
				program.Strings[0] += "!"
			}
			binary := filepath.Join(test.TempDir(), "program.wasm")
			// The wasm32 refusals TestWASIAgreesWithNode holds by name (wasiRefusedFixtures), at build or at run.
			reason := buildWASI(test, native.C(program), binary)
			var actual run
			if reason == "" {
				actual = onWasmtime(test, inputRun{}, binary)
				reason = wasiRefusalReason(actual)
			}
			if assertWASIRefusal(test, fixture.path, reason, wasiRefusedFixtures) {
				boundary = true
				return
			}
			expected, limited = wasmtimeHostLimits(test, fixture.path, expected)
			compareWasmtime(test, expected, actual, inputRun{}, binary)
		})
	}
}

// TestWasmtimeInputAgreesWithNode runs every input fixture's wasm32 build on wasmtime, as a family (families_test.go),
// as TestWASIInputAgreesWithNode runs them on V8: within a shard the fixtures run in order, and each shard logs its
// own counts.
func runWasmtimeInputShard(test *testing.T, shard int) {
	test.Helper()
	requireWasmtime(test)
	rows := familyCases(test, "TestWasmtimeInputAgreesWithNode", shard, inputFixturePaths())
	if len(rows) == 0 {
		return
	}
	planted, probing := familyProbe("TestWasmtimeInputAgreesWithNode")
	pass, limits, refused, fail, skip := 0, 0, 0, 0, 0
	defer func() {
		test.Logf(
			"wasmtime input shard=%d fixtures=%d pass=%d limitations=%d refused=%d fail=%d skip=%d",
			shard,
			pass+limits+refused+fail+skip,
			pass,
			limits,
			refused,
			fail,
			skip,
		)
	}()
	for _, index := range rows {
		fixture := inputFixtures[index]
		engine := wasmtimeEngine
		if probing && fixture.path == planted {
			// TestFamiliesCatchAPlantedCase's mutant: one byte more on what wasmtime printed.
			engine.run = func(test *testing.T, how inputRun, binary string) run {
				actual := onWasmtime(test, how, binary)
				actual.stdout = append(actual.stdout, '!')
				return actual
			}
		}
		test.Run(fixture.path, func(test *testing.T) {
			limited, wasRefused := false, false
			defer func() {
				if test.Failed() {
					fail++
					test.Log("FAIL", fixture.path)
				} else if test.Skipped() {
					skip++
					test.Log("SKIP", fixture.path)
				} else if wasRefused {
					refused++
				} else if limited {
					limits++
					test.Log("LIMIT", fixture.path)
				} else {
					pass++
					test.Log("PASS", fixture.path)
				}
			}()
			limited, wasRefused = runWASIInputFixture(
				test,
				fixture.path,
				fixture.arguments,
				fixture.unreadable,
				fixture.writes,
				fixture.plainNames,
				engine,
			)
		})
	}
}

func TestWasmtimeOracleCatchesMutants(test *testing.T) {
	test.Parallel()
	requireWasmtime(test)
	path, absolutePathError := filepath.Abs(filepath.Join(repository, "dedication", "dedication.a"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	expected := onNodeWith(test, inputRun{}, path)
	original := native.C(program)
	probe := func(source string) run {
		binary := filepath.Join(sharedDirectory(test), "mutant.wasm")
		if buildError := native.Build(source, binary, native.Options{Target: "wasm32-wasi"}); buildError != nil {
			test.Fatal(buildError)
		}
		return onWasmtime(test, inputRun{}, binary)
	}
	if difference := disagreement(expected, probe(original)); difference != "" {
		test.Fatalf("control: %s", difference)
	}
	program.Strings[0] += "!"
	if difference := disagreement(expected, probe(native.C(program))); difference != "stdout differs" {
		test.Fatalf("stdout mutant: %q", difference)
	}
	test.Log("stdout mutant caught by exact byte comparison")
	// Change the same fixture's emitted main return, preserving all output: generated main returns the host's
	// exit status after its ordinary cleanup.
	const mainReturn = "return adamic_host_exit_status();"
	index := strings.LastIndex(original, mainReturn)
	if index < 0 {
		test.Fatal("main return missing")
	}
	mutant := original[:index] + strings.Replace(original[index:], mainReturn, "return 23;", 1)
	if difference := disagreement(expected, probe(mutant)); difference != "exit codes differ" {
		test.Fatalf("exit mutant: %q", difference)
	}
	test.Log("exit mutant caught by exit status comparison")
	stderrMutant := "#include <stdio.h>\n" + original[:index] + "fputs(\"mutant\\n\", stderr); " + original[index:]
	if difference := disagreement(expected, probe(stderrMutant)); difference != "stderr differs" {
		test.Fatalf("stderr mutant: %q", difference)
	}
	test.Log("stderr mutant caught by exact byte comparison")
}

func TestWasmtimeCheckedWitness(test *testing.T) {
	test.Parallel()
	requireWasmtime(test)
	path, absolutePathError := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/writes_past_end.a"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	raw := onNodeWith(test, inputRun{}, path)
	checked := onJavaScriptBackend(test, program)
	if disagreement(raw, checked) != "exit codes differ" {
		test.Fatal("checked control must differ from raw source exit")
	}
	binary := filepath.Join(sharedDirectory(test), "checked.wasm")
	if buildError := native.Build(native.C(program), binary, native.Options{Target: "wasm32-wasi"}); buildError != nil {
		test.Fatal(buildError)
	}
	if difference := disagreement(checked, onWasmtime(test, inputRun{}, binary)); difference != "" {
		test.Fatal(difference)
	}
	test.Logf("raw exit=%d checked exit=%d; wasmtime agrees with checked backend", raw.exitCode, checked.exitCode)
}

// The named exceptions still compare all three fields. Each independent mutant
// must fail that exact comparison; changed host behavior cannot become a pass.
func TestWASINamedBehaviorCatchesMutants(test *testing.T) {
	test.Parallel()
	for _, name := range []string{
		"write_stdout_order.a",
		"write_stderr_order.a",
		"prompt_then_read.a",
		"stdout_edges/read_after_print.a",
		"arguments.a",
	} {
		test.Run(name, func(test *testing.T) {
			expected, limited := wasmtimeHostLimits(test, filepath.Join("internal/oracle/testdata", name), run{})
			if !limited {
				test.Fatal("missing named assertion")
			}
			stdout := expected
			stdout.stdout = append(append([]byte{}, expected.stdout...), '!')
			stderr := expected
			stderr.stderr = append(append([]byte{}, expected.stderr...), '!')
			exit := expected
			exit.exitCode++
			for _, mutant := range []run{stdout, stderr, exit} {
				if disagreement(expected, mutant) == "" {
					test.Fatal("named behavior mutant escaped")
				}
			}
		})
	}
}

// walk.a's pinned limit replaces only the invalid-UTF-8 listing: Node's empty-path lines
// must survive it, and a host that started listing the directory must be noticed.
func TestWasmtimeWalkLimitCatchesMutants(test *testing.T) {
	test.Parallel()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		test.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	path, absolutePathError := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/walk.a"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	shared := sharedDirectory(test)
	how := inputRun{directory: filepath.Dir(path), arguments: []string{writable(test, shared, "written")}}
	if os.Geteuid() == 0 {
		how.credential = &syscall.Credential{Uid: 65534, Gid: 65534}
	}
	expected, limited := wasmtimeHostLimits(test, path, onNodeWith(test, how, path))
	if !limited {
		test.Fatal("walk: wasmtime pins its host limit")
	}
	holdsNodeEmptyPathLines(test, expected)
	changed := expected
	changed.stdout = []byte(strings.Replace(
		string(expected.stdout),
		"cannot read directory <dir>: failed\n",
		"host directory now works\n",
		1,
	))
	if disagreement(expected, changed) != "stdout differs" {
		test.Fatal("walk host mutant escaped")
	}
	test.Log("walk empty-path and invalid-UTF-8 host assertion mutants caught")
}
