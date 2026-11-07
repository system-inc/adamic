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
// process entry, which names the recipe, is written once per process, and an earlier test in this one may have taken
// it.
func testReadsRig(test *testing.T) *rig {
	test.Helper()
	rig := newRig(test)
	journalProcess = sync.Once{}
	test.Setenv("ADAMIC_TRACE_TEST_KEY", testReadSetKey)
	return rig
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
func findTestReads(test *testing.T) Finding {
	test.Helper()
	forgetHeldForLife()
	forgetTracked()
	findings, findError := Find([]string{testReadSetKey})
	if findError != nil {
		test.Fatal(findError)
	}
	return findings[0]
}

// A test read set is the whole run's reads (#sfqdxvy): what the test read outside any build (its data file), and what
// its build read through its product's key (#mmpt615). cmd/find finds it on an unchanged tree, still finds it after a
// change to a file the run never read, and misses it after a change to either file it did read, so a test that only
// found its products is run again exactly when something it read changed. Mutant that fails it: the set without the
// reads outside every build (notes.txt dropped), which finds it after the data file changed.
// Not parallel: newRig.
func TestATestReadSetHoldsTheWholeRunsReads(test *testing.T) {
	rig := testReadsRig(test)
	Product(test, port, building("built"))
	rig.open(rig.pid, "notes.txt")
	rig.window(rig.built(test)["port"].Build, func() { rig.open(rig.pid, "source/a.c") })
	rig.add(rig.pid, "+++ exited with 0 +++")
	settlement := rig.settle(test)
	if len(settlement.Settled) != 1 || !strings.HasPrefix(settlement.TestReads, "test read set ") {
		test.Fatalf("settled %q, test reads %q", settlement.Settled, settlement.TestReads)
	}
	test.Setenv("ADAMIC_BUILD_TRACE", "")
	finding := findTestReads(test)
	if finding.Outcome != FindingFound || finding.Name != TestReadSetName {
		test.Fatalf("an unchanged tree: %+v", finding)
	}
	// The report names the set under the key cmd/find finds it by, for Loom to vouch with it and no other set.
	if line := testReadsReport(
		settlement,
	); line == nil || line.NameKey != testReadSetKey || line.Key != finding.Key || line.Reads == 0 {
		test.Fatalf("the report's read set line is %+v, want %s under %s", line, testReadSetKey, finding.Key)
	}
	write(test, rig.root, "include/x.h", "#define X 2\n")
	if finding := findTestReads(test); finding.Outcome != FindingFound {
		test.Fatalf("a change to a file the run never read: %+v", finding)
	}
	for _, read := range []struct{ path, content string }{
		{"notes.txt", "another note\n"},
		{"source/a.c", "int a2;\n"},
	} {
		gitIn(test, rig.root, "checkout", "-q", "--", ".")
		write(test, rig.root, read.path, read.content)
		if finding := findTestReads(
			test,
		); finding.Outcome != FindingMissing || !strings.Contains(finding.Why, read.path) {
			test.Fatalf("a change to %s, which the run read: %+v", read.path, finding)
		}
	}
}

// Only a run that can stand for the test's next one records a test read set: a command that failed or was killed
// records none, and says why, so its test runs on every tree (Builder's review of #sfqdxvy). Mutant that fails it: a
// set recorded whatever the command's status.
// Not parallel: newRig.
func TestAFailedRunRecordsNoTestReadSet(test *testing.T) {
	for _, exit := range []string{"+++ exited with 2 +++", "+++ killed by SIGKILL +++"} {
		test.Run(exit, func(test *testing.T) {
			rig := testReadsRig(test)
			Product(test, port, building("built"))
			rig.open(rig.pid, "notes.txt")
			rig.window(rig.built(test)["port"].Build, func() { rig.open(rig.pid, "source/a.c") })
			rig.add(rig.pid, "%s", exit)
			if settlement := rig.settle(
				test,
			); !strings.HasPrefix(settlement.TestReads, "no test read set: ") || testReadsReport(settlement) != nil {
				test.Fatalf("a run that %s: %q, report %+v", exit, settlement.TestReads, testReadsReport(settlement))
			}
			test.Setenv("ADAMIC_BUILD_TRACE", "")
			if finding := findTestReads(test); finding.Outcome != FindingMissing {
				test.Fatalf("a run that %s left a test read set: %+v", exit, finding)
			}
		})
	}
}

// The key Loom names a test read set by is no input to any product: a product's name key is the same with it set or
// not, so naming the set moves no product (Builder's review of #sfqdxvy). Mutant that fails it: the variable read
// into a name key.
// Not parallel: newRig.
func TestTheTestReadSetKeyKeysNoProduct(test *testing.T) {
	rig := newRig(test)
	test.Setenv("ADAMIC_TRACE_TEST_KEY", "")
	without := NameKey(rig.root, port)
	test.Setenv("ADAMIC_TRACE_TEST_KEY", testReadSetKey)
	if with := NameKey(rig.root, port); with != without {
		test.Fatalf("ADAMIC_TRACE_TEST_KEY moved a name key: %s, without it %s", with, without)
	}
}

// A process the command started that outlives it still reads for the run: what it reads after the command's exit is in
// the test read set too (Mentat's review of 8.8). Mutant that fails it: an orphan's reads dropped when it has no live
// ancestor, so the edit to what only it read stays found.
// Not parallel: newRig.
func TestATestReadSetHoldsWhatAProcessReadAfterTheCommandExited(test *testing.T) {
	rig := testReadsRig(test)
	Product(test, port, building("built"))
	child := rig.pid + 1
	rig.add(rig.pid, `clone(child_stack=NULL, flags=SIGCHLD) = %d`, child)
	rig.window(rig.built(test)["port"].Build, func() { rig.open(rig.pid, "source/a.c") })
	rig.add(rig.pid, "+++ exited with 0 +++")
	rig.open(child, "include/x.h")
	rig.add(child, "+++ exited with 0 +++")
	if settlement := rig.settle(test); !strings.HasPrefix(settlement.TestReads, "test read set ") {
		test.Fatalf("test reads %q", settlement.TestReads)
	}
	test.Setenv("ADAMIC_BUILD_TRACE", "")
	if finding := findTestReads(test); finding.Outcome != FindingFound {
		test.Fatalf("an unchanged tree: %+v", finding)
	}
	write(test, rig.root, "include/x.h", "#define X 2\n")
	if finding := findTestReads(
		test,
	); finding.Outcome != FindingMissing || !strings.Contains(finding.Why, "include/x.h") {
		test.Fatalf("a change to what only the outliving process read: %+v", finding)
	}
}

// A test read set's marker is placed by a rename, never written in place: a seeded cache links its seed's files, so a
// write through the link would change the seed's (Mentat's re-review of #sfqdxvy). Mutant that fails it: the marker
// written with os.WriteFile.
func TestATestReadSetsMarkerNeverWritesThroughALink(test *testing.T) {
	test.Parallel()
	seed, cache := test.TempDir(), test.TempDir()
	key := strings.Repeat("8", 64)
	seeded := filepath.Join(seed, "test-read-set")
	if writeError := os.WriteFile(seeded, []byte("the seed's\n"), 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	if createDirectoryError := os.MkdirAll(filepath.Join(cache, key), 0o755); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	if linkError := os.Link(seeded, filepath.Join(cache, key, "test-read-set")); linkError != nil {
		test.Fatal(linkError)
	}
	marker := placeTestMarker(cache, key, testReadSetKey)
	if content, _ := os.ReadFile(seeded); string(content) != "the seed's\n" {
		test.Fatalf("the seed's file reads %q after the marker was placed", content)
	}
	if content, _ := os.ReadFile(filepath.Join(marker, "test-read-set")); string(content) != testReadSetKey+"\n" {
		test.Fatalf("the marker reads %q", content)
	}
}

// stat is a stat of name that found it.
func (rig *rig) stat(pid int, name string) {
	rig.add(pid, `newfstatat(AT_FDCWD, %q, {st_mode=S_IFREG|0644, st_size=7, ...}, 0) = 0`, rig.path(name))
}

// testSet is the test read set the settlement recorded: its key and its entries, newest first.
func testSet(test *testing.T, rig *rig, settlement Settlement) (string, []string) {
	test.Helper()
	line := testReadsReport(settlement)
	if line == nil {
		test.Fatalf("no test read set: %q", settlement.TestReads)
	}
	return line.Key, rig.entries(test, testReadSetKey)
}

// A test that built its product and one that found it record one test read set (#mmpt615, Node's ruling): the build's
// reads are its product's, named by the product's key, never the test's, so the test's own stat of a file its build
// read in full is the same entry either way. #s57yq6y's A/B moved 111 test sets on exactly this, each by its products'
// reads alone. Mutant that fails it: the test set taking every read of the run, windows included.
// Not parallel: newRig.
func TestATestReadSetIsTheSameWhetherItsProductWasBuiltOrFound(test *testing.T) {
	rig := testReadsRig(test)
	built := Product(test, port, building("built"))
	rig.stat(rig.pid, "source/a.c")
	rig.window(
		rig.built(test)["port"].Build,
		func() { rig.open(rig.pid, "source/a.c"); rig.open(rig.pid, "include/x.h") },
	)
	rig.open(rig.pid, filepath.Join(built, "product"))
	rig.add(rig.pid, "+++ exited with 0 +++")
	builtKey, builtEntries := testSet(test, rig, rig.settle(test))
	rig.retrace(test, "found")
	journalProcess = sync.Once{}
	forgetHeldForLife()
	found := Product(test, port, building("built"))
	if strings.Contains(found, "pending") {
		test.Fatalf("the second run built port again, at %s", found)
	}
	rig.stat(rig.pid, "source/a.c")
	rig.open(rig.pid, filepath.Join(found, "product"))
	rig.add(rig.pid, "+++ exited with 0 +++")
	foundKey, foundEntries := testSet(test, rig, rig.settle(test))
	if foundKey != builtKey {
		test.Fatalf(
			"the found run's test set is %s, the built run's %s:\nbuilt %q\nfound %q",
			foundKey[:12],
			builtKey[:12],
			builtEntries,
			foundEntries,
		)
	}
	if !slices.Contains(builtEntries, "product key "+NameKey(rig.root, port)) ||
		slices.Contains(builtEntries, "content include/x.h") {
		test.Fatalf("the built run's test set %q: want port by its key, and none of its build's reads", builtEntries)
	}
}

// A product a test used stands in its test read set by its key, so the test's key moves with every file the product's
// build read though the test's own reads don't (Node's guard on #mmpt615), whether the run built the product or found
// it. By its key and never its bytes: the window a test set leaves out holds every read of its owner's process, a
// parallel subtest's too, so a product rebuilt to the same bytes from a changed file still runs the test. Mutants that
// each fail it: the entry valued by the product's bytes; no entry for a product the run built; a found product left as
// a product entry, by its bytes.
// Not parallel: newRig.
func TestATestReadSetMovesWithItsProductsInputsByItsKey(test *testing.T) {
	for _, used := range []string{"built", "found"} {
		test.Run(used, func(test *testing.T) {
			rig := testReadsRig(test)
			if used == "found" {
				// An earlier run builds port, recording no test set, and the test's run finds it and reads it.
				test.Setenv("ADAMIC_TRACE_TEST_KEY", "")
				Product(test, port, building("built"))
				rig.window(
					rig.built(test)["port"].Build,
					func() { rig.open(rig.pid, "source/a.c"); rig.open(rig.pid, "include/x.h") },
				)
				rig.add(rig.pid, "+++ exited with 0 +++")
				rig.settle(test)
				rig.retrace(test, "test")
				test.Setenv("ADAMIC_TRACE_TEST_KEY", testReadSetKey)
				journalProcess = sync.Once{}
				forgetHeldForLife()
				rig.open(rig.pid, filepath.Join(Product(test, port, building("built")), "product"))
			} else {
				// The run builds port and never reads its directory after, so only the build stands port in the set.
				Product(test, port, building("built"))
				rig.window(
					rig.built(test)["port"].Build,
					func() { rig.open(rig.pid, "source/a.c"); rig.open(rig.pid, "include/x.h") },
				)
			}
			rig.add(rig.pid, "+++ exited with 0 +++")
			testSet(test, rig, rig.settle(test))
			test.Setenv("ADAMIC_BUILD_TRACE", "")
			if finding := findTestReads(test); finding.Outcome != FindingFound {
				test.Fatalf("an unchanged tree: %+v", finding)
			}
			// include/x.h, which only port's build read, changes, and another traced run builds port again to the same
			// bytes.
			write(test, rig.root, "include/x.h", "#define X 2\n")
			gitIn(test, rig.root, "commit", "-q", "-am", "x")
			forgetTracked()
			rig.retrace(test, "rebuilt")
			test.Setenv("ADAMIC_TRACE_TEST_KEY", "")
			journalProcess = sync.Once{}
			forgetHeldForLife()
			Product(test, port, building("built"))
			rig.window(
				rig.built(test)["port"].Build,
				func() { rig.open(rig.pid, "source/a.c"); rig.open(rig.pid, "include/x.h") },
			)
			rig.add(rig.pid, "+++ exited with 0 +++")
			if settlement := rig.settle(test); len(settlement.Settled) != 1 {
				test.Fatalf("port wasn't settled again: %q, refused %q", settlement.Settled, settlement.Refused)
			}
			test.Setenv("ADAMIC_BUILD_TRACE", "")
			if finding := findTestReads(
				test,
			); finding.Outcome != FindingMissing || !strings.Contains(finding.Why, "product key") {
				test.Fatalf(
					"port rebuilt to the same bytes from a changed include/x.h, which only its build read: %+v",
					finding,
				)
			}
		})
	}
}

// A build that doesn't settle has no key to stand for its window, so the window's reads stay in the test read set: a
// read no key can name holds the test's set back as it holds the product's (#mmpt615). Mutant that fails it: a
// refused build's window left out, which records a test set missing what the run read.
// Not parallel: newRig.
func TestARefusedBuildsReadsStayInTheTestReadSet(test *testing.T) {
	rig := testReadsRig(test)
	write(test, rig.root, "scratch.txt", "untracked\n")
	Product(test, port, building("built"))
	rig.window(
		rig.built(test)["port"].Build,
		func() { rig.open(rig.pid, "source/a.c"); rig.open(rig.pid, "scratch.txt") },
	)
	rig.add(rig.pid, "+++ exited with 0 +++")
	settlement := rig.settle(test)
	if len(settlement.Refused) != 1 {
		test.Fatalf("port read an untracked file and wasn't refused: settled %q", settlement.Settled)
	}
	if !strings.HasPrefix(settlement.TestReads, "no test read set: ") ||
		!strings.Contains(settlement.TestReads, "scratch.txt") ||
		testReadsReport(settlement) != nil {
		test.Fatalf("a refused build's untracked read: test reads %q", settlement.TestReads)
	}
}
