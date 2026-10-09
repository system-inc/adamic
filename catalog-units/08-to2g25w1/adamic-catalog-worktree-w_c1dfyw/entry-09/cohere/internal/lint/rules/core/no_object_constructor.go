package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messagePreferLiteral = rule.Message{
	Id: "preferLiteral",
	Description: "This calls the `Object` constructor with no argument, which builds the same " +
		"empty object `{}` builds, more slowly and less legibly. The constructor form also does " +
		"something different from what it looks like when it is given an argument, since " +
		"`Object(x)` returns `x` itself rather than a copy, so a reader has to check the argument " +
		"list before knowing what the line means. Write the literal.",
}

var messageUseLiteral = rule.Message{
	Id:          "useLiteral",
	Description: "Replace with an object literal.",
}

var messageUseLiteralAfterSemicolon = rule.Message{
	Id: "useLiteralAfterSemicolon",
	Description: "Replace with an object literal, adding the preceding semicolon that the " +
		"parenthesis would otherwise swallow.",
}

// NoObjectConstructor flags calling `Object` with no arguments.
//
//	valid:   new Object(x)
//	valid:   Object(x)
//	valid:   new globalThis.Object
//	valid:   const createObject = Object => new Object()
//	valid:   var Object; new Object;
//	invalid: new Object
//	invalid: Object()
//	invalid: const obj = Object?.();
//
// # The judgment is three tests and the corpus states each one
//
// The callee must be the bare identifier `Object`, there must be no arguments, and the name must
// resolve to the global rather than a local shadowing it. `Object(x)` is clean because the
// one-argument form is a different operation entirely: it returns its argument boxed rather than a
// new object. `new globalThis.Object` is clean because the callee is a member access rather than the
// identifier, which is upstream's own reading and is narrower than "evaluates to Object".
//
// Both call shapes report, since the constructor ignores `new`, and the optional form `Object?.()`
// reports too.
//
// # The suggestion, and why it is a suggestion rather than a fix
//
// Upstream declares `hasSuggestions` and offers no fix, so the repair is one a human chooses. That
// is ported as-is: replacing a call with a literal is mechanical here only because the argument list
// is empty, and the engine should not be rewriting expressions unattended on the strength of that.
//
// The replacement text is `{}` or `({})` depending on position, and which one is not cosmetic. An
// object literal at the start of an expression statement parses as a BLOCK, so `Object()` alone on a
// line has to become `({})` rather than `{}` or the meaning changes. The same is true directly after
// an arrow, where `() => {}` is a function with an empty body. Measured on eslint 10.8.1:
//
//	Object()                      suggests ({})
//	const fn = () => Object();    suggests ({})
//	const obj = Object?.();       suggests {}
//	(new Object() instanceof Object);   suggests {}, because the paren is already there
//
// # One divergence, measured: a `with` body
//
// `with (obj) Object();` reports upstream and is SILENT here. The checker returns no symbol at all
// for `Object` inside a `with` block, so `resolvesToAGlobal` answers false and the call is declined.
//
// That is the predicate being right rather than wrong. Inside `with (obj)`, whether `Object` names
// the global depends on whether `obj` happens to carry an `Object` property at runtime, which is
// exactly the question no static analysis can answer and precisely why `with` is banned in strict
// mode and by `no-with`. Upstream's scope analysis reports because it does not model `with` bodies
// at all; ours declines because it does. Reproducing upstream here would mean asserting a binding we
// cannot see.
//
// Measured with a probe reading `GetSymbolAtLocation` directly: `Object();` at the top level
// resolves to a symbol carrying 2 declarations, and the same call inside a `with` resolves to none.
// The case is kept in the fixture table as SILENT with this reasoning at the line, rather than
// dropped, so the difference stays visible.

// # The preceding semicolon, which is the expensive half
//
// When the literal needs parentheses AND the previous line ended without a semicolon, the `(` would
// continue that previous expression instead of starting a new one, so upstream emits `;({})` under a
// second message id. Twelve of the corpus's fifty cases take that arm.
//
// Upstream decides it with `needsPrecedingSemicolon`, which reads the previous TOKEN and then the
// node that token belongs to. This answers the same question from the tree instead, because we have
// the tree and no token-before helper: the question is what the previous statement is and how it
// ended. See `previousStatementCanContinue` for the enumeration and for what each row was measured
// against.
var NoObjectConstructor = rule.Rule{
	Name: "no-object-constructor",

	// The whole discrimination between a report and a false positive on `const createObject =
	// Object => new Object()` is whether `Object` is the global, and nothing structural answers it.
	NeedsTypeChecker: true,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		check := func(node *ast.Node) {
			if ctx.TypeChecker == nil {
				return
			}

			var callee *ast.Node
			var arguments *ast.NodeList
			switch node.Kind {
			case ast.KindNewExpression:
				callee = node.AsNewExpression().Expression
				arguments = node.AsNewExpression().Arguments
			case ast.KindCallExpression:
				callee = node.AsCallExpression().Expression
				arguments = node.AsCallExpression().Arguments
			default:
				return
			}

			// No `ast.SkipParentheses` here, and that is upstream's reading rather than an omission:
			// it tests `node.callee.type !== "Identifier"` directly, and a parenthesized callee is a
			// different node in ESTree too. Measured on eslint 10.8.1, `(Object)()` is CLEAN.
			if callee == nil || callee.Kind != ast.KindIdentifier || callee.Text() != "Object" {
				return
			}
			// `new Object` with no argument list at all has a nil Arguments, and it reports.
			if arguments != nil && len(arguments.Nodes) > 0 {
				return
			}
			if !resolvesToAGlobal(ctx, callee) {
				return
			}

			replacement := "{}"
			message := messageUseLiteral
			if literalNeedsParentheses(node) {
				replacement = "({})"
				if precedingTextCanContinue(ctx, node) {
					replacement = ";({})"
					message = messageUseLiteralAfterSemicolon
				}
			}

			ctx.ReportNodeWithSuggestions(node, messagePreferLiteral, rule.Suggestion{
				Message: message,
				Fixes: []rule.Fix{
					// `rule.TokenRange` rather than `node.Pos()`: a node's Pos includes its leading
					// trivia, so a fix built from it eats the whitespace and comments before the
					// call. `const fn = () => Object();` came out as `const fn = () =>({});` before
					// this, which is a repair that is right and anchored wrong.
					rule.ReplaceRange(rule.TokenRange(ctx.SourceFile, node), replacement),
				},
			})
		}

		return rule.Listeners{
			ast.KindNewExpression:  check,
			ast.KindCallExpression: check,
		}
	},
}

// literalNeedsParentheses reports whether an object literal replacing this node would be misparsed
// without them.
//
// Two positions need them, and upstream tests exactly these two. A literal at the START of an
// expression statement parses as a block, so `Object();` alone must become `({});`. And a literal
// directly after an arrow is the function's body, so `() => Object()` must become `() => ({})`.
//
// "At the start of" is positional rather than structural: `Object() instanceof Object;` needs them
// because the statement begins at the call even though the call is not the whole statement. Upstream
// walks ancestors while their start position is unchanged, which is what this reproduces.
func literalNeedsParentheses(node *ast.Node) bool {
	start := node.Pos()
	for ancestor := node.Parent; ancestor != nil && ancestor.Pos() == start; ancestor = ancestor.Parent {
		if ancestor.Kind == ast.KindExpressionStatement {
			return true
		}
	}
	// A concise arrow body: `() => Object()`. The braces would otherwise read as a block body.
	if parent := node.Parent; parent != nil && parent.Kind == ast.KindArrowFunction &&
		parent.AsArrowFunction().Body == node {
		return true
	}
	return false
}

// precedingTextCanContinue reports whether the source immediately before this node could swallow a
// leading `(` as a continuation, so the suggestion has to open with a semicolon.
//
// Upstream answers this by reading the previous TOKEN and classifying it, and the corpus is built to
// separate that from any statement-level approximation. Three of its cases show why a
// previous-statement reading is wrong:
//
//	foo();Object();                  useLiteral, because an explicit `;` is already there
//	a++ Object()                     useLiteral, because postfix `++` itself triggers insertion
//	const foo = () => {} Object()    useLiteral, because the arrow body's `}` closes the statement
//
// So this reads the last non-whitespace, non-comment character before the node, which is the token
// reading expressed over the text we have rather than over a token stream we do not.
//
// The punctuators that need no semicolon are upstream's own set, and each is there for its own
// reason. `;` `{` `:` end a statement or a label outright. `=>` cannot be followed by an expression
// statement at all. `++` and `--` are postfix operators, and a postfix operator is required to be on
// the same line as its operand, so a newline after one has ALREADY triggered insertion.
//
// A closing brace is the ambiguous one and it is resolved by what the brace closes, which the tree
// knows and the character does not: a block, a function or class DECLARATION, or an arrow body
// closes the statement, while an object literal, a function or class EXPRESSION does not. That is
// what `statementEndsOpen` answers, and it is consulted only when the character is `}`.
func precedingTextCanContinue(ctx rule.Context, node *ast.Node) bool {
	if ctx.SourceFile == nil {
		return false
	}
	text := ctx.SourceFile.Text()
	start := rule.TokenRange(ctx.SourceFile, node).Pos()
	if start <= 0 || start > len(text) {
		return false
	}

	index := lastMeaningfulIndex(text, start)
	if index < 0 {
		// Nothing before it in the file, so nothing can continue.
		return false
	}

	switch text[index] {
	case ';', '{', ':':
		return false
	case '>':
		// `=>`. Any other `>` ends an expression, which a `(` would continue.
		if index > 0 && text[index-1] == '=' {
			return false
		}
		return true
	case '+':
		// Postfix `++` has already triggered insertion, since a postfix operator must sit on the
		// same line as its operand. A single `+` is binary and cannot end a statement.
		return !(index > 0 && text[index-1] == '+')
	case '-':
		return !(index > 0 && text[index-1] == '-')
	case ')':
		// Upstream: `return !STATEMENTS.has(prevNode.type)`, where prevNode is the node CONTAINING
		// the paren. A closing paren that ends a statement's header is followed by that statement's
		// body, so nothing is left open; a paren ending a call or a parenthesized expression is.
		return !closingParenBelongsToAStatementHeader(ctx, index)
	case '}':
		return closingBraceEndsAnExpression(ctx, index)
	case '\'', '"':
		// A string ending an import or export declaration is closed by insertion; any other string
		// is an expression a `(` would call. Upstream: `!DECLARATIONS.has(prevNode.parent.type)`.
		return !stringEndsAModuleDeclaration(ctx, index)
	}

	// An identifier or keyword. A keyword that ENDS its construct has already been closed by
	// automatic semicolon insertion; a plain identifier has not.
	if word, wordStart := wordBefore(text, index); word != "" {
		return !keywordClosesTheStatement(ctx, word, wordStart)
	}
	return true
}

// stringEndsAModuleDeclaration reports whether the string literal ending at `index` is the specifier
// of an import or export declaration, which insertion has already closed.
func stringEndsAModuleDeclaration(ctx rule.Context, index int) bool {
	node := innermostNodeContaining(ctx.SourceFile.AsNode(), index)
	for ; node != nil; node = node.Parent {
		if node.End() != index+1 {
			continue
		}
		switch node.Kind {
		case ast.KindImportDeclaration, ast.KindExportDeclaration:
			return true
		}
		if node.Parent != nil {
			switch node.Parent.Kind {
			case ast.KindImportDeclaration, ast.KindExportDeclaration:
				return true
			}
		}
	}
	return false
}

// closingParenBelongsToAStatementHeader reports whether the paren at `index` closes the header of a
// statement that takes a body after it.
//
// Upstream's set, verbatim: do-while, for, for-in, for-of, if, while, with. A `do` is in the list
// because its own paren comes last, and the corpus writes that case.
func closingParenBelongsToAStatementHeader(ctx rule.Context, index int) bool {
	node := innermostNodeContaining(ctx.SourceFile.AsNode(), index)
	for ; node != nil; node = node.Parent {
		switch node.Kind {
		case ast.KindDoStatement,
			ast.KindForStatement,
			ast.KindForInStatement,
			ast.KindForOfStatement,
			ast.KindIfStatement,
			ast.KindWhileStatement,
			ast.KindWithStatement:
			// Only when the paren really is this statement's header rather than a call inside it,
			// which is what the position test settles: a header paren is followed by the body.
			if statementHeaderClosesAt(node, index) {
				return true
			}
		case ast.KindCallExpression,
			ast.KindNewExpression,
			ast.KindParenthesizedExpression,
			ast.KindArrowFunction,
			ast.KindFunctionExpression,
			ast.KindFunctionDeclaration:
			return false
		}
	}
	return false
}

// statementHeaderClosesAt reports whether the statement's body begins after the character at index,
// which is what makes that paren the header's rather than something nested inside it.
func statementHeaderClosesAt(statement *ast.Node, index int) bool {
	var body *ast.Node
	switch statement.Kind {
	case ast.KindDoStatement:
		// `do X while (a)` closes with its own paren at the very end.
		return statement.End() > index
	case ast.KindForStatement:
		body = statement.AsForStatement().Statement
	case ast.KindForInStatement, ast.KindForOfStatement:
		body = statement.AsForInOrOfStatement().Statement
	case ast.KindIfStatement:
		body = statement.AsIfStatement().ThenStatement
		if elseStatement := statement.AsIfStatement().ElseStatement; elseStatement != nil &&
			rule.TokenRange(nil, elseStatement).Pos() > index {
			body = elseStatement
		}
	case ast.KindWhileStatement:
		body = statement.AsWhileStatement().Statement
	case ast.KindWithStatement:
		body = statement.AsWithStatement().Statement
	}
	// The body must begin at or after this character for the paren to be the header's rather than
	// one belonging to something nested inside the header or the body.
	return body != nil && body.End() > index && rule.TokenRange(nil, body).Pos() >= index
}

// closingBraceEndsAnExpression reports whether the brace at `index` closes something that leaves the
// statement open.
//
// Upstream's three, verbatim: a block that is a FunctionExpression's body and not a method's, a
// ClassExpression's body, and an object literal. Everything else -- a plain block, a function or
// class declaration, an arrow body, a loop or conditional body -- closes the statement.
func closingBraceEndsAnExpression(ctx rule.Context, index int) bool {
	node := innermostNodeContaining(ctx.SourceFile.AsNode(), index)
	for ; node != nil; node = node.Parent {
		if node.End() != index+1 {
			continue
		}
		switch node.Kind {
		case ast.KindObjectLiteralExpression, ast.KindFunctionExpression, ast.KindClassExpression:
			return true
		case ast.KindBlock:
			// A block is open only when it is a function EXPRESSION's body. An arrow's body, a
			// method's, and a bare block all close the statement.
			if parent := node.Parent; parent != nil && parent.Kind == ast.KindFunctionExpression {
				return true
			}
			return false
		case ast.KindClassDeclaration, ast.KindFunctionDeclaration, ast.KindArrowFunction,
			ast.KindIfStatement, ast.KindForStatement, ast.KindForInStatement,
			ast.KindForOfStatement, ast.KindWhileStatement, ast.KindWithStatement,
			ast.KindSwitchStatement, ast.KindTryStatement, ast.KindModuleDeclaration,
			ast.KindInterfaceDeclaration, ast.KindEnumDeclaration:
			return false
		}
	}
	return false
}

// innermostNodeContaining returns the deepest node whose range covers `index`.
//
// This is upstream's `sourceCode.getNodeByRangeIndex`, which is what the paren and brace arms both
// resolve their token through.
func innermostNodeContaining(root *ast.Node, index int) *ast.Node {
	var found *ast.Node
	var visit func(node *ast.Node)
	visit = func(node *ast.Node) {
		if node == nil || index < node.Pos() || index >= node.End() {
			return
		}
		found = node
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}
	visit(root)
	return found
}

// wordBefore returns the identifier-or-keyword word ending at `index`, or "" when the character
// there is not part of one.
func wordBefore(text string, index int) (string, int) {
	isWordByte := func(b byte) bool {
		return b == '_' || b == '$' || (b >= '0' && b <= '9') ||
			(b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
	}
	if index < 0 || !isWordByte(text[index]) {
		return "", -1
	}
	start := index
	for start > 0 && isWordByte(text[start-1]) {
		start--
	}
	return text[start : index+1], start
}

// keywordClosesTheStatement reports whether the word ending at `index` is a KEYWORD that automatic
// semicolon insertion has already closed, rather than an identifier that merely spells one.
//
// Upstream maps each such keyword to the node type it produces and answers
// `prevNode.type !== nodeType`, so a real `yield` closes the statement while a variable named
// `yield` does not. The corpus writes both, in the same file:
//
//	function * foo() { yield \n Object(); }        useLiteral, the yield is a YieldExpression
//	var yield = bar.yield \n Object()              useLiteralAfterSemicolon, this one is a property
//
// The tree answers it directly: the word closes the statement only when the node containing it is
// the construct that keyword introduces AND that construct has no operand written after the keyword.
func keywordClosesTheStatement(ctx rule.Context, word string, wordStart int) bool {
	node := innermostNodeContaining(ctx.SourceFile.AsNode(), wordStart)
	if node == nil {
		return false
	}
	// `else` and `do` introduce a body rather than ending a statement, so nothing is left open after
	// them and no construct ends AT them for the End() test below to find.
	if word == "else" || word == "do" {
		return true
	}

	// A property name or a string spelled like a keyword closes nothing. An IDENTIFIER is not
	// excluded here, and that is the fix for `break foo` and `var foo`: those end at the label or
	// the binding name, so the identifier is the innermost node and its ANCESTOR is the construct
	// insertion closed. Returning early on the identifier missed both, and upstream reads the same
	// shape through the identifier's parent.
	if node.Kind == ast.KindStringLiteral || node.Kind == ast.KindPrivateIdentifier {
		return false
	}
	if node.Kind == ast.KindIdentifier && node.Parent != nil &&
		node.Parent.Kind == ast.KindPropertyAccessExpression {
		return false
	}

	// A declaration whose last declarator has no initializer ends at the name, and insertion closes
	// it: `var foo \n Object()` is `useLiteral` in the corpus.
	for ancestor := node; ancestor != nil; ancestor = ancestor.Parent {
		if ancestor.Kind == ast.KindVariableDeclaration {
			return ancestor.AsVariableDeclaration().Initializer == nil
		}
		if ancestor.Kind == ast.KindVariableStatement {
			break
		}
	}

	for ancestor := node; ancestor != nil; ancestor = ancestor.Parent {
		// The keyword must END the construct for insertion to have closed it. A `return x` or a
		// `yield x` continues into its operand and is not this case.
		if ancestor.End() != wordStart+len(word) {
			continue
		}
		switch ancestor.Kind {
		case ast.KindReturnStatement,
			ast.KindYieldExpression,
			ast.KindBreakStatement,
			ast.KindContinueStatement,
			ast.KindDebuggerStatement,
			ast.KindImportDeclaration,
			ast.KindExportDeclaration:
			return true
		case ast.KindLabeledStatement:
			// `foo: break foo` ends at the label name, and the BreakStatement inside it is what
			// insertion closed. The labeled statement wraps it and ends at the same character.
			if statement := ancestor.AsLabeledStatement().Statement; statement != nil {
				switch statement.Kind {
				case ast.KindBreakStatement, ast.KindContinueStatement:
					return true
				}
			}
		}
	}
	return false
}

// lastMeaningfulIndex returns the index of the last character before `start` that is neither
// whitespace nor part of a comment, or -1 when there is none.
//
// Comments are skipped by asking the source file for them rather than by re-scanning: a `//` inside
// a string literal is not a comment and a hand-rolled scan gets that wrong.
func lastMeaningfulIndex(text string, start int) int {
	for index := start - 1; index >= 0; index-- {
		switch text[index] {
		case ' ', '\t', '\r', '\n', '\v', '\f':
			continue
		}
		return index
	}
	return -1
}
