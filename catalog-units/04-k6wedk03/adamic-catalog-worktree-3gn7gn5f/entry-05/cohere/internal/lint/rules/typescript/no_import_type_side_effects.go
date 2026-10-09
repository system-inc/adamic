package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/imports"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageUseTopLevelQualifier = rule.Message{
	Id: "useTopLevelQualifier",
	Description: "Every name in this import carries its own inline `type` qualifier, so TypeScript " +
		"erases the names and leaves the bare `import 'module'` behind, which still runs the " +
		"module for its side effects. The import reads as type-only and is not. Move the " +
		"qualifier to the top level, where it removes the whole statement.",
}

// NoImportTypeSideEffects flags an import whose specifiers all carry an inline `type` qualifier.
//
//	valid:   import { type T, U } from 'mod'
//	valid:   import type { T } from 'mod'
//	valid:   import T, { type U } from 'mod'
//	invalid: import { type A } from 'mod'
//	invalid: import { type A as AA, type B as BB } from 'mod'
//
// Ported from typescript-eslint's `no-import-type-side-effects`, which is the authority.
//
// # The judgment is a universal over the specifier list, and both of its guards are load-bearing
//
// The original returns early on three separate conditions and each one is a real class of input
// rather than defensive coding. Measured against the installed rule at version 8.67.0:
//
//	import 'mod';                            silent, no clause at all
//	import {} from 'mod';                    silent, an empty specifier list
//	import type { T } from 'mod';            silent, already top-level, nothing to move
//	import A, { type B, type C } from 'mod'; silent, the default binding is not a type specifier
//	import * as T from 'mod';                silent, a namespace binding is not a type specifier
//	import { type T, U } from 'mod';         silent, one plain specifier disarms the whole statement
//
// The empty-list guard is the one that is easy to drop, because a universal quantifier over an empty
// list is vacuously true and the rule would then report `import {} from 'mod'` while offering a fix
// that produces `import type {} from 'mod'`. Upstream writes the guard explicitly and the corpus
// does not cover it, so it is measured above and pinned by a fixture below.
//
// A default or namespace binding is not an `ImportSpecifier` in ESTree, and upstream's loop bails on
// the first specifier that is not one. Here the same distinction is structural rather than a type
// test: `ImportClause.Name()` holds the default binding and `NamedBindings` holds either a namespace
// or the element list, so a clause carrying a default or a namespace is declined before the elements
// are ever examined.
//
// # The fixer's removal span is positional, not name-based
//
// Upstream removes `[typeKeyword.range[0], specifier.imported.range[0]]`, that is, from the start of the
// `type` keyword to the start of the imported name, and then inserts ` type` after the `import`
// keyword. That span choice, rather than "delete the word type", is what makes these all correct,
// each measured against the installed rule:
//
//	import { type as } from 'mod';           -> import type { as } from 'mod';
//	import { type as as as } from 'mod';     -> import type { as as as } from 'mod';
//	import { type type } from 'mod';         -> import type { type } from 'mod';
//	import { type default as A } from 'mod'; -> import type { default as A } from 'mod';
//	import { type 'a' as A } from 'mod';     -> import type { 'a' as A } from 'mod';
//
// A name-based removal would have to decide which `as` or which `type` is the keyword, and would
// have nothing to say about a string-literal module export name. The positional span never asks.
//
// It also decides what happens to whitespace, which the corpus never varies and which a port would
// otherwise guess at. Only the run between the keyword and the name is consumed; everything else in
// the statement survives byte for byte. Measured:
//
//	import {type A} from 'mod';              -> import type {A} from 'mod';
//	import   {   type   A   }   from 'mod';  -> import type   {   A   }   from 'mod';
//
// The second is the one worth reading twice: the three spaces after `import` are untouched and the
// inserted qualifier lands immediately after the `import` keyword rather than before the brace, so
// the result carries `import type   {`.
//
// The imported name is the SOURCE side of a rename, not the local binding. In `type A as AA`, upstream
// removes up to `A`. Here that is `PropertyName` when a rename is present and the specifier's own name
// otherwise, which is the opposite of the reflexive choice: `Name()` alone would remove `type A as `
// and produce `import type { AA } from 'mod'`, silently changing which export is imported.
var NoImportTypeSideEffects = rule.Rule{
	Name: "@typescript-eslint/no-import-type-side-effects",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindImportDeclaration: func(node *ast.Node) {
				declaration := node.AsImportDeclaration()
				if declaration.ImportClause == nil {
					return
				}
				clause := declaration.ImportClause.AsImportClause()
				if clause.IsTypeOnly() {
					return
				}
				// A default binding (`import A, {...}`) is not a type specifier, and neither is a
				// namespace binding (`import * as A`). Either one leaves a runtime import behind
				// no matter what the named elements say, and BindingsOf reports each separately.
				bindings := imports.BindingsOf(node)
				if bindings.Default != nil || bindings.Namespace != nil {
					return
				}
				if len(bindings.Named) == 0 {
					return
				}
				for _, element := range bindings.Named {
					if element.Kind != ast.KindImportSpecifier || !element.AsImportSpecifier().IsTypeOnly {
						return
					}
				}

				fixes := make([]rule.Fix, 0, len(bindings.Named)+1)
				for _, element := range bindings.Named {
					// Both ends are token ranges rather than node ranges. `Pos()` on the imported
					// name sits immediately after the `type` keyword, before the whitespace that
					// separates them, so ending the removal there leaves that whitespace behind and
					// writes `{  A }` for `{ type A }`. The token range starts at the name itself.
					fixes = append(fixes, rule.RemoveRange(
						rule.TokenRange(ctx.SourceFile, element).
							WithEnd(rule.TokenRange(ctx.SourceFile, importedNameOf(element)).Pos()),
					))
				}
				// The qualifier goes immediately after the `import` keyword, which is the first
				// token of the declaration, so its own token range's end is the insertion point.
				importKeyword := rule.TokenRange(ctx.SourceFile, node)
				fixes = append(fixes, rule.ReplaceRange(
					importKeyword.WithEnd(importKeyword.Pos()+len("import")).
						WithPos(importKeyword.Pos()+len("import")),
					" type",
				))

				ctx.ReportNodeWithFixes(node, messageUseTopLevelQualifier, fixes...)
			},
		}
	},
}

// importedNameOf returns the SOURCE-side name of an import specifier, which is what upstream's
// `specifier.imported` names.
//
// For `A as AA` that is `A`, carried on PropertyName; for a bare `A` there is no PropertyName and the
// specifier's own name is both sides at once. Taking Name() unconditionally would remove `type A as `
// from a renaming specifier and rewrite `import { type A as AA }` into `import type { AA }`, which
// imports a different export and still compiles wherever `AA` happens to exist upstream.
func importedNameOf(specifier *ast.Node) *ast.Node {
	if propertyName := specifier.AsImportSpecifier().PropertyName; propertyName != nil {
		return propertyName
	}
	return specifier.Name()
}
