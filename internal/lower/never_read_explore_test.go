package lower

import (
	"fmt"
	"testing"
)

func TestNeverReadExploreRoads(t *testing.T) {
	enumLet := "enum E { A, B }\nfunction describe(start: E): string {\n\tlet e: E = start;\n\tconst flip = (): void => { e = E.A; };\n\tif (e === E.A) return 'a';\n\tflip();\n\tif (e === E.B) return 'b';\n\t{ const seen: unknown = e; return String(seen); }\n}\nconsole.log(describe(E.B));\n"
	for _, probe := range []struct{ name, source string }{
		{"typeof", neverReadLet + "\treturn typeof node;\n}\nconsole.log(describe({ kind: K.B, b: 2 }));\n"},
		{"template", neverReadLet + "\treturn `${node}`;\n}\nconsole.log(describe({ kind: K.B, b: 2 }));\n"},
		{"equals", neverReadLet + "\treturn String(node === undefined);\n}\nconsole.log(describe({ kind: K.B, b: 2 }));\n"},
		{"string", neverReadLet + "\treturn String(node);\n}\nconsole.log(describe({ kind: K.B, b: 2 }));\n"},
		{"shorthand", neverReadLet + "\t{ const wrapped: { node: unknown } = { node }; return stray(wrapped.node); }\n}\nconsole.log(describe({ kind: K.B, b: 2 }));\n"},
		{"enum_let", enumLet},
	} {
		_, err := lowerSourceAs(t, probe.source, ".ts")
		text := fmt.Sprint(err)
		if len(text) > 170 {
			text = text[:170]
		}
		t.Logf("EXPLORE road %s: %s", probe.name, text)
	}
}

func TestNeverReadExplore(t *testing.T) {
	for _, probe := range []struct{ name, source string }{
		{"lie_ignore", neverReadShapes + "// @ts-ignore\nconst lie: Shape = \"triangle\";\nconsole.log(String(corners(lie)));\n"},
		{"unknown_param", "type Shape = \"circle\" | \"square\";\nfunction fail(shape: unknown): never { throw new Error(\"unexpected shape \" + String(shape)); }\nfunction corners(shape: Shape): number {\n\tswitch (shape) {\n\t\tcase \"circle\": return 0;\n\t\tcase \"square\": return 4;\n\t}\n\treturn fail(shape);\n}\nconsole.log(String(corners(\"circle\")));\n"},
		{"while", "type Shape = \"circle\" | \"square\";\nfunction fail(shape: unknown): never { throw new Error(\"x\"); }\nfunction total(shape: Shape, times: number): number {\n\tlet sum = 0;\n\tfor (let i = 0; i < times; i++) {\n\t\tswitch (shape) { case \"circle\": sum += 1; break; case \"square\": sum += 2; break; default: fail(shape); }\n\t}\n\treturn sum;\n}\nconsole.log(String(total(\"circle\", 2)));\n"},
		{"nv1_main_shape_direct", neverReadHolder + "\tif (holder.node.kind === K.B) return 'b';\n\t{ const seen: unknown = holder.node; return stray(seen); }\n}\nconsole.log(describe({ node: { kind: K.B, b: 2 } }));\n"},
	} {
		for _, extension := range []string{".ts", ".a"} {
			_, err := lowerSourceAs(t, probe.source, extension)
			text := fmt.Sprint(err)
			if len(text) > 200 {
				text = text[:200]
			}
			t.Logf("EXPLORE %s%s: %s", probe.name, extension, text)
		}
	}
}
