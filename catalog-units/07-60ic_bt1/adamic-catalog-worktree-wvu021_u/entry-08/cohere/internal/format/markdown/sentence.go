package markdown

import (
	"github.com/system-inc/cohere/internal/format/doc"
	"github.com/system-inc/cohere/internal/format/printing"
)

// src/language-markdown/print/sentence.js.

// printSentence is upstream's printSentence. Upstream grows each content part by wrapping it, one
// two-element concat per word: [[["", a], b], c]. Here a part is one flat concat of the same docs, ["", a, b,
// c], which prints the same, and every part's docs share one array, so a sentence allocates a few slices
// rather than one per word (#vbjv3d6).
func printSentence(path *astPath, print printing.PrintFunc) doc.Doc {
	var parts []doc.Doc
	// contents holds every content part's docs in order. A part is contents[partStart:], closed with a
	// full slice expression, so the next part's appends never write into it.
	contents := []doc.Doc{doc.Text("")}
	partStart := 0
	closePart := func() doc.Doc {
		if len(contents)-partStart == 1 {
			return contents[partStart]
		}
		return doc.Concat(contents[partStart:len(contents):len(contents)])
	}

	path.Each(func(path *astPath, _ int, _ any) {
		node := currentNode(path)
		printed := print(nil, nil)
		if node.NodeType == "whitespace" {
			if _, isText := printed.(doc.Text); !isText {
				parts = append(parts, closePart(), printed)
				partStart = len(contents)
				contents = append(contents, doc.Text(""))
				return
			}
		}
		contents = append(contents, printed)
	}, "children")

	return doc.NewFill(append(parts, closePart()))
}
