package oracle

import (
	"bytes"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/optional_function_presence.a", true, false})
}

// Keep the supplied probes intact. Their generic captures still hit the independent
// cycle-capable guard on main.
func TestOptionalFunctionMemoizeBoundary(t *testing.T) {
	for _, name := range []string{"optional_function_memoize_a.a", "optional_function_memoize_b.a"} {
		t.Run(name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", name))
			if err != nil {
				t.Fatal(err)
			}
			observed := onNode(t, path)
			if observed.exitCode != 0 || !bytes.Equal(observed.stdout, []byte("42 42 1\n")) || len(observed.stderr) != 0 {
				t.Fatalf("Node: exit %d, stdout %q, stderr %q", observed.exitCode, observed.stdout, observed.stderr)
			}
			_, err = lowered(t, path)
			var refused *lower.Refused
			if !errors.As(err, &refused) || !strings.Contains(err.Error(), "adamic/cycle-capable") {
				t.Fatalf("got %v, want the separate generic capture cycle refusal", err)
			}
		})
	}
}
