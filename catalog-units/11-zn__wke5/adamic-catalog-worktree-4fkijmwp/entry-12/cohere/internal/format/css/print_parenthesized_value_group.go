package css

// src/language-css/print/parenthesized-value-group.js

import (
	"strings"

	"github.com/system-inc/cohere/internal/format/doc"
	"github.com/system-inc/cohere/internal/format/estree"
	"github.com/system-inc/cohere/internal/format/printing"
)

func hasComma(path *astPath, options *printerOptions) bool {
	node := currentNode(path)
	parent, _ := path.Parent()
	return node.Truthy("source") &&
		strings.HasSuffix(trimEnd(sliceText(options.OriginalText, locStart(node), locStart(parent.Child("close")))), ",")
}

func printTrailingComma(path *astPath, options *printerOptions) doc.Doc {
	grandparent, _ := path.Grandparent()
	if isVarFunctionNode(grandparent) && hasComma(path, options) {
		return doc.Text(",")
	}

	node := currentNode(path)
	if node.Type() != "value-comment" &&
		!(node.Type() == "value-comma_group" &&
			everyNode(node.List("groups"), func(group *estree.Node) bool { return group.Type() == "value-comment" })) &&
		shouldPrintTrailingComma(options) &&
		printing.CallParent(path, func(path *astPath) bool { return isSCSSMapItemNode(path, options) }, 0) {
		return doc.NewIfBreak(doc.Text(","), nil, nil)
	}

	return doc.Text("")
}

func printParenthesizedValueGroup(path *astPath, options *printerOptions, print printing.PrintFunc) doc.Doc {
	node := currentNode(path)
	parent, hasParent := path.Parent()
	// A url() argument the value parser cannot handle is a plain string among the groups
	// (parse-value.js), which is why groups are walked as values here and not as nodes.
	groupDocs := printing.Map(path, func(path *astPath, _ int, _ any) doc.Doc {
		if text, isString := path.Value().(string); isString {
			return doc.Text(text)
		}
		return print(nil, nil)
	}, "groups")
	groupCount := len(groupDocs)
	groups := node.List("groups")

	if hasParent &&
		isURLFunctionNode(parent) &&
		(groupCount == 1 ||
			(groupCount > 0 &&
				len(groups) > 0 &&
				groups[0].Type() == "value-comma_group" &&
				len(groups[0].List("groups")) > 0 &&
				groups[0].List("groups")[0].Type() == "value-word" &&
				strings.HasPrefix(groups[0].List("groups")[0].String("value"), "data:"))) {
		return doc.Concat{
			choose(node.Child("open") != nil, func() doc.Doc { return print("open", nil) }),
			doc.Join(doc.Text(","), groupDocs),
			choose(node.Child("close") != nil, func() doc.Doc { return print("close", nil) }),
		}
	}

	if node.Child("open") == nil {
		forceHardLine := shouldBreakList(path)
		withComma := chunk(doc.Join(doc.Text(","), groupDocs), 2)
		var separator doc.Doc = doc.LineDoc
		if forceHardLine {
			separator = doc.Hardline
		}
		parts := doc.Join(separator, withComma)
		if forceHardLine {
			return doc.NewIndent(doc.Concat{doc.Hardline, parts})
		}
		var leading doc.Doc = doc.Text("")
		if shouldPrecededBySoftline(path) {
			leading = doc.Softline
		}
		return doc.NewIndent(doc.NewGroup(doc.Concat{leading, doc.NewFill(parts)}, doc.GroupOptions{}))
	}

	parts := printing.Map(path, func(path *astPath, index int, _ any) doc.Doc {
		child := currentNode(path)
		isLast := path.IsLast()
		groupDoc := groupDocs[index]

		// Key/Value pair in open paren already indented
		childGroups := child.List("groups")
		if isKeyValuePairNode(child) &&
			child.Type() == "value-comma_group" &&
			childGroups != nil &&
			childGroups[0].Type() != "value-paren_group" &&
			len(childGroups) > 2 && childGroups[2].Type() == "value-paren_group" {
			if outer, isGroup := groupDoc.(*doc.Group); isGroup {
				if middle, isIndent := outer.Contents.(*doc.Indent); isIndent {
					if _, isFill := middle.Contents.(*doc.Fill); isFill {
						groupDoc = doc.NewGroup(doc.Dedent(groupDoc), doc.GroupOptions{})
					}
				}
			}
		}

		var separator doc.Doc = doc.Text(",")
		if isLast {
			separator = printTrailingComma(path, options)
		}
		parts := doc.Concat{groupDoc, separator}

		if !isLast && child.Type() == "value-comma_group" && len(childGroups) > 0 {
			last := childGroups[len(childGroups)-1]

			// `value-paren_group` does not have location info, but its closing parenthesis does.
			if !last.Truthy("source") && last.Child("close") != nil {
				last = last.Child("close")
			}

			if last.Truthy("source") && isNextLineEmpty(options.OriginalText, locEnd(last)) {
				parts = append(parts, doc.Hardline)
			}
		}

		return parts
	}, "groups")

	isKey := isKeyInValuePairNode(node, parent)
	isConfiguration := isConfigurationNode(node, parent)
	isSCSSMapItem := isSCSSMapItemNode(path, options)
	shouldBreak := isConfiguration || (isSCSSMapItem && !isKey)
	shouldDedent := isConfiguration || isKey

	printed := doc.NewGroup(
		doc.Concat{
			choose(node.Child("open") != nil, func() doc.Doc { return print("open", nil) }),
			doc.NewIndent(doc.Concat{doc.Softline, doc.Join(doc.LineDoc, parts)}),
			doc.Softline,
			doc.LineSuffixBoundary,
			choose(node.Child("close") != nil, func() doc.Doc { return print("close", nil) }),
		},
		doc.GroupOptions{ShouldBreak: shouldBreak},
	)

	if shouldDedent {
		return doc.Dedent(printed)
	}
	return printed
}

func shouldBreakList(path *astPath) bool {
	return path.Match(
		func(node any, _ any, _ int, _ bool) bool {
			value := asNode(node)
			return value.Type() == "value-paren_group" &&
				value.Child("open") == nil &&
				someNode(value.List("groups"), func(group *estree.Node) bool { return group.Type() == "value-comma_group" })
		},
		func(node any, key any, _ int, _ bool) bool {
			return key == "group" && asNode(node).Type() == "value-value"
		},
		func(node any, key any, _ int, _ bool) bool {
			return key == "group" && asNode(node).Type() == "value-root"
		},
		func(node any, key any, _ int, _ bool) bool {
			value := asNode(node)
			return key == "value" &&
				((value.Type() == "css-decl" && !strings.HasPrefix(value.String("prop"), "--")) ||
					(value.Type() == "css-atrule" && value.Truthy("variable")))
		},
	)
}

func shouldPrecededBySoftline(path *astPath) bool {
	return path.Match(
		func(node any, _ any, _ int, _ bool) bool {
			value := asNode(node)
			return value.Type() == "value-paren_group" && value.Child("open") == nil
		},
		func(node any, key any, _ int, _ bool) bool {
			return key == "group" && asNode(node).Type() == "value-value"
		},
		func(node any, key any, _ int, _ bool) bool {
			return key == "group" && asNode(node).Type() == "value-root"
		},
		func(node any, key any, _ int, _ bool) bool {
			return key == "value" && asNode(node).Type() == "css-decl"
		},
	)
}

// chunk splits a joined doc array into arrays of size elements, as upstream's chunk does.
func chunk(array doc.Concat, size int) []doc.Doc {
	var result []doc.Doc
	for index := 0; index < len(array); index += size {
		result = append(result, append(doc.Concat(nil), array[index:min(index+size, len(array))]...))
	}
	return result
}
