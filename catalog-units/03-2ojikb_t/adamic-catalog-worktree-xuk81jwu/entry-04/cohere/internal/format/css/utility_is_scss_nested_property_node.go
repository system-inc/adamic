package css

// src/language-css/utilities/is-scss-nested-property-node.js.

import (
	"regexp"
	"strings"

	"github.com/system-inc/cohere/internal/format/estree"
)

// JavaScript's `.` stops at every line terminator, Go's only at \n.
var (
	scssNestedPropertyBlockComment  = regexp.MustCompile(`/\*[^\n\r\x{2028}\x{2029}]*?\*/`)
	scssNestedPropertyInlineComment = regexp.MustCompile(`//[^\n\r\x{2028}\x{2029}]*\n`)
)

func isScssNestedPropertyNode(node *estree.Node, options *parseOptions) bool {
	if options.parser != "scss" {
		return false
	}

	selector, isString := node.Get("selector").(string)
	if !isString || selector == "" {
		return false
	}

	// String.prototype.replace with a regular expression that is not global replaces the first match
	// (replaceFirst is print_misc.go's).
	selector = replaceFirst(scssNestedPropertyBlockComment, selector, "")
	selector = replaceFirst(scssNestedPropertyInlineComment, selector, "")
	return strings.HasSuffix(estree.TrimEndJavaScript(selector), ":")
}
