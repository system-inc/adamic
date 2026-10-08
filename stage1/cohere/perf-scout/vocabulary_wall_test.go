package perfscout

import (
	"os"
	"os/exec"
	"testing"
)

func TestVocabularyWallAndCacheGuards(t *testing.T) {
	t.Parallel()
	if os.Getenv("ADAMIC_SCOUT_VOCABULARY") == "" {
		t.Fatal("ADAMIC_SCOUT_VOCABULARY is required; no skipped profiles")
	}
	output, err := exec.Command("python3", "-B", "-m", "unittest", "-v", "test_vocabulary_wall").CombinedOutput()
	if err != nil {
		t.Fatalf("%v\n%s", err, output)
	}
	t.Log(string(output))
}
