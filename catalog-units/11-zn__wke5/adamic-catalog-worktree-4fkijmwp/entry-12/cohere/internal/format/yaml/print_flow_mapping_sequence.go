package yaml

import (
	"github.com/system-inc/cohere/internal/format/doc"
	"github.com/system-inc/cohere/internal/format/printing"
)

// src/language-yaml/print/flow-mapping-sequence.js. Upstream exports printFlowMapping twice, the second
// time as printFlowSequence; here both node types call printFlowMapping.

// printFlowMapping is upstream's printFlowMapping, for flow mappings and flow sequences alike.
func printFlowMapping(path *astPath, options *options, print printing.PrintFunc) doc.Doc {
	node := currentNode(path)
	settings := settingsOf(options)
	isMapping := node.NodeType == "flowMapping"
	openMarker, closeMarker := "[", "]"
	if isMapping {
		openMarker, closeMarker = "{", "}"
	}

	bracketSpacing := doc.Softline
	if isMapping && len(node.Children) > 0 && settings.bracketSpacing {
		bracketSpacing = doc.LineDoc
	}
	lastItem := childAt(node, len(node.Children)-1)
	isLastItemEmptyMappingItem := lastItem != nil && lastItem.NodeType == "flowMappingItem" &&
		isEmptyNode(childAt(lastItem, 0)) &&
		isEmptyNode(childAt(lastItem, 1))

	// Upstream builds this array in order, so the children print before the end comments. Printing a
	// comment records its end offset for printNextEmptyLine, so the order is kept.
	contents := doc.Concat{bracketSpacing, printChildren(path, options, print)}
	if settings.trailingComma == "none" {
		contents = append(contents, doc.Text(""))
	} else {
		contents = append(contents, doc.NewIfBreak(doc.Text(","), nil, nil))
	}
	if hasEndComments(node) {
		contents = append(contents, doc.Concat{doc.Hardline, doc.Join(doc.Hardline, mapPrint(path, print, "endComments"))})
	} else {
		contents = append(contents, doc.Text(""))
	}

	closingSpacing := bracketSpacing
	if isLastItemEmptyMappingItem {
		closingSpacing = doc.Text("")
	}
	return doc.Concat{
		doc.Text(openMarker),
		alignWithSpaces(settings.tabWidth, contents),
		closingSpacing,
		doc.Text(closeMarker),
	}
}

// printChildren is upstream's printChildren.
func printChildren(path *astPath, options *options, print printing.PrintFunc) doc.Doc {
	settings := settingsOf(options)
	return doc.Concat(printing.Map(path, func(path *astPath, _ int, _ any) doc.Doc {
		printed := print(nil, nil)
		if path.IsLast() {
			return doc.Concat{printed, doc.Text("")}
		}
		node := currentNode(path)
		next := nextNode(path)
		var emptyLine doc.Doc = doc.Text("")
		if node.Position.Start.Line != next.Position.Start.Line {
			emptyLine = printNextEmptyLine(path, options.OriginalText, settings)
		}
		return doc.Concat{printed, doc.Concat{doc.Text(","), doc.LineDoc, emptyLine}}
	}, "children"))
}
