package postcss

import (
	"errors"
	"fmt"

	"github.com/system-inc/cohere/internal/format/estree"
	"github.com/system-inc/cohere/internal/format/printing"
)

// postcss 8.5.16, lib/parse.js, called the way Prettier's parser-postcss.js parseWithParser calls it:
// postcssParse.default(text, { map: false }), with no `from`, so the hint parse appends to a
// CssSyntaxError's message for a .scss, .sass or .less file never applies.

// Parse is postcss's parse: the Root of the tree, or the CssSyntaxError the parser threw, marked as a
// parse failure with printing.Syntax. Recover with errors.As(err, new(*CssSyntaxError)) for the name,
// reason, line and column Prettier builds its error from.
//
// The tree is postcss's, as estree.Nodes. A node's type is postcss's type ("root", "rule", "atrule",
// "decl", "comment") and its properties are the object's own enumerable fields in the order postcss
// creates them, leaving out parent and source.input:
//
//   - raws, a map[string]any: before, after, between, semicolon, afterName, important, ownSemicolon,
//     left, right, and selector, params or value as {raw, value} maps where Parser.raw kept the source
//     text apart from the clean value.
//   - nodes, a []*estree.Node, on the root, rules and at-rules with a block.
//   - source, a map[string]any of start and end, each {offset, line, column}: offset in bytes into the
//     text after any byte order mark, line and column postcss's (column one-based, in UTF-16 units). An
//     at-rule ending the file with nothing after its name has no end, as in postcss.
//   - selector; name and params; prop, important and value; text.
//
// A node's Range is [source.start.offset, source.end.offset] in bytes where postcss set both, and
// [0, 0] where it set no end. Prettier's positions are loc.js's, which the glue computes from source.
func Parse(text string) (*estree.Node, error) {
	return parseWith(text, newParser)
}

// parseWith runs a parser, postcss's or postcss-scss's, over the text: new Input(text), new Parser(input),
// parser.parse(), parser.root.
func parseWith(text string, construct func(in *input) *parser) (root *estree.Node, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			root = nil
			if syntaxError, isSyntaxError := recovered.(*CssSyntaxError); isSyntaxError {
				// Marked as a parse failure, so the format phase skips the file rather than reporting a
				// broken formatter.
				err = printing.Syntax(syntaxError)
				return
			}
			if typeError, isTypeError := recovered.(*TypeError); isTypeError {
				// postcss-scss reaching past its tokens, a TypeError upstream that Prettier lets escape: the
				// file does not parse, so it is marked the same way.
				err = printing.Syntax(typeError)
				return
			}
			// Anything else is a bug in the port, where upstream would have thrown a TypeError; it is
			// reported, not marked as a syntax error, and never escapes as a panic.
			err = errors.New(fmt.Sprint("postcss: ", recovered))
		}
	}()

	in := newInput(text)
	p := construct(in)
	p.parse()

	return toEstree(p.root, in), nil
}

// toEstree hands a node out as an estree.Node, offsets crossing from UTF-16 units to bytes.
func toEstree(each *node, in *input) *estree.Node {
	start, end := 0, 0
	if each.source != nil && each.source.start != nil && each.source.end != nil {
		start = in.byteOffset(each.source.start.offset)
		end = in.byteOffset(each.source.end.offset)
	}
	result := estree.New(each.nodeType, start, end, "raws", rawsToMap(each.raws))
	for _, key := range each.keys {
		switch key {
		case "nodes":
			children := make([]*estree.Node, len(each.nodes))
			for index, child := range each.nodes {
				children[index] = toEstree(child, in)
			}
			result.Set("nodes", children)
		case "source":
			result.Set("source", sourceToMap(each.source, in))
		case "selector":
			result.Set("selector", each.selector)
		case "name":
			result.Set("name", each.name)
		case "params":
			result.Set("params", each.params)
		case "prop":
			result.Set("prop", each.prop)
		case "important":
			result.Set("important", each.important)
		case "value":
			result.Set("value", each.value)
		case "text":
			result.Set("text", each.text)
		case "isNested":
			result.Set("isNested", each.isNested)
		default:
			panic("postcss: no field " + key)
		}
	}
	return result
}

func rawsToMap(raws map[string]any) map[string]any {
	result := make(map[string]any, len(raws))
	for key, value := range raws {
		if raw, isRaw := value.(*rawValue); isRaw {
			rawMap := map[string]any{"raw": raw.raw, "value": raw.value}
			if raw.hasScss {
				rawMap["scss"] = raw.scss
			}
			result[key] = rawMap
			continue
		}
		result[key] = value
	}
	return result
}

func sourceToMap(each *source, in *input) map[string]any {
	result := map[string]any{"start": positionToMap(each.start, in)}
	if each.end != nil {
		result["end"] = positionToMap(each.end, in)
	}
	return result
}

// positionToMap is a position with its offset in bytes; column stays in UTF-16 units, as postcss counts.
func positionToMap(each *position, in *input) map[string]any {
	return map[string]any{
		"column": each.column,
		"line":   each.line,
		"offset": in.byteOffset(each.offset),
	}
}
