package cloud

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/toolchain"
)

// submit-gates.py's tests, run here so a gate runs them (#v3q9ayd): each rule of cloud/gates/diff-to-gates.json firing
// and not, the fresh base rule, a recipe whose hash differs, and a head with one gate's job deleted, refused by name.
// They need nothing but python3 and git: a disposable repository with its own identity and no global configuration,
// and job rows written by the test, no network and no Loom.
func TestSubmitGates(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "submit-gates-tests.txt")
	output, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	command := toolchain.Command("python3", "-B", "-m", "unittest", "-v", "submit_gates_test")
	if command.Err == nil {
		command.Env = toolchain.GitEnv()
	}
	command.Stdout = output
	command.Stderr = output
	runError := command.Run()
	if err := output.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if runError != nil {
		t.Fatalf("submit-gates tests: %v\n%s", runError, data)
	}
	t.Logf("submit-gates tests passed:\n%s", data)
}
