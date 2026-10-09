package main

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/cohere/internal/format/formatfiles"
	"github.com/system-inc/cohere/internal/types/program"
)

// nestedTree writes a project whose library is a declared submodule, with a submodule declared inside
// the library, the shape of ahra, libraries/structure and libraries/structure/libraries/nexus, and a
// clone under projects/ that nobody declared, as ahra holds six.
func nestedTree(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, directory := range []string{".git", "library/.git", "library/inner/.git", "library/source", "modules", "projects/clone/.git"} {
		if err := os.MkdirAll(filepath.Join(root, directory), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	writeTree(t, root, map[string]string{
		"modules/Host.ts":          "export const host   =   1\n",
		"library/source/Thing.ts":  "export const thing   =   1\n",
		"library/source/Tidy.ts":   "export const tidy = 1\n",
		"library/inner/Deep.ts":    "export const deep   =   1\n",
		"library/inner/Settled.ts": "export const settled = 1\n",
		"library/.gitignore":       "ignored/\n",
		"library/ignored/Skip.ts":  "export const skip   =   1\n",
		"projects/clone/Clone.ts":  "export const clone   =   1\n",
		".gitmodules":              "[submodule \"library\"]\n\tpath = library\n\turl = ../library\n",
		"library/.gitmodules":      "[submodule \"inner\"]\n\tpath = inner\n\turl = ../inner\n",
	})
	return root
}

// TestARunInsideALibraryWritesTheLibrary: the repository a run writes is the nearest one at or above
// where it started, inside the project; the project itself everywhere else.
func TestARunInsideALibraryWritesTheLibrary(t *testing.T) {
	t.Parallel()
	root := nestedTree(t)
	for _, testCase := range []struct {
		start string
		want  string
	}{
		{".", "."},
		{"modules", "."},
		{"library", "library"},
		{"library/source", "library"},
		{"library/inner", "library/inner"},
	} {
		got := writeRepositoryRoot(filepath.Join(root, testCase.start), root)
		if got != filepath.Join(root, testCase.want) {
			t.Errorf("a run started in %s writes %s, want %s", testCase.start, got, filepath.Join(root, testCase.want))
		}
	}

	// A project that is not a repository at all writes where it always did.
	plain := t.TempDir()
	if err := os.MkdirAll(filepath.Join(plain, "source"), 0o755); err != nil {
		t.Fatal(err)
	}
	if got := writeRepositoryRoot(filepath.Join(plain, "source"), plain); got != plain {
		t.Errorf("a plain directory's run writes %s, want the project %s", got, plain)
	}
}

// TestARunWritesOnlyItsOwnRepository: from inside a library, a file in the project above and a file in
// a repository nested in the library are both somebody else's, named; the library's own file is not.
func TestARunWritesOnlyItsOwnRepository(t *testing.T) {
	t.Parallel()
	root := nestedTree(t)
	library := filepath.Join(root, "library")

	if why := unwritableRepository(library, filepath.Join(library, "source", "Thing.ts")); why != "" {
		t.Errorf("the library's own file was refused: %s", why)
	}
	if why := unwritableRepository(library, filepath.Join(root, "modules", "Host.ts")); why != "the project outside "+library {
		t.Errorf("the project's file was not refused as outside the library: %q", why)
	}
	if why := unwritableRepository(library, filepath.Join(library, "inner", "Deep.ts")); why != "nested repository inner" {
		t.Errorf("a file in a repository nested in the library was not refused under it: %q", why)
	}
	if why := unwritableRepository(root, filepath.Join(library, "source", "Thing.ts")); why != "nested repository library" {
		t.Errorf("from the project, the library's file was not refused under the library: %q", why)
	}
}

// TestTheProjectReadsItsLibrariesDrift: every declared submodule is read, recursively, each with its own
// ignore layers, and each file its own run would rewrite is reported under that repository. A clone
// nobody declared is not read, and nothing is written.
func TestTheProjectReadsItsLibrariesDrift(t *testing.T) {
	t.Parallel()
	root := nestedTree(t)
	engine := prettierLike()
	engine.enumerate = func(walkRoot string) (formatfiles.Enumeration, error) {
		return formatfiles.Enumerate(walkRoot, engine.Handles)
	}
	before := treeSnapshot(t, root)

	check, err := checkNestedRepositories(engine, root, false)
	if err != nil {
		t.Fatal(err)
	}

	got := []string{}
	for _, drift := range check.Drift {
		relative, _ := filepath.Rel(root, drift.FileName)
		got = append(got, drift.Repository+": "+relative)
	}
	want := []string{"library/inner: library/inner/Deep.ts", "library: library/source/Thing.ts"}
	if strings.Join(got, ", ") != strings.Join(want, ", ") {
		t.Fatalf("drift %v, want %v (the ignored file and the undeclared clone skipped, the project's own file not read here)", got, want)
	}
	if check.Repositories != 2 || check.Files != 4 {
		t.Fatalf("read %d repositories and %d files, want 2 and 4", check.Repositories, check.Files)
	}
	assertTreeUnchanged(t, root, before, "checkNestedRepositories")
}

// TestTheNestedCheckKeepsEachRepositorysOwnRecord: the drift check formats only what is not on a nested
// repository's own record at its current bytes, keeps that record in the repository's own cache table and
// never the parent's, and still reports every drifted file: one recorded clean and then misformatted, and
// one whose recorded entry was written under another formatter's key or other options.
func TestTheNestedCheckKeepsEachRepositorysOwnRecord(t *testing.T) {
	t.Parallel()
	root := nestedTree(t)
	library := filepath.Join(root, "library")
	inner := filepath.Join(library, "inner")
	engine := prettierLike()
	engine.enumerate = func(walkRoot string) (formatfiles.Enumeration, error) {
		return formatfiles.Enumerate(walkRoot, engine.Handles)
	}
	check := func(label string, wantChecked int, wantDrift ...string) {
		t.Helper()
		result, err := checkNestedRepositories(engine, root, false)
		if err != nil {
			t.Fatal(err)
		}
		got := []string{}
		for _, drift := range result.Drift {
			relative, _ := filepath.Rel(root, drift.FileName)
			got = append(got, relative)
		}
		if strings.Join(got, ", ") != strings.Join(wantDrift, ", ") {
			t.Fatalf("%s: drift %v, want %v", label, got, wantDrift)
		}
		if result.Files != 4 || result.Checked != wantChecked {
			t.Fatalf("%s: %d files, %d checked, want 4 files and %d checked", label, result.Files, result.Checked, wantChecked)
		}
	}
	drifted := []string{"library/inner/Deep.ts", "library/source/Thing.ts"}

	// A repository nobody has run cohere in keeps no cache, and a check that only reads it creates none:
	// every file goes through the formatter, every time. A run inside each repository makes its cache.
	for range 2 {
		check("a check before either repository keeps a cache", 4, drifted...)
	}
	for _, repository := range []string{library, inner} {
		if _, err := os.Stat(cacheDirectory(repository)); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("a check that only reads %s made its cache directory (%v)", repository, err)
		}
		if err := os.MkdirAll(cacheDirectory(repository), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	// Nothing on record yet: every file goes through the formatter.
	check("the first check", 4, drifted...)

	// Each repository holds its own files and no other's; the parent holds none of them.
	entries := func(repository string) []string {
		section := readFormatSection(repository)
		if section == nil {
			return nil
		}
		names := []string{}
		for fileName := range section.Entries {
			relative, _ := filepath.Rel(root, fileName)
			names = append(names, relative)
		}
		sort.Strings(names)
		return names
	}
	if got := entries(root); got != nil {
		t.Fatalf("the nested check wrote the parent's record: %v", got)
	}
	if got := strings.Join(entries(library), ", "); got != "library/source/Thing.ts, library/source/Tidy.ts" {
		t.Fatalf("the library's record holds %s", got)
	}
	if got := strings.Join(entries(inner), ", "); got != "library/inner/Deep.ts, library/inner/Settled.ts" {
		t.Fatalf("the inner repository's record holds %s", got)
	}

	// The settled files are on record and skipped; the drifted ones are recorded at their formatted text,
	// so their bytes on disk stay out of the record and stay reported.
	check("the second check", 2, drifted...)

	// The positive control: a file the last check recorded clean, misformatted, is drift again.
	writeTree(t, root, map[string]string{"library/source/Tidy.ts": "export const tidy   =   1\n"})
	check("a recorded file misformatted", 3, "library/inner/Deep.ts", "library/source/Thing.ts", "library/source/Tidy.ts")
	writeTree(t, root, map[string]string{"library/source/Tidy.ts": "export const tidy = 1\n"})

	// An entry vouching for Thing.ts's misformatted bytes. Under this formatter's key and options the
	// record is believed, which is what proves the check consults it; under another key, or options the
	// file no longer formats with, it says nothing and the drift is reported.
	thing := filepath.Join(library, "source", "Thing.ts")
	contents := readForTest(t, thing)
	options, err := engine.OptionsFingerprint(thing)
	if err != nil {
		t.Fatal(err)
	}
	key, err := formatRecordKey(library)
	if err != nil {
		t.Fatal(err)
	}
	vouch := func(key string, options string) {
		t.Helper()
		section := readFormatSection(library)
		section.Key = key
		section.Entries[thing] = program.FormatEntry{Sum: formatRecordSum([]byte(contents)), Options: options}
		if err := writeFormatSection(library, section); err != nil {
			t.Fatal(err)
		}
	}
	vouch(key, options)
	check("an entry vouching for the bytes under this key", 1, "library/inner/Deep.ts")
	vouch("formatter another", options)
	check("an entry under another formatter's key", 3, drifted...)
	vouch(key, options+" changed")
	check("an entry under other options", 2, drifted...)

	// `--format-all` reads every nested file: an entry vouching for Thing.ts's misformatted bytes, under
	// this formatter's key and options, says nothing to it.
	vouch(key, options)
	all, err := checkNestedRepositories(engine, root, true)
	if err != nil {
		t.Fatal(err)
	}
	if all.Checked != 4 || len(all.Drift) != 2 {
		t.Fatalf("--format-all checked %d of 4 files and found %d drifted, want 4 and 2 (the record's word taken for Thing.ts)", all.Checked, len(all.Drift))
	}
}

// TestASignedEntryStillReadsAnEditedFile is the positive control through the record's cheap path. An
// entry signed with its file's size and modification time vouches for the file without reading it, so
// an edit, which moves both, must send the file back through the formatter.
func TestASignedEntryStillReadsAnEditedFile(t *testing.T) {
	t.Parallel()
	root := nestedTree(t)
	library := filepath.Join(root, "library")
	if err := os.MkdirAll(cacheDirectory(library), 0o755); err != nil {
		t.Fatal(err)
	}
	engine := prettierLike()
	engine.enumerate = func(walkRoot string) (formatfiles.Enumeration, error) {
		return formatfiles.Enumerate(walkRoot, engine.Handles)
	}
	// Older than the second a fresh write is left unsigned for.
	tidy := filepath.Join(library, "source", "Tidy.ts")
	past := time.Now().Add(-time.Hour)
	if err := os.Chtimes(tidy, past, past); err != nil {
		t.Fatal(err)
	}
	drifted := func() []string {
		t.Helper()
		check, err := checkNestedRepositories(engine, root, false)
		if err != nil {
			t.Fatal(err)
		}
		names := []string{}
		for _, drift := range check.Drift {
			names = append(names, filepath.Base(drift.FileName))
		}
		return names
	}

	// Recorded, then signed on the next check, which finds its bytes on record.
	drifted()
	drifted()
	if entry := readFormatSection(library).Entries[tidy]; entry.ModifiedNanoseconds != past.UnixNano() {
		t.Fatalf("Tidy.ts was not signed with its modification time: %+v", entry)
	}

	writeTree(t, root, map[string]string{"library/source/Tidy.ts": "export const tidy   =   1\n"})
	if got := strings.Join(drifted(), ", "); got != "Deep.ts, Thing.ts, Tidy.ts" {
		t.Fatalf("an edit to a signed file was taken on the record's word: drift %s", got)
	}
}

// TestTheNestedLineSaysWhatItRead: the files the formatter checked and the files taken on the record's
// word are counted apart, and "would change" is said only of checked files.
func TestTheNestedLineSaysWhatItRead(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		check nestedDriftCheck
		want  string
	}{
		{nestedDriftCheck{Repositories: 2, Files: 2073},
			"nested repositories: 2 read, 2073 files: none checked by the formatter, all 2073 taken on their record's word as formatted at their current bytes"},
		{nestedDriftCheck{Repositories: 2, Files: 2073, Checked: 2073, Drift: make([]nestedDrift, 3)},
			"nested repositories: 2 read, 2073 files: 2073 checked by the formatter, 3 would change under their own run"},
		{nestedDriftCheck{Repositories: 2, Files: 2073, Checked: 23, Drift: make([]nestedDrift, 1)},
			"nested repositories: 2 read, 2073 files: 23 checked by the formatter, 1 would change under their own run; 2050 taken on their record's word as formatted at their current bytes"},
		{nestedDriftCheck{},
			"nested repositories: 0 read, 0 files: 0 checked by the formatter, 0 would change under their own run"},
	} {
		if got := nestedSummary(testCase.check); got != testCase.want {
			t.Errorf("got  %s\nwant %s", got, testCase.want)
		}
	}
}

// TestALibraryIsFormattedFromItsOwnRootAndReadFromItsProject is the ruling end to end, through the
// binary: the project's check reports the library's drift and writes nothing there; a run started
// inside the library formats and fixes the library and writes nothing in the project above it; and the
// project's check is then clean of it.
func TestALibraryIsFormattedFromItsOwnRootAndReadFromItsProject(t *testing.T) {
	t.Parallel()
	binary := buildCohere(t)
	root := t.TempDir()
	writeTree(t, root, map[string]string{
		"tsconfig.json":            fixScopeTsconfig,
		"CohereSettings.json":      `{"extends":"./NexusCohereSettings.json","rules":{"no-debugger":"error"}}`,
		"NexusCohereSettings.json": `{"format":{"ignore":[]}}`,
		// The project's own files: a fixable finding and a formatting one, which no run inside the
		// library may touch.
		"Producer.ts":         "export function value(): number {\n  debugger;\n  return 1;\n}\n",
		"Ugly.ts":             "export const ugly   =   1\n",
		".gitmodules":         "[submodule \"library\"]\n\tpath = library\n\turl = ../library\n",
		"library/.git":        "gitdir: ../.git/modules/library\n",
		"library/Inner.ts":    "export const inner   =   1\n",
		"library/Debugged.ts": "export function debugged(): number {\n  debugger;\n  return 1;\n}\n",
	})
	project := func() map[string]string {
		snapshot := treeSnapshot(t, root)
		for path := range snapshot {
			if strings.HasPrefix(path, filepath.Join(root, "library")+string(filepath.Separator)) {
				delete(snapshot, path)
			}
		}
		return snapshot
	}
	library := filepath.Join(root, "library")

	// The project's check reads the library and reports the file its formatter would rewrite, writing
	// nothing anywhere. The library's fixable finding is lint's to report, at its position, and the run
	// says its fix was not applied.
	before := treeSnapshot(t, root)
	output, code := runCohere(t, binary, root, "--no-fix", "--format")
	if code == 0 || !strings.Contains(output, "nested repositories: 1 read, 2 files: 2 checked by the formatter, 1 would change under their own run") ||
		!strings.Contains(output, "1 fix not applied: in nested repository library") ||
		!strings.Contains(output, filepath.Join(library, "Inner.ts")+":1:1 - the formatter would rewrite this file in nested repository library") {
		t.Fatalf("the project's check did not report the library's drift, exit %d:\n%s", code, output)
	}
	assertTreeUnchanged(t, root, before, output)

	// A run started inside the library walks the library and writes only there.
	projectBefore := project()
	output, code = runCohere(t, binary, library, "--fix", "--format")
	if code != 0 {
		t.Fatalf("the library's run exited %d:\n%s", code, output)
	}
	if !strings.Contains(output, "walked "+library) {
		t.Fatalf("the library's run did not walk the library:\n%s", output)
	}
	if got := readForTest(t, filepath.Join(library, "Inner.ts")); got != "export const inner = 1;\n" {
		t.Fatalf("the library's run did not format the library:\n%s\noutput:\n%s", got, output)
	}
	if got := readForTest(t, filepath.Join(library, "Debugged.ts")); strings.Contains(got, "debugger") {
		t.Fatalf("the library's run did not fix the library:\n%s", got)
	}
	projectAfter := project()
	for path, contents := range projectBefore {
		if projectAfter[path] != contents {
			t.Errorf("the library's run wrote %s in the project above it:\n%s", path, projectAfter[path])
		}
	}
	for path := range projectAfter {
		if _, existed := projectBefore[path]; !existed {
			t.Errorf("the library's run created %s in the project above it", path)
		}
	}

	// The project's check now reads the library clean, and its own files are still its own to fix.
	// The library's run left its record in the library, and the project's check reads it, so nothing in
	// the library is formatted again.
	output, _ = runCohere(t, binary, root, "--no-fix", "--format")
	if !strings.Contains(output, "nested repositories: 1 read, 2 files: none checked by the formatter, all 2 taken on their record's word as formatted at their current bytes") {
		t.Fatalf("the library is still reported, or formatted again, after its own run formatted it:\n%s", output)
	}

	// `--format-all` reads every nested file whatever the record says.
	output, _ = runCohere(t, binary, root, "--no-fix", "--format-all")
	if !strings.Contains(output, "nested repositories: 1 read, 2 files: 2 checked by the formatter, 0 would change under their own run\n") {
		t.Fatalf("--format-all took the library's record's word:\n%s", output)
	}
	if !strings.Contains(output, filepath.Join(root, "Producer.ts")+":1:1 - --fix would rewrite this file") {
		t.Fatalf("the project's own drift went missing:\n%s", output)
	}

	// The positive control, right after a check that read the file clean from the record: a misformat
	// planted in it is drift again, on this check and the next.
	writeTree(t, root, map[string]string{"library/Inner.ts": "export const inner   =   1;\n"})
	for _, label := range []string{"the check after the plant", "the check after that"} {
		output, code = runCohere(t, binary, root, "--no-fix", "--format")
		if code == 0 || !strings.Contains(output, "nested repositories: 1 read, 2 files: 1 checked by the formatter, 1 would change under their own run; 1 taken on their record's word as formatted at their current bytes") ||
			!strings.Contains(output, filepath.Join(library, "Inner.ts")+":1:1 - the formatter would rewrite this file in nested repository library") {
			t.Fatalf("%s did not report the planted misformat, exit %d:\n%s", label, code, output)
		}
	}
}

// TestAFreshClonesLibrariesGetTheirCacheAtStartup: a checkout where nobody has run cohere inside a library, a
// fresh clone or a CI machine, kept no cache there, so every check formatted every library file, every run
// (#v0etgwx). The run's start now makes the cache, only in a library whose own .gitignore ignores it, and the
// next check reads that library's record. The control is the inner library, whose .gitignore does not ignore
// .cache: it gets no directory, and every one of its files is still checked.
func TestAFreshClonesLibrariesGetTheirCacheAtStartup(t *testing.T) {
	t.Parallel()
	root := nestedTree(t)
	writeTree(t, root, map[string]string{"library/.gitignore": "ignored/\n.cache/\n"})
	library := filepath.Join(root, "library")
	inner := filepath.Join(library, "inner")
	engine := prettierLike()
	engine.enumerate = func(walkRoot string) (formatfiles.Enumeration, error) {
		return formatfiles.Enumerate(walkRoot, engine.Handles)
	}

	prepareNestedCacheDirectories(root)
	if information, err := os.Stat(cacheDirectory(library)); err != nil || !information.IsDir() {
		t.Fatalf("the library, whose .gitignore ignores .cache, has no cache after the run's start (%v)", err)
	}
	if _, err := os.Stat(cacheDirectory(inner)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("the inner library, whose .gitignore does not ignore .cache, got a cache directory (%v)", err)
	}

	checked := func(label string) int {
		t.Helper()
		result, err := checkNestedRepositories(engine, root, false)
		if err != nil {
			t.Fatal(err)
		}
		if len(result.Drift) != 2 {
			t.Fatalf("%s: %d files drift, want Thing.ts and Deep.ts", label, len(result.Drift))
		}
		return result.Checked
	}
	// The first check formats all four and records the library's. The second takes the library's clean file
	// on its record's word; its misformatted file and the inner library's two are checked again.
	if got := checked("the first check"); got != 4 {
		t.Fatalf("the first check formatted %d files, want 4", got)
	}
	if got := checked("the second check"); got != 3 {
		t.Fatalf("the second check formatted %d files, want 3: the library's record was not read", got)
	}

	// Run again, the start leaves an existing cache as it is.
	prepareNestedCacheDirectories(root)
	if got := checked("the third check"); got != 3 {
		t.Fatalf("after a second start the check formatted %d files, want 3", got)
	}
}
