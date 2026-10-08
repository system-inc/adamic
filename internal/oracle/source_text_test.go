package oracle

import (
	"os"
	"path/filepath"
	"testing"
)

// V8 observes these exact function spans in Array.prototype.shift's TypeError.
// Direct Node type stripping is the independent control for our source loader.
func TestNodeOraclePreservesFunctionSource(t *testing.T) {
	t.Parallel()
	path := filepath.Join(repository, "internal/oracle/testdata/library_array_shift_function.a")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	truthPath := filepath.Join(t.TempDir(), "truth.mts")
	if err := os.WriteFile(truthPath, data, 0600); err != nil {
		t.Fatal(err)
	}
	truth := execute(t, "node", "--disable-warning=ExperimentalWarning", truthPath)
	actual := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), path)
	if truth.exitCode != 0 || actual.exitCode != 0 || len(truth.stderr) != 0 || len(actual.stderr) != 0 {
		t.Fatalf("source text control must finish cleanly: truth=%+v actual=%+v", truth, actual)
	}
	if difference := disagreement(truth, actual); difference != "" {
		t.Fatal(difference)
	}
}

func TestNodeOracleTransformsRuntimeTypeScript(t *testing.T) {
	t.Parallel()
	data := []byte("enum Choice { No, Yes } console.log(`${Choice.Yes}`);\n")
	dir := t.TempDir()
	path, truthPath := filepath.Join(dir, "source.a"), filepath.Join(dir, "truth.mts")
	for _, name := range []string{path, truthPath} {
		if err := os.WriteFile(name, data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	truth := execute(t, "node", "--disable-warning=ExperimentalWarning", "--experimental-transform-types", truthPath)
	actual := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), path)
	if truth.exitCode != 0 || string(truth.stdout) != "1\n" || len(truth.stderr) != 0 {
		t.Fatalf("bad independent Node enum control: %+v", truth)
	}
	if difference := disagreement(truth, actual); difference != "" {
		t.Fatal(difference)
	}
}
