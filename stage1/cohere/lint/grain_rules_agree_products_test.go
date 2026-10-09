package lint

import (
	"testing"
	"time"
)

// C emission exceeds the build phase's 60s grain. Shards still prepare
// lowered/native products themselves, before starting their own deadlines.
func TestProduct_RulesAgreeOracle(t *testing.T) {
	t.Parallel()
	started := time.Now()
	rulesAgreeOracle(t)
	if time.Since(started) >= 60*time.Second {
		t.Fatal("RulesAgree oracle product over 60s")
	}
}

func TestProduct_RulesAgreeCapture(t *testing.T) {
	t.Parallel()
	started := time.Now()
	rulesAgreeCapture(t)
	if time.Since(started) >= 60*time.Second {
		t.Fatal("RulesAgree capture product over 60s")
	}
}
