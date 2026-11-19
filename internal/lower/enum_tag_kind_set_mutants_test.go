package lower

import (
	"errors"
	"os/exec"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
)

type kindSetMutation struct{ name, before, after, pin, witness string }

var kindSetMutations = []kindSetMutation{
	{"Disjunction", "return r.comparisons(b.Left) && r.comparisons(b.Right)", "return false", "TestKindSetBodyGrammar/disjunction", "kind-set grammar disjunction"},
	{"Switch", "return r.switchBody(statements[0], len(statements) == 2)", "return false", "TestKindSetBodyGrammar/switch", "kind-set grammar switch"},
	{"SwitchTail", "return hasDefault || tailFalse", "return hasDefault", "TestKindSetBodyGrammar/switch_tail", "kind-set grammar switch_tail"},
	{"Alias", "r.alias = r.l.checker.GetSymbolAtLocation(d.Name())", "r.alias = nil", "TestKindSetBodyGrammar/alias", "kind-set grammar alias"},
	{"Guard", "return r.guard(b.Left) && r.comparisons(b.Right)", "return false", "TestKindSetBodyGrammar/guard", "kind-set grammar guard"},
	{"DoubleGuard", "if n.Kind == ast.KindPrefixUnaryExpression && n.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken {", "if false && n.Kind == ast.KindPrefixUnaryExpression && n.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken {", "TestKindSetBodyGrammar/double_guard", "kind-set grammar double_guard"},
	{"IfGuard", "return test.ElseStatement == nil && test.ThenStatement.Kind == ast.KindBlock && r.guard(test.Expression) && r.body(test.ThenStatement)", "_ = test; return false", "TestKindSetBodyGrammar/if_guard_alias", "kind-set grammar if_guard_alias"},
	{"FalseSide", "// Preserve the singleton rule exactly, including never (an empty K(S)).", "return l.kindSetMeets(tag, set) // mutant: drop any tag that meets the set", "TestKindSetFalseSideKeepsPartialTag", "kind-set partial tag was dropped from the false side"},
}

func kindSetMutantBinary(t *testing.T, m kindSetMutation) string {
	t.Helper()
	return buildcache.GoTestMutated(t, "", "kind-set-"+m.name+".test", "./internal/lower", nil, []buildcache.Mutation{{File: "internal/lower/enum_tag_kind_set.go", Before: m.before, After: m.after}})
}
func TestProduct_KindSetDisjunctionMutant(t *testing.T) {
	t.Parallel()
	kindSetMutantBinary(t, kindSetMutations[0])
}
func TestProduct_KindSetSwitchMutant(t *testing.T) {
	t.Parallel()
	kindSetMutantBinary(t, kindSetMutations[1])
}
func TestProduct_KindSetSwitchTailMutant(t *testing.T) {
	t.Parallel()
	kindSetMutantBinary(t, kindSetMutations[2])
}
func TestProduct_KindSetAliasMutant(t *testing.T) {
	t.Parallel()
	kindSetMutantBinary(t, kindSetMutations[3])
}
func TestProduct_KindSetGuardMutant(t *testing.T) {
	t.Parallel()
	kindSetMutantBinary(t, kindSetMutations[4])
}
func TestProduct_KindSetDoubleGuardMutant(t *testing.T) {
	t.Parallel()
	kindSetMutantBinary(t, kindSetMutations[5])
}
func TestProduct_KindSetIfGuardMutant(t *testing.T) {
	t.Parallel()
	kindSetMutantBinary(t, kindSetMutations[6])
}
func TestProduct_KindSetFalseSideMutant(t *testing.T) {
	t.Parallel()
	kindSetMutantBinary(t, kindSetMutations[7])
}
func TestKindSetBodyMutantsAreCaught(t *testing.T) {
	t.Parallel()
	for _, m := range kindSetMutations {
		t.Run(m.name, func(t *testing.T) {
			t.Parallel()
			binary := kindSetMutantBinary(t, m)
			command := exec.CommandContext(t.Context(), binary, "-test.run", "^"+strings.ReplaceAll(m.pin, "/", "$/^")+"$", "-test.count=1", "-test.v")
			stdout, stderr, err := childguard.Capture(t.Context(), command)
			var exit *exec.ExitError
			if !errors.As(err, &exit) || exit.ExitCode() != 1 || !strings.Contains(string(stdout), m.witness) {
				t.Fatalf("%s mutant missed %s: %v\n%s\n%s", m.name, m.pin, err, stdout, stderr)
			}
			t.Logf("%s mutant caught by %s:\n%s", m.name, m.pin, stdout)
		})
	}
}
