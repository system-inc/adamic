package coherepin

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCaptureAndCheckoutMustMatchGitlink(t *testing.T) {
	root := t.TempDir()
	child := filepath.Join(root, "cohere")
	if err := os.Mkdir(child, 0755); err != nil {
		t.Fatal(err)
	}
	git := func(dir string, args ...string) string {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", dir, "-c", "user.name=Pin fixture", "-c", "user.email=pin@example.invalid"}, args...)...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	git(child, "init")
	git(child, "commit", "--allow-empty", "-m", "Captured source")
	pin := git(child, "rev-parse", "HEAD")
	git(root, "init")
	git(root, "update-index", "--add", "--cacheinfo", "160000,"+pin+",cohere")
	git(root, "commit", "-m", "Pin source")
	if err := Check(root, pin); err != nil {
		t.Fatal(err)
	}
	if err := Check(root, strings.Repeat("0", 40)); err == nil || !strings.Contains(err.Error(), "recapture") {
		t.Fatalf("stale capture accepted: %v", err)
	}
	git(child, "commit", "--allow-empty", "-m", "Unpinned source")
	if err := Check(root, pin); err == nil || !strings.Contains(err.Error(), "checkout") {
		t.Fatalf("checkout drift accepted: %v", err)
	}
}
