package oracle

import (
	"errors"
	"github.com/system-inc/adamic/internal/lower"
	"path/filepath"
	"strings"
	"testing"
)

func TestHiddenTNodeValueExactNotYet(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/notyet/hidden_boundary_generic_tnode_value.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	var gap *lower.NotYet
	if program != nil || !errors.As(err, &gap) || gap.What != "a generic function as a value" || !strings.HasSuffix(gap.Where, "hidden_boundary_generic_tnode_value.a:3:36") {
		t.Fatalf("wanted exact generic-value stop at 3:36: %v", err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "node\n" || len(truth.stderr) != 0 {
		t.Fatalf("Node: %+v", truth)
	}
}

func init() {
	for _, path := range []string{"internal/oracle/testdata/hidden_boundary_generic_tnode.a", "internal/oracle/testdata/hidden_boundary_generic_tnode_constraints.a"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/notyet/hidden_boundary_generic_tnode_value.a", false, false})
}

func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/notyet/hidden_boundary_generic_tnode_mutation.a", false, false})
}
