package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageUnexpectedDebugger = rule.Message{
	Id: "unexpectedDebugger",
	Description: "A debugger statement stops execution when devtools are open and does nothing " +
		"otherwise, so it is a breakpoint written into the source. Shipped, it halts the page for " +
		"anyone with devtools open and is invisible to everyone else, which is the worst pairing " +
		"of symptoms for reproducing a report. Set the breakpoint in the debugger instead.",
}

// NoDebugger flags a debugger statement.
//
//	valid:   console.log(value)
//	invalid: debugger;
//	invalid: if(condition) debugger;
//
// This is the cheapest rule in the catalog and one of the most valuable, because the failure it
// prevents is asymmetric: a stray debugger costs the author nothing (their devtools are open, the
// pause is expected) and costs a user with devtools open a frozen page. The author is the least
// likely person to notice.
//
// The fix removes the statement rather than commenting it out. There is no form of this statement
// that belongs in committed code, so there is nothing to preserve. ESLint offers no fix at all; this
// one is ours, and it is offered only where the statement sits in a list of statements. As the whole
// body of an unbraced `if`, a loop or a label, removing it leaves `if (foo)` with no statement, which
// does not parse (ESLint's own corpus row, caught by the harness's fix-parses check, #kq9vtva), so
// there the finding stands without a repair.
var NoDebugger = rule.Rule{
	Name: "no-debugger",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindDebuggerStatement: func(node *ast.Node) {
				if !noDebuggerSitsInAStatementList(node) {
					ctx.ReportNode(node, messageUnexpectedDebugger)
					return
				}
				ctx.ReportNodeWithFixes(node, messageUnexpectedDebugger, ctx.RemoveNode(node))
			},
		}
	},
}

// noDebuggerSitsInAStatementList reports whether removing the statement leaves its parent whole: a
// block, a file, a module body or a case clause holds a list, and losing one entry of it is fine.
func noDebuggerSitsInAStatementList(node *ast.Node) bool {
	if node.Parent == nil {
		return false
	}
	switch node.Parent.Kind {
	case ast.KindBlock, ast.KindSourceFile, ast.KindModuleBlock, ast.KindCaseClause, ast.KindDefaultClause:
		return true
	}
	return false
}
