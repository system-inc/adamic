package lower

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

// Probe only, never lands: which viewed arrays an object pattern destructures with no element check.
func TestZZProbeForOfObjectPatternView(t *testing.T) {
	arms := []struct{ name, decl, arm, pattern string }{
		{"map-union", "", "Node | Map<string, number>", "size"},
		{"map-alone", "", "Map<string, number>", "size"},
		{"set-union", "", "Node | Set<number>", "size"},
		{"set-alone", "", "Set<number>", "size"},
		{"date-union", "", "Node | Date", "size"},
		{"regexp-union", "", "Node | RegExp", "size"},
		{"regexp-alone", "", "RegExp", "source"},
		{"error-union", "", "Node | Error", "size"},
		{"error-alone", "", "Error", "message"},
		{"class-union", "class Counted { readonly size: number = 1; readonly kind: number = 2; }\n", "Node | Counted", "size"},
		{"class-alone", "class Counted { readonly size: number = 1; readonly kind: number = 2; }\n", "Counted", "size"},
		{"function-union", "", "Node | (() => number)", "size"},
		{"bytes-alone", "", "Uint8Array", "length"},
		{"array-alone", "", "readonly number[]", "length"},
		{"record-alone", "", "Record<string, number>", "size"},
		{"optional-union", "", "Node | undefined", "size"},
		{"null-union", "", "Node | null", "size"},
		{"partial", "", "Partial<Node>", "size"},
		{"node", "", "Node", "size"},
		{"two-nodes", "interface Leaf { readonly kind: 3; readonly size: number; readonly length: number; }\n", "Node | Leaf", "size"},
		{"string-union", "", "Node | string", "length"},
		{"number-union", "", "Node | number", "size"},
	}
	for _, arm := range arms {
		for _, shape := range []string{"field", "widened"} {
			if shape == "widened" && !strings.HasPrefix(arm.arm, "Node | ") {
				continue
			}
			field := arm.arm
			values := "const values = view._children;\n"
			if shape == "widened" {
				field = "Node"
				values = "const values: readonly (" + arm.arm + ")[] = view._children;\n"
			}
			prelude := `interface Node { readonly kind: number; readonly size: number; readonly length: number; }
interface Base { readonly tag: 1; }
` + arm.decl + `interface SyntaxList extends Base { readonly _children: readonly (` + field + `)[]; }
const raw = { tag: 1 as const, _children: [7] };
const base: Base = raw;
const view = base as SyntaxList;
` + values
			for _, loop := range []struct{ name, text string }{
				{"identifier", "for (const item of values) { console.log(String(item !== undefined)); }\n"},
				{"pattern", fmt.Sprintf("for (const { %s } of values) { console.log(String(%s)); }\n", arm.pattern, arm.pattern)},
			} {
				program, err := load.Load([]string{writeAgreementSource(t, "main.ts", prelude+loop.text)})
				if err != nil {
					fmt.Printf("PROBE %-15s %-8s %-10s load: %s\n", arm.name, shape, loop.name, strings.ReplaceAll(err.Error(), "\n", " | "))
					continue
				}
				_, err = Lower(context.Background(), program)
				fmt.Printf("PROBE %-15s %-8s %-10s %v\n", arm.name, shape, loop.name, err)
			}
		}
	}
}
