//go:build typeofunionprobe

package lower

import "testing"

// Lower's probe on #5np1c51 -12: getter_plain_union_field.ts prints nothing natively where Node prints true.
func TestTypeofUnionProbe(t *testing.T) {
	head := "interface Held { readonly kind: 'Holder'; readonly amount: number | string; }\n" +
		"class Wrapper {\n" +
		"	readonly held: Held;\n" +
		"	constructor(base: Held) { this.held = base; }\n" +
		"	get isOne(): boolean { return this.check(); }\n" +
		"	check(): boolean { return typeof this.held.amount === \"number\"; }\n" +
		"}\n" +
		"function read(value: object, key: string): unknown {\n" +
		"	return Reflect.get(Object(value), key, value);\n" +
		"}\n" +
		"const raw = { kind: 'Holder' as const, amount: 1 };\n"
	for _, row := range []struct{ name, source string }{
		{"reflect", head + "console.log(String(read(new Wrapper(raw), 'isOne')));\n"},
		{"direct getter", head + "console.log(String(new Wrapper(raw).isOne));\n"},
		{"direct method", head + "console.log(String(new Wrapper(raw).check()));\n"},
		{"no class", "interface Held { readonly kind: 'Holder'; readonly amount: number | string; }\nconst raw: Held = { kind: 'Holder', amount: 1 };\nconsole.log(String(typeof raw.amount === \"number\"));\n"},
		{"reflect equality", head + "console.log(String(read(new Wrapper(raw), 'isOne') === true));\n"},
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
