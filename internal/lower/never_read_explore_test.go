package lower

import (
	"fmt"
	"testing"
)

const exploreKinds = "function stray(value: unknown): string { return \"stray \" + String(value); }\ninterface NodeA { kind: \"a\"; a: number }\ninterface NodeB { kind: \"b\"; b: number }\ntype N = NodeA | NodeB;\n"

func TestNeverReadExploreLanguage(t *testing.T) {
	describe := "function describe(node: N): string {\n\tif (node.kind === \"a\") return \"a\";\n\ttouch(node);\n\tif (node.kind === \"b\") return \"b\";\n\t{ const seen: unknown = node; return stray(seen); }\n}\nconsole.log(describe({ kind: \"b\", b: 2 }));\n"
	for _, probe := range []struct{ name, source string }{
		{"h1", exploreKinds + "function touch(node: N): void { node.kind = \"a\"; }\n" + describe},
		{"widen", exploreKinds + "function touch(node: { kind: string }): void { node.kind = \"a\"; }\n" + describe},
		{"getter", "interface Circle { readonly kind: \"circle\"; readonly radius: number }\ninterface Square { readonly kind: \"square\"; readonly side: number }\ntype Shape = Circle | Square;\nclass Ring implements Circle {\n\tget kind(): \"circle\" { return \"circle\"; }\n\treadonly radius: number = 2;\n}\nfunction corners(shape: Shape): string {\n\tif (shape.kind === \"circle\") return \"circle \" + String(shape.radius);\n\tif (shape.kind === \"square\") return \"square\";\n\t{ const seen: unknown = shape; return \"stray \" + typeof seen; }\n}\nconsole.log(corners(new Ring()) + \" \" + corners({ kind: \"square\", side: 1 }));\n"},
		{"enum_switch", "enum K { A, B }\nfunction stray(value: unknown): string { return \"stray \" + String(value); }\nfunction name(k: K): string {\n\tswitch (k) {\n\t\tcase K.A: return \"a\";\n\t\tcase K.B: return \"b\";\n\t\tdefault: return stray(k);\n\t}\n}\nconsole.log(name(K.A) + \" \" + name(K.B));\n"},
		{"enum_number", "enum K { A, B }\nfunction stray(value: unknown): string { return \"stray \" + String(value); }\nfunction name(k: K): string {\n\tswitch (k) {\n\t\tcase K.A: return \"a\";\n\t\tcase K.B: return \"b\";\n\t\tdefault: return stray(k);\n\t}\n}\nfunction other(): number { return 7; }\nconsole.log(name(other()));\n"},
	} {
		for _, extension := range []string{".ts", ".a"} {
			program, err := lowerSourceAs(t, probe.source, extension)
			traps := []string{}
			if err == nil {
				for _, function := range program.Functions {
					if function.Name == "never_read" || function.Name == "enum_never" {
						traps = append(traps, function.Name)
					}
				}
			}
			text := fmt.Sprint(err)
			if len(text) > 190 {
				text = text[:190]
			}
			t.Logf("EXPLORE %s%s: %v %s", probe.name, extension, traps, text)
		}
	}
}
