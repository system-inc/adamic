package repowalk_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/system-inc/adamic/internal/adamiclint/repowalk"
)

// The fixtures under testdata/src carry a // want on every line the rule must flag, and analysistest fails on a
// finding no want expects as much as on a want nothing found, so the allowed shapes prove the rule stays quiet.
func TestRepoWalk(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), repowalk.Analyzer, "flagged", "allowed",
		"github.com/system-inc/adamic/productwalk")
}
