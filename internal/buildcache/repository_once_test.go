package buildcache

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
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

// watchWaiters has waitingOnRepository call waiting with each asker's count as it parks on another's reading, for the
// rest of t.
func watchWaiters(t *testing.T, waiting func(count int64)) {
	t.Helper()
	var waiters atomic.Int64
	original := waitingOnRepository
	waitingOnRepository = func() { waiting(waiters.Add(1)) }
	t.Cleanup(func() { waitingOnRepository = original })
}

// waitFor waits on ready for at most 30 seconds, failing t if it never closes, so a broken wait can't hang the test.
func waitFor(t *testing.T, ready <-chan struct{}, what string) {
	t.Helper()
	select {
	case <-ready:
	case <-time.After(30 * time.Second):
		t.Errorf("%s never happened", what)
	}
}

// Every asker waiting on a reading under way wakes to that reading's answer, never to the nil its run holds before
// then, which Tracked reads as a directory with no repository, everything tracked: a silent wrong answer. The listing
// holds until every other asker has parked on it, so each of them was waiting before the answer was in.
// Mutant: readRepositoryRun frees the run (closes done) before its reading, so the waiters wake to nil.
// Not parallel: listRepository, waitingOnRepository and repositories are the process's.
func TestEveryWaiterWakesToTheReadingsAnswer(t *testing.T) {
	const askers = 16
	listings := countListings(t, nil)
	listing := listRepository
	parked := make(chan struct{})
	watchWaiters(t, func(count int64) {
		if count == askers-1 {
			close(parked)
		}
	})
	listRepository = func(directory string) ([]byte, []byte, error) {
		waitFor(t, parked, "every other asker waiting on the reading")
		return listing(directory)
	}
	directory := fakeCheckout(t)
	answers := make([]*trackedFiles, askers)
	var asking sync.WaitGroup
	for index := range askers {
		asking.Go(func() { answers[index] = repositoryAt(directory) })
	}
	asking.Wait()
	for index, files := range answers {
		if files == nil || files.all || !files.paths["a.txt"] || len(files.paths) != 1 {
			t.Fatalf("asker %d of %d got %+v, want the listing's one tracked file, a.txt", index, askers, files)
		}
		if files != answers[0] {
			t.Fatalf("asker %d got another answer than asker 0's, want the one reading's", index)
		}
	}
	if got := listings.Load(); got != 1 {
		t.Fatalf("one checkout was listed %d times, want once", got)
	}
}

// A reading that panics frees the asker waiting on it and is never kept: the waiter asks again and gets the next
// reading's answer, and so does a later asker, where a run left unfreed would hold them forever (the timeout says so).
// Mutant: readRepositoryRun's free not deferred, so the panic leaves done open and the waiter hangs.
// Not parallel: listRepository, waitingOnRepository and repositories are the process's.
func TestAReadingThatPanicsFreesItsWaiters(t *testing.T) {
	listings := countListings(t, nil)
	listing := listRepository
	parked := make(chan struct{})
	watchWaiters(t, func(count int64) {
		if count == 1 {
			close(parked)
		}
	})
	var readings atomic.Int64
	listRepository = func(directory string) ([]byte, []byte, error) {
		if readings.Add(1) == 1 {
			waitFor(t, parked, "an asker waiting on the reading")
			panic("git's listing panicked")
		}
		return listing(directory)
	}
	directory := fakeCheckout(t)
	type answer struct {
		files     *trackedFiles
		recovered any
	}
	answers := make(chan answer, 2)
	for range 2 {
		go func() {
			var files *trackedFiles
			defer func() { answers <- answer{files, recover()} }()
			files = repositoryAt(directory)
		}()
	}
	var answered *trackedFiles
	panics := 0
	for range 2 {
		select {
		case got := <-answers:
			if got.recovered != nil {
				if got.recovered != "git's listing panicked" {
					t.Fatalf("an asker panicked with %v, want the listing's own panic", got.recovered)
				}
				panics++
				continue
			}
			if got.files == nil || got.files.all || !got.files.paths["a.txt"] {
				t.Fatalf("the asker waiting on a reading that panicked got %+v, want the next reading's a.txt", got.files)
			}
			answered = got.files
		case <-time.After(30 * time.Second):
			t.Fatal("an asker of a reading that panicked never got an answer: its waiters were never freed")
		}
	}
	if panics != 1 || answered == nil {
		t.Fatalf("%d of 2 askers panicked, want the reader alone", panics)
	}
	later := make(chan *trackedFiles, 1)
	go func() { later <- repositoryAt(directory) }()
	select {
	case files := <-later:
		if files != answered {
			t.Fatalf("a later asker got %+v, want the kept reading's answer", files)
		}
	case <-time.After(30 * time.Second):
		t.Fatal("a later asker never got an answer after a reading panicked")
	}
	if got := readings.Load(); got != 2 || listings.Load() != 1 {
		t.Fatalf("%d readings, %d answered, want the one that panicked and one more", got, listings.Load())
	}
}
