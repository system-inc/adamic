package oracle

import (
	"errors"
	"github.com/system-inc/adamic/internal/lower"
	"path/filepath"
	"strings"
	"testing"
)

func TestGeneratorCycleFrame(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/notyet/generator_cycle.a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 {
		t.Fatalf("Node: %+v", truth)
	}
	program, err := lowered(t, path)
	if err != nil {
		var refused *lower.Refused
		if !errors.As(err, &refused) || !strings.Contains(refused.What, "cycle-capable generator frame slot saved without weak") {
			t.Fatal(err)
		}
		return
	}
	// A mutant admitting the slot must reach the leak witness rather than failing
	// compilation. Node finishes; the suspended counted frame cannot free itself.
	got, binary := nativelyUncached(t, program)
	if diff := disagreement(truth, got); diff != "" {
		t.Fatal(diff)
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
	t.Fatal("cycle-capable frame admitted without weak")
}
