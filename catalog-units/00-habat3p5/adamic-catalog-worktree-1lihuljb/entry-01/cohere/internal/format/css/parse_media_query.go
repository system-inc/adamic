package css

// src/language-css/parse/parse-media-query.js.

import (
	"github.com/system-inc/cohere/internal/format/css/mediaquery"
	"github.com/system-inc/cohere/internal/format/estree"
)

func parseMediaQuery(params string) *estree.Node {
	result, err := mediaquery.Parse(params)
	if err != nil {
		// Ignore bad media queries
		/* c8 ignore next 4 */
		return estree.New("selector-unknown", 0, 0, "value", params)
	}

	addTypePrefix(addMissingType(result), "media-", nil)
	return result
}
