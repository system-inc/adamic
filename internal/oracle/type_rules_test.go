package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

var typeRuleProbes = []string{"conditional_copy", "parameter_min", "parameter_array", "parameter_map", "parameter_field", "weak_min", "weak_field", "weak_method", "weak_callback", "union_min", "union_field", "union_method", "union_callback"}

// Register this unit's programs with the normal differential and counts gates.
func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/type_rules_controls.a", true, false})
	if os.Getenv("ADAMIC_TYPE_RULE_PROBES") != "1" {
		return
	}
	for _, name := range typeRuleProbes {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/type_rules_" + name + ".a", true, false})
	}
}

// The probes are executable oracle programs when a guard is removed. With the
// final compiler, reject them before emitting C and independently record Node.
func TestTypeRuleProbesStayRefused(t *testing.T) {
	if os.Getenv("ADAMIC_TYPE_RULE_PROBES") == "1" {
		t.Skip("unguarded oracle experiment")
	}
	for _, name := range typeRuleProbes {
		t.Run(name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/type_rules_"+name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			node := onNode(t, path)
			if node.exitCode != 0 {
				t.Fatalf("Node: exit %d stderr %q", node.exitCode, node.stderr)
			}
			_, err = lowered(t, path)
			var refused *lower.Refused
			id := "adamic/native-nominal-view"
			if name == "conditional_copy" {
				id = "adamic/invariant-mutable"
			}
			if strings.HasPrefix(name, "parameter_") {
				id = "adamic/invariant-constraint-view"
			}
			if !errors.As(err, &refused) || !strings.Contains(err.Error(), id) {
				t.Fatalf("want %s, got %v", id, err)
			}
			t.Logf("Node stdout %q; refused: %v", node.stdout, err)
		})
	}
}
