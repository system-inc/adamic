package cloud

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"
)

// cohereGateVariable turns the cohere baseline on. Ordinary loops skip it: it builds cohere and runs four
// phases over the repository, about two minutes and 232s of CPU. Every gate run sets it, and the skip census
// classes this skip as a required input, so a gate that forgets it fails rather than passing silently.
const cohereGateVariable = "ADAMIC_GATE_COHERE"

// Keep the repository baseline in the gate, where new findings cannot be overlooked.
func TestRepositoryPassesCohereBaseline(t *testing.T) {
	t.Parallel()
	if os.Getenv(cohereGateVariable) != "1" {
		// census: required-input The dedicated cohere gate shard sets ADAMIC_GATE_COHERE=1 and runs TestRepositoryPassesCohereBaseline. cloud/cohere_gate.py builds the pinned cohere submodule and checks repository sources and docs against cloud/cohere-baseline.json; setup supplies Node v24.19.0, Go and Python 3.
		t.Skipf("%s is not 1: the cohere baseline runs in the gate, not in ordinary loops", cohereGateVariable)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, "python3", "cohere_gate.py")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("cohere baseline: %v\n%s", err, output)
	}
	t.Logf("%s", output)
}
