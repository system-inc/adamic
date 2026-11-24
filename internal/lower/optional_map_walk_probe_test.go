//go:build optionalmapprobe

package lower

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// Lower's probe on #gtc2ms2: optionalMapStructuralReceiver's answer at an optional Map get, asked directly,
// for a program whose only other construction is new (Links as any)() and for the same with Links a class.
// The census lowers one unit, so the any cast elsewhere never stops it, but this walk is program-wide.
func TestOptionalMapWalkProbe(t *testing.T) {
	get := "function lookup(m: ReadonlyMap<string, number> | undefined): number | undefined { return m?.get(\"a\"); }\n"
	for _, row := range []struct{ name, source string }{
		{"any-typed function constructor", "function Links(this: any) { this.flags = 0; }\nfunction make(): unknown { return new (Links as any)(); }\n" + get},
		{"class instead", "class Links { flags = 0; }\nfunction make(): unknown { return new Links(); }\n" + get},
		{"no construction", get},
	} {
		t.Run(row.name, func(t *testing.T) {
			program, err := load.Load([]string{writeAgreementSource(t, "main.ts", row.source)})
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			entry := program.Files()[0]
			checked, release := program.Checker(context.Background(), entry)
			defer release()
			l := &lowering{program: program, checker: checked, result: &ir.Program{}, this: -1, functionIndex: -1, entry: entry, elementKey: &viewElementKey{}}
			l.moduleFormat = ModuleFormatEcmaScript
			l.result.Source = filepath.Base(program.FileName(entry))
			var receiver *ast.Node
			var find ast.Visitor
			find = func(node *ast.Node) bool {
				if node.Kind == ast.KindPropertyAccessExpression && node.AsPropertyAccessExpression().Name().Text() == "get" {
					receiver = node.AsPropertyAccessExpression().Expression
					return true
				}
				node.ForEachChild(find)
				return false
			}
			entry.AsNode().ForEachChild(find)
			if receiver == nil {
				t.Fatal("no optional get found")
			}
			t.Logf("ROW %s: structural receiver hazard %v", row.name, l.optionalMapStructuralReceiver(receiver))
		})
	}
}
