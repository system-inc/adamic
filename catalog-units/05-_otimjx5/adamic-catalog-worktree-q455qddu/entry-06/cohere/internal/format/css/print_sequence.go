package css

// src/language-css/print/sequence.js

import (
	"github.com/system-inc/cohere/internal/format/doc"
	"github.com/system-inc/cohere/internal/format/estree"
	"github.com/system-inc/cohere/internal/format/printing"
)

func printSequence(path *astPath, options *printerOptions, print printing.PrintFunc) doc.Doc {
	parts := doc.Concat{}
	path.Each(func(path *astPath, _ int, _ any) {
		node := currentNode(path)
		previous, _ := path.Previous()
		if previous.Type() == "css-comment" && trim(previous.String("text")) == "prettier-ignore" {
			parts = append(parts, doc.Text(options.OriginalText[locStart(node):locEnd(node)]))
		} else {
			parts = append(parts, print(nil, nil))
		}

		if path.IsLast() {
			return
		}

		next, _ := path.Next()
		if (next.Type() == "css-comment" &&
			!printing.HasNewline(options.OriginalText, locStart(next), true) &&
			!isFrontMatter(node)) ||
			(next.Type() == "css-atrule" &&
				next.String("name") == "else" &&
				node.Type() != "css-comment") {
			parts = append(parts, doc.Text(" "))
		} else {
			// options.__isHTMLStyleAttribute is an embed-only option, false here, so this is hardline.
			parts = append(parts, doc.Hardline)
			if isNextLineEmpty(options.OriginalText, locEnd(node)) && !isFrontMatter(node) {
				parts = append(parts, doc.Hardline)
			}
		}
	}, "nodes")

	return parts
}

// isFrontMatter is src/main/front-matter/is-front-matter.js. Upstream tests a symbol the front matter
// parser sets; the glue gives the front matter node the type "front-matter", which nothing else has.
func isFrontMatter(node *estree.Node) bool {
	return node.Type() == "front-matter"
}
