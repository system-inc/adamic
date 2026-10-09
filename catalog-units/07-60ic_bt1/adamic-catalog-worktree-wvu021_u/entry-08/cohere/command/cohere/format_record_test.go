package main

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/cohere/internal/format/formatfiles"
)

// The format record decides what the default format scope holds, so each test below asserts the exact
// set it puts in scope, name by name. "Some files were in scope" is what a record that skips the wrong
// ones also produces.

// recordFixture is a project on disk with a declared submodule and an undeclared clone, and an engine
// that walks it for real. The record lives in the real cache table, under a home of the test's own.
type recordFixture struct {
	root   string
	engine *fakeEngine
}

func newRecordFixture(t *testing.T) recordFixture {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CACHE_HOME", "")
	root := t.TempDir()
	writeTree(t, root, map[string]string{
		"A.ts":         "export const a = 1;\n",
		"B.ts":         "export const b = 2;\n",
		"notes.txt":    "not formatted by anything\n",
		".gitignore":   "ignored/\n",
		"ignored/C.ts": "export const c = 3;\n",
		".gitmodules":  "[submodule \"library\"]\n\tpath = library\n\turl = ../library\n",
		// A submodule's `.git` is a gitlink file.
		"library/.git":     "gitdir: ../.git/modules/library\n",
		"library/Inner.ts": "export const inner = 4;\n",
		// A clone nobody declared, the shape of ahra's `projects/*`.
		"projects/clone/.git/HEAD": "ref: refs/heads/main\n",
		"projects/clone/Outer.ts":  "export const outer = 5;\n",
	})
	engine := &fakeEngine{handled: []string{".ts"}}
	engine.enumerate = func(directory string) (formatfiles.Enumeration, error) {
		return formatfiles.Enumerate(directory, engine.Handles)
	}
	return recordFixture{root: root, engine: engine}
}

func (fixture recordFixture) path(name string) string {
	return filepath.Join(fixture.root, filepath.FromSlash(name))
}

// scope loads the record the way a run does and returns the default scope's files, relative and
// sorted, with the universe and the record it read.
func (fixture recordFixture) scope(t *testing.T) ([]string, []string, *formatRecord, string) {
	t.Helper()
	record := loadFormatRecord(fixture.root)
	scope, universe := unformattedScope(fixture.engine, record, fixture.root)
	return fixture.relative(t, scope.FileNames), universe, record, scope.Description
}

func (fixture recordFixture) relative(t *testing.T, files []string) []string {
	t.Helper()
	out := []string{}
	for _, file := range files {
		relative, err := filepath.Rel(fixture.root, file)
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, filepath.ToSlash(relative))
	}
	return out
}

// formatEverything runs the files in scope through an observed transform the way the fix phase does,
// with a formatter that finds every file already formatted, and saves the record.
func (fixture recordFixture) formatEverything(t *testing.T) {
	t.Helper()
	record := loadFormatRecord(fixture.root)
	scope, universe := unformattedScope(fixture.engine, record, fixture.root)
	transform := record.observe(formatTransform(fixture.engine), fixture.engine.OptionsFingerprint)
	for _, fileName := range scope.FileNames {
		if _, err := transform(fileName, readForTest(t, fileName), nil); err != nil {
			t.Fatal(err)
		}
	}
	if err := record.save(universe); err != nil {
		t.Fatal(err)
	}
}

// aged moves a file's modification time into the past, past the second inside which no signature is
// recorded, so the next read can be answered by stat alone.
func aged(t *testing.T, path string) {
	t.Helper()
	old := time.Now().Add(-time.Minute)
	if err := os.Chtimes(path, old, old); err != nil {
		t.Fatal(err)
	}
}

func assertScope(t *testing.T, got []string, want ...string) {
	t.Helper()
	if want == nil {
		want = []string{}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("in scope %v, expected %v", got, want)
	}
}

// TestWithNoRecordEveryFileOfTheRepositoryIsInScopeAndNoSubmoduleIs: the universe stops at the
// repository's own boundary. Writes stay inside one repository (@system_cohere, 2026-10-03), so the
// declared submodule is formatted by a run inside it and only read from here, by the check (see
// checkNestedRepositories); it used to be walked into and written.
// Not parallel: it sets HOME and XDG_CACHE_HOME with t.Setenv through newRecordFixture, which a parallel test may not
func TestWithNoRecordEveryFileOfTheRepositoryIsInScopeAndNoSubmoduleIs(t *testing.T) {
	fixture := newRecordFixture(t)
	files, universe, _, description := fixture.scope(t)

	// The ignored file, the undeclared clone and the declared submodule are all outside the universe.
	assertScope(t, files, "A.ts", "B.ts")
	if len(universe) != 2 {
		t.Fatalf("universe %v, expected the two files in scope", fixture.relative(t, universe))
	}
	for _, want := range []string{"all 2 files, because no earlier check is on record", "skipped nested repositories library, projects/clone"} {
		if !strings.Contains(description, want) {
			t.Errorf("the scope line does not say %q:\n%s", want, description)
		}
	}
}

// Not parallel: it sets HOME and XDG_CACHE_HOME with t.Setenv through newRecordFixture, which a parallel test may not
func TestTheRecordScopesWhatChangedSinceCohereLastLooked(t *testing.T) {
	fixture := newRecordFixture(t)
	fixture.formatEverything(t)

	files, _, _, description := fixture.scope(t)
	assertScope(t, files)
	if !strings.Contains(description, "0 of 2 files not on record as formatted") {
		t.Errorf("an unchanged tree's scope line: %s", description)
	}

	original := readForTest(t, fixture.path("A.ts"))
	writeTree(t, fixture.root, map[string]string{"A.ts": "export const a = 10;\n"})
	files, _, _, _ = fixture.scope(t)
	assertScope(t, files, "A.ts")

	// Edited back to the bytes on record: unchanged, whatever its modification time says.
	writeTree(t, fixture.root, map[string]string{"A.ts": original})
	files, _, _, _ = fixture.scope(t)
	assertScope(t, files)

	// A new file nobody has formatted is in scope, untracked or not: nothing here asks git.
	writeTree(t, fixture.root, map[string]string{"New.ts": "export const fresh = 6;\n"})
	// An edit inside the submodule is the submodule's run's to format, not this one's: the project's
	// check reports it as the submodule's drift instead.
	writeTree(t, fixture.root, map[string]string{"library/Inner.ts": "export const inner = 40;\n"})
	files, _, _, _ = fixture.scope(t)
	assertScope(t, files, "New.ts")
}

// Not parallel: it sets HOME and XDG_CACHE_HOME with t.Setenv through newRecordFixture, which a parallel test may not
func TestADeletedFileLeavesTheScopeAndTheRecord(t *testing.T) {
	fixture := newRecordFixture(t)
	fixture.formatEverything(t)
	if err := os.Remove(fixture.path("B.ts")); err != nil {
		t.Fatal(err)
	}

	files, universe, record, _ := fixture.scope(t)
	assertScope(t, files)
	if err := record.save(universe); err != nil {
		t.Fatal(err)
	}
	stored := readFormatSection(fixture.root)
	if stored == nil {
		t.Fatal("the record was not written to the cache table")
	}
	if _, kept := stored.Entries[fixture.path("B.ts")]; kept {
		t.Fatal("the record kept an entry for a deleted file")
	}
	if len(stored.Entries) != 1 {
		t.Fatalf("the record holds %d entries, expected the one file that remains", len(stored.Entries))
	}
}

// An unformatted file checked under --no-fix records the text the formatter would write, which is not
// what is on disk, so the file stays in scope and keeps being reported until it is written.
// Not parallel: it sets HOME and XDG_CACHE_HOME with t.Setenv through newRecordFixture, which a parallel test may not
func TestAnUnformattedFileStaysInScopeUntilItIsWritten(t *testing.T) {
	fixture := newRecordFixture(t)
	fixture.engine.format = func(fileName string, text string) (string, error) {
		return strings.ReplaceAll(text, "  ", " "), nil
	}
	writeTree(t, fixture.root, map[string]string{"A.ts": "export const a  = 1;\n"})
	fixture.formatEverything(t)

	files, _, _, _ := fixture.scope(t)
	assertScope(t, files, "A.ts")

	// Written as the formatter would write it, it is on record.
	writeTree(t, fixture.root, map[string]string{"A.ts": "export const a = 1;\n"})
	files, _, _, _ = fixture.scope(t)
	assertScope(t, files)
}

// Not parallel: it sets HOME and XDG_CACHE_HOME with t.Setenv through newRecordFixture, which a parallel test may not
func TestAnOptionsChangeSendsItsFilesBackThroughTheFormatter(t *testing.T) {
	fixture := newRecordFixture(t)
	fixture.formatEverything(t)
	fixture.engine.options = func(fileName string) (string, error) {
		if strings.HasSuffix(fileName, "B.ts") {
			return "printWidth 80", nil
		}
		return "fake options", nil
	}

	files, _, _, _ := fixture.scope(t)
	assertScope(t, files, "B.ts")
}

// Not parallel: it sets HOME and XDG_CACHE_HOME with t.Setenv through newRecordFixture, which a parallel test may not
func TestARecordFromAnotherCohereSaysNothing(t *testing.T) {
	fixture := newRecordFixture(t)
	fixture.formatEverything(t)

	stored := readFormatSection(fixture.root)
	stored.Key = "another cohere"
	if err := writeFormatSection(fixture.root, stored); err != nil {
		t.Fatal(err)
	}

	files, _, _, description := fixture.scope(t)
	assertScope(t, files, "A.ts", "B.ts")
	if !strings.Contains(description, "the formatter changed since the last check") {
		t.Errorf("the scope line does not say why every file is in scope: %s", description)
	}

	// An unreadable table is the same answer, never an error.
	if err := os.WriteFile(filepath.Join(cacheDirectory(fixture.root), "format.gob"), []byte("not a cache table"), 0o644); err != nil {
		t.Fatal(err)
	}
	files, _, _, _ = fixture.scope(t)
	assertScope(t, files, "A.ts", "B.ts")
}

// The stat shortcut is only taken once the bytes were read and matched, and a file whose bytes change
// is caught whether or not its size does.
// Not parallel: it sets HOME and XDG_CACHE_HOME with t.Setenv through newRecordFixture, which a parallel test may not
func TestAVerifiedFileIsAnsweredByStatAndAnEditIsStillSeen(t *testing.T) {
	fixture := newRecordFixture(t)
	fixture.formatEverything(t)
	for _, name := range []string{"A.ts", "B.ts"} {
		aged(t, fixture.path(name))
	}

	// The first read verifies by content and records each signature.
	files, universe, record, _ := fixture.scope(t)
	assertScope(t, files)
	if err := record.save(universe); err != nil {
		t.Fatal(err)
	}
	reloaded := loadFormatRecord(fixture.root)
	if entry := reloaded.entries[fixture.path("A.ts")]; entry.ModifiedNanoseconds == 0 {
		t.Fatal("a file read and matched was given no signature, so every run reads it again")
	}

	// Same size, new bytes.
	writeTree(t, fixture.root, map[string]string{"A.ts": "export const a = 9;\n"})
	files, _, _, _ = fixture.scope(t)
	assertScope(t, files, "A.ts")
}

func TestDeclaredSubmodulesReadsTheGitmodulesPaths(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	writeTree(t, directory, map[string]string{".gitmodules": "" +
		"[submodule \"libraries/structure\"]\n" +
		"\tpath = libraries/structure\n" +
		"\turl = git@github.com:ahraia/structure-next.git\n" +
		"\tbranch = main\n" +
		"[submodule \"quoted\"]\n" +
		"\tpath = \"libraries/with space\"\n"})

	declared, err := declaredSubmodules(directory)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]struct{}{"libraries/structure": {}, "libraries/with space": {}}
	if !reflect.DeepEqual(declared, want) {
		t.Fatalf("declared %v, expected %v", declared, want)
	}

	if none, err := declaredSubmodules(t.TempDir()); err != nil || len(none) != 0 {
		t.Fatalf("a repository with no .gitmodules declared %v (err %v)", none, err)
	}
}

// The record follows the formatter, not the binary. Three builds of this command: two from different
// commits stamped with the same formatter identity, and a third whose formatter differs. The second
// must read what the first recorded, and the third must start over. This is the property that keeps a
// lint-only cohere commit from reformatting the tree, and the one that must not keep a record across a
// printer change.
func TestTheRecordFollowsTheFormatterNotTheBinary(t *testing.T) {
	t.Parallel()
	// Every build compiles one snapshot of the source, taken once, so the three binaries differ only in
	// the two stamps the test sets. Built from the shared working tree, they could carry different trees
	// whenever another node edited it between builds, and the second build then discarded the record:
	// the test failed about one run in three, measuring the tree's churn rather than the record's rule.
	snapshot := sourceSnapshot(t)
	// The three differ only in what the linker stamps. The first builds alone, compiling every package,
	// and the other two link at once from what it left in the build cache: one after another the links
	// were most of this test's time (#nxgt2ca), and all three started together each compiled every
	// package, since go does not share a compile in flight between processes: three times the work, beside
	// the pool's budget (load 58 on 16 cores with two runs, #91m7c16). -trimpath keeps the snapshot's
	// directory out of every package's cache key, so a run reuses what the last run compiled rather than
	// compiling the whole program cold from a new directory each time.
	stamps := []struct{ selfCommit, formatter string }{
		{strings.Repeat("1", 40), "formatterA"},
		{strings.Repeat("2", 40), "formatterA"},
		{strings.Repeat("2", 40), "formatterB"},
	}
	binaries := make([]string, len(stamps))
	failures := make([]string, len(stamps))
	build := func(index int) {
		const packaging = "github.com/system-inc/cohere/internal/release/packaging"
		stamp := stamps[index]
		binaries[index] = filepath.Join(t.TempDir(), "cohere")
		command := exec.Command("go", "build", "-buildvcs=false", "-trimpath", "-o", binaries[index],
			"-ldflags=-X "+packaging+".selfCommit="+stamp.selfCommit+" -X "+packaging+".formatterIdentity="+stamp.formatter, "./command/cohere")
		command.Dir = snapshot
		if output, err := command.CombinedOutput(); err != nil {
			failures[index] = fmt.Sprintf("cannot build cohere stamped %s, %s: %v\n%s", stamp.selfCommit, stamp.formatter, err, output)
		}
	}
	build(0)
	if failures[0] != "" {
		t.Fatal(failures[0])
	}
	var links sync.WaitGroup
	for index := 1; index < len(stamps); index++ {
		links.Go(func() { build(index) })
	}
	links.Wait()
	for _, failure := range failures {
		if failure != "" {
			t.Fatal(failure)
		}
	}
	first, lintOnly, printerChanged := binaries[0], binaries[1], binaries[2]

	root := t.TempDir()
	home := t.TempDir()
	writeTree(t, root, map[string]string{
		"tsconfig.json":            fixScopeTsconfig,
		"CohereSettings.json":      "{ \"extends\": \"./NexusCohereSettings.json\", \"rules\": {} }\n",
		"NexusCohereSettings.json": "{ \"format\": { \"ignore\": [] } }\n",
		"Tidy.ts":                  "export const tidy = 1;\n",
		"Ugly.ts":                  "export const ugly   =   1\n",
	})
	run := func(binary string, arguments ...string) string {
		t.Helper()
		command := exec.Command(binary, verboseArguments(arguments)...)
		command.Dir = root
		command.Env = append(os.Environ(), "HOME="+home, "XDG_CACHE_HOME=")
		output, _ := command.CombinedOutput()
		return string(output)
	}

	run(first, "--fix", "--format")
	if output := run(lintOnly, "--no-fix", "--format"); !strings.Contains(output, "0 of 5 files not on record as formatted") {
		t.Fatalf("a build from another commit with the same formatter did not read the record:\n%s", output)
	}
	if output := run(printerChanged, "--no-fix", "--format"); !strings.Contains(output, "all 5 files, because the formatter changed since the last check") {
		t.Fatalf("a build with another formatter trusted the record:\n%s", output)
	}
}

// sourceSnapshot copies the module's source, as it is on disk now, into a directory of the test's own,
// and returns the copy's root. Builds from it cannot see an edit anyone makes to the working tree
// afterward. The compiler pin (TypeScript, a submodule nobody edits in passing) is linked rather than
// copied. It is a copy of the working tree rather than of a commit, so the test still builds the code
// someone is changing: an export of HEAD would quietly test the old record logic instead.
//
// It copies only what `go build` reads: test files and testdata directories are left out, since the
// build ignores both and no go:embed reaches into either. They were most of the copy's 4,000 files and
// 95 MB.
func sourceSnapshot(t *testing.T) string {
	t.Helper()
	moduleRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	snapshot := t.TempDir()
	for _, name := range []string{"go.mod", "go.sum", "go.work", "go.work.sum"} {
		copySnapshotFile(t, filepath.Join(moduleRoot, name), filepath.Join(snapshot, name))
	}
	// Every directory holding source go builds: the module's own, and the modules go.work names beside
	// it. A module left out is one the snapshot's go.work names and cannot find.
	for _, directory := range []string{"command", "internal", "policy", "TypeScript-shim", "static_single_assignment", "mutation_aliasing"} {
		source := filepath.Join(moduleRoot, directory)
		err := filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkError error) error {
			if walkError != nil {
				return walkError
			}
			relative, err := filepath.Rel(moduleRoot, path)
			if err != nil {
				return err
			}
			if entry.IsDir() {
				if entry.Name() == "testdata" {
					return filepath.SkipDir
				}
				return os.MkdirAll(filepath.Join(snapshot, relative), 0o755)
			}
			if !entry.Type().IsRegular() || strings.HasSuffix(entry.Name(), "_test.go") {
				return nil
			}
			copySnapshotFile(t, path, filepath.Join(snapshot, relative))
			return nil
		})
		if err != nil {
			t.Fatalf("snapshotting %s: %v", directory, err)
		}
	}
	if err := os.Symlink(filepath.Join(moduleRoot, "TypeScript"), filepath.Join(snapshot, "TypeScript")); err != nil {
		t.Fatal(err)
	}
	return snapshot
}

func copySnapshotFile(t *testing.T, source string, destination string) {
	t.Helper()
	contents, err := os.ReadFile(source)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		t.Fatal(err)
	}
	if err := os.WriteFile(destination, contents, 0o644); err != nil {
		t.Fatal(err)
	}
}
