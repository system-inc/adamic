package lower

import (
	"fmt"
	"testing"
)

// Probe only, never lands: which viewed arrays an object pattern destructures with no element check.
func TestZZProbeForOfObjectPatternView(t *testing.T) {
	arms := []struct{ name, decl, arm, pattern string }{
		{"map-union", "", "Node | Map<string, number>", "size"},
		{"map-alone", "", "Map<string, number>", "size"},
		{"set-union", "", "Node | Set<number>", "size"},
		{"date-union", "", "Node | Date", "size"},
		{"regexp-union", "", "Node | RegExp", "size"},
		{"error-union", "", "Node | Error", "size"},
		{"class-union", "class Counted { readonly size = 1; readonly kind = 2; }\n", "Node | Counted", "size"},
		{"class-alone", "class Counted { readonly size = 1; readonly kind = 2; }\n", "Counted", "size"},
		{"function-union", "", "Node | (() => number)", "size"},
		{"object-union", "", "Node | object", "size"},
		{"bytes-union", "", "Node | Uint8Array", "length"},
		{"promise-union", "", "Node | Promise<number>", "size"},
		{"array-union", "", "Node | readonly number[]", "length"},
		{"record-union", "", "Node | Record<string, number>", "size"},
		{"optional-union", "", "Node | undefined", "size"},
		{"null-union", "", "Node | null", "size"},
		{"generic", "", "Partial<Node>", "size"},
	}
	for _, arm := range arms {
		prelude := `interface Node { readonly kind: number; readonly size: number; readonly length: number; }
interface Base { readonly tag: 1; }
interface SyntaxList extends Base { readonly _children: readonly Node[]; }
` + arm.decl + `const raw = { tag: 1 as const, _children: [7] };
const base: Base = raw;
const view = base as SyntaxList;
const values: readonly (` + arm.arm + `)[] = view._children;
`
		for _, loop := range []struct{ name, text string }{
			{"identifier", "for (const item of values) { console.log(String(item !== undefined)); }\n"},
			{"pattern", fmt.Sprintf("for (const { %s } of values) { console.log(String(%s)); }\n", arm.pattern, arm.pattern)},
		} {
			_, err := lowerSourceAs(t, prelude+loop.text, ".ts")
			fmt.Printf("PROBE %-15s %-10s %v\n", arm.name, loop.name, err)
		}
	}
}
