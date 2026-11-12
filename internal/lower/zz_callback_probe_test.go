package lower

import (
	"os"
	"strings"
	"testing"
)

// A probe, never for landing: which call shapes put the overload census of assertNode under a concrete T,
// as tsc's destructuring.ts call does (#68x1t2y). The cap is 32 on this branch, so a shape whose walk reads
// the deep N refuses.
func TestZZCallbackProbe(t *testing.T) {
	os.Setenv("ADAMIC_PROBE_CALLBACK", "1")
	assertDecl := `function assertNode<T extends N, U extends T>(node: T | undefined, test: (node: T) => node is U, message?: string): asserts node is U;
function assertNode(node: N | undefined, test: ((node: N) => boolean) | undefined, message?: string): void;
function assertNode(node: N | undefined, test: ((node: N) => boolean) | undefined, message?: string): void {
    check(node !== undefined && (test === undefined || test(node)), message || "Unexpected node.");
}
`
	nsDecl := "namespace Debug {\n" + strings.ReplaceAll(assertDecl, "function assertNode", "export function assertNode") + "}\n"
	body := "const id: Id = { kind: 1, text: \"a\" };\n"
	shapes := []struct{ name, source string }{
		{"A module function, module call", assertDecl + body + "const s: Sub = id;\nassertNode(s, isId);\nconsole.log(s.text);\n"},
		{"B namespace function, module call", nsDecl + body + "const s: Sub = id;\nDebug.assertNode(s, isId);\nconsole.log(s.text);\n"},
		{"C namespace function, call in a function", nsDecl + body + "function use(s: Sub): void { Debug.assertNode(s, isId); console.log(s.text); }\nuse(id);\n"},
		{"D module function, call in a function", assertDecl + body + "function use(s: Sub): void { assertNode(s, isId); console.log(s.text); }\nuse(id);\n"},
		{"E module function, call in a function declared first", "function use(s: Sub): void { assertNode(s, isId); console.log(s.text); }\n" + assertDecl + body + "use(id);\n"},
	}
	for _, shape := range shapes {
		source := overloadCallbackDeepNodes(overloadCallbackNodes) + shape.source
		os.Stderr.WriteString("PROBE shape " + shape.name + "\n")
		_, err := lowerSourceAs(t, source, ".ts")
		t.Logf("SHAPE %s: %v", shape.name, err)
	}
}
