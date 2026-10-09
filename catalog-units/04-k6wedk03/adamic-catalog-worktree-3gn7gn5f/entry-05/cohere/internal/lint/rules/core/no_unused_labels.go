package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/ecmascript/comments"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageUnusedLabel = rule.Message{
	Id: "unusedLabel",
	Description: "This label is declared and nothing jumps to it. A label only does anything " +
		"when a `break` or `continue` names it, so an unnamed one is dead syntax, and it is " +
		"usually what is left over when the jump that used it was removed or rewritten.",
}

// NoUnusedLabels flags a label no `break` or `continue` names.
//
//	valid:   A: break A;
//	valid:   A: for (var i = 0; i < 10; ++i) { if (a) continue A; }
//	valid:   outer: { function f() { inner: { break inner; } } break outer; }
//	invalid: A: var foo = 0;
//	invalid: A: for (var i = 0; i < 10; ++i) { if (a) break; }
//	invalid: A: B: 'foo'
//
// # Both jump forms count, and missing one halves the rule
//
// A label is used by a labeled `break` or a labeled `continue`, and the two are not interchangeable:
// `continue A` is legal only against a loop, `break A` against any labeled statement. A port hooking
// only `break` reports every loop whose sole jump is `continue A`, which is a false positive on
// correct code and the most common way this rule is ported wrong. Both listeners are here for that
// reason, and the clean fixture pair covers each separately.
//
// # Scope is lexical, so this needs no checker
//
// A jump reaches only the labels enclosing it, so the question is which labels contain this jump,
// and the AST already answers that. Nothing here asks what declaration a name binds to, which is the
// only question that would need type information. `A: { var A = 0; console.log(A); break A; }` is
// the case that proves the distinction: the variable `A` shadows nothing, because a label and a
// variable live in different namespaces entirely, and a rule resolving `A` as a value would find the
// variable and conclude the label was unused.
//
// The walk goes upward from each jump rather than downward from each label, and stops at the first
// enclosing label whose name matches. Stopping matters: `A: { A: { break A; } }` has two labels of
// one name, the break belongs to the inner one, and marking every match would call the outer label
// used. A function or class boundary stops the walk too, since a label does not cross one.
//
// # The fix, and the three cases it refuses
//
// Removing an unused label means deleting the text from the label through the colon, leaving the
// body. That preserves meaning in most cases and is applied unattended, but not always, and this is
// the one place the port diverges from oxc.
//
// oxc declares this fix unconditionally and rewrites the labeled statement to its body every time.
// ESLint gates the same repair behind three refusals, and each one corresponds to a rewrite that
// parses and means something different. The engine's guard checks that the rewritten file parses,
// which makes it blind to exactly these: a fix producing source that parses and misbehaves is the
// one failure it structurally cannot refuse. So the refusals are ported and the divergence is
// deliberate.
//
//	comment    `A: /* comment */ foo` would discard the comment, and in `A /* comment */: foo`
//	           there is no span that removes the label and keeps it.
//	directive  `A: "use strict"` would promote the string into a directive prologue, changing which
//	           programs are legal rather than only how they read.
//	ASI        `foo()\nLABEL: [1].forEach(x => x)` would rejoin as `foo()[1]`, a different program.
//
// The finding still fires in all three. Only the repair is withheld, which is the right trade: a
// missing fix costs a manual edit, and a wrong one costs a silent behavior change.
// The verdict is about the whole file rather than about one node, since a label may be used by a
// jump written anywhere below it, so nothing can be decided at the label itself. The walk here is
// pre-order and a source-file listener fires before its children, so a listener collecting labels
// and a separate one reading them would see an empty list. Gathering and judging therefore happen in
// the same place, which is what an exit hook would otherwise buy.
var NoUnusedLabels = rule.Rule{
	Name: "no-unused-labels",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				labeledStatements := []*ast.Node{}
				// Labels some jump names, keyed by the labeled statement's own position. Position
				// rather than name, because two labels may share a name and only one is the target.
				usedLabels := make(map[int]bool)

				var visit func(*ast.Node) bool
				visit = func(current *ast.Node) bool {
					switch {
					case ast.IsLabeledStatement(current):
						labeledStatements = append(labeledStatements, current)
					case current.Kind == ast.KindBreakStatement:
						markTargetLabelUsed(current, current.AsBreakStatement().Label, usedLabels)
					case current.Kind == ast.KindContinueStatement:
						markTargetLabelUsed(current, current.AsContinueStatement().Label, usedLabels)
					}
					current.ForEachChild(visit)
					return false
				}
				node.ForEachChild(visit)

				for _, labeled := range labeledStatements {
					if usedLabels[labeled.Pos()] {
						continue
					}
					statement := labeled.AsLabeledStatement()
					if statement == nil || statement.Label == nil {
						continue
					}
					if removal, ok := labelRemovalFix(ctx, labeled, statement); ok {
						ctx.ReportNodeWithFixes(statement.Label, messageUnusedLabel, removal)
						continue
					}
					ctx.ReportNode(statement.Label, messageUnusedLabel)
				}
			},
		}
	},
}

// markTargetLabelUsed records the label a jump names, if it names one.
//
// The walk stops at the first enclosing label of a matching name rather than marking every match,
// because a jump belongs to exactly one label and the nearest one wins. It also stops at a function
// or class boundary, since a jump cannot name a label outside the function holding it: without that
// stop, the inner `label` in `label: while (true) { (() => { label: while (false) {} })(); }` would
// mark the outer one and the rule would report one finding where upstream reports two.
func markTargetLabelUsed(jump *ast.Node, label *ast.Node, usedLabels map[int]bool) {
	if label == nil {
		return
	}
	name := label.Text()
	for ancestor := jump.Parent; ancestor != nil; ancestor = ancestor.Parent {
		if ast.IsFunctionLikeDeclaration(ancestor) || ast.IsClassLike(ancestor) {
			return
		}
		if !ast.IsLabeledStatement(ancestor) {
			continue
		}
		labeled := ancestor.AsLabeledStatement()
		if labeled != nil && labeled.Label != nil && labeled.Label.Text() == name {
			usedLabels[ancestor.Pos()] = true
			return
		}
	}
}

// labelRemovalFix builds the repair that deletes a label, or declines to.
//
// The span runs from the label's own token start to the body's token start, which is what removes
// the name and the colon while leaving the body untouched. Token starts rather than Pos(), because
// Pos() sits before leading trivia: for the inner label of `A: B: 'foo'` the raw Pos() covers the
// space before `B`, and a fix built from it eats whitespace that belongs to nobody.
func labelRemovalFix(ctx rule.Context, labeled *ast.Node, statement *ast.LabeledStatement) (rule.Fix, bool) {
	labelStart := scanner.GetTokenPosOfNode(statement.Label, ctx.SourceFile, false)
	bodyStart := scanner.GetTokenPosOfNode(statement.Statement, ctx.SourceFile, false)

	if commentLiesBetween(ctx, labelStart, bodyStart) {
		return rule.Fix{}, false
	}
	if wouldBecomeADirective(labeled, statement) {
		return rule.Fix{}, false
	}
	if wouldRejoinWithThePreviousLine(ctx, labelStart, bodyStart) {
		return rule.Fix{}, false
	}

	return rule.RemoveRange(core.NewTextRange(labelStart, bodyStart)), true
}

// commentLiesBetween says whether any comment sits in the span the fix would delete.
//
// A comment there is either destroyed by the removal (`A: /* comment */ foo`) or sits where no span
// can remove the label without it (`A /* comment */: foo`), so both decline. ESLint asks this by
// comparing the token after the label with the token before the body and checking they are the same
// token; asking the comment list directly answers the same question over the same span.
//
// The comment cache is shared per file, so this costs one scan across every label in a file rather
// than one per label.
func commentLiesBetween(ctx rule.Context, start int, end int) bool {
	for _, comment := range comments.ForFile(ctx) {
		if comment.Range.Pos() < end && comment.Range.End() > start {
			return true
		}
	}
	return false
}

// wouldBecomeADirective says whether removing the label would promote the body into a directive.
//
// A string expression statement is a directive only in a directive prologue, which exists at the top
// of a Program and at the top of a function body and nowhere else. So the question is what the
// labeled statement's outermost non-label ancestor is: while the label is there the string is an
// ordinary statement, and removing it can put the string somewhere a parser reads as `"use strict"`.
//
// That changes which programs are legal rather than only how they read, which is why it disqualifies
// a fix rather than merely looking untidy. A parenthesized string is not a directive under any
// circumstances, so `A: ("use strict")` is refused for the ASI reason below rather than this one.
func wouldBecomeADirective(labeled *ast.Node, statement *ast.LabeledStatement) bool {
	// The immediate body, deliberately not the bottom of a label chain.
	//
	// This port briefly walked through nested labels to the statement they eventually wrap, on the
	// reasoning that removing every label in `A: B: "use strict"` ends at a bare directive. That is
	// true and it is not this label's problem: removing only `A` leaves `B: "use strict"`, which is
	// still labeled and so still not a directive. Each label is judged on what its own removal
	// produces, and the last one to go is refused on its own pass, which is what keeps the chain
	// from ever collapsing into a directive.
	//
	// Checked against ESLint rather than reasoned about, after the chain walk started withholding
	// repairs upstream offers: ESLint's own Linter fixes `A: B: 'foo'` to `B: 'foo'`, and its
	// isFixable reads node.body directly for this reason.
	if statement.Statement == nil || !ast.IsExpressionStatement(statement.Statement) {
		return false
	}
	expression := statement.Statement.AsExpressionStatement().Expression
	if expression == nil {
		return false
	}
	// A no-substitution template is the template form of the same hazard, and a parenthesized
	// expression is deliberately not included: parentheses stop it being a directive at all.
	if expression.Kind != ast.KindStringLiteral && expression.Kind != ast.KindNoSubstitutionTemplateLiteral {
		return false
	}

	// The outermost enclosing label, since nested labels are removed one at a time and the position
	// that matters is the whole chain's.
	outermost := labeled
	for outermost.Parent != nil && ast.IsLabeledStatement(outermost.Parent) {
		outermost = outermost.Parent
	}
	parent := outermost.Parent
	if parent == nil {
		return false
	}
	if ast.IsSourceFile(parent) {
		return true
	}
	return ast.IsBlock(parent) && parent.Parent != nil && ast.IsFunctionLikeDeclaration(parent.Parent)
}

// wouldRejoinWithThePreviousLine says whether removing the label creates an ASI hazard.
//
// `foo()\nLABEL: [1].forEach(x => x)` becomes `foo()\n[1].forEach(x => x)`, which parses as
// `foo()[1]` rather than as two statements: the label was the only thing keeping them apart. The
// rewrite parses cleanly and means something else, so nothing downstream would refuse it.
//
// The test is upstream's and it is deliberately about text rather than about the tree. If the token
// before the label already terminates a statement then nothing can rejoin, and otherwise a body
// opening with one of the continuation characters would bind to it.
func wouldRejoinWithThePreviousLine(ctx rule.Context, labelStart int, bodyStart int) bool {
	source := ctx.SourceFile.Text()
	if bodyStart >= len(source) {
		return false
	}

	previous := 0
	for index := labelStart - 1; index >= 0; index-- {
		if character := source[index]; character != ' ' && character != '\t' &&
			character != '\n' && character != '\r' {
			previous = int(character)
			break
		}
	}
	// Nothing before the label, or a token that already closes a statement, so no rejoining.
	if previous == 0 || previous == ':' || previous == ';' || previous == '{' {
		return false
	}

	switch source[bodyStart] {
	case '(', '[', '-', '+', '/', '`':
		return true
	}
	return false
}
