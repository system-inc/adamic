package oracle

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func init() {
	for _, path := range []string{
		"internal/oracle/testdata/notyet_void_undefined_push.a",
		"internal/oracle/testdata/notyet_void_undefined_pop.a",
	} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path, true, false})
	}
}

// TypeScript allows a number-returning function in a void callback slot. Optional
// access adds undefined to the call type without erasing the number on Node.
func TestOptionalVoidValueStillNotYet(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/notyet_void_undefined_value.a"))
	if err != nil {
		t.Fatal(err)
	}
	observed := onNode(t, path)
	if observed.exitCode != 0 || string(observed.stdout) != "number\nundefined\n" || len(observed.stderr) != 0 {
		t.Fatalf("Node observation: %#v", observed)
	}
	_, err = lowered(t, path)
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) {
		t.Fatalf("want a named value-use stop, got %v", err)
	}
	t.Logf("preserved value-use stop: %v; Node prints %q", err, observed.stdout)
}
