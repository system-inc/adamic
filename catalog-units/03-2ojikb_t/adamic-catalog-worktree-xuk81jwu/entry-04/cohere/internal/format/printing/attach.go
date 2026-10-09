package printing

import (
	"fmt"
	"sort"
	"strings"
)

// AddLeadingComment is upstream's addLeadingComment, src/main/comments/utilities.js.
func AddLeadingComment[N Node[N]](node N, comment N) {
	fields := comment.CommentData()
	fields.Leading = true
	fields.Trailing = false
	addComment(node, comment)
}

// AddDanglingComment is upstream's addDanglingComment. An empty marker sets none.
func AddDanglingComment[N Node[N]](node N, comment N, marker string) {
	fields := comment.CommentData()
	fields.Leading = false
	fields.Trailing = false
	if marker != "" {
		fields.Marker = marker
	}
	addComment(node, comment)
}

// AddTrailingComment is upstream's addTrailingComment.
func AddTrailingComment[N Node[N]](node N, comment N) {
	fields := comment.CommentData()
	fields.Leading = false
	fields.Trailing = true
	addComment(node, comment)
}

// addComment is upstream's addCommentHelper, without the debugging description.
func addComment[N Node[N]](node N, comment N) {
	data := node.CommentData()
	data.Comments = append(data.Comments, comment)
	comment.CommentData().Printed = false
}

// decorated is upstream's decorateComment result: the enclosing, preceding and following nodes.
type decorated[N Node[N]] struct {
	enclosing, preceding, following          N
	hasEnclosing, hasPreceding, hasFollowing bool
}

// AttachComments is upstream's attachComments, src/main/comments/attach.js: decide, for every comment,
// which node it belongs to and whether it leads, trails or dangles there.
func AttachComments[N Node[N]](ast N, comments []N, options *Options[N]) {
	printer := options.Printer
	if len(comments) == 0 || printer.CanAttachComment == nil {
		return
	}

	childNodesCache := map[N][]N{}
	text := options.OriginalText

	contexts := make([]*CommentContext[N], len(comments))
	var noEnclosing N
	for index, comment := range comments {
		found := decorateComment(ast, comment, options, noEnclosing, false, nil, childNodesCache)
		contexts[index] = &CommentContext[N]{
			Comment:       comment,
			Text:          text,
			Options:       options,
			Ast:           ast,
			IsLastComment: index == len(comments)-1,
			Enclosing:     found.enclosing, HasEnclosing: found.hasEnclosing,
			Preceding: found.preceding, HasPreceding: found.hasPreceding,
			Following: found.following, HasFollowing: found.hasFollowing,
		}
	}

	var tiesToBreak []*CommentContext[N]
	for index, context := range contexts {
		comment := context.Comment
		placement := "remaining"
		switch {
		case isOwnLineComment(text, options, contexts, index):
			placement = "ownLine"
		case isEndOfLineComment(text, options, contexts, index):
			placement = "endOfLine"
		}
		context.Placement = placement

		fields := comment.CommentData()
		fields.Enclosing, fields.hasEnclosing = context.Enclosing, context.HasEnclosing
		fields.Preceding, fields.hasPreceding = context.Preceding, context.HasPreceding
		fields.Following, fields.hasFollowing = context.Following, context.HasFollowing
		fields.Placement = placement

		handle := func(handler func(*CommentContext[N]) bool) bool {
			return handler != nil && handler(context)
		}

		switch placement {
		case "ownLine":
			// A comment on its own line prefers to lead what follows it.
			switch {
			case handle(printer.HandleComments.OwnLine):
			case context.HasFollowing:
				AddLeadingComment(context.Following, comment)
			case context.HasPreceding:
				AddTrailingComment(context.Preceding, comment)
			case context.HasEnclosing:
				AddDanglingComment(context.Enclosing, comment, "")
			default:
				AddDanglingComment(ast, comment, "")
			}
		case "endOfLine":
			// Content before it on the line and none after: prefer trailing the previous node.
			switch {
			case handle(printer.HandleComments.EndOfLine):
			case context.HasPreceding:
				AddTrailingComment(context.Preceding, comment)
			case context.HasFollowing:
				AddLeadingComment(context.Following, comment)
			case context.HasEnclosing:
				AddDanglingComment(context.Enclosing, comment, "")
			default:
				AddDanglingComment(ast, comment, "")
			}
		default:
			switch {
			case handle(printer.HandleComments.Remaining):
			case context.HasPreceding && context.HasFollowing:
				// Text on both sides on the same line: defer to the tie-breaker.
				if len(tiesToBreak) > 0 {
					lastTie := tiesToBreak[len(tiesToBreak)-1]
					if lastTie.Following != context.Following {
						breakTies(tiesToBreak, options)
						tiesToBreak = tiesToBreak[:0]
					}
				}
				tiesToBreak = append(tiesToBreak, context)
			case context.HasPreceding:
				AddTrailingComment(context.Preceding, comment)
			case context.HasFollowing:
				AddLeadingComment(context.Following, comment)
			case context.HasEnclosing:
				AddDanglingComment(context.Enclosing, comment, "")
			default:
				AddDanglingComment(ast, comment, "")
			}
		}
	}
	breakTies(tiesToBreak, options)

	// Upstream deletes these to break cycles; here they are cleared so nothing reads stale neighbours.
	for _, comment := range comments {
		fields := comment.CommentData()
		var zero N
		fields.Enclosing, fields.Preceding, fields.Following = zero, zero, zero
		fields.hasEnclosing, fields.hasPreceding, fields.hasFollowing = false, false, false
	}
}

// decorateComment is upstream's decorateComment: binary-search the sorted children of node for the
// nodes around comment, descending into a child that contains it.
func decorateComment[N Node[N]](node N, comment N, options *Options[N], enclosing N, hasEnclosing bool, ancestors []N, cache map[N][]N) decorated[N] {
	printer := options.Printer
	commentStart := printer.LocStart(comment)
	commentEnd := printer.LocEnd(comment)

	childNodes := sortedChildNodes(node, ancestors, options, cache)
	result := decorated[N]{enclosing: enclosing, hasEnclosing: hasEnclosing}

	left, right := 0, len(childNodes)
	for left < right {
		middle := (left + right) >> 1
		child := childNodes[middle]
		start := printer.LocStart(child)
		end := printer.LocEnd(child)

		if start <= commentStart && commentEnd <= end {
			// Completely contained by this child: abandon the search here and descend.
			return decorateComment(child, comment, options, child, true, append([]N{child}, ancestors...), cache)
		}
		if end <= commentStart {
			// Completely before the comment: the closest preceding node so far.
			result.preceding, result.hasPreceding = child, true
			left = middle + 1
			continue
		}
		if commentEnd <= start {
			// Completely after the comment: the closest following node so far.
			result.following, result.hasFollowing = child, true
			right = middle
			continue
		}
		panic(fmt.Sprintf("printing: comment at %d-%d overlaps a %s at %d-%d", commentStart, commentEnd, child.Type(), start, end))
	}

	// Comments inside one expression of a template literal must not move to another expression.
	if hasEnclosing && printer.TemplateQuasis != nil {
		if quasis, isTemplate := printer.TemplateQuasis(enclosing); isTemplate {
			commentIndex := findExpressionIndexForComment(quasis, comment, options)
			if result.hasPreceding && findExpressionIndexForComment(quasis, result.preceding, options) != commentIndex {
				var zero N
				result.preceding, result.hasPreceding = zero, false
			}
			if result.hasFollowing && findExpressionIndexForComment(quasis, result.following, options) != commentIndex {
				var zero N
				result.following, result.hasFollowing = zero, false
			}
		}
	}
	return result
}

// sortedChildNodes is upstream's getSortedChildNodes, src/main/utilities/get-sorted-child-nodes.js: the
// attachable descendants of node, flattening through nodes a comment cannot attach to, sorted by
// position. Cached per node, as upstream caches in childNodesCache.
func sortedChildNodes[N Node[N]](node N, ancestors []N, options *Options[N], cache map[N][]N) []N {
	if cached, present := cache[node]; present {
		return cached
	}
	printer := options.Printer
	if printer.CanAttachComment == nil {
		return nil
	}

	// Upstream's `getCommentChildNodes(...) ?? getChildren(...)`: a printer that answers with an empty
	// list means no children, which is not the same as not answering.
	var children []N
	provided := false
	if printer.GetCommentChildNodes != nil {
		children, provided = printer.GetCommentChildNodes(node, options)
	}
	if !provided {
		children = childNodes(node, printer.VisitorKeys)
	}

	childAncestors := append([]N{node}, ancestors...)
	var result []N
	for _, child := range children {
		if printer.CanAttachComment(child, childAncestors) {
			result = append(result, child)
		} else {
			result = append(result, sortedChildNodes(child, childAncestors, options, cache)...)
		}
	}
	sort.SliceStable(result, func(left, right int) bool {
		leftStart, rightStart := printer.LocStart(result[left]), printer.LocStart(result[right])
		if leftStart != rightStart {
			return leftStart < rightStart
		}
		return printer.LocEnd(result[left]) < printer.LocEnd(result[right])
	})
	cache[node] = result
	return result
}

// childNodes is upstream's getChildren, src/utilities/ast.js: the nodes under each visitor key, in key
// order, skipping absent values.
func childNodes[N Node[N]](node N, visitorKeys func(N) []string) []N {
	var children []N
	for _, key := range visitorKeys(node) {
		value := node.Field(key)
		if isNil(value) {
			continue
		}
		if isSlice(value) {
			for index := 0; index < sliceLength(value); index++ {
				element, _ := sliceElement(value, index)
				if child, isNode := asNode[N](element); isNode && !isNil(element) {
					children = append(children, child)
				}
			}
			continue
		}
		if child, isNode := asNode[N](value); isNode {
			children = append(children, child)
		}
	}
	return children
}

// isAllEmptyAndNoLineBreak is upstream's !/[\S\n  ]/.test(text): only non-newline whitespace.
func isAllEmptyAndNoLineBreak(text string) bool {
	for _, character := range text {
		if character == '\n' || character == 0x2028 || character == 0x2029 || !isJavaScriptWhitespace(character) {
			return false
		}
	}
	return true
}

// isOwnLineComment is upstream's isOwnLineComment.
func isOwnLineComment[N Node[N]](text string, options *Options[N], contexts []*CommentContext[N], commentIndex int) bool {
	current := contexts[commentIndex]
	start := options.Printer.LocStart(current.Comment)
	if current.HasPreceding {
		// Find the first comment on the same line.
		for index := commentIndex - 1; index >= 0; index-- {
			previous := contexts[index]
			if !previous.HasPreceding || previous.Preceding != current.Preceding ||
				!isAllEmptyAndNoLineBreak(text[options.Printer.LocEnd(previous.Comment):start]) {
				break
			}
			start = options.Printer.LocStart(previous.Comment)
		}
	}
	return HasNewline(text, start, true)
}

// isEndOfLineComment is upstream's isEndOfLineComment.
func isEndOfLineComment[N Node[N]](text string, options *Options[N], contexts []*CommentContext[N], commentIndex int) bool {
	current := contexts[commentIndex]
	end := options.Printer.LocEnd(current.Comment)
	if current.HasFollowing {
		// Find the last comment on the same line.
		for index := commentIndex + 1; index < len(contexts); index++ {
			next := contexts[index]
			if !next.HasFollowing || next.Following != current.Following ||
				!isAllEmptyAndNoLineBreak(text[end:options.Printer.LocStart(next.Comment)]) {
				break
			}
			end = options.Printer.LocEnd(next.Comment)
		}
	}
	return HasNewline(text, end, false)
}

// breakTies is upstream's breakTies: of comments between the same preceding and following node, the
// trailing run of ones separated from the following node by only gaps lead it; the rest trail.
func breakTies[N Node[N]](tiesToBreak []*CommentContext[N], options *Options[N]) {
	tieCount := len(tiesToBreak)
	if tieCount == 0 {
		return
	}
	preceding := tiesToBreak[0].Preceding
	following := tiesToBreak[0].Following
	printer := options.Printer

	gapEnd := printer.LocStart(following)
	indexOfFirstLeadingComment := tieCount
	for ; indexOfFirstLeadingComment > 0; indexOfFirstLeadingComment-- {
		tie := tiesToBreak[indexOfFirstLeadingComment-1]
		if tie.Preceding != preceding || tie.Following != following {
			panic("printing: a tie to break does not share its neighbours")
		}
		gap := options.OriginalText[printer.LocEnd(tie.Comment):gapEnd]
		if isGap(gap, options) {
			gapEnd = printer.LocStart(tie.Comment)
		} else {
			break
		}
	}

	for index, tie := range tiesToBreak {
		if index < indexOfFirstLeadingComment {
			AddTrailingComment(preceding, tie.Comment)
		} else {
			AddLeadingComment(following, tie.Comment)
		}
	}

	for _, node := range []N{preceding, following} {
		comments := node.CommentData().Comments
		if len(comments) > 1 {
			sort.SliceStable(comments, func(left, right int) bool {
				return printer.LocStart(comments[left]) < printer.LocStart(comments[right])
			})
		}
	}
}

// isGap is upstream's `printer.isGap?.(gap, options) ?? /^[\s(]*$/.test(gap)`: the printer decides if
// it has an opinion, and otherwise a gap is whitespace and opening parentheses.
func isGap[N Node[N]](gap string, options *Options[N]) bool {
	if options.Printer.IsGap != nil {
		if answer, decided := options.Printer.IsGap(gap, options); decided {
			return answer
		}
	}
	return strings.IndexFunc(gap, func(character rune) bool {
		return character != '(' && !isJavaScriptWhitespace(character)
	}) < 0
}

// findExpressionIndexForComment is upstream's findExpressionIndexForComment.
func findExpressionIndexForComment[N Node[N]](quasis []N, comment N, options *Options[N]) int {
	startPosition := options.Printer.LocStart(comment) - 1
	for index := 1; index < len(quasis); index++ {
		if startPosition < options.Printer.LocStart(quasis[index]) {
			return index - 1
		}
	}
	return 0
}
