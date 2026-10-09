package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoContinueUnexpected = rule.Message{
	Id: "unexpected",
	Description: "This loop uses `continue`, which jumps to the next iteration from the middle of " +
		"the body. A reader tracing what happens on one pass has to hold every `continue` above " +
		"their position in mind, because any of them may have skipped the code they are looking " +
		"at, and in a `for` loop the update expression still runs while the rest of the body does " +
		"not. Inverting the condition and wrapping the remainder in an `if` says the same thing " +
		"with the skip visible in the shape of the code rather than in a jump.",
}

// NoContinue flags a `continue` statement.
//
//	valid:   for (let i = 0; i < 10; i++) { if (i < 5) { doSomething(); } }
//	valid:   for (let i = 0; i < 10; i++) { break; }
//	invalid: for (let i = 0; i < 10; i++) { if (i >= 5) { continue; } doSomething(); }
//	invalid: outer: for (;;) { for (;;) { continue outer; } }
//
// Ported from `no-continue` in ESLint, read from the clone at `lib/rules/no-continue.js`. No options
// (`schema: []`), one message, no fixer. The rule body is four lines: every `ContinueStatement`
// reports, unconditionally.
//
// The whole 6-case corpus was extracted from upstream's own tester and replayed against the rule
// through the ESLint Linter API before any code was written. It agreed on all 6.
//
// # There is genuinely no discrimination here, and that is worth saying out loud
//
// Most rules in this package decide something. This one does not: it has no options, no exemptions,
// and no shape it declines. A labeled `continue` reports exactly like a bare one, and a `continue`
// in any loop kind reports the same way. The only thing a port can get wrong is failing to see the
// statement at all.
//
// That makes the mutation sweep unusually weak here, which is a property of the rule rather than of
// the fixtures: there is one branch, so there is one thing to mutate. The fixtures compensate by
// covering the shapes the corpus does not write -- a labeled continue, and each loop kind -- so that
// a port which somehow saw only one of them would be caught.
var NoContinue = rule.Rule{
	Name: "no-continue",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindContinueStatement: func(node *ast.Node) {
				ctx.ReportNode(node, messageNoContinueUnexpected)
			},
		}
	},
}
