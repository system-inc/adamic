package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageRedundantBlock = rule.Message{
	Id: "redundantBlock",
	Description: "These braces do nothing. A block standing where statements already go creates no " +
		"scope unless something inside it is block-scoped, so the braces only indent the code and " +
		"suggest a grouping the language does not enforce. A reader spends attention deciding " +
		"whether the nesting means something. Delete the braces and keep the statements, or if " +
		"the grouping was the point, move it into a function whose name says what it groups.",
}

var messageRedundantNestedBlock = rule.Message{
	Id: "redundantNestedBlock",
	Description: "These braces do nothing, and they sit directly inside another block, so the " +
		"nesting reads as two scopes where the language has one. Nothing inside is block-scoped, " +
		"which means the inner braces neither contain a binding nor limit one. Delete them and " +
		"keep the statements at the level they already run at.",
}

// NoLoneBlocks flags a block whose braces create no scope.
//
//	valid:   if (foo) { if (bar) { baz(); } }
//	valid:   { let x = 1; }
//	valid:   { class Bar {} }
//	valid:   { function bar() {} }                     (strict, which every module is)
//	valid:   switch (foo) { case bar: { baz; } }
//	valid:   class C { static { { let block; } something; } }
//	invalid: {}
//	invalid: {var x = 1;}
//	invalid: foo(); {} bar();
//	invalid: if (foo) { bar(); {} baz(); }              redundantNestedBlock
//	invalid: { {let y = 1;} }                           two findings
//
// A block standing where a statement list already goes creates no scope of its own, so the braces
// are decoration that reads as structure. The exception is what the rule is really about: a block
// containing a block-scoped binding is doing real work, and `var` is the case that catches people
// out, because it looks identical and is function-scoped.
//
// # Which message, and it is decided by the PARENT
//
// A block whose parent is another block or a static block reports `redundantNestedBlock`; anything
// else reports `redundantBlock`. Both messages are in upstream's `meta.messages` and the corpus
// asserts which fires per case, so a port with one message passes no fixture here.
//
// # Two ways to be redundant, and the second is not obvious
//
// Upstream's exit handler has two arms and they answer different questions:
//
//	the block is lone AND holds no block-scoped binding      redundant on its own
//	the block is the ONLY statement of an enclosing block     redundant regardless of its contents
//
// The second is what makes `{ {let y = 1;} }` report TWICE. The inner block holds a `let` so it is
// not redundant by the first test, but it is its parent's only statement, so the outer braces add
// nothing and the inner ones are reported for being the sole occupant. Measured against the
// installed eslint 10.8.1 build, which returns `redundantBlock` then `redundantNestedBlock`.
//
// A port implementing only the first arm passes 26 of the 27 imported failing cases.
//
// # What counts as a block-scoped binding, and the one that reads as a version question
//
// `let`, `const`, `using`, `await using`, a class declaration, and a function declaration all bind
// to the block. `var` does not.
//
// The function declaration is the subtle one. Upstream gates it on `sourceCode.getScope(node).isStrict`,
// because a function declaration in a block binds to the block in strict mode and hoists out of it
// in sloppy mode. Upstream's corpus runs `{ function bar() {} }` twice at the SAME ecmaVersion and
// gets opposite verdicts, differing only in `impliedStrict`.
//
// cohere lints TypeScript, and a TypeScript file is a module, and a module body is always strict.
// So the sloppy arm has no reachable configuration and the function declaration always binds.
// Measured rather than reasoned: driving eslint 10.8.1 on that input gives zero findings under
// `sourceType: "module"` and one under `"script"`. Reproducing upstream's scope query would be
// reproducing a workaround for a constraint we do not have.
//
// # A switch case is a statement list, and the exemption is positional
//
// A block inside a `case` is lone unless it is the case's ONLY statement. `case bar: { baz; }` is
// clean and `case 1: foo(); { bar; }` reports, and so does `case 1: { bar; } foo();`. The braces
// are how you give a case its own scope, which only reads as deliberate when they wrap the whole
// case.
//
// # Why this is one source-file listener rather than a block listener
//
// The judgment needs a block's whole subtree before it can answer, since a `let` written anywhere
// directly inside it makes the braces load-bearing. Upstream gets that from `BlockStatement:exit`.
// There is no `rule.OnExit` here and the walk is pre-order, so the walk happens inside a
// `KindSourceFile` listener, which fires before its children.
//
// No fix. Removing the braces is safe only when nothing inside them is block-scoped, which is the
// rule's own predicate, but the repair still has to reindent the body and decide what to do with
// comments attached to the braces. Upstream ships no fixer either.
var NoLoneBlocks = rule.Rule{
	Name: "no-lone-blocks",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				type finding struct {
					node    *ast.Node
					message rule.Message
				}
				var findings []finding

				report := func(block *ast.Node) {
					message := messageRedundantBlock
					// Upstream's test is `parent is BlockStatement or StaticBlock`. In our tree
					// a static block's statements live in a `KindBlock` body hanging off the
					// `KindClassStaticBlockDeclaration`, so a block directly inside a static block
					// already has a `KindBlock` parent and this one test answers both of
					// upstream's. Measured: every `class C { static { { ... } } }` case in the
					// corpus asserts `redundantNestedBlock` and gets it.
					if parent := block.Parent; parent != nil && parent.Kind == ast.KindBlock {
						message = messageRedundantNestedBlock
					}
					findings = append(findings, finding{block, message})
				}

				var visit func(*ast.Node) bool
				visit = func(current *ast.Node) bool {
					// No static-block-body guard, and that is measured. A static block's own
					// body is a `KindBlock` whose parent is the
					// `KindClassStaticBlockDeclaration`, which is neither of the two parent kinds
					// the arms below accept, so both decline it without a guard in front. An
					// explicit guard was written first, and a mutant deleting it survived every
					// fixture; deleting it AND widening `isLoneBlock` to accept a static block
					// parent is caught by 29 lines, which is what showed the guard was a second
					// spelling of a decision the helpers already make rather than a filter of its
					// own.
					//
					// This still matters to read, because in ESTree the static block IS the
					// statement list and this position does not exist. Every arm below therefore
					// has to keep declining a `KindClassStaticBlockDeclaration` parent, and the
					// corpus's eight clean static-block cases are what will notice if one stops.
					if current.Kind == ast.KindBlock {
						switch {
						case isLoneBlock(current) && !holdsBlockScopedBinding(current):
							report(current)
						case isOnlyStatementOfABlock(current):
							// The second arm, and the reason `{ {let y = 1;} }` reports twice: a
							// block that IS its parent's whole contents adds nothing, whatever it
							// holds. Upstream reaches this through the `else if` after its lone
							// check, so a block satisfying both reports once rather than twice.
							report(current)
						}
					}
					current.ForEachChild(visit)
					return false
				}
				node.ForEachChild(visit)

				for _, found := range findings {
					ctx.ReportNode(found.node, found.message)
				}
			},
		}
	},
}

// isLoneBlock reports whether a block stands where a statement list already goes.
//
// Upstream's `isLoneBlock`: the parent is a block, a static block, the program, or a switch case
// where the block is not the case's only statement. Everything else -- an `if` consequent, a loop
// body, a function body, a `try` block, a labeled statement -- is a block the syntax required.
func isLoneBlock(block *ast.Node) bool {
	parent := block.Parent
	if parent == nil {
		return false
	}

	switch parent.Kind {
	case ast.KindBlock, ast.KindSourceFile:
		return true

	case ast.KindCaseClause, ast.KindDefaultClause:
		// A block wrapping the WHOLE case is how a case is given its own scope, and upstream
		// exempts exactly that. `case bar: { baz; }` is clean; `case 1: foo(); { bar; }` and
		// `case 1: { bar; } foo();` both report, so the test is on the case's length as well as on
		// which statement the block is.
		// No nil guard on the list: probed with `switch(a){ case 1: }`, `case 1: case 2: b;`,
		// `default: }` and a truncated `case }`, and the parser gives an EMPTY list every time,
		// never nil. A mutant flipping the guard's verdict survived every fixture, which is what
		// sent us to probe rather than to write a case for it.
		statements := parent.AsCaseOrDefaultClause().Statements

		// The identity comparison is equivalent to the length test alone and is kept for what it
		// SAYS rather than for what it decides. A block reached here is by construction one of its
		// parent clause's statements, so a clause with exactly one statement has that statement be
		// this block. Checked exhaustively over sixteen nesting shapes with a probe asserting the
		// invariant directly: zero violations. Two mutants deleting the comparison, one per arm,
		// survive for that reason and are equivalent rather than uncovered.
		//
		// The first hypothesis about them was wrong and is worth recording: `{ if (a) { bar; } }`
		// looks like it should distinguish them, since `Nodes[0]` is the `IfStatement` rather than
		// the consequent block. It cannot, because the consequent block's parent is the
		// `IfStatement` and the kind guard above declines it before any of this runs.
		onlyStatement := len(statements.Nodes) == 1 && statements.Nodes[0] == block
		return !onlyStatement
	}

	return false
}

// isOnlyStatementOfABlock reports whether a block is the entire contents of an enclosing block.
//
// Upstream's second arm. Separate from `isLoneBlock` because it answers a different question and
// reaches a different set: a block holding a `let` is not redundant on its own, and is still
// redundant when it is all its parent contains. A static block counts as the parent, which is what
// upstream's `node.parent.type === "StaticBlock"` adds.
func isOnlyStatementOfABlock(block *ast.Node) bool {
	parent := block.Parent
	if parent == nil {
		return false
	}

	if parent.Kind != ast.KindBlock {
		return false
	}

	// The identity comparison here is equivalent to the length test for the same reason as in
	// `isLoneBlock`, and is kept for the same reason: it states the invariant the arm relies on.
	// See the note there for the probe and for the hypothesis that was wrong about it.
	statements := parent.AsBlock().Statements
	return statements != nil && len(statements.Nodes) == 1 && statements.Nodes[0] == block
}

// holdsBlockScopedBinding reports whether anything DIRECTLY inside a block binds to it.
//
// Directly, not recursively: a `let` inside a nested block belongs to that block, and upstream
// answers the same way because its marker only fires for a declaration whose immediate parent is
// the block being judged.
//
// A function declaration counts unconditionally here, where upstream gates it on strictness. That
// is the one place this deliberately does not reproduce upstream's mechanism, because a TypeScript
// file is a module and a module body is always strict, so upstream's sloppy arm has no reachable
// configuration. See the rule doc for the measurement.
func holdsBlockScopedBinding(block *ast.Node) bool {
	statements := block.AsBlock().Statements
	if statements == nil {
		return false
	}

	for _, statement := range statements.Nodes {
		switch statement.Kind {
		case ast.KindClassDeclaration, ast.KindFunctionDeclaration:
			return true

		// TypeScript's own block-scoped declarations, which upstream has no arm for because its
		// parser cannot produce them: eslint 10.8.1 reports a fatal parse error on all four.
		//
		// This is a false-positive class no imported fixture can see, and it is not a judgment
		// call. Measured with tsc: a file writing `{ enum E {} interface I {} type T = number; }`
		// and then referencing `E`, `I` and `T` below the block gets TS2304 "Cannot find name" for
		// every one of them, so all three really do bind to the block exactly as `let` does.
		// `namespace` is TS1235 inside a block, which is not legal at all rather than not scoped,
		// and it is listed so a recovered parse does not produce a finding on already-broken code.
		//
		// Without this, `{ type T = number; }` reports as a redundant block while the braces are
		// the only thing keeping `T` local.
		case ast.KindEnumDeclaration, ast.KindInterfaceDeclaration,
			ast.KindTypeAliasDeclaration, ast.KindModuleDeclaration:
			return true

		case ast.KindVariableStatement:
			// `var` is function-scoped and does not make the braces load-bearing, which is the
			// distinction the rule exists for: `{var x = 1;}` reports and `{ let x = 1; }` does
			// not, and they are one keyword apart. `using` and `await using` bind to the block the
			// way `let` does.
			declarationList := statement.AsVariableStatement().DeclarationList
			if declarationList == nil {
				continue
			}
			// `NodeFlagsBlockScoped` is `Let | Const | Using` at
			// `TypeScript/tsc/internal/ast/nodeflags.go:49`, so it already covers `using` and
			// `await using` and a separate flag test for them would be dead. Checked rather than
			// assumed, because the name reads narrower than the constant is.
			if declarationList.Flags&ast.NodeFlagsBlockScoped != 0 {
				return true
			}
		}
	}

	return false
}
