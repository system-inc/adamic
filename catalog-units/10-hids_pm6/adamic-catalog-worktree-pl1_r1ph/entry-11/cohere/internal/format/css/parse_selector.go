package css

// src/language-css/parse/parse-selector.js.

import (
	"regexp"

	"github.com/system-inc/cohere/internal/format/css/selector"
	"github.com/system-inc/cohere/internal/format/estree"
)

var (
	selectorQuotedStringPattern = regexp.MustCompile(`"[^"]+"|'[^']+'`)
	selectorCommentPattern      = regexp.MustCompile(`/[/*]`)
)

func parseSelector(selectorText string) *estree.Node {
	// If there's a comment inside of a selector, the parser tries to parse
	// the content of the comment as selectors which turns it into complete
	// garbage. Better to print the whole selector as-is and not try to parse
	// and reformat it.
	if selectorCommentPattern.MatchString(selectorQuotedStringPattern.ReplaceAllString(selectorText, "")) {
		return estree.New("selector-unknown", 0, 0, "value", trim(selectorText))
	}

	result, err := selector.Parse(selectorText)
	if err != nil {
		// Fail silently. It's better to print it as is than to try and parse it
		// Note: A common failure is for SCSS nested properties. `background:
		// none { color: red; }` is parsed as a NestedDeclaration by
		// postcss-scss, while `background: { color: red; }` is parsed as a Rule
		// with a selector ending with a colon. See:
		// https://github.com/postcss/postcss-scss/issues/39
		//
		// Upstream's catch takes every throw, so every error lands here, the port's own included. A
		// selector the library never returns from (selector.ErrLoopsForever) hangs Prettier; here it is
		// printed as is.
		return estree.New("selector-unknown", 0, 0, "value", selectorText)
	}

	addTypePrefix(result, "selector-", nil)
	return result
}
