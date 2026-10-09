// Package next holds the rules from `@next/eslint-plugin-next`, ported from Vercel's originals.
//
// These are ports rather than adaptations. The judgment in them is Vercel's, made against how Next
// itself behaves, so rewriting the reasoning here would substitute our interpretation of a framework
// we do not own for the framework author's own. Where a rule's wording is expanded it is to say why
// rather than to change what fires.
//
// # Naming
//
// A rule here is named for itself and never for its family: `no-img-element`, not
// `next-no-img-element`. The config writes `nextjs/no-img-element` and the matcher strips the
// namespace on a `/` boundary, so a family prefix in the rule's own name means the config entry
// cannot match it. The rule then runs on zero files while every one of its tests passes, because a
// fixture exercises the rule directly and never reads the configuration.
//
// The first rule in this package shipped with a `next-` prefix and was inert. Nothing caught it
// except the coverage line, and only because someone planted a violation and noticed it did not
// fire. Twenty-two rules named the wrong way would have been twenty-two silent passes.
package next

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoAssignModuleVariable = rule.Message{
	Id: "noAssignModuleVariable",
	Description: "Declaring a variable named `module` shadows the CommonJS `module` object that " +
		"the bundler relies on, so code compiled into a CommonJS context loses its reference to " +
		"its own exports. The failure surfaces at runtime as an export that is silently missing " +
		"rather than as a build error. Rename the variable.",
}

// NoAssignModuleVariable flags a variable declaration named `module`.
//
//	valid:   let m = {}
//	valid:   const moduleName = 'x'
//	invalid: let module = {}
//	invalid: let a, module = {}
//
// Ported from `@next/next/no-assign-module-variable`.
//
// The rule keys on the declaration list rather than the individual declarator, which is Vercel's
// choice and is preserved: `let a, module = {}` reports once against the whole statement rather than
// against the offending name. Reporting per declarator would be a narrower range and a different
// rule, and choosing it here would be us overriding the upstream judgment on a rule we ported.
var NoAssignModuleVariable = rule.Rule{
	// The name carries no family prefix, and that is load-bearing rather than stylistic. The config
	// writes `nextjs/no-assign-module-variable`, and matching strips the namespace on a `/` boundary,
	// so a rule named `next-no-assign-module-variable` matches nothing and runs on no files while its
	// own tests pass. That is not hypothetical: this rule shipped that way and was inert until a
	// planted violation failed to fire.
	Name: "@next/next/no-assign-module-variable",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindVariableStatement: func(node *ast.Node) {
				statement := node.AsVariableStatement()
				if statement == nil || statement.DeclarationList == nil {
					return
				}

				declarationList := statement.DeclarationList.AsVariableDeclarationList()
				if declarationList == nil || declarationList.Declarations == nil {
					return
				}

				for _, declaration := range declarationList.Declarations.Nodes {
					// Only a plain identifier can shadow `module`. A destructuring pattern binds its
					// own names and never introduces one called `module` by writing the word, so
					// checking the name of a binding pattern would match text rather than a binding.
					name := declaration.AsVariableDeclaration().Name()
					if name == nil || name.Kind != ast.KindIdentifier {
						continue
					}
					if name.Text() != "module" {
						continue
					}

					ctx.ReportNode(node, messageNoAssignModuleVariable)
					return
				}
			},
		}
	},
}
