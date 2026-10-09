package corpus

import (
	"os"
	"os/exec"
	"testing"
)

// runProbe runs the test named name again in a child process with the Ahra corpus's variable set to
// value and COHERE_CORPUS_PROBE set to probe, and returns its verbose output.
func runProbe(t *testing.T, name, probe, value string) string {
	t.Helper()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command(executable, "-test.run", "^"+name+"$", "-test.v")
	command.Env = append(os.Environ(), "COHERE_CORPUS_PROBE="+probe, Ahra.Variable+"="+value)
	output, _ := command.CombinedOutput()
	return string(output)
}
