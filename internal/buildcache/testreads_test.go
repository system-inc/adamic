package buildcache

import (
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

// A test read set is the whole run's reads (#sfqdxvy): what the test read outside any build (its data file) as well as
// what its build read. cmd/find finds it on an unchanged tree, still finds it after a change to a file the run never
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
