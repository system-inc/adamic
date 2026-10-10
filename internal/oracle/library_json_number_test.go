package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// A generic Union body can be shared by a scalar and a container instantiation. Neither may
// borrow the other's descriptor. The independent source on Node proves these are valid programs.
func TestLibraryJSONGenericInstantiations(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ file, output string }{
		{"library_json_generic.a", "\"xx\"\n[1,2]\n"},
		{"library_json_generic_shared.a", "3\n[1,2]\n"},
	} {
		t.Run(probe.file, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "docs/library-json-number/probes", probe.file))
			if err != nil {
				t.Fatal(err)
			}
			node := onNode(t, path)
			if node.exitCode != 0 || string(node.stdout) != probe.output {
				t.Fatalf("Node: exit %d stdout %q stderr %q", node.exitCode, node.stdout, node.stderr)
			}
			_, err = lowered(t, path)
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) || (!strings.Contains(err.Error(), "generic union without per-instantiation container metadata") && !strings.Contains(err.Error(), "adamic/json-union-array")) {
				t.Fatalf("want compile-time generic metadata refusal, got %v", err)
			}
			t.Logf("Node prints %q; compiler refuses: %v", node.stdout, err)
		})
	}
}

// Register slice fixtures here: oracle_test.go belongs to integration. The common oracle and
// counts gate then test these programs exactly like its existing fixtures.
func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/library_number_immediate.a", true, false})
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/library_json_gap_nul.a", true, false})
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/library_json_shapes.a", true, false})
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/library_number_math_edges.a", true, false})
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/library_number_own_names.a", true, false})
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/library_number_convert_sources.a", true, false})
}
