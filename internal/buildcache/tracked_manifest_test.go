package buildcache

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/toolchain"
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
		command := toolchain.GitCommand(append([]string{"-C", directory, "-c", "user.name=t", "-c", "user.email=t@t", "-c", "commit.gpgsign=false"}, arguments...)...)
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

// A runner's source values a clean file by the id its manifest's listing holds, unread, as a checkout values one by
// its index's (#szswjtc: a runner read and hashed every content entry of a read set, 8 to 27 s a lookup): a file no
// one may read keys, and keys and values as the checkout does. A file changed after the listing was written is hashed,
// never taken by the listing's id, which names the content it was unpacked with. A submodule's entry names a commit,
// not a file's content, so a file where the listing has a gitlink is hashed. Mutants: a file changed after the listing
// taken by the manifest's id (the listing's time taken as when it was read); the change time not asked (a file edited
// and its modification time set back takes the listing's id); the listing's submodule entries taken as files.
func TestAnUnpackedSourceValuesACleanFileByItsManifest(t *testing.T) {
	t.Parallel()
	tree := checkout(t)
	listing, err := toolchain.GitCommand("-C", tree, "ls-tree", "-r", "-z", "--full-tree", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	clean := key(t, tree, declared)
	// runner is declaredTree unpacked an hour ago, its listing (with extra entries after the tree's) a minute ago.
	written := time.Now().Add(-time.Minute)
	runner := func(t *testing.T, extra string) string {
		t.Helper()
		root := unpacked(t, nil)
		old := time.Now().Add(-time.Hour)
		for name := range declaredTree {
			if err := os.Chtimes(filepath.Join(root, name), old, old); err != nil {
				t.Fatal(err)
			}
		}
		write(t, root, ".tracked/files", string(listing)+extra)
		if err := os.Chtimes(filepath.Join(root, ".tracked/files"), written, written); err != nil {
			t.Fatal(err)
		}
		return root
	}
	t.Run("clean", func(t *testing.T) {
		t.Parallel()
		root := runner(t, "")
		unreadable := filepath.Join(root, "source/a.c")
		if err := os.Chmod(unreadable, 0); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { os.Chmod(unreadable, 0o644) })
		if _, err := os.ReadFile(unreadable); err == nil {
			t.Skip("this user reads a file with no permissions, so a file valued unread can't be told from one read")
		}
		// The chmod set a.c's change time to now, and nothing sets one back: the listing is written after it, as Loom
		// writes it after the files it lists, past the racy margin.
		time.Sleep(1100 * time.Millisecond)
		if err := os.Chtimes(filepath.Join(root, ".tracked/files"), time.Now(), time.Now()); err != nil {
			t.Fatal(err)
		}
		got, err := Key(root, declared)
		if err != nil {
			t.Fatalf("keying a clean file of an unpacked source read it: %v", err)
		}
		if got != clean {
			t.Fatalf("an unpacked source valued by its manifest keys %s, its checkout %s", got, clean)
		}
		if value, want := readTreeValue(root, "content", "source/a.c"), readTreeValue(tree, "content", "source/a.c"); value != want {
			t.Fatalf("a read set values the unpacked source's a.c %q, the checkout's %q", value, want)
		}
	})
	t.Run("changed after the listing", func(t *testing.T) {
		t.Parallel()
		root := runner(t, "")
		changed := map[string]string{"source/a.c": "int Z;\n"}
		write(t, root, "source/a.c", changed["source/a.c"])
		// Changed half a minute after the listing was written: a lookup asking now finds it changed long before.
		since := written.Add(30 * time.Second)
		if err := os.Chtimes(filepath.Join(root, "source/a.c"), since, since); err != nil {
			t.Fatal(err)
		}
		// Its change time is now; past the racy margin, it reads old to a lookup that took the listing's time as when
		// it was read, as the mutant does.
		time.Sleep(1100 * time.Millisecond)
		if got, want := key(t, root, declared), key(t, unpacked(t, changed), declared); got != want || got == clean {
			t.Fatalf("a file changed after the listing keys %s: its new content keys %s, the listing's id %s", got, want, clean)
		}
	})
	t.Run("changed after the listing, its time set back", func(t *testing.T) {
		t.Parallel()
		root := runner(t, "")
		changed := map[string]string{"source/a.c": "int Z;\n"}
		write(t, root, "source/a.c", changed["source/a.c"])
		// Edited, then its modification time set back an hour before the listing (Operations' scout): only its change
		// time, now, says it moved.
		old := written.Add(-time.Hour)
		if err := os.Chtimes(filepath.Join(root, "source/a.c"), old, old); err != nil {
			t.Fatal(err)
		}
		if got, want := key(t, root, declared), key(t, unpacked(t, changed), declared); got != want || got == clean {
			t.Fatalf("a file changed after the listing, its time set back, keys %s: its new content keys %s, the listing's id %s", got, want, clean)
		}
	})
	t.Run("a gitlink's path", func(t *testing.T) {
		t.Parallel()
		root := runner(t, "160000 commit "+strings.Repeat("ab", 20)+"\tvendor\x00")
		write(t, root, "vendor", "a file where the commit has a submodule\n")
		old := time.Now().Add(-time.Hour)
		if err := os.Chtimes(filepath.Join(root, "vendor"), old, old); err != nil {
			t.Fatal(err)
		}
		if _, held := repositoryAt(root).objects["vendor"]; held {
			t.Fatal("the listing's submodule entry is held as a file's id")
		}
		if value, want := readTreeValue(root, "content", "vendor"), "file false "+blobID([]byte("a file where the commit has a submodule\n"), 40); value != want {
			t.Fatalf("a file at a gitlink's path values %q, its content %q", value, want)
		}
	})
}
