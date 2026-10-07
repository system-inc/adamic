package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/rule"
)

type Wave01Stand struct {
	Parameters, Declared, Resolved []int
	ArgumentCount, ArgumentIndex   int
	Rest                           bool
	ReturnType, PositionType       int
	Expected                       map[int][]int
}

func Wave01AwaitStandCapture(ctx rule.Context, n *ast.Node) []Wave01Stand {
	resolved := checker.Checker_getResolvedSignature(ctx.TypeChecker, n, nil, checker.CheckModeNormal)
	if resolved == nil || resolved.Target() == nil {
		return nil
	}
	declared := resolved.Target()
	if len(declared.TypeParameters()) == 0 {
		return nil
	}
	var out []Wave01Stand
	for index := range n.Arguments() {
		c := Wave01Stand{Parameters: []int{}, Declared: []int{}, Resolved: []int{}, ArgumentCount: len(n.Arguments()), ArgumentIndex: index, Rest: declared.HasRestParameter(), Expected: map[int][]int{}}
		for _, t := range declared.TypeParameters() {
			c.Parameters = append(c.Parameters, int(t.Id()))
		}
		for _, p := range declared.Parameters() {
			c.Declared = append(c.Declared, int(checker.Checker_getTypeOfSymbol(ctx.TypeChecker, p).Id()))
		}
		for _, p := range resolved.Parameters() {
			c.Resolved = append(c.Resolved, int(checker.Checker_getTypeOfSymbol(ctx.TypeChecker, p).Id()))
		}
		c.ReturnType = int(checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, declared).Id())
		if t := checker.Checker_getContextualType(ctx.TypeChecker, n, checker.ContextFlagsNone); t != nil {
			c.PositionType = int(t.Id())
		}
		for t, parts := range requireAwaitTypeParameterStandIns(ctx, n, resolved, declared, index) {
			for _, part := range parts {
				c.Expected[int(t.Id())] = append(c.Expected[int(t.Id())], int(part.Id()))
			}
		}
		out = append(out, c)
	}
	return out
}
