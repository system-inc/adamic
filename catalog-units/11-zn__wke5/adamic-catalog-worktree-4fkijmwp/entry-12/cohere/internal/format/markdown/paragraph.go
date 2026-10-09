package markdown

import (
	"github.com/system-inc/cohere/internal/format/doc"
	"github.com/system-inc/cohere/internal/format/printing"
)

// src/language-markdown/print/paragraph.js.

func printParagraph(path *astPath, print printing.PrintFunc) doc.Doc {
	parts := printing.Map(path, func(*astPath, int, any) doc.Doc { return print(nil, nil) }, "children")
	return flattenFill(parts)
}

// flattenFill joins the children's fills into one. We assume parts always meet following conditions:
// parts.length is odd, and odd elements are line-like docs that come from odd elements of inner fills.
func flattenFill(docs []doc.Doc) doc.Doc {
	parts := []doc.Doc{doc.Text("")}

	var rec func(docArray []doc.Doc)
	rec = func(docArray []doc.Doc) {
		for _, document := range docArray {
			if array, isArray := doc.Parts(document); isArray {
				rec(array)
				continue
			}

			head := document
			var rest []doc.Doc
			if fill, isFill := document.(*doc.Fill); isFill {
				head = fill.Parts[0]
				rest = fill.Parts[1:]
			}

			parts[len(parts)-1] = doc.Concat{parts[len(parts)-1], head}
			parts = append(parts, rest...)
		}
	}
	rec(docs)

	return doc.NewFill(parts)
}
