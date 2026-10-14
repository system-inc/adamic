package cloud

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/toolchain"
)

// The gate tools' own tests (#74z4tth): run.py decides what every change runs and what it may declare, so a change to
// it runs fast-gate/run_test.py, every class, as this package's test. python3 is toolchain's, and it runs with go's
// environment, since the tests drive go list and git over this tree and over fixtures of their own.
func TestGateToolsTests(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "gate-tools-tests.txt")
	output, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	command := toolchain.CommandIn("fast-gate", "python3", "-B", "-m", "unittest", "run_test")
	if command.Err == nil {
		command.Env = toolchain.Env()
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
		t.Fatalf("gate tools tests: %v\n%s", runError, data)
	}
	t.Logf("gate tools tests passed:\n%s", data)
}
