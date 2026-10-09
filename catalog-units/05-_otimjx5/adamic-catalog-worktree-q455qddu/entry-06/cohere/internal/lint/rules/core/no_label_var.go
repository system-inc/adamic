package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoLabelVar = rule.Message{
	Id: "identifierClashWithLabel",
	Description: "This label has the same name as a variable in scope, so the same word means two " +
		"different things a few lines apart and neither reading is wrong on sight. A reader who " +
		"sees `break x` has to know whether `x` is the label or the variable before they can say " +
		"what the statement does. Rename the label.",
}

// NoLabelVar flags a label whose name is already a variable in scope.
//
//	valid:   x: for(;;) { break x; }
//	valid:   function bar() { var x = foo; q: for(;;) { break q; } }
//	valid:   { let x = 1; } x: for(;;) { break x; }
//	invalid: function bar(x) { x: for(;;) { break x; } }
//	invalid: let x = 1; x: for(;;) { break x; }
//	invalid: Object: for(;;) { break Object; }
//
// Labels and variables live in different namespaces, so this collision is legal and the program
// runs. It is a readability judgment: `break x` reads as either one until you have found which.
//
// # It is a scope question, which is the whole port
//
// Upstream asks `eslint-scope` for the innermost scope at the labeled statement and then for a
// variable of that name, which walks the chain outward to the global scope. `GetSymbolsInScope` asks
// the same question of the checker, and the meaning it is asked for is what has to match.
//
// `SymbolFlagsValue` is that meaning, measured rather than chosen. Upstream's `getVariableByName`
// returns ANY scope variable, so a function declaration, a class, a `let` and a `const` all clash,
// not only a `var`. Probed against sixteen inputs on the installed build at 10.8.1, the two agree on
// every one, including the four `SymbolFlagsVariable` gets wrong.
//
// # Two upstream behaviours that read as false positives, reproduced
//
// `Object: for(;;) {}` and `undefined: for(;;) {}` both REPORT, measured. The scope chain ends at
// the global scope, where every standard library name is a variable, so any label named after a
// builtin clashes. That is upstream's actual behaviour rather than an oversight in this port, and
// widening or narrowing it here would be a different rule.
//
// # And two that read as gaps and are the point
//
// A binding whose block has already closed does not clash: `{ let x = 1; } x: for(;;) {}` is clean.
// Nor does one inside the label's own body: `x: for(;;) { let x = 1; }` is clean, because the label
// sits outside the block it labels. Both measured, and both would report for any rule that searched
// the file for the name instead of asking what is in scope at this point.
var NoLabelVar = rule.Rule{
	Name:             "no-label-var",
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindLabeledStatement: func(node *ast.Node) {
				// Without this the rule goes completely silent under the plain harness rather than
				// crashing, which is the more dangerous failure: every clean fixture would pass
				// vacuously. `TestNoLabelVarRequiresTheTypedHarness` pins it.
				if ctx.TypeChecker == nil {
					return
				}

				labeled := node.AsLabeledStatement()
				if labeled.Label == nil {
					return
				}
				labelName := labeled.Label.Text()
				if labelName == "" {
					return
				}

				// Asked at the labeled statement rather than inside it, which is upstream's own
				// anchor and is what makes a binding declared in the label's body not clash.
				for _, symbol := range ctx.TypeChecker.GetSymbolsInScope(node, ast.SymbolFlagsValue) {
					if symbol.Name == labelName {
						ctx.ReportNode(node, messageNoLabelVar)
						return
					}
				}
			},
		}
	},
}
