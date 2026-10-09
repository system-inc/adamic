package nexus

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/ecmascript/jsx"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

const correctnessNoIdenticalBranchesId = "identicalBranches"

// correctnessNoIdenticalBranchesText is the rule's message, whose wording lives in
// `policy/messages/correctness-no-identical-branches.json`.
var correctnessNoIdenticalBranchesText = policy.MessageOf("nexus/correctness-no-identical-branches", correctnessNoIdenticalBranchesId)

func correctnessNoIdenticalBranchesMessage() rule.Message {
	return rule.Message{Id: correctnessNoIdenticalBranchesId, Description: correctnessNoIdenticalBranchesText.Render(nil)}
}

// CorrectnessNoIdenticalBranches reports a conditional whose every branch is the same code, so the
// condition is silently ignored.
//
//	invalid: { name: 'model_mode', value: highQuality ? 'pro' : 'pro' }
//	invalid: if (ready) { start(); } else { start(); }
//	invalid: if (a) { start(); } else if (b) { start(); } else { start(); }
//	invalid: if (a) { stop(); } else if (b) { start(); } else { start(); }   (b is ignored)
//	valid:   { name: 'model_mode', value: highQuality ? 'pro' : 'std' }
//	valid:   if (a) { start(); } else if (b) { start(); } else { stop(); }
//	valid:   if (a) { start(); } else if (b) { start(); }
//	valid:   typeof value === 'string' ? format(value) : format(value)   (the branches pick overloads)
//
// # Where it came from
//
// `modules/kling/KlingApi.ts:156` and `:187` in ahra, `highQuality ? 'pro' : 'pro'`: a quality flag
// that silently did nothing in two request builders. Found by both research passes of the new-rules
// sweep (`#tevhg3f`), built as task `#052jffy`.
//
// # Equality is structural
//
// Two branches are the same when they are the same tree written with the same tokens: whitespace,
// line breaks and comments do not count, and nothing else is forgiven. `'pro'` and `"pro"` are
// different tokens and stay different, `[]` and `[ ]` are the same. Literals, identifiers and
// template text are compared exactly as written, so a regular expression, a string or JSX text that
// happens to contain `/* */` or doubled spaces is never collapsed by the comment and whitespace
// rule. Every way this can be wrong is a missed finding, never a false one.
//
// At the top of a branch only, a braced block holding one statement is the same as that statement
// unbraced, and parentheses around a whole ternary branch are dropped, because neither changes what
// runs.
//
// # Same tokens can still be different code, when the condition narrows a type
//
// `typeof value === 'string' ? format(value) : format(value)` is two token-identical branches that
// do different work at compile time: with `format` overloaded for `string` and for `number` and no
// signature taking the union, each branch resolves its own overload and the merged call would not
// type-check. The condition is load-bearing, so reporting it would be a false positive. The rule
// therefore declares the type checker and, after the branches match token for token, walks them in
// parallel asking whether one merged copy would type-check the same way: every call-like node must
// resolve the same signature with arguments of the same types, and every operand of an operator
// that rejects unions (arithmetic, bitwise, relational, `in`, `instanceof`) must have the same type.
// Any difference leaves the conditional alone. `correctnessNoIdenticalBranchesSameTypes` gives the
// reasoning for why those positions and no others.
//
// The first draft compared the type of every identifier, `this`, property and element access and
// call, and on the three trees it silenced three conditionals whose condition really is ignored:
// `node.type === ImportDeclaration ? node.source : node.source` (a discriminated union narrowed,
// read through a property, and merging compiles) and `typeof parsed.expiresAt === 'number' ? new
// Date(parsed.expiresAt) : new Date(parsed.expiresAt)` twice (the field is a `string`, so the first
// branch narrows it to `never`). Both shapes are fixtures. Without a checker the rule registers
// nothing rather than reporting without the guard.
//
// # What it reports, and what it leaves alone, by decision
//
//   - A ternary (`ConditionalExpression`) whose two branches match. Type-level conditionals
//     (`T extends U ? X : X`) are not reported: they distribute over a union, so identical branches
//     there still do work.
//   - An `if` with an `else` whose two branches match. An `if` with no `else` has no second branch.
//   - **An `else if` chain is judged from its last `if`**, the one with the final `else`: when its
//     two branches match, that condition is ignored. The finding then widens up the chain over
//     every earlier `if` whose branch is the same too, so `if (a) X else if (b) X else X` is one
//     finding on the whole chain, and `if (a) Y else if (b) X else X` is one finding on `if (b)`.
//     Two equal branches with a different one after them (`if (a) X else if (b) X else Y`) are
//     **not** reported: that is `if (a || b) X else Y` written long, both conditions still decide
//     between X and Y, and whether to merge them is a style choice. Equal branches that are not
//     adjacent (`if (a) X else if (b) Y else if (c) X`) cannot even be merged without reordering.
//     A chain with no final `else` is never reported, because when no condition holds nothing
//     runs, so every condition decides something.
//   - The same widening applies to a ternary chain through its false branch:
//     `a ? x : b ? x : x` is one finding on the whole expression.
//   - **Empty branches are not reported.** `if (a) {} else {}`, or two branches holding only
//     comments, are already `no-empty`'s finding when empty and, when commented, carry their
//     intent in the comments, the one thing this comparison ignores. Reporting them would be
//     reporting branches the author told apart in the only way an empty branch can be.
//   - **Branches that are only `return;` (or `break;`, `continue;`, `throw x;`) are reported.**
//     `if (a) return; else return;` ignores its condition exactly as the ternary does, and
//     nothing about a bare exit makes the two branches different.
//   - **`switch` is not covered.** Two cases with the same tokens can still run differently,
//     since one may end in `break` and fall nowhere while the other falls through into the next
//     case, and the default case is usually the one written without the `break`, so a
//     token-equality check would miss the common shape and a normalized one is a semantic
//     judgment about fallthrough rather than a structural one. Measured on the three trees as
//     zero candidates; see the document beside this file.
//
// # No fix
//
// Which branch was meant is the author's call: the KlingApi site needed `'std'` on one side, and
// deleting the conditional would have kept the bug.
var CorrectnessNoIdenticalBranches = rule.Rule{
	Name:             "nexus/correctness-no-identical-branches",
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.TypeChecker == nil {
			// The narrowing guard needs types. Reporting without it would report the overload
			// pattern above, so with no checker the rule declines the file.
			return nil
		}
		return rule.Listeners{
			ast.KindConditionalExpression: func(node *ast.Node) {
				conditional := node.AsConditionalExpression()
				whenTrue := ast.SkipParentheses(conditional.WhenTrue)
				if !correctnessNoIdenticalBranchesSame(ctx, whenTrue, ast.SkipParentheses(conditional.WhenFalse)) {
					return
				}
				anchor := node
				for {
					parent := correctnessNoIdenticalBranchesParentSkippingParentheses(anchor)
					if parent == nil || parent.Kind != ast.KindConditionalExpression {
						break
					}
					parentConditional := parent.AsConditionalExpression()
					if ast.SkipParentheses(parentConditional.WhenFalse) != anchor {
						break
					}
					if !correctnessNoIdenticalBranchesSame(ctx, ast.SkipParentheses(parentConditional.WhenTrue), whenTrue) {
						break
					}
					anchor = parent
				}
				ctx.ReportNode(anchor, correctnessNoIdenticalBranchesMessage())
			},
			ast.KindIfStatement: func(node *ast.Node) {
				ifStatement := node.AsIfStatement()
				if ifStatement.ElseStatement == nil {
					return
				}
				// An `if` whose `else` is another `if` compares its branch against that whole `if`,
				// which differs from it except in the rare `if (a) { if (c) p(); else q(); } else if
				// (c) p(); else q();`, where `a` really is ignored. The usual chain is judged at its
				// last `if` and reaches back up through the widening below.
				thenBranch := correctnessNoIdenticalBranchesUnwrapBlock(ifStatement.ThenStatement)
				elseBranch := correctnessNoIdenticalBranchesUnwrapBlock(ifStatement.ElseStatement)
				if correctnessNoIdenticalBranchesIsEmpty(thenBranch) || correctnessNoIdenticalBranchesIsEmpty(elseBranch) {
					return
				}
				if !correctnessNoIdenticalBranchesSame(ctx, thenBranch, elseBranch) {
					return
				}
				anchor := node
				for {
					parent := anchor.Parent
					if parent == nil || parent.Kind != ast.KindIfStatement || parent.AsIfStatement().ElseStatement != anchor {
						break
					}
					if !correctnessNoIdenticalBranchesSame(ctx, correctnessNoIdenticalBranchesUnwrapBlock(parent.AsIfStatement().ThenStatement), thenBranch) {
						break
					}
					anchor = parent
				}
				ctx.ReportNode(anchor, correctnessNoIdenticalBranchesMessage())
			},
		}
	},
}

// correctnessNoIdenticalBranchesParentSkippingParentheses is the node a conditional is a branch of,
// looking out through any parentheses around it.
func correctnessNoIdenticalBranchesParentSkippingParentheses(node *ast.Node) *ast.Node {
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	return parent
}

// correctnessNoIdenticalBranchesUnwrapBlock reads `{ statement; }` as `statement;`, so a braced
// branch and an unbraced one holding the same single statement compare equal.
func correctnessNoIdenticalBranchesUnwrapBlock(statement *ast.Node) *ast.Node {
	if statement != nil && statement.Kind == ast.KindBlock {
		statements := statement.AsBlock().Statements
		if statements != nil && len(statements.Nodes) == 1 {
			return statements.Nodes[0]
		}
	}
	return statement
}

// correctnessNoIdenticalBranchesIsEmpty says whether a branch runs nothing: `{}` (comments or not)
// or a lone `;`.
func correctnessNoIdenticalBranchesIsEmpty(statement *ast.Node) bool {
	switch statement.Kind {
	case ast.KindEmptyStatement:
		return true
	case ast.KindBlock:
		statements := statement.AsBlock().Statements
		return statements == nil || len(statements.Nodes) == 0
	}
	return false
}

// correctnessNoIdenticalBranchesSame says whether two branches are the same code: the same tokens,
// and, where the condition could have narrowed a type the branch reads, the same types.
func correctnessNoIdenticalBranchesSame(ctx rule.Context, left *ast.Node, right *ast.Node) bool {
	if left == nil || right == nil {
		return false
	}
	if !correctnessNoIdenticalBranchesSameTokens(ctx.SourceFile, left, right) {
		return false
	}
	return correctnessNoIdenticalBranchesSameTypes(ctx, left, right)
}

// correctnessNoIdenticalBranchesChildren lists a node's children in source order.
func correctnessNoIdenticalBranchesChildren(node *ast.Node) []*ast.Node {
	var children []*ast.Node
	node.ForEachChild(func(child *ast.Node) bool {
		children = append(children, child)
		return false
	})
	return children
}

// correctnessNoIdenticalBranchesSameTokens compares two subtrees node by node: the same kind, the
// same number of children, the same punctuation in every gap between children, and the same text
// in every leaf.
//
// The kind recovers operators the child walk never visits (a prefix unary's operator is a field,
// not a child), and the gaps recover the rest: the dot in `a.b`, the `?.` that makes it a different
// expression, `const` against `let`. Gaps are scanned for tokens, so whitespace and comments drop
// out; leaves are compared as written, so a literal is never re-scanned by a scanner that cannot
// tell a regular expression from a division or JSX text from code. This is the shape of
// `tokenSignature` in `core/token_signature.go`, written here as a parallel walk that stops at the
// first difference, because a rule may not import another rule package.
func correctnessNoIdenticalBranchesSameTokens(sourceFile *ast.SourceFile, left *ast.Node, right *ast.Node) bool {
	if left.Kind != right.Kind {
		return false
	}
	leftChildren := correctnessNoIdenticalBranchesChildren(left)
	rightChildren := correctnessNoIdenticalBranchesChildren(right)
	if len(leftChildren) != len(rightChildren) {
		return false
	}
	leftRange := rule.TokenRange(sourceFile, left)
	rightRange := rule.TokenRange(sourceFile, right)
	if len(leftChildren) == 0 {
		if correctnessNoIdenticalBranchesIsPunctuationLeaf(left) {
			return correctnessNoIdenticalBranchesSameGap(sourceFile, leftRange.Pos(), leftRange.End(), rightRange.Pos(), rightRange.End())
		}
		text := sourceFile.Text()
		return text[leftRange.Pos():leftRange.End()] == text[rightRange.Pos():rightRange.End()]
	}
	leftCursor := leftRange.Pos()
	rightCursor := rightRange.Pos()
	for index := range leftChildren {
		leftChildRange := rule.TokenRange(sourceFile, leftChildren[index])
		rightChildRange := rule.TokenRange(sourceFile, rightChildren[index])
		if !correctnessNoIdenticalBranchesSameGap(sourceFile, leftCursor, leftChildRange.Pos(), rightCursor, rightChildRange.Pos()) {
			return false
		}
		if !correctnessNoIdenticalBranchesSameTokens(sourceFile, leftChildren[index], rightChildren[index]) {
			return false
		}
		leftCursor = leftChildRange.End()
		rightCursor = rightChildRange.End()
	}
	return correctnessNoIdenticalBranchesSameGap(sourceFile, leftCursor, leftRange.End(), rightCursor, rightRange.End())
}

// correctnessNoIdenticalBranchesIsPunctuationLeaf names the childless nodes made only of
// punctuation, `[ ]`, `{ }` and `;`, whose inner whitespace is trivia. Every other leaf is compared
// as written, which can only miss a match, never invent one.
func correctnessNoIdenticalBranchesIsPunctuationLeaf(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindArrayLiteralExpression, ast.KindObjectLiteralExpression, ast.KindBlock, ast.KindEmptyStatement:
		return true
	}
	return false
}

// correctnessNoIdenticalBranchesSameGap says whether two source ranges hold the same tokens,
// ignoring the whitespace and comments between them.
func correctnessNoIdenticalBranchesSameGap(sourceFile *ast.SourceFile, leftStart int, leftEnd int, rightStart int, rightEnd int) bool {
	leftTokens := correctnessNoIdenticalBranchesTokens(sourceFile, leftStart, leftEnd)
	rightTokens := correctnessNoIdenticalBranchesTokens(sourceFile, rightStart, rightEnd)
	if len(leftTokens) != len(rightTokens) {
		return false
	}
	for index := range leftTokens {
		if leftTokens[index] != rightTokens[index] {
			return false
		}
	}
	return true
}

// correctnessNoIdenticalBranchesTokens scans the tokens in a half-open range. The scanner returned
// by GetScannerForSourceFile has already scanned the token at the start, so its current token is
// read before advancing; scanning first drops the first token of every range.
func correctnessNoIdenticalBranchesTokens(sourceFile *ast.SourceFile, start int, end int) []string {
	if start >= end {
		return nil
	}
	var tokens []string
	tokenScanner := scanner.GetScannerForSourceFile(sourceFile, start)
	for tokenScanner.Token() != ast.KindEndOfFile && tokenScanner.TokenStart() < end {
		tokens = append(tokens, tokenScanner.TokenText())
		tokenScanner.Scan()
	}
	return tokens
}

// correctnessNoIdenticalBranchesSameTypes walks two token-identical branches in parallel and says
// whether merging them into one copy would type-check the same way. It does not ask whether every
// expression has the same type on both sides, because the condition narrowing a reference is the
// ordinary case and most of it is harmless: a branch compiles against each narrowed type, and
// assignment, property access, element access and an argument to a single signature all accept a
// union exactly when they accept each of its members, so the merged copy compiles too.
//
// Narrowing changes what the same tokens mean in two places only, and those are the two it checks:
//
//   - **Which signature a call resolves to.** An overloaded function, a method on a union of
//     receivers (`Array.isArray(v) ? v.slice(0) : v.slice(0)`), a constructor or a JSX component
//     can resolve to a different signature per branch, and the merged call may resolve to none
//     (`format(v)` above does not; `v.slice(0)` happens to, and the rule does not try to tell which).
//     Every call-like node compares its resolved signature. When both branches resolve the same
//     one, the merged call resolves it too: an earlier overload accepting the union would have
//     accepted each member, and so would have been chosen in each branch. Only a spread argument
//     can still differ, a union of tuples that cannot be spread, so it compares its type as well.
//     An intrinsic JSX element (`<p>`) gets a fresh signature per element from the checker, so it
//     is skipped: its props are fixed by its tag, which the token comparison already matched.
//   - **A binary operator that restricts its operand types.** `(number | bigint) * x`,
//     `(string | number) + 1` and `(string | number) < 3` do not compile where each member alone
//     does (checked with tsc), so arithmetic, bitwise and relational operators, their compound
//     assignments, and `in` and `instanceof` compare their operands' types. The unary operators do
//     not: `-`, `~`, `++` and `--` all accept `number | bigint` (also checked), and every type
//     they reject alone would already fail in one of the branches.
//
// A type of `never` on one side matches anything on the other: the condition narrowed that
// reference to nothing in one branch, so the other branch already sees its whole type, and the
// merged copy is exactly that branch as the checker sees it.
func correctnessNoIdenticalBranchesSameTypes(ctx rule.Context, left *ast.Node, right *ast.Node) bool {
	switch left.Kind {
	case ast.KindCallExpression, ast.KindNewExpression, ast.KindTaggedTemplateExpression,
		ast.KindJsxOpeningElement, ast.KindJsxSelfClosingElement:
		if !correctnessNoIdenticalBranchesIsIntrinsicJsx(left) &&
			ctx.TypeChecker.GetResolvedSignature(left) != ctx.TypeChecker.GetResolvedSignature(right) {
			return false
		}
		for index, leftArgument := range correctnessNoIdenticalBranchesArguments(left) {
			if leftArgument.Kind == ast.KindSpreadElement &&
				!correctnessNoIdenticalBranchesSameTypeAt(ctx, leftArgument.AsSpreadElement().Expression,
					correctnessNoIdenticalBranchesArguments(right)[index].AsSpreadElement().Expression) {
				return false
			}
		}
	case ast.KindBinaryExpression:
		if correctnessNoIdenticalBranchesTypedOperators[left.AsBinaryExpression().OperatorToken.Kind] &&
			(!correctnessNoIdenticalBranchesSameTypeAt(ctx, left.AsBinaryExpression().Left, right.AsBinaryExpression().Left) ||
				!correctnessNoIdenticalBranchesSameTypeAt(ctx, left.AsBinaryExpression().Right, right.AsBinaryExpression().Right)) {
			return false
		}
	}
	leftChildren := correctnessNoIdenticalBranchesChildren(left)
	rightChildren := correctnessNoIdenticalBranchesChildren(right)
	for index := range leftChildren {
		if !correctnessNoIdenticalBranchesSameTypes(ctx, leftChildren[index], rightChildren[index]) {
			return false
		}
	}
	return true
}

// correctnessNoIdenticalBranchesTypedOperators are the binary operators that can reject a union
// whose members they each accept. Equality, the logical operators, `??`, `,` and plain assignment
// accept any union of what they accept, so narrowing cannot change whether they compile.
var correctnessNoIdenticalBranchesTypedOperators = map[ast.Kind]bool{
	ast.KindPlusToken: true, ast.KindMinusToken: true, ast.KindAsteriskToken: true,
	ast.KindAsteriskAsteriskToken: true, ast.KindSlashToken: true, ast.KindPercentToken: true,
	ast.KindLessThanLessThanToken: true, ast.KindGreaterThanGreaterThanToken: true,
	ast.KindGreaterThanGreaterThanGreaterThanToken: true, ast.KindAmpersandToken: true,
	ast.KindBarToken: true, ast.KindCaretToken: true,
	ast.KindLessThanToken: true, ast.KindGreaterThanToken: true,
	ast.KindLessThanEqualsToken: true, ast.KindGreaterThanEqualsToken: true,
	ast.KindInKeyword: true, ast.KindInstanceOfKeyword: true,
	ast.KindPlusEqualsToken: true, ast.KindMinusEqualsToken: true, ast.KindAsteriskEqualsToken: true,
	ast.KindAsteriskAsteriskEqualsToken: true, ast.KindSlashEqualsToken: true,
	ast.KindPercentEqualsToken: true, ast.KindLessThanLessThanEqualsToken: true,
	ast.KindGreaterThanGreaterThanEqualsToken: true, ast.KindGreaterThanGreaterThanGreaterThanEqualsToken: true,
	ast.KindAmpersandEqualsToken: true, ast.KindBarEqualsToken: true, ast.KindCaretEqualsToken: true,
}

// correctnessNoIdenticalBranchesSameTypeAt compares the checker's type for one pair of matching
// expressions, with `never` on either side matching anything (see above).
func correctnessNoIdenticalBranchesSameTypeAt(ctx rule.Context, left *ast.Node, right *ast.Node) bool {
	leftType := ctx.TypeChecker.GetTypeAtLocation(left)
	rightType := ctx.TypeChecker.GetTypeAtLocation(right)
	if leftType == rightType {
		return true
	}
	return type_checking.IsTypeFlagSet(leftType, checker.TypeFlagsNever) || type_checking.IsTypeFlagSet(rightType, checker.TypeFlagsNever)
}

// correctnessNoIdenticalBranchesArguments lists a call's or a construction's arguments.
func correctnessNoIdenticalBranchesArguments(node *ast.Node) []*ast.Node {
	switch node.Kind {
	case ast.KindCallExpression:
		return node.AsCallExpression().Arguments.Nodes
	case ast.KindNewExpression:
		if node.AsNewExpression().Arguments != nil {
			return node.AsNewExpression().Arguments.Nodes
		}
	}
	return nil
}

// correctnessNoIdenticalBranchesIsIntrinsicJsx says whether a JSX element names a host element
// (`<p>`, `<my-element>`, `<svg:rect>`) rather than a component, the way the checker decides it.
func correctnessNoIdenticalBranchesIsIntrinsicJsx(node *ast.Node) bool {
	tagName, _ := jsx.ElementParts(node)
	if tagName == nil {
		return false
	}
	return tagName.Kind == ast.KindJsxNamespacedName ||
		(tagName.Kind == ast.KindIdentifier && scanner.IsIntrinsicJsxName(tagName.Text()))
}
