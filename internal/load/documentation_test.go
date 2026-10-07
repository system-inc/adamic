package load

import (
	"os/exec"
	"testing"
)

// The audit covers tracked source, comments, links and reports, while preserving
// upstream documentation and the gitignore oracle's synthetic paths.
func TestDocumentationPathsStayRenamed(t *testing.T) {
	t.Parallel()
	command := exec.Command("python3", "cloud/check-documentation-paths.py")
	command.Dir = "../.."
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("documentation path audit: %v\n%s", err, output)
	}
}
