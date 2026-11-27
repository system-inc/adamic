package lower

import "testing"

// A probe, never for main (#xfpsqtn): a rest tuple literal read past its built length, beside Node.
func TestZZRestTuplePastLengthProbe(t *testing.T) {
	t.Parallel()
	const head = "enum E { A = 1, B = 2 }\nconst endings: readonly [E, ...E[]] = [E.A, E.B];\n"
	for _, row := range []struct{ name, source string }{
		{"past", head + "const third = endings[2];\nconsole.log(String(third));\n"},
		{"length", head + "console.log(String(endings.length));\n"},
		{"loop", head + "let out = '';\nfor (const e of endings) { out += String(e); }\nconsole.log(out);\n"},
		{"param past", "enum E { A = 1, B = 2 }\nfunction third(endings: readonly [E, ...E[]]): string { return String(endings[2]); }\nconsole.log(third([E.A, E.B]));\n"},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			program, err := lowerSourceAs(t, row.source, ".ts")
			if err != nil {
				t.Logf("PROBE %s lowering: %v", row.name, err)
				return
			}
			node := runAgreementNode(t, writeAgreementSource(t, "main.ts", row.source))
			code, stdout, stderr := brandedNative(t, program)
			t.Logf("PROBE %s native exit %d stdout %q stderr %.300q | node exit %d stdout %q", row.name, code, stdout, stderr, node.code, node.stdout)
		})
	}
}
