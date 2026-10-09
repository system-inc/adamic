package yaml

import (
	"strings"

	"github.com/system-inc/cohere/internal/format/doc"
	"github.com/system-inc/cohere/internal/format/printing"
	"github.com/system-inc/cohere/internal/format/yaml/unist"
)

// src/language-yaml/print/mapping-item.js.

// printMappingItem is upstream's printMappingItem, for block and flow mapping items.
func printMappingItem(path *astPath, options *options, print printing.PrintFunc) doc.Doc {
	node := currentNode(path)
	parent := parentNode(path)
	settings := settingsOf(options)
	key, value := childAt(node, 0), childAt(node, 1)

	isEmptyMappingKey := isEmptyNode(key)
	isEmptyMappingValue := isEmptyNode(value)

	if isEmptyMappingKey && isEmptyMappingValue {
		return doc.Text(": ")
	}

	printedKey := print("key", nil)
	spaceBeforeColon := doc.Text("")
	if needsSpaceInFrontOfMappingValue(node) {
		spaceBeforeColon = doc.Text(" ")
	}

	if isEmptyMappingValue {
		if node.NodeType == "flowMappingItem" && parent.NodeType == "flowMapping" {
			return printedKey
		}

		// parent.tag?.value !== "tag:yaml.org,2002:set": a missing tag is not that tag.
		if node.NodeType == "mappingItem" &&
			isAbsolutelyPrintedAsSingleLineNode(contentOf(key), options, settings) &&
			!hasTrailingComment(contentOf(key)) &&
			(parent.Tag == nil || parent.Tag.Value != "tag:yaml.org,2002:set") {
			return doc.Concat{printedKey, spaceBeforeColon, doc.Text(":")}
		}

		return doc.Concat{doc.Text("? "), alignWithSpaces(2, printedKey)}
	}

	printedValue := print("value", nil)
	if isEmptyMappingKey {
		return doc.Concat{doc.Text(": "), alignWithSpaces(2, printedValue)}
	}

	// force explicit Key
	if hasLeadingComments(value) || !isInlineNode(contentOf(key)) {
		parts := doc.Concat{doc.Text("? "), alignWithSpaces(2, printedKey), doc.Hardline}
		for _, printed := range printing.Map(path, func(*astPath, int, any) doc.Doc {
			return doc.Concat{print(nil, nil), doc.Hardline}
		}, "value", "leadingComments") {
			parts = append(parts, printed)
		}
		return append(parts, doc.Text(": "), alignWithSpaces(2, printedValue))
	}

	// force singleline
	if isSingleLineNode(contentOf(key)) &&
		!hasLeadingComments(contentOf(key)) &&
		!hasMiddleComments(contentOf(key)) &&
		!hasTrailingComment(contentOf(key)) &&
		!hasEndComments(key) &&
		!hasLeadingComments(contentOf(value)) &&
		!hasMiddleComments(contentOf(value)) &&
		!hasEndComments(value) &&
		isAbsolutelyPrintedAsSingleLineNode(contentOf(value), options, settings) &&
		isAbsolutelyPrintedAsSingleLineNode(contentOf(key), options, settings) {
		return doc.Concat{printedKey, spaceBeforeColon, doc.Text(": "), printedValue}
	}

	groupID := doc.NewGroupID("mappingKey")
	groupedKey := doc.NewGroup(doc.Concat{
		doc.NewIfBreak(doc.Text("? "), nil, nil),
		doc.NewGroup(alignWithSpaces(2, printedKey), doc.GroupOptions{ID: groupID}),
	}, doc.GroupOptions{})

	// Construct both explicit and implicit mapping values.
	explicitMappingValue := doc.Concat{
		doc.Hardline,
		doc.Text(": "),
		alignWithSpaces(2, printedValue),
	}
	// In the implicit case, it's convenient to treat everything from the key's colon
	// as part of the mapping value
	implicitMappingValueParts := doc.Concat{spaceBeforeColon, doc.Text(":")}
	valueContent := contentOf(value)
	switch {
	case hasEndComments(value) &&
		valueContent != nil &&
		isNode(valueContent, "flowMapping", "flowSequence") &&
		len(valueContent.Children) == 0:
		implicitMappingValueParts = append(implicitMappingValueParts, doc.Text(" "))
	case hasLeadingComments(valueContent) ||
		(hasEndComments(value) &&
			valueContent != nil &&
			!isNode(valueContent, "mapping", "sequence")) ||
		(parent.NodeType == "mapping" &&
			hasTrailingComment(contentOf(key)) &&
			isInlineNode(valueContent)) ||
		// value.content.tag === null: a mapping or sequence always has the field, null when absent.
		(isNode(valueContent, "mapping", "sequence") &&
			valueContent.Tag == nil &&
			valueContent.Anchor == nil):
		implicitMappingValueParts = append(implicitMappingValueParts, doc.Hardline)
	case valueContent != nil:
		implicitMappingValueParts = append(implicitMappingValueParts, doc.LineDoc)
	case hasTrailingComment(value):
		implicitMappingValueParts = append(implicitMappingValueParts, doc.Text(" "))
	}
	implicitMappingValueParts = append(implicitMappingValueParts, printedValue)
	implicitMappingValue := alignWithSpaces(settings.tabWidth, implicitMappingValueParts)

	// If a key is definitely single-line, forcibly use implicit style to avoid edge cases (very long
	// keys) that would otherwise trigger explicit style as if it was multiline.
	// In those cases, explicit style makes the line even longer and causes confusion.
	if isAbsolutelyPrintedAsSingleLineNode(contentOf(key), options, settings) &&
		!hasLeadingComments(contentOf(key)) &&
		!hasMiddleComments(contentOf(key)) &&
		!hasTrailingComment(contentOf(key)) &&
		!hasEndComments(key) {
		return doc.ConditionalGroup([]doc.Doc{doc.Concat{printedKey, implicitMappingValue}}, doc.GroupOptions{})
	}

	// Use explicit mapping syntax if the key breaks, implicit otherwise
	return doc.ConditionalGroup([]doc.Doc{
		doc.Concat{
			groupedKey,
			doc.NewIfBreak(explicitMappingValue, implicitMappingValue, groupID),
		},
	}, doc.GroupOptions{})
}

// isAbsolutelyPrintedAsSingleLineNode is upstream's isAbsolutelyPrintedAsSingleLineNode.
func isAbsolutelyPrintedAsSingleLineNode(node *unist.Node, options *options, settings *settings) bool {
	if node == nil {
		return true
	}

	switch node.NodeType {
	case "plain", "quoteSingle", "quoteDouble":
	case "alias":
		return true
	default:
		return false
	}

	if settings.proseWrap == "preserve" {
		return node.Position.Start.Line == node.Position.End.Line
	}

	if
	// backslash-newline
	hasBackslashAtLineEnd(options.OriginalText[node.Position.Start.Offset:node.Position.End.Offset]) {
		return false
	}

	switch settings.proseWrap {
	case "never":
		return !strings.Contains(node.Value, "\n")
	case "always":
		return !strings.ContainsAny(node.Value, "\n ")
	default:
		return false
	}
}

// hasBackslashAtLineEnd is /\\$/m.test(text): a backslash before the end of the text or before a line
// terminator, which for JavaScript's m flag is \n, \r, U+2028 or U+2029.
func hasBackslashAtLineEnd(text string) bool {
	for index := 0; index < len(text); index++ {
		if text[index] != '\\' {
			continue
		}
		rest := text[index+1:]
		if rest == "" || rest[0] == '\n' || rest[0] == '\r' ||
			strings.HasPrefix(rest, "\u2028") || strings.HasPrefix(rest, "\u2029") {
			return true
		}
	}
	return false
}

// needsSpaceInFrontOfMappingValue is upstream's needsSpaceInFrontOfMappingValue: an alias key needs a
// space before its colon, which would otherwise read as part of the alias.
func needsSpaceInFrontOfMappingValue(node *unist.Node) bool {
	keyContent := contentOf(childAt(node, 0))
	return keyContent != nil && keyContent.NodeType == "alias"
}

// isSingleLineNode is upstream's isSingleLineNode.
func isSingleLineNode(node *unist.Node) bool {
	if node == nil {
		return true
	}

	switch node.NodeType {
	case "plain", "quoteDouble", "quoteSingle":
		return node.Position.Start.Line == node.Position.End.Line
	case "alias":
		return true
	default:
		return false
	}
}
