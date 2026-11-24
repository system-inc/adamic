package lower

import (
	"errors"
	"strings"
	"testing"
)

// typeParameterKeyPrelude is readPackageJsonField's shape (moduleNameResolver.ts:368, #e24qsqx): a generic
// accessor whose key is K extends keyof a fixed interface, read as the instance's one field name.
const typeParameterKeyPrelude = `interface Pkg { readonly name: string; readonly version: string; readonly count: number }
function field<K extends keyof Pkg>(p: Pkg, key: K): Pkg[K] {
    const value = p[key];
    return value;
}
const p: Pkg = { name: "a", version: "1", count: 3 };
`

// Each call's K is one name, so each instance's read is the finite read of that field, and it agrees with
// Node natively. Mutant: finiteKeyFields reads the key's type as written (no concrete), and this stops at
// "an ElementAccessExpression" again.
func TestTypeParameterKeyReadLowers(t *testing.T) {
	t.Parallel()
	source := typeParameterKeyPrelude + "console.log(field(p, \"name\"));\nconsole.log(field(p, \"version\"));\nconsole.log(String(field(p, \"count\")));\n"
	program := lowersAndAgreesWithNodeAs(t, source, ".ts")
	if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != "a\n1\n3\n" {
		t.Fatalf("native: exit %d, stdout %q, stderr %q; want what Node prints", code, stdout, stderr)
	}
}

// Adamic's condition on #e24qsqx: a K bound to two names whose fields are held in different representations
// stops at the finite read's own words, so the road never reads past one field's representation.
func TestTypeParameterKeyOfDifferingFieldsStops(t *testing.T) {
	t.Parallel()
	source := typeParameterKeyPrelude + "function pick(n: number): \"name\" | \"count\" { return n > 0 ? \"name\" : \"count\"; }\nconsole.log(String(field(p, pick(1))));\n"
	_, err := lowerSourceAs(t, source, ".ts")
	var notYet *NotYet
	if !errors.As(err, &notYet) || !strings.Contains(notYet.What, "a computed field with differently represented alternatives") {
		t.Fatalf("want NotYet at the differently represented alternatives, got %v", err)
	}
}
