package oracle

import (
	"testing"

	"github.com/system-inc/adamic/internal/buildcache/testdata/oracle/helper"
)

// The helper package is imported only here, so only a key that lists the test's dependencies holds it.
func TestAnswer(t *testing.T) {
	if Answer() != helper.Expected {
		t.Fatalf("answer %d", Answer())
	}
}
