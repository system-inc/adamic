package buildcache

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
)

// countListings has listRepository count each directory's git listing for the rest of t, answering one tracked file
// (a.txt, clean), or failing every one when fail says so, so no git runs.
func countListings(t *testing.T, fail func() bool) *atomic.Int64 {
	t.Helper()
	var listings atomic.Int64
	original := listRepository
	listRepository = func(directory string) ([]byte, []byte, error) {
		listings.Add(1)
		if fail != nil && fail() {
			return nil, nil, errors.New("git couldn't list it")
		}
		return []byte("H 100644 " + "e69de29bb2d1d6434b8b29ae775ad8c2e48c5391" + " 0\ta.txt\x00"), nil, nil
	}
	t.Cleanup(func() { listRepository = original })
	repositories.Range(func(key, _ any) bool { repositories.Delete(key); return true })
	t.Cleanup(func() { repositories.Range(func(key, _ any) bool { repositories.Delete(key); return true }) })
	return &listings
}

// fakeCheckout is a directory with a .git, which repositoryAt reads as a checkout and lists through listRepository.
func fakeCheckout(t *testing.T) string {
	t.Helper()
	directory := t.TempDir()
	if err := os.Mkdir(filepath.Join(directory, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	return directory
}

// Parallel askers of one checkout wait on one git listing (#npd568j: git's diff-files stats every entry of
// cohere/TypeScript's index, and a product test's racing askers each ran their own, three in one test).
// Mutant: repositoryAt's wait dropped, so each asker lists it again.
// Not parallel: listRepository and repositories are the process's.
func TestOneCheckoutIsListedOnceHoweverManyAsk(t *testing.T) {
	listings := countListings(t, nil)
	directory := fakeCheckout(t)
	var asking sync.WaitGroup
	for range 32 {
		asking.Go(func() { repositoryAt(directory) })
	}
	asking.Wait()
	for range 100 {
		repositoryAt(directory)
	}
	if got := listings.Load(); got != 1 {
		t.Fatalf("one checkout was listed %d times, want once", got)
	}
}

// A checkout asked by a relative path, its absolute path and a link to it is one checkout, listed once.
// Not parallel: listRepository, repositories and the working directory are the process's.
func TestOneCheckoutByThreeNamesIsListedOnce(t *testing.T) {
	listings := countListings(t, nil)
	directory := fakeCheckout(t)
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(directory, link); err != nil {
		t.Fatal(err)
	}
	t.Chdir(filepath.Dir(directory))
	for _, name := range []string{filepath.Base(directory), directory, link, directory + string(filepath.Separator)} {
		repositoryAt(name)
	}
	if got := listings.Load(); got != 1 {
		t.Fatalf("one checkout by four spellings was listed %d times, want once", got)
	}
}

// A listing git failed is answered to the askers waiting on it, as every file tracked, and never kept: the next asker
// lists it again. Mutant: a failed reading kept, so the second asker is answered from it.
// Not parallel: listRepository and repositories are the process's.
func TestAFailedListingIsNeverKept(t *testing.T) {
	failing := true
	listings := countListings(t, func() bool { return failing })
	directory := fakeCheckout(t)
	if files := repositoryAt(directory); files == nil || !files.all {
		t.Fatalf("a failed listing gave %+v, want every file tracked", files)
	}
	failing = false
	if files := repositoryAt(directory); files == nil || files.all {
		t.Fatalf("the next asker after a failure got %+v, want a fresh listing", files)
	}
	if got := listings.Load(); got != 2 {
		t.Fatalf("a failed listing and the next were %d listings, want 2", got)
	}
}
