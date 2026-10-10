package lint

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"regexp"
	"testing"
)

// Selecting one shard in a fresh process must look up its own shared inputs: both JSX oracles and the bundle that
// consumes them, each by its own key, with nothing carried over from another test. The child inherits this run's
// products (ADAMIC_BUILD_CACHE, ADAMIC_BUILD_CACHE_DIR), so on a runner it reads Workshop's and never builds; that each
// builds alone from an empty cache is what its TestProduct_ proves on Workshop.
func TestJsxLintTreesSetupIsolation(t *testing.T) {
	t.Parallel()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	command := exec.CommandContext(context.Background(), executable, "-test.run=^TestJsxLintTrees_001$", "-test.v", "-test.skip=^TestProduct_")
	command.Env = append(os.Environ(), "ADAMIC_TEST_SHARD=", "ADAMIC_JSX_SHARD_CHILD=")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("standalone shard failed: %v\n%s", err, output)
	}
	for _, product := range []string{"jsx-parser", "jsx-membership", "jsx-tree-setup-v1"} {
		if !productLookedUp(output, regexp.QuoteMeta(product)) {
			t.Fatalf("the child never looked up %s itself:\n%s", product, output)
		}
	}
	if !bytes.Contains(output, []byte("--- PASS: TestJsxLintTrees_001")) {
		t.Fatalf("the selected shard didn't pass:\n%s", output)
	}
}

// productLookedUp is whether output holds buildcache's census line for a product whose whole name matches name, a
// regular expression: the process found it (hit, fetched, audited) or built it (miss) itself.
func productLookedUp(output []byte, name string) bool {
	return regexp.MustCompile(`(?m)build (` + name + `) [0-9a-f]{12} (hit|miss|fetched|audited) `).Match(output)
}
