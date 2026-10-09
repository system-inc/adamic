package oracle

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
)

var selfCompareFixtures = []string{"closure", "array-fill", "sweep"}

func init() {
	for _, name := range selfCompareFixtures {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"stage3/fixtures/self-compare/" + name + ".a", true, false})
	}
}

func TestSelfCompare(t *testing.T) {
	for _, name := range selfCompareFixtures {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "stage3/fixtures/self-compare", name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			expected := onNode(t, path)
			compare := func(mode string, actual run) {
				t.Helper()
				if difference := disagreement(expected, actual); difference != "" {
					t.Fatalf("%s: %s; got exit %d stdout %q stderr %q", mode, difference, actual.exitCode, actual.stdout, actual.stderr)
				}
			}
			compare("JavaScript", onJavaScriptBackend(t, program))
			compare("release", releasedUncached(t, program))
			sanitized, binary := nativelyUncached(t, program)
			compare("sanitized", sanitized)
			if report := leaksUncached(t, program, binary); report != "" {
				t.Fatal(report)
			}
			countedBinary := filepath.Join(t.TempDir(), "counted")
			if err := native.Build(native.C(program), countedBinary, native.Options{Count: true}); err != nil {
				t.Fatal(err)
			}
			counted := execute(t, countedBinary)
			if report := leakcheck.Unbalanced(leakRun(counted)); report != "" {
				t.Fatal(report)
			}
			counted.stderr = countsLine.ReplaceAll(counted.stderr, nil)
			compare("counted", counted)
			if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
				compare("wasm32-wasi", onWASI(t, native.C(program)))
			}
		})
	}
}

func TestSelfCompareConstantMutant(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "stage3/fixtures/self-compare/sweep.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	calls := regexp.MustCompile(`adamic_same_number\((adamic_local_[A-Za-z0-9_]+), (adamic_local_[A-Za-z0-9_]+)\)`)
	replaced := false
	source = calls.ReplaceAllStringFunc(source, func(call string) string {
		operands := calls.FindStringSubmatch(call)
		if !replaced && operands[1] == operands[2] {
			replaced = true
			return "true"
		}
		return call
	})
	if !replaced {
		t.Fatal("mutant did not replace numeric self-equality")
	}
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	actual := execute(t, binary)
	if actual.exitCode != 0 || len(actual.stderr) != 0 {
		t.Fatalf("mutant failed outside the oracle: %+v", actual)
	}
	if difference := disagreement(onNode(t, path), actual); difference != "stdout differs" {
		t.Fatalf("mutant: %q", difference)
	}
	if !strings.Contains(string(actual.stdout), "NaN: true true\n") {
		t.Fatalf("NaN did not catch folding: %q", actual.stdout)
	}
	t.Log("constant self-equality mutant compiled and ran; sweep.a caught NaN: true true instead of false true")
}
