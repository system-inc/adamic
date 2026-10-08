package perfscout

import (
	"os"
	"os/exec"
	"testing"
)

func TestBoundaryAcceptanceAndMutants(t *testing.T) {
	t.Parallel()
	if os.Getenv("ADAMIC_SCOUT_BOUNDARY") == "" {
		t.Fatal("ADAMIC_SCOUT_BOUNDARY required; no skipped inputs")
	}
	output, err := exec.Command("python3", "-B", "-m", "unittest", "-v", "test_boundary").CombinedOutput()
	if err != nil {
		t.Fatalf("%v\n%s", err, output)
	}
	t.Log(string(output))
}
