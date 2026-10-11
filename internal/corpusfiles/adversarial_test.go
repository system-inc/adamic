package corpusfiles

// The independent review's adversarial cases for a source with no .git (#cyasrr4), kept as tests.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/tracked"
)

func g(t *testing.T, dir string, args ...string) string {
	t.Helper()
	c := exec.Command("git", append([]string{"-C", dir, "-c", "user.name=x", "-c", "user.email=x@x", "-c", "commit.gpgsign=false", "-c", "protocol.file.allow=always"}, args...)...)
	out, err := c.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func w(t *testing.T, root, name, text string) {
	t.Helper()
	p := filepath.Join(root, name)
	os.MkdirAll(filepath.Dir(p), 0o755)
	if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

// Copy a checkout's worktree (minus .git) into dst.
func unpack(t *testing.T, src, dst string) {
	t.Helper()
	filepath.WalkDir(src, func(p string, e os.DirEntry, err error) error {
		rel, _ := filepath.Rel(src, p)
		if e.Name() == ".git" {
			if e.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if e.IsDir() {
			return nil
		}
		data, _ := os.ReadFile(p)
		w(t, dst, rel, string(data))
		return nil
	})
}

// A1: Repository mode, a manifest from an older commit than the tree: a corpus file added since is named, never
// silently dropped from a selection smaller than git's.
func TestAdversarialStaleManifestRepository(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	g(t, repo, "init", "-q")
	w(t, repo, "docs/a.md", "a\n")
	g(t, repo, "add", ".")
	g(t, repo, "commit", "-qm", "1")
	source := t.TempDir()
	if err := tracked.Write(repo, source); err != nil { // manifest at commit 1
		t.Fatal(err)
	}
	w(t, repo, "docs/b.md", "this one would fail\n")
	g(t, repo, "add", ".")
	g(t, repo, "commit", "-qm", "2")
	unpack(t, repo, source) // tree at commit 2
	viaGit, _, _, err := selectFiles(repo, "", []string{"docs"}, []string{"*.md"})
	if err != nil {
		t.Fatal(err)
	}
	viaManifest, _, _, err := selectFiles(source, "", []string{"docs"}, []string{"*.md"})
	t.Logf("git: %d files %v", len(viaGit), viaGit)
	t.Logf("stale manifest: %d files %v err=%v", len(viaManifest), viaManifest, err)
	if err == nil || !strings.Contains(err.Error(), "docs/b.md") {
		t.Errorf("a stale manifest passes Repository with %d files where git selects %d", len(viaManifest), len(viaGit))
	}
}

// A2: Upstream mode: HEAD names the pin and matches the parent's gitlink, but files (and the tree) are another
// commit's. The commit object binds them: the listing rebuilds a tree the pinned commit doesn't name.
func TestAdversarialHeadNotBoundToFiles(t *testing.T) {
	t.Parallel()
	sub := t.TempDir()
	g(t, sub, "init", "-q")
	w(t, sub, "x/a.css", "a{}\n")
	g(t, sub, "add", ".")
	g(t, sub, "commit", "-qm", "1")
	old := g(t, sub, "rev-parse", "HEAD")
	oldFiles := g(t, sub, "ls-tree", "-r", "--full-tree", "HEAD")
	_ = oldFiles
	w(t, sub, "x/a.css", "b{}\n")
	g(t, sub, "commit", "-qam", "2")
	pin := g(t, sub, "rev-parse", "HEAD")
	parent := t.TempDir()
	g(t, parent, "init", "-q")
	w(t, parent, "readme", "r\n")
	g(t, parent, "add", ".")
	g(t, parent, "-c", "protocol.file.allow=always", "submodule", "add", "-q", sub, "cohere")
	g(t, parent, "commit", "-qm", "p")
	source := t.TempDir()
	if err := tracked.Write(parent, source); err != nil {
		t.Fatal(err)
	}
	// Replace cohere's files with the old commit's listing and the tree with old content; HEAD stays the pin.
	listing, _ := exec.Command("git", "-C", sub, "ls-tree", "-r", "-z", "--full-tree", old).Output()
	os.WriteFile(filepath.Join(source, ".tracked/cohere/files"), listing, 0o644)
	w(t, source, "readme", "r\n")
	w(t, source, "cohere/x/a.css", "a{}\n")
	files, head, _, err := selectFiles(filepath.Join(source, "cohere"), pin, []string{"x"}, []string{"*.css"})
	t.Logf("head=%s pin=%s files=%v err=%v", head, pin, files, err)
	if err == nil || !strings.Contains(err.Error(), "aren't that commit's") {
		t.Errorf("Upstream passes on pin %s while the tree and files are commit %s: %v", pin, old, err)
	}
}

// A3: an extra file under a root is refused by Upstream, as git status does, and by Repository too, since without
// .git it can't be told from a file a stale manifest left out.
func TestAdversarialUntracked(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	g(t, repo, "init", "-q")
	w(t, repo, "r/a.md", "a\n")
	g(t, repo, "add", ".")
	g(t, repo, "commit", "-qm", "1")
	pin := g(t, repo, "rev-parse", "HEAD")
	source := t.TempDir()
	tracked.Write(repo, source)
	unpack(t, repo, source)
	w(t, source, "r/extra.md", "x\n")
	_, _, _, errU := selectFiles(source, pin, []string{"r"}, []string{"*.md"})
	files, _, _, errR := selectFiles(source, "", []string{"r"}, []string{"*.md"})
	t.Logf("upstream: %v ; repository: %v %v", errU, files, errR)
	if errU == nil || !strings.Contains(errU.Error(), "r/extra.md") || errR == nil || !strings.Contains(errR.Error(), "r/extra.md") {
		t.Errorf("an extra file passed: upstream %v, repository %v", errU, errR)
	}
}

// A4: broken .git file (worktree pointing nowhere) with a valid manifest present: must not fall back.
func TestAdversarialBrokenGitFile(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	g(t, repo, "init", "-q")
	w(t, repo, "r/a.md", "a\n")
	g(t, repo, "add", ".")
	g(t, repo, "commit", "-qm", "1")
	source := t.TempDir()
	tracked.Write(repo, source)
	unpack(t, repo, source)
	w(t, source, ".git", "gitdir: /nonexistent/worktrees/x\n")
	_, _, _, err := selectFiles(source, "", []string{"r"}, []string{"*.md"})
	t.Logf("broken .git: %v", err)
	if err == nil {
		t.Errorf("broken .git fell back to the manifest")
	}
	// shallow clone answers through git
	shallow := filepath.Join(t.TempDir(), "s")
	g(t, repo, "clone", "-q", "--depth", "1", "file://"+repo, shallow)
	if _, _, _, err := selectFiles(shallow, "", []string{"r"}, []string{"*.md"}); err != nil {
		t.Errorf("shallow: %v", err)
	}
}
