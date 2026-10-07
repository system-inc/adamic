package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	checking "github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
)

type Wave01ContextFacts struct {
	Property, StringIndex, NumberIndex int
	Tuple                              bool
	Elements, Returns                  []int
}
type Wave01Context struct {
	Kind, Index int
	Facts       []Wave01ContextFacts
	Expected    []int
}

func Wave01AwaitContextCapture(ctx rule.Context, n *ast.Node) []Wave01Context {
	var types []*checker.Type
	for _, s := range ctx.SourceFile.Statements.Nodes {
		if s.Kind == ast.KindTypeAliasDeclaration {
			types = append(types, ctx.TypeChecker.GetTypeAtLocation(s.Type()))
		}
	}
	id := func(t *checker.Type) int {
		if t == nil {
			return 0
		}
		return int(t.Id())
	}
	var out []Wave01Context
	for kind := 0; kind < 3; kind++ {
		for index := 0; index < 4; index++ {
			step := requireAwaitContextStep{kind: requireAwaitContextStepKind(kind), name: "run", index: index}
			c := Wave01Context{Kind: kind, Index: index, Expected: []int{}}
			for _, t := range types {
				f := Wave01ContextFacts{Elements: []int{}, Returns: []int{}}
				property := checker.Checker_getPropertyOfType(ctx.TypeChecker, t, "run")
				if property != nil {
					f.Property = id(checker.Checker_getTypeOfSymbol(ctx.TypeChecker, property))
				}
				f.StringIndex = id(checker.Checker_getIndexTypeOfType(ctx.TypeChecker, t, checker.Checker_stringType(ctx.TypeChecker)))
				f.NumberIndex = id(checker.Checker_getIndexTypeOfType(ctx.TypeChecker, t, checker.Checker_numberType(ctx.TypeChecker)))
				f.Tuple = checker.Checker_isArrayOrTupleType(ctx.TypeChecker, t) && !checker.Checker_isArrayType(ctx.TypeChecker, t)
				if f.Tuple {
					for _, e := range checker.Checker_getTypeArguments(ctx.TypeChecker, t) {
						f.Elements = append(f.Elements, id(e))
					}
				}
				for _, s := range checking.GetCallSignatures(ctx.TypeChecker, t) {
					f.Returns = append(f.Returns, id(checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, s)))
				}
				c.Facts = append(c.Facts, f)
			}
			for _, t := range requireAwaitApplyContextStep(ctx, types, step) {
				c.Expected = append(c.Expected, id(t))
			}
			out = append(out, c)
		}
	}
	return out
}
