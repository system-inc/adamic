// A port of cohere's internal/format/css/mediaquery/index.go to Adamic 0.1: postcss-media-query-parser
// 0.2.3, dist/index.js, the parser Prettier's src/language-css/parse/parse-media-query.js runs on the
// params of @media and @custom-media.
//
// The tree is the library's own: types unprefixed, and a type the library could not decide left
// undefined (the node's type is ''). Prettier's addMissingType and addTypePrefix, and its fallback to a
// selector-unknown node when the library throws, belong to the language-css glue, not here.

import type { MediaNode } from './nodes.ts';
import { newContainer } from './nodes.ts';
import { parseMediaList, type Result } from './parsers.ts';
import { trim } from './whitespace.ts';

/**
 * index.go: Parse.
 *
 * Parses a media query list into an array of nodes. A typical node signature:
 *  {string} node.type -- one of: 'media-query', 'media-type', 'keyword',
 *    'media-feature-expression', 'media-feature', 'colon', 'value'
 *  {string} node.value -- the contents of a particular element, trimmed
 *    e.g.: `screen`, `max-width`, `1024px`
 *  {string} node.after -- whitespaces that follow the element
 *  {string} node.before -- whitespaces that precede the element
 *  {string} node.sourceIndex -- the index of the element in a source media
 *    query list, 0-based
 *
 * The library's default export, parseMedia(value): the media-query-list container. Where the library
 * throws (or would never return) the result is an Error with the Go's message; the Go marks it
 * printing.Syntax, and Prettier catches it and prints the params as they are. sourceIndex is in UTF-16
 * units, the library's own, where the Go's is in bytes.
 */
export function parse(params: string): Result<MediaNode> {
	const nodes = parseMediaList(params);
	if (nodes.kind === 'Error') {
		return nodes;
	}
	// after, before and sourceIndex are undefined in the options, so Container derives them.
	return { kind: 'Ok', value: newContainer(undefined, undefined, 'media-query-list', trim(params), undefined, nodes.value) };
}
