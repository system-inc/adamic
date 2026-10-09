package base

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// ConsistencyNoConsole flags every `console.*` access in source.
//
//	valid:   myConsole.log('x');
//	valid:   window.console.log('x');
//	valid:   console;
//	invalid: console.log('x');
//	invalid: const log = console.log;
//	invalid: foo(console.error);
//	invalid: console['log']('x');
//
// This is one of Kirk's own base rules rather than an upstream port, so the source in
// `api-phi-health/libraries/base/code-quality/lint/rules/ConsistencyNoConsoleRule.ts` is the specification and
// there is no imported corpus. Its reasoning is worth carrying rather than restating: a log line is
// an ingested artifact, competing for the 256 KB trace budget per invocation on exactly the request
// that failed, and walked by the consumer in every batch. The narrow version of this rule rots on
// contact, because "use the logger for errors, console.log is fine for debugging" collapses the
// first time somebody debugs an error.
//
// # The anchor is a member ACCESS, not a call, and that is the whole design
//
// Upstream registers a `MemberExpression` visitor rather than a `CallExpression` one, with a comment
// saying why: the aliasing routes are the point. `const log = console.log` and passing
// `console.error` as a callback both escape a call-shaped check while reporting themselves clean,
// which is worse than not having the rule.
//
// So this anchors on property and element access rather than on the call, and the corresponding
// spellings all report:
//
//	console.log('x')       reports, method log
//	const log = console.log   reports, method log
//	foo(console.error)     reports, method error
//	console['log']('x')    reports, method log (a literal key takes the fallback)
//
// Measured against the real rule driven over the eslint interface, not inferred from the source.
//
// # What is deliberately silent
//
// The receiver must be the bare identifier `console`. `window.console.log('x')` is silent, because
// the object of the outer access is `window.console` rather than `console`, and `myConsole.log` is
// silent for the obvious reason. A bare `console;` with no member access is silent too, since there
// is no member expression at all. All three measured.
//
// That receiver test is upstream's `node.object.type !== 'Identifier' || node.object.name !==
// 'console'`, reproduced exactly. It means a shadowed local named `console` still reports, which is
// upstream's behaviour: the rule reads a name rather than resolving a binding, so it needs no type
// checker and cannot be defeated by a local that happens to be called console.
//
// # The method name, and the fallback that is easy to miss
//
// The message interpolates the accessed member. Upstream reads it as
// `node.property.type === 'Identifier' ? node.property.name : 'log'`, and the subtlety is that
// estree's `property` is whatever expression sits in the brackets rather than a marker of
// computedness. So the fallback fires on the KEY'S KIND, not on the access being computed:
//
//	console.error       renders console.error
//	console[key]        renders console.key       an Identifier property
//	console['error']    renders console.log       a Literal property, so the fallback
//
// That middle row is the one that reads wrong from the source and is right in fact. It was written
// the other way here first, and only the fixture measured against the real rule caught it. A port
// that helpfully read the string literal would diverge on a message nobody would think to check.
//
// # No repair
//
// Upstream ships no fixer and says why: replacing a console call means naming what failed, which is
// the judgment the whole design is asking for and the one thing a rule cannot do.
//
// # Cost
//
// Two common anchors, both exiting on a single identifier comparison. No checker, no program.
var ConsistencyNoConsole = rule.Rule{
	Name: "base/consistency-no-console",

	Run: func(ctx rule.Context, options any) rule.Listeners {
		report := func(node *ast.Node, receiver *ast.Node, methodName string) {
			// Parentheses are stepped through before the receiver is read, and this line was
			// missing on the first writing.
			//
			// estree has no parenthesized-expression node, so upstream's `node.object` for
			// `(console).log('x')` IS the identifier and it reports. Our parser keeps
			// `KindParenthesizedExpression`, so a receiver test without this unwrap goes silent on
			// that shape. Measured against the real rule driven over the eslint interface; it was
			// the one disagreement in twenty-five probed inputs.
			//
			// Written as a loop with its own nil check rather than `ast.SkipParentheses`, which
			// dereferences its argument, and because `((console))` nests.
			for receiver != nil && receiver.Kind == ast.KindParenthesizedExpression {
				inner := receiver.AsParenthesizedExpression().Expression
				if inner == nil {
					break
				}
				receiver = inner
			}

			// The receiver must be the bare identifier `console`. A name comparison rather than a
			// resolved binding, matching upstream, which is what keeps this rule free of the
			// checker.
			if receiver == nil || receiver.Kind != ast.KindIdentifier || receiver.Text() != "console" {
				return
			}
			ctx.ReportNode(node, buildNoConsoleMessage(methodName))
		}

		return rule.Listeners{
			ast.KindPropertyAccessExpression: func(node *ast.Node) {
				access := node.AsPropertyAccessExpression()
				name := access.Name()
				if name == nil {
					return
				}
				// A private name (`console.#x`) is not an Identifier, so it takes upstream's
				// fallback rather than its own text.
				methodName := "log"
				if name.Kind == ast.KindIdentifier {
					methodName = name.Text()
				}
				report(node, access.Expression, methodName)
			},

			ast.KindElementAccessExpression: func(node *ast.Node) {
				access := node.AsElementAccessExpression()

				// The rendered member name for a computed access is NOT always the fallback, and
				// reading upstream's ternary got this backwards on the first writing.
				//
				// estree keeps `property` as whatever expression sits in the brackets, so
				// `console[key]` has an Identifier property named `key` and renders `console.key`,
				// while `console['error']` has a Literal property and takes the fallback, rendering
				// `console.log`. Both measured against the real rule:
				//
				//	console[key]        renders console.key
				//	console['error']    renders console.log
				//	console['log']      renders console.log
				//
				// So the test is on the KEY'S KIND rather than on the access being computed.
				methodName := "log"
				if key := access.ArgumentExpression; key != nil && key.Kind == ast.KindIdentifier {
					methodName = key.Text()
				}
				report(node, access.Expression, methodName)
			},
		}
	},
}

// consistencyNoConsoleText is the rule's message, whose wording lives in
// `policy/messages/consistency-no-console.json`.
var consistencyNoConsoleText = policy.MessageOf("base/consistency-no-console", "consistencyNoConsole")

func buildNoConsoleMessage(methodName string) rule.Message {
	return rule.Message{
		Id:          consistencyNoConsoleText.Id,
		Description: consistencyNoConsoleText.Render(map[string]string{"methodName": methodName}),
	}
}
