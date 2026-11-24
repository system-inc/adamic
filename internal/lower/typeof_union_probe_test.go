//go:build typeofunionprobe

package lower

import "testing"

// Lower's probe on #5np1c51 -12: getter_plain_union_field.ts segfaults natively, and so does a direct call
// of its method. These rows hold no Reflect.get anywhere, to see whether main does it too.
func TestTypeofUnionProbe(t *testing.T) {
	class := "interface Held { readonly kind: 'Holder'; readonly amount: number | string; }\n" +
		"class Wrapper {\n" +
		"	readonly held: Held;\n" +
		"	constructor(base: Held) { this.held = base; }\n" +
		"	get isOne(): boolean { return this.check(); }\n" +
		"	check(): boolean { return typeof this.held.amount === \"number\"; }\n" +
		"}\n"
	for _, row := range []struct{ name, source string }{
		{"method, unannotated literal", class + "const raw = { kind: 'Holder' as const, amount: 1 };\nconsole.log(String(new Wrapper(raw).check()));\n"},
		{"getter, unannotated literal", class + "const raw = { kind: 'Holder' as const, amount: 1 };\nconsole.log(String(new Wrapper(raw).isOne));\n"},
		{"method, literal typed Held", class + "const raw: Held = { kind: 'Holder', amount: 1 };\nconsole.log(String(new Wrapper(raw).check()));\n"},
		{"no class, unannotated literal", "interface Held { readonly kind: 'Holder'; readonly amount: number | string; }\nfunction check(held: Held): boolean { return typeof held.amount === \"number\"; }\nconst raw = { kind: 'Holder' as const, amount: 1 };\nconsole.log(String(check(raw)));\n"},
		{"no class, value read", "interface Held { readonly kind: 'Holder'; readonly amount: number | string; }\nfunction value(held: Held): number | string { return held.amount; }\nconst raw = { kind: 'Holder' as const, amount: 1 };\nconsole.log(String(value(raw)));\n"},
	} {
		t.Run(row.name, func(t *testing.T) {
			program, err := lowerSourceAs(t, row.source, ".ts")
			if err != nil {
				t.Logf("ROW %s: STOPS %v", row.name, err)
				return
			}
			code, stdout, stderr := brandedNative(t, program)
			t.Logf("ROW %s: native exit %d stdout %q stderr %.300q", row.name, code, stdout, stderr)
		})
	}
}
