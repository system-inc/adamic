package lower

import "testing"

// A container of void values stays legal wherever no position is seen wider (#p7t8tmh): its length,
// its identity, a loop that reads no element, and passing it where void[] is expected. The refused
// side is internal/oracle/testdata/void_widening_refused/contained_*.
func TestVoidContainersStayLegal(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"length", `console.log(String(all.length));`},
		{"identity", `const other = all; console.log(String(all === other));`},
		{"loop without a read", `let n = 0; for (const x of all) { n += 1; } console.log(String(n));`},
		{"passed as void[]", `function size(values: readonly void[]): number { return values.length; } console.log(String(size(all)));`},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			lowersAndAgreesWithNode(t, "const make: () => void = () => 42; const all: void[] = [1, 2].map(() => make());\n"+probe.source)
		})
	}
}
