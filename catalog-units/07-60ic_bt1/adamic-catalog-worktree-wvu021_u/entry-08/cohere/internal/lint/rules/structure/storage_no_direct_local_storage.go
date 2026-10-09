package structure

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// messageDirectLocalStorageText is the rule's `directLocalStorage` message, whose wording lives in
// `policy/messages/storage-no-direct-local-storage.json`.
var messageDirectLocalStorageText = policy.MessageOf("structure/storage-no-direct-local-storage", "directLocalStorage")

// messageDirectLocalStorage is the finding, rendered when it is reported so the text comes from the
// current catalog.
func messageDirectLocalStorage() rule.Message {
	return rule.Message{Id: messageDirectLocalStorageText.Id, Description: messageDirectLocalStorageText.Render(nil)}
}

// messageWindowLocalStorageText is the rule's `windowLocalStorage` message, whose wording lives in
// `policy/messages/storage-no-direct-local-storage.json`.
var messageWindowLocalStorageText = policy.MessageOf("structure/storage-no-direct-local-storage", "windowLocalStorage")

// messageWindowLocalStorage is the finding, rendered when it is reported so the text comes from the
// current catalog.
func messageWindowLocalStorage() rule.Message {
	return rule.Message{Id: messageWindowLocalStorageText.Id, Description: messageWindowLocalStorageText.Render(nil)}
}

// StorageNoDirectLocalStorage flags any direct use of localStorage outside the service.
//
//	valid:   localStorageService.set('key', value)
//	valid:   localStorage.getItem('key')          inside services/local-storage/
//	invalid: localStorage.getItem('key')
//	invalid: window.localStorage.getItem('key')
//	invalid: const storage = localStorage
//	invalid: createPersister({ storage: window.localStorage })
//
// Two message ids because the two spellings are found by different searches. Someone grepping for
// `localStorage.` finds the first and not the second, so telling them apart in the finding is what
// makes a sweep complete.
//
// # A deliberate divergence from the TypeScript original, and the one place tonight where the gate
// # and the correct rule come apart
//
// The original's window branch requires the `window.localStorage` node to be the object of an outer
// member access, so it matches `window.localStorage.getItem(...)` and never `window.localStorage`
// passed as a value. Its identifier branch does not catch that either, since it skips anything
// whose parent is a member access, which this is. Bare `window.localStorage` as a value falls
// between the two branches.
//
// There is exactly one occurrence in the tree, `NetworkService.ts:228`, passing it as a persister's
// storage, with no disable comment. It is a real violation of what the rule says it enforces and
// the gate cannot see it.
//
// So this rule reports it and the gate does not, and the differential will classify it
// `only-cohere`. That is correct rather than a regression: reproducing the gap would mean shipping
// a bug to keep a number matching. Ruled by `@system_cohere` after the discrepancy was measured.
//
// The six other direct uses in `services/network/internal/` all carry per-line disable comments, so
// the rule fires on them and the suppression layer withholds the findings, exactly as under the
// gate.
var StorageNoDirectLocalStorage = rule.Rule{
	Name: "structure/storage-no-direct-local-storage",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if FileContextFor(ctx.SourceFile.FileName().AsString()).IsLocalStorageServiceFile {
			// The service itself has to reach the real thing, or the rule forbids the thing it
			// recommends.
			return nil
		}

		return rule.Listeners{
			ast.KindIdentifier: func(node *ast.Node) {
				if node.Text() != "localStorage" {
					return
				}

				parent := node.Parent
				if parent == nil {
					return
				}

				switch parent.Kind {
				case ast.KindPropertyAccessExpression:
					access := parent.AsPropertyAccessExpression()

					// `window.localStorage`, whether called or passed. Reported once, on the whole
					// access, so the finding covers the expression a reader has to replace.
					if access.Name() == node {
						receiver := ast.SkipParentheses(access.Expression)
						if receiver != nil && receiver.Kind == ast.KindIdentifier && receiver.Text() == "window" {
							ctx.ReportNode(parent, messageWindowLocalStorage())
						}
						// A `localStorage` property on anything else is somebody's own field.
						return
					}

					// `localStorage.getItem(...)`, where the identifier is the receiver.
					ctx.ReportNode(node, messageDirectLocalStorage())

				case ast.KindPropertyAssignment:
					// A key named localStorage is a name, not a use. `{ localStorage: x }` stores
					// nothing. The shorthand form is a real use and is handled below, since its
					// node is a ShorthandPropertyAssignment rather than this.
					if parent.AsPropertyAssignment().Name() == node {
						return
					}
					ctx.ReportNode(node, messageDirectLocalStorage())

				case ast.KindPropertyDeclaration, ast.KindMethodDeclaration,
					ast.KindPropertySignature, ast.KindMethodSignature,
					ast.KindBindingElement, ast.KindParameter, ast.KindVariableDeclaration,
					ast.KindFunctionDeclaration, ast.KindClassDeclaration, ast.KindImportSpecifier:
					// Declaring or binding a name is not reaching for the global one.
					//
					// Only when the identifier IS the name. `const storage = localStorage` has a
					// VariableDeclaration parent too, and there the identifier is the initializer,
					// which is a real use. Matching on the parent kind alone silenced exactly that
					// case, and a fixture caught it.
					if declaredName(parent) == node {
						return
					}
					ctx.ReportNode(node, messageDirectLocalStorage())

				default:
					// Every other position is the value being used: an argument, an initializer, a
					// return, an element of an array.
					ctx.ReportNode(node, messageDirectLocalStorage())
				}
			},
		}
	},
}

// declaredName returns the name a declaration binds, or nil when the node has no simple name.
//
// Used to tell a declaration's name apart from the rest of the declaration. The kinds handled here
// are the ones whose parent kind alone would otherwise read as "this is a name".
func declaredName(node *ast.Node) *ast.Node {
	switch node.Kind {
	case ast.KindVariableDeclaration:
		return node.AsVariableDeclaration().Name()
	case ast.KindParameter:
		return node.AsParameterDeclaration().Name()
	case ast.KindBindingElement:
		return node.AsBindingElement().Name()
	case ast.KindPropertyDeclaration:
		return node.AsPropertyDeclaration().Name()
	case ast.KindPropertySignature:
		return node.AsPropertySignatureDeclaration().Name()
	case ast.KindMethodDeclaration:
		return node.AsMethodDeclaration().Name()
	case ast.KindMethodSignature:
		return node.AsMethodSignatureDeclaration().Name()
	case ast.KindFunctionDeclaration:
		return node.AsFunctionDeclaration().Name()
	case ast.KindClassDeclaration:
		return node.AsClassDeclaration().Name()
	case ast.KindImportSpecifier:
		return node.AsImportSpecifier().Name()
	}
	return nil
}
