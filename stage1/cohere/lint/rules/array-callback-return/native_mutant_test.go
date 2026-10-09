package cfg_rule

import (
	"os/exec"
	"testing"
)

func TestNativeRuleMutant(t *testing.T) {
	t.Parallel()
	command := exec.Command("python3", "testdata/native_mutant.py")
	out, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("native rule mutant: %v\n%s", err, out)
	}
	t.Log(string(out))
}
