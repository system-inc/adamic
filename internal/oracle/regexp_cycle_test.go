package oracle

import (
	"reflect"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// The native regex slice is on codex/stage1-css, not main. Register this fixture
// when that slice's program representation is present, including on the scratch
// merge. Main still runs the regex IR proof tests in internal/fresh. This avoids
// importing the regex emitter and runtime into the cycle-proof change.
func nativeRegexSlicePresent() bool {
	_, present := reflect.TypeFor[ir.Program]().FieldByName("Regexps")
	return present
}

func init() {
	if nativeRegexSlicePresent() {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/fresh/testdata/regexp_tree.ts", true, false})
	}
}

func TestRegexCycleFixtureHasItsNativeDependency(t *testing.T) {
	t.Parallel()
	if !nativeRegexSlicePresent() {
		t.Skip("native regex lowering/emission is on codex/stage1-css; run this fixture on the scratch merge")
	}
}
