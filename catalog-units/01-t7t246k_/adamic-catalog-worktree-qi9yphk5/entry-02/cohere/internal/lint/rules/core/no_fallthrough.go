package core

import (
	"regexp"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/ecmascript/comments"
	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoFallthroughOptions configures which comments excuse a fallthrough and which clauses are exempt.
//
// This is the whole option surface, read off ESLint's `meta.schema`: one object with exactly three
// properties and `additionalProperties: false`. oxc's config struct carries the same three under
// `deny_unknown_fields`, so the two agree and there is nothing else to support.
type NoFallthroughOptions struct {
	// CommentPattern replaces the default set of accepted fallthrough comments with a regular
	// expression. Compiled case-insensitively, matching oxc, which prepends `(?iu)` to whatever the
	// caller wrote. An unparseable pattern falls back to the default set rather than matching
	// everything, because a rule that silently stops reporting is worse than one that ignores a
	// misconfiguration.
	CommentPattern string `json:"commentPattern"`

	// AllowEmptyCase lets a clause with no statements at all fall through even when a blank line
	// separates it from the next one. Off by default: a clause followed by a blank line reads as a
	// body somebody forgot to write rather than as a deliberate grouping.
	AllowEmptyCase bool `json:"allowEmptyCase"`

	// ReportUnusedFallthroughComment reports a fallthrough comment on a clause that cannot fall
	// through, which is a comment describing behavior the code no longer has.
	ReportUnusedFallthroughComment bool `json:"reportUnusedFallthroughComment"`
}

// NoFallthrough reports a switch clause whose body runs on into the next clause.
//
// Falling from one case into the next is legal and occasionally deliberate, but far more often it
// is a missing `break`, and the two are indistinguishable from the source. So the rule asks for the
// deliberate case to be said out loud in a comment, and reports the rest.
//
// Examples of incorrect code:
//
//	switch (foo) { case 0: doSomething(); case 1: doSomethingElse(); }
//	switch (foo) { case 0: if (a) { break; } default: b(); }
//
// Examples of correct code:
//
//	switch (foo) { case 0: doSomething(); break; case 1: doSomethingElse(); }
//	switch (foo) { case 0: doSomething(); /* falls through */ case 1: doSomethingElse(); }
//	switch (foo) { case 0: case 1: doSomething(); }
//
// # Why there is no control flow graph here
//
// Upstream reaches for oxc's control flow graph (`ctx.cfg()`) and, in the same file, spends fifty
// lines warning that the reach is a hack: `get_switch_semantic_cases` is labelled black magic, told
// not to copy, and pinned to an open issue asking for a semantic interface instead. We have no such
// graph, and TypeScript's flow nodes are not one. They exist for *narrowing*, keyed on expression
// positions where a type could change, and `ast.FlowNode` carries `Antecedent` and `Antecedents`
// with no successor edges at all. There is no break instruction in it to find.
//
// So the analysis here is structural, over the statement tree, and for this question the tree is
// the better instrument rather than the weaker one. "Does control reach the end of this clause" is
// a property of how statements nest, and the constructs that make it interesting are all visible:
//
//	break/return/throw/continue   leave the clause outright
//	if                            leaves only when it has an else and both arms leave
//	try                           leaves when the finally leaves, or the try and the catch both do
//	switch                        a nested switch's break is its own; it never leaves the outer one
//	loops                         a break inside one belongs to the loop, not the switch
//
// The loop rule is what upstream's control flow graph gets right through a bypass edge and what the
// tree gets right by declining to credit the body, and both of upstream's `while (a) { break; }` and
// `do { break; } while (a)` fail cases turn on exactly it.
//
// The judgment here is close to `getter_return.go`'s `bodyDefinitelyExits` and is deliberately not
// shared with it, because the two ask different questions. That one asks whether every path leaves
// the *function*, so a `break` is not an exit and a `continue` is not an exit. This one asks
// whether every path leaves the *switch*, so both are, and a nested loop or switch swallows them.
// Sharing one walk would need a mode flag threaded through every arm, and the arms that would then
// differ are the majority.
var NoFallthrough = rule.Rule{
	Name: "no-fallthrough",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		parsed, _ := rule.OptionsAs[NoFallthroughOptions](options)
		matchesFallthroughComment := fallthroughCommentMatcher(parsed.CommentPattern)

		return rule.Listeners{
			ast.KindSwitchStatement: func(node *ast.Node) {
				clauses := node.AsSwitchStatement().CaseBlock.AsCaseBlock().Clauses.Nodes
				sourceText := ctx.SourceFile.Text()

				for index := 0; index+1 < len(clauses); index++ {
					clause := clauses[index]
					nextClause := clauses[index+1]

					// A clause's `Pos()` is where the previous clause ended, so the trivia between
					// two clauses belongs to the *next* one and every window measured to
					// `nextClause.Pos()` is empty. What both the comment scan and the blank-line
					// test want is the next clause's first token, which is what `TokenRange` finds.
					// This cost the first eighteen fixtures and is the single most load-bearing
					// line in the rule.
					nextClauseStart := rule.TokenRange(ctx.SourceFile, nextClause).Pos()

					if clauseAlwaysLeavesTheSwitch(clause) {
						// The clause cannot fall through, so a comment saying it does is describing
						// behavior the code does not have. Only reported when asked for, since a
						// comment left behind by a later `break` is a nuisance rather than a bug.
						if parsed.ReportUnusedFallthroughComment {
							if commentRange, found := fallthroughCommentBetween(ctx, clause,
								nextClauseStart, matchesFallthroughComment); found {
								ctx.ReportRange(commentRange, rule.Message{
									Id: "unusedFallthroughComment",
									Description: "This comment says the case falls through, but " +
										"the case cannot fall through. Remove the comment, or " +
										"remove whatever leaves the case.",
								})
							}
						}
						continue
					}

					// An empty clause is `case 0: case 1:`, a deliberate grouping that reads as one
					// clause with several labels, so it is exempt. The exception is an empty clause
					// separated from the next by a *blank* line, which reads as a body somebody
					// meant to write. `allowEmptyCase` turns that exception off.
					if len(clause.AsCaseOrDefaultClause().Statements.Nodes) == 0 {
						if parsed.AllowEmptyCase ||
							!hasBlankLineBetween(sourceText, clause.End(), nextClauseStart) {
							continue
						}
					}

					if _, found := fallthroughCommentBetween(ctx, clause, nextClauseStart,
						matchesFallthroughComment); found {
						continue
					}

					// The finding points at the clause fallen *into*, not the one that failed to
					// break, because that is the line a reader has to look at to see the two
					// bodies running together. Upstream's snapshot underlines the same span.
					messageId := "case"
					if nextClause.Kind == ast.KindDefaultClause {
						messageId = "default"
					}
					ctx.ReportNode(nextClause, rule.Message{
						Id: messageId,
						Description: "Control runs from the previous case into this one, which is " +
							"almost always a missing break. Add a break, or say the fallthrough " +
							"is deliberate with a `falls through` comment.",
					})
				}
			},
		}
	},
}

// fallthroughCommentMatcher builds the predicate deciding whether a comment excuses a fallthrough.
//
// The default set is oxc's four exact spellings rather than ESLint's `/falls?\s?through/iu`, which
// is a real divergence between the two upstreams and is reproduced rather than resolved: oxc fails
// `/* falling through */` and ESLint's pattern would pass it, and that input is in oxc's corpus as
// a fail case. Following ESLint here would turn one of the imported fail fixtures green.
func fallthroughCommentMatcher(pattern string) func(string) bool {
	if pattern != "" {
		// oxc prepends `(?iu)`, so a lower-case pattern matches an upper-case comment. The `u` flag
		// has no Go equivalent and needs none: Go's regexp is Unicode-aware by default and its
		// classes already match what the Rust crate's `u` turns on.
		if compiled, compileError := regexp.Compile("(?i)" + pattern); compileError == nil {
			return func(comment string) bool { return compiled.MatchString(comment) }
		}
		// A pattern that will not compile falls through to the default set. Treating it as
		// "matches everything" would take the rule silent across every file the option touches,
		// which is the failure mode a misconfiguration must not be able to cause.
	}
	return func(comment string) bool {
		switch strings.ToLower(text.TrimWhitespace(comment)) {
		case "falls through", "fall through", "fallsthrough", "fallthrough":
			return true
		}
		return false
	}
}

// fallthroughCommentBetween finds the comment, if any, that excuses falling out of a clause.
//
// Two windows, checked in order, and both are upstream's. When the clause's only statement is a
// block, the comment may sit *inside* it before the closing brace, which is the shape
// `case 0: { a(); /* falls through */ }` relies on. Otherwise, and as a fallback, the comment sits
// between the end of the clause's last statement and the start of the next clause.
//
// Only the last comment in a window counts. That is what separates upstream's clean
// `{ a(); /* falls through */ } /* comment */` from its failing
// `/* no break */ /* todo: fix readability */`: in the first the matching comment is last in the
// block window, in the second a non-matching comment follows it in the only window there is.
func fallthroughCommentBetween(ctx rule.Context, clause *ast.Node, nextClauseStart int,
	matches func(string) bool) (core.TextRange, bool) {
	statements := clause.AsCaseOrDefaultClause().Statements.Nodes
	fileComments := comments.ForFile(ctx)

	// The window's start is the clause's end, which is also its last statement's end.
	//
	// It was written as "the last statement's end, or the clause's end when there are none" and a
	// mutation collapsing it to `clause.End()` alone survived every fixture. Probed rather than
	// argued: a case clause carries no token after its final statement, so across seven shapes
	// including a trailing empty statement, an unterminated call, and a `do-while` with no
	// semicolon, the two positions were equal on every one. The branch was subsumed rather than
	// unseen, so it is deleted with the reason at the line instead of tested with a fixture that
	// could assert nothing.
	windowStart := clause.End()

	if len(statements) == 1 && statements[0].Kind == ast.KindBlock {
		block := statements[0]
		blockStatements := block.AsBlock().Statements.Nodes
		blockWindowStart := block.Pos()
		if len(blockStatements) > 0 {
			blockWindowStart = blockStatements[len(blockStatements)-1].End()
		}
		if commentRange, found := lastCommentIn(fileComments, blockWindowStart, block.End(),
			matches); found {
			return commentRange, true
		}
	}

	return lastCommentIn(fileComments, windowStart, nextClauseStart, matches)
}

// lastCommentIn returns the last comment fully inside a range, when it matches.
//
// This used to carry its own scanner sweep, because `internal/utilities/comments` could not see a
// comment that was the only content of a block and this rule's corpus contains exactly that:
// `switch(foo) { case 0: { /* falls through */ } case 1: b(); }` returned zero comments from the
// shelf. The shelf was fixed, so the local scan is gone and this filters the shared scan instead.
// The gap turned out to be wider than a block, and the fix is described where it lives.
//
// Filtering rather than scanning is also why this is now cheap. `comments.ForFile` is computed once
// per file and shared with the four other comment rules, so the two windows this rule asks about
// cost a walk over that file's comments rather than a fresh sweep of the source per window.
func lastCommentIn(fileComments []comments.Comment, start int, end int,
	matches func(string) bool) (core.TextRange, bool) {
	if start < 0 || start >= end {
		return core.TextRange{}, false
	}

	// The last one wins, and the comments arrive in source order, so the final match in the window
	// is the answer. That ordering is what separates upstream's clean
	// `{ a(); /* falls through */ } /* comment */` from its failing
	// `/* no break */ /* todo: fix readability */`: in the first the matching comment is last in the
	// block window, in the second a non-matching comment follows it in the only window there is.
	var lastRange core.TextRange
	var lastText string
	found := false
	for _, comment := range fileComments {
		if comment.Range.Pos() < start || comment.Range.End() > end {
			continue
		}
		lastRange = comment.Range
		lastText = comment.Text
		found = true
	}

	if !found {
		return core.TextRange{}, false
	}
	if !isFallthroughComment(lastText, matches) {
		return core.TextRange{}, false
	}
	return lastRange, true
}

// isFallthroughComment answers whether a comment's content excuses a fallthrough.
//
// The directive guard comes first and is not part of the pattern. A comment opening `eslint-` or
// `oxlint-` is an instruction to the linter, and a custom `commentPattern` broad enough to match
// one must not turn a suppression directive into a fallthrough excuse. Upstream checks this before
// consulting the pattern for the same reason.
func isFallthroughComment(commentText string, matches func(string) bool) bool {
	content := commentText
	switch {
	case strings.HasPrefix(content, "//"):
		content = content[2:]
	case strings.HasPrefix(content, "/*"):
		content = strings.TrimSuffix(content[2:], "*/")
	}
	// ESLint trims with JavaScript's trim (no-fallthrough.js:29), the rule's authority over oxc's str::trim.
	content = text.TrimWhitespace(content)

	if strings.HasPrefix(content, "oxlint-") || strings.HasPrefix(content, "eslint-") {
		return false
	}
	return matches(content)
}

// hasBlankLineBetween answers whether two positions are separated by a blank line.
//
// Two newlines rather than one, because a clause and the next one are normally on separate lines
// and that first newline is formatting. A second means somebody left a gap, and a gap after an
// empty clause reads as a body that was meant to go there.
func hasBlankLineBetween(sourceText string, start int, end int) bool {
	if start < 0 || end > len(sourceText) || start >= end {
		return false
	}
	return strings.Count(sourceText[start:end], "\n") >= 2
}

// clauseAlwaysLeavesTheSwitch answers whether control can reach the end of a clause's body.
//
// This is the judgment the control flow graph was doing upstream, asked structurally. It is close
// to `getter_return.go`'s `bodyDefinitelyExits` and deliberately separate: that one asks whether
// every path leaves the enclosing *function*, this one whether every path leaves the enclosing
// *switch*, and a `break`, a `continue`, and a nested loop are all read oppositely by the two.
//
// An empty clause answers false and is handled by the caller rather than here, because whether an
// empty clause may fall through is an option rather than a fact about control flow.
func clauseAlwaysLeavesTheSwitch(clause *ast.Node) bool {
	for _, statement := range clause.AsCaseOrDefaultClause().Statements.Nodes {
		if statementLeavesTheSwitch(statement) {
			return true
		}
	}
	return false
}

// statementLeavesTheSwitch answers whether every path through a statement leaves the enclosing
// switch clause.
//
// Anything not named answers false, which is the reporting direction: an unrecognized statement is
// assumed to fall through, so the rule speaks up rather than going quiet. That is the right bias
// for a correctness rule, and it is why the recognized set covers every construct the corpus draws
// a line with.
func statementLeavesTheSwitch(statement *ast.Node) bool {
	if statement == nil {
		return false
	}

	switch statement.Kind {
	case ast.KindBreakStatement:
		// A bare `break` leaves the switch. A labeled one leaves whatever it names, and the label
		// is necessarily on something enclosing this clause, so it leaves the switch too. The case
		// where a labeled break does *not* leave the switch is a label on a loop or block inside
		// the clause, and that break is inside the loop, so this arm never sees it: the loop arm
		// below declines the whole loop first.
		return true

	case ast.KindContinueStatement:
		// `continue` can only target a loop, and no loop encloses this statement within the clause,
		// so the loop it targets is outside the switch and control leaves.
		return true

	case ast.KindReturnStatement, ast.KindThrowStatement:
		return true

	case ast.KindBlock:
		for _, inner := range statement.AsBlock().Statements.Nodes {
			if statementLeavesTheSwitch(inner) {
				return true
			}
		}
		return false

	case ast.KindIfStatement:
		// Both arms have to leave, and an absent else is an arm that does not. Upstream fails
		// `case 0: if (a) { break; } default:` for exactly this and passes the same input with an
		// `else { throw 0; }`. Nil is rejected on the first line of this function, so an absent
		// else needs no separate test.
		ifStatement := statement.AsIfStatement()
		return statementLeavesTheSwitch(ifStatement.ThenStatement) &&
			statementLeavesTheSwitch(ifStatement.ElseStatement)

	case ast.KindTryStatement:
		return tryStatementLeavesTheSwitch(statement.AsTryStatement())

	case ast.KindSwitchStatement:
		// A nested switch never lends its `break` outward: that break belongs to the inner switch.
		// The inner switch can still leave the outer one, but only through a return, a throw, or a
		// labeled break, and crediting that would need every clause of it to leave. Upstream does
		// not credit a nested switch at all, and its `case 0: switch(bar) { case 2: break; }` pass
		// case is clean only because of its comment, which is what pins the choice.
		return false

	case ast.KindLabeledStatement:
		// A label around a statement does not change whether it leaves, except that a `break` at
		// this label now lands here instead. That distinction is the loop and block arms' concern,
		// and both decline anyway, so the label is transparent.
		return statementLeavesTheSwitch(statement.AsLabeledStatement().Statement)

	case ast.KindWithStatement:
		return statementLeavesTheSwitch(statement.AsWithStatement().Statement)
	}

	// Loops land here and answer false. A `break` inside a loop breaks the loop and lands back in
	// the clause, and the body may run zero times besides, so nothing inside a loop leaves the
	// switch on every path. Upstream fails both `while (a) { break; }` and `do { break; } while (a)`
	// and this reproduces both. A `do-while` body that *throws* does leave, and upstream passes
	// `do { throw 0; } while(a)` for that reason, which this reproduces too by way of the loop
	// arm below.
	if statement.Kind == ast.KindDoStatement {
		// The one loop whose body is guaranteed to run. A throw or return inside it therefore
		// happens, and upstream's `case 0: do { throw 0; } while(a); default: b();` pass case turns
		// on it. A `break` inside is still the loop's own, which is why this asks a narrower
		// question than the general walk.
		return doWhileBodyLeavesTheSwitch(statement.AsDoStatement().Statement)
	}
	return false
}

// doWhileBodyLeavesTheSwitch asks whether a do-while body leaves the switch without crediting any
// break inside it.
//
// A do-while runs its body at least once, so a `return` or `throw` in it is unconditional. A
// `break` is not: it targets the loop and lands after it, still inside the clause. So this is a
// second, stricter walk rather than a reuse of the general one, and it is small because only the
// nesting constructs need to recurse.
func doWhileBodyLeavesTheSwitch(statement *ast.Node) bool {
	if statement == nil {
		return false
	}
	switch statement.Kind {
	case ast.KindReturnStatement, ast.KindThrowStatement:
		return true
	case ast.KindBlock:
		for _, inner := range statement.AsBlock().Statements.Nodes {
			if doWhileBodyLeavesTheSwitch(inner) {
				return true
			}
		}
		return false
	case ast.KindIfStatement:
		ifStatement := statement.AsIfStatement()
		return doWhileBodyLeavesTheSwitch(ifStatement.ThenStatement) &&
			doWhileBodyLeavesTheSwitch(ifStatement.ElseStatement)
	case ast.KindLabeledStatement:
		return doWhileBodyLeavesTheSwitch(statement.AsLabeledStatement().Statement)
	}
	return false
}

// tryStatementLeavesTheSwitch answers whether a try statement leaves the switch on every path.
//
// A finally that leaves wins outright, since it runs on every path out of the try and the catch
// both. Otherwise the try block has to leave, and if a catch exists it has to leave as well,
// because the try may throw partway through and then only the catch runs. Upstream pins the pair:
// `try { break; } finally {}` passes and `try { throw 0; } catch (err) {}` fails.
func tryStatementLeavesTheSwitch(statement *ast.TryStatement) bool {
	if statement.FinallyBlock != nil &&
		statementLeavesTheSwitch(statement.FinallyBlock.AsNode()) {
		return true
	}
	if !statementLeavesTheSwitch(statement.TryBlock.AsNode()) {
		return false
	}
	if statement.CatchClause == nil {
		return true
	}
	return statementLeavesTheSwitch(statement.CatchClause.AsCatchClause().Block.AsNode())
}
