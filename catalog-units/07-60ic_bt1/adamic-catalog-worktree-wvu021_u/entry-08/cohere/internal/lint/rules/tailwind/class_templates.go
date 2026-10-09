package tailwind

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
)

// ClassTemplate is a template literal on a class surface, broken at its interpolation boundaries.
//
// This is deliberately a separate reading from ClassLiteral. A template is only partly knowable, so
// rules that rewrite class text must not see it, and the reader that feeds them keeps excluding it.
// Only a rule that reasons about the boundaries themselves wants this shape.
type ClassTemplate struct {
	Node *ast.Node
	// Boundaries are the seams between static text and interpolation, in source order.
	Boundaries []TemplateBoundary
	Origin     ClassLiteralOrigin
}

// TemplateBoundary is one seam where static class text meets an interpolation.
//
// Side says which way the static text faces: text ending at a hole is on the hole's left, text
// starting after a hole is on its right. Fragment is the class text actually touching the seam, and
// is empty when whitespace separates them, which is the case that is fine.
type TemplateBoundary struct {
	Side     TemplateBoundarySide
	Fragment string
	// Range points at the static text that owns the seam, so a finding lands on the fragment the
	// author wrote rather than on the whole template.
	Range core.TextRange
}

// TemplateBoundarySide distinguishes the two directions a fragment can touch a hole from.
type TemplateBoundarySide string

const (
	// TemplateBoundarySideBeforeHole is static text running up to an interpolation: the `px-` of
	// `px-${size}`.
	TemplateBoundarySideBeforeHole TemplateBoundarySide = "BeforeHole"
	// TemplateBoundarySideAfterHole is static text resuming after one: the `-4` of `${prefix}-4`.
	TemplateBoundarySideAfterHole TemplateBoundarySide = "AfterHole"
)

// ClassTemplatesIn returns the template literal a node carries on a class surface, if any.
//
// Mirrors ClassLiteralsIn's dispatch so that a rule reading templates covers the same three surfaces
// a rule reading strings does. Anything else is a difference in coverage that nobody declared.
func (r *ClassLiteralReader) ClassTemplatesIn(node *ast.Node) []ClassTemplate {
	if node == nil {
		return nil
	}

	switch node.Kind {
	case ast.KindJsxAttribute:
		attribute := node.AsJsxAttribute()
		if attribute == nil {
			return nil
		}
		name := attribute.Name()
		if name == nil || !r.attributes.matches(name.Text()) {
			return nil
		}
		return templatesFrom(templateExpressionOf(attribute.Initializer), ClassLiteralOriginAttribute)

	case ast.KindCallExpression:
		call := node.AsCallExpression()
		if call == nil || call.Expression == nil || call.Arguments == nil {
			return nil
		}
		if !r.readsCallee(call.Expression) {
			return nil
		}
		var templates []ClassTemplate
		for _, argument := range call.Arguments.Nodes {
			templates = append(templates, templatesFrom(templateExpressionOf(argument), ClassLiteralOriginCallee)...)
		}
		return templates

	case ast.KindVariableDeclaration:
		declaration := node.AsVariableDeclaration()
		if declaration == nil || declaration.Initializer == nil {
			return nil
		}
		name := declaration.Name()
		if name == nil || name.Kind != ast.KindIdentifier {
			return nil
		}
		if !r.variables.matches(name.Text()) {
			return nil
		}
		return templatesFrom(templateExpressionOf(declaration.Initializer), ClassLiteralOriginVariable)
	}

	return nil
}

// templateExpressionOf unwraps a template literal from the shapes it arrives in.
//
// A template with no substitutions is not returned: it has no interpolation boundary, so it is a
// plain string as far as this reading is concerned and belongs to the string reader.
func templateExpressionOf(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}

	switch node.Kind {
	case ast.KindTemplateExpression:
		return node

	case ast.KindJsxExpression:
		expression := node.AsJsxExpression()
		if expression == nil {
			return nil
		}
		return templateExpressionOf(expression.Expression)

	// Same omission as the string reader had: parentheses are real nodes here, so `{(`px-${s}`)}`
	// is invisible to a reader that only knows the kinds it expects. See collectClassValues for why
	// stripping is safe for these rules specifically and not a pattern to copy.
	case ast.KindParenthesizedExpression:
		parenthesized := node.AsParenthesizedExpression()
		if parenthesized == nil {
			return nil
		}
		return templateExpressionOf(parenthesized.Expression)
	}

	return nil
}

// templatesFrom breaks a template into its boundaries.
//
// The head runs up to the first hole, and each span's literal runs from its own hole to the next, so
// every span contributes a boundary on its left and, unless it is the last, one on its right. That
// enumeration is the whole rule: a boundary is a defect exactly when no whitespace sits at the seam.
func templatesFrom(node *ast.Node, origin ClassLiteralOrigin) []ClassTemplate {
	if node == nil {
		return nil
	}

	template := node.AsTemplateExpression()
	if template == nil || template.Head == nil || template.TemplateSpans == nil {
		return nil
	}

	classTemplate := ClassTemplate{Node: node, Origin: origin}

	// The head faces the first hole on its right.
	if boundary, hasBoundary := boundaryBeforeHole(template.Head.Text(), template.Head.Loc); hasBoundary {
		classTemplate.Boundaries = append(classTemplate.Boundaries, boundary)
	}

	spans := template.TemplateSpans.Nodes
	for index, spanNode := range spans {
		span := spanNode.AsTemplateSpan()
		if span == nil || span.Literal == nil {
			continue
		}
		text := span.Literal.Text()

		// Text resuming after this span's hole.
		if boundary, hasBoundary := boundaryAfterHole(text, span.Literal.Loc); hasBoundary {
			classTemplate.Boundaries = append(classTemplate.Boundaries, boundary)
		}

		// And, unless this is the final span, the text also runs up to the next hole.
		if index < len(spans)-1 {
			if boundary, hasBoundary := boundaryBeforeHole(text, span.Literal.Loc); hasBoundary {
				classTemplate.Boundaries = append(classTemplate.Boundaries, boundary)
			}
		}
	}

	if len(classTemplate.Boundaries) == 0 {
		return nil
	}
	return []ClassTemplate{classTemplate}
}

// boundaryBeforeHole reports the class fragment left touching an interpolation on its right.
//
// Empty text is fine: `${a}${b}` has nothing glued to anything, so there is no fragment whose name
// got broken. Text ending in whitespace is also fine, and is the shape the rule is teaching people
// to write.
func boundaryBeforeHole(spanText string, textRange core.TextRange) (TemplateBoundary, bool) {
	if spanText == "" || endsWithWhitespace(spanText) {
		return TemplateBoundary{}, false
	}

	fields := text.WhitespaceFields(spanText)
	if len(fields) == 0 {
		return TemplateBoundary{}, false
	}

	return TemplateBoundary{
		Side:     TemplateBoundarySideBeforeHole,
		Fragment: fields[len(fields)-1],
		Range:    textRange,
	}, true
}

// boundaryAfterHole reports the class fragment right touching an interpolation on its left.
func boundaryAfterHole(spanText string, textRange core.TextRange) (TemplateBoundary, bool) {
	if spanText == "" || startsWithWhitespace(spanText) {
		return TemplateBoundary{}, false
	}

	fields := text.WhitespaceFields(spanText)
	if len(fields) == 0 {
		return TemplateBoundary{}, false
	}

	return TemplateBoundary{
		Side:     TemplateBoundarySideAfterHole,
		Fragment: fields[0],
		Range:    textRange,
	}, true
}

// endsWithWhitespace and startsWithWhitespace ask about any whitespace, not only a space.
//
// A tab separates two class names as well as a space does, and upstream treats it that way: a
// tab-separated pair is not a finding for this rule or for the whitespace rule. Testing for `' '`
// alone would report every tab-indented multi-line class list in the tree.
func endsWithWhitespace(text string) bool {
	if text == "" {
		return false
	}
	trimmed := strings.TrimRightFunc(text, isSpace)
	return len(trimmed) != len(text)
}

func startsWithWhitespace(text string) bool {
	if text == "" {
		return false
	}
	trimmed := strings.TrimLeftFunc(text, isSpace)
	return len(trimmed) != len(text)
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == '\v' || r == '\f'
}
