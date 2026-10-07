package oracle

import (
	"errors"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestArrayHolesMilestone(t *testing.T) {
	for _, name := range []string{"library_array_holes_scanner_probe.a", "array-holes-boundaries/library_array_holes_length.a", "array-holes-boundaries/library_array_holes_range.a", "library_array_holes_references.a", "library_array_holes_callbacks.a", "library_array_holes_resize.a", "library_array_holes_toString.a", "library_array_holes_keys.a", "library_array_holes_catch.a", "library_array_holes_objects.a"} {
		t.Run(name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", name))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			native, binary := natively(t, program)
			for _, got := range []run{native, released(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(truth, got); difference != "" {
					t.Fatalf("%s: Node %q %q; got %q %q", difference, truth.stdout, truth.stderr, got.stdout, got.stderr)
				}
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatal(report)
			}
			t.Logf("Node stdout: %s", truth.stdout)
		})
	}
}

func init() {
	for _, name := range []string{"library_array_holes_scanner_probe.a", "array-holes-boundaries/library_array_holes_length.a", "array-holes-boundaries/library_array_holes_range.a", "library_array_holes_references.a", "library_array_holes_callbacks.a", "library_array_holes_resize.a", "library_array_holes_toString.a", "library_array_holes_keys.a", "library_array_holes_catch.a", "library_array_holes_objects.a"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/" + name, true, false})
	}
}

func TestArrayHolesRefusals(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join(repository, "internal/oracle/testdata/array-holes-refused/library_array_holes_refuse_*.a"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 31 {
		t.Fatalf("want all 31 refusal fixtures, got %d", len(paths))
	}
	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			absolute, err := filepath.Abs(path)
			if err != nil {
				t.Fatal(err)
			}
			var truth run
			if filepath.Base(path) == "library_array_holes_refuse_console.a" {
				truth = execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), absolute)
			} else {
				truth = onNode(t, absolute)
			}
			if truth.exitCode != 0 || len(truth.stderr) != 0 {
				t.Fatalf("Node source did not finish: %d %s", truth.exitCode, truth.stderr)
			}
			if filepath.Base(path) == "library_array_holes_refuse_console.a" {
				_, err := load.Load([]string{absolute})
				var check *load.CheckError
				if !errors.As(err, &check) {
					t.Fatalf("want the existing typed console refusal, got %v", err)
				}
				t.Logf("%v; Node stdout %q", err, truth.stdout)
				return
			}
			_, err = lowered(t, absolute)
			var notYet *lower.NotYet
			var refused *lower.Refused
			if !errors.As(err, &notYet) && !errors.As(err, &refused) {
				t.Fatalf("want a named compile refusal, got %v", err)
			}
			if err == nil || (!strings.Contains(err.Error(), "array") && !strings.Contains(err.Error(), "Array") && !strings.Contains(err.Error(), "in") && !strings.Contains(err.Error(), "hasOwnProperty") && !strings.Contains(err.Error(), "console")) {
				t.Fatalf("refusal has no operation reason: %v", err)
			}
			t.Logf("%v; Node stdout %q", err, truth.stdout)
		})
	}
}

// Array fixtures need no host capabilities. Compare real WASI commands with Node
// without treating target refusals or traps as successful observations.
func TestArrayHolesWasmtime(t *testing.T) {
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	runner := os.Getenv("ADAMIC_WASMTIME")
	if runner == "" {
		var err error
		runner, err = exec.LookPath("wasmtime")
		if err != nil {
			t.Fatal(err)
		}
	}
	count := 0
	for _, fixture := range fixtures {
		if !strings.Contains(fixture.path, "library_array_holes_") {
			continue
		}
		count++
		t.Run(fixture.path, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, fixture.path))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			binary := filepath.Join(t.TempDir(), "array.wasm")
			if err := native.Build(native.C(program), binary, native.Options{Target: "wasm32-wasi"}); err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			actual := execute(t, runner, "run", binary)
			if difference := disagreement(truth, actual); difference != "" {
				t.Fatalf("%s: Node exit %d stdout %q stderr %q; wasmtime exit %d stdout %q stderr %q", difference, truth.exitCode, truth.stdout, truth.stderr, actual.exitCode, actual.stdout, actual.stderr)
			}
			t.Logf("Node and wasmtime stdout: %s", truth.stdout)
		})
	}
	if count != 10 {
		t.Fatalf("want 10 Array fixtures, got %d", count)
	}
}

func TestArrayHolesWasmtimeRunnerMutants(t *testing.T) {
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	runner := os.Getenv("ADAMIC_WASMTIME")
	if runner == "" {
		var err error
		runner, err = exec.LookPath("wasmtime")
		if err != nil {
			t.Fatal(err)
		}
	}
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_array_holes_scanner_probe.a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	for _, probe := range []struct{ name, source, difference string }{
		{"control", "#include <stdio.h>\nint main(void) { puts(\"4\"); return 0; }\n", ""},
		{"stdout", "#include <stdio.h>\nint main(void) { puts(\"5\"); return 0; }\n", "stdout differs"},
		{"exit", "#include <stdio.h>\nint main(void) { puts(\"4\"); return 23; }\n", "exit codes differ"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			binary := filepath.Join(t.TempDir(), "runner.wasm")
			if err := native.Build(probe.source, binary, native.Options{Target: "wasm32-wasi"}); err != nil {
				t.Fatal(err)
			}
			if difference := disagreement(truth, execute(t, runner, "run", binary)); difference != probe.difference {
				t.Fatalf("want %q, got %q", probe.difference, difference)
			}
		})
	}
}
