package oracle

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

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

// On a disagreement, run the same artifact under V8 too. This is diagnostic
// evidence only: the source or checked JavaScript backend remains the witness.
func compareWasmtime(t *testing.T, expected, actual run, how inputRun, binary string) {
	t.Helper()
	if difference := disagreement(expected, actual); difference != "" {
		v8 := onV8Input(t, how, binary)
		t.Errorf("%s\nexpected witness/known behavior: exit %d, stdout %q, stderr %q\nwasmtime: exit %d, stdout %q, stderr %q\nv8 wasm: exit %d, stdout %q, stderr %q", difference, expected.exitCode, expected.stdout, expected.stderr, actual.exitCode, actual.stdout, actual.stderr, v8.exitCode, v8.stdout, v8.stderr)
	}
}

func buildEngineFixture(t *testing.T, compiler, path, directory string) string {
	t.Helper()
	binary := filepath.Join(directory, "program.wasm")
	command := bounded(t, compiler, "build", "--target", "wasm32-wasi", path, "-o", binary)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("adamic build: %v\n%s", err, output)
	}
	return binary
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
			expected, limited = expectedEngineBehavior(t, fixture.path, expected, inputRun{}, true)
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
			limited = runWASIInputFixture(t, compiler, fixture.path, fixture.arguments, fixture.unreadable, fixture.writes, true)
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
