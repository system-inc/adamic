package lower

import (
	"strings"
	"testing"
)

// Throwaway measurement for #5qx6qv4, never committed: each source's lowering outcome, printed.
func TestSourceSlotPendingProbe(t *testing.T) {
	const tail = "function isIdent(n: Node): n is Ident { return n.kind === 80; }\nfunction setValueDeclaration(symbol: Sym, node: Decl): void {\n\tsymbol.valueDeclaration = node;\n}\nconst raw = { kind: 80, text: \"x\" };\nconst n: Node = raw;\nif (isIdent(n)) console.log(n.text);\n"
	const head = "interface Sym { flags: number; valueDeclaration?: Decl | undefined }\n"
	const rest = "interface Decl extends Node { _declarationBrand: undefined }\ninterface Ident extends Node { readonly kind: 80; text: string }\n"
	prelude := "interface FlowNode { readonly id: number }\ninterface Node { readonly kind: \"a\" | \"b\"; flowNode?: FlowNode | undefined }\ninterface NodeA extends Node { readonly kind: \"a\" }\nconst node: Node = { kind: \"a\", flowNode: { id: 1 } };\nconst view = node as NodeA;\nexport function setIt(t: NodeA, v: FlowNode): void {\n\tt.flowNode = v;\n}\n"
	for _, row := range []struct{ name, source string }{
		{"exported uncalled", prelude + "console.log(String(view.kind));\n"},
		{"163 map", head + "interface Node { readonly kind: number; symbol?: Sym | undefined; locals?: Map<string, Sym> | undefined }\n" + rest + tail},
		{"163 tuple", head + "interface Node { readonly kind: number; symbol?: Sym | undefined; parts?: [number, ...string[]] | undefined }\n" + rest + tail},
		{"163 plain", head + "interface Node { readonly kind: number; symbol?: Sym | undefined }\n" + rest + tail},
		{"163 map called", head + "interface Node { readonly kind: number; symbol?: Sym | undefined; locals?: Map<string, Sym> | undefined }\n" + rest + strings.Replace(tail, "if (isIdent(n))", "const sym: Sym = { flags: 0 };\nsetValueDeclaration(sym, { kind: 1, _declarationBrand: undefined });\nif (isIdent(n))", 1)},
	} {
		_, err := lowerSourceAs(t, row.source, ".ts")
		outcome := "lowers"
		if err != nil {
			outcome = err.Error()
			if at := strings.Index(outcome, "main.ts:"); at >= 0 {
				outcome = outcome[at:]
			}
		}
		t.Logf("PROBE %s: %s", row.name, outcome)
	}
}
