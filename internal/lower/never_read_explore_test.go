package lower

import (
	"fmt"
	"testing"
)

func TestNeverReadExplore(t *testing.T) {
	for _, probe := range []struct{ name, source string }{
		{"lie_cast", neverReadShapes + "console.log(String(corners(\"circle\")));\nconsole.log(String(corners(\"triangle\" as unknown as Shape)));\n"},
		{"unknown_param", "type Shape = \"circle\" | \"square\";\nfunction fail(shape: unknown): never { throw new Error(\"unexpected shape \" + String(shape)); }\nfunction corners(shape: Shape): number {\n\tswitch (shape) {\n\t\tcase \"circle\": return 0;\n\t\tcase \"square\": return 4;\n\t}\n\treturn fail(shape);\n}\nconsole.log(String(corners(\"circle\")));\n"},
		{"while", "type Shape = \"circle\" | \"square\";\nfunction fail(shape: never): never { throw new Error(\"x\"); }\nfunction total(shape: Shape, times: number): number {\n\tlet sum = 0;\n\tfor (let i = 0; i < times; i++) {\n\t\tswitch (shape) { case \"circle\": sum += 1; break; case \"square\": sum += 2; break; default: fail(shape); }\n\t}\n\treturn sum;\n}\nconsole.log(String(total(\"circle\", 2)));\n"},
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
