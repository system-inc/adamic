package consistentreturn

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"testing"
)

func TestAdamicReturnControls(t *testing.T) {
	texts := []string{
		"function foo(){return;} function Foo(){return undefined;} function é(){return (undefined);} function É(){return void 1;}",
		"class A{static async *#foo(){return 1;} get bar(){return 2;} set bar(x){return;} constructor(){return;} ['computed'](){return 3;} [dynamic](){return 4;} 0x10(){return 5;}}",
		"const x={foo:function named(){return true;}};const y=()=>{return undefined};function *g(){return 1};",
	}
	for _, s := range texts {
		f := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/control.ts", PathKey: "/control.ts"}, s, core.ScriptKindTS)
		var visit func(*ast.Node) bool
		visit = func(n *ast.Node) bool {
			IsScope(n)
			IsGenerator(n)
			isExemptFromEndJudgment(n)
			if n.Kind == ast.KindSourceFile || adamicRawIsScope(n) {
				Name(n, false)
				Name(n, true)
				staticName(n)
				ReportRange(n, func(a *ast.Node) core.TextRange { return a.Loc })
			}
			if n.Kind == ast.KindReturnStatement {
				for _, u := range []bool{false, true} {
					for _, h := range []int{0, 1, 2} {
						hooks := Hooks{}
						if h > 0 {
							value := h == 2
							hooks.TreatsArgumentAsUnspecified = func(*ast.Node) bool { return value }
						}
						HasValue(n, Settings{TreatUndefinedAsUnspecified: u}, hooks)
					}
				}
			}
			return n.ForEachChild(visit)
		}
		visit(f.AsNode())
	}
	for _, s := range []string{"", "missingReturnValue", "unexpectedReturnValue", "missingReturn", "other"} {
		Verb(s)
	}
	// Actual Name callers start with an ASCII kind; cover that complete first-byte domain.
	for i := 0; i < 128; i++ {
		capitaliseFirst(string(rune(i))+"tail", false)
		capitaliseFirst(string(rune(i))+"tail", true)
	}
	capitaliseFirst("", true)
}
