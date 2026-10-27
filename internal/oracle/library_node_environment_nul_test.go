package oracle

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

func TestProcessEnvironmentNULNamesMatchNode(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	path := filepath.Join(directory, "names.a")
	if err := os.WriteFile(path, []byte("import process from 'node:process';\nconsole.log(process.env['a\\0b']);\nconsole.log(process.env['A\\0B']);\n"), 0600); err != nil {
		t.Fatal(err)
	}
	// In particular, A is present while a is absent; never inherit the box's env.
	environment := []string{"A=set"}
	loader := filepath.Join(repository, "oracle/node.mjs")
	truth := processEnvironmentRun(t, environment, "node", "--disable-warning=ExperimentalWarning", loader, path)
	if truth.exitCode != 0 || len(truth.stderr) != 0 || string(truth.stdout) != "undefined\nset\n" {
		t.Fatalf("Node NUL witness changed: %+v", truth)
	}
	t.Logf("Node stdout: %s", truth.stdout)
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(directory, "native")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(directory, "program.mjs")
	if err := os.WriteFile(script, []byte(javascript.JavaScript(program)), 0600); err != nil {
		t.Fatal(err)
	}
	for backend, got := range map[string]run{
		"native":     processEnvironmentRun(t, environment, binary),
		"JavaScript": processEnvironmentRun(t, environment, "node", "--disable-warning=ExperimentalWarning", loader, script),
	} {
		t.Logf("%s stdout: %s", backend, got.stdout)
		if difference := disagreement(truth, got); difference != "" {
			t.Fatalf("%s: %s", backend, difference)
		}
	}
}
