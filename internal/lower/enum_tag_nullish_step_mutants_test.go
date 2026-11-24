package lower

import (
	"errors"
	"os/exec"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
)

// Each mutant of enum_tag_nullish_step.go and the pin rows it must turn red.
type nullishStepMutation struct {
	name, before, after, pin string
	rows                     []string
}

var nullishStepMutations = []nullishStepMutation{
	// The true side outside the local tests asks proves again, as before U8.
	{"Local", "if w.tests == nil || w.inLocalTests(at) {", "if true {", "TestEnumTagNullishStepLowers", []string{"early_return,_then_the_read_re-tested", "ternary_chain"}},
	// The non-null step outside the local tests counts a proof.
	{"Proves", "\t\treturn true, true\n\t}\n\tif before.Flags()", "\t\tw.proved++\n\t\treturn true, true\n\t}\n\tif before.Flags()", "TestEnumTagNullishStepRefusals", []string{"premise,_call_then_predicate", "premise,_call_then_if", "premise,_nothing_between"}},
	// A false side passes without being exactly before's nullish part.
	{"Exact", "\tvar left []*checker.Type\n", "\tif others > 0 {\n\t\treturn true, true\n\t}\n\tvar left []*checker.Type\n", "TestEnumTagNullishStepRefusals", []string{"typeof_leaves_an_object"}},
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
func TestProduct_NullishStepExactMutant(t *testing.T) {
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
