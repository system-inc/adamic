package cloud

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// The CLI tests use only disposable local Git origins and synthetic Go events.
// Real three-way worker execution is recorded separately in the unit report.
func TestLintWaveCheck(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "lint-wave-check-tests.txt")
	output, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command("python3", "-B", "lint-wave-check-test.py")
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
		t.Fatalf("worker check CLI tests: %v\n%s", runError, data)
	}
	t.Logf("worker check CLI tests passed:\n%s", data)
}
