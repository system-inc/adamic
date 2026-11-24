//go:build optionalmapprobe

package lower

import "testing"

// Lower's probe on #gtc2ms2: one any-typed construction anywhere stops every optional Map get through a view.
func TestOptionalMapProbe(t *testing.T) {
	get := "function lookup(m: ReadonlyMap<string, number> | undefined): number | undefined { return m?.get(\"a\"); }\n" +
		"console.log(String(lookup(new Map([[\"a\", 1]]))));\n"
	for _, row := range []struct{ name, source string }{
		{"any-typed function constructor", "function Links(this: any) { this.flags = 0; }\nconst links = new (Links as any)();\nconsole.log(String(links.flags));\n" + get},
		{"class instead", "class Links { flags = 0; }\nconst links = new Links();\nconsole.log(String(links.flags));\n" + get},
		{"structural map literal", "const fake: ReadonlyMap<string, number> = { get: (k: string) => 7, has: (k: string) => true, size: 1, forEach: () => {}, entries: () => new Map<string, number>().entries(), keys: () => new Map<string, number>().keys(), values: () => new Map<string, number>().values(), [Symbol.iterator]: () => new Map<string, number>().entries() };\nconsole.log(String(fake.size));\n" + get},
		{"no construction", get},
	} {
		t.Run(row.name, func(t *testing.T) {
			program, err := lowerSourceAs(t, row.source, ".ts")
			if err != nil {
				t.Logf("ROW %s: STOPS %v", row.name, err)
				return
			}
			code, stdout, stderr := brandedNative(t, program)
			t.Logf("ROW %s: LOWERS native exit %d stdout %q stderr %.200q", row.name, code, stdout, stderr)
		})
	}
}
