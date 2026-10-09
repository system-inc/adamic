package oracle

import (
	"path/filepath"
	"testing"
)

func TestGeneratorBasic(t *testing.T)          { t.Parallel(); generatorOracle(t, "basic") }
func TestGeneratorCancellation(t *testing.T)   { t.Parallel(); generatorOracle(t, "cancellation") }
func TestGeneratorThrow(t *testing.T)          { t.Parallel(); generatorOracle(t, "throw") }
func TestGeneratorDefaults(t *testing.T)       { t.Parallel(); generatorOracle(t, "defaults") }
func TestGeneratorClose(t *testing.T)          { t.Parallel(); generatorOracle(t, "close") }
func TestGeneratorDelegate(t *testing.T)       { t.Parallel(); generatorOracle(t, "delegate") }
func TestGeneratorDelegateThrow(t *testing.T)  { t.Parallel(); generatorOracle(t, "delegate_throw") }
func TestGeneratorReentrant(t *testing.T)      { t.Parallel(); generatorOracle(t, "reentrant") }
func TestGeneratorOwned(t *testing.T)          { t.Parallel(); generatorOracle(t, "owned") }
func TestGeneratorCaptures(t *testing.T)       { t.Parallel(); generatorOracle(t, "captures") }
func TestGeneratorDefaultError(t *testing.T)   { t.Parallel(); generatorOracle(t, "default_error") }
func TestGeneratorMapIterator(t *testing.T)    { t.Parallel(); generatorOracle(t, "map_iterator") }
func TestGeneratorChecker(t *testing.T)        { t.Parallel(); generatorOracle(t, "checker") }
func TestGeneratorFirstReference(t *testing.T) { t.Parallel(); generatorOracle(t, "first_reference") }
func TestGeneratorCompletions(t *testing.T)    { t.Parallel(); generatorOracle(t, "completions") }
func generatorOracle(t *testing.T, name string) {
	t.Helper()
	generatorOraclePath(t, filepath.Join(repository, "internal/oracle/testdata/generators", name+".a"))
}
func generatorOraclePath(t *testing.T, source string) {
	t.Helper()
	path, err := filepath.Abs(source)
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 {
		t.Fatalf("Node: %+v", truth)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	js := onJavaScriptBackend(t, program)
	if diff := disagreement(truth, js); diff != "" {
		t.Fatalf("JavaScript: %s: %s", diff, string(js.stdout)+string(js.stderr))
	}
	got, binary := nativelyUncached(t, program)
	if diff := disagreement(truth, got); diff != "" {
		t.Fatalf("sanitized: %s: stdout=%s stderr=%s", diff, got.stdout, got.stderr)
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
	if diff := disagreement(truth, releasedUncached(t, program)); diff != "" {
		t.Fatalf("release: %s", diff)
	}
}

func init() {
	for _, name := range []string{"basic", "cancellation", "throw", "defaults", "close", "delegate", "delegate_throw", "reentrant", "owned", "captures", "default_error", "map_iterator", "checker", "first_reference", "completions"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/generators/" + name + ".a", true, false})
	}
}

func TestGeneratorRegExpIteratorResultCompatibility(t *testing.T) {
	t.Parallel()
	generatorOraclePath(t, filepath.Join(repository, "internal/oracle/testdata/regexp.a"))
}
func TestGeneratorRegExpTreeResultCompatibility(t *testing.T) {
	t.Parallel()
	generatorOraclePath(t, filepath.Join(repository, "internal/fresh/testdata/regexp_tree.ts"))
}
