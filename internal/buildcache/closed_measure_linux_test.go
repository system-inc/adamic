//go:build linux && !mips && !mipsle && !mips64 && !mips64le && closurewalk

package buildcache

import (
	"os"
	"testing"
	"time"
)

// A measure, not a check: how many entries the no-write proof's walk visits over a tree, and how long it takes, with
// the machine's own owner and access reads. The process is given a uid no file has, so no entry is refused as owned,
// and every writable answer is asked and then set aside, so the walk visits everything. Skipped unless
// ADAMIC_MEASURE_CLOSURE_WALK names the tree to walk: from the repository's root,
//
//	ADAMIC_MEASURE_CLOSURE_WALK=$PWD go test -tags closurewalk -run '^TestMeasureTheClosureWalk$' -v ./internal/buildcache
//
// The tree comes from the variable, not from walking up to go.mod, so the test reaches outside its package only when
// asked to (repowalk). It builds only under the closurewalk tag, as a measurement, never a pin, so the lane's skip
// census never meets its skip (internal/lower's loose_slots_census_test.go is the same).
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
