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

func TestSelfCompareClosure(test *testing.T) {
	test.Parallel()
	selfCompare(test, "closure")
}

func TestSelfCompareArrayFill(test *testing.T) {
	test.Parallel()
	selfCompare(test, "array-fill")
}

func TestSelfCompareSweep(test *testing.T) {
	test.Parallel()
	selfCompare(test, "sweep")
}

func selfCompare(test *testing.T, name string) {
	test.Helper()
	path, absolutePathError := filepath.Abs(filepath.Join(repository, "stage3/fixtures/self-compare", name+".a"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	expected := onNode(test, path)
	compare := func(mode string, actual run) {
		test.Helper()
		if difference := disagreement(expected, actual); difference != "" {
			test.Fatalf(
				"%s: %s; got exit %d stdout %q stderr %q",
				mode,
				difference,
				actual.exitCode,
				actual.stdout,
				actual.stderr,
			)
		}
	}
	compare("JavaScript", onJavaScriptBackend(test, program))
	compare("release", releasedUncached(test, program))
	sanitized, binary := nativelyUncached(test, program)
	compare("sanitized", sanitized)
	if report := leaksUncached(test, program, binary); report != "" {
		test.Fatal(report)
	}
	countedBinary := filepath.Join(test.TempDir(), "counted")
	if buildError := native.Build(native.C(program), countedBinary, native.Options{Count: true}); buildError != nil {
		test.Fatal(buildError)
	}
	counted := execute(test, countedBinary)
	if report := leakcheck.Unbalanced(leakRun(counted)); report != "" {
		test.Fatal(report)
	}
	counted.stderr = countsLine.ReplaceAll(counted.stderr, nil)
	compare("counted", counted)
	if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
		compare("wasm32-wasi", onWASI(test, native.C(program)))
	}
}

func TestSelfCompareConstantMutant(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(filepath.Join(repository, "stage3/fixtures/self-compare/sweep.a"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
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
		test.Fatal("mutant did not replace numeric self-equality")
	}
	binary := filepath.Join(test.TempDir(), "mutant")
	if buildError := native.Build(source, binary, native.Options{Sanitize: true}); buildError != nil {
		test.Fatal(buildError)
	}
	actual := execute(test, binary)
	if actual.exitCode != 0 || len(actual.stderr) != 0 {
		test.Fatalf("mutant failed outside the oracle: %+v", actual)
	}
	if difference := disagreement(onNode(test, path), actual); difference != "stdout differs" {
		test.Fatalf("mutant: %q", difference)
	}
	if !strings.Contains(string(actual.stdout), "NaN: true true\n") {
		test.Fatalf("NaN did not catch folding: %q", actual.stdout)
	}
	test.Log("constant self-equality mutant compiled and ran; sweep.a caught NaN: true true instead of false true")
}
