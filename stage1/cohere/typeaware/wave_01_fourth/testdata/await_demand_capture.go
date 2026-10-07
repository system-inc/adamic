package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	checking "github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
)

type Wave01Demand struct {
	Start      int
	Signatures [][]bool
	Expected   bool
}

func Wave01AwaitDemandCapture(ctx rule.Context, n *ast.Node) Wave01Demand {
	result := Wave01Demand{Start: n.Pos()}
	contextual := checker.Checker_getContextualType(ctx.TypeChecker, n, checker.ContextFlagsNone)
	if contextual == nil {
		return result
	}
	for _, part := range checking.UnionTypeParts(contextual) {
		for _, signature := range checking.GetCallSignatures(ctx.TypeChecker, part) {
			var branches []bool
			returns := checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, signature)
			for _, branch := range checking.UnionTypeParts(returns) {
				branches = append(branches, checking.IsThenableType(ctx.TypeChecker, n, branch))
			}
			result.Signatures = append(result.Signatures, branches)
		}
	}
	result.Expected = requireAwaitTypesDemandPromise(ctx, n, []*checker.Type{contextual}, nil)
	return result
}
