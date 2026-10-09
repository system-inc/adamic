package program_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/system-inc/cohere/internal/types/program"
)

// listEverything is what discovery does for a build: every directory under root, listed once.
func listEverything(t *testing.T, root string) *program.DirectoryListings {
	t.Helper()
	listings := program.NewDirectoryListings()
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || !entry.IsDir() {
			return err
		}
		entries, err := os.ReadDir(path)
		if err != nil {
			return err
		}
		listings.Add(path, entries)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return listings
}

// listingsReport is what a build makes of its tree: the files the tsconfig named, and every input the run
// cache would record.
func listingsReport(t *testing.T, directory string, listings *program.DirectoryListings) (ours []string, present []string, absent []string) {
	t.Helper()
	recorder := program.NewInputRecorder()
	graph, err := program.Build(program.Options{ConfigFileName: "tsconfig.json", CurrentDirectory: directory, Inputs: recorder, Listings: listings})
	if err != nil {
		t.Fatalf("building: %v", err)
	}
	for _, sourceFile := range graph.ProjectFiles() {
		ours = append(ours, sourceFile.FileName().AsString())
	}
	slices.Sort(ours)
	present, absent, _ = recorder.Inputs()
	slices.Sort(present)
	slices.Sort(absent)
	return ours, present, absent
}

// A build that lists its directories from discovery's listings names the same files, in the same order, and
// records the same inputs as one that lists them from disk (#bjv0tg4). The tree holds what the listing has to
// answer as the disk does: nested directories, a symbolic link to a file and one to a directory (both
// followed, as the disk follows them), a dangling link (left out), and a directory the include does not
// reach. The control: a listing that has gone stale, one file short, names one file fewer, so the build
// really reads the listings and the comparison can fail.
func TestABuildListsFromDiscoverysListingsAsFromDisk(t *testing.T) {
	t.Parallel()
	directory, err := filepath.EvalSymlinks(writeProject(t, map[string]string{
		"tsconfig.json":          `{"compilerOptions": {"target": "ES2022", "strict": true, "noEmit": true}, "include": ["src/**/*.ts"]}`,
		"src/index.ts":           "export const index = 1;\n",
		"src/nested/deep/a.ts":   "export const a = 1;\n",
		"src/nested/b.ts":        "export const b = 1;\n",
		"elsewhere/linked.ts":    "export const linked = 1;\n",
		"elsewhere/dir/inner.ts": "export const inner = 1;\n",
		"outside/ignored.ts":     "export const ignored = 1;\n",
	}))
	if err != nil {
		t.Fatal(err)
	}
	for link, target := range map[string]string{
		"src/linked.ts":   "../elsewhere/linked.ts",
		"src/linkeddir":   "../elsewhere/dir",
		"src/dangling.ts": "../nowhere.ts",
	} {
		if err := os.Symlink(target, filepath.Join(directory, link)); err != nil {
			t.Fatal(err)
		}
	}

	diskOurs, diskPresent, diskAbsent := listingsReport(t, directory, nil)
	listings := listEverything(t, directory)
	listedOurs, listedPresent, listedAbsent := listingsReport(t, directory, listings)
	if !slices.Equal(diskOurs, listedOurs) {
		t.Fatalf("the files named differ:\ndisk   %v\nlisted %v", diskOurs, listedOurs)
	}
	if len(diskOurs) != 5 {
		t.Fatalf("the fixture's five included files (two through links) came back as %v", diskOurs)
	}
	if !slices.Equal(diskPresent, listedPresent) || !slices.Equal(diskAbsent, listedAbsent) {
		t.Fatalf("the recorded inputs differ:\ndisk   %v %v\nlisted %v %v", diskPresent, diskAbsent, listedPresent, listedAbsent)
	}

	stale := program.NewDirectoryListings()
	stale.Add(filepath.Join(directory, "src", "nested"), nil)
	if staleOurs, _, _ := listingsReport(t, directory, stale); len(staleOurs) != len(diskOurs)-2 {
		t.Fatalf("control: a listing emptied of src/nested named %v, so the build did not read it", staleOurs)
	}
}

// Listing answers for a directory that was listed, by either separator, and says plainly when one was not,
// so a caller reads that one from disk rather than taking it for empty.
func TestListingAnswersForWhatWasListedAndOnlyThat(t *testing.T) {
	t.Parallel()
	directory := writeProject(t, map[string]string{"a.ts": "export const a = 1;\n", "sub/b.ts": "export const b = 1;\n"})
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	listings := program.NewDirectoryListings()
	listings.Add(directory, entries)

	got, listed := listings.Listing(directory + string(filepath.Separator))
	if !listed || len(got) != 2 || got[0].Name() != "a.ts" || got[1].Name() != "sub" || !got[1].IsDir() {
		t.Fatalf("the listed directory answered %v, %t", got, listed)
	}
	if got, listed := listings.Listing(filepath.Join(directory, "sub")); listed || got != nil {
		t.Fatalf("a directory nobody listed answered %v, %t, as if it were listed", got, listed)
	}
}
