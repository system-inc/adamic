package nexus

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// ImportStyle is how a package expects to be bound.
//
// The two forms are not interchangeable: a namespace object is frozen and, under esModuleInterop,
// is not the value the default import yields. So the correct style is a property of the package
// rather than a house preference, and pinning the name without pinning the form would leave half
// the drift in place.
type ImportStyle string

const (
	ImportStyleDefault   ImportStyle = "Default"
	ImportStyleNamespace ImportStyle = "Namespace"
)

// ModuleAlias is the binding one package must be imported under.
//
// Style is empty rather than absent when unset, and empty reads as Default, which is what almost
// every package wants. Leaving it genuinely optional would make the shorthand quietly weaker than
// it looks: it would pin the name while permitting both shapes, which is the drift the rule exists
// to stop.
type ModuleAlias struct {
	Name  string      `json:"name"`
	Style ImportStyle `json:"style"`
}

// ImportRequireModuleAliasOptions names extra packages to pin, merged over the built-in defaults.
type ImportRequireModuleAliasOptions struct {
	Modules map[string]ModuleAlias `json:"modules"`
}

// defaultModuleAliases are the packages every consumer gets without asking.
//
// These are not house style, which is why they can live in a framework-agnostic library: both have
// a single correct answer that no codebase disagrees with. React's ecosystem binds it as `React`
// and emits the default form everywhere, and `typescript.d.ts` ends in `export = ts` with no
// default export at all, so the namespace form is the only accurate one rather than a preference.
//
// A consumer that genuinely wants something else re-states the entry and wins, since the configured
// map is merged over this one.
var defaultModuleAliases = map[string]ModuleAlias{
	"react":      {Name: "React", Style: ImportStyleDefault},
	"typescript": {Name: "TypeScript", Style: ImportStyleNamespace},
}

// importRequireModuleAliasRequireAliasNameText is the rule's `requireAliasName` message,
// whose wording lives in `policy/messages/import-require-module-alias.json`.
var importRequireModuleAliasRequireAliasNameText = policy.MessageOf("nexus/import-require-module-alias", "requireAliasName")

func messageRequireAliasName(source string, expected string, actual string) rule.Message {
	return rule.Message{
		Id: "requireAliasName",
		Description: importRequireModuleAliasRequireAliasNameText.Render(map[string]string{
			"source":   source,
			"expected": expected,
			"actual":   actual,
		}),
	}
}

// importRequireModuleAliasRequireDefaultStyleText is the rule's `requireDefaultStyle` message,
// whose wording lives in `policy/messages/import-require-module-alias.json`.
var importRequireModuleAliasRequireDefaultStyleText = policy.MessageOf("nexus/import-require-module-alias", "requireDefaultStyle")

func messageRequireDefaultStyle(source string, expected string) rule.Message {
	return rule.Message{
		Id: "requireDefaultStyle",
		Description: importRequireModuleAliasRequireDefaultStyleText.Render(map[string]string{
			"source":   source,
			"expected": expected,
		}),
	}
}

// importRequireModuleAliasRequireNamespaceStyleText is the rule's `requireNamespaceStyle` message,
// whose wording lives in `policy/messages/import-require-module-alias.json`.
var importRequireModuleAliasRequireNamespaceStyleText = policy.MessageOf("nexus/import-require-module-alias", "requireNamespaceStyle")

func messageRequireNamespaceStyle(source string, expected string) rule.Message {
	return rule.Message{
		Id: "requireNamespaceStyle",
		Description: importRequireModuleAliasRequireNamespaceStyleText.Render(map[string]string{
			"source":   source,
			"expected": expected,
		}),
	}
}

// ImportRequireModuleAlias pins the binding a package is imported under, so one package reads the
// same way in every file.
//
//	valid:   import React from 'react'
//	valid:   import type React from 'react'
//	valid:   import * as TypeScript from 'typescript'
//	invalid: import Reakt from 'react'
//	invalid: import * as React from 'react'
//	invalid: import * as ts from 'typescript'
//	invalid: import TypeScript from 'typescript'
//
// Why pin the name: a package imported as `ts` in one file and `typeScript` in another is the same
// dependency wearing three faces, and every reader pays to recognize it again. Pinning it makes
// `TypeScript.isStringLiteral` mean the same thing everywhere and makes the package greppable
// across the tree by a single spelling.
//
// Named imports are untouched. `import { useState } from 'react'` is a different question, owned by
// `react-import-no-destructuring` for React and left alone elsewhere. This rule only has an opinion
// once a default or namespace binding exists.
//
// No fix, deliberately, and this is a departure from the TypeScript original.
//
// The original autofixes both halves. The name fix guards itself with a scope query, rewriting the
// specifier only when the old binding is referenced nowhere else in the file, because a fix that
// renames a binding without rewriting its references breaks the file. We have no scope analysis, so
// that guard cannot be ported, and the fix without its guard is exactly the unsafe rewrite the
// guard exists to prevent.
//
// The style fix has the same defect and no guard at all in the original: rewriting `* as ts` to
// `TypeScript` changes the binding name in the same stroke, so every existing `ts.` reference in the
// file dangles. Porting the fix would be shipping a rewrite that compiles to a broken file, and the
// edit engine's parse guard cannot catch it, since the corrupted output still parses.
//
// The repair is a rename, which an editor does correctly with the whole file in view. The rule's
// job here is to say which name and which form, and to say why.
var ImportRequireModuleAlias = rule.Rule{
	Name: "nexus/import-require-module-alias",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		// Configured entries win, so a consumer can override a default rather than only add to it.
		configuredModules := make(map[string]ModuleAlias, len(defaultModuleAliases))
		for source, alias := range defaultModuleAliases {
			configuredModules[source] = alias
		}
		if settings, hasSettings := rule.OptionsAs[ImportRequireModuleAliasOptions](options); hasSettings {
			for source, alias := range settings.Modules {
				if alias.Name == "" {
					continue
				}
				if alias.Style == "" {
					alias.Style = ImportStyleDefault
				}
				configuredModules[source] = alias
			}
		}

		return rule.Listeners{
			ast.KindImportDeclaration: func(node *ast.Node) {
				declaration := node.AsImportDeclaration()
				if declaration == nil || declaration.ModuleSpecifier == nil {
					return
				}
				if !ast.IsStringLiteralLike(declaration.ModuleSpecifier) {
					return
				}

				alias, isConfigured := configuredModules[declaration.ModuleSpecifier.Text()]
				if !isConfigured {
					return
				}
				source := declaration.ModuleSpecifier.Text()

				// `import 'package'` for its side effects has no clause and so no binding to pin.
				if declaration.ImportClause == nil {
					return
				}
				clause := declaration.ImportClause.AsImportClause()
				if clause == nil {
					return
				}

				// The TypeScript parser hands the original one flat `specifiers` list holding the
				// default binding, the namespace binding, and the named ones together, so it filters
				// out `ImportSpecifier` and treats what remains uniformly. Here those are three
				// different places in the tree: the default binding is the clause's own name, and
				// the namespace binding lives under NamedBindings alongside NamedImports. Checking
				// only one of them would silently pin half the forms, so both are read, and
				// NamedImports is skipped where the original's filter skipped it.
				defaultBinding := clause.Name()
				var namespaceBinding *ast.Node
				if clause.NamedBindings != nil && clause.NamedBindings.Kind == ast.KindNamespaceImport {
					namespaceBinding = clause.NamedBindings
				}

				// Style first, because a namespace import bound to the right name is still the wrong
				// shape, and reporting the name would say nothing about what to change.
				expectsNamespace := alias.Style == ImportStyleNamespace

				if defaultBinding != nil {
					if expectsNamespace {
						ctx.ReportNode(defaultBinding, messageRequireNamespaceStyle(source, alias.Name))
					} else if defaultBinding.Text() != alias.Name {
						ctx.ReportNode(
							defaultBinding,
							messageRequireAliasName(source, alias.Name, defaultBinding.Text()),
						)
					}
				}

				if namespaceBinding != nil {
					namespaceName := namespaceBinding.AsNamespaceImport().Name()
					if !expectsNamespace {
						ctx.ReportNode(namespaceBinding, messageRequireDefaultStyle(source, alias.Name))
					} else if namespaceName != nil && namespaceName.Text() != alias.Name {
						ctx.ReportNode(
							namespaceName,
							messageRequireAliasName(source, alias.Name, namespaceName.Text()),
						)
					}
				}
			},
		}
	},
}
