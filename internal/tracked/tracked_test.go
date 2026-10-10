package tracked

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// runGit runs git in directory with no user or system configuration, so a fixture is the same on every machine.
func runGit(t *testing.T, directory string, arguments ...string) string {
	t.Helper()
	command := exec.Command("git", append([]string{"-C", directory, "-c", "user.name=Tracked fixture", "-c", "user.email=tracked@example.invalid",
		"-c", "commit.gpgsign=false", "-c", "init.defaultBranch=main", "-c", "protocol.file.allow=always"}, arguments...)...)
	command.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s: %v\n%s", arguments, directory, err, output)
	}
	return string(output)
}

func write(t *testing.T, name, text string, mode os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(name, []byte(text), mode); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(name, mode); err != nil {
		t.Fatal(err)
	}
}

func repository(t *testing.T, files map[string]string) string {
	t.Helper()
	directory := t.TempDir()
	runGit(t, directory, "init", "-q")
	for name, text := range files {
		mode := os.FileMode(0o644)
		if strings.HasSuffix(name, ".sh") {
			mode = 0o755
		}
		write(t, filepath.Join(directory, filepath.FromSlash(name)), text, mode)
	}
	runGit(t, directory, "add", ".")
	runGit(t, directory, "commit", "-qm", "fixture")
	return directory
}

// checkout is adamic's shape in small: a tree with a submodule at cohere, which has its own at TypeScript. Names that
// sort beside the gitlink (cohere-x/, cohere.json) put git's order apart from a sorted listing: git prints a
// submodule's files where its gitlink sorts, before cohere-x/, and a byte sort puts cohere/ after cohere.json. In a
// tree object a directory sorts as if it ended in "/", so a.b.md comes before a/ there and after it in a byte sort.
func checkout(t *testing.T) string {
	t.Helper()
	leaf := repository(t, map[string]string{"src/compiler/a.ts": "a;\n", "README.md": "# leaf\n"})
	middle := repository(t, map[string]string{"internal/x.go": "package x\n", "TypeScript-shim/y.md": "y\n"})
	runGit(t, middle, "submodule", "add", "-q", leaf, "TypeScript")
	runGit(t, middle, "commit", "-qm", "TypeScript")
	top := repository(t, map[string]string{"cohere-x/file.ts": "x;\n", "cohere.json": "{}\n", "a/b.md": "b\n", "a.b.md": "beside a/\n", "tool.sh": "#!/bin/sh\n"})
	if err := os.Symlink("a/b.md", filepath.Join(top, "link.md")); err != nil {
		t.Fatal(err)
	}
	runGit(t, top, "add", "link.md")
	runGit(t, top, "submodule", "add", "-q", middle, "cohere")
	runGit(t, top, "submodule", "update", "-q", "--init", "--recursive")
	runGit(t, top, "commit", "-qm", "cohere")
	return top
}

// unpack is the source as a runner holds it: every tracked file, submodules' included, with its link or executable
// bit, and no .git anywhere; plus the manifest build-tree writes, when manifest is set.
func unpack(t *testing.T, checkout string, manifest bool) string {
	t.Helper()
	directory := t.TempDir()
	for _, name := range strings.Split(strings.TrimSuffix(runGit(t, checkout, "ls-files", "--recurse-submodules", "-z"), "\x00"), "\x00") {
		source, destination := filepath.Join(checkout, filepath.FromSlash(name)), filepath.Join(directory, filepath.FromSlash(name))
		info, err := os.Lstat(source)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
			t.Fatal(err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(source)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(target, destination); err != nil {
				t.Fatal(err)
			}
			continue
		}
		data, err := os.ReadFile(source)
		if err != nil {
			t.Fatal(err)
		}
		write(t, destination, string(data), info.Mode().Perm())
	}
	if manifest {
		if err := Write(checkout, directory); err != nil {
			t.Fatal(err)
		}
	}
	return directory
}

// A source with no .git answers from its manifest exactly what git answers in the checkout it came from: each
// repository's commit and entries, the recursive listing in git's own order, and every file as recorded.
func TestManifestAnswersWhatGitAnswers(t *testing.T) {
	t.Parallel()
	checkout := checkout(t)
	source := unpack(t, checkout, true)
	for _, repository := range []string{".", "cohere", "cohere/TypeScript"} {
		if Git(filepath.Join(source, repository)) || !Git(filepath.Join(checkout, repository)) {
			t.Fatalf("%s: the checkout must have .git and the source none", repository)
		}
		gitHead, err := Head(filepath.Join(checkout, repository))
		if err != nil {
			t.Fatal(err)
		}
		head, err := Head(filepath.Join(source, repository))
		if err != nil || head != gitHead {
			t.Fatalf("%s: HEAD %s (%v), git says %s", repository, head, err, gitHead)
		}
		gitFiles, err := Files(filepath.Join(checkout, repository))
		if err != nil {
			t.Fatal(err)
		}
		files, err := Files(filepath.Join(source, repository))
		if err != nil || !reflect.DeepEqual(files, gitFiles) {
			t.Fatalf("%s: files %v (%v), git says %v", repository, files, err, gitFiles)
		}
		missing, changed, err := Changed(filepath.Join(source, repository), files)
		if err != nil || len(missing)+len(changed) != 0 {
			t.Fatalf("%s: an unpacked source reads as missing %q, changed %q (%v)", repository, missing, changed, err)
		}
	}
	gitRecursive, err := Recursive(checkout)
	if err != nil {
		t.Fatal(err)
	}
	recursive, err := Recursive(source)
	if err != nil || !reflect.DeepEqual(recursive, gitRecursive) {
		t.Fatalf("recursive listing %q (%v), git says %q", recursive, err, gitRecursive)
	}
	if sort.StringsAreSorted(recursive) || len(recursive) != 12 {
		t.Fatalf("the fixture no longer tells git's order from a sorted one: %q", recursive)
	}
	t.Logf("recursive, in git's order: %q", recursive)
}

// Changed names each way a file can differ from its record: content of the same size, the executable bit, a link
// made a file, and a file gone.
func TestChangedNamesEveryDifference(t *testing.T) {
	t.Parallel()
	source := unpack(t, checkout(t), true)
	files, err := Files(source)
	if err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(source, "a/b.md"), "c\n", 0o644)
	if err := os.Chmod(filepath.Join(source, "tool.sh"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(source, "link.md")); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(source, "link.md"), "a/b.md", 0o644)
	if err := os.Remove(filepath.Join(source, "cohere.json")); err != nil {
		t.Fatal(err)
	}
	missing, changed, err := Changed(source, files)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(missing, []string{"cohere.json"}) || !reflect.DeepEqual(changed, []string{"a/b.md", "link.md", "tool.sh"}) {
		t.Fatalf("missing %q, changed %q", missing, changed)
	}
}

// Where neither git nor the manifest can answer, every question fails, naming the fact that is missing.
func TestNoGitAndNoManifestFailsByName(t *testing.T) {
	t.Parallel()
	checkout := checkout(t)
	source := unpack(t, checkout, false)
	if _, err := Head(filepath.Join(source, "cohere")); err == nil || !strings.Contains(err.Error(), ManifestDirectory) || !strings.Contains(err.Error(), "commit") {
		t.Fatalf("no manifest: %v", err)
	}
	if _, err := Files(source); err == nil || !strings.Contains(err.Error(), "tracked files") {
		t.Fatalf("no manifest: %v", err)
	}
	if _, err := Recursive(source); err == nil || !strings.Contains(err.Error(), "no .git") {
		t.Fatalf("no manifest: %v", err)
	}

	lost := unpack(t, checkout, true)
	if err := os.Remove(filepath.Join(lost, ManifestDirectory, "cohere", "files")); err != nil {
		t.Fatal(err)
	}
	if _, err := Files(filepath.Join(lost, "cohere")); err == nil || !strings.Contains(err.Error(), "no tracked files for it") {
		t.Fatalf("a repository's files lost from the manifest: %v", err)
	}
	if _, err := Recursive(lost); err == nil || !strings.Contains(err.Error(), "no tracked files for it") {
		t.Fatalf("a submodule's files lost from the manifest: %v", err)
	}
}

// record replaces repository's record in source with the one Write makes of directory as it stands.
func record(t *testing.T, directory, source, repository string) {
	t.Helper()
	scratch := t.TempDir()
	if err := Write(directory, scratch); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{headFile, commitFile, filesFile} {
		data, err := os.ReadFile(filepath.Join(scratch, ManifestDirectory, name))
		if err != nil {
			t.Fatal(err)
		}
		write(t, filepath.Join(source, ManifestDirectory, filepath.FromSlash(repository), name), string(data), 0o644)
	}
}

// A submodule's recorded commit must be the one its parent's gitlink pins: a record of another commit, whole and
// bound, is refused.
func TestManifestCommitMustMatchTheParentGitlink(t *testing.T) {
	t.Parallel()
	checkout := checkout(t)
	source := unpack(t, checkout, true)
	leaf := filepath.Join(checkout, "cohere", "TypeScript")
	write(t, filepath.Join(leaf, "README.md"), "# moved on\n", 0o644)
	runGit(t, leaf, "commit", "-qam", "unpinned")
	record(t, leaf, source, "cohere/TypeScript")
	if _, err := Head(filepath.Join(source, "cohere", "TypeScript")); err == nil || !strings.Contains(err.Error(), "gitlink") {
		t.Fatalf("a commit its parent doesn't pin: %v", err)
	}
	if _, err := Head(filepath.Join(source, "cohere")); err != nil {
		t.Fatalf("the parent's own record still answers: %v", err)
	}
}

// Where .git is there, git answers, whatever manifest lies beside it.
func TestGitAnswersOverAManifest(t *testing.T) {
	t.Parallel()
	checkout := checkout(t)
	want := strings.TrimSpace(runGit(t, checkout, "rev-parse", "HEAD"))
	write(t, filepath.Join(checkout, ManifestDirectory, "HEAD"), strings.Repeat("2", 40)+"\n", 0o644)
	write(t, filepath.Join(checkout, ManifestDirectory, "files"), "", 0o644)
	if head, err := Head(checkout); err != nil || head != want {
		t.Fatalf("HEAD %s (%v), git says %s", head, err, want)
	}
	if files, err := Files(checkout); err != nil || len(files) != 8 {
		t.Fatalf("files %v (%v)", files, err)
	}
}

// A record's three files are one commit's: the commit object must hash to HEAD, and the tree it names must be the
// tree the listing rebuilds. A listing from another commit, a commit object edited, or one missing, is refused.
func TestManifestFilesAreBoundToItsCommit(t *testing.T) {
	t.Parallel()
	checkout := checkout(t)
	leaf := filepath.Join(checkout, "cohere", "TypeScript")
	older := t.TempDir()
	if err := Write(leaf, older); err != nil {
		t.Fatal(err)
	}
	// The tree and cohere as checked out, TypeScript as Write recorded it before any later commit: bound.
	prepare := func() string {
		source := unpack(t, checkout, false)
		record(t, checkout, source, "")
		record(t, filepath.Join(checkout, "cohere"), source, "cohere")
		for _, file := range []string{headFile, commitFile, filesFile} {
			data, err := os.ReadFile(filepath.Join(older, ManifestDirectory, file))
			if err != nil {
				t.Fatal(err)
			}
			write(t, filepath.Join(source, ManifestDirectory, "cohere", "TypeScript", file), string(data), 0o644)
		}
		return source
	}
	for name, plant := range map[string]func(source string){
		"files from another commit": func(source string) {
			write(t, filepath.Join(leaf, "src/compiler/b.ts"), "b;\n", 0o644)
			runGit(t, leaf, "add", ".")
			runGit(t, leaf, "commit", "-qm", "later")
			later := t.TempDir()
			if err := Write(leaf, later); err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(filepath.Join(later, ManifestDirectory, filesFile))
			if err != nil {
				t.Fatal(err)
			}
			write(t, filepath.Join(source, ManifestDirectory, "cohere", "TypeScript", filesFile), string(data), 0o644)
		},
		"commit object edited": func(source string) {
			name := filepath.Join(source, ManifestDirectory, "cohere", "TypeScript", commitFile)
			data, err := os.ReadFile(name)
			if err != nil {
				t.Fatal(err)
			}
			write(t, name, strings.Replace(string(data), "fixture", "fixturf", 1), 0o644)
		},
		"commit object missing": func(source string) {
			if err := os.Remove(filepath.Join(source, ManifestDirectory, "cohere", "TypeScript", commitFile)); err != nil {
				t.Fatal(err)
			}
		},
	} {
		source := prepare()
		if _, err := Files(filepath.Join(source, "cohere", "TypeScript")); err != nil {
			t.Fatalf("%s: the record before the fault: %v", name, err)
		}
		source = prepare()
		plant(source)
		_, err := Files(filepath.Join(source, "cohere", "TypeScript"))
		want := map[string]string{"files from another commit": "aren't that commit's", "commit object edited": "hashes to", "commit object missing": "no commit object"}[name]
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Fatalf("%s: %v", name, err)
		}
		t.Logf("%s: %v", name, err)
	}
}

// git's rule for the executable bit is the owner's: a group or other x alone changes nothing, and a 100755 file its
// owner can't execute has changed mode.
func TestExecutableBitIsTheOwners(t *testing.T) {
	t.Parallel()
	source := unpack(t, checkout(t), true)
	files, err := Files(source)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filepath.Join(source, "a/b.md"), 0o654); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filepath.Join(source, "tool.sh"), 0o645); err != nil {
		t.Fatal(err)
	}
	missing, changed, err := Changed(source, files)
	if err != nil || len(missing) != 0 || !reflect.DeepEqual(changed, []string{"tool.sh"}) {
		t.Fatalf("missing %q, changed %q (%v)", missing, changed, err)
	}
}

// A file the record leaves out is a stale manifest: Unrecorded names it, and Recursive refuses rather than list the
// tree without it. What the tree's own .gitignore files ignore (a provisioned node_modules, a log) is no file of the
// source, as git status has it, and a submodule's files are its own record's. Where git answers, it is nil.
func TestUnrecordedFilesAreNamed(t *testing.T) {
	t.Parallel()
	checkout := checkout(t)
	write(t, filepath.Join(checkout, ".gitignore"), "node_modules/\n*.log\n", 0o644)
	runGit(t, checkout, "add", ".gitignore")
	runGit(t, checkout, "commit", "-qm", "ignore rules")
	source := unpack(t, checkout, true)
	write(t, filepath.Join(source, "a", "node_modules", "pkg", "README.md"), "provisioned\n", 0o644)
	write(t, filepath.Join(source, "run.log"), "log\n", 0o644)
	if extra, err := Unrecorded(source, "."); err != nil || len(extra) != 0 {
		t.Fatalf("ignored leftovers named: %q (%v)", extra, err)
	}
	if _, err := Recursive(source); err != nil {
		t.Fatalf("ignored leftovers refused: %v", err)
	}
	write(t, filepath.Join(source, "a", "new", "c.md"), "added since\n", 0o644)
	write(t, filepath.Join(source, "cohere", "internal", "y.go"), "package x\n", 0o644)
	if extra, err := Unrecorded(source, "a"); err != nil || !reflect.DeepEqual(extra, []string{"a/new/c.md"}) {
		t.Fatalf("unrecorded under a: %q (%v)", extra, err)
	}
	if extra, err := Unrecorded(filepath.Join(source, "cohere"), "."); err != nil || !reflect.DeepEqual(extra, []string{"internal/y.go"}) {
		t.Fatalf("unrecorded in cohere: %q (%v)", extra, err)
	}
	if _, err := Recursive(source); err == nil || !strings.Contains(err.Error(), "internal/y.go") {
		t.Fatalf("a stale submodule record listed: %v", err)
	}
	if err := os.Remove(filepath.Join(source, "cohere", "internal", "y.go")); err != nil {
		t.Fatal(err)
	}
	if _, err := Recursive(source); err == nil || !strings.Contains(err.Error(), "a/new/c.md") {
		t.Fatalf("a stale record listed: %v", err)
	}
	write(t, filepath.Join(checkout, "a", "new", "c.md"), "untracked in a checkout\n", 0o644)
	if extra, err := Unrecorded(checkout, "."); err != nil || extra != nil {
		t.Fatalf("git's own checkout: %q (%v)", extra, err)
	}
}
