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
	{"Guard", "return r.guard(b.Left) && r.comparisons(b.Right)", "_ = b; return false", "TestKindSetBodyGrammar/guard", "kind-set grammar guard"},
	{"DoubleGuard", "if n.Kind == ast.KindPrefixUnaryExpression && n.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken {", "if false && n.Kind == ast.KindPrefixUnaryExpression && n.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken {", "TestKindSetBodyGrammar/double_guard", "kind-set grammar double_guard"},
	{"IfGuard", "return test.ElseStatement == nil && test.ThenStatement.Kind == ast.KindBlock && r.guard(test.Expression) && r.body(test.ThenStatement)", "_ = test; return false", "TestKindSetBodyGrammar/if_guard_alias", "kind-set grammar if_guard_alias"},
	{"FalseSide", "// Preserve the singleton rule exactly, including never (an empty K(S)).", "return l.kindSetMeets(tag, set) // mutant: drop any tag that meets the set", "TestKindSetFalseSideKeepsPartialTag", "kind-set partial tag was dropped from the false side"},
	{"FirstKindFit", "for _, kind := range call.members {", "for _, kind := range call.members[:1] {", "TestKindSetClosedSourcesFitEveryKind/one_does_not_fit", "kind-set all-kinds fit"},
	{"NotEqual", "if b.OperatorToken.Kind != ast.KindEqualsEqualsEqualsToken {", "if b.OperatorToken.Kind != ast.KindEqualsEqualsEqualsToken && b.OperatorToken.Kind != ast.KindExclamationEqualsEqualsToken {", "TestKindSetPredicateFallbacksKeepFamilyOne/not_equal", "kind-set body fallback not_equal"},
	{"EnumIdentity", "if identity == nil || r.enum != nil && r.enum != identity {", "if identity == nil {", "TestKindSetPredicateFallbacksKeepFamilyOne/two_enums", "kind-set body fallback two_enums"},
	{"PrimitiveGuard", "if part.Flags()&checker.TypeFlagsPrimitive != 0 && part.Flags()&(checker.TypeFlagsNull|checker.TypeFlagsUndefined) == 0 {", "if false && part.Flags()&checker.TypeFlagsPrimitive != 0 && part.Flags()&(checker.TypeFlagsNull|checker.TypeFlagsUndefined) == 0 {", "TestKindSetPredicateFallbacksKeepFamilyOne/primitive_guard", "kind-set body fallback primitive_guard"},
	{"TrueDefault", "if falls || c.Kind == ast.KindDefaultClause && outcome {", "if falls {", "TestKindSetPredicateFallbacksKeepFamilyOne/true_default", "kind-set body fallback true_default"},
	{"DuplicateCase", "if checker.Checker_isTypeIdenticalTo(r.l.checker, earlier, value) {", "if false && checker.Checker_isTypeIdenticalTo(r.l.checker, earlier, value) {", "TestKindSetPredicateFallbacksKeepFamilyOne/(duplicate_case|k04)", "kind-set body fallback duplicate_case"},
	{"UnionClaimFit", "if !w.l.closedKindSourcesFit(member, call.field, kind, claimed) {", "if claimed.Flags()&checker.TypeFlagsUnion == 0 && !w.l.closedKindSourcesFit(member, call.field, kind, claimed) {", "TestKindSetUnionClaimFitsEverySource/odd_.*", "kind-set union claim"},
}

func kindSetMutantBinary(t *testing.T, m kindSetMutation) string {
	t.Helper()
	file := "internal/lower/enum_tag_kind_set.go"
	if m.name == "FirstKindFit" || m.name == "UnionClaimFit" {
		file = "internal/lower/enum_tag_flow.go"
	}
	return buildcache.GoTestMutated(t, "", "kind-set-"+m.name+".test", "./internal/lower", nil, []buildcache.Mutation{{File: file, Before: m.before, After: m.after}})
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
func TestProduct_KindSetFirstKindFitMutant(t *testing.T) {
	t.Parallel()
	kindSetMutantBinary(t, kindSetMutations[8])
}
func TestProduct_KindSetNotEqualMutant(t *testing.T) {
	t.Parallel()
	kindSetMutantBinary(t, kindSetMutations[9])
}
func TestProduct_KindSetEnumIdentityMutant(t *testing.T) {
	t.Parallel()
	kindSetMutantBinary(t, kindSetMutations[10])
}
func TestProduct_KindSetPrimitiveGuardMutant(t *testing.T) {
	t.Parallel()
	kindSetMutantBinary(t, kindSetMutations[11])
}
func TestProduct_KindSetTrueDefaultMutant(t *testing.T) {
	t.Parallel()
	kindSetMutantBinary(t, kindSetMutations[12])
}
func TestProduct_KindSetDuplicateCaseMutant(t *testing.T) {
	t.Parallel()
	kindSetMutantBinary(t, kindSetMutations[13])
}
func TestProduct_KindSetUnionClaimFitMutant(t *testing.T) {
	t.Parallel()
	kindSetMutantBinary(t, kindSetMutations[14])
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
			if m.name == "DuplicateCase" && !strings.Contains(string(stdout), "kind-set body fallback k04") {
				t.Fatalf("duplicate-case mutant missed the equal-value k04 row:\n%s", stdout)
			}
			if m.name == "UnionClaimFit" {
				for _, row := range []string{"odd_declared", "odd_built", "odd_anonymous"} {
					if !strings.Contains(string(stdout), "kind-set union claim "+row+":") {
						t.Fatalf("union-claim mutant missed %s:\n%s", row, stdout)
					}
				}
			}
			t.Logf("%s mutant caught by %s:\n%s", m.name, m.pin, stdout)
		})
	}
}
