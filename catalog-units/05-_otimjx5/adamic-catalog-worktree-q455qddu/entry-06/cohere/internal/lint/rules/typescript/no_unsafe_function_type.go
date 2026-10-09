// Package typescript holds the rules from `@typescript-eslint/eslint-plugin`, ported from the
// originals.
//
// These are ports rather than adaptations. The judgment is typescript-eslint's, made against how
// TypeScript's own type system behaves, so rewriting the reasoning would substitute our reading of
// the language for the people who track it release by release.
//
// # Naming
//
// A rule here is named for itself and never for its family: `no-unsafe-function-type`, not
// `typescript-no-unsafe-function-type`. The config writes `typescript/no-unsafe-function-type` and
// the matcher strips the namespace on a `/` boundary, so a family prefix in the rule's own name
// means the config entry cannot match it. The rule then runs on zero files while every one of its
// tests passes, because a fixture exercises the rule directly and never reads the configuration.
//
// The first rule in the `next` package shipped with that defect and was inert until a planted
// violation failed to fire. `TestEveryRegisteredRuleIsReachableFromTheLiveConfig` catches it now,
// but reading the config costs ten seconds and does not depend on remembering that the guard exists.
package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/module"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageBannedFunctionType = rule.Message{
	Id: "bannedFunctionType",
	Description: "The `Function` type accepts any function-like value, so it constrains nothing " +
		"beyond callability: it says nothing about parameters, nothing about the return type, and " +
		"a value typed this way can be called with any arguments and produces `any`. That is the " +
		"opposite of what a type annotation is for. Write the signature the value actually has, " +
		"such as `(input: string) => number`, or `() => void` when it takes nothing and returns " +
		"nothing.",
}

// NoUnsafeFunctionType flags the built-in `Function` type.
//
//	valid:   let handler: () => void
//	valid:   let handler: (input: string) => number
//	invalid: let handler: Function
//	invalid: interface Thing extends Function {}
//	invalid: class Thing implements Function {}
//
// Ported from `@typescript-eslint/no-unsafe-function-type`.
//
// The original guards the identifier with `isReferenceToGlobalFunction`, so a locally declared
// `Function` is not flagged. That guard needs symbol resolution, which our checker shim does not
// expose yet — see the note at the reference below. Rather than silently drop it, the syntactic
// approximation here is narrowed in the one direction that matters: a file that declares its own
// `Function` type or interface is skipped entirely, so a local declaration cannot be flagged. The
// remaining divergence from the original is an import of a non-global `Function`, which shadows the
// global without declaring it locally. That is rare enough to be worth naming rather than worth
// blocking the rule on, and the clean fixtures pin the behavior either way.
var NoUnsafeFunctionType = rule.Rule{
	Name: "@typescript-eslint/no-unsafe-function-type",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		// One pass over the file's own statements, before any node is visited, to learn whether the
		// name is shadowed here. Declining costs this scan; being wrong costs a false positive on a
		// type the author defined themselves.
		if module.DeclaresTypeNamed(ctx.SourceFile, "Function") {
			return nil
		}

		report := func(node *ast.Node) {
			if node == nil || node.Kind != ast.KindIdentifier || node.Text() != "Function" {
				return
			}
			ctx.ReportNode(node, messageBannedFunctionType)
		}

		// One listener covers all three positions, which is a real difference from the original and
		// worth stating. The ESLint tree this was ported from distinguishes `TSTypeReference` from
		// `TSInterfaceHeritage` and `TSClassImplements`, so the original registers three visitors.
		// typescript-go represents all three as a type reference: verified by walking
		// `interface Thing extends Function {}` and `class Thing implements Function {}` and observing
		// only `TypeReference` fire.
		//
		// A second listener for `ExpressionWithTypeArguments` was written here first, mirroring the
		// original's shape, and it was dead code. The fixtures passed with it present and passed
		// identically with it removed, which is how it was found: mutation testing, not review.
		return rule.Listeners{
			ast.KindTypeReference: func(node *ast.Node) {
				typeReference := node.AsTypeReferenceNode()
				if typeReference == nil {
					return
				}
				report(typeReference.TypeName)
			},
		}
	},
}
