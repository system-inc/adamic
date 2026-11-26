package lower

import (
	"strings"
	"testing"
)

// A property narrowed to an object arm and kept by the checker across a call that may rewrite it stops at
// compile time where a stale value's field could be read at the wrong representation (#8nj5pn4, Compiler's -5
// ruling): if (m.n.kind === 'a') { reset(m); return String(m.n.v + 1); } with v a number in A and a string in B
// reads the string's pointer as a number natively. A field only one arm declares, Optimizer's m.n.a, is read
// by name or by a slot every layout holding it shares, and a value lacking it stops loudly there (Emit's guard),
// so it isn't this stop's. Rows: string tags and an open enum's, read directly and handed on, and the honest
// twin whose call writes an a, which the rule, keyed by the name, stops too; a getter named n stops even with
// nothing between, since it answers each read itself. The same read where both arms hold v as a string, rows
// with nothing between the check and the read, a held binding the call can't rewrite, and an unrelated getter
// lower and agree with Node. Language's r08 is the same hole through an element, arr[0] after a call that writes
// x[0], pinned with its honest twin and its nothing-between control. Mutants (narrowed_arm_property.mutants.md): the property dispatch dropped, the
// write test dropped, and the silent test dropped (the same-representation row stops).
func TestNarrowedArmPropertyRead(t *testing.T) {
	t.Parallel()
	const enumTypes = "enum K { A = 1, B = 2 }\ninterface A { readonly kind: K.A; readonly v: number }\ninterface B { readonly kind: K.B; readonly v: string }\ntype N = A | B;\ninterface M { n: N }\nconst a: A = { kind: K.A, v: 1 };\nconst b: B = { kind: K.B, v: \"x\" };\nfunction reset(m: M): void { m.n = b; }\nfunction keep(m: M): void { m.n = a; }\nfunction use(x: A): string { return String(x.v + 1); }\n"
	const stringTypes = "type A = { readonly kind: 'a'; readonly v: number };\ntype B = { readonly kind: 'b'; readonly v: string };\ntype N = A | B;\ninterface M { n: N }\nconst a: A = { kind: 'a', v: 1 };\nconst b: B = { kind: 'b', v: \"x\" };\nfunction reset(m: M): void { m.n = b; }\nfunction keep(m: M): void { m.n = a; }\nfunction use(x: A): string { return String(x.v + 1); }\n"
	const sameTypes = "type A = { readonly kind: 'a'; readonly v: string };\ntype B = { readonly kind: 'b'; readonly v: string };\ntype N = A | B;\ninterface M { n: N }\nconst a: A = { kind: 'a', v: \"y\" };\nconst b: B = { kind: 'b', v: \"x\" };\nfunction reset(m: M): void { m.n = b; }\n"
	const stop = "narrowed to A where a call may have rewritten it"
	const elements = "function resetFirst(x: N[]): void { x[0] = b; }\nfunction keepFirst(x: N[]): void { x[0] = a; }\n"
	for _, row := range []struct{ name, source, stdout string }{
		{"string direct lie", stringTypes + "function show(m: M): string { if (m.n.kind === 'a') { reset(m); return String(m.n.v + 1); } return 'b'; }\nconsole.log(show({ n: a }));\n", ""},
		{"string handed on lie", stringTypes + "function show(m: M): string { if (m.n.kind === 'a') { reset(m); return use(m.n); } return 'b'; }\nconsole.log(show({ n: a }));\n", ""},
		{"enum direct lie", enumTypes + "function show(m: M): string { if (m.n.kind === K.A) { reset(m); return String(m.n.v + 1); } return 'b'; }\nconsole.log(show({ n: a }));\n", ""},
		{"enum handed on lie", enumTypes + "function show(m: M): string { if (m.n.kind === K.A) { reset(m); return use(m.n); } return 'b'; }\nconsole.log(show({ n: a }));\n", ""},
		{"string honest twin", stringTypes + "function show(m: M): string { if (m.n.kind === 'a') { keep(m); return String(m.n.v + 1); } return 'b'; }\nconsole.log(show({ n: a }));\n", ""},
		{"same representation", sameTypes + "function show(m: M): string { if (m.n.kind === 'a') { reset(m); return m.n.v; } return 'b'; }\nconsole.log(show({ n: a }));\n", "x\n"},
		{"string nothing between", stringTypes + "function show(m: M): string { if (m.n.kind === 'a') { return String(m.n.v + 1); } return 'b'; }\nconsole.log(show({ n: a }));\nconsole.log(show({ n: b }));\n", "2\nb\n"},
		{"enum nothing between", enumTypes + "function show(m: M): string { if (m.n.kind === K.A) { return String(m.n.v + 1); } return 'b'; }\nconsole.log(show({ n: a }));\n", "2\n"},
		{"getter named n", stringTypes + "const g = { get n(): N { return b; } };\nconsole.log(g.n.kind);\nfunction show(m: M): string { if (m.n.kind === 'a') { return String(m.n.v + 1); } return 'b'; }\nconsole.log(show({ n: a }));\n", ""},
		{"unrelated getter", stringTypes + "const g = { get other(): number { return 1; } };\nfunction show(m: M): string { if (m.n.kind === 'a') { return String(m.n.v + 1); } return 'b'; }\nconsole.log(show({ n: a }));\nconsole.log(String(g.other));\n", "2\n1\n"},
		{"element lie", stringTypes + elements + "function show(arr: N[]): string { if (arr[0] !== undefined && arr[0].kind === 'a') { resetFirst(arr); return String(arr[0].v + 1); } return 'b'; }\nconsole.log(show([a]));\n", ""},
		{"element honest twin", stringTypes + elements + "function show(arr: N[]): string { if (arr[0] !== undefined && arr[0].kind === 'a') { keepFirst(arr); return String(arr[0].v + 1); } return 'b'; }\nconsole.log(show([a]));\n", ""},
		{"element nothing between", stringTypes + elements + "function show(arr: N[]): string { if (arr[0] !== undefined && arr[0].kind === 'a') { return String(arr[0].v + 1); } return 'b'; }\nconsole.log(show([a]));\nconsole.log(show([b]));\n", "2\nb\n"},
		{"string held binding", stringTypes + "function show(m: M): string { const n = m.n; if (n.kind === 'a') { reset(m); return String(n.v + 1); } return 'b'; }\nconsole.log(show({ n: a }));\n", "2\n"},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			if row.stdout == "" {
				_, err := lowerSourceAs(t, row.source, ".ts")
				if err == nil || !strings.Contains(err.Error(), stop) {
					t.Fatalf("want a not yet with %q, got %v", stop, err)
				}
				return
			}
			program := lowersAndAgreesWithNodeAs(t, row.source, ".ts")
			if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != row.stdout {
				t.Fatalf("native: exit %d, stdout %q, stderr %q; want 0 and %q", code, stdout, stderr, row.stdout)
			}
		})
	}
}
