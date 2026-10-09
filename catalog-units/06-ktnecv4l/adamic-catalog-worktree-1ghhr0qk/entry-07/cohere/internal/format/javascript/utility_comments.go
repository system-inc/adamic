package javascript

import "github.com/system-inc/cohere/internal/format/estree"

// utilities/comments.js, utilities/comment-types.js, utilities/is-prettier-ignore-comment.js.

// CommentCheckFlags are upstream's CommentCheckFlags.
type commentCheckFlags int

const (
	commentLeading        commentCheckFlags = 1 << 1
	commentTrailing       commentCheckFlags = 1 << 2
	commentDangling       commentCheckFlags = 1 << 3
	commentBlock          commentCheckFlags = 1 << 4
	commentLine           commentCheckFlags = 1 << 5
	commentPrettierIgnore commentCheckFlags = 1 << 6
	commentFirst          commentCheckFlags = 1 << 7
	commentLast           commentCheckFlags = 1 << 8
)

// commentsOf is node.comments.
func commentsOf(node Node) []Node {
	if node == nil {
		return nil
	}
	return node.Comments
}

// commentTest is upstream's getCommentTestFunction. A nil result means "every comment matches".
func commentTest(flags commentCheckFlags, fn func(Node) bool) func(comment Node, index int, comments []Node) bool {
	if flags == 0 && fn == nil {
		return nil
	}
	return func(comment Node, index int, comments []Node) bool {
		data := comment.CommentData()
		return !(flags&commentLeading != 0 && !data.Leading ||
			flags&commentTrailing != 0 && !data.Trailing ||
			flags&commentDangling != 0 && (data.Leading || data.Trailing) ||
			flags&commentBlock != 0 && !isBlockComment(comment) ||
			flags&commentLine != 0 && !isLineComment(comment) ||
			flags&commentFirst != 0 && index != 0 ||
			flags&commentLast != 0 && index != len(comments)-1 ||
			flags&commentPrettierIgnore != 0 && !isPrettierIgnoreComment(comment) ||
			fn != nil && !fn(comment))
	}
}

// hasComment is upstream's hasComment(node, flags, fn).
func hasComment(node Node, flags commentCheckFlags, fn func(Node) bool) bool {
	comments := commentsOf(node)
	if len(comments) == 0 {
		return false
	}
	test := commentTest(flags, fn)
	if test == nil {
		return true
	}
	for index, comment := range comments {
		if test(comment, index, comments) {
			return true
		}
	}
	return false
}

// hasAnyComment is upstream's hasComment(node) with no flags.
func hasAnyComment(node Node) bool { return hasComment(node, 0, nil) }

// getComments is upstream's getComments(node, flags, fn).
func getComments(node Node, flags commentCheckFlags, fn func(Node) bool) []Node {
	comments := commentsOf(node)
	test := commentTest(flags, fn)
	if test == nil {
		return comments
	}
	var result []Node
	for index, comment := range comments {
		if test(comment, index, comments) {
			result = append(result, comment)
		}
	}
	return result
}

// isBlockComment is upstream's isBlockComment.
func isBlockComment(comment Node) bool { return comment.Is("Block", "CommentBlock", "MultiLine") }

// isLineComment is upstream's isLineComment.
func isLineComment(comment Node) bool {
	return comment.Is("Line", "CommentLine", "SingleLine", "HashbangComment", "HTMLOpen", "HTMLClose", "Hashbang",
		"InterpreterDirective")
}

// isPrettierIgnoreComment is upstream's isPrettierIgnoreComment. `unignore` is set by
// comments/handle-comments.js when it moves a prettier-ignore onto a union member or a mapped type.
func isPrettierIgnoreComment(comment Node) bool {
	return estree.TrimJavaScript(comment.String("value")) == "prettier-ignore" && !comment.Bool("unignore")
}
