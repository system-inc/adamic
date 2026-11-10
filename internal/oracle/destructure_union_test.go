package oracle

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

var destructureUnionFixtures = []string{"semver", "null", "former_stops", "object"}

func init() {
	for _, name := range destructureUnionFixtures {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{
			"internal/oracle/testdata/destructure_union_" + name + ".a", true, false,
		})
	}
}

// A destructured Union reads the boxed field just as a property read does. The
// null fixture distinguishes a default from ??, and the object fixture exercises
// ownership of a boxed reference copied into a local.
func TestDestructureUnionAgrees(t *testing.T) {
	t.Parallel()
	for _, name := range destructureUnionFixtures {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			relative := "internal/oracle/testdata/destructure_union_" + name + ".a"
			for _, backend := range []string{"native", "sanitized", "javascript", "wasi"} {
				t.Run(backend, func(t *testing.T) {
					t.Parallel()
					if backend != "wasi" {
						loomViewsDAgreesBackend(t, relative, backend)
						return
					}
					if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
						t.Skip("set ADAMIC_ORACLE_WASI=1 to run the WASI oracle")
					}
					path, err := filepath.Abs(filepath.Join(repository, relative))
					if err != nil {
						t.Fatal(err)
					}
					want := onNode(t, path)
					program, err := lowered(t, path)
					if err != nil {
						t.Fatal(err)
					}
					got := onWASI(t, native.C(program))
					if diff := disagreement(want, got); diff != "" {
						t.Errorf("wasi: %s, stderr %q", diff, got.stderr)
					}
				})
			}
		})
	}
}
