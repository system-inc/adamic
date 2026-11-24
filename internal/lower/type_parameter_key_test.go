package lower

import "testing"

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
// (a string and a number). The instance's result is Pkg[K], which holds both, so the finite read reads each
// field at its own storage and fits it to that union, as main does for a literal-union key: it lowers, and
// both picks agree with Node natively. The read's own stop ("a computed field with differently represented
// alternatives") fires only where a field can't be fitted to the result, which this road can't reach,
// since the result is always the indexed access of the names K binds.
func TestTypeParameterKeyOfDifferingFieldsAgrees(t *testing.T) {
	t.Parallel()
	source := typeParameterKeyPrelude + "function pick(n: number): \"name\" | \"count\" { return n > 0 ? \"name\" : \"count\"; }\nconst picked = [field(p, pick(1)), field(p, pick(0))];\nconsole.log(typeof picked[0]);\nconsole.log(typeof picked[1]);\nconsole.log(String(picked[0]));\nconsole.log(String(picked[1]));\n"
	program := lowersAndAgreesWithNodeAs(t, source, ".ts")
	if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != "string\nnumber\na\n3\n" {
		t.Fatalf("native: exit %d, stdout %q, stderr %q; want what Node prints", code, stdout, stderr)
	}
}
