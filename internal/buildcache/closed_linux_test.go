//go:build linux && !mips && !mipsle && !mips64 && !mips64le

package buildcache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// The kernel's own answer, asked live: a directory this process made is writable to it (the writable-directory case),
// and a file it made read-only, in a directory it made read-only, isn't. Skipped as root, whose access ignores mode
// bits.
// Mutant: writableByThisProcess's `case err == nil:` answering `return false, nil`: "a directory it made" goes red.
func TestTheKernelAnswersWhatThisProcessMayWrite(t *testing.T) {
	t.Parallel()
	if os.Geteuid() == 0 {
		t.Skip("running as root, whose access ignores mode bits, so nothing here is closed to it")
	}
	directory := t.TempDir()
	closed := filepath.Join(directory, "closed")
	write(t, closed, "file.c", "int file;\n")
	file := filepath.Join(closed, "file.c")
	if err := os.Chmod(file, 0o444); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(closed, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(closed, 0o755) })
	for _, each := range []struct {
		name, path string
		writable   bool
	}{
		{"a directory it made", directory, true},
		{"a read-only file in a read-only directory", file, false},
	} {
		if writable, err := writableByThisProcess(each.path); err != nil || writable != each.writable {
			t.Errorf("%s: writable %t (%v), want %t", each.name, writable, err, each.writable)
		}
	}
}

// A measure, not a check: how many entries the no-write proof's walk visits over a tree, and how long it takes, with
// the machine's own owner and access reads. The process is given a uid no file has, so no entry is refused as owned,
// and every writable answer is asked and then set aside, so the walk visits everything. Skipped unless
// ADAMIC_MEASURE_CLOSURE_WALK names the tree to walk: from the repository's root,
//
//	ADAMIC_MEASURE_CLOSURE_WALK=$PWD go test -run '^TestMeasureTheClosureWalk$' -v ./internal/buildcache
//
// The tree comes from the variable, not from walking up to go.mod, so the test reaches outside its package only when
// asked to (repowalk).
func TestMeasureTheClosureWalk(t *testing.T) {
	t.Parallel()
	root := os.Getenv("ADAMIC_MEASURE_CLOSURE_WALK")
	if root == "" {
		t.Skip("ADAMIC_MEASURE_CLOSURE_WALK names no tree to walk")
	}
	visited, asked := 0, 0
	reads := closureReads{
		// 4294967294 is -2 as a uid: one no file has.
		status: statusForTest(noCapabilities, noCapabilities, "4294967294\t4294967294\t4294967294\t4294967294", oneUser),
		owner: func(path string) (int, error) {
			visited++
			return ownerOf(path)
		},
		writable: func(path string) (bool, error) {
			asked++
			writableByThisProcess(path)
			return false, nil
		},
	}
	started := time.Now()
	_, err := noWriteProof(root, reads)
	t.Logf("walked %s: %d entries visited, %d asked whether writable, in %s (err %v)", root, visited, asked, time.Since(started), err)
}
