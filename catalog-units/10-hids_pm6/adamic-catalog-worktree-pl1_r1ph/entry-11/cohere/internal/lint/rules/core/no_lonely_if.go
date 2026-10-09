package core

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageUnexpectedLonelyIf = rule.Message{
	Id: "unexpectedLonelyIf",
	Description: "This `else` block holds nothing but an `if`, so the braces add a level of " +
		"indentation without adding a branch. Written as `else if`, the chain reads as the one " +
		"sequence of alternatives it already is, and every later condition lines up with the " +
		"first instead of stepping right.",
}

// NoLonelyIf flags an `if` that is the only statement inside an `else` block.
//
//	valid:   if (a) {;} else if (b) {;}
//	valid:   if (a) {;} else { if (b) {;} ; }
//	valid:   if (a) if (a) {} else { if (b) {} } else {}
//	invalid: if (a) {;} else { if (b) {;} }
//	invalid: if (foo) {} else { if (bar) baz(); }
//
// # The shape being matched
//
// An `if` whose parent is a block, where that block holds exactly one statement, and where the
// block is the ELSE branch of an enclosing `if`. The last test is identity against the enclosing
// statement's else branch rather than a kind test, because the then branch of an `if` is a block
// holding one `if` just as often and is not this rule's subject.
//
// # The third clean case is a dangling-else guard, and it is the whole of upstream's helper
//
// `if (a) if (a) {} else { if (b) {} } else {}` is clean, and nothing about "one statement in an
// else block" explains that. The braces are load-bearing: remove them and the trailing `else {}`
// would bind to the inner `if (b)` rather than to the outer one, which is a different program.
// Upstream calls `astUtils.areBracesNecessary` for this, and the part that fires here is
// `hasUnsafeIf(statement) && isFollowedByElseKeyword(node)`.
//
// The other half of that helper, `isLexicalDeclaration`, cannot fire from this rule at all: the
// statement being tested is the block's only statement AND this listener already knows it is an
// `if`, so it is never a `let`, `const`, `using`, function or class declaration. It is left out
// rather than transcribed, and the reasoning is here so the omission does not read as a gap.
//
// `hasUnsafeIf` is recursive and the recursion is load-bearing. It asks whether the code would end
// with an `if` that has no `else` of its own, following the alternate chain down and stepping
// through loop and label bodies, which are the constructs whose body a dangling `else` can reach
// through. Measured against the installed build at 10.8.1:
//
//	if (a) if (a) {} else { if (b) {} } else {}                clean, the inner if has no else
//	if (a) if (a) {} else { if (b) {} }                        REPORTS, nothing follows to bind
//	if (a) if (a) {} else { if (b) {} else {} } else {}        REPORTS, the chain ends with an else
//	if (a) if (a) {} else { if (b) {} else if (c) {} } else {} clean, the chain ends without one
//
// The third of those produces broken output when upstream fixes it, which is upstream's own defect
// and is reproduced rather than corrected, because a port that silently improves on the original
// is a divergence nobody can see.
//
// # The fixer, and the five things it refuses
//
// Upstream declares `fixable: "code"` and rewrites the else block to `else if`, replacing the
// braces with the inner statement's own text. The corpus asserts the output text on ten cases and
// declines on five, and those declines are the port rather than a detail:
//
//	a comment before the inner if      the text between `{` and the `if` is not just whitespace,
//	                                   so removing the braces would drop it
//	a comment after the inner if       same, on the other side
//	the fix would create a syntax
//	  error or change meaning by
//	  automatic semicolon insertion    three separate shapes, all where the inner consequent has
//	                                   no braces and no terminating semicolon
//
// The semicolon-insertion group is the subtle one, and each shape is in the corpus. An unbraced
// consequent whose statement is not terminated joins the line after it: `baz()` followed by
// `qux();` on the same line becomes one call, `baz()` followed by a line opening `[` becomes a
// subscript, and `baz++` followed by anything becomes an operand. The same sources WITH a semicolon
// are fixed, and both spellings are in the corpus for the comparison.
//
// A wrong fix here is worse than no fix, because the engine's only guard is that the result parses
// and every one of these parses. So the finding still fires in all five and only the repair is
// withheld, which is upstream's own trade.
var NoLonelyIf = rule.Rule{
	Name: "no-lonely-if",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindIfStatement: func(node *ast.Node) {
				block := node.Parent
				if block == nil || block.Kind != ast.KindBlock {
					return
				}
				if len(block.AsBlock().Statements.Nodes) != 1 {
					return
				}
				enclosing := block.Parent
				if enclosing == nil || enclosing.Kind != ast.KindIfStatement {
					return
				}
				// Identity against the ELSE branch. A then branch is the same shape and is not the
				// subject, so a kind test here would report every `if (a) { if (b) {} }`.
				if enclosing.AsIfStatement().ElseStatement != block {
					return
				}
				if noLonelyIfBracesAreNecessary(ctx, node, block) {
					return
				}

				if fix, canFix := noLonelyIfFix(ctx, node, block, enclosing); canFix {
					ctx.ReportNodeWithFixes(node, messageUnexpectedLonelyIf, fix)
					return
				}
				ctx.ReportNode(node, messageUnexpectedLonelyIf)
			},
		}
	},
}

// noLonelyIfBracesAreNecessary is the half of upstream's `areBracesNecessary` that can fire here.
//
// Upstream's helper is `isLexicalDeclaration(statement) || (hasUnsafeIf(statement) &&
// isFollowedByElseKeyword(node))`. The first disjunct is unreachable from this rule, because the
// statement it tests is the one this listener already knows to be an `if`, so only the second is
// transcribed. The reasoning is recorded rather than the code, so the omission does not read as a
// gap somebody should fill.
func noLonelyIfBracesAreNecessary(ctx rule.Context, node *ast.Node, block *ast.Node) bool {
	return noLonelyIfHasUnsafeIf(node) && noLonelyIfIsFollowedByElse(ctx, block)
}

// noLonelyIfHasUnsafeIf answers whether the code would end with an `if` carrying no `else`.
//
// Upstream's `hasUnsafeIf`, transcribed including its recursion. An `if` with no alternate is the
// unsafe case outright; one with an alternate defers to that alternate, which is how a long
// `else if` chain is judged by its last link. Loop bodies and labeled statements are stepped
// through because a dangling `else` reaches into them.
//
// `WithStatement` is upstream's fifth stepped-through construct and has no counterpart to name
// here: this parser has no `with` node kind, since the construct is illegal in the strict mode
// every TypeScript module is in. Its absence is a fact about the input language rather than a
// narrowing of the rule.
func noLonelyIfHasUnsafeIf(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindIfStatement:
		alternate := node.AsIfStatement().ElseStatement
		if alternate == nil {
			return true
		}
		return noLonelyIfHasUnsafeIf(alternate)
	case ast.KindForStatement:
		return noLonelyIfHasUnsafeIf(node.AsForStatement().Statement)
	case ast.KindForInStatement, ast.KindForOfStatement:
		return noLonelyIfHasUnsafeIf(node.AsForInOrOfStatement().Statement)
	case ast.KindWhileStatement:
		return noLonelyIfHasUnsafeIf(node.AsWhileStatement().Statement)
	case ast.KindLabeledStatement:
		return noLonelyIfHasUnsafeIf(node.AsLabeledStatement().Statement)
	}
	return false
}

// noLonelyIfIsFollowedByElse answers whether an `else` keyword comes directly after the block.
//
// Upstream reads the next token and compares it to `else`. Here the same question is asked of the
// enclosing statement rather than of the token stream: the block is some enclosing `if`'s else
// branch, so the only way an `else` can follow it is if a further enclosing `if` holds that
// statement as ITS then branch. Both spellings were compared against the installed build on the
// four dangling-else shapes above and they agree.
func noLonelyIfIsFollowedByElse(ctx rule.Context, block *ast.Node) bool {
	text := ctx.SourceFile.Text()
	position := scanner.SkipTrivia(text, block.End())
	return strings.HasPrefix(text[position:], "else") &&
		(position+4 >= len(text) || !noLonelyIfIsIdentifierCharacter(text[position+4]))
}

// noLonelyIfIsIdentifierCharacter keeps `elsewhere` from reading as `else`.
//
// A prefix test alone would match any identifier starting with those four letters, and the token
// after a block is exactly where such an identifier can appear.
func noLonelyIfIsIdentifierCharacter(character byte) bool {
	switch {
	case character >= 'a' && character <= 'z',
		character >= 'A' && character <= 'Z',
		character >= '0' && character <= '9',
		character == '_', character == '$':
		return true
	}
	return false
}

// noLonelyIfFix builds the repair, or declines.
//
// The repair replaces the else block's braces, inclusive, with the inner `if`'s own text, adding a
// space when the `else` keyword sits directly against the opening brace. Reproduced from upstream's
// fixer and checked against every `output` its corpus asserts.
func noLonelyIfFix(ctx rule.Context, node *ast.Node, block *ast.Node, enclosing *ast.Node) (rule.Fix, bool) {
	sourceText := ctx.SourceFile.Text()

	blockStart := rule.TokenRange(ctx.SourceFile, block).Pos()
	blockEnd := block.End()
	innerStart := rule.TokenRange(ctx.SourceFile, node).Pos()
	innerEnd := node.End()

	// A comment on either side of the inner statement would be discarded by the rewrite, and
	// upstream declines rather than dropping it. The test is upstream's own: whether the text
	// between the brace and the statement is anything but whitespace.
	if text.TrimWhitespace(sourceText[blockStart+1:innerStart]) != "" {
		return rule.Fix{}, false
	}
	if text.TrimWhitespace(sourceText[innerEnd:blockEnd-1]) != "" {
		return rule.Fix{}, false
	}

	if noLonelyIfFixWouldChangeMeaning(ctx, node, block) {
		return rule.Fix{}, false
	}

	separator := ""
	elseEnd := noLonelyIfElseKeywordEnd(ctx, enclosing, blockStart)
	if elseEnd == blockStart {
		separator = " "
	}

	return rule.ReplaceRange(core.NewTextRange(blockStart, blockEnd),
		separator+sourceText[innerStart:innerEnd]), true
}

// noLonelyIfElseKeywordEnd finds where the `else` keyword ends.
//
// Only its end is wanted, and only to decide whether a space has to be inserted: `else{ ... }`
// becomes `else if ...` rather than `elseif ...`. The keyword lies between the enclosing
// statement's then branch and the block, and it is the only `else` that can appear in that gap.
func noLonelyIfElseKeywordEnd(ctx rule.Context, enclosing *ast.Node, blockStart int) int {
	text := ctx.SourceFile.Text()
	thenEnd := enclosing.AsIfStatement().ThenStatement.End()
	if thenEnd < 0 || thenEnd > blockStart {
		return -1
	}
	offset := strings.LastIndex(text[thenEnd:blockStart], "else")
	if offset < 0 {
		return -1
	}
	return thenEnd + offset + len("else")
}

// noLonelyIfFixWouldChangeMeaning reproduces upstream's automatic-semicolon-insertion refusal.
//
// The hazard exists only when the inner `if`'s consequent is NOT a block and does not already end
// in a semicolon, because then removing the braces lets the following line join it. Upstream then
// declines on three signals, each of which has a corpus case:
//
//	the next token starts on the consequent's own line, so the two run together outright
//	the next token opens with one of `( [ / + ` -`, each of which continues an expression
//	the consequent ends in `++` or `--`, which take the next token as an operand
func noLonelyIfFixWouldChangeMeaning(ctx rule.Context, node *ast.Node, block *ast.Node) bool {
	sourceText := ctx.SourceFile.Text()
	consequent := node.AsIfStatement().ThenStatement
	if consequent == nil || consequent.Kind == ast.KindBlock {
		return false
	}

	lastToken := text.TrimWhitespace(sourceText[rule.TokenRange(ctx.SourceFile, consequent).Pos():consequent.End()])
	if strings.HasSuffix(lastToken, ";") {
		return false
	}

	afterBlock := scanner.SkipTrivia(sourceText, block.End())
	if afterBlock >= len(sourceText) {
		// Nothing follows, so nothing can join. `if (foo) {} else { if (bar) baz() }` at the end of
		// a file is fixable, which is why this is a guard rather than a refusal.
		return false
	}

	if noLonelyIfOnSameLine(sourceText, consequent.End(), afterBlock) {
		return true
	}
	switch sourceText[afterBlock] {
	case '(', '[', '/', '+', '`', '-':
		return true
	}
	return strings.HasSuffix(lastToken, "++") || strings.HasSuffix(lastToken, "--")
}

// noLonelyIfOnSameLine reports whether two offsets sit on one line.
func noLonelyIfOnSameLine(text string, from int, to int) bool {
	if from < 0 || to > len(text) || from > to {
		return false
	}
	return !strings.ContainsAny(text[from:to], "\n\r")
}
