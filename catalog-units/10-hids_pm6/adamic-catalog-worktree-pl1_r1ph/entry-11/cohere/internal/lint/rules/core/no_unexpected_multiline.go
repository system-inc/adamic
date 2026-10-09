package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageUnexpectedMultilineFunction = rule.Message{
	Id: "function",
	Description: "A newline sits between this expression and the `(` that follows it, so what reads " +
		"as two statements is one call. Automatic semicolon insertion does not fire before an open " +
		"paren, so the line above is being invoked rather than ended.",
}

var messageUnexpectedMultilineProperty = rule.Message{
	Id: "property",
	Description: "A newline sits between this expression and the `[` that follows it, so what reads " +
		"as an array on its own line is a computed property access on the line above. Automatic " +
		"semicolon insertion does not fire before an open bracket.",
}

var messageUnexpectedMultilineTaggedTemplate = rule.Message{
	Id: "taggedTemplate",
	Description: "A newline sits between this expression and the template literal that follows it, " +
		"so the template is being used as a tag argument rather than standing alone. Automatic " +
		"semicolon insertion does not fire before a backtick.",
}

var messageUnexpectedMultilineDivision = rule.Message{
	Id: "division",
	Description: "A newline sits between the numerator and the `/` that follows it, and what comes " +
		"after reads as a regular expression with flags rather than a division. The two parse " +
		"differently and only one of them is what the line above looks like.",
}

// NoUnexpectedMultiline flags a newline that looks like it ends a statement and does not.
//
//	valid:   (x || y).aFunction()
//	valid:   var a = b;\n(x || y).doSomething()
//	valid:   f(\n(x)\n)
//	valid:   foo\n/ bar /2
//	invalid: var a = b\n(x || y).doSomething()
//	invalid: var a = b\n[a, b, c].forEach(doSomething)
//	invalid: let x = function() {}\n `hello`
//	invalid: foo\n/ bar /gym
//
// # What the rule is actually about
//
// Automatic semicolon insertion does not fire before `(`, `[`, a backtick, or `/`. Each of those
// four continues the previous line instead of starting a new one, so source laid out as two
// statements is one expression. That is the whole subject, and it is why this is a correctness rule
// rather than a formatting preference: no formatter can express the difference, because both
// readings are the same characters.
//
// # Four judgments, one shared test
//
// Every arm asks the same question in the end. Take the expression on the left, find the token that
// opens what follows it, and compare their lines. Upstream shares this as `checkForBreakAfter` and
// so does this port.
//
// # Where our parser makes upstream's paren skip unnecessary
//
// Upstream calls `getTokenAfter(node, astUtils.isNotClosingParenToken)`, and that filter is
// load-bearing for it: its parser folds `ParenthesizedExpression` away, so for
// `var a = (a || b)\n(x || y).doSomething()` the callee node ends at `b` and the very next token is
// the `)` that closes the parenthesis rather than the `(` that opens the call. It has to walk past
// the closers to reach the token the rule is about.
//
// Our parser keeps `KindParenthesizedExpression` as a real node whose End() is after its own
// closing paren, so `Expression.End()` already sits past every closer and `SkipTrivia` from there
// lands on the open token directly. Measured across the four parenthesized cases in upstream's
// corpus, and the skip has nothing to do: the port would produce the same answer with the filter
// and cannot express it without one.
//
// This is the parenthesis hazard running the other way for once. The usual failure is a port seeing
// a node upstream deleted; here the extra node removes a step rather than adding one.
//
// # Why no unwrap loop
//
// Nothing in this rule walks THROUGH an expression to reason about what is inside it. Every arm
// reads a boundary, where one node ends and a token begins, and a parenthesis is part of that
// boundary rather than something to see past. Unwrapping would move the boundary to the wrong place
// and report `var a = (a || b)\n(x).doSomething()` at the paren rather than the call.
//
// # The division arm, and the stale flag list it carries
//
// The shape is a `/` binary whose LEFT child is also a `/` binary, which is how `foo / bar /gym`
// parses. If the token after the second slash is an identifier that is entirely regular expression
// flags AND is adjacent to the slash with no space, the line reads as `foo` followed by a regular
// expression literal rather than as two divisions.
//
// Upstream matches flags with `/^[gimsuy]+$/u`, which predates both `d` (hasIndices, ES2022) and
// `v` (unicodeSets, ES2024). Measured against the installed build at 10.8.1:
//
//	foo\n/bar/g     REPORTS
//	foo\n/bar/s     REPORTS
//	foo\n/bar/d     clean
//	foo\n/bar/v     clean
//
// So `d` and `v` are missed, and the miss is upstream's rather than this port's. Reproduced rather
// than corrected: a port that silently improves on its original is a divergence nobody can see, and
// widening the set here would report inputs the tool being compared against calls clean.
//
// The adjacency test is not decoration either. `foo\n/ bar / g` is clean because the space means
// the `g` cannot be a flag, and `foo\n/ bar /GYM` is clean because the flag set is case sensitive.
// Both are in the corpus.
//
// # What the arms decline, and why each decline is a case rather than an oversight
//
//	a call with no arguments      `b\n()` invokes nothing, so nothing was misread as a statement
//	an optional call or access    `b?.\n(x)` cannot be an accident: the `?.` is explicit
//	a non-computed member         `b\n.x` is unambiguous, a leading dot cannot start a statement
//	a NewExpression               upstream listens on CallExpression alone, and `new b\n(x)` is
//	                              clean against the installed build
//
// # Type arguments, which our parser sees and upstream's default parser cannot
//
// `var a = foo<T>\n(x)` parses here as a call with type arguments, and `Expression.End()` is before
// the `<`. SkipTrivia lands on the `<`, which is on the callee's own line, so the arm declines.
// That is the same answer upstream gives under the TypeScript parser, measured, where
// `getTokenAfter` also returns the `<`. The agreement is worth stating because it is a coincidence
// of two different routes rather than a shared implementation.
var NoUnexpectedMultiline = rule.Rule{
	Name: "no-unexpected-multiline",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		text := ctx.SourceFile.Text()

		// reportIfLineBreaks is upstream's checkForBreakAfter. The token that opens what follows
		// `expression` is the first non-trivia character after it, and the finding is anchored on
		// that single character rather than on either node, which is what upstream's `loc: openParen.loc`
		// does and what its corpus asserts to the column.
		reportIfLineBreaks := func(expression *ast.Node, message rule.Message) {
			if expression == nil {
				return
			}
			openToken := scanner.SkipTrivia(text, expression.End())
			if openToken >= len(text) {
				return
			}
			// The expression's own last line, read from its last character rather than from its
			// End(), which sits one past it.
			//
			// Measured equivalent to reading End() directly, and recorded rather than fixtured
			// because no input can distinguish them: they differ only if the character AT End() is
			// on a later line than the one before it, which requires the expression's final
			// character to be a line terminator, and no production ends an expression with one.
			// Probed over six shapes including calls and accesses whose callee itself ends in `)`
			// or `]`, and the two positions reported the same line every time. The subtraction is
			// kept because it is what the sentence above means, not because it changes an answer.
			if scanner.GetECMALineOfPosition(ctx.SourceFile, expression.End()-1) ==
				scanner.GetECMALineOfPosition(ctx.SourceFile, openToken) {
				return
			}
			ctx.ReportRange(core.NewTextRange(openToken, openToken+1), message)
		}

		return rule.Listeners{
			ast.KindElementAccessExpression: func(node *ast.Node) {
				access := node.AsElementAccessExpression()
				// An optional access spells its own continuation, so the newline is not a surprise.
				if access.QuestionDotToken != nil {
					return
				}
				reportIfLineBreaks(access.Expression, messageUnexpectedMultilineProperty)
			},

			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				if call.QuestionDotToken != nil {
					return
				}
				// A call with no arguments misreads nothing: there is no statement-shaped thing on
				// the next line that was swallowed.
				if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
					return
				}
				reportIfLineBreaks(call.Expression, messageUnexpectedMultilineFunction)
			},

			ast.KindTaggedTemplateExpression: func(node *ast.Node) {
				tagged := node.AsTaggedTemplateExpression()
				if tagged.Template == nil {
					return
				}
				// The template's Pos() sits before its leading trivia, and the trivia is exactly
				// what this arm is measuring across: `aaaa<test>/* comment */`+"`foo`" reports at
				// the backtick, four lines below the comment's start. GetRangeOfTokenAtPosition is
				// what moves the position past the trivia to the backtick itself.
				templateToken := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, tagged.Template.Pos()).Pos()
				if templateToken >= len(text) {
					return
				}
				// The token that ends the tag is what upstream compares against, via
				// `getTokenBefore(quasi)`, and when type arguments are present that token is the
				// closing `>` rather than anything either node's End() names. The type argument
				// LIST ends before its own `>`, so skipping trivia forward from there lands on the
				// `>` itself. That is a forward scan over the same gap upstream walks backwards,
				// and it reaches the same token without a second implementation of comment lexing.
				//
				// `tag<\n generic\n>`+"`x`" is clean for exactly this reason: the tag ends three
				// lines above the backtick while the `>` shares its line.
				lastTokenEnd := tagged.Tag.End()
				if tagged.TypeArguments != nil && tagged.TypeArguments.End() > lastTokenEnd {
					closingAngle := scanner.SkipTrivia(text, tagged.TypeArguments.End())
					if closingAngle >= len(text) {
						return
					}
					lastTokenEnd = closingAngle + 1
				}
				// End() is one past the last character, and a line is read from a character.
				if lastTokenEnd <= 0 {
					return
				}
				if scanner.GetECMALineOfPosition(ctx.SourceFile, lastTokenEnd-1) ==
					scanner.GetECMALineOfPosition(ctx.SourceFile, templateToken) {
					return
				}
				ctx.ReportRange(core.NewTextRange(templateToken, templateToken+1),
					messageUnexpectedMultilineTaggedTemplate)
			},

			ast.KindBinaryExpression: func(node *ast.Node) {
				binary := node.AsBinaryExpression()
				if binary.OperatorToken == nil || binary.OperatorToken.Kind != ast.KindSlashToken {
					return
				}
				// Upstream's selector is `BinaryExpression[operator='/'] > BinaryExpression[operator='/'].left`,
				// so the node it checks is the LEFT child and the second slash is this node's own
				// operator. Anchoring on the parent instead would check the wrong division in
				// `foo\n/ bar / baz /g`, which is clean because the break is before `/bar`.
				left := binary.Left
				if left == nil || left.Kind != ast.KindBinaryExpression {
					return
				}
				leftBinary := left.AsBinaryExpression()
				if leftBinary.OperatorToken == nil || leftBinary.OperatorToken.Kind != ast.KindSlashToken {
					return
				}
				// The token after the second slash has to be an identifier made only of regular
				// expression flags AND adjacent to the slash, or the source does not read as a
				// regular expression literal at all.
				//
				// Read as a TOKEN rather than as the right operand's node. Upstream calls
				// `getTokenAfter(secondSlash)`, and the two differ whenever the flags are followed
				// by anything: in `foo\n/ bar /g.test(baz)` the right operand is the whole
				// `g.test(baz)` call while the token is the bare `g`. Three of upstream's five
				// division cases have that shape, and a node-kind test silently declines all of
				// them.
				flagsStart := binary.OperatorToken.End()
				flagsEnd := flagsStart
				for flagsEnd < len(text) && noUnexpectedMultilineIsFlagCharacter(text[flagsEnd]) {
					flagsEnd++
				}
				if flagsEnd == flagsStart {
					return
				}
				// A longer identifier that merely STARTS with flag characters is not a flag run:
				// `foo\n/ bar /gymnasium` reads as a division by an ordinary name. Upstream gets
				// this from `getTokenAfter` returning the whole identifier and then anchoring the
				// pattern, so the boundary has to be a non-identifier character.
				if flagsEnd < len(text) && scanner.IsIdentifierPart(rune(text[flagsEnd])) {
					return
				}
				reportIfLineBreaks(leftBinary.Left, messageUnexpectedMultilineDivision)
			},
		}
	},
}

// noUnexpectedMultilineIsFlagCharacter reproduces the character class in upstream's
// `/^[gimsuy]+$/u`.
//
// Deliberately missing `d` and `v`, which are real regular expression flags upstream's pattern
// predates. See the rule doc comment for the measurement against the installed build.
func noUnexpectedMultilineIsFlagCharacter(character byte) bool {
	switch character {
	case 'g', 'i', 'm', 's', 'u', 'y':
		return true
	}
	return false
}
