package core

import (
	"encoding/json"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/system-inc/cohere/internal/lint/rule"
)

type wave01SyntaxCase struct {
	Nodes   [][]any
	Results [][]any
}

func Wave01AtomicSyntaxCapture() []byte {
	texts := []string{"foo = await bar;", "foo += await bar;", "foo.bar = await baz;", "foo[bar].baz += await q;", "a.foo = a.foo + await baz;", "let foo = await bar;", "function f(foo){foo.bar = 1;}", "const {foo: bar = foo} = obj;", "foo.bar + bar.foo;", "foo[bar] = 1;", "foo.bar ??= await baz;", "foo.bar &&= baz;", "foo.bar ||= baz;", "foo.bar **= baz;", "foo.bar <<= baz;", "foo.bar >>= baz;", "foo.bar >>>= baz;", "foo.bar &= baz;", "foo.bar |= baz;", "foo.bar ^= baz;", "foo.bar %= baz;", "foo.bar /= baz;", "foo.bar -= baz;", "foo.bar *= baz;"}
	var captures []wave01SyntaxCase
	for _, text := range texts {
		file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/workspace/atomic.ts"}, text, core.ScriptKindTS)
		var nodes []*ast.Node
		ids := map[*ast.Node]int{nil: -1}
		var visit func(*ast.Node) bool
		visit = func(n *ast.Node) bool {
			ids[n] = len(nodes)
			nodes = append(nodes, n)
			n.ForEachChild(visit)
			return false
		}
		visit(file.AsNode())
		c := wave01SyntaxCase{}
		for i, n := range nodes {
			member, left, name := -1, -1, -1
			operator := ""
			switch n.Kind {
			case ast.KindPropertyAccessExpression:
				member = ids[n.AsPropertyAccessExpression().Expression]
			case ast.KindElementAccessExpression:
				member = ids[n.AsElementAccessExpression().Expression]
			case ast.KindBinaryExpression:
				left = ids[n.AsBinaryExpression().Left]
				t := n.AsBinaryExpression().OperatorToken
				r := rule.TokenRange(file, t)
				operator = text[r.Pos():r.End()]
			case ast.KindVariableDeclaration, ast.KindBindingElement, ast.KindParameter:
				name = ids[n.Name()]
			}
			c.Nodes = append(c.Nodes, []any{ids[n.Parent], member, left, name, operator})
			if n.Kind == ast.KindIdentifier {
				assignment, ok := propertyAssignmentHeadedBy(n)
				assignmentID := -1
				if ok {
					assignmentID = ids[assignment]
				}
				c.Results = append(c.Results, []any{i, isDeclarationName(n), isPlainAssignmentTarget(n), assignmentID})
			}
		}
		captures = append(captures, c)
	}
	data, err := json.Marshal(captures)
	if err != nil {
		panic(err)
	}
	return data
}
