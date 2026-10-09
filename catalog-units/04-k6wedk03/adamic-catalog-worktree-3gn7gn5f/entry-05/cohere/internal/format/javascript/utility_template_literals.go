package javascript

import (
	"strings"

	"github.com/system-inc/cohere/internal/format/estree"
)

// utilities/template-literal-has-new-lines.js, utilities/is-template-on-its-own-line.js.

// templateElementRaw is upstream's `element.value.raw` on a TemplateElement, whose value the tree
// holds as an *estree.TemplateValue.
func templateElementRaw(element Node) string {
	value, _ := element.Get("value").(*estree.TemplateValue)
	if value == nil {
		return ""
	}
	return value.Raw
}

// templateLiteralHasNewLines is upstream's templateLiteralHasNewLines.
func templateLiteralHasNewLines(template Node) bool {
	for _, quasi := range template.List("quasis") {
		if strings.Contains(templateElementRaw(quasi), "\n") {
			return true
		}
	}
	return false
}

// isTemplateOnItsOwnLine is upstream's isTemplateOnItsOwnLine.
func isTemplateOnItsOwnLine(node Node, text string) bool {
	return (node.Is("TemplateLiteral") && templateLiteralHasNewLines(node) ||
		node.Is("TaggedTemplateExpression") &&
			templateLiteralHasNewLines(node.Child("quasi"))) &&
		!hasNewlineBackwards(text, locStart(node))
}
