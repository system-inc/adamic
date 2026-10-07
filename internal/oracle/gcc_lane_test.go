package oracle

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// This lane deliberately has no observation cache. The compiler can be switched to clang to
// measure the same work on the same fixtures, rather than comparing different harnesses.
func TestGCCAgreesWithNode(t *testing.T) {
	if os.Getenv("ADAMIC_GCC_LANE") != "1" {
		t.Skip("opt in with ADAMIC_GCC_LANE=1")
	}
	compiler := os.Getenv("ADAMIC_LANE_CC")
	if compiler == "" {
		compiler = "gcc"
	}
	sanitize := os.Getenv("ADAMIC_LANE_SANITIZE") == "1"
	options := native.Options{Compiler: compiler, Sanitize: sanitize}
	t.Logf("compiler=%s flags=%s", compiler, strings.Join(native.Flags(options), " "))
	// Build once before the fixture fan-out. A failed runtime still allows every emitted
	// translation unit to be checked independently for GCC diagnostics.
	_, runtimeError := native.RuntimeLibrary("", options)
	for _, fixture := range fixtures {
		if !fixture.lowers {
			continue
		}
		t.Run(fixture.path, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, fixture.path))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			// Execute Node directly: even a shell without ADAMIC_GATE_UNCACHED cannot reuse results.
			oracle := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle", "node.mjs"), path)
			if fixture.checked {
				oracle = onJavaScriptBackend(t, program)
			}
			binary := filepath.Join(t.TempDir(), "program")
			code := native.C(program)
			if runtimeError != nil {
				source := filepath.Join(filepath.Dir(binary), "main.c")
				if err := os.WriteFile(source, []byte(code), 0o644); err != nil {
					t.Fatal(err)
				}
				arguments := append(native.Flags(options), "-I", filepath.Join(repository, "internal/native/runtime"), "-c", source, "-o", binary+".o")
				output, err := exec.Command(compiler, arguments...).CombinedOutput()
				if err != nil {
					t.Errorf("emitted C: %v\n%s", err, output)
				}
				t.Fatalf("runtime blocks linking: %v", runtimeError)
			}
			if err := native.Build(code, binary, options); err != nil {
				t.Fatal(err)
			}
			actual := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary)
			if difference := disagreement(oracle, actual); difference != "" {
				t.Errorf("%s: Node exit %d, native exit %d\nstdout %s\nstderr %s", difference, oracle.exitCode, actual.exitCode, firstLaneDifference(oracle.stdout, actual.stdout), firstLaneDifference(oracle.stderr, actual.stderr))
			}
			if fixture.checked && actual.exitCode != 70 {
				t.Errorf("inserted check did not fire: exit %d", actual.exitCode)
			}
			if sanitize && oracle.exitCode == 0 {
				if report := leakSanitizer(t, binary); report != "" {
					t.Errorf("leaks: %s", report)
				}
			}
		})
	}
}

func firstLaneDifference(want, got []byte) string {
	if bytes.Equal(want, got) {
		return "identical"
	}
	left, right := bytes.Split(want, []byte("\n")), bytes.Split(got, []byte("\n"))
	for line := 0; line < len(left) || line < len(right); line++ {
		var a, b []byte
		if line < len(left) {
			a = left[line]
		}
		if line < len(right) {
			b = right[line]
		}
		if !bytes.Equal(a, b) || line >= len(left) || line >= len(right) {
			return fmt.Sprintf("line %d: Node %q, native %q", line+1, a, b)
		}
	}
	return "different bytes"
}

// The comparison uses the same C and Node runners on a real fixture with one changed byte.
// Clang keeps this proof runnable when GCC cannot compile the runtime under the lane's flags.
func TestGCCLaneComparisonCatchesMutants(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "dedication/dedication.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	program.Strings[0] += "!"
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(native.C(program), binary, native.Options{}); err != nil {
		t.Fatal(err)
	}
	truth := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), path)
	mutant := execute(t, binary)
	if difference := disagreement(truth, mutant); difference != "stdout differs" {
		t.Fatalf("mutant: %q", difference)
	}
	if difference := firstLaneDifference(truth.stdout, mutant.stdout); !strings.HasPrefix(difference, "line 1:") {
		t.Fatal(difference)
	}
}
