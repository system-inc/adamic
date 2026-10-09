package lint

import (
	"bytes"
	"context"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestCompleteSuggestionSerialization_IsolatedShard(t *testing.T) {
	t.Parallel()
	// A fresh product cache proves the selected shard builds, rather than
	// accidentally relying on products prepared by another top-level test.
	cache := t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()
	command := completeSuggestionCommand(ctx, os.Args[0], "-test.run=^TestCompleteSuggestionSerialization_001$", "-test.timeout=600s", "-test.v")
	command.Env = append(os.Environ(), "ADAMIC_BUILD_CACHE_DIR="+cache, "ADAMIC_BUILD_CACHE=on", "ADAMIC_COMPLETE_SUGGESTION_PLANT=0")
	output, err := command.CombinedOutput()
	if command.Process != nil {
		_ = syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
	}
	if err != nil || !bytes.Contains(output, []byte("--- PASS: TestCompleteSuggestionSerialization_001 ")) {
		t.Fatalf("isolated shard: %v\n%s", err, output)
	}
	if !bytes.Contains(output, []byte("build complete-suggestion-go-oracle ")) || !bytes.Contains(output, []byte(" miss ")) {
		t.Fatalf("isolated shard did not build cold products:\n%s", output)
	}
	t.Log("isolated shard built its shared products and passed without Setup")
}
