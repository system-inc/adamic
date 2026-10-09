package lint

import (
	"testing"
	"time"
)

// C emission and upstream capture exceed the build phase's 60s grain
// (#c5k975w, #3he8f8g). Shards still prepare those products themselves,
// before starting their own deadlines.
func TestProduct_RulesAgreeOracle(t *testing.T) {
	t.Parallel()
	started := time.Now()
	rulesAgreeOracle(t)
	t.Logf("%s: %.3fs", t.Name(), time.Since(started).Seconds())
}
