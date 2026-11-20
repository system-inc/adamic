package lower

import (
	"errors"
	"os/exec"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
)

type kindAliasMutation struct {
	name, before, after, pin string
	rows                     []string
}

var kindAliasMutations = []kindAliasMutation{
	{"Initializer", "return w.fieldRead(variable.Initializer)", "return ast.SkipParentheses(variable.Initializer).Name().Text()", "TestKindAliasFallbacksKeepFamilyOne/(other_reference|getter)", []string{"other_reference", "getter"}},
	{"RunsFirst", "|| !kindAliasRunsFirst(list.Parent, statement) {", "|| false && !kindAliasRunsFirst(list.Parent, statement) {", "TestKindAliasFallbacksKeepFamilyOne/(after|other_function|loop)", []string{"after", "other_function", "loop"}},
	{"Rewritten", "if w.kindAliasRewritten(", "if false && w.kindAliasRewritten(", "TestKindAliasFallbacksKeepFamilyOne/(assigned|destructured|shorthand)", []string{"assigned", "destructured", "shorthand"}},
	{"Let", "list.Flags&ast.NodeFlagsBlockScoped != ast.NodeFlagsConst", "list.Flags&ast.NodeFlagsBlockScoped == 0", "TestKindAliasFallbacksKeepFamilyOne/let", []string{"let"}},
}

func kindAliasMutantBinary(t *testing.T, m kindAliasMutation) string {
	t.Helper()
	return buildcache.GoTestMutated(t, "", "kind-alias-"+m.name+".test", "./internal/lower", nil, []buildcache.Mutation{{File: "internal/lower/enum_tag_kind_alias.go", Before: m.before, After: m.after}})
}
func TestProduct_KindAliasInitializerMutant(t *testing.T) {
	t.Parallel()
	kindAliasMutantBinary(t, kindAliasMutations[0])
}
func TestProduct_KindAliasRunsFirstMutant(t *testing.T) {
	t.Parallel()
	kindAliasMutantBinary(t, kindAliasMutations[1])
}
func TestProduct_KindAliasRewrittenMutant(t *testing.T) {
	t.Parallel()
	kindAliasMutantBinary(t, kindAliasMutations[2])
}
func TestProduct_KindAliasLetMutant(t *testing.T) {
	t.Parallel()
	kindAliasMutantBinary(t, kindAliasMutations[3])
}

func TestKindAliasMutantsAreCaught(t *testing.T) {
	t.Parallel()
	for _, m := range kindAliasMutations {
		t.Run(m.name, func(t *testing.T) {
			t.Parallel()
			binary := kindAliasMutantBinary(t, m)
			command := exec.CommandContext(t.Context(), binary, "-test.run", "^"+strings.ReplaceAll(m.pin, "/", "$/^")+"$", "-test.count=1", "-test.v")
			stdout, stderr, err := childguard.Capture(t.Context(), command)
			var exit *exec.ExitError
			if !errors.As(err, &exit) || exit.ExitCode() != 1 {
				t.Fatalf("%s alias mutant missed %s: %v\n%s\n%s", m.name, m.pin, err, stdout, stderr)
			}
			for _, row := range m.rows {
				if !strings.Contains(string(stdout), "kind-alias fallback "+row+" was recognized") {
					t.Fatalf("%s alias mutant missed %s:\n%s", m.name, row, stdout)
				}
			}
			t.Logf("%s alias mutant caught by %s:\n%s", m.name, m.pin, stdout)
		})
	}
}
