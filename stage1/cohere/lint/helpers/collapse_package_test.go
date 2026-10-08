package helpers

import (
	"os/exec"
	"testing"
)

func TestCollapsePackage(t *testing.T) {
	command := exec.Command("go", "test", "./tailwind/collapse", "-count=1", "-v", "-timeout=20m")
	output, err := command.CombinedOutput()
	t.Logf("%s", output)
	if err != nil {
		t.Fatal(err)
	}
}
