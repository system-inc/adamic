package javascript

import (
	"github.com/system-inc/cohere/internal/format/printing"
)

// print/if-statement.js, for-statement.js, for-x-statement.js, while-statement.js,
// do-while-statement.js, try-statement.js, switch-statement.js and clause.js.
//
// These files carry most of the fork's customizations, each marked @system-inc upstream and here:
// no space between a keyword and its `(`, `else`, `catch` and `finally` on their own lines.

// printIfStatement is upstream's printIfStatement.
func printIfStatement(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)
	opening := groupIn(path, concatIn(path, // @system-inc: no space between `if` and `(`.
		"if(",
		printIfStatementCondition(path, options, print),
		")",
		printIfStatementConsequent(path, options, print),
	))
	if current.Child("alternate") == nil {
		return opening
	}

	// @system-inc: always put `else` on its own line, never cuddled with `}`.
	parts := []any{opening, hardline}
	danglingComments := getComments(current, commentDangling, nil)
	if len(danglingComments) > 0 {
		firstComment := danglingComments[0]
		// @system-inc: the hardline before `else` is already pushed, so only the extra blank line for a
		// comment that had one before it in the source is still needed.
		if isPreviousLineEmptyBefore(firstComment, options) {
			parts = append(parts, hardline)
		}
		var separator Doc = concatIn(path, " ")
		if needsHardlineAfterDanglingComment(current) ||
			hasNewline(options.OriginalText, locEnd(danglingComments[len(danglingComments)-1])) {
			separator = hardline
		}
		parts = append(parts, printing.PrintDanglingComments(path, options, printing.DanglingOptions[Node]{}), separator)
	}
	parts = append(parts, "else", groupIn(path, printIfStatementAlternate(path, options, print)))
	return concatIn(path, parts...)
}

// printForStatement is upstream's printForStatement.
func printForStatement(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)
	body := printClause(path, options, print, "body")

	// Dangling comments stay above the loop.
	dangling := printing.PrintDanglingComments(path, options, printing.DanglingOptions[Node]{})
	var printedComments Doc = emptyDoc
	if !isEmptyString(dangling) {
		printedComments = concatIn(path, dangling, softline)
	}

	if current.Child("init") == nil && current.Child("test") == nil && current.Child("update") == nil {
		return concatIn(path, printedComments, groupIn(path, concatIn(path /* @system-inc: no space */, "for(;;)", body)))
	}

	var update Doc = emptyDoc
	if current.Child("update") != nil {
		update = concatIn(path, line, print("update", nil))
	}
	return concatIn(path, printedComments,
		groupIn(path, concatIn(path, // @system-inc: no space between `for` and `(`.
			"for(",
			groupIn(path, concatIn(path, indentIn(path, concatIn(path, softline, print("init", nil), ";", line, print("test", nil), ";", update)),
				softline,
			)),
			")",
			body,
		)),
	)
}

// printForXStatement is upstream's printForXStatement.
func printForXStatement(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)
	isForOfStatement := current.Is("ForOfStatement")
	await, keyword := "", "in"
	if isForOfStatement {
		keyword = "of"
		if current.Bool("await") {
			await = " await"
		}
	}
	return groupIn(path, concatIn(path, "for",
		await,
		// @system-inc: no space between `for` and `(`.
		"(",
		print("left", nil),
		" ",
		keyword,
		" ",
		print("right", nil),
		")",
		printClause(path, options, print, "body"),
	))
}

// printWhileStatement is upstream's printWhileStatement.
func printWhileStatement(path *Path, options *Options, print PrintFunc) Doc {
	keyword := "while"
	if node(path).Is("WithStatement") {
		keyword = "with"
	}
	return groupIn(path, concatIn(path, keyword,
		// @system-inc: no space between the keyword and `(`.
		"(",
		printWhileStatementCondition(path, options, print),
		")",
		printClause(path, options, print, "body"),
	))
}

// printDoWhileStatement is upstream's printDoWhileStatement.
func printDoWhileStatement(path *Path, options *Options, print PrintFunc) Doc {
	var separator Doc = hardline
	if node(path).Child("body").Is("BlockStatement") {
		separator = concatIn(path, " ")
	}
	return concatIn(path, groupIn(path, concatIn(path, "do", printClause(path, options, print, "body"))),
		separator,
		// @system-inc: no space between `while` and `(`.
		"while(",
		printDoWhileStatementCondition(path, options, print),
		")",
		printSemicolon(options),
	)
}

// printTryStatement is upstream's printTryStatement.
func printTryStatement(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)
	var handler, finalizer Doc = emptyDoc, emptyDoc
	if current.Child("handler") != nil {
		handler = print("handler", nil)
	}
	if current.Child("finalizer") != nil {
		finalizer = concatIn(path, hardline, "finally ", print("finalizer", nil))
	}
	// @system-inc: block boundaries always break the line, so neither `catch` nor `finally` is cuddled
	// with the preceding `}`. printCatchClause emits its own leading hardline; `finally` gets one here.
	return concatIn(path, "try ", print("block", nil), handler, finalizer)
}

// printCatchClause is upstream's printCatchClause.
func printCatchClause(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)
	if param := current.Child("param"); param != nil {
		text := options.OriginalText
		parameterHasComments := hasComment(param, 0, func(comment Node) bool {
			data := comment.CommentData()
			return !isBlockComment(comment) ||
				data.Leading && hasNewline(text, locEnd(comment)) ||
				data.Trailing && hasNewlineBackwards(text, locStart(comment))
		})
		printedParam := print("param", nil)
		var parameter Doc
		if parameterHasComments {
			parameter = concatIn(path, "(", indentIn(path, concatIn(path, softline, printedParam)), softline, ") ")
		} else {
			parameter = concatIn(path, "(", printedParam, ") ")
		}
		// @system-inc: `catch` starts on its own line, never cuddled with `}`.
		return concatIn(path, hardline, "catch", parameter, print("body", nil))
	}
	return concatIn(path, hardline, "catch ", print("body", nil))
}

// printSwitchStatement is upstream's printSwitchStatement.
func printSwitchStatement(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)
	var cases Doc
	if len(current.List("cases")) > 0 {
		printed := mapPath(path, func(path *Path, _ int) Doc {
			var blank Doc = emptyDoc
			if !path.IsLast() && isNextLineEmptyAfter(node(path), options) {
				blank = hardline
			}
			return concatIn(path, print(nil, nil), blank)
		}, "cases")
		cases = indentIn(path, concatIn(path, hardline, join(hardline, printed)))
	} else {
		cases = printing.PrintDanglingComments(path, options, printing.DanglingOptions[Node]{Indent: true})
	}
	return concatIn(path, groupIn(path, concatIn(path, // @system-inc: no space between `switch` and `(`.
		"switch(",
		indentIn(path, concatIn(path, softline, print("discriminant", nil))),
		softline,
		")",
	)),
		" {",
		cases,
		hardline,
		"}",
	)
}

// printSwitchCase is upstream's printSwitchCase.
func printSwitchCase(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)
	var parts []any
	if current.Child("test") != nil {
		parts = append(parts, "case ", print("test", nil), ":")
	} else {
		parts = append(parts, "default:")
	}
	if hasComment(current, commentDangling, nil) {
		parts = append(parts, " ", printing.PrintDanglingComments(path, options, printing.DanglingOptions[Node]{}))
	}
	var consequent []Node
	for _, statement := range current.List("consequent") {
		if !statement.Is("EmptyStatement") {
			consequent = append(consequent, statement)
		}
	}
	if len(consequent) > 0 {
		printed := printStatementSequence(path, options, print, "consequent")
		if len(consequent) == 1 && consequent[0].Is("BlockStatement") {
			parts = append(parts, concatIn(path, " ", printed))
		} else {
			parts = append(parts, indentIn(path, concatIn(path, hardline, printed)))
		}
	}
	return concatIn(path, parts...)
}

// shouldPrintLeadingHardline is upstream's shouldPrintLeadingHardline, print/clause.js.
func shouldPrintLeadingHardline(node Node, options *Options) bool {
	leadingComments := getComments(node, commentLeading, nil)
	if len(leadingComments) == 0 {
		return false
	}
	firstComment := leadingComments[0]
	text := options.OriginalText
	start := locStart(firstComment)
	return hasNewlineInRange(text, start, locEnd(firstComment)) || hasNewlineBackwards(text, start)
}

// printClause is upstream's printClause, exported upstream as printDoWhileStatementBody,
// printForXStatementBody and printWhileStatementBody.
func printClause(path *Path, options *Options, print PrintFunc, property string) Doc {
	return call(path, func(path *Path) Doc {
		current := node(path)
		printed := print(nil, nil)
		if current.Is("EmptyStatement") {
			if hasComment(current, commentLeading, nil) {
				return concatIn(path, " ", printed)
			}
			return printed
		}
		isBlockStatement := current.Is("BlockStatement")
		if shouldPrintLeadingHardline(current, options) {
			if isBlockStatement {
				return concatIn(path, hardline, printed)
			}
			return indentIn(path, concatIn(path, hardline, printed))
		}
		if isBlockStatement || current.Is("IfStatement") && parentOf(path).Is("IfStatement") && keyOf(path) == "alternate" {
			return concatIn(path, " ", printed)
		}
		return indentIn(path, concatIn(path, line, printed))
	}, property)
}

// printIfStatementConsequent is upstream's printIfStatementConsequent.
func printIfStatementConsequent(path *Path, options *Options, print PrintFunc) Doc {
	return printClause(path, options, print, "consequent")
}

// printIfStatementAlternate is upstream's printIfStatementAlternate.
func printIfStatementAlternate(path *Path, options *Options, print PrintFunc) Doc {
	return printClause(path, options, print, "alternate")
}
