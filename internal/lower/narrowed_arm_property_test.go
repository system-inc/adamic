package lower

import (
	"strings"
	"testing"
)

// A property narrowed to an object arm and kept by the checker across a call that may rewrite it stops at
// compile time (#8nj5pn4): if (m.n.kind === 'a') { reset(m); return String(m.n.a + 1); } read b's storage
// through a's layout natively and printed 1, where Node prints NaN. Optimizer's rows: string tags and an
// open enum's, read directly and handed on, and the honest twin whose call writes an a, which the rule, keyed
// by the name, stops too. Rows with nothing between the check and the read, and a held binding the call
// can't rewrite, lower and agree with Node. A getter named n anywhere stops the read even with nothing between,
// since it answers each read itself, and an unrelated getter elsewhere doesn't. Mutant (narrowed_arm_property.mutants.md): narrowedArmReadGap
// without its property dispatch, and the lie rows lower again.
func TestNarrowedArmPropertyRead(t *testing.T) {
	t.Parallel()
	const enumTypes = "enum K { A = 1, B = 2 }\ninterface A { readonly kind: K.A; readonly a: number }\ninterface B { readonly kind: K.B; readonly b: string }\ntype N = A | B;\ninterface M { n: N }\nconst a: A = { kind: K.A, a: 1 };\nconst b: B = { kind: K.B, b: \"x\" };\nfunction reset(m: M): void { m.n = b; }\nfunction keep(m: M): void { m.n = a; }\nfunction use(x: A): string { return String(x.a + 1); }\n"
	const stringTypes = "type A = { readonly kind: 'a'; readonly a: number };\ntype B = { readonly kind: 'b'; readonly b: string };\ntype N = A | B;\ninterface M { n: N }\nconst a: A = { kind: 'a', a: 1 };\nconst b: B = { kind: 'b', b: \"x\" };\nfunction reset(m: M): void { m.n = b; }\nfunction keep(m: M): void { m.n = a; }\nfunction use(x: A): string { return String(x.a + 1); }\n"
	const stop = "a read of property n narrowed to A where a call may have rewritten it"
	for _, row := range []struct{ name, source, stdout string }{
		{"string direct lie", stringTypes + "function show(m: M): string { if (m.n.kind === 'a') { reset(m); return String(m.n.a + 1); } return 'b'; }\nconsole.log(show({ n: a }));\n", ""},
		{"string handed on lie", stringTypes + "function show(m: M): string { if (m.n.kind === 'a') { reset(m); return use(m.n); } return 'b'; }\nconsole.log(show({ n: a }));\n", ""},
		{"enum direct lie", enumTypes + "function show(m: M): string { if (m.n.kind === K.A) { reset(m); return String(m.n.a + 1); } return 'b'; }\nconsole.log(show({ n: a }));\n", ""},
		{"enum handed on lie", enumTypes + "function show(m: M): string { if (m.n.kind === K.A) { reset(m); return use(m.n); } return 'b'; }\nconsole.log(show({ n: a }));\n", ""},
		{"string honest twin", stringTypes + "function show(m: M): string { if (m.n.kind === 'a') { keep(m); return String(m.n.a + 1); } return 'b'; }\nconsole.log(show({ n: a }));\n", ""},
		{"string nothing between", stringTypes + "function show(m: M): string { if (m.n.kind === 'a') { return String(m.n.a + 1); } return 'b'; }\nconsole.log(show({ n: a }));\nconsole.log(show({ n: b }));\n", "2\nb\n"},
		{"enum nothing between", enumTypes + "function show(m: M): string { if (m.n.kind === K.A) { return String(m.n.a + 1); } return 'b'; }\nconsole.log(show({ n: a }));\n", "2\n"},
		{"getter named n", stringTypes + "const g = { get n(): N { return b; } };\nconsole.log(g.n.kind);\nfunction show(m: M): string { if (m.n.kind === 'a') { return String(m.n.a + 1); } return 'b'; }\nconsole.log(show({ n: a }));\n", ""},
		{"unrelated getter", stringTypes + "const g = { get other(): number { return 1; } };\nfunction show(m: M): string { if (m.n.kind === 'a') { return String(m.n.a + 1); } return 'b'; }\nconsole.log(show({ n: a }));\nconsole.log(String(g.other));\n", "2\n1\n"},
		{"string held binding", stringTypes + "function show(m: M): string { const n = m.n; if (n.kind === 'a') { reset(m); return String(n.a + 1); } return 'b'; }\nconsole.log(show({ n: a }));\n", "2\n"},
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
