//go:build lintoracle

package decorators

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"testing"
)

func TestAdamicDecoratorControls(t *testing.T) {
	for _, source := range []string{"@Foo() @ns.Bar() class C { @Bare public method(@Foo() value:string){} }", "class C { @Foo<T>() @Bar() static prop=1; constructor(@Foo() public x:number){} }", "@((Foo))() class C {}", "@日本語() class C {}"} {
		file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: tspath.RootedFilePath("/controls.ts"), PathKey: tspath.PathKey("/controls.ts")}, source, core.ScriptKindTS)
		var visit func(*ast.Node)
		visit = func(n *ast.Node) {
			CallName(n)
			Of(n)
			HasDecoratorInSet(n, map[string]struct{}{"": {}, "Foo": {}, "日本語": {}})
			n.ForEachChild(func(c *ast.Node) bool { visit(c); return false })
		}
		visit(file.AsNode())
	}
	CallName(nil)
	Of(nil)
	HasDecoratorInSet(nil, map[string]struct{}{"": {}})
}
