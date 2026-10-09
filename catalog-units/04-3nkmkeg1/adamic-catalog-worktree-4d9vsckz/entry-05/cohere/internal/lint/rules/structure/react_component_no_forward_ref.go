package structure

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/imports"
	"github.com/system-inc/cohere/internal/lint/ecmascript/react"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// messageNoForwardRefCallText is the rule's `noForwardRefCall` message, whose wording lives in
// `policy/messages/react-component-no-forward-ref.json`.
var messageNoForwardRefCallText = policy.MessageOf("structure/react-component-no-forward-ref", "noForwardRefCall")

// messageNoForwardRefCall is the finding, rendered when it is reported so the text comes from the
// current catalog.
func messageNoForwardRefCall() rule.Message {
	return rule.Message{Id: messageNoForwardRefCallText.Id, Description: messageNoForwardRefCallText.Render(nil)}
}

// messageNoForwardRefImportText is the rule's `noForwardRefImport` message, whose wording lives in
// `policy/messages/react-component-no-forward-ref.json`.
var messageNoForwardRefImportText = policy.MessageOf("structure/react-component-no-forward-ref", "noForwardRefImport")

// messageNoForwardRefImport is the finding, rendered when it is reported so the text comes from the
// current catalog.
func messageNoForwardRefImport() rule.Message {
	return rule.Message{Id: messageNoForwardRefImportText.Id, Description: messageNoForwardRefImportText.Render(nil)}
}

// forwardRefTypeNames are the React types that only exist to describe a forwardRef wrapper.
//
// Flagged because they are how the wrapper survives a refactor: someone removes the forwardRef call
// and leaves the annotation, and the type keeps the old shape alive in every consumer's editor.
var forwardRefTypeNames = map[string]bool{
	"ForwardRefExoticComponent": true,
	"ForwardedRef":              true,
}

// ReactComponentNoForwardRef flags forwardRef in each of the three shapes it appears in.
//
//	valid:   function Field(properties: { ref?: Ref<HTMLInputElement> }) { ... }
//	invalid: import { forwardRef } from 'react'
//	invalid: React.forwardRef(function Field() { ... })
//	invalid: forwardRef(function Field() { ... })
//	invalid: const Field: React.ForwardRefExoticComponent<Properties> = ...
//
// Three shapes rather than one, because removing forwardRef from a codebase means removing all
// three and each survives the others. A call can go while the import stays; both can go while a
// type annotation keeps the wrapper's shape in every consumer's editor.
//
// The bare call is matched on the name alone, with no check that it came from react. That is the
// original's choice and its known limit: a local function named forwardRef would be reported. It is
// the right direction here, since the alternative needs the binding resolved and would go quiet on
// the aliased import (`import { forwardRef as fr }`) that the import arm exists to catch anyway.
var ReactComponentNoForwardRef = rule.Rule{
	Name: "structure/react-component-no-forward-ref",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindImportDeclaration: func(node *ast.Node) {
				reportForwardRefImport(ctx, node)
			},

			ast.KindCallExpression: func(node *ast.Node) {
				callee := ast.SkipParentheses(node.AsCallExpression().Expression)
				if callee == nil {
					return
				}

				switch callee.Kind {
				case ast.KindIdentifier:
					if callee.Text() == "forwardRef" {
						ctx.ReportNode(callee, messageNoForwardRefCall())
					}
				case ast.KindPropertyAccessExpression:
					if react.IsNamespacedMember(callee, func(name string) bool { return name == "forwardRef" }) {
						ctx.ReportNode(callee, messageNoForwardRefCall())
					}
				}
			},

			// A qualified type name, which is where the wrapper hides after the call is gone.
			ast.KindQualifiedName: func(node *ast.Node) {
				qualified := node.AsQualifiedName()
				left, right := qualified.Left, qualified.Right
				if left == nil || right == nil {
					return
				}
				if left.Kind == ast.KindIdentifier && left.Text() == "React" &&
					right.Kind == ast.KindIdentifier && forwardRefTypeNames[right.Text()] {
					ctx.ReportNode(node, messageNoForwardRefCall())
				}
			},
		}
	},
}

// reportForwardRefImport flags a named import of forwardRef from react, aliased or not.
//
// The alias is deliberately not what is tested. `import { forwardRef as wrap }` imports the same
// function under another name, so the imported name is what decides and the local name is what the
// call arm would have to have resolved. Reporting the specifier rather than the whole statement
// keeps the finding on the one name at fault when several are imported together.
func reportForwardRefImport(ctx rule.Context, node *ast.Node) {
	declaration := node.AsImportDeclaration()
	if declaration.ModuleSpecifier == nil || !ast.IsStringLiteralLike(declaration.ModuleSpecifier) {
		return
	}
	if declaration.ModuleSpecifier.Text() != "react" {
		return
	}
	if declaration.ImportClause == nil {
		return
	}

	// A default or namespace import brings in no name to test here. `import React from 'react'` is
	// how React.forwardRef is reached, and that is the call arm's business. `BindingsOf` separates
	// the three shapes, so reading `Named` alone is that distinction rather than an omission.
	for _, element := range imports.BindingsOf(node).Named {
		if imports.ImportedNameOf(element) == "forwardRef" {
			// The specifier node rather than the local name, so an aliased import blames the whole
			// `forwardRef as forward` rather than just the alias.
			ctx.ReportNode(element, messageNoForwardRefImport())
		}
	}
}
