package corpusfiles

// The independent review's adversarial cases for a source with no .git (#cyasrr4), kept as tests.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/toolchain"
	"github.com/system-inc/adamic/internal/tracked"
)

func g(test *testing.T, dir string, args ...string) string {
	test.Helper()
	c := toolchain.GitCommand(append([]string{
		"-C",
		dir,
		"-c",
		"user.name=x",
		"-c",
		"user.email=x@x",
		"-c",
		"commit.gpgsign=false",
		"-c",
		"protocol.file.allow=always",
	}, args...)...)
	out, err := c.CombinedOutput()
	if err != nil {
		test.Fatalf("git %v: %v %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func w(test *testing.T, root, name, text string) {
	test.Helper()
	p := filepath.Join(root, name)
	os.MkdirAll(filepath.Dir(p), 0o755)
	if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
		test.Fatal(err)
	}
}

// Copy a checkout's worktree (minus .git) into dst.
func unpack(test *testing.T, src, dst string) {
	test.Helper()
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
		w(test, dst, rel, string(data))
		return nil
	})
}

// A1: Repository mode, a manifest from an older commit than the tree: a corpus file added since is named, never
// silently dropped from a selection smaller than git's.
func TestAdversarialStaleManifestRepository(test *testing.T) {
	test.Parallel()
	repo := test.TempDir()
	g(test, repo, "init", "-q")
	w(test, repo, "docs/a.md", "a\n")
	g(test, repo, "add", ".")
	g(test, repo, "commit", "-qm", "1")
	source := test.TempDir()
	if err := tracked.Write(repo, source); err != nil { // manifest at commit 1
		test.Fatal(err)
	}
	w(test, repo, "docs/b.md", "this one would fail\n")
	g(test, repo, "add", ".")
	g(test, repo, "commit", "-qm", "2")
	unpack(test, repo, source) // tree at commit 2
	viaGit, _, _, err := selectFiles(repo, "", []string{"docs"}, []string{"*.md"}, nil)
	if err != nil {
		test.Fatal(err)
	}
	viaManifest, _, _, err := selectFiles(source, "", []string{"docs"}, []string{"*.md"}, nil)
	test.Logf("git: %d files %v", len(viaGit), viaGit)
	test.Logf("stale manifest: %d files %v err=%v", len(viaManifest), viaManifest, err)
	if err == nil || !strings.Contains(err.Error(), "docs/b.md") {
		test.Errorf(
			"a stale manifest passes Repository with %d files where git selects %d",
			len(viaManifest),
			len(viaGit),
		)
	}
}

// A2: Upstream mode: HEAD names the pin and matches the parent's gitlink, but files (and the tree) are another
// commit's. The commit object binds them: the listing rebuilds a tree the pinned commit doesn't name.
func TestAdversarialHeadNotBoundToFiles(test *testing.T) {
	test.Parallel()
	sub := test.TempDir()
	g(test, sub, "init", "-q")
	w(test, sub, "x/a.css", "a{}\n")
	g(test, sub, "add", ".")
	g(test, sub, "commit", "-qm", "1")
	old := g(test, sub, "rev-parse", "HEAD")
	oldFiles := g(test, sub, "ls-tree", "-r", "--full-tree", "HEAD")
	_ = oldFiles
	w(test, sub, "x/a.css", "b{}\n")
	g(test, sub, "commit", "-qam", "2")
	pin := g(test, sub, "rev-parse", "HEAD")
	parent := test.TempDir()
	g(test, parent, "init", "-q")
	w(test, parent, "readme", "r\n")
	g(test, parent, "add", ".")
	g(test, parent, "-c", "protocol.file.allow=always", "submodule", "add", "-q", sub, "cohere")
	g(test, parent, "commit", "-qm", "p")
	source := test.TempDir()
	if err := tracked.Write(parent, source); err != nil {
		test.Fatal(err)
	}
	// Replace cohere's files with the old commit's listing and the tree with old content; HEAD stays the pin.
	listing, _ := toolchain.GitCommand("-C", sub, "ls-tree", "-r", "-z", "--full-tree", old).Output()
	os.WriteFile(filepath.Join(source, ".tracked/cohere/files"), listing, 0o644)
	w(test, source, "readme", "r\n")
	w(test, source, "cohere/x/a.css", "a{}\n")
	files, head, _, err := selectFiles(filepath.Join(source, "cohere"), pin, []string{"x"}, []string{"*.css"}, nil)
	test.Logf("head=%s pin=%s files=%v err=%v", head, pin, files, err)
	if err == nil || !strings.Contains(err.Error(), "aren't that commit's") {
		test.Errorf("Upstream passes on pin %s while the tree and files are commit %s: %v", pin, old, err)
	}
}

// A3: an extra file under a root is refused by Upstream, as git status does, and by Repository too, since without
// .git it can't be told from a file a stale manifest left out.
func TestAdversarialUntracked(test *testing.T) {
	test.Parallel()
	repo := test.TempDir()
	g(test, repo, "init", "-q")
	w(test, repo, "r/a.md", "a\n")
	g(test, repo, "add", ".")
	g(test, repo, "commit", "-qm", "1")
	pin := g(test, repo, "rev-parse", "HEAD")
	source := test.TempDir()
	tracked.Write(repo, source)
	unpack(test, repo, source)
	w(test, source, "r/extra.md", "x\n")
	_, _, _, errU := selectFiles(source, pin, []string{"r"}, []string{"*.md"}, nil)
	files, _, _, errR := selectFiles(source, "", []string{"r"}, []string{"*.md"}, nil)
	test.Logf("upstream: %v ; repository: %v %v", errU, files, errR)
	if errU == nil ||
		!strings.Contains(errU.Error(), "r/extra.md") ||
		errR == nil ||
		!strings.Contains(errR.Error(), "r/extra.md") {
		test.Errorf("an extra file passed: upstream %v, repository %v", errU, errR)
	}
}

// A4: broken .git file (worktree pointing nowhere) with a valid manifest present: must not fall back.
func TestAdversarialBrokenGitFile(test *testing.T) {
	test.Parallel()
	repo := test.TempDir()
	g(test, repo, "init", "-q")
	w(test, repo, "r/a.md", "a\n")
	g(test, repo, "add", ".")
	g(test, repo, "commit", "-qm", "1")
	source := test.TempDir()
	tracked.Write(repo, source)
	unpack(test, repo, source)
	w(test, source, ".git", "gitdir: /nonexistent/worktrees/x\n")
	_, _, _, err := selectFiles(source, "", []string{"r"}, []string{"*.md"}, nil)
	test.Logf("broken .git: %v", err)
	if err == nil {
		test.Errorf("broken .git fell back to the manifest")
	}
	// shallow clone answers through git
	shallow := filepath.Join(test.TempDir(), "s")
	g(test, repo, "clone", "-q", "--depth", "1", "file://"+repo, shallow)
	if _, _, _, err := selectFiles(shallow, "", []string{"r"}, []string{"*.md"}, nil); err != nil {
		test.Errorf("shallow: %v", err)
	}
}
