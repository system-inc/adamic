package typescripthelpers

import (
	"os/exec"
	"testing"
)

// Not parallel: this test compiles the baseline and three mutants in sequence to bound memory.
func TestTypescriptHelperCapturesAndMutants(t *testing.T) {
	output, err := exec.Command("python3", "validate.py").CombinedOutput()
	t.Log(string(output))
	if err != nil {
		t.Fatal(err)
	}
}
