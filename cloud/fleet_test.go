package cloud

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestFleetFastRelaunch(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "fleet-tests.log")
	output, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command("python3", "-B", "test_fleet.py")
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
		t.Fatalf("fleet tests: %v\n%s", runError, data)
	}
	t.Logf("fleet tests passed:\n%s", data)
}
