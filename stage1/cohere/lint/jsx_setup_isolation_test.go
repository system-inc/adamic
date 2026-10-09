package lint

import (
	"bytes"
	"os"
	"testing"
)

// Selecting one shard in a fresh process must build its own shared inputs.
func TestJsxLintTreesSetupIsolation(t *testing.T) {
	t.Parallel()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	command := lintBuildPhaseChild(t, executable, "-test.run=^TestJsxLintTrees_001$", "-test.v", "-test.timeout=600s")
	command.Env = append(os.Environ(), "ADAMIC_BUILD_CACHE_DIR="+t.TempDir(), "ADAMIC_BUILD_CACHE=on", "ADAMIC_TEST_SHARD=", "ADAMIC_JSX_SHARD_CHILD=")
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
