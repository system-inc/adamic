package core

import (
	"encoding/json"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/system-inc/cohere/internal/lint/rule"
)

type wave01AwaitReport struct {
	Start, End, AsyncStart, AsyncEnd  int
	Arrow, Method, Constructor, Named bool
	Name, Replacement                 string
	Expected                          []any
}

func Wave01AwaitReportCapture() []byte {
	texts := []string{"async function foo(){work();}", "const f = async function named(){work();};", "const f = async function(){work();};", "const f = async () => {work();};", "const o = { run: async function own(){work();} };", "const o = { run: async () => {work();} };", "class A { async run(){work();} }", "class A { async \"\"(){work();} }", "class A { async [key](){work();} }", "class A { field=0\nasync [key](){work();} }", "class A { field=0\nasync in(){work();} }", "class A { field=0;\nasync [key](){work();} }", "// 😀\nasync function 工作(){work();}", "async /* keep */ function foo(){work();}", "export default async function foo(){work();}"}
	var results []wave01AwaitReport
	for _, text := range texts {
		file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/workspace/await-report.ts"}, text, core.ScriptKindTS)
		var visit func(*ast.Node) bool
		visit = func(n *ast.Node) bool {
			if n.ModifierFlags()&ast.ModifierFlagsAsync != 0 {
				switch n.Kind {
				case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction, ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor:
					holder := n
					method := n.Kind == ast.KindMethodDeclaration || n.Kind == ast.KindGetAccessor || n.Kind == ast.KindSetAccessor
					if p := n.Parent; p != nil && p.Kind == ast.KindPropertyAssignment && p.AsPropertyAssignment().Initializer == n {
						holder = p
						method = true
					}
					name, named := "", false
					if holder.Name() != nil {
						name, named = requireAwaitNameText(holder.Name())
					}
					c := wave01AwaitReport{Arrow: n.Kind == ast.KindArrowFunction, Method: method, Constructor: n.Kind == ast.KindConstructor, Named: named, Name: name, AsyncStart: -1}
					ctx := rule.Context{SourceFile: file, Report: func(d rule.Diagnostic) {
						c.Start = d.Range.Pos()
						c.End = d.Range.End()
						c.Expected = []any{d.Range.Pos(), d.Range.End(), "require-await", d.Message.Id, d.Message.Description, len(d.Fixes), len(d.Suggestions)}
						for _, fix := range d.Fixes {
							c.Expected = append(c.Expected, fix.Range.Pos(), fix.Range.End(), fix.Text)
						}
						for _, suggestion := range d.Suggestions {
							c.Expected = append(c.Expected, suggestion.Message.Id, suggestion.Message.Description, len(suggestion.Fixes))
							for _, fix := range suggestion.Fixes {
								c.Expected = append(c.Expected, fix.Range.Pos(), fix.Range.End(), fix.Text)
							}
						}
						results = append(results, c)
					}}
					if span, ok := requireAwaitAsyncKeywordRange(ctx, n); ok {
						c.AsyncStart = span.Pos()
						c.AsyncEnd = span.End()
						c.Replacement = requireAwaitReplacement(ctx, n)
					}
					checkRequireAwait(ctx, n)
				}
			}
			n.ForEachChild(visit)
			return false
		}
		visit(file.AsNode())
	}
	data, err := json.Marshal(results)
	if err != nil {
		panic(err)
	}
	return data
}
