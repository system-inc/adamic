package oracle

import (
	"errors"
	"github.com/system-inc/adamic/internal/lower"
	"path/filepath"
	"testing"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/overload_primitive_comparison.a", true, false})
}

func TestOverloadComparisonRefusedSourcesOnNode(t *testing.T) {
	for _, probe := range []struct{ name, output string }{{"mixed", "true\n"}, {"written", "false\n"}} {
		path, err := filepath.Abs(filepath.Join(repository, "internal/lower/testdata/overload_comparison_"+probe.name+".a"))
		if err != nil {
			t.Fatal(err)
		}
		if got := onNode(t, path); disagreement(run{stdout: []byte(probe.output)}, got) != "" {
			t.Fatalf("Node %s: %+v", probe.name, got)
		}
		_, err = lowered(t, path)
		var notYet *lower.NotYet
		if !errors.As(err, &notYet) || notYet.What != "a BinaryExpression with a union of differently held members and a union of differently held members" {
			t.Fatalf("%s want pinned unsupported comparison, got %v", probe.name, err)
		}
	}
}
