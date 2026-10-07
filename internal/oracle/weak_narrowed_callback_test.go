package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// Register the sound neighbors without changing the shared fixture list's source file.
func init() {
	for _, name := range []string{"map", "each", "reduce", "find", "some", "sort"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{
			"internal/oracle/testdata/9984394_narrowed_" + name + "_explicit.a", true, false,
		})
	}
}

// Node runs these while both targets have strong owners. Stage 0 must not hand their
// handles to callback parameters that read objects. The annotation is the sound neighbor.
func TestWeakNarrowedCallbacksStayNotYet(t *testing.T) {
	t.Parallel()
	for name, output := range map[string]string{
		"map": "a1,b1\n", "each": "a1\nb1\n", "reduce": "4\n",
		"find": "b1\nb1\nb1\n", "some": "true\n", "sort": "true\n",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/weak_callback_refused/9984394_narrowed_"+name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			if got := onNode(t, path); disagreement(run{stdout: []byte(output)}, got) != "" {
				t.Fatalf("Node: exit %d, stdout %q, stderr %q", got.exitCode, got.stdout, got.stderr)
			}
			program, err := lowered(t, path)
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) {
				if err == nil {
					got, _ := natively(t, program)
					t.Logf("accepted native run: exit %d, stdout %q, stderr %q", got.exitCode, got.stdout, got.stderr)
				}
				t.Fatalf("want NotYet for a Weak handle callback, got %v", err)
			}
			if !strings.Contains(notYet.Where, filepath.Base(path)+":") || !strings.Contains(notYet.What, "callback parameter") || !strings.Contains(notYet.What, "Weak<") {
				t.Fatalf("want path, reason and annotation fix, got %v", err)
			}
			t.Log(err)
		})
	}
}
