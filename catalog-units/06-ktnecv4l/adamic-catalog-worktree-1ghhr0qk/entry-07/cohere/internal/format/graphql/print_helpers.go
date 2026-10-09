package graphql

// The five helpers in src/language-graphql/print/, one section each, plus the utilities they and
// printer-graphql.js import (is-non-empty-array.js, is-next-line-empty.js) and the doc builders spelled
// the way upstream spells them.

import (
	"github.com/system-inc/cohere/internal/format/doc"
	"github.com/system-inc/cohere/internal/format/estree"
	"github.com/system-inc/cohere/internal/format/printing"
)

// group, indent, ifBreak, join and the lines are src/document/builders.js.
func group(contents ...doc.Doc) doc.Doc {
	return doc.NewGroup(doc.Concat(contents), doc.GroupOptions{})
}
func indent(contents ...doc.Doc) doc.Doc {
	return doc.NewIndent(doc.Concat(contents))
}
func ifBreak(breakContents doc.Doc, flatContents doc.Doc) doc.Doc {
	return doc.NewIfBreak(breakContents, flatContents, nil)
}

// join is upstream's join, which returns an array; a Concat here, so a caller can spread it.
func join(separator doc.Doc, parts []doc.Doc) doc.Concat {
	return doc.Join(separator, parts)
}

var (
	hardline = doc.Hardline
	line     = doc.LineDoc
	softline = doc.Softline
	empty    = doc.Text("")
)

// text is a string doc.
func text(value string) doc.Doc { return doc.Text(value) }

// isNonEmptyArray is utilities/is-non-empty-array.js on a node's list property.
func isNonEmptyArray(node *estree.Node, key string) bool {
	return len(node.List(key)) > 0
}

// currentNode is upstream's path.node.
func currentNode(path *astPath) *estree.Node {
	node, _ := path.Node()
	return node
}

// printAll is upstream's path.map(print, ...names).
func printAll(path *astPath, print printing.PrintFunc, names ...any) []doc.Doc {
	return printing.Map(path, func(*astPath, int, any) doc.Doc { return print(nil, nil) }, names...)
}

// isNextLineEmpty is utilities/is-next-line-empty.js. It skips JavaScript's comment syntax, not
// GraphQL's, exactly as upstream does for every language.
func isNextLineEmpty(text string, startIndex int) bool {
	// upstream's `let oldIdx = null`: a value no index can equal.
	oldIndex, hasOldIndex := 0, false
	index := startIndex
	for !hasOldIndex || index != oldIndex {
		// We need to skip all the potential trailing inline comments
		oldIndex, hasOldIndex = index, true
		index = printing.SkipToLineEnd(text, index, false)
		index = printing.SkipInlineComment(text, index)
		index = printing.SkipSpaces(text, index, false)
	}
	index = printing.SkipTrailingComment(text, index)
	index = printing.SkipNewline(text, index, false)
	// upstream's `idx !== false`: the skip functions' false is a negative position here other than -1.
	return (index >= -1) && printing.HasNewline(text, index, false)
}

// print/arguments.js

func printArguments(path *astPath, options *printerOptions, print printing.PrintFunc) doc.Doc {
	node := currentNode(path)
	if !isNonEmptyArray(node, "arguments") {
		return empty
	}

	return group(
		text("("),
		indent(
			softline,
			join(
				doc.Concat{ifBreak(empty, text(", ")), softline},
				printSequence(path, options, print, "arguments"),
			),
		),
		softline,
		text(")"),
	)
}

// print/description.js

func printDescription(path *astPath, options *printerOptions, print printing.PrintFunc) doc.Doc {
	node := currentNode(path)
	if node.Child("description") == nil {
		return empty
	}

	parts := doc.Concat{print("description", nil)}
	if node.Is("InputValueDefinition") && !node.Child("description").Bool("block") {
		parts = append(parts, line)
	} else {
		parts = append(parts, hardline)
	}

	return parts
}

// print/directives.js

func printDirectives(path *astPath, print printing.PrintFunc) doc.Doc {
	node := currentNode(path)

	if !isNonEmptyArray(node, "directives") {
		return empty
	}

	printed := join(line, printAll(path, print, "directives"))

	if node.Is("FragmentDefinition") ||
		node.Is("OperationDefinition") {
		return group(line, printed)
	}

	return doc.Concat{text(" "), group(indent(softline, printed))}
}

// print/sequence.js

func printSequence(path *astPath, options *printerOptions, print printing.PrintFunc, property string) []doc.Doc {
	return printing.Map(path, func(path *astPath, _ int, _ any) doc.Doc {
		printed := print(nil, nil)

		if !path.IsLast() && isNextLineEmpty(options.OriginalText, locEnd(currentNode(path))) {
			return doc.Concat{printed, hardline}
		}

		return printed
	}, property)
}

// print/variable-definitions.js

func printVariableDefinitions(path *astPath, print printing.PrintFunc) doc.Doc {
	node := currentNode(path)
	if !isNonEmptyArray(node, "variableDefinitions") {
		return empty
	}
	return group(
		text("("),
		indent(
			softline,
			join(
				doc.Concat{ifBreak(empty, text(", ")), softline},
				printAll(path, print, "variableDefinitions"),
			),
		),
		softline,
		text(")"),
	)
}
