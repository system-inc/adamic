package oracle

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// The wasmtime witness: the same artifacts on a second engine. It is opt-in
// (ADAMIC_ORACLE_WASMTIME=1) and every test that runs wasmtime lives here.

func requireWasmtime(t *testing.T) {
	t.Helper()
	if os.Getenv("ADAMIC_ORACLE_WASMTIME") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASMTIME=1")
	}
	if _, err := exec.LookPath("wasmtime"); err != nil {
		t.Fatal(err)
	}
	if err := native.ValidateOptions(native.Options{Target: "wasm32-wasi"}); err != nil {
		t.Fatal(err)
	}
}

// The root mapping deliberately matches oracle/wasi.mjs. Restricting it to the
// working directory would change absolute paths and permission probes.
func onWasmtime(t *testing.T, how inputRun, binary string) run {
	t.Helper()
	arguments := []string{"run", "--dir", "/::/"}
	// WASI libc resolves relative paths against a dot preopen, not the host
	// process cwd. Keep root access and add that alias for input fixtures.
	if how.directory != "" {
		arguments = append(arguments, "--dir", how.directory+"::.",
			"--dir", filepath.Dir(binary)+"::"+filepath.Dir(binary), "--dir", "/dev::/dev")
	}
	arguments = append(arguments, binary)
	return executeInput(t, how, nil, "wasmtime", append(arguments, how.arguments...)...)
}

var wasmtimeEngine = wasmEngine{run: onWasmtime, limits: wasmtimeHostLimits}

// Exact named observations, never a blanket permission to disagree.
func wasmtimeHostLimits(t *testing.T, path string, expected run) (run, bool) {
	t.Helper()
	// Limits are pinned for fixtures directly in testdata; a same-named file in a subdirectory, such as
	// regexp_replace/arguments.a, is a different program and answers to Node unchanged.
	name := filepath.Base(path)
	if filepath.Base(filepath.Dir(path)) != "testdata" {
		name = ""
	}
	limited := false
	replace := func(old, changed string) {
		t.Helper()
		if strings.Count(string(expected.stdout), old) != 1 {
			t.Fatalf("%s witness missing exact control %q", name, old)
		}
		expected.stdout = []byte(strings.Replace(string(expected.stdout), old, changed, 1))
	}
	switch name {
	case "walk.a":
		limited = true
		replace("<dir>: 4 names, bad� name closed locked unlisted\n  bad� name: not looked up\n  closed: directory, true\n    cannot read directory <dir>/closed: permission denied\n  locked: directory, true\n    <dir>/locked: 0 names, \n  unlisted: directory, true\n    cannot read directory <dir>/unlisted: permission denied\n", "cannot read directory <dir>: failed\n")
		t.Log("HOST LIMIT walk.a: wasmtime rejects directory entries with non-UTF-8 filenames")
	case "write_stdout_order.a":
		expected = run{stdout: []byte("first\nthird, after cannot write /dev/stdout: permission denied\n")}
		limited = true
	case "write_stderr_order.a":
		expected = run{stdout: []byte("first\nthird\n")}
		limited = true
	case "prompt_then_read.a":
		expected = run{stdout: []byte("ready\ncannot read /dev/stdin: permission denied\n")}
		limited = true
	case "arguments.a":
		expected = run{exitCode: 1, stderr: []byte("Error: failed to convert \"a\\xFFb\" to utf-8\n")}
		limited = true
	}
	if limited && name != "walk.a" {
		t.Log("HOST LIMIT", name, "exact pinned wasmtime stdout/stderr/exit asserted")
	}
	return expected, limited
}

// On a disagreement, run the same artifact under V8 too. This is diagnostic
// evidence only: the source or checked JavaScript backend remains the witness.
func compareWasmtime(t *testing.T, expected, actual run, how inputRun, binary string) {
	t.Helper()
	if difference := disagreement(expected, actual); difference != "" {
		v8 := onV8Input(t, how, binary)
		t.Errorf("%s\nexpected witness/known behavior: exit %d, stdout %q, stderr %q\nwasmtime: exit %d, stdout %q, stderr %q\nv8 wasm: exit %d, stdout %q, stderr %q", difference, expected.exitCode, expected.stdout, expected.stderr, actual.exitCode, actual.stdout, actual.stderr, v8.exitCode, v8.stdout, v8.stderr)
	}
}

// Not parallel: retain deterministic per-fixture counts and avoid compiling the
// entire runtime for hundreds of modules simultaneously.
func TestWasmtimeAgreesWithNode(t *testing.T) {
	requireWasmtime(t)
	compiler := filepath.Join(sharedDirectory(t), "adamic")
	command := bounded(t, "go", "build", "-o", compiler, filepath.Join(repository, "cmd", "adamic"))
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("compiler: %v\n%s", err, output)
	}
	passed, failed, skipped, limitations := 0, 0, 0, 0
	defer func() {
		t.Logf("fixtures=%d pass=%d limitations=%d fail=%d skip=%d", passed+failed+skipped+limitations, passed, limitations, failed, skipped)
	}()
	for _, fixture := range fixtures {
		t.Run(fixture.path, func(t *testing.T) {
			limited := false
			// Fatal build/run errors must count as failures too.
			defer func() {
				if t.Skipped() {
					skipped++
					t.Log("SKIP", fixture.path)
				} else if t.Failed() {
					failed++
					t.Log("FAIL", fixture.path)
				} else if limited {
					limitations++
					t.Log("LIMIT", fixture.path)
				} else {
					passed++
					t.Log("PASS", fixture.path)
				}
			}()
			path, err := filepath.Abs(filepath.Join(repository, fixture.path))
			if err != nil {
				t.Fatal(err)
			}
			if !fixture.lowers {
				_, err := lowered(t, path)
				var notYet *lower.NotYet
				if !errors.As(err, &notYet) {
					t.Fatalf("want NotYet, got %v", err)
				}
				t.Skip("fixture does not lower")
			}
			expected := onNodeWith(t, inputRun{}, path)
			if fixture.checked {
				program, err := lowered(t, path)
				if err != nil {
					t.Fatal(err)
				}
				expected = onJavaScriptBackend(t, program)
			}
			binary := buildEngineFixture(t, compiler, path, sharedDirectory(t))
			actual := onWasmtime(t, inputRun{}, binary)
			expected, limited = wasmtimeHostLimits(t, fixture.path, expected)
			compareWasmtime(t, expected, actual, inputRun{}, binary)
		})
	}
	for _, fixture := range inputFixtures {
		t.Run(fixture.path, func(t *testing.T) {
			limited := false
			defer func() {
				if t.Failed() {
					failed++
					t.Log("FAIL", fixture.path)
				} else if t.Skipped() {
					skipped++
					t.Log("SKIP", fixture.path)
				} else if limited {
					limitations++
					t.Log("LIMIT", fixture.path)
				} else {
					passed++
					t.Log("PASS", fixture.path)
				}
			}()
			limited = runWASIInputFixture(t, compiler, fixture.path, fixture.arguments, fixture.unreadable, fixture.writes, wasmtimeEngine)
		})
	}
}

func TestWasmtimeOracleCatchesMutants(t *testing.T) {
	requireWasmtime(t)
	path, err := filepath.Abs(filepath.Join(repository, "dedication", "dedication.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	expected := onNodeWith(t, inputRun{}, path)
	original := native.C(program)
	probe := func(source string) run {
		binary := filepath.Join(sharedDirectory(t), "mutant.wasm")
		if err := native.Build(source, binary, native.Options{Target: "wasm32-wasi"}); err != nil {
			t.Fatal(err)
		}
		return onWasmtime(t, inputRun{}, binary)
	}
	if difference := disagreement(expected, probe(original)); difference != "" {
		t.Fatalf("control: %s", difference)
	}
	program.Strings[0] += "!"
	if difference := disagreement(expected, probe(native.C(program))); difference != "stdout differs" {
		t.Fatalf("stdout mutant: %q", difference)
	}
	t.Log("stdout mutant caught by exact byte comparison")
	// Change the same fixture's emitted main return, preserving all output.
	index := strings.LastIndex(original, "return 0;")
	if index < 0 {
		t.Fatal("main return missing")
	}
	mutant := original[:index] + strings.Replace(original[index:], "return 0;", "return 23;", 1)
	if difference := disagreement(expected, probe(mutant)); difference != "exit codes differ" {
		t.Fatalf("exit mutant: %q", difference)
	}
	t.Log("exit mutant caught by exit status comparison")
	stderrMutant := "#include <stdio.h>\n" + original[:index] + "fputs(\"mutant\\n\", stderr); " + original[index:]
	if difference := disagreement(expected, probe(stderrMutant)); difference != "stderr differs" {
		t.Fatalf("stderr mutant: %q", difference)
	}
	t.Log("stderr mutant caught by exact byte comparison")
}

func TestWasmtimeCheckedWitness(t *testing.T) {
	t.Parallel()
	requireWasmtime(t)
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/writes_past_end.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	raw := onNodeWith(t, inputRun{}, path)
	checked := onJavaScriptBackend(t, program)
	if disagreement(raw, checked) != "exit codes differ" {
		t.Fatal("checked control must differ from raw source exit")
	}
	binary := filepath.Join(sharedDirectory(t), "checked.wasm")
	if err := native.Build(native.C(program), binary, native.Options{Target: "wasm32-wasi"}); err != nil {
		t.Fatal(err)
	}
	if difference := disagreement(checked, onWasmtime(t, inputRun{}, binary)); difference != "" {
		t.Fatal(difference)
	}
	t.Logf("raw exit=%d checked exit=%d; wasmtime agrees with checked backend", raw.exitCode, checked.exitCode)
}

// The named exceptions still compare all three fields. Each independent mutant
// must fail that exact comparison; changed host behavior cannot become a pass.
func TestWASINamedBehaviorCatchesMutants(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"write_stdout_order.a", "write_stderr_order.a", "prompt_then_read.a", "arguments.a"} {
		t.Run(name, func(t *testing.T) {
			expected, limited := wasmtimeHostLimits(t, filepath.Join("internal/oracle/testdata", name), run{})
			if !limited {
				t.Fatal("missing named assertion")
			}
			stdout := expected
			stdout.stdout = append(append([]byte{}, expected.stdout...), '!')
			stderr := expected
			stderr.stderr = append(append([]byte{}, expected.stderr...), '!')
			exit := expected
			exit.exitCode++
			for _, mutant := range []run{stdout, stderr, exit} {
				if disagreement(expected, mutant) == "" {
					t.Fatal("named behavior mutant escaped")
				}
			}
		})
	}
}

// walk.a's pinned limit replaces only the invalid-UTF-8 listing: Node's empty-path lines
// must survive it, and a host that started listing the directory must be noticed.
func TestWasmtimeWalkLimitCatchesMutants(t *testing.T) {
	t.Parallel()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/walk.a"))
	if err != nil {
		t.Fatal(err)
	}
	shared := sharedDirectory(t)
	how := inputRun{directory: filepath.Dir(path), arguments: []string{writable(t, shared, "written")}}
	if os.Geteuid() == 0 {
		how.credential = &syscall.Credential{Uid: 65534, Gid: 65534}
	}
	expected, limited := wasmtimeHostLimits(t, path, onNodeWith(t, how, path))
	if !limited {
		t.Fatal("walk: wasmtime pins its host limit")
	}
	holdsNodeEmptyPathLines(t, expected)
	changed := expected
	changed.stdout = []byte(strings.Replace(string(expected.stdout), "cannot read directory <dir>: failed\n", "host directory now works\n", 1))
	if disagreement(expected, changed) != "stdout differs" {
		t.Fatal("walk host mutant escaped")
	}
	t.Log("walk empty-path and invalid-UTF-8 host assertion mutants caught")
}
