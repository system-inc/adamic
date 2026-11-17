package lower

import (
	"testing"
)

// A probe, never for landing (#x2cjvy9): does lowering ever compile a write of a field through any?
// tsc's clone loop writes (clone as any)[key] = (node as any)[key] (nodeFactory.ts), and the closed-kind
// rule's build check counts such a write as reaching every supertype. If no write through any lowers,
// in .ts or .a, the rule may leave them out, since a program holding one never compiles.
func TestProbeWriteThroughAny(t *testing.T) {
	t.Parallel()
	prelude := "enum K { A = 1, B = 2 }\ninterface Ex { readonly kind: K; readonly e: number; readonly _b: undefined }\nconst x: Ex = { kind: K.A, e: 1, _b: undefined };\n"
	for _, probe := range []struct{ name, source string }{
		{"dot", "(x as any).kind = 2;\nconsole.log(String(x.kind));\n"},
		{"key", "const key: string = 'kind';\n(x as any)[key] = 2;\nconsole.log(String(x.kind));\n"},
		{"clone", "function clone(n: Ex): Ex {\n  const c = { kind: K.A, e: 0, _b: undefined } as Ex;\n  for (const key in n) { (c as any)[key] = (n as any)[key]; }\n  return c;\n}\nconsole.log(String(clone(x).e));\n"},
		{"any binding", "const y: any = x;\ny.kind = 2;\nconsole.log(String(x.kind));\n"},
		{"any parameter", "function set(o: any): void { o.kind = 2; }\nset(x);\nconsole.log(String(x.kind));\n"},
	} {
		for _, extension := range []string{".ts", ".a"} {
			_, err := lowerSourceAs(t, prelude+probe.source, extension)
			t.Logf("PROBE %s%s lower=%v", probe.name, extension, err)
		}
	}
}
