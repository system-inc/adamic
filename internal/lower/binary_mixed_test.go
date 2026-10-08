package lower

import (
	"errors"
	"strings"
	"testing"
)

// Null and undefined share no boxed representation yet. Preserve that explicit boundary.
func TestMixedLogicalBoxedNullBoundary(t *testing.T) {
	_, err := lowerSource(t, `function choose(value: RegExpExecArray | null): RegExpExecArray | true { return value || true; }
`)
	var notYet *NotYet
	if !errors.As(err, &notYet) || !strings.Contains(notYet.What, "boxed null distinct from undefined") {
		t.Fatalf("expected explicit boxed-null boundary, got %v", err)
	}
}
