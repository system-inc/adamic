package helpers

import (
	"os"
	"os/exec"
	"testing"
)

// Keep the module package in the shared helpers' ordinary gate.
func TestModulePackage(t *testing.T) {
	cmd := exec.Command("go", "test", "./module", "-count=1", "-v", "-timeout=20m")
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()
	t.Logf("%s", output)
	if err != nil {
		t.Fatal(err)
	}
}
