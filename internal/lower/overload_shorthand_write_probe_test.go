package lower

import "testing"

// A throwaway probe (#0kkpw2e set 3's red row): censusReturnProof finds a write to a parameter through
// GetSymbolAtLocation, which on a shorthand name answers the property, so ({ x } = other) may go unseen and an
// overload's narrow result may be taken as proven. Each row logs its stop or, where it lowers, Node's output
// beside native's.
func TestOverloadShorthandWriteProbe(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, extension, source string }{
		{"literal", ".ts", "const other: { x: number } = { x: 7 };\nfunction pick(x: 1): 1;\nfunction pick(x: number): number {\n    ({ x } = other);\n    return x;\n}\nconst one: 1 = pick(1);\nconsole.log(`${one === 1 ? \"one\" : \"other\"} ${one + 1}`);\n"},
		{"literal_a", ".a", "const other: { x: number } = { x: 7 };\nfunction pick(x: 1): 1;\nfunction pick(x: number): number {\n    ({ x } = other);\n    return x;\n}\nconst one: 1 = pick(1);\nconsole.log(`${one === 1 ? \"one\" : \"other\"} ${one + 1}`);\n"},
		{"object", ".ts", "interface Wide { readonly kind: string; }\ninterface Narrow extends Wide { readonly size: number; }\nconst wide: Wide = { kind: \"wide\" };\nconst other: { x: Wide } = { x: wide };\nfunction pick(x: Narrow): Narrow;\nfunction pick(x: Wide): Wide {\n    ({ x } = other);\n    return x;\n}\nconst got = pick({ kind: \"narrow\", size: 3 });\nconsole.log(`${got.kind} ${got.size}`);\n"},
		{"control_plain_write", ".ts", "const other: { x: number } = { x: 7 };\nfunction pick(x: 1): 1;\nfunction pick(x: number): number {\n    x = other.x;\n    return x;\n}\nconst one: 1 = pick(1);\nconsole.log(`${one === 1 ? \"one\" : \"other\"} ${one + 1}`);\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, err := lowerSourceAs(t, probe.source, probe.extension)
			if err != nil {
				t.Logf("ROW %s stops: %v", probe.name, err)
				return
			}
			node := runAgreementNode(t, writeAgreementSource(t, "main.ts", probe.source))
			code, stdout, stderr := brandedNative(t, program)
			t.Logf("ROW %s LOWERS: node exit %d %q; native exit %d %q stderr %q", probe.name, node.code, node.stdout, code, stdout, stderr)
		})
	}
}
