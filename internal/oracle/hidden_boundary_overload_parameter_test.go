package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/hidden_boundary_overload_parameter.a", true, false})
}

// TypeScript accepts these signatures, but the implementation parameter does
// not serve every admitted value. A successful Node run is not a type proof.
func TestHiddenBoundaryOverloadParameterRefused(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ file, stdout string }{
		{"hidden_boundary_overload_parameter.a", "8\n"},
		{"hidden_boundary_overload_generic.a", "0\n"},
	} {
		t.Run(probe.file, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/refusals", probe.file))
			if err != nil {
				t.Fatal(err)
			}
			observed := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), path)
			if difference := disagreement(run{stdout: []byte(probe.stdout)}, observed); difference != "" {
				t.Fatalf("source on Node: %s", difference)
			}
			if probe.file == "hidden_boundary_overload_parameter.a" {
				_, err := load.Load([]string{path})
				if err == nil || !strings.Contains(err.Error(), "TS2394") {
					t.Fatalf("expected the incompatible overload diagnostic, got %v", err)
				}
				return
			}
			_, err = lowered(t, path)
			var refused *lower.Refused
			if !errors.As(err, &refused) || refused.What != "overload 1 of visitBindingElement parameter elem cannot be served by implementation parameter elem" || !strings.Contains(refused.Error(), path) {
				t.Fatalf("expected the parameter contract refusal, got %v", err)
			}
		})
	}
}
