// Package mediaquery is postcss-media-query-parser 0.2.3 (dist/), the parser Prettier's
// src/language-css/parse/parse-media-query.js runs on the params of @media and @custom-media.
//
// The tree is the library's own: types unprefixed, and a type the library could not decide left
// undefined (the node's type is ""). Prettier's addMissingType and addTypePrefix, and its fallback to
// a selector-unknown node when the library throws, belong to the language-css glue, not here.
package mediaquery

import (
	"github.com/system-inc/cohere/internal/format/estree"
	"github.com/system-inc/cohere/internal/format/printing"
)

// postcss-media-query-parser 0.2.3, dist/index.js.

/**
 * Parses a media query list into an array of nodes. A typical node signature:
 *  {string} node.type -- one of: 'media-query', 'media-type', 'keyword',
 *    'media-feature-expression', 'media-feature', 'colon', 'value'
 *  {string} node.value -- the contents of a particular element, trimmed
 *    e.g.: `screen`, `max-width`, `1024px`
 *  {string} node.after -- whitespaces that follow the element
 *  {string} node.before -- whitespaces that precede the element
 *  {string} node.sourceIndex -- the index of the element in a source media
 *    query list, 0-based
 *  {object} node.parent -- a link to the parent node (a container)
 *
 * Some nodes (media queries, media feature expressions) contain other nodes.
 * They additionally have:
 *  {array} node.nodes -- an array of nodes of the type described here
 *  {funciton} node.each -- traverses direct children of the node, calling
 *    a callback for each one
 *  {funciton} node.walk -- traverses ALL descendants of the node, calling
 *    a callback for each one
 */

// Parse is the library's default export, parseMedia(value): the media-query-list container. Where the
// library throws (or would never return) the error is marked printing.Syntax; Prettier catches it and
// prints the params as they are.
//
// sourceIndex is in bytes into params, where the library's is in UTF-16 units. Node ranges are [0, 0]:
// the library records only a start, and the glue computes Prettier's positions.
func Parse(params string) (*estree.Node, error) {
	nodes, err := parseMediaList(params)
	if err != nil {
		return nil, printing.Syntax(err)
	}
	// after, before and sourceIndex are undefined in the options, so Container derives them.
	return newContainer(nil, nil, "media-query-list", trim(params), nil, nodes), nil
}
