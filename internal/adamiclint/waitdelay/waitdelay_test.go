package waitdelay_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/system-inc/adamic/internal/adamiclint/waitdelay"
)

// The fixtures under testdata/src carry a // want on every line the rule must flag, and analysistest fails on a
// finding no want expects as much as on a want nothing found, so the allowed shapes prove the rule stays quiet.
func TestWaitDelay(t *testing.T) {
	t.Parallel()
	analysistest.Run(errorsOnly{t}, analysistest.TestData(), waitdelay.Analyzer,
		"flagged", "allowed", "github.com/system-inc/adamic/childguardcalls",
		"github.com/system-inc/adamic/internal/childguard")
}

// errorsOnly hands analysistest.Run a reporter that isn't a testing.TB. Given a testing.TB, Run first proves the go
// command works by building a throwaway program, and Loom refuses a test that builds anything no key covers. Loading
// the fixtures needs only go list, a read-only query.
type errorsOnly struct{ t *testing.T }

func (reporter errorsOnly) Errorf(format string, arguments ...any) {
	reporter.t.Helper()
	reporter.t.Errorf(format, arguments...)
}
