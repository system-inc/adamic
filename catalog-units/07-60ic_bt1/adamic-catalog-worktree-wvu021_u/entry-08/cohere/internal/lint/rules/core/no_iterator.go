package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/ecmascript/property"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoIterator = rule.Message{
	Id: "noIterator",
	Description: "This uses the reserved name `__iterator__`. That was SpiderMonkey's " +
		"pre-standard iteration protocol, removed from Firefox in 2015 and never implemented " +
		"anywhere else, so the property does nothing in any engine running today. Use " +
		"`[Symbol.iterator]`, which is the standard protocol that replaced it.",
}

var messageUseSymbolIterator = rule.Message{
	Id:          "useSymbolIterator",
	Description: "Replace `__iterator__` with `[Symbol.iterator]`.",
}

// NoIterator flags a property access naming `__iterator__`.
//
//	valid:   var a = test[__iterator__];
//	valid:   var __iterator__ = null;
//	valid:   foo[`__iterator`] = null;
//	invalid: var a = test.__iterator__;
//	invalid: var a = test['__iterator__'];
//	invalid: var a = test[`__iterator__`];
//
// The three valid cases are the whole discrimination and each fails a different way if the rule is
// written as a text match. `test[__iterator__]` is a computed access through a *variable* of that
// name, which is a different property entirely and may not be a string at all. A declaration binding
// the name is not a property access. And a template literal whose text merely resembles it,
// including one carrying a newline, is a different key.
//
// # Static property name is the predicate, not the spelling
//
// Upstream asks the member expression for its static property name, which answers for the dotted
// form, a string-literal subscript, and a template literal with no substitutions alike, and answers
// nothing for a computed access whose key is an expression. That is the same question this asks, and
// writing it as "does the source text contain __iterator__" would report all three valid cases.
//
// A template literal with substitutions has no static name even when its cooked text would match,
// because the value is not known until it runs.
//
// # The suggestion, and why it is not a fix
//
// Replacing `.__iterator__` with `[Symbol.iterator]` preserves meaning only if the author meant the
// iteration protocol, and that is an inference about intent rather than a mechanical equivalence: a
// codebase that defined its own `__iterator__` property gets a different property under the same
// name. So it is offered as a suggestion a human chooses rather than applied unattended.
//
// The replaced span runs from the end of the object expression to the end of the member expression,
// which is what deletes the dot in `test.__iterator__` and the brackets in `test['__iterator__']`
// without either case needing its own arm.
var NoIterator = rule.Rule{
	Name: "no-iterator",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			// Both access kinds, because `staticPropertyName` answers for both and a rule watching
			// only the dotted form is silent on `test['__iterator__']`, which upstream reports.
			ast.KindPropertyAccessExpression: reportIfIteratorAccess(ctx),
			ast.KindElementAccessExpression:  reportIfIteratorAccess(ctx),
		}
	},
}

// reportIfIteratorAccess reports an access whose statically-known property is `__iterator__`.
//
// The predicate is `staticPropertyName`, which this package already had for `no-self-assign` and
// which answers exactly upstream's question: the dotted form, a string-literal subscript, and a
// no-substitution template all resolve, while a computed access through a variable does not. That is
// what makes `test[__iterator__]` a pass case rather than a missed finding, and writing a second
// spelling of it here would have been the drift the shelf guard exists to refuse.
//
// It is broader than this rule needs in one way, accepting a numeric-literal subscript, which is
// harmless: no number spells `__iterator__`.
func reportIfIteratorAccess(ctx rule.Context) func(*ast.Node) {
	return func(node *ast.Node) {
		// The second return is not read, and that is measured rather than sloppy: it is false only
		// when the name is "", which the comparison below already rejects. A mutation dropping the
		// check survived the whole fixture set, and reading `staticPropertyName` shows why. Kept out
		// rather than kept in, because an unreachable guard reads as a case somebody handled.
		name, _ := property.AccessedName(node, property.Static)
		if name != "__iterator__" {
			return
		}
		object := accessedObject(node)
		if object == nil {
			return
		}
		ctx.ReportNodeWithSuggestions(node, messageNoIterator, rule.Suggestion{
			Message: messageUseSymbolIterator,
			Fixes: []rule.Fix{
				// The span runs from the end of the object to the end of the access, which removes
				// the dot in `test.__iterator__` and the brackets in `test['__iterator__']` without
				// either spelling needing its own arm.
				//
				// `rule.ReplaceRange` rather than `ctx.ReplaceRange`: the fix families split by
				// receiver on purpose. Node forms are Context methods because they need the context
				// to trim trivia; range forms are package-level because the caller already computed
				// the exact span and nothing should be trimmed. Reaching for `ctx.` here is the
				// reasonable instinct and is wrong for exactly these two.
				rule.ReplaceRange(core.NewTextRange(object.End(), node.End()), "[Symbol.iterator]"),
			},
		})
	}
}
