package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/ecmascript/comments"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageExtraLabel = rule.Message{
	Id: "unexpected",
	Description: "This label is unnecessary. The nearest enclosing breakable statement is already " +
		"the one this jump targets, so naming it changes nothing and only invites the reader to " +
		"go looking for the outer statement it appears to reach.",
}

// NoExtraLabel flags a labeled `break` or `continue` whose label names the statement the jump would
// have reached anyway.
//
//	valid:   A: break A;
//	valid:   A: while (a) { while (b) { break A; } }
//	valid:   A: while (a) { switch (b) { case 0: break A; } }
//	invalid: A: while (a) break A;
//	invalid: A: do { break A; } while (a);
//	invalid: A: switch (a) { case 0: break A; }
//
// # The judgment is about the nearest scope, not about matching names
//
// An unlabeled `break` leaves the innermost enclosing breakable statement, which is a loop or a
// switch. So `break A` is redundant exactly when the innermost breakable statement is itself
// labeled `A`, and it is load-bearing whenever anything breakable sits between the jump and the
// label. That is why `A: while (a) { while (b) { break A; } }` is clean while
// `A: while (a) break A;` is not: the inner `while` intercepts the plain break in the first, and in
// the second nothing does.
//
// The distinction is easy to lose by matching names. A rule reporting whenever a jump's label names
// some enclosing label reports every one of upstream's fifteen passing cases, because in all of
// them the label does enclose the jump. The scope stack is the whole rule.
//
// # A labeled statement whose body is breakable contributes ONE scope, not two
//
// In `A: while (a) ...` the label and the loop are one target: an unlabeled break leaves that loop,
// and the loop is what `A` names. Upstream builds the stack so a labeled breakable statement pushes
// a single entry carrying both facts, and a labeled non-breakable statement (`A: { ... }`) pushes an
// entry that is labeled and not breakable. Pushing two entries for the labeled loop would put a
// breakable-but-unlabeled frame above the labeled one, and `A: while (a) break A;` would go silent,
// which is the exact case the rule exists for.
//
// The walk therefore reads a breakable statement's own parent to learn whether it is labeled, rather
// than the labeled statement pushing on its behalf.
//
// # A divergence our parser creates, and upstream cannot express
//
// `A: switch (a) { case 0: continue A; }` reports here and is a SYNTAX ERROR upstream, because
// `continue` may only target a loop and ESLint refuses to parse it. Our parser recovers and hands
// the rule a continue statement, so the rule judges it and finds the label redundant against the
// switch frame. The same applies to any duplicate label in one scope, which is likewise illegal.
//
// Measured rather than reasoned: driving ESLint's own Linter over these inputs returns a finding
// with a null ruleId and the message "Label 'A' is already declared" or "Unsyntactic break", which
// reads as a count of one and is not a finding at all. Four of the ten inputs in an early probe here
// were that shape, and reading only the counts would have recorded a rule level divergence that does
// not exist. Sixteen well formed inputs were then measured against the installed rule and all
// sixteen agree.
//
// # `continue` reads the same stack, and it is not a separate judgment
//
// Upstream routes `break` and `continue` through one function, and the stack it consults counts a
// switch as breakable. That is deliberate rather than sloppy: `continue` cannot target a switch, so
// a `continue A` inside a switch inside loop `A` is legal source that this rule leaves alone by
// stopping the search at the switch frame. Reproduced rather than corrected, because the alternative
// is a rule that reports a jump whose label is doing real work. Upstream's
// `A: while (a) { switch (b) { case 0: continue A; } }` is a passing case and pins it.
//
// # The fix, and the comment it refuses to destroy
//
// Removing a redundant label deletes the text between the end of the `break` or `continue` keyword
// and the end of the label, which leaves `break A;` as `break;`. Upstream withholds the repair when
// a comment sits inside that span, because the removal would silently delete it, and reports the
// finding anyway. Both halves are ported: four of upstream's nineteen failing cases carry
// `output: null` for exactly this reason, and a comment written just outside the span still fixes,
// which is what separates `break/**/ A` (declines) from `/*comment*/break A` (fixes).
//
// The span runs from the keyword's END rather than from the label's start, so a comment written
// after the label survives: upstream rewrites `continue A/*comment*/;` to `continue/*comment*/;`.
//
// The verdict at any jump depends on every enclosing statement, which the pre-order walk has not
// visited yet when the jump's own listener would fire. So the scope stack is built and read inside
// one source-file listener, which fires before its children and lets this rule own its own descent.
var NoExtraLabel = rule.Rule{
	Name: "no-extra-label",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				// The enclosing scope stack, innermost last. Built during the descent and read at
				// each jump, which is why the walk is written here rather than as sibling listeners.
				scopes := []labelScope{}

				var visit func(*ast.Node) bool
				visit = func(current *ast.Node) bool {
					pushed := false

					switch {
					case isBreakableStatement(current):
						// A breakable statement carries its label when a labeled statement wraps it
						// directly, which is the one-frame-not-two rule above.
						var label *ast.Node
						if current.Parent != nil && ast.IsLabeledStatement(current.Parent) {
							label = current.Parent.AsLabeledStatement().Label
						}
						scopes = append(scopes, labelScope{label: label, breakable: true})
						pushed = true

					case ast.IsLabeledStatement(current):
						// A labeled breakable statement is handled by the branch above when the
						// walk reaches the body, so this pushes only for the other bodies.
						//
						// The breakable half of that guard is EQUIVALENT rather than load bearing,
						// and it is kept because it states the intent. Dropping it makes a labeled
						// loop push two frames, the label frame first and the breakable frame for
						// its own body second; the search runs innermost first and always stops at
						// a breakable frame, so the label frame beneath it can never be read.
						// Measured: 80 enumerated sources over five outer statement forms, both
						// label names, both jump kinds and both targets produced byte identical
						// findings under both spellings.
						//
						// The OTHER half is load bearing and its mutant is caught. A labeled block
						// must push, so that a jump naming it stops there instead of reaching an
						// outer loop of the same name. That shape needs a function boundary to be
						// legal source, which is why upstream's corpus cannot reach it and why
						// TestNoExtraLabelStopsAtALabeledBlockOfTheSameName exists.
						labeled := current.AsLabeledStatement()
						if labeled != nil && !isBreakableStatement(labeled.Statement) {
							scopes = append(scopes, labelScope{label: labeled.Label, breakable: false})
							pushed = true
						}

					case current.Kind == ast.KindBreakStatement:
						reportIfLabelIsRedundant(ctx, current, current.AsBreakStatement().Label, scopes)

					case current.Kind == ast.KindContinueStatement:
						reportIfLabelIsRedundant(ctx, current, current.AsContinueStatement().Label, scopes)
					}

					current.ForEachChild(visit)

					if pushed {
						scopes = scopes[:len(scopes)-1]
					}
					return false
				}
				node.ForEachChild(visit)
			},
		}
	},
}

// labelScope is one frame of the enclosing-statement stack.
//
// A frame is breakable when an unlabeled `break` would leave it, and carries a label when a labeled
// statement names it. A labeled loop is both, which is the case the whole rule turns on.
type labelScope struct {
	label     *ast.Node
	breakable bool
}

// isBreakableStatement says whether an unlabeled `break` inside this statement would leave it.
//
// This is upstream's `astUtils.isBreakableStatement`: the four loop forms and a switch. A labeled
// block is deliberately absent, because a plain `break` cannot leave one, which is what makes
// `A: { while (b) { break A; } }` clean.
func isBreakableStatement(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindWhileStatement, ast.KindDoStatement, ast.KindForStatement,
		ast.KindForInStatement, ast.KindForOfStatement, ast.KindSwitchStatement:
		return true
	}
	return false
}

// reportIfLabelIsRedundant walks outward from a jump and reports when its label names the frame the
// jump would have reached without it.
//
// The search stops at the first frame that either is breakable or carries the jump's own label,
// because that frame is where the jump lands one way or the other. It reports only when the same
// frame is both, which is exactly upstream's condition and the reason a plain outward name search
// is wrong: in `A: while (a) { while (b) { break A; } }` the inner loop is the first breakable frame
// and it carries no label, so the search stops there having reported nothing.
func reportIfLabelIsRedundant(ctx rule.Context, jump *ast.Node, label *ast.Node, scopes []labelScope) {
	if label == nil {
		return
	}
	name := label.Text()

	for index := len(scopes) - 1; index >= 0; index-- {
		scope := scopes[index]
		namesThisScope := scope.label != nil && scope.label.Text() == name
		if !scope.breakable && !namesThisScope {
			continue
		}
		if scope.breakable && namesThisScope {
			if removal, ok := extraLabelRemovalFix(ctx, jump, label); ok {
				ctx.ReportNodeWithFixes(label, messageExtraLabel, removal)
			} else {
				ctx.ReportNode(label, messageExtraLabel)
			}
		}
		// Either way this frame is where the jump lands, so nothing above it can matter.
		return
	}
}

// extraLabelRemovalFix builds the repair that deletes a redundant label, or declines to.
//
// The span runs from the end of the `break` or `continue` keyword to the end of the label, which
// turns `break A` into `break` while leaving anything written after the label in place. Upstream
// computes the same two offsets from the first token of the jump and from the label's range end.
//
// It declines when a comment sits inside that span, because the removal would delete it with no
// warning and the engine applies a fix unattended. A comment outside the span is untouched and the
// repair is still offered, which is why `/*comment*/break A;` fixes and `break/**/ A;` does not.
func extraLabelRemovalFix(ctx rule.Context, jump *ast.Node, label *ast.Node) (rule.Fix, bool) {
	keywordStart := scanner.GetTokenPosOfNode(jump, ctx.SourceFile, false)
	keywordEnd := keywordStart + len(jumpKeywordText(jump))
	labelEnd := label.End()

	// An ordering guard rather than a behavioral filter, kept as crash protection on a range the
	// engine would otherwise be handed inverted.
	//
	// Its mutant survives and no input reaches it: a jump carrying a label always writes that label
	// after the keyword, and a jump carrying none exits at the nil check above. Probed over eight
	// shapes including a bare `break`, a break split across a newline, a break with no semicolon and
	// a comment butted against the keyword, and none produced an inverted span. The verdict names
	// those callers, so adding another caller voids it.
	//
	// It stays because a fix is applied unattended and an inverted range is the one repair the edit
	// engine's parse check cannot refuse, and because keywordEnd is computed from a token LENGTH
	// rather than scanned, so a future jump kind reaching jumpKeywordText would land here silently.
	if keywordEnd >= labelEnd {
		return rule.Fix{}, false
	}
	for _, comment := range comments.ForFile(ctx) {
		if comment.Range.Pos() < labelEnd && comment.Range.End() > keywordEnd {
			return rule.Fix{}, false
		}
	}

	return rule.RemoveRange(core.NewTextRange(keywordEnd, labelEnd)), true
}

// jumpKeywordText gives the keyword a jump statement opens with.
//
// Read as a length rather than rescanned, because the keyword is fixed by the node's kind and both
// spellings are known. A jump reaching here is always one of the two, since the only callers are the
// two switch arms that matched those kinds.
func jumpKeywordText(jump *ast.Node) string {
	if jump.Kind == ast.KindContinueStatement {
		return "continue"
	}
	return "break"
}
