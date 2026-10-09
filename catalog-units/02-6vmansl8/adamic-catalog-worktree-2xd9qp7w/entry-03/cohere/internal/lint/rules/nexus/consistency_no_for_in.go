package nexus

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

const consistencyNoForInId = "forIn"

// consistencyNoForInText is the rule's message, whose wording lives in
// `policy/messages/consistency-no-for-in.json`.
var consistencyNoForInText = policy.MessageOf("nexus/consistency-no-for-in", consistencyNoForInId)

func consistencyNoForInMessage() rule.Message {
	return rule.Message{
		Id:          consistencyNoForInId,
		Description: consistencyNoForInText.Render(nil),
	}
}

// ConsistencyNoForIn bans the `for...in` statement.
//
//	invalid: for (const key in object) { use(object[key]); }
//	invalid: for (const key in object) { if (!Object.hasOwn(object, key)) continue; use(object[key]); }
//	valid:   for (const [key, value] of Object.entries(object)) { use(key, value); }
//	valid:   for (const key of Object.keys(object)) { use(key); }
//	valid:   if ('name' in object) { use(object.name); }
//
// # Why this replaces `guard-for-in`
//
// Upstream's `guard-for-in` accepts a `for...in` whose body opens with an `if`, so it asks every loop
// over an object to carry a guard, and it cannot tell a guard from any other `if` (a body opening
// with `if (key === 'parent') continue;` passes while still visiting inherited keys). Removing the
// statement removes the question: `Object.entries` and `Object.keys` return own enumerable string
// keys, which is what a guarded `for...in` computes, so the replacement is the guarded loop's
// meaning with the guard built in. Kirk's ruling, 2026-10-01: one loop shape for objects.
//
// # What it reports
//
// Every `for...in` statement, guarded or not, in every file. There is no condition to satisfy and no
// option: a guarded loop is the same shape the rule exists to retire, and an `Object.hasOwn` guard in
// front of it is exactly the ceremony the replacement deletes. The `in` operator
// (`'name' in object`) is a different construct, a membership test, and is not reported; neither is
// `for...of` or `for await...of`, which share a node shape with `for...in` in the parser and are told
// apart by kind.
//
// # No fix
//
// The rewrite is mechanical in the common case and not meaning-preserving in general. A `for...in`
// over a class instance or an object with a prototype that carries enumerable properties visits keys
// `Object.keys` does not, an unguarded loop that mutates the object while iterating sees different
// keys than a snapshot does, and choosing between `Object.keys` and `Object.entries` (and how to
// rename the value reads in the body) is the author's call. So the finding says what to write and
// the author writes it.
var ConsistencyNoForIn = rule.Rule{
	Name: "nexus/consistency-no-for-in",

	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			// Keyed on the `for...in` kind alone. `for...of` shares `ForInOrOfStatement` in this
			// parser, and a listener on the shared shape would report every `for...of` in the tree.
			ast.KindForInStatement: func(node *ast.Node) {
				ctx.ReportNode(node, consistencyNoForInMessage())
			},
		}
	},
}
