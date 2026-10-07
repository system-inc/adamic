package cloud

import (
	"context"
	"os/exec"
	"testing"
	"time"
)

// Keep the repository baseline in the ordinary Go gate, where new findings cannot be overlooked.
func TestRepositoryPassesCohereBaseline(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, "python3", "cohere_gate.py")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("cohere baseline: %v\n%s", err, output)
	}
	t.Logf("%s", output)
}
