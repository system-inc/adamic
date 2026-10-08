package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func init() {
	for _, name := range []string{"method", "method_parameter"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{
			"internal/oracle/testdata/9984394_view_" + name + "_matching.a", true, false,
		})
	}
}

// Method thunks cannot reinterpret a strong result or argument as a Weak handle.
// Node establishes the source behavior; stage 0 must diagnose the incompatible view.
func TestWeakMethodViewsStayNotYet(t *testing.T) {
	t.Parallel()
	for name, output := range map[string]string{"method": "true b1\n", "method_parameter": "t1 b1\n"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/weak_view_refused/9984394_view_"+name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			if got := onNode(t, path); disagreement(run{stdout: []byte(output)}, got) != "" {
				t.Fatalf("Node: exit %d, stdout %q, stderr %q", got.exitCode, got.stdout, got.stderr)
			}
			_, err = lowered(t, path)
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) {
				t.Fatalf("want NotYet for incompatible method keeping, got %v", err)
			}
			if !strings.Contains(notYet.Where, filepath.Base(path)+":") || !strings.Contains(notYet.What, "weakly") || !strings.Contains(notYet.What, "keep the Weak annotations") {
				t.Fatalf("want path, keeping reason and fix, got %v", err)
			}
			t.Log(err)
		})
	}
}
