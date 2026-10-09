package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/reference"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoUnassignedVars = rule.Message{
	Id: "noUnassignedVars",
	Description: "This declares a variable without a value and nothing ever assigns one, so every " +
		"read of it is `undefined`. The code below it is written as though the value arrives from " +
		"somewhere, which is what makes this hard to see in review: the comparison, the property " +
		"access, the fallback all read as ordinary until you notice nothing on the other side ever " +
		"writes. The usual cause is an assignment that was deleted, moved into a branch that no " +
		"longer runs, or never written. Give it an initializer, assign it where the value is " +
		"known, or delete the declaration.",
}

// NoUnassignedVars flags a `let` or `var` that is read but never assigned.
//
//	valid:   let x;
//	valid:   let y = undefined; log(y);
//	valid:   let one; if (one !== two) one = two;
//	valid:   declare let c: string | undefined; log(c);
//	valid:   for (let p of paths) { p.remove() }
//	invalid: let user; greet(user);
//	invalid: let flag; while (!flag) { }
//	invalid: let x: number; log(x);
//
// The variable is always `undefined`, so every read of it is a read of `undefined`, and the code
// around it is written as though a value arrives. `while (!flag)` loops forever, `config?.enabled`
// is always `undefined`, and `error || 'Unknown error'` always takes the fallback. None of those
// fail at the declaration, which is why the defect survives: the line that is wrong looks fine and
// the line that misbehaves looks fine too.
//
// A declaration nobody reads is deliberately not reported. That is `no-unused-vars`' finding, and
// duplicating it here would put two findings on every unused local in the tree. Upstream states
// this in a comment and this reproduces it.
//
// # Why this reads the checker, and why it is the conservative direction
//
// The sibling rules in this family anchor on a declaration and report a write to it. This one
// inverts the question: it reports a declaration precisely because no write to it exists. That
// flips which mistake is expensive. There, a write the rule fails to recognize is a missed finding
// on code that is already wrong. Here, a write the rule fails to recognize is a finding on code
// that is correct, accusing a variable of never being assigned three characters away from the line
// that assigns it. So the rule reports only when it can see the whole picture, and declines
// wherever it cannot.
//
// Name matching would fail in that expensive direction. `let x; function f(x) { x = 1; } log(x);`
// has a write spelled exactly like the binding, to a parameter that shadows it, and the outer `x`
// really is never assigned. But the mistake name matching makes there is the cheap one: it sees a
// same-named write and stays silent. The costly direction is the other one, and it is the reason
// for the checker:
//
//	let x; log(x); { let x; x = 1; log(x); }
//
// Two `let` bindings, one write. Name matching sees a write to "x" and declines both, missing the
// outer finding. Symbol identity separates them and reports the outer one.
//
// Node identity rather than declaration kind, which is the trap this family keeps setting. The two
// anchors above are both `let` declarations, so a comparison on kind calls the inner declaration a
// match for the outer one, concludes the outer is assigned, and reports the inner instead. That
// produces exactly one finding, same as the correct behavior, so a fixture asserting the count
// passes while the rule names the wrong variable. `TestNoUnassignedVarsNamesTheOuterBindingWhenOnlyTheInnerIsWritten`
// asserts the offset for that reason.
//
// # Why the checker is not sufficient on its own
//
// Symbol identity says which binding an identifier names; it says nothing about whether the
// occurrence writes. `x.y = 1` and `foo(x)` both resolve to the binding and neither assigns it, so a
// rule that treated every resolved occurrence as a write would go silent on most real findings.
// `reference.WritesToBinding` is the structural half, shared with the sibling rules, and the rule
// concludes "never assigned" only where no occurrence satisfies both halves.
//
// # What is exempt, and why each one
//
//	an initializer          `let x = undefined;` assigns, even to undefined
//	const and using         cannot be reassigned, so a missing value is a different mistake
//	a binding pattern       no single name to report, and `Name()` panics on one outright
//	a for-head declaration  the loop assigns on every iteration
//	an ambient declaration  `declare let c: string;` promises a value from elsewhere
//	any module declaration  reproduced from oxc, which exempts a plain `namespace` too
//
// The last one is a stated divergence from what our checker alone would give. `NodeFlagsAmbient`
// covers `declare let`, `declare module` and `declare global` in a single flag, but a plain
// `namespace N { let x: string; log(x); }` carries no ambient flag while oxc exempts it anyway: its
// guard walks ancestors for `TSNamespaceDeclaration` without asking whether `declare` is present.
// Reproduced rather than improved on, so the two implementations agree.
//
// No fix. The repair is either an initializer whose value the rule cannot know, or a deletion that
// is only correct if every read was also a mistake.
var NoUnassignedVars = rule.Rule{
	Name: "no-unassigned-vars",

	// See the doc above: the discrimination is which binding a write names, and the shape that
	// needs it is a shadowed redeclaration of the same kind.
	NeedsTypeChecker: true,
	// Reads only this file's declarations (rule.DeclarationsIn), so its findings key on imports' shapes.
	TypeReach: rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindVariableDeclaration: func(node *ast.Node) {
				// The engine hands every rule a nil checker when the program could not be built,
				// and this rule can answer nothing without one. Reporting on a nil checker would be
				// the worst available behavior here, since it would mean reporting every readable
				// declaration in the file as never-assigned.
				if ctx.TypeChecker == nil {
					return
				}

				declaration := node.AsVariableDeclaration()

				// An initializer assigns, including `= undefined`. Upstream exits here first and so
				// does this, because it is the cheapest of the exemptions and the most common.
				if declaration.Initializer != nil {
					return
				}

				// A binding pattern has no single name to report, and this guard is also what keeps
				// the rule from crashing: `Name()` on a pattern panics rather than returning nil.
				// A pattern without an initializer is not valid TypeScript, but the linter reads
				// files mid-edit and a parse error does not stop the walk.
				name := node.Name()
				if name == nil || name.Kind != ast.KindIdentifier {
					return
				}

				// Combined flags rather than the node's own, since `const` and `declare` live on the
				// declaration list and the ambient flag is inherited from an enclosing `declare`
				// block. `IsVarConstLike` covers `using` and `await using` alongside `const`, all of
				// which refuse reassignment and so make "never assigned" a different complaint.
				if ast.IsVarConstLike(node) ||
					ast.GetCombinedNodeFlags(node)&ast.NodeFlagsAmbient != 0 {
					return
				}

				// A catch parameter is bound by the runtime when the exception is thrown, so no
				// assignment to it exists anywhere in the source. That is exactly the shape this
				// rule reports, which is why it read as 445 findings on our own tree and every one
				// of them was wrong.
				//
				// Upstream never faces this because it requires two nodes: a `VariableDeclarator`
				// whose parent is a `VariableDeclaration` (oxc `no_unassigned_vars.rs:59` and `:66`).
				// A catch parameter fails that parent check. TypeScript's AST spells the declarator
				// itself `KindVariableDeclaration`, so the port collapsed the two into one anchor
				// and the parent check went with it. It was not redundant; it was this exemption.
				if node.Parent != nil && node.Parent.Kind == ast.KindCatchClause {
					return
				}

				// A for-head declaration is assigned by the loop on every iteration. All three heads
				// count: `for (let i; ;)` is upstream's `ForStatement` arm and is the one that looks
				// most like an ordinary unassigned declaration.
				if list := node.Parent; list != nil && list.Parent != nil {
					switch list.Parent.Kind {
					case ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement:
						return
					}
				}

				if isInsideModuleDeclaration(node) {
					return
				}

				anchor := declarationAnchoredAt(ctx, name)
				if anchor == nil {
					return
				}

				sourceFile := ast.GetSourceFileOfNode(node)
				if sourceFile == nil {
					return
				}

				// The whole file, not the enclosing scope. A write can sit above the declaration
				// inside a hoisted function, below it, or nested arbitrarily deep, and all three
				// name the same binding. Anchoring the search on the file and the match on the
				// symbol is what makes the position of the write irrelevant, which is the same
				// reasoning the sibling rules use for the same walk.
				hasRead := false
				hasWrite := false

				var visit func(*ast.Node)
				visit = func(current *ast.Node) {
					if current == nil || hasWrite {
						return
					}

					// The text comparison is a pre-filter rather than a discrimination: symbol
					// identity already implies it, since an identifier spelled differently cannot
					// resolve to this declaration. It is here because it is far cheaper than a
					// checker call and this walk visits every identifier in the file.
					if current.Kind == ast.KindIdentifier && current.Text() == name.Text() &&
						current != name && resolvesToDeclaration(ctx, current, anchor) {
						if reference.WritesToBinding(current) {
							// One write is enough to settle the question, and finding it stops the
							// walk. Upstream returns on its first `is_write()` for the same reason.
							hasWrite = true
							return
						}
						hasRead = true
					}

					current.ForEachChild(func(child *ast.Node) bool {
						visit(child)
						return false
					})
				}
				visit(sourceFile.AsNode())

				if hasWrite || !hasRead {
					return
				}

				ctx.ReportNode(name, messageNoUnassignedVars)
			},
		}
	},
}

// isInsideModuleDeclaration reports whether a node sits inside a `namespace`, `module` or `global`
// block.
//
// Separate from the ambient-flag check above because the two answer different questions and only
// overlap. `declare module 'm' { let x: string; }` sets the ambient flag on the declaration, so the
// flag check already covers it. A plain `namespace N { let x: string; log(x); }` does not set the
// flag, and oxc exempts it anyway: its guard tests the ancestor's AST kind and never asks whether
// `declare` is present. Both checks are therefore kept, and this one is what makes the plain
// namespace clean.
//
// Whether that exemption is right is a separate question from whether we reproduce it. A binding
// declared in a plain namespace and read there is as much a finding as one at file scope, and
// upstream's guard reads more like an over-broad reach for the ambient case than a decision. It is
// reproduced rather than narrowed so the two implementations agree on the same inputs, and stated
// here so a later reader knows this is a copied gap rather than an oversight of ours.
func isInsideModuleDeclaration(node *ast.Node) bool {
	for ancestor := node.Parent; ancestor != nil; ancestor = ancestor.Parent {
		if ancestor.Kind == ast.KindModuleDeclaration {
			return true
		}
	}
	return false
}
