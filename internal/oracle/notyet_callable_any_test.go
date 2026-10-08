package oracle

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/notyet_call_boundaries/notyet_callable_any.a", false, false})
}

// The table's timer callback returns any. Discarding it does not grant a proven
// return ABI; docs/0.1.md requires unknown and narrowing instead of accepting any.
func TestCallableAnyResultStaysNotYet(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/notyet_call_boundaries/notyet_callable_any.a"))
	if err != nil {
		t.Fatal(err)
	}
	observed := onNode(t, path)
	if observed.exitCode != 0 || string(observed.stdout) != "tick\n" || len(observed.stderr) != 0 {
		t.Fatalf("Node observation: %#v", observed)
	}
	_, err = lowered(t, path)
	var stop *lower.NotYet
	if !errors.As(err, &stop) || stop.What != "a call returning any" {
		t.Fatalf("want the unrepresented any return stop, got %v", err)
	}
}
