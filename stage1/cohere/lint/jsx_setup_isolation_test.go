package lint

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Selecting one shard in a fresh process must build its own shared inputs.
func TestJsxLintTreesSetupIsolation(t *testing.T) {
	t.Parallel()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	// Preserve build-phase dependencies, but require the fresh child to build
	// both JSX oracles and the bundle that consumes them itself.
	ready := jsxPrepareTrees(t)
	excluded := map[string]bool{
		filepath.Base(ready):                                true,
		filepath.Base(filepath.Dir(jsxParserOracle(t))):     true,
		filepath.Base(filepath.Dir(jsxMembershipOracle(t))): true,
	}
	cache := t.TempDir()
	entries, err := os.ReadDir(filepath.Dir(ready))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if !entry.IsDir() || excluded[entry.Name()] {
			continue
		}
		if err := os.Symlink(filepath.Join(filepath.Dir(ready), entry.Name()), filepath.Join(cache, entry.Name())); err != nil {
			t.Fatal(err)
		}
	}
	command := exec.CommandContext(context.Background(), executable, "-test.run=^TestJsxLintTrees_001$", "-test.v", "-test.skip=^TestProduct_")
	command.Env = append(os.Environ(), "ADAMIC_BUILD_CACHE_DIR="+cache, "ADAMIC_BUILD_CACHE=on", "ADAMIC_TEST_SHARD=", "ADAMIC_JSX_SHARD_CHILD=")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("standalone cold shard failed: %v\n%s", err, output)
	}
	for _, evidence := range []string{"build jsx-parser", "build jsx-membership", " miss ", "--- PASS: TestJsxLintTrees_001"} {
		if !bytes.Contains(output, []byte(evidence)) {
			t.Fatalf("missing preparation proof %q:\n%s", evidence, output)
		}
	}
}
