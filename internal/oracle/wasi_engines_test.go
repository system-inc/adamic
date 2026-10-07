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
	return executeInput(t, how, nil, "wasmtime", append([]string{"run", "--dir", "/::/", binary}, how.arguments...)...)
}

// On a disagreement, run the same artifact under V8 too. This is diagnostic
// evidence only: Node running the source still decides whether the test passes.
func compareWasmtime(t *testing.T, expected, actual run, how inputRun, binary string) {
	t.Helper()
	if difference := disagreement(expected, actual); difference != "" {
		runner, err := filepath.Abs(filepath.Join(repository, "oracle", "wasi.mjs"))
		if err != nil {
			t.Fatal(err)
		}
		v8 := executeInput(t, how, nil, "node", append([]string{"--disable-warning=ExperimentalWarning", runner, binary}, how.arguments...)...)
		t.Errorf("%s\nnode source: exit %d, stdout %q, stderr %q\nwasmtime: exit %d, stdout %q, stderr %q\nv8 wasm: exit %d, stdout %q, stderr %q", difference, expected.exitCode, expected.stdout, expected.stderr, actual.exitCode, actual.stdout, actual.stderr, v8.exitCode, v8.stdout, v8.stderr)
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
	passed, failed, skipped := 0, 0, 0
	defer func() { t.Logf("fixtures=%d pass=%d fail=%d skip=%d", passed+failed+skipped, passed, failed, skipped) }()
	for _, fixture := range fixtures {
		t.Run(fixture.path, func(t *testing.T) {
			// Fatal build/run errors must count as failures too.
			defer func() {
				if t.Skipped() {
					skipped++
					t.Log("SKIP", fixture.path)
				} else if t.Failed() {
					failed++
					t.Log("FAIL", fixture.path)
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
			binary := buildEngineFixture(t, compiler, path, sharedDirectory(t))
			actual := onWasmtime(t, inputRun{}, binary)
			compareWasmtime(t, expected, actual, inputRun{}, binary)
		})
	}
	for _, fixture := range inputFixtures {
		t.Run(fixture.path, func(t *testing.T) {
			defer func() {
				if t.Skipped() {
					skipped++
					t.Log("SKIP", fixture.path)
				} else if t.Failed() {
					failed++
					t.Log("FAIL", fixture.path)
				} else {
					passed++
					t.Log("PASS", fixture.path)
				}
			}()
			path, err := filepath.Abs(filepath.Join(repository, fixture.path))
			if err != nil {
				t.Fatal(err)
			}
			shared := sharedDirectory(t)
			binary := buildEngineFixture(t, compiler, path, shared)
			how := inputRun{directory: filepath.Dir(path), arguments: append([]string{}, fixture.arguments...)}
			if os.Geteuid() == 0 {
				how.credential = &syscall.Credential{Uid: 65534, Gid: 65534}
			}
			if fixture.unreadable {
				unreadable := filepath.Join(shared, "unreadable.txt")
				if err := os.WriteFile(unreadable, []byte("secret\n"), 0); err != nil {
					t.Fatal(err)
				}
				how.arguments = append(how.arguments, unreadable)
			}
			// Reuse the exact argument path, resetting its contents between hosts.
			// Separate paths would make fixtures that print paths differ spuriously.
			if fixture.writes {
				how.arguments = append([]string{writable(t, shared, "written")}, how.arguments...)
			}
			expected := onNodeWith(t, how, path)
			var left map[string]string
			if fixture.writes {
				left = snapshot(t, how.arguments[0])
				if err := os.RemoveAll(how.arguments[0]); err != nil {
					t.Fatal(err)
				}
				writable(t, shared, "written")
			}
			actual := onWasmtime(t, how, binary)
			compareWasmtime(t, expected, actual, how, binary)
			if fixture.writes {
				if difference := filesDiffer(left, snapshot(t, how.arguments[0])); difference != "" {
					t.Errorf("files differ: %s", difference)
				}
			}
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
}
