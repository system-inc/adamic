package postcss

import "github.com/system-inc/cohere/internal/format/estree"

// postcss-scss 4.0.9, lib/scss-parse.js, called the way Prettier's parser-postcss.js parseScss calls it:
// parseWithParser(postcssScssParse, text) runs postcssScssParse(text, { map: false }), so, as for Parse,
// Input has no `from` and no map.

// ParseSCSS is postcss-scss's parse: the Root of the tree ScssParser builds, or the CssSyntaxError it
// threw, marked with printing.Syntax; a TypeError upstream throws (see TypeError) is marked the same way.
//
// The tree is Parse's, with what postcss-scss adds:
//
//   - a `//` comment is a comment with raws.inline true, its text with any "/*" or "*/" rewritten as
//     "*//*", and raws.text, the text as written, where it is not blank;
//   - a nested property with a value (`margin: 0 { left: 1px }`) is a "decl" with isNested true and
//     nodes, its fields in NestedDeclaration's order (isNested, nodes, source, prop, important, value);
//   - raws.selector, raws.params and raws.value, where Parser.raw keeps them, carry scss, the text with
//     its `//` comments as written, when raw (with those comments as /* */) differs.
func ParseSCSS(text string) (*estree.Node, error) {
	return parseWith(text, newScssParser)
}
