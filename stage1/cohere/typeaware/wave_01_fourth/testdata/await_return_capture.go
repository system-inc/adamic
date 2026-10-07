package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

type Wave01Return struct {
	Expression, Arrow, Declared bool
	Flags                       int
	Expected                    bool
}

func Wave01AwaitReturnCapture(ctx rule.Context, n *ast.Node) []Wave01Return {
	return []Wave01Return{{Expression: ast.IsFunctionExpression(n), Arrow: ast.IsArrowFunction(n), Declared: n.Type() != nil, Flags: int(ast.GetFunctionFlags(n)), Expected: requireAwaitReturnTakesContextualType(n)}}
}
