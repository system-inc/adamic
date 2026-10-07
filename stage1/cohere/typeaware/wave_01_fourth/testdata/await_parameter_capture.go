package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/rule"
)

type Wave01Parameter struct {
	Parameters                   []int
	Rest                         bool
	RestElement, Index, Expected int
}

func Wave01AwaitParameterCapture(ctx rule.Context, n *ast.Node) []Wave01Parameter {
	resolved := checker.Checker_getResolvedSignature(ctx.TypeChecker, n, nil, checker.CheckModeNormal)
	if resolved == nil {
		return nil
	}
	signature := resolved.Target()
	if signature == nil {
		signature = resolved
	}
	parameters := signature.Parameters()
	var ids []int
	for _, p := range parameters {
		ids = append(ids, int(checker.Checker_getTypeOfSymbol(ctx.TypeChecker, p).Id()))
	}
	if ids == nil {
		ids = []int{}
	}
	element := 0
	if signature.HasRestParameter() {
		rest := checker.Checker_getTypeOfSymbol(ctx.TypeChecker, parameters[len(parameters)-1])
		if t := checker.Checker_getIndexTypeOfType(ctx.TypeChecker, rest, checker.Checker_numberType(ctx.TypeChecker)); t != nil {
			element = int(t.Id())
		}
	}
	var out []Wave01Parameter
	for index := 0; index < len(parameters)+3; index++ {
		c := Wave01Parameter{Parameters: ids, Rest: signature.HasRestParameter(), RestElement: element, Index: index}
		if t := requireAwaitParameterTypeAt(ctx, signature, index); t != nil {
			c.Expected = int(t.Id())
		}
		out = append(out, c)
	}
	return out
}
