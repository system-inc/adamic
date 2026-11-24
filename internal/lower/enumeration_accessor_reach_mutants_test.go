package lower

import (
	"errors"
	"os/exec"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
)

type enumerationReachMutation struct {
	name, before, after, pin string
	rows                     []string
}

var enumerationReachMutations = []enumerationReachMutation{
	{"WholeGraph", "if reach.unknown {", "if true || reach.unknown {", "TestEnumerationReachUnrelatedGetterLowers", []string{"unrelated was refused"}},
	{"UnionMembers", "case flags&checker.TypeFlagsUnion != 0:", "case false && flags&checker.TypeFlagsUnion != 0:", "TestEnumerationReachUnionLowers", []string{"union was refused"}},
	{"DeepReach", "if !deep {", "if true {", "TestEnumerationReachDeepGetterRefused/(field|element)", []string{"field was admitted", "element was admitted"}},
	{"DiscriminatedUnion", "types = append(types, next.proven) // the union itself", "// the union itself is left unread", "TestEnumerationReachReachingGetterRefused/discriminated", []string{"discriminated was admitted"}},
	{"GenericLiteral", "	if holder == nil {\n		return nil, false\n	}\n	if holder.Kind == ast.KindObjectLiteralExpression {", "	if holder == nil || enumerationGenericContext(holder) {\n		return nil, false\n	}\n	if holder.Kind == ast.KindObjectLiteralExpression {", "TestEnumerationReachGenericLiteralLowers", []string{"generic literal was refused"}},
	{"LiteralTypeParameter", "		if l.eachReached(owner, 0,", "		if false && l.eachReached(owner, 0,", "TestEnumerationReachReachingGetterRefused/held", []string{"held was admitted"}},
	{"DeclaredOwner", "		if declared, named := l.enumerationLiteralDeclaredOwner(holder); named {", "		if declared, named := l.enumerationLiteralDeclaredOwner(holder); false && named {", "TestEnumerationReachDeclaredLiteralLowers", []string{"declared literal was refused"}},
	{"GenericClass", "		if enumerationGenericContext(class.Parent) {", "		if enumerationGenericContext(class) {", "TestEnumerationReachGenericClassLowers", []string{"generic class was refused"}},
	{"AnyInstance", "		owners = append(owners, l.enumerationAnyInstance(side, static))", "		owners = append(owners, side)", "TestEnumerationReachReachingGetterRefused/generic", []string{"generic was admitted"}},
	{"Contextual", "		if contextual := l.enumerationLiteralContextualOwners(holder); contextual != nil {", "		if contextual := l.enumerationLiteralContextualOwners(holder); false && contextual != nil {", "TestEnumerationReachContextualLiteralLowers", []string{"contextual literal was refused"}},
	{"ContextualScope", "	if len(enclosing) != 0 && l.eachReached(contextual,", "	if false && len(enclosing) != 0 && l.eachReached(contextual,", "TestEnumerationReachReachingGetterRefused/held", []string{"held was admitted"}},
	{"OneWay", "l.checker.IsTypeAssignableTo(owner, read) || l.checker.IsTypeAssignableTo(read, owner)", "l.checker.IsTypeAssignableTo(owner, read)", "TestEnumerationReachDowncastRefused", []string{"downcast was admitted"}},
	{"ViewPass", "func (l *lowering) checkEnumerationReachViews() error {", "func (l *lowering) checkEnumerationReachViews() error {\n\treturn nil // mutant: no admitted enumeration is held against the checked views", "TestEnumerationReachViewRefused/sibling", []string{"sibling was admitted"}},
}

func enumerationReachMutantBinary(t *testing.T, m enumerationReachMutation) string {
	t.Helper()
	return buildcache.GoTestMutated(t, "", "enumeration-reach-"+m.name+".test", "./internal/lower", nil, []buildcache.Mutation{{File: "internal/lower/enumeration_accessor_reach.go", Before: m.before, After: m.after}})
}
func TestProduct_EnumerationReachWholeGraphMutant(t *testing.T) {
	t.Parallel()
	enumerationReachMutantBinary(t, enumerationReachMutations[0])
}
func TestProduct_EnumerationReachUnionMembersMutant(t *testing.T) {
	t.Parallel()
	enumerationReachMutantBinary(t, enumerationReachMutations[1])
}
func TestProduct_EnumerationReachDeepReachMutant(t *testing.T) {
	t.Parallel()
	enumerationReachMutantBinary(t, enumerationReachMutations[2])
}

func TestProduct_EnumerationReachDiscriminatedUnionMutant(t *testing.T) {
	t.Parallel()
	enumerationReachMutantBinary(t, enumerationReachMutations[3])
}

func TestProduct_EnumerationReachGenericLiteralMutant(t *testing.T) {
	t.Parallel()
	enumerationReachMutantBinary(t, enumerationReachMutations[4])
}

func TestProduct_EnumerationReachLiteralTypeParameterMutant(t *testing.T) {
	t.Parallel()
	enumerationReachMutantBinary(t, enumerationReachMutations[5])
}

func TestProduct_EnumerationReachDeclaredOwnerMutant(t *testing.T) {
	t.Parallel()
	enumerationReachMutantBinary(t, enumerationReachMutations[6])
}

func TestProduct_EnumerationReachGenericClassMutant(t *testing.T) {
	t.Parallel()
	enumerationReachMutantBinary(t, enumerationReachMutations[7])
}

func TestProduct_EnumerationReachAnyInstanceMutant(t *testing.T) {
	t.Parallel()
	enumerationReachMutantBinary(t, enumerationReachMutations[8])
}

func TestProduct_EnumerationReachContextualMutant(t *testing.T) {
	t.Parallel()
	enumerationReachMutantBinary(t, enumerationReachMutations[9])
}

func TestProduct_EnumerationReachContextualScopeMutant(t *testing.T) {
	t.Parallel()
	enumerationReachMutantBinary(t, enumerationReachMutations[10])
}

func TestProduct_EnumerationReachOneWayMutant(t *testing.T) {
	t.Parallel()
	enumerationReachMutantBinary(t, enumerationReachMutations[11])
}

func TestProduct_EnumerationReachViewPassMutant(t *testing.T) {
	t.Parallel()
	enumerationReachMutantBinary(t, enumerationReachMutations[12])
}

// The mutants run one at a time, so at most one mutant binary competes with this package's own tests on a
// small box. Each pin fails fast under its mutant, so the walk stays far inside the test budget.
func TestEnumerationReachMutantsAreCaught(t *testing.T) {
	t.Parallel()
	for _, m := range enumerationReachMutations {
		t.Run(m.name, func(t *testing.T) {
			binary := enumerationReachMutantBinary(t, m)
			command := exec.CommandContext(t.Context(), binary, "-test.run", "^"+strings.ReplaceAll(m.pin, "/", "$/^")+"$", "-test.count=1", "-test.v")
			stdout, stderr, err := childguard.Capture(t.Context(), command)
			var exit *exec.ExitError
			if !errors.As(err, &exit) || exit.ExitCode() != 1 {
				t.Fatalf("%s enumeration reach mutant missed %s: %v\n%s\n%s", m.name, m.pin, err, stdout, stderr)
			}
			for _, row := range m.rows {
				if !strings.Contains(string(stdout), "enumeration reach row "+row) {
					t.Fatalf("%s enumeration reach mutant missed %s:\n%s", m.name, row, stdout)
				}
			}
			t.Logf("%s enumeration reach mutant caught by %s:\n%s", m.name, m.pin, stdout)
		})
	}
}
