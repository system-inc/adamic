package oracle

import (
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
	}{
		"internal/oracle/testdata/8e67677_structural_matching.a", true, false,
	})
}

// A library prototype member is not a field or an own method in the native shape.
// The check belongs at the view boundary, including optional and nested slots.
func TestStructuralLibraryViewsStayNotYet(t *testing.T) {
	t.Parallel()
	for name, output := range map[string]string{
		"method": "true\n", "optional": "has test\nno test\n",
		"assignment": "true\n", "return": "true\n", "shorthand": "true\n", "array": "1\n", "metadata": "a\n", "map": "0\n",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_view_refused/8e67677_structural_"+name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			if got := onNode(t, path); disagreement(run{stdout: []byte(output)}, got) != "" {
				t.Fatalf("Node: exit %d, stdout %q, stderr %q", got.exitCode, got.stdout, got.stderr)
			}
			_, err = lowered(t, path)
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) {
				t.Fatalf("want NotYet for a structural library prototype view, got %v", err)
			}
			member := "test"
			if name == "metadata" {
				member = "source"
			} else if name == "map" {
				member = "size"
			}
			if !strings.Contains(notYet.Where, filepath.Base(path)+":") || !strings.Contains(notYet.What, "prototype member "+member) || !strings.Contains(notYet.What, "wrapper object") || !strings.Contains(notYet.What, "arrow") {
				t.Fatalf("want path, member and wrapper or arrow fix, got %v", err)
			}
			t.Log(err)
		})
	}
}
