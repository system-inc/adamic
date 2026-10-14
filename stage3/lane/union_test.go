package lane

import (
	"context"
	"testing"
	"time"
)

func TestUnionIdentityOutcomes(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	command := boundedCommand(ctx, repositoryRoot(t), "python3", "-B", "-m", "unittest", "test_union")
	command.Dir = repositoryRoot(t) + "/stage3/lane"
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("union fixture: %v: %s", err, output)
	}
}
