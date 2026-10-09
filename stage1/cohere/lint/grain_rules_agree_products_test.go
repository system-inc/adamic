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
	if time.Since(started) >= 60*time.Second {
		t.Fatal("RulesAgree oracle product over 60s")
	}
}
