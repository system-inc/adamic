package lower

import (
	"strings"
	"testing"
)

// The enum-tag walk's nullish-only steps (Adamic's ruling on #s8gy89r, Nov 22, family 1's U8): a
// step that only removes undefined and null, or leaves exactly them, may sit outside an access
// reference's local tests and passes there counting no proof. tsc writes it as
// `if (node.name && nodeIsMissing(node.name)) return;` before `node.name && isX(node.name)`
// (checker.ts:43181) and as the ternary chain at transformers/esDecorators.ts:550.

const nullishStepHolder = `function isB(n: N): n is B { return n.kind === K.B; }
interface O { n: N | undefined }
function clear(o: O): void { o.n = undefined; }
function missing(n: N): boolean { return isB(n) && n.b === ''; }
interface Q { n: N | null | undefined }
`

// Each lowers and agrees with Node, sanitized natively and on the JavaScript backend. Mutants, in
// enum_tag_nullish_step_mutants_test.go: (Local) the true side outside the local tests asks proves
// again, as before U8, and the early-return rows go back to the open-enum refusal; (FalseSide) the
// false-side arm gone, and every row does, since each continues on a falsy test's false side.
func TestEnumTagNullishStepLowers(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, stdout string }{
		{"early return, then the read re-tested", "function show(o: O): string {\n    if (o.n && missing(o.n)) return 'missing';\n    return o.n && isA(o.n) ? `a ${o.n.a}` : 'other';\n}\nconsole.log(show({ n: a }));\nconsole.log(show({ n: b }));\nconsole.log(show({ n: undefined }));\n", "a 1\nother\nother\n"},
		{"ternary chain", "function show(o: O): string {\n    return o.n && isB(o.n) ? `b ${o.n.b}` : o.n && isA(o.n) ? `a ${o.n.a}` : 'member';\n}\nconsole.log(show({ n: a }));\nconsole.log(show({ n: b }));\nconsole.log(show({ n: undefined }));\n", "a 1\nb x\nmember\n"},
		{"a resetting call before the re-test", "function show(o: O): string {\n    if (o.n && missing(o.n)) return 'missing';\n    clear(o);\n    return o.n && isA(o.n) ? `a ${o.n.a}` : 'other';\n}\nconsole.log(show({ n: a }));\n", "other\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			source := kindPredicateTypes + accessHolder + nullishStepHolder + probe.source
			for _, extension := range []string{".ts", ".a"} {
				if _, err := lowerSourceAs(t, source, extension); err != nil {
					t.Fatalf("%s: want it lowered, got %v", extension, err)
				}
			}
			program := lowersAndAgreesWithNodeAs(t, source, ".ts")
			if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != probe.stdout {
				t.Fatalf("native: exit %d, stdout %q, stderr %q; want 0 and %q", code, stdout, stderr, probe.stdout)
			}
		})
	}
}

// What stays refused, each with the open-enum refusal at the read. The premise rows: a non-null
// step outside the read's local tests proves nothing, so a read that relies on it after a call that
// writes undefined is refused (Node throws a TypeError), and so is the same with no call between, an
// over-refusal Adamic noted and doesn't owe. A step that removes null and keeps undefined is neither
// shape. Mutants: (Proves) that non-null step counting a proof lowers the row with nothing between;
// the rows with a call are held by codeBetween as well, so it leaves them refused. (Exact) a false
// side passing without being exactly the nullish part lowers the null-removed row.
func TestEnumTagNullishStepRefusals(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"premise, call then predicate", "function show(o: O): string {\n    if (!o.n) return 'none';\n    clear(o);\n    return isA(o.n) ? `a ${o.n.a}` : 'other';\n}\nconsole.log(show({ n: a }));\n"},
		{"premise, call then if", "function show(o: O): string {\n    if (!o.n) return 'none';\n    clear(o);\n    if (isA(o.n)) return `a ${o.n.a}`;\n    return 'other';\n}\nconsole.log(show({ n: a }));\n"},
		{"premise, nothing between", "function show(o: O): string {\n    if (!o.n) return 'none';\n    return isA(o.n) ? `a ${o.n.a}` : 'other';\n}\nconsole.log(show({ n: a }));\n"},
		{"null removed, undefined kept", "function show(q: Q): string {\n    return q.n !== null && q.n !== undefined && isA(q.n) ? `a ${q.n.a}` : 'other';\n}\nconsole.log(show({ n: a }));\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, kindPredicateTypes+accessHolder+nullishStepHolder+probe.source, ".ts")
			if err == nil || !strings.Contains(err.Error(), kindPredicateRefusal) {
				t.Fatalf("want %q, got %v", kindPredicateRefusal, err)
			}
		})
	}
}

// Outside the walk, the same hazard lowers and stops loudly: the inserted check names the call that
// put undefined back, where Node throws a TypeError.
func TestEnumTagNullishStepHazardStopsAtRunTime(t *testing.T) {
	t.Parallel()
	const source = `enum Kind { Identifier = 80, Body = 90 }
interface Name { readonly kind: Kind; readonly text: string }
interface Body extends Name { readonly kind: Kind.Body; readonly body: string }
function isBody(name: Name): name is Body { return name.kind === Kind.Body; }
interface Holder { name: Name | undefined }
function clear(node: Holder): void { node.name = undefined; }
function show(node: Holder): string {
    if (!node.name) return 'none';
    clear(node);
    return isBody(node.name) ? node.name.body : 'other';
}
const b: Body = { kind: Kind.Body, text: 'b', body: 'hello' };
console.log(show({ name: b }));
`
	program, err := lowerSourceAs(t, source, ".ts")
	if err != nil {
		t.Fatalf("want it lowered, got %v", err)
	}
	const message = "undefined where the checker narrowed it away: a call since the narrowing put it back"
	if code, stdout, stderr := brandedNative(t, program); code != 70 || stdout != "" || !strings.Contains(stderr, message) {
		t.Fatalf("native: exit %d, stdout %q, stderr %q; want 70, nothing printed and %q", code, stdout, stderr, message)
	}
}
