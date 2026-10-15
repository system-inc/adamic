package buildcache

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
)

// testReadsRig is a rig whose run journals this process afresh, as a traced test's process always does: the journal's
// process entry, which names the recipe, is written once per process, and an earlier test in this one may have taken it.
func testReadsRig(t *testing.T) *rig {
	t.Helper()
	r := newRig(t)
	journalProcess = sync.Once{}
	t.Setenv("ADAMIC_TRACE_TEST_KEY", testReadSetKey)
	return r
}

// testReadsReport is the settlement's read set line, or nil.
func testReadsReport(settlement Settlement) *Report {
	for index := range settlement.Reports {
		if settlement.Reports[index].Outcome == ReportReadSet {
			return &settlement.Reports[index]
		}
	}
	return nil
}

// testReadSetKey is a test read set's name key as Loom's tree builder would name it (#sfqdxvy).
var testReadSetKey = strings.Repeat("7", 64)

// findTestReads is cmd/find's answer for the test read set, as a new process asks it, holding nothing an earlier
// lookup read.
func findTestReads(t *testing.T) Finding {
	t.Helper()
	forgetHeldForLife()
	forgetTracked()
	findings, err := Find([]string{testReadSetKey})
	if err != nil {
		t.Fatal(err)
	}
	return findings[0]
}

// A test read set is the whole run's reads (#sfqdxvy): what the test read outside any build (its data file), and what
// its build read through its product's key (#mmpt615). cmd/find finds it on an unchanged tree, still finds it after a change to a file the run never
// read, and misses it after a change to either file it did read, so a test that only found its products is run again
// exactly when something it read changed. Mutant that fails it: the set without the reads outside every build (notes.txt
// dropped), which finds it after the data file changed.
// Not parallel: newRig.
func TestATestReadSetHoldsTheWholeRunsReads(t *testing.T) {
	r := testReadsRig(t)
	Product(t, port, building("built"))
	r.open(r.pid, "notes.txt")
	r.window(r.built(t)["port"].Build, func() { r.open(r.pid, "source/a.c") })
	r.add(r.pid, "+++ exited with 0 +++")
	settlement := r.settle(t)
	if len(settlement.Settled) != 1 || !strings.HasPrefix(settlement.TestReads, "test read set ") {
		t.Fatalf("settled %q, test reads %q", settlement.Settled, settlement.TestReads)
	}
	t.Setenv("ADAMIC_BUILD_TRACE", "")
	finding := findTestReads(t)
	if finding.Outcome != FindingFound || finding.Name != TestReadSetName {
		t.Fatalf("an unchanged tree: %+v", finding)
	}
	// The report names the set under the key cmd/find finds it by, for Loom to vouch with it and no other set.
	if line := testReadsReport(settlement); line == nil || line.NameKey != testReadSetKey || line.Key != finding.Key || line.Reads == 0 {
		t.Fatalf("the report's read set line is %+v, want %s under %s", line, testReadSetKey, finding.Key)
	}
	write(t, r.root, "include/x.h", "#define X 2\n")
	if finding := findTestReads(t); finding.Outcome != FindingFound {
		t.Fatalf("a change to a file the run never read: %+v", finding)
	}
	for _, read := range []struct{ path, content string }{{"notes.txt", "another note\n"}, {"source/a.c", "int a2;\n"}} {
		gitIn(t, r.root, "checkout", "-q", "--", ".")
		write(t, r.root, read.path, read.content)
		if finding := findTestReads(t); finding.Outcome != FindingMissing || !strings.Contains(finding.Why, read.path) {
			t.Fatalf("a change to %s, which the run read: %+v", read.path, finding)
		}
	}
}

// Only a run that can stand for the test's next one records a test read set: a command that failed or was killed
// records none, and says why, so its test runs on every tree (Builder's review of #sfqdxvy). Mutant that fails it: a
// set recorded whatever the command's status.
// Not parallel: newRig.
func TestAFailedRunRecordsNoTestReadSet(t *testing.T) {
	for _, exit := range []string{"+++ exited with 2 +++", "+++ killed by SIGKILL +++"} {
		t.Run(exit, func(t *testing.T) {
			r := testReadsRig(t)
			Product(t, port, building("built"))
			r.open(r.pid, "notes.txt")
			r.window(r.built(t)["port"].Build, func() { r.open(r.pid, "source/a.c") })
			r.add(r.pid, "%s", exit)
			if settlement := r.settle(t); !strings.HasPrefix(settlement.TestReads, "no test read set: ") || testReadsReport(settlement) != nil {
				t.Fatalf("a run that %s: %q, report %+v", exit, settlement.TestReads, testReadsReport(settlement))
			}
			t.Setenv("ADAMIC_BUILD_TRACE", "")
			if finding := findTestReads(t); finding.Outcome != FindingMissing {
				t.Fatalf("a run that %s left a test read set: %+v", exit, finding)
			}
		})
	}
}

// The key Loom names a test read set by is no input to any product: a product's name key is the same with it set or
// not, so naming the set moves no product (Builder's review of #sfqdxvy). Mutant that fails it: the variable read
// into a name key.
// Not parallel: newRig.
func TestTheTestReadSetKeyKeysNoProduct(t *testing.T) {
	r := newRig(t)
	t.Setenv("ADAMIC_TRACE_TEST_KEY", "")
	without := NameKey(r.root, port)
	t.Setenv("ADAMIC_TRACE_TEST_KEY", testReadSetKey)
	if with := NameKey(r.root, port); with != without {
		t.Fatalf("ADAMIC_TRACE_TEST_KEY moved a name key: %s, without it %s", with, without)
	}
}

// A process the command started that outlives it still reads for the run: what it reads after the command's exit is in
// the test read set too (Mentat's review of 8.8). Mutant that fails it: an orphan's reads dropped when it has no live
// ancestor, so the edit to what only it read stays found.
// Not parallel: newRig.
func TestATestReadSetHoldsWhatAProcessReadAfterTheCommandExited(t *testing.T) {
	r := testReadsRig(t)
	Product(t, port, building("built"))
	child := r.pid + 1
	r.add(r.pid, `clone(child_stack=NULL, flags=SIGCHLD) = %d`, child)
	r.window(r.built(t)["port"].Build, func() { r.open(r.pid, "source/a.c") })
	r.add(r.pid, "+++ exited with 0 +++")
	r.open(child, "include/x.h")
	r.add(child, "+++ exited with 0 +++")
	if settlement := r.settle(t); !strings.HasPrefix(settlement.TestReads, "test read set ") {
		t.Fatalf("test reads %q", settlement.TestReads)
	}
	t.Setenv("ADAMIC_BUILD_TRACE", "")
	if finding := findTestReads(t); finding.Outcome != FindingFound {
		t.Fatalf("an unchanged tree: %+v", finding)
	}
	write(t, r.root, "include/x.h", "#define X 2\n")
	if finding := findTestReads(t); finding.Outcome != FindingMissing || !strings.Contains(finding.Why, "include/x.h") {
		t.Fatalf("a change to what only the outliving process read: %+v", finding)
	}
}

// A test read set's marker is placed by a rename, never written in place: a seeded cache links its seed's files, so a
// write through the link would change the seed's (Mentat's re-review of #sfqdxvy). Mutant that fails it: the marker
// written with os.WriteFile.
func TestATestReadSetsMarkerNeverWritesThroughALink(t *testing.T) {
	t.Parallel()
	seed, cache := t.TempDir(), t.TempDir()
	key := strings.Repeat("8", 64)
	seeded := filepath.Join(seed, "test-read-set")
	if err := os.WriteFile(seeded, []byte("the seed's\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(cache, key), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Link(seeded, filepath.Join(cache, key, "test-read-set")); err != nil {
		t.Fatal(err)
	}
	marker := placeTestMarker(cache, key, testReadSetKey)
	if content, _ := os.ReadFile(seeded); string(content) != "the seed's\n" {
		t.Fatalf("the seed's file reads %q after the marker was placed", content)
	}
	if content, _ := os.ReadFile(filepath.Join(marker, "test-read-set")); string(content) != testReadSetKey+"\n" {
		t.Fatalf("the marker reads %q", content)
	}
}

// stat is a stat of name that found it.
func (r *rig) stat(pid int, name string) {
	r.add(pid, `newfstatat(AT_FDCWD, %q, {st_mode=S_IFREG|0644, st_size=7, ...}, 0) = 0`, r.path(name))
}

// testSet is the test read set the settlement recorded: its key and its entries, newest first.
func testSet(t *testing.T, r *rig, settlement Settlement) (string, []string) {
	t.Helper()
	line := testReadsReport(settlement)
	if line == nil {
		t.Fatalf("no test read set: %q", settlement.TestReads)
	}
	return line.Key, r.entries(t, testReadSetKey)
}

// A test that built its product and one that found it record one test read set (#mmpt615, Node's ruling): the build's
// reads are its product's, named by the product's key, never the test's, so the test's own stat of a file its build
// read in full is the same entry either way. #s57yq6y's A/B moved 111 test sets on exactly this, each by its products'
// reads alone. Mutant that fails it: the test set taking every read of the run, windows included.
// Not parallel: newRig.
func TestATestReadSetIsTheSameWhetherItsProductWasBuiltOrFound(t *testing.T) {
	r := testReadsRig(t)
	built := Product(t, port, building("built"))
	r.stat(r.pid, "source/a.c")
	r.window(r.built(t)["port"].Build, func() { r.open(r.pid, "source/a.c"); r.open(r.pid, "include/x.h") })
	r.open(r.pid, filepath.Join(built, "product"))
	r.add(r.pid, "+++ exited with 0 +++")
	builtKey, builtEntries := testSet(t, r, r.settle(t))
	r.retrace(t, "found")
	journalProcess = sync.Once{}
	forgetHeldForLife()
	found := Product(t, port, building("built"))
	if strings.Contains(found, "pending") {
		t.Fatalf("the second run built port again, at %s", found)
	}
	r.stat(r.pid, "source/a.c")
	r.open(r.pid, filepath.Join(found, "product"))
	r.add(r.pid, "+++ exited with 0 +++")
	foundKey, foundEntries := testSet(t, r, r.settle(t))
	if foundKey != builtKey {
		t.Fatalf("the found run's test set is %s, the built run's %s:\nbuilt %q\nfound %q", foundKey[:12], builtKey[:12], builtEntries, foundEntries)
	}
	if !slices.Contains(builtEntries, "product key "+NameKey(r.root, port)) || slices.Contains(builtEntries, "content include/x.h") {
		t.Fatalf("the built run's test set %q: want port by its key, and none of its build's reads", builtEntries)
	}
}

// A product a test used stands in its test read set by its key, so the test's key moves with every file the product's
// build read though the test's own reads don't (Node's guard on #mmpt615), whether the run built the product or found
// it. By its key and never its bytes: the window a test set leaves out holds every read of its owner's process, a
// parallel subtest's too, so a product rebuilt to the same bytes from a changed file still runs the test. Mutants that
// each fail it: the entry valued by the product's bytes; no entry for a product the run built; a found product left as
// a product entry, by its bytes.
// Not parallel: newRig.
func TestATestReadSetMovesWithItsProductsInputsByItsKey(t *testing.T) {
	for _, used := range []string{"built", "found"} {
		t.Run(used, func(t *testing.T) {
			r := testReadsRig(t)
			if used == "found" {
				// An earlier run builds port, recording no test set, and the test's run finds it and reads it.
				t.Setenv("ADAMIC_TRACE_TEST_KEY", "")
				Product(t, port, building("built"))
				r.window(r.built(t)["port"].Build, func() { r.open(r.pid, "source/a.c"); r.open(r.pid, "include/x.h") })
				r.add(r.pid, "+++ exited with 0 +++")
				r.settle(t)
				r.retrace(t, "test")
				t.Setenv("ADAMIC_TRACE_TEST_KEY", testReadSetKey)
				journalProcess = sync.Once{}
				forgetHeldForLife()
				r.open(r.pid, filepath.Join(Product(t, port, building("built")), "product"))
			} else {
				// The run builds port and never reads its directory after, so only the build stands port in the set.
				Product(t, port, building("built"))
				r.window(r.built(t)["port"].Build, func() { r.open(r.pid, "source/a.c"); r.open(r.pid, "include/x.h") })
			}
			r.add(r.pid, "+++ exited with 0 +++")
			testSet(t, r, r.settle(t))
			t.Setenv("ADAMIC_BUILD_TRACE", "")
			if finding := findTestReads(t); finding.Outcome != FindingFound {
				t.Fatalf("an unchanged tree: %+v", finding)
			}
			// include/x.h, which only port's build read, changes, and another traced run builds port again to the same
			// bytes.
			write(t, r.root, "include/x.h", "#define X 2\n")
			gitIn(t, r.root, "commit", "-q", "-am", "x")
			forgetTracked()
			r.retrace(t, "rebuilt")
			t.Setenv("ADAMIC_TRACE_TEST_KEY", "")
			journalProcess = sync.Once{}
			forgetHeldForLife()
			Product(t, port, building("built"))
			r.window(r.built(t)["port"].Build, func() { r.open(r.pid, "source/a.c"); r.open(r.pid, "include/x.h") })
			r.add(r.pid, "+++ exited with 0 +++")
			if settlement := r.settle(t); len(settlement.Settled) != 1 {
				t.Fatalf("port wasn't settled again: %q, refused %q", settlement.Settled, settlement.Refused)
			}
			t.Setenv("ADAMIC_BUILD_TRACE", "")
			if finding := findTestReads(t); finding.Outcome != FindingMissing || !strings.Contains(finding.Why, "product key") {
				t.Fatalf("port rebuilt to the same bytes from a changed include/x.h, which only its build read: %+v", finding)
			}
		})
	}
}

// A build that doesn't settle has no key to stand for its window, so the window's reads stay in the test read set: a
// read no key can name holds the test's set back as it holds the product's (#mmpt615). Mutant that fails it: a
// refused build's window left out, which records a test set missing what the run read.
// Not parallel: newRig.
func TestARefusedBuildsReadsStayInTheTestReadSet(t *testing.T) {
	r := testReadsRig(t)
	write(t, r.root, "scratch.txt", "untracked\n")
	Product(t, port, building("built"))
	r.window(r.built(t)["port"].Build, func() { r.open(r.pid, "source/a.c"); r.open(r.pid, "scratch.txt") })
	r.add(r.pid, "+++ exited with 0 +++")
	settlement := r.settle(t)
	if len(settlement.Refused) != 1 {
		t.Fatalf("port read an untracked file and wasn't refused: settled %q", settlement.Settled)
	}
	if !strings.HasPrefix(settlement.TestReads, "no test read set: ") || !strings.Contains(settlement.TestReads, "scratch.txt") || testReadsReport(settlement) != nil {
		t.Fatalf("a refused build's untracked read: test reads %q", settlement.TestReads)
	}
}
