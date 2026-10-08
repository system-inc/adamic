package oracle

import (
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// Register this unit's fixtures without changing the shared oracle driver.
func init() {
	for _, path := range []string{
		"internal/oracle/testdata/never_string.a",
		"internal/oracle/testdata/never_values.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}

// Compile every lowering oracle's emitted C, including stack_overflow, which
// deliberately has no counts row. This audit runs no unrelated oracle programs.
func TestOracleCWarningFree(t *testing.T) {
	t.Parallel()
	for _, fixture := range fixtures {
		if !fixture.lowers {
			continue
		}
		t.Run(fixture.path, func(t *testing.T) {
			t.Parallel()
			program, err := lowered(t, filepath.Join(repository, fixture.path))
			if err != nil {
				t.Fatal(err)
			}
			if err := native.Build(native.C(program), filepath.Join(t.TempDir(), "program"), native.Options{}); err != nil {
				t.Fatal(err)
			}
		})
	}
}
