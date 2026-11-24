package lower

import (
	"errors"
	"os/exec"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
)

// Each mutant of enum_tag_nullish_step.go and the pin rows it must turn red. The false side's
// exactness (after is before's nullish part and nothing else) has no mutant: under the walk it can't
// be the only guard. A proving step before a false side has already removed undefined and null, so
// before holds none there, and with no proving step before it, the first read after that false side
// has no proof of its own and is refused (enumTagRefinementProven's proved == 0). It stays as the
// ruling's own words, so the rule holds if the walk's proof count ever changes.
type nullishStepMutation struct {
	name, before, after, pin string
	rows                     []string
}

var nullishStepMutations = []nullishStepMutation{
	// The true side outside the local tests asks proves again, as before U8.
	{"Local", "if w.tests == nil || w.inLocalTests(at) {", "if true {", "TestEnumTagNullishStepLowers", []string{"early_return,_then_the_read_re-tested", "a_resetting_call_before_the_re-test"}},
	// The non-null step outside the local tests counts a proof, as proves would count it: firstProof
	// moves back to the early return, and codeBetween then sees the call in its test.
	{"Proves", "\t\treturn true, true\n\t}\n\tif before.Flags()", "\t\tw.proved++\n\t\tif at != nil && (w.firstProof < 0 || at.Pos() < w.firstProof) {\n\t\t\tw.firstProof = at.Pos()\n\t\t}\n\t\treturn true, true\n\t}\n\tif before.Flags()", "TestEnumTagNullishStepLowers", []string{"early_return,_then_the_read_re-tested", "a_resetting_call_before_the_re-test"}},
	// The false-side arm gone: a falsy test's false side refuses, as before U8.
	{"FalseSide", "\tif before.Flags()&checker.TypeFlagsUnion == 0 {\n", "\tif true {\n", "TestEnumTagNullishStepLowers", []string{"early_return,_then_the_read_re-tested", "ternary_chain", "a_resetting_call_before_the_re-test"}},
}

func nullishStepMutantBinary(t *testing.T, m nullishStepMutation) string {
	t.Helper()
	return buildcache.GoTestMutated(t, "", "nullish-step-"+m.name+".test", "./internal/lower", nil, []buildcache.Mutation{{File: "internal/lower/enum_tag_nullish_step.go", Before: m.before, After: m.after}})
}
func TestProduct_NullishStepLocalMutant(t *testing.T) {
	t.Parallel()
	nullishStepMutantBinary(t, nullishStepMutations[0])
}
func TestProduct_NullishStepProvesMutant(t *testing.T) {
	t.Parallel()
	nullishStepMutantBinary(t, nullishStepMutations[1])
}
func TestProduct_NullishStepFalseSideMutant(t *testing.T) {
	t.Parallel()
	nullishStepMutantBinary(t, nullishStepMutations[2])
}

func TestNullishStepMutantsAreCaught(t *testing.T) {
	t.Parallel()
	for _, m := range nullishStepMutations {
		t.Run(m.name, func(t *testing.T) {
			t.Parallel()
			binary := nullishStepMutantBinary(t, m)
			command := exec.CommandContext(t.Context(), binary, "-test.run", "^"+m.pin+"$", "-test.count=1", "-test.v")
			stdout, stderr, err := childguard.Capture(t.Context(), command)
			var exit *exec.ExitError
			if !errors.As(err, &exit) || exit.ExitCode() != 1 {
				t.Fatalf("%s mutant missed %s: %v\n%s\n%s", m.name, m.pin, err, stdout, stderr)
			}
			for _, row := range m.rows {
				if !strings.Contains(string(stdout), "--- FAIL: "+m.pin+"/"+row+" ") {
					t.Fatalf("%s mutant left %s/%s green:\n%s", m.name, m.pin, row, stdout)
				}
			}
			t.Logf("%s mutant caught by %s", m.name, m.pin)
		})
	}
}
