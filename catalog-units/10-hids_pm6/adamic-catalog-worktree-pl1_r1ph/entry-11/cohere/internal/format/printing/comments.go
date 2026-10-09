package printing

import (
	"fmt"

	"github.com/system-inc/cohere/internal/format/doc"
)

// CommentFilter is upstream's printOptions.filter: which comments a print call should include.
type CommentFilter[N Node[N]] func(comment N) bool

// printComment is upstream's printComment, src/main/comments/print.js: mark it printed, then print it.
func printComment[N Node[N]](path *AstPath[N], options *Options[N]) doc.Doc {
	comment, _ := path.Node()
	comment.CommentData().Printed = true
	return options.Printer.PrintComment(path, options)
}

// printLeadingComment is upstream's printLeadingComment.
func printLeadingComment[N Node[N]](path *AstPath[N], options *Options[N]) doc.Doc {
	comment, _ := path.Node()
	parts := doc.Concat{printComment(path, options)}
	printer := options.Printer
	text := options.OriginalText

	if printer.IsBlockComment != nil && printer.IsBlockComment(comment) {
		var lineBreak doc.Doc = doc.Text(" ")
		if HasNewline(text, printer.LocEnd(comment), false) {
			if HasNewline(text, printer.LocStart(comment), true) {
				lineBreak = doc.Hardline
			} else {
				lineBreak = doc.LineDoc
			}
		}
		parts = append(parts, lineBreak)
	} else {
		parts = append(parts, doc.Hardline)
	}

	index := SkipNewline(text, SkipSpaces(text, printer.LocEnd(comment), false), false)
	if index != notFound && HasNewline(text, index, false) {
		parts = append(parts, doc.Hardline)
	}
	return parts
}

// trailingPrinted is upstream's printTrailingComment result.
type trailingPrinted struct {
	doc           doc.Doc
	isBlock       bool
	hasLineSuffix bool
	present       bool
}

// printTrailingComment is upstream's printTrailingComment.
func printTrailingComment[N Node[N]](path *AstPath[N], options *Options[N], previous trailingPrinted) trailingPrinted {
	comment, _ := path.Node()
	printed := printComment(path, options)
	printer := options.Printer
	text := options.OriginalText
	isBlock := printer.IsBlockComment != nil && printer.IsBlockComment(comment)

	if (previous.present && previous.hasLineSuffix && !previous.isBlock) || HasNewline(text, printer.LocStart(comment), true) {
		// This allows comments at the end of nested structures:
		// {
		//   x: 1,
		//   y: 2
		//   // A comment
		// }
		// Those kinds of comments are almost always leading comments, but here it doesn't go "outside"
		// the block and turns it into a trailing comment for `2`. We can simulate the above by checking
		// if this a comment on its own line; normal trailing comments are always at the end of another
		// expression.
		var blank doc.Doc = doc.Text("")
		if IsPreviousLineEmpty(text, printer.LocStart(comment)) {
			blank = doc.Hardline
		}
		return trailingPrinted{
			doc:           doc.NewLineSuffix(doc.Concat{doc.Hardline, blank, printed}),
			isBlock:       isBlock,
			hasLineSuffix: true,
			present:       true,
		}
	}

	if !isBlock || (previous.present && previous.hasLineSuffix) {
		return trailingPrinted{
			doc:           doc.Concat{doc.NewLineSuffix(doc.Concat{doc.Text(" "), printed}), doc.BreakParent},
			isBlock:       isBlock,
			hasLineSuffix: true,
			present:       true,
		}
	}
	return trailingPrinted{doc: doc.Concat{doc.Text(" "), printed}, isBlock: isBlock, present: true}
}

// DanglingOptions are upstream's printDanglingComments options.
type DanglingOptions[N Node[N]] struct {
	Indent bool
	Marker string
	Filter CommentFilter[N]
}

// PrintDanglingComments is upstream's printDanglingComments: the comments attached to the current node
// that neither lead nor trail it, joined by hardlines.
func PrintDanglingComments[N Node[N]](path *AstPath[N], options *Options[N], danglingOptions DanglingOptions[N]) doc.Doc {
	node, isNode := path.Node()
	if !isNode || isNil(node) {
		return doc.Text("")
	}
	dangling := map[N]bool{}
	for _, comment := range node.CommentData().Comments {
		fields := comment.CommentData()
		if fields.Leading || fields.Trailing || fields.Marker != danglingOptions.Marker ||
			(danglingOptions.Filter != nil && !danglingOptions.Filter(comment)) {
			continue
		}
		dangling[comment] = true
	}
	if len(dangling) == 0 {
		return doc.Text("")
	}

	var parts []doc.Doc
	for _, printed := range Map(path, func(path *AstPath[N], _ int, _ any) doc.Doc {
		comment, _ := path.Node()
		if dangling[comment] {
			return printComment(path, options)
		}
		return nil
	}, "comments") {
		// Upstream's .filter(Boolean) drops an empty string as well as a missing doc, so a comment that
		// prints as "" does not leave an extra hardline in the join.
		if printed != nil && !isEmptyText(printed) {
			parts = append(parts, printed)
		}
	}

	joined := doc.Join(doc.Hardline, parts)
	if danglingOptions.Indent {
		return doc.NewIndent(doc.Concat{doc.Hardline, joined})
	}
	return joined
}

// PrintLeadingComments is upstream's printLeadingComments. A nil filter includes every comment.
func PrintLeadingComments[N Node[N]](path *AstPath[N], options *Options[N], filter CommentFilter[N]) doc.Doc {
	node, isNode := path.Node()
	if !isNode || isNil(node) {
		return doc.Text("")
	}
	leading := map[N]bool{}
	for _, comment := range node.CommentData().Comments {
		if !options.printedComments[comment] && comment.CommentData().Leading && (filter == nil || filter(comment)) {
			leading[comment] = true
		}
	}
	if len(leading) == 0 {
		return doc.Text("")
	}

	var parts doc.Concat
	for _, printed := range Map(path, func(path *AstPath[N], _ int, _ any) doc.Doc {
		comment, _ := path.Node()
		if leading[comment] {
			return printLeadingComment(path, options)
		}
		return nil
	}, "comments") {
		if printed != nil {
			parts = append(parts, printed)
		}
	}
	return parts
}

// PrintTrailingComments is upstream's printTrailingComments. Each trailing comment is printed even when
// filtered out, because the next one's layout depends on the previous one's; only the kept ones join the
// result, exactly as upstream threads printedTrailingComment through every trailing comment.
func PrintTrailingComments[N Node[N]](path *AstPath[N], options *Options[N], filter CommentFilter[N]) doc.Doc {
	node, isNode := path.Node()
	if !isNode || isNil(node) {
		return doc.Text("")
	}
	comments := node.CommentData().Comments
	trailing := map[N]bool{}
	shouldPrint := map[N]bool{}
	for _, comment := range comments {
		if comment.CommentData().Trailing {
			trailing[comment] = true
			if !options.printedComments[comment] && (filter == nil || filter(comment)) {
				shouldPrint[comment] = true
			}
		}
	}
	if len(shouldPrint) == 0 {
		return doc.Text("")
	}

	var docs doc.Concat
	var previous trailingPrinted
	path.Each(func(path *AstPath[N], _ int, _ any) {
		comment, _ := path.Node()
		if !trailing[comment] {
			return
		}
		previous = printTrailingComment(path, options, previous)
		if shouldPrint[comment] {
			docs = append(docs, previous.doc)
		}
	}, "comments")
	return docs
}

// PrintCommentsSeparately is upstream's printCommentsSeparately.
func PrintCommentsSeparately[N Node[N]](path *AstPath[N], options *Options[N], filter CommentFilter[N]) (leading doc.Doc, trailing doc.Doc) {
	return PrintLeadingComments(path, options, filter), PrintTrailingComments(path, options, filter)
}

// PrintComments is upstream's printComments: the node's doc between its leading and trailing comments,
// keeping its label.
func PrintComments[N Node[N]](path *AstPath[N], printed doc.Doc, options *Options[N], filter CommentFilter[N]) doc.Doc {
	leading := PrintLeadingComments(path, options, filter)
	trailing := PrintTrailingComments(path, options, filter)
	if isEmptyText(leading) && isEmptyText(trailing) {
		return printed
	}
	return doc.InheritLabel(printed, func(contents doc.Doc) doc.Doc {
		return doc.Concat{leading, contents, trailing}
	})
}

// isEmptyText is upstream's falsy test on printLeadingComments' result, which is "" when there is
// nothing and an array otherwise. An empty array is truthy in JavaScript, and so is non-empty here.
func isEmptyText(document doc.Doc) bool {
	text, isText := document.(doc.Text)
	return isText && text == ""
}

// ensureAllCommentsPrinted is upstream's ensureAllCommentsPrinted: a comment no printer printed is lost
// output, so it is an error rather than a silent deletion.
func ensureAllCommentsPrinted[N Node[N]](options *Options[N]) error {
	for _, comment := range options.Comments {
		if !comment.CommentData().Printed && !options.printedComments[comment] {
			start := options.Printer.LocStart(comment)
			end := options.Printer.LocEnd(comment)
			return fmt.Errorf("comment %q was not printed; please report this error", options.OriginalText[start:end])
		}
	}
	return nil
}
