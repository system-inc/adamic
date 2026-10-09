package formatfiles

import (
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// listedTree is a tree with everything the walk decides on: git's ignore files at two depths, the house list
// and ignorePatterns, a nested clone and a submodule's gitlink, symbolic links to a file, a directory and
// nothing, a file of a type no engine formats, and Adamic `.a` files, one held back and one in an ignored
// directory, so the hold-back and the ignore layer before it both decide something.
func listedTree(t *testing.T) string {
	t.Helper()
	root := settingsTree(t, `["pnpm-lock.yaml"]`, `["archived/**"]`, map[string]string{
		".gitignore":               "dist/\n*.log\n",
		"a.ts":                     "export const a = 1;\n",
		"pnpm-lock.yaml":           "lockfileVersion: 9\n",
		"debug.log":                "noise\n",
		"source/.gitignore":        "generated/\n!keep.log\n",
		"source/b.ts":              "export const b = 1;\n",
		"source/keep.log":          "kept\n",
		"source/generated/g.ts":    "export const g = 1;\n",
		"source/deep/c.md":         "# c\n",
		"source/deep/deeper/d.css": "a { b: c; }\n",
		"dist/built.ts":            "export const built = 1;\n",
		"dist/built.a":             "built\n",
		"source/x.a":               "x\n",
		"archived/old.md":          "# old\n",
		"vendor/.git/HEAD":         "ref: refs/heads/main\n",
		"vendor/theirs.ts":         "export const theirs = 1;\n",
		"library/.git":             "gitdir: ../.git/modules/library\n",
		"library/l.ts":             "export const l = 1;\n",
		"tool.go":                  "package tool\n",
	})
	for link, target := range map[string]string{
		"source/alias.ts": filepath.Join(root, "a.ts"),
		"zone.js":         filepath.Join(root, "source", "deep"),
		"dangling.ts":     filepath.Join(root, "missing.ts"),
	} {
		if err := os.Symlink(target, filepath.Join(root, link)); err != nil {
			t.Skipf("this filesystem cannot make symbolic links: %v", err)
		}
	}
	return root
}

// diskListings is every directory under root as os.ReadDir lists it, ignored or not, the way discovery keeps
// them: by absolute path.
func diskListings(t *testing.T, root string) map[string][]os.DirEntry {
	t.Helper()
	listings := map[string][]os.DirEntry{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || !entry.IsDir() {
			return err
		}
		entries, err := os.ReadDir(path)
		listings[path] = entries
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return listings
}

func listingOf(listings map[string][]os.DirEntry) func(directory string) ([]os.DirEntry, bool) {
	return func(directory string) ([]os.DirEntry, bool) {
		entries, listed := listings[filepath.Clean(directory)]
		return entries, listed
	}
}

// without is a copy of listings with name removed from directory's listing.
func without(listings map[string][]os.DirEntry, directory string, name string) map[string][]os.DirEntry {
	copied := make(map[string][]os.DirEntry, len(listings))
	for key, entries := range listings {
		copied[key] = entries
	}
	var kept []os.DirEntry
	for _, entry := range listings[directory] {
		if entry.Name() != name {
			kept = append(kept, entry)
		}
	}
	copied[directory] = kept
	return copied
}

// A walk read from listings is the walk of the disk (#g3046x5): every field of the Enumeration, with every
// directory listed, with half of them read from the disk instead, and with none. Each listing the walk reads
// decides something, so dropping one entry from a listing changes the answer: a file, a `.gitignore`, a
// `.git`. Those three mutants are the proof the listings are read at all.
func TestAWalkReadFromListingsIsTheWalkOfTheDisk(t *testing.T) {
	t.Parallel()
	root := listedTree(t)
	disk, err := Enumerate(root, handlesEveryLanguage)
	if err != nil {
		t.Fatal(err)
	}
	if len(disk.Files) < 4 || disk.SymbolicLinks != 3 || len(disk.NestedRepositories) != 2 || len(disk.Adamic) != 1 {
		t.Fatalf("the fixture's walk is too thin to prove anything: %+v", disk)
	}
	listings := diskListings(t, root)

	// Reversed, so a listing that is not sorted by name is read in Walk's order all the same.
	reversed := map[string][]os.DirEntry{}
	for directory, entries := range listings {
		backwards := make([]os.DirEntry, len(entries))
		for index, entry := range entries {
			backwards[len(entries)-1-index] = entry
		}
		reversed[directory] = backwards
	}
	half := map[string][]os.DirEntry{}
	index := 0
	for directory, entries := range listings {
		if index%2 == 0 {
			half[directory] = entries
		}
		index++
	}
	for name, listing := range map[string]map[string][]os.DirEntry{
		"every directory listed": listings,
		"half of them listed":    half,
		"none listed":            {},
		"every listing reversed": reversed,
	} {
		listed, err := EnumerateListed(root, handlesEveryLanguage, listingOf(listing))
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if !reflect.DeepEqual(listed, disk) {
			t.Errorf("%s: the listed walk differs from the walk of the disk\nlisted: %+v\ndisk:   %+v", name, listed, disk)
		}
	}

	for name, mutant := range map[string]map[string][]os.DirEntry{
		"a file dropped":         without(listings, filepath.Join(root, "source"), "b.ts"),
		"a .gitignore dropped":   without(listings, filepath.Join(root, "source"), ".gitignore"),
		"a .git dropped":         without(listings, filepath.Join(root, "vendor"), ".git"),
		"a gitlink dropped":      without(listings, filepath.Join(root, "library"), ".git"),
		"a directory dropped":    without(listings, root, "source"),
		"an Adamic file dropped": without(listings, filepath.Join(root, "source"), "x.a"),
	} {
		listed, err := EnumerateListed(root, handlesEveryLanguage, listingOf(mutant))
		if err == nil && reflect.DeepEqual(listed, disk) {
			t.Errorf("%s: the listed walk still equals the disk's, so it did not read that listing", name)
		}
	}
}
