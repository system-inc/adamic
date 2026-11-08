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
func TestAKeyFromARunnersSourceIsTheCheckoutsKey(test *testing.T) {
	git := func(directory string, arguments ...string) []byte {
		test.Helper()
		command := toolchain.GitCommand(append(
			[]string{"-C", directory, "-c", "user.name=t", "-c", "user.email=t@t", "-c", "commit.gpgsign=false"},
			arguments...,
		)...)
		output, outputError := command.Output()
		if outputError != nil {
			test.Fatalf("git %v: %v", arguments, outputError)
		}
		return output
	}
	checkout, runner := test.TempDir(), test.TempDir()
	for _, root := range []string{checkout, runner} {
		write(test, root, "go.mod", "module example\n")
		write(test, root, "data/answer.txt", "forty-two\n")
		write(test, root, "data/nested/deep.txt", "deep\n")
		// Untracked in the checkout, shipped in the runner's source: an npm package beside the tree's files.
		write(test, root, "data/node_modules/package/index.js", "module.exports = 1;\n")
	}
	git(checkout, "init", "-q")
	git(checkout, "add", "go.mod", "data/answer.txt", "data/nested/deep.txt")
	git(checkout, "commit", "-q", "-m", "tree")
	write(test, runner, ".tracked/HEAD", string(git(checkout, "rev-parse", "HEAD")))
	write(test, runner, ".tracked/commit", string(git(checkout, "cat-file", "commit", "HEAD")))
	write(test, runner, ".tracked/files", string(git(checkout, "ls-tree", "-r", "-z", "--full-tree", "HEAD")))
	inputs := Inputs{Name: "runner source", Files: []string{"data"}}

	forgetTracked()
	want := key(test, checkout, inputs)
	forgetTracked()
	if got := key(test, runner, inputs); got != want {
		test.Fatalf("the runner's source keys data %s, the checkout %s", got, want)
	}
	if Tracked(runner, filepath.Join(runner, "data/node_modules/package/index.js"), false) {
		test.Error("an untracked shipped file counts as the tree's")
	}
	if Tracked(runner, filepath.Join(runner, ".tracked/files"), false) {
		test.Error("the manifest counts as the tree's")
	}
	if !Tracked(runner, filepath.Join(runner, "data/nested/deep.txt"), false) {
		test.Error("a tracked file doesn't count as the tree's")
	}

	// The check can fail: without the manifest the shipped package is keyed, and the key moves.
	if renameError := os.Rename(
		filepath.Join(runner, ".tracked"),
		filepath.Join(test.TempDir(), "moved"),
	); renameError != nil {
		test.Fatal(renameError)
	}
	forgetTracked()
	if key(test, runner, inputs) == want {
		test.Fatal(
			"with no manifest the runner's source still keys as the checkout; the untracked package isn't reaching the key",
		)
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
func TestAnUnpackedSourceValuesACleanFileByItsManifest(test *testing.T) {
	test.Parallel()
	tree := checkout(test)
	listing, outputError := toolchain.GitCommand("-C", tree, "ls-tree", "-r", "-z", "--full-tree", "HEAD").Output()
	if outputError != nil {
		test.Fatal(outputError)
	}
	clean := key(test, tree, declared)
	// runner is declaredTree unpacked an hour ago, its listing (with extra entries after the tree's) a minute ago.
	written := time.Now().Add(-time.Minute)
	runner := func(test *testing.T, extra string) string {
		test.Helper()
		root := unpacked(test, nil)
		old := time.Now().Add(-time.Hour)
		for name := range declaredTree {
			if changeTimesError := os.Chtimes(filepath.Join(root, name), old, old); changeTimesError != nil {
				test.Fatal(changeTimesError)
			}
		}
		write(test, root, ".tracked/files", string(listing)+extra)
		if changeTimesError := os.Chtimes(
			filepath.Join(root, ".tracked/files"),
			written,
			written,
		); changeTimesError != nil {
			test.Fatal(changeTimesError)
		}
		return root
	}
	test.Run("clean", func(test *testing.T) {
		test.Parallel()
		root := runner(test, "")
		unreadable := filepath.Join(root, "source/a.c")
		if changeModeError := os.Chmod(unreadable, 0); changeModeError != nil {
			test.Fatal(changeModeError)
		}
		test.Cleanup(func() { os.Chmod(unreadable, 0o644) })
		if _, readError := os.ReadFile(unreadable); readError == nil {
			test.Skip("this user reads a file with no permissions, so a file valued unread can't be told from one read")
		}
		// The chmod set a.c's change time to now, and nothing sets one back: the listing is written after it, as Loom
		// writes it after the files it lists, past the racy margin.
		time.Sleep(1100 * time.Millisecond)
		if changeTimesError := os.Chtimes(
			filepath.Join(root, ".tracked/files"),
			time.Now(),
			time.Now(),
		); changeTimesError != nil {
			test.Fatal(changeTimesError)
		}
		got, keyError := Key(root, declared)
		if keyError != nil {
			test.Fatalf("keying a clean file of an unpacked source read it: %v", keyError)
		}
		if got != clean {
			test.Fatalf("an unpacked source valued by its manifest keys %s, its checkout %s", got, clean)
		}
		if value, want := readTreeValue(
			root,
			"content",
			"source/a.c",
		), readTreeValue(tree, "content", "source/a.c"); value != want {
			test.Fatalf("a read set values the unpacked source's a.c %q, the checkout's %q", value, want)
		}
	})
	test.Run("changed after the listing", func(test *testing.T) {
		test.Parallel()
		root := runner(test, "")
		changed := map[string]string{"source/a.c": "int Z;\n"}
		write(test, root, "source/a.c", changed["source/a.c"])
		// Changed half a minute after the listing was written: a lookup asking now finds it changed long before.
		since := written.Add(30 * time.Second)
		if changeTimesError := os.Chtimes(filepath.Join(root, "source/a.c"), since, since); changeTimesError != nil {
			test.Fatal(changeTimesError)
		}
		// Its change time is now; past the racy margin, it reads old to a lookup that took the listing's time as when
		// it was read, as the mutant does.
		time.Sleep(1100 * time.Millisecond)
		if got, want := key(
			test,
			root,
			declared,
		), key(test, unpacked(test, changed), declared); got != want || got == clean {
			test.Fatalf(
				"a file changed after the listing keys %s: its new content keys %s, the listing's id %s",
				got,
				want,
				clean,
			)
		}
	})
	test.Run("changed after the listing, its time set back", func(test *testing.T) {
		test.Parallel()
		root := runner(test, "")
		changed := map[string]string{"source/a.c": "int Z;\n"}
		write(test, root, "source/a.c", changed["source/a.c"])
		// Edited, then its modification time set back an hour before the listing (Operations' scout): only its change
		// time, now, says it moved.
		old := written.Add(-time.Hour)
		if changeTimesError := os.Chtimes(filepath.Join(root, "source/a.c"), old, old); changeTimesError != nil {
			test.Fatal(changeTimesError)
		}
		if got, want := key(
			test,
			root,
			declared,
		), key(test, unpacked(test, changed), declared); got != want || got == clean {
			test.Fatalf(
				"a file changed after the listing, its time set back, keys %s: its new content keys %s, the listing's id %s",
				got,
				want,
				clean,
			)
		}
	})
	test.Run("a gitlink's path", func(test *testing.T) {
		test.Parallel()
		root := runner(test, "160000 commit "+strings.Repeat("ab", 20)+"\tvendor\x00")
		write(test, root, "vendor", "a file where the commit has a submodule\n")
		old := time.Now().Add(-time.Hour)
		if changeTimesError := os.Chtimes(filepath.Join(root, "vendor"), old, old); changeTimesError != nil {
			test.Fatal(changeTimesError)
		}
		if _, held := repositoryAt(root).objects["vendor"]; held {
			test.Fatal("the listing's submodule entry is held as a file's id")
		}
		if value, want := readTreeValue(
			root,
			"content",
			"vendor",
		), "file false "+blobId([]byte("a file where the commit has a submodule\n"), 40); value != want {
			test.Fatalf("a file at a gitlink's path values %q, its content %q", value, want)
		}
	})
}

// An unpacked source's manifest is read as keying, as a checkout's .git is: a traced product test from a source with
// no .git never records .tracked in its read set, where it is refused as untracked and the test falls back untraced
// (#asrz3cy). The look for the manifest at each parent and the read of its listing both count. Mutant: manifestListing
// called outside keying.
// Not parallel: test.Setenv, statManifest, openManifest and forgetTracked.
func TestAnUnpackedSourcesManifestIsReadAsKeying(test *testing.T) {
	test.Setenv("ADAMIC_BUILD_TRACE", test.TempDir())
	root := unpacked(test, nil)
	listing := "100644 blob " + blobId([]byte(declaredTree["flags.txt"]), 40) + "\tflags.txt\x00"
	write(test, root, ".tracked/files", listing)
	var looks, unkeyed []string
	savedStat, savedOpen := statManifest, openManifest
	statManifest = func(name string) (os.FileInfo, error) {
		looks = append(looks, name)
		if keyingNow.Load() == 0 {
			unkeyed = append(unkeyed, name)
		}
		return savedStat(name)
	}
	openManifest = func(name string) (*os.File, error) {
		looks = append(looks, name)
		if keyingNow.Load() == 0 {
			unkeyed = append(unkeyed, name)
		}
		return savedOpen(name)
	}
	test.Cleanup(func() { statManifest, openManifest = savedStat, savedOpen })
	forgetTracked()
	test.Cleanup(forgetTracked)
	files := repositoryAt(root)
	if files == nil || !files.paths["flags.txt"] || files.paths["source/a.c"] {
		test.Fatalf("the manifest's listing isn't the source's tracked files: %+v", files)
	}
	// A directory under the root the manifest doesn't record looks for it at each parent, and isn't a repository.
	if repositoryAt(filepath.Join(root, "source", "nested")) != nil {
		test.Fatal("a directory the manifest doesn't record reads as a repository")
	}
	if len(looks) < 5 || len(unkeyed) != 0 {
		test.Fatalf("read the manifest at %q; outside keying at %q", looks, unkeyed)
	}
}
