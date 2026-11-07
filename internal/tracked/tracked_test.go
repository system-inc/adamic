package tracked

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/toolchain"
)

// runGit runs git in directory with no user or system configuration (toolchain's git environment), so a fixture is the
// same on every machine.
func runGit(test *testing.T, directory string, arguments ...string) string {
	test.Helper()
	command := toolchain.GitCommand(
		append(
			[]string{"-C", directory, "-c", "user.name=Tracked fixture", "-c", "user.email=tracked@example.invalid",
				"-c", "commit.gpgsign=false", "-c", "init.defaultBranch=main", "-c", "protocol.file.allow=always"},
			arguments...,
		)...,
	)
	output, err := command.CombinedOutput()
	if err != nil {
		test.Fatalf("git %v in %s: %v\n%s", arguments, directory, err, output)
	}
	return string(output)
}

func write(test *testing.T, name, text string, mode os.FileMode) {
	test.Helper()
	if err := os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
		test.Fatal(err)
	}
	if err := os.WriteFile(name, []byte(text), mode); err != nil {
		test.Fatal(err)
	}
	if err := os.Chmod(name, mode); err != nil {
		test.Fatal(err)
	}
}

func repository(test *testing.T, files map[string]string) string {
	test.Helper()
	directory := test.TempDir()
	runGit(test, directory, "init", "-q")
	for name, text := range files {
		mode := os.FileMode(0o644)
		if strings.HasSuffix(name, ".sh") {
			mode = 0o755
		}
		write(test, filepath.Join(directory, filepath.FromSlash(name)), text, mode)
	}
	runGit(test, directory, "add", ".")
	runGit(test, directory, "commit", "-qm", "fixture")
	return directory
}

// checkout is adamic's shape in small: a tree with a submodule at cohere, which has its own at TypeScript. Names that
// sort beside the gitlink (cohere-x/, cohere.json) put git's order apart from a sorted listing: git prints a
// submodule's files where its gitlink sorts, before cohere-x/, and a byte sort puts cohere/ after cohere.json. In a
// tree object a directory sorts as if it ended in "/", so a.b.md comes before a/ there and after it in a byte sort.
func checkout(test *testing.T) string {
	test.Helper()
	leaf := repository(test, map[string]string{"src/compiler/a.ts": "a;\n", "README.md": "# leaf\n"})
	middle := repository(test, map[string]string{"internal/x.go": "package x\n", "TypeScript-shim/y.md": "y\n"})
	runGit(test, middle, "submodule", "add", "-q", leaf, "TypeScript")
	runGit(test, middle, "commit", "-qm", "TypeScript")
	top := repository(test, map[string]string{
		"cohere-x/file.ts": "x;\n",
		"cohere.json":      "{}\n",
		"a/b.md":           "b\n",
		"a.b.md":           "beside a/\n",
		"tool.sh":          "#!/bin/sh\n",
	})
	if err := os.Symlink("a/b.md", filepath.Join(top, "link.md")); err != nil {
		test.Fatal(err)
	}
	runGit(test, top, "add", "link.md")
	runGit(test, top, "submodule", "add", "-q", middle, "cohere")
	runGit(test, top, "submodule", "update", "-q", "--init", "--recursive")
	runGit(test, top, "commit", "-qm", "cohere")
	return top
}

// unpack is the source as a runner holds it: every tracked file, submodules' included, with its link or executable
// bit, and no .git anywhere; plus the manifest build-tree writes, when manifest is set.
func unpack(test *testing.T, checkout string, manifest bool) string {
	test.Helper()
	directory := test.TempDir()
	for _, name := range strings.Split(
		strings.TrimSuffix(runGit(test, checkout, "ls-files", "--recurse-submodules", "-z"), "\x00"),
		"\x00",
	) {
		source, destination := filepath.Join(
			checkout,
			filepath.FromSlash(name),
		), filepath.Join(directory, filepath.FromSlash(name))
		info, err := os.Lstat(source)
		if err != nil {
			test.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
			test.Fatal(err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(source)
			if err != nil {
				test.Fatal(err)
			}
			if err := os.Symlink(target, destination); err != nil {
				test.Fatal(err)
			}
			continue
		}
		data, err := os.ReadFile(source)
		if err != nil {
			test.Fatal(err)
		}
		write(test, destination, string(data), info.Mode().Perm())
	}
	if manifest {
		if err := Write(checkout, directory); err != nil {
			test.Fatal(err)
		}
	}
	return directory
}

// A source with no .git answers from its manifest exactly what git answers in the checkout it came from: each
// repository's commit and entries, the recursive listing in git's own order, and every file as recorded.
func TestManifestAnswersWhatGitAnswers(test *testing.T) {
	test.Parallel()
	checkout := checkout(test)
	source := unpack(test, checkout, true)
	for _, repository := range []string{".", "cohere", "cohere/TypeScript"} {
		if Git(filepath.Join(source, repository)) || !Git(filepath.Join(checkout, repository)) {
			test.Fatalf("%s: the checkout must have .git and the source none", repository)
		}
		gitHead, err := Head(filepath.Join(checkout, repository))
		if err != nil {
			test.Fatal(err)
		}
		head, err := Head(filepath.Join(source, repository))
		if err != nil || head != gitHead {
			test.Fatalf("%s: HEAD %s (%v), git says %s", repository, head, err, gitHead)
		}
		gitFiles, err := Files(filepath.Join(checkout, repository))
		if err != nil {
			test.Fatal(err)
		}
		files, err := Files(filepath.Join(source, repository))
		if err != nil || !reflect.DeepEqual(files, gitFiles) {
			test.Fatalf("%s: files %v (%v), git says %v", repository, files, err, gitFiles)
		}
		missing, changed, err := Changed(filepath.Join(source, repository), files)
		if err != nil || len(missing)+len(changed) != 0 {
			test.Fatalf(
				"%s: an unpacked source reads as missing %q, changed %q (%v)",
				repository,
				missing,
				changed,
				err,
			)
		}
	}
	gitRecursive, err := Recursive(checkout)
	if err != nil {
		test.Fatal(err)
	}
	recursive, err := Recursive(source)
	if err != nil || !reflect.DeepEqual(recursive, gitRecursive) {
		test.Fatalf("recursive listing %q (%v), git says %q", recursive, err, gitRecursive)
	}
	if sort.StringsAreSorted(recursive) || len(recursive) != 12 {
		test.Fatalf("the fixture no longer tells git's order from a sorted one: %q", recursive)
	}
	test.Logf("recursive, in git's order: %q", recursive)
}

// Changed names each way a file can differ from its record: content of the same size, the executable bit, a link
// made a file, and a file gone.
func TestChangedNamesEveryDifference(test *testing.T) {
	test.Parallel()
	source := unpack(test, checkout(test), true)
	files, err := Files(source)
	if err != nil {
		test.Fatal(err)
	}
	write(test, filepath.Join(source, "a/b.md"), "c\n", 0o644)
	if err := os.Chmod(filepath.Join(source, "tool.sh"), 0o644); err != nil {
		test.Fatal(err)
	}
	if err := os.Remove(filepath.Join(source, "link.md")); err != nil {
		test.Fatal(err)
	}
	write(test, filepath.Join(source, "link.md"), "a/b.md", 0o644)
	if err := os.Remove(filepath.Join(source, "cohere.json")); err != nil {
		test.Fatal(err)
	}
	missing, changed, err := Changed(source, files)
	if err != nil {
		test.Fatal(err)
	}
	if !reflect.DeepEqual(missing, []string{"cohere.json"}) ||
		!reflect.DeepEqual(changed, []string{"a/b.md", "link.md", "tool.sh"}) {
		test.Fatalf("missing %q, changed %q", missing, changed)
	}
}

// Where neither git nor the manifest can answer, every question fails, naming the fact that is missing.
func TestNoGitAndNoManifestFailsByName(test *testing.T) {
	test.Parallel()
	checkout := checkout(test)
	source := unpack(test, checkout, false)
	if _, err := Head(
		filepath.Join(source, "cohere"),
	); err == nil || !strings.Contains(err.Error(), ManifestDirectory) || !strings.Contains(err.Error(), "commit") {
		test.Fatalf("no manifest: %v", err)
	}
	if _, err := Files(source); err == nil || !strings.Contains(err.Error(), "tracked files") {
		test.Fatalf("no manifest: %v", err)
	}
	if _, err := Recursive(source); err == nil || !strings.Contains(err.Error(), "no .git") {
		test.Fatalf("no manifest: %v", err)
	}

	lost := unpack(test, checkout, true)
	if err := os.Remove(filepath.Join(lost, ManifestDirectory, "cohere", "files")); err != nil {
		test.Fatal(err)
	}
	if _, err := Files(
		filepath.Join(lost, "cohere"),
	); err == nil || !strings.Contains(err.Error(), "no tracked files for it") {
		test.Fatalf("a repository's files lost from the manifest: %v", err)
	}
	if _, err := Recursive(lost); err == nil || !strings.Contains(err.Error(), "no tracked files for it") {
		test.Fatalf("a submodule's files lost from the manifest: %v", err)
	}
}

// record replaces repository's record in source with the one Write makes of directory as it stands.
func record(test *testing.T, directory, source, repository string) {
	test.Helper()
	scratch := test.TempDir()
	if err := Write(directory, scratch); err != nil {
		test.Fatal(err)
	}
	for _, name := range []string{headFile, commitFile, filesFile} {
		data, err := os.ReadFile(filepath.Join(scratch, ManifestDirectory, name))
		if err != nil {
			test.Fatal(err)
		}
		write(test, filepath.Join(source, ManifestDirectory, filepath.FromSlash(repository), name), string(data), 0o644)
	}
}

// A submodule's recorded commit must be the one its parent's gitlink pins: a record of another commit, whole and
// bound, is refused.
func TestManifestCommitMustMatchTheParentGitlink(test *testing.T) {
	test.Parallel()
	checkout := checkout(test)
	source := unpack(test, checkout, true)
	leaf := filepath.Join(checkout, "cohere", "TypeScript")
	write(test, filepath.Join(leaf, "README.md"), "# moved on\n", 0o644)
	runGit(test, leaf, "commit", "-qam", "unpinned")
	record(test, leaf, source, "cohere/TypeScript")
	if _, err := Head(
		filepath.Join(source, "cohere", "TypeScript"),
	); err == nil || !strings.Contains(err.Error(), "gitlink") {
		test.Fatalf("a commit its parent doesn't pin: %v", err)
	}
	if _, err := Head(filepath.Join(source, "cohere")); err != nil {
		test.Fatalf("the parent's own record still answers: %v", err)
	}
}

// Where .git is there, git answers, whatever manifest lies beside it.
func TestGitAnswersOverAManifest(test *testing.T) {
	test.Parallel()
	checkout := checkout(test)
	want := strings.TrimSpace(runGit(test, checkout, "rev-parse", "HEAD"))
	write(test, filepath.Join(checkout, ManifestDirectory, "HEAD"), strings.Repeat("2", 40)+"\n", 0o644)
	write(test, filepath.Join(checkout, ManifestDirectory, "files"), "", 0o644)
	if head, err := Head(checkout); err != nil || head != want {
		test.Fatalf("HEAD %s (%v), git says %s", head, err, want)
	}
	if files, err := Files(checkout); err != nil || len(files) != 8 {
		test.Fatalf("files %v (%v)", files, err)
	}
}

// A record's three files are one commit's: the commit object must hash to HEAD, and the tree it names must be the
// tree the listing rebuilds. A listing from another commit, a commit object edited, or one missing, is refused.
func TestManifestFilesAreBoundToItsCommit(test *testing.T) {
	test.Parallel()
	checkout := checkout(test)
	leaf := filepath.Join(checkout, "cohere", "TypeScript")
	older := test.TempDir()
	if err := Write(leaf, older); err != nil {
		test.Fatal(err)
	}
	// The tree and cohere as checked out, TypeScript as Write recorded it before any later commit: bound.
	prepare := func() string {
		source := unpack(test, checkout, false)
		record(test, checkout, source, "")
		record(test, filepath.Join(checkout, "cohere"), source, "cohere")
		for _, file := range []string{headFile, commitFile, filesFile} {
			data, err := os.ReadFile(filepath.Join(older, ManifestDirectory, file))
			if err != nil {
				test.Fatal(err)
			}
			write(test, filepath.Join(source, ManifestDirectory, "cohere", "TypeScript", file), string(data), 0o644)
		}
		return source
	}
	for name, plant := range map[string]func(source string){
		"files from another commit": func(source string) {
			write(test, filepath.Join(leaf, "src/compiler/b.ts"), "b;\n", 0o644)
			runGit(test, leaf, "add", ".")
			runGit(test, leaf, "commit", "-qm", "later")
			later := test.TempDir()
			if err := Write(leaf, later); err != nil {
				test.Fatal(err)
			}
			data, err := os.ReadFile(filepath.Join(later, ManifestDirectory, filesFile))
			if err != nil {
				test.Fatal(err)
			}
			write(
				test,
				filepath.Join(source, ManifestDirectory, "cohere", "TypeScript", filesFile),
				string(data),
				0o644,
			)
		},
		"commit object edited": func(source string) {
			name := filepath.Join(source, ManifestDirectory, "cohere", "TypeScript", commitFile)
			data, err := os.ReadFile(name)
			if err != nil {
				test.Fatal(err)
			}
			write(test, name, strings.Replace(string(data), "fixture", "fixturf", 1), 0o644)
		},
		"commit object missing": func(source string) {
			if err := os.Remove(
				filepath.Join(source, ManifestDirectory, "cohere", "TypeScript", commitFile),
			); err != nil {
				test.Fatal(err)
			}
		},
	} {
		source := prepare()
		if _, err := Files(filepath.Join(source, "cohere", "TypeScript")); err != nil {
			test.Fatalf("%s: the record before the fault: %v", name, err)
		}
		source = prepare()
		plant(source)
		_, err := Files(filepath.Join(source, "cohere", "TypeScript"))
		want := map[string]string{
			"files from another commit": "aren't that commit's",
			"commit object edited":      "hashes to",
			"commit object missing":     "no commit object",
		}[name]
		if err == nil || !strings.Contains(err.Error(), want) {
			test.Fatalf("%s: %v", name, err)
		}
		test.Logf("%s: %v", name, err)
	}
}

// git's rule for the executable bit is the owner's: a group or other x alone changes nothing, and a 100755 file its
// owner can't execute has changed mode.
func TestExecutableBitIsTheOwners(test *testing.T) {
	test.Parallel()
	source := unpack(test, checkout(test), true)
	files, err := Files(source)
	if err != nil {
		test.Fatal(err)
	}
	if err := os.Chmod(filepath.Join(source, "a/b.md"), 0o654); err != nil {
		test.Fatal(err)
	}
	if err := os.Chmod(filepath.Join(source, "tool.sh"), 0o645); err != nil {
		test.Fatal(err)
	}
	missing, changed, err := Changed(source, files)
	if err != nil || len(missing) != 0 || !reflect.DeepEqual(changed, []string{"tool.sh"}) {
		test.Fatalf("missing %q, changed %q (%v)", missing, changed, err)
	}
}

// A file the record leaves out is a stale manifest: Unrecorded names it, and Recursive refuses rather than list the
// tree without it. What the tree's own .gitignore files ignore (a provisioned node_modules, a log) is no file of the
// source, as git status has it, and a submodule's files are its own record's. Where git answers, it is nil.
func TestUnrecordedFilesAreNamed(test *testing.T) {
	test.Parallel()
	checkout := checkout(test)
	write(test, filepath.Join(checkout, ".gitignore"), "node_modules/\n*.log\n", 0o644)
	runGit(test, checkout, "add", ".gitignore")
	runGit(test, checkout, "commit", "-qm", "ignore rules")
	source := unpack(test, checkout, true)
	write(test, filepath.Join(source, "a", "node_modules", "pkg", "README.md"), "provisioned\n", 0o644)
	write(test, filepath.Join(source, "run.log"), "log\n", 0o644)
	if extra, err := Unrecorded(source, "."); err != nil || len(extra) != 0 {
		test.Fatalf("ignored leftovers named: %q (%v)", extra, err)
	}
	if _, err := Recursive(source); err != nil {
		test.Fatalf("ignored leftovers refused: %v", err)
	}
	write(test, filepath.Join(source, "a", "new", "c.md"), "added since\n", 0o644)
	write(test, filepath.Join(source, "cohere", "internal", "y.go"), "package x\n", 0o644)
	if extra, err := Unrecorded(source, "a"); err != nil || !reflect.DeepEqual(extra, []string{"a/new/c.md"}) {
		test.Fatalf("unrecorded under a: %q (%v)", extra, err)
	}
	if extra, err := Unrecorded(
		filepath.Join(source, "cohere"),
		".",
	); err != nil || !reflect.DeepEqual(extra, []string{"internal/y.go"}) {
		test.Fatalf("unrecorded in cohere: %q (%v)", extra, err)
	}
	if _, err := Recursive(source); err == nil || !strings.Contains(err.Error(), "internal/y.go") {
		test.Fatalf("a stale submodule record listed: %v", err)
	}
	if err := os.Remove(filepath.Join(source, "cohere", "internal", "y.go")); err != nil {
		test.Fatal(err)
	}
	if _, err := Recursive(source); err == nil || !strings.Contains(err.Error(), "a/new/c.md") {
		test.Fatalf("a stale record listed: %v", err)
	}
	write(test, filepath.Join(checkout, "a", "new", "c.md"), "untracked in a checkout\n", 0o644)
	if extra, err := Unrecorded(checkout, "."); err != nil || extra != nil {
		test.Fatalf("git's own checkout: %q (%v)", extra, err)
	}
}

// An untracked name git would read as pathspec magic (":(glob)...", ":!...") is a name like any other: check-ignore
// takes it literally, so Unrecorded names it rather than failing on it, and an ignored one stays ignored.
func TestUnrecordedNamesArePathsNotPathspecs(test *testing.T) {
	test.Parallel()
	checkout := checkout(test)
	write(test, filepath.Join(checkout, ".gitignore"), "*.log\n", 0o644)
	runGit(test, checkout, "add", ".gitignore")
	runGit(test, checkout, "commit", "-qm", "ignore rules")
	source := unpack(test, checkout, true)
	// At the repository's top, where the walk hands them to git as they are.
	for _, name := range []string{":(glob)x.md", ":!y.md", ":(top)z.log"} {
		write(test, filepath.Join(source, name), "magic\n", 0o644)
	}
	extra, err := Unrecorded(source, ".")
	if err != nil || !reflect.DeepEqual(extra, []string{":!y.md", ":(glob)x.md"}) {
		test.Fatalf("names that look like pathspecs: %q (%v)", extra, err)
	}
}

// A caller's git variables never redirect what this package asks a checkout: a hook that runs the tests sets GIT_DIR,
// GIT_INDEX_FILE and GIT_WORK_TREE for its own repository, and Head, Files and Recursive answer as they do without
// them. Mutant: git run with the process's whole environment.
// Not parallel: test.Setenv.
func TestACallersGitVariablesNeverRedirectAnAnswer(test *testing.T) {
	top := checkout(test)
	answers := func() (string, []Entry, []string) {
		test.Helper()
		head, err := Head(top)
		if err != nil {
			test.Fatal(err)
		}
		files, err := Files(top)
		if err != nil {
			test.Fatal(err)
		}
		recursive, err := Recursive(top)
		if err != nil {
			test.Fatal(err)
		}
		return head, files, recursive
	}
	wantHead, wantFiles, wantRecursive := answers()
	other := repository(test, map[string]string{"other.txt": "other\n"})
	test.Setenv("GIT_DIR", filepath.Join(other, ".git"))
	test.Setenv("GIT_INDEX_FILE", filepath.Join(other, ".git", "index"))
	test.Setenv("GIT_WORK_TREE", other)
	gotHead, gotFiles, gotRecursive := answers()
	if gotHead != wantHead ||
		!reflect.DeepEqual(gotFiles, wantFiles) ||
		!reflect.DeepEqual(gotRecursive, wantRecursive) {
		test.Fatalf(
			"with another repository's GIT_DIR, tracked answered HEAD %s and %d files, %d recursive; without, %s, %d and %d",
			gotHead,
			len(gotFiles),
			len(gotRecursive),
			wantHead,
			len(wantFiles),
			len(wantRecursive),
		)
	}
}

// A directory skip names is left unentered below root, and nothing under it is listed, in git's answer and in the
// walk Workshop's trace takes alike; a file or directory that only shares the name, and root itself, stay (#hp8gq3e).
// Not parallel: it sets ADAMIC_BUILD_TRACE, which buildcache.Tracing reads for the whole process.
func TestPathsUnderLeavesSkippedDirectoriesUnentered(test *testing.T) {
	directory := repository(test, map[string]string{
		"stage1/a.ts": "a;\n", "stage1/evidence.ts": "named like the folder;\n", "stage1/evidence-based/b.ts": "b;\n",
		"stage1/rules/x/rule.ts": "rule;\n", "stage1/rules/x/evidence/log.txt": "evidence\n",
		"stage1/rules/x/testdata/witness.ts": "witness;\n", "stage1/testdata/deep/fixture.ts": "fixture;\n",
		"other/evidence/kept.ts": "outside root;\n",
	})
	want := []string{"stage1/a.ts", "stage1/evidence-based/b.ts", "stage1/evidence.ts", "stage1/rules/x/rule.ts"}
	listed, err := PathsUnder(directory, "stage1", "evidence", "testdata")
	if err != nil {
		test.Fatal(err)
	}
	if !reflect.DeepEqual(listed, want) {
		test.Fatalf("git's answer: %q, want %q", listed, want)
	}
	test.Setenv("ADAMIC_BUILD_TRACE", test.TempDir())
	walked, err := PathsUnder(directory, "stage1", "evidence", "testdata")
	if err != nil {
		test.Fatal(err)
	}
	if !reflect.DeepEqual(walked, want) {
		test.Fatalf("the traced walk: %q, want %q", walked, want)
	}
	// A root that is itself a skipped name is still listed: skip applies below root.
	if rooted, err := PathsUnder(
		directory,
		"stage1/rules/x/evidence",
		"evidence",
	); err != nil || !reflect.DeepEqual(rooted, []string{"stage1/rules/x/evidence/log.txt"}) {
		test.Fatalf("a skipped name as root: %q %v", rooted, err)
	}
}

// An unpacked source's manifest is read as key-time work (asKeying, buildcache.Keying), as git's answers are, so a
// traced product test from a source with no .git never records .tracked in its read set (#asrz3cy). The counting
// wrapper puts the manifest in place only while it runs, so a read outside it, of a record or of a parent's HEAD, finds
// no manifest and fails. Mutant: readManifest's reads not wrapped.
// Not parallel: asKeying.
func TestManifestIsReadAsKeying(test *testing.T) {
	checkout := checkout(test)
	source := unpack(test, checkout, true)
	away := filepath.Join(test.TempDir(), "away")
	if err := os.Rename(filepath.Join(source, ManifestDirectory), away); err != nil {
		test.Fatal(err)
	}
	calls := 0
	saved := asKeying
	asKeying = func(compute func()) {
		calls++
		if err := os.Rename(away, filepath.Join(source, ManifestDirectory)); err != nil {
			test.Fatal(err)
		}
		defer func() {
			if err := os.Rename(filepath.Join(source, ManifestDirectory), away); err != nil {
				test.Fatal(err)
			}
		}()
		saved(compute)
	}
	test.Cleanup(func() { asKeying = saved })
	for _, repository := range []string{"cohere/TypeScript", "cohere", "."} {
		gitHead, err := Head(filepath.Join(checkout, repository))
		if err != nil {
			test.Fatal(err)
		}
		if head, err := Head(filepath.Join(source, repository)); err != nil || head != gitHead {
			test.Fatalf("%s: HEAD %s (%v), git says %s", repository, head, err, gitHead)
		}
	}
	// TypeScript's record and the look for cohere's HEAD, cohere's and the look for the root's HEAD, the root's.
	if calls < 5 {
		test.Fatalf("the manifest was read through asKeying %d times", calls)
	}
}
