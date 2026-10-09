package estree

import (
	"strings"
)

/*
 * Prettier's parse postprocess, src/language-js/parse/postprocess/index.js, for astType "typescript".
 *
 * The printer assumes the reshaped tree, so this runs before anything else reads it. Branches for
 * other parsers (Babel's ParenthesizedExpression and interpreter directive, Flow's tuple types, oxc,
 * Hermes) are not ported: typescript-estree never produces those nodes, and porting dead branches
 * would be code nothing could test.
 */

// Postprocess reshapes a converted Program the way Prettier does, and returns the root. comments is
// the file's comment list, which it may merge in place, so it returns that too, and the text stripped of
// those comments, for the printer to share.
func Postprocess(program *Node, comments []*Node, text string) (*Node, []*Node, *StrippedText) {
	comments = mergeNestledJsdocComments(comments)

	// In typescript the Program does not count leading and trailing whitespace and comments.
	program.Range = [2]int{0, len(text)}

	processor := &postprocessor{text: text, stripped: NewStrippedText(text, comments)}
	program = processor.visit(program)
	return program, comments, processor.stripped
}

type postprocessor struct {
	text     string
	stripped *StrippedText
}

// StrippedText is a file's text with its comments blanked, StripComments, made the first time it is asked
// for. The postprocess and the printer both strip the file's text of the file's comments: the same string,
// and the same list, merged before either reads it and never moved after, so they share one strip rather
// than make two copies of the file, 37 MB apiece on a cold run on ahra (#wcgw0n4).
type StrippedText struct {
	text     string
	comments []*Node
	stripped string
	made     bool
}

// NewStrippedText is text stripped of comments, made when first asked for.
func NewStrippedText(text string, comments []*Node) *StrippedText {
	return &StrippedText{text: text, comments: comments}
}

// Text is the stripped text.
func (stripped *StrippedText) Text() string {
	if !stripped.made {
		stripped.stripped = StripComments(stripped.text, stripped.comments)
		stripped.made = true
	}
	return stripped.stripped
}

// visit is upstream's visitNode, postprocess/visit-node.js.
func (processor *postprocessor) visit(node *Node) *Node {
	if node == nil {
		return nil
	}
	if result := processor.onEnter(node); result != node {
		return processor.visit(result)
	}
	for _, key := range VisitorKeys(node) {
		switch value := node.Get(key).(type) {
		case *Node:
			node.Set(key, processor.visit(value))
		case []*Node:
			for index := range value {
				value[index] = processor.visit(value[index])
			}
		default:
			// Upstream assigns node[key] for every visitor key, so a key the converter never set
			// exists afterwards with an undefined value. `key in node` can see that.
			if !node.Has(key) {
				node.Set(key, nil)
			}
		}
	}
	return processor.onLeave(node)
}

func (processor *postprocessor) onEnter(node *Node) *Node {
	processor.setContentEnd(node)

	switch node.Type() {
	case "TemplateElement":
		// typescript follows the espree style positions.
		start := LocStart(node) + 1
		end := LocEnd(node) - 2
		if node.Bool("tail") {
			end = LocEnd(node) - 1
		}
		node.Range = [2]int{start, end}
	case "TSParenthesizedType":
		return node.Child("typeAnnotation")
	case "TSUnionType", "TSIntersectionType":
		if types := node.List("types"); len(types) == 1 {
			return types[0]
		}
	}
	return node
}

func (processor *postprocessor) onLeave(node *Node) *Node {
	if node.Is("LogicalExpression") && isUnbalancedLogicalTree(node) {
		return rebalanceLogicalTree(node)
	}
	return node
}

func isUnbalancedLogicalTree(node *Node) bool {
	return node.Is("LogicalExpression") && node.Child("right").Is("LogicalExpression") &&
		node.String("operator") == node.Child("right").String("operator")
}

// rebalanceLogicalTree is upstream's rebalanceLogicalTree: a && (b && c) becomes (a && b) && c.
func rebalanceLogicalTree(node *Node) *Node {
	if !isUnbalancedLogicalTree(node) {
		return node
	}
	right := node.Child("right")
	left := rebalanceLogicalTree(New("LogicalExpression", LocStart(node.Child("left")), LocEnd(right.Child("left")),
		"operator", node.String("operator"),
		"left", node.Child("left"),
		"right", right.Child("left")))
	return rebalanceLogicalTree(New("LogicalExpression", LocStart(node), LocEnd(node),
		"operator", node.String("operator"),
		"left", left,
		"right", right.Child("right")))
}

// setContentEnd is upstream's setContentEnd: the end of a statement before its semicolon, with the
// comments between them stripped out of consideration.
func (processor *postprocessor) setContentEnd(node *Node) {
	if !ShouldAddContentEnd(node) {
		return
	}
	end := LocEndWithFullText(node)
	if end == 0 || processor.text[end-1] != ';' {
		return
	}
	end--
	textBeforeSemicolon := processor.stripped.Text()[LocStart(node):end]
	cleaned := TrimEndJavaScript(textBeforeSemicolon)
	node.ContentEnd = end - (len(textBeforeSemicolon) - len(cleaned))
	node.HasContentEnd = true
}

// StripComments is upstream's stripComments: every comment's characters, except line feeds, become
// spaces, so offsets are unchanged. Bytes rather than characters, which keeps byte offsets unchanged.
func StripComments(text string, comments []*Node) string {
	bytes := []byte(text)
	for _, comment := range comments {
		for index := LocStart(comment); index < LocEnd(comment); index++ {
			if bytes[index] != '\n' {
				bytes[index] = ' '
			}
		}
	}
	return string(bytes)
}

// TrimEndJavaScript is String.prototype.trimEnd: JavaScript whitespace and line terminators.
func TrimEndJavaScript(text string) string {
	end := len(text)
	for end > 0 {
		size := lastWhitespaceSize(text[:end])
		if size == 0 {
			break
		}
		end -= size
	}
	return text[:end]
}

// TrimStartJavaScript is String.prototype.trimStart.
func TrimStartJavaScript(text string) string {
	start := 0
	for start < len(text) {
		size := whitespaceSize(text, start)
		if size == 0 {
			break
		}
		start += size
	}
	return text[start:]
}

// TrimJavaScript is String.prototype.trim.
func TrimJavaScript(text string) string {
	return TrimStartJavaScript(TrimEndJavaScript(text))
}

// lastWhitespaceSize is the byte size of the whitespace character ending text, or 0.
func lastWhitespaceSize(text string) int {
	for size := 1; size <= 3 && size <= len(text); size++ {
		start := len(text) - size
		if size > 1 && text[start]&0xC0 == 0x80 {
			continue
		}
		if whitespaceSize(text, start) == size {
			return size
		}
		return 0
	}
	return 0
}

// mergeNestledJsdocComments is upstream's postprocess/merge-nestled-jsdoc-comments.js: two
// indentable block comments touching end to start become one.
func mergeNestledJsdocComments(comments []*Node) []*Node {
	if len(comments) < 2 {
		return comments
	}
	var following *Node
	for index := len(comments) - 1; index >= 0; index-- {
		comment := comments[index]
		if following != nil && LocEnd(comment) == LocStart(following) &&
			IsIndentableBlockComment(comment) && IsIndentableBlockComment(following) {
			comments = append(comments[:index+1], comments[index+2:]...)
			comment.Set("value", comment.String("value")+"*//*"+following.String("value"))
			comment.Range = [2]int{LocStart(comment), LocEnd(following)}
		}
		following = comment
	}
	return comments
}

// IndentableBlockCommentLines is upstream's getIndentableBlockCommentLines: the lines of a multi-line
// block comment whose every line starts with a star, trimmed, or nil.
func IndentableBlockCommentLines(comment *Node) []string {
	if !comment.Is("Block") {
		return nil
	}
	value := comment.String("value")
	if !strings.Contains(value, "\n") {
		return nil
	}
	lines := strings.Split("*"+value+"*", "\n")
	trimmed := make([]string, 0, len(lines))
	for _, line := range lines {
		line = TrimStartJavaScript(line)
		if !strings.HasPrefix(line, "*") {
			return nil
		}
		trimmed = append(trimmed, line)
	}
	return trimmed
}

// IsIndentableBlockComment is upstream's isIndentableBlockComment.
func IsIndentableBlockComment(comment *Node) bool {
	return len(IndentableBlockCommentLines(comment)) > 0
}
