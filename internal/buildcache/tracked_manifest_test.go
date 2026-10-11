package buildcache

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// A Loom runner's source has no .git: build-tree unpacks the tree's tracked files from chunks and ships git's answers
// in a .tracked manifest beside them, with whatever else the source carries (npm packages, the manifest itself).
// A directory's key there is its key in the checkout Workshop built it from, because the manifest, not the directory
// listing, says which files are the tree's (#59paqn3). Without the manifest the untracked file enters the key, which
// is the miss Loom's key check found.
// Not parallel: forgetTracked clears the process's repository memo.
func TestAKeyFromARunnersSourceIsTheCheckoutsKey(t *testing.T) {
	git := func(directory string, arguments ...string) []byte {
		t.Helper()
		command := exec.Command("git", append([]string{"-C", directory, "-c", "user.name=t", "-c", "user.email=t@t", "-c", "commit.gpgsign=false"}, arguments...)...)
		output, err := command.Output()
		if err != nil {
			t.Fatalf("git %v: %v", arguments, err)
		}
		return output
	}
	checkout, runner := t.TempDir(), t.TempDir()
	for _, root := range []string{checkout, runner} {
		write(t, root, "go.mod", "module example\n")
		write(t, root, "data/answer.txt", "forty-two\n")
		write(t, root, "data/nested/deep.txt", "deep\n")
		// Untracked in the checkout, shipped in the runner's source: an npm package beside the tree's files.
		write(t, root, "data/node_modules/package/index.js", "module.exports = 1;\n")
	}
	git(checkout, "init", "-q")
	git(checkout, "add", "go.mod", "data/answer.txt", "data/nested/deep.txt")
	git(checkout, "commit", "-q", "-m", "tree")
	write(t, runner, ".tracked/HEAD", string(git(checkout, "rev-parse", "HEAD")))
	write(t, runner, ".tracked/commit", string(git(checkout, "cat-file", "commit", "HEAD")))
	write(t, runner, ".tracked/files", string(git(checkout, "ls-tree", "-r", "-z", "--full-tree", "HEAD")))
	inputs := Inputs{Name: "runner source", Files: []string{"data"}}

	forgetTracked()
	want := key(t, checkout, inputs)
	forgetTracked()
	if got := key(t, runner, inputs); got != want {
		t.Fatalf("the runner's source keys data %s, the checkout %s", got, want)
	}
	if Tracked(runner, filepath.Join(runner, "data/node_modules/package/index.js"), false) {
		t.Error("an untracked shipped file counts as the tree's")
	}
	if Tracked(runner, filepath.Join(runner, ".tracked/files"), false) {
		t.Error("the manifest counts as the tree's")
	}
	if !Tracked(runner, filepath.Join(runner, "data/nested/deep.txt"), false) {
		t.Error("a tracked file doesn't count as the tree's")
	}

	// The check can fail: without the manifest the shipped package is keyed, and the key moves.
	if err := os.Rename(filepath.Join(runner, ".tracked"), filepath.Join(t.TempDir(), "moved")); err != nil {
		t.Fatal(err)
	}
	forgetTracked()
	if key(t, runner, inputs) == want {
		t.Fatal("with no manifest the runner's source still keys as the checkout; the untracked package isn't reaching the key")
	}
	forgetTracked()
}
