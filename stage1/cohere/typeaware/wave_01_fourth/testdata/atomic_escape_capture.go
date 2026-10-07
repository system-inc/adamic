package core

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/system-inc/cohere/internal/lint/rule"
	"strings"
)

func Wave01AtomicEscapeCapture() string {
	file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/workspace/escape.ts"}, "let outside; function f(parameter){let inside;} function g(other){}", core.ScriptKindTS)
	root := file.Statements.Nodes[1]
	other := file.Statements.Nodes[2]
	var inner, outer *ast.Node
	var walk func(*ast.Node) bool
	walk = func(n *ast.Node) bool {
		if n.Kind == ast.KindVariableDeclaration {
			if n.Name().Text() == "inside" {
				inner = n
			} else {
				outer = n
			}
		}
		n.ForEachChild(walk)
		return false
	}
	walk(file.AsNode())
	var out strings.Builder
	for mode := 0; mode < 64; mode++ {
		declaration := outer
		if mode&8 != 0 {
			declaration = inner
		}
		if mode&1 != 0 {
			owner := other
			if mode&8 != 0 {
				owner = root
			}
			declaration = owner.AsFunctionDeclaration().Parameters.Nodes[0]
		}
		symbol := &ast.Symbol{Declarations: []*ast.Node{declaration}}
		if mode&4 != 0 {
			symbol.Declarations = nil
		}
		if mode&32 != 0 {
			symbol = nil
		}
		memo := map[*ast.Symbol]bool{}
		if mode&16 != 0 {
			memo[symbol] = true
		}
		ctx := rule.Context{SourceFile: file}
		binding := escapesEnclosingFunction(ctx, symbol, root, false, memo)
		member := escapesEnclosingFunction(ctx, symbol, root, mode&2 != 0, memo)
		fmt.Fprintf(&out, "%d %t %t\n", mode, binding, member)
	}
	return out.String()
}
