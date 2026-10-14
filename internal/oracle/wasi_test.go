package oracle

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// TestWASIAgreesWithNode runs every ordinary fixture as a WASI command, as a family (families_test.go):
// TestWASIAgreesWithNode_000 and on each run the fixtures their shard owns.
// Opt in with ADAMIC_ORACLE_WASI=1; these observations always run uncached.
func runWASIAgreesShard(t *testing.T, shard int) {
	t.Helper()
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
	runWASIFixtures(t, familyCases(t, "TestWASIAgreesWithNode", shard, fixturePaths()), mutate)
}

func wasiFixtureDifference(name string, expected, actual run) error {
	if difference := disagreement(expected, actual); difference != "" {
		return fmt.Errorf("%s: %s\nnode: exit %d, stdout %q, stderr %q\nwasi: exit %d, stdout %q, stderr %q", name, difference, expected.exitCode, expected.stdout, expected.stderr, actual.exitCode, actual.stdout, actual.stderr)
	}
	return nil
}

// Setup belongs inside the selected shard. Node identity does not build native
// archives; each WASI program fetches its own feature/options runtime product.
func runWASIFixtures(t *testing.T, rows []int, mutate func(string, *ir.Program)) {
	t.Helper()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1 to run the WASI oracle")
	}
	started := time.Now()
	if err := native.ValidateOptions(native.Options{Target: "wasm32-wasi"}); err != nil {
		t.Fatal(err)
	}
	identity(t)
	t.Logf("WASI shard setup (Node identity): %.3fs", time.Since(started).Seconds())
	for _, fixtureIndex := range rows {
		fixture := fixtures[fixtureIndex]
		t.Run(fixture.path, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, fixture.path))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if assertPrototypeFixtureRefusal(t, fixture.path, err) {
				return
			}
			if !fixture.lowers {
				var notYet *lower.NotYet
				if !errors.As(err, &notYet) {
					t.Fatalf("want NotYet, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			expected := onNode(t, path)
			if fixture.checked {
				expected = onJavaScriptBackend(t, program)
			}
			if mutate != nil {
				mutate(fixture.path, program)
			}
			actual := onWASI(t, native.C(program))
			if err := wasiFixtureDifference(t.Name(), expected, actual); err != nil {
				t.Error(err)
			}
		})
	}
}

func onWASI(t *testing.T, source string) run {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "program.wasm")
	if err := native.Build(source, binary, native.Options{Target: "wasm32-wasi"}); err != nil {
		t.Fatal(err)
	}
	return execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle", "wasi.mjs"), binary)
}

func TestWASIOracleCatchesMutants(t *testing.T) {
	t.Parallel()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	path, err := filepath.Abs(filepath.Join(repository, "dedication", "dedication.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	expected := onNode(t, path)
	program.Strings[0] += "!"
	if difference := disagreement(expected, onWASI(t, native.C(program))); difference != "stdout differs" {
		t.Fatalf("byte mutant: got %q", difference)
	}
	if difference := disagreement(expected, onWASI(t, "int main(void) { return 23; }\n")); difference != "exit codes differ" {
		t.Fatalf("exit mutant: got %q", difference)
	}
}

// This probe holds the runner to real Wasm artifacts even before the runtime port is available.
func TestWASIRunnerCatchesMutants(t *testing.T) {
	t.Parallel()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	sysroot := os.Getenv("WASI_SYSROOT")
	compiler := filepath.Join(filepath.Dir(filepath.Dir(sysroot)), "bin", "clang")
	probe := func(text string, exit int) run {
		directory := t.TempDir()
		source, binary := filepath.Join(directory, "probe.c"), filepath.Join(directory, "probe.wasm")
		if err := os.WriteFile(source, []byte(fmt.Sprintf("#include <stdio.h>\nint main(void) { fputs(%q, stdout); return %d; }\n", text, exit)), 0644); err != nil {
			t.Fatal(err)
		}
		command := bounded(t, compiler, "--target=wasm32-wasi", "--sysroot="+sysroot, "-O2", "-mexec-model=command", source, "-o", binary)
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("probe build: %v\n%s", err, output)
		}
		return execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle", "wasi.mjs"), binary)
	}
	expected := execute(t, "node", "-e", "process.stdout.write('probe\\n')")
	if difference := disagreement(expected, probe("probe\n", 0)); difference != "" {
		t.Fatalf("control: %s", difference)
	}
	if difference := disagreement(expected, probe("probe!\n", 0)); difference != "stdout differs" {
		t.Fatalf("byte mutant: %q", difference)
	}
	if difference := disagreement(expected, probe("probe\n", 23)); difference != "exit codes differ" {
		t.Fatalf("exit mutant: %q", difference)
	}
}

// Compile every emitted translation unit for the 32-bit ABI independently of runtime linking, as a family
// (families_test.go): TestWASIEmission_000 and on each compile the fixtures their shard owns.
func runWASIEmissionShard(t *testing.T, shard int) {
	t.Helper()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	options := native.Options{Target: "wasm32-wasi"}
	if err := native.ValidateOptions(options); err != nil {
		t.Fatal(err)
	}
	compiler := filepath.Join(filepath.Dir(filepath.Dir(os.Getenv("WASI_SYSROOT"))), "bin", "clang")
	_, probing := familyProbe("TestWASIEmission")
	for _, index := range familyCases(t, "TestWASIEmission", shard, fixturePaths()) {
		fixture := fixtures[index]
		t.Run(fixture.path, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, fixture.path))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if !fixture.lowers {
				t.Skip("fixture does not lower")
			}
			if err != nil {
				t.Fatal(err)
			}
			code := native.C(program)
			if probing {
				// TestFamiliesCatchAPlantedCase's mutant: C that no target compiles.
				code += "\nint planted = ;\n"
			}
			directory := t.TempDir()
			source := filepath.Join(directory, "main.c")
			if err := os.WriteFile(source, []byte(code), 0644); err != nil {
				t.Fatal(err)
			}
			flags := append(native.Flags(options), "-I", filepath.Join(repository, "internal", "native", "runtime"), "-c", source, "-o", filepath.Join(directory, "main.o"))
			if output, err := bounded(t, compiler, flags...).CombinedOutput(); err != nil {
				t.Fatalf("emitted C: %v\n%s", err, output)
			}
		})
	}
}
