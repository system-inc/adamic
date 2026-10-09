package javascript

import (
	"github.com/system-inc/cohere/internal/format/doc"
)

// print/object.js.

// isPrintingImportAttributes is upstream's isPrintingImportAttributes.
func isPrintingImportAttributes(node Node) bool {
	return node.Is(
		"ImportDeclaration",
		"ExportDefaultDeclaration",
		"ExportNamedDeclaration",
		"ExportAllDeclaration",
		"DeclareExportDeclaration",
		"DeclareExportAllDeclaration",
	)
}

/*
- `ObjectExpression`
- `ObjectPattern`
- `ImportDeclaration`
- `ExportDefaultDeclaration`
- `ExportNamedDeclaration`
- `ExportAllDeclaration`
- `TSEnumDeclaration`(TypeScript)

Upstream's Flow enum bodies (`EnumBody` and friends) and their `hasUnknownMembers` branch never reach a
TypeScript tree, so isFlowEnumBody and hasUnknownMembers are always false here and the `...` member
printing is dropped.
*/

// printObject is upstream's printObject.
func printObject(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)
	parent := parentOf(path)

	isEnumBody := current.Is("TSEnumBody")
	isImportAttributes := isPrintingImportAttributes(current)

	property := "properties"
	if isEnumBody {
		property = "members"
	} else if isImportAttributes {
		property = "attributes"
	}
	children := current.List(property)

	shouldBreak := isEnumBody ||
		(current.Is("ObjectPattern") &&
			!parent.Is("FunctionDeclaration") &&
			!parent.Is("FunctionExpression") &&
			!parent.Is("ArrowFunctionExpression") &&
			!parent.Is("ObjectMethod") &&
			!parent.Is("ClassMethod") &&
			!parent.Is("ClassPrivateMethod") &&
			!parent.Is("AssignmentPattern") &&
			!parent.Is("CatchClause") &&
			someNode(current.List("properties"), func(property Node) bool {
				return property.Child("value") != nil &&
					(property.Child("value").Is("ObjectPattern") ||
						property.Child("value").Is("ArrayPattern"))
			})) ||
		// options.objectWrap is always "preserve" in our configs.
		(!current.Is("ObjectPattern") &&
			len(children) > 0 &&
			hasNewLineAfterOpeningBrace(current, children[0], options))

	var separatorParts []any
	parts := mapPath(path, func(path *Path, _ int) Doc {
		result := concatIn(path, append(append([]any{}, separatorParts...), print(nil, nil))...)
		separatorParts = []any{",", line}
		if isNextLineEmptyAfter(node(path), options) {
			separatorParts = append(separatorParts, hardline)
		}
		return result
	}, property)

	canHaveTrailingSeparator := !(len(children) > 0 && children[len(children)-1].Is("RestElement"))

	var content Doc
	if len(parts) == 0 {
		content = groupIn(path, concatIn(path, "{",
			printDanglingCommentsInList(path, options, nil),
			"}",
			printOptionalToken(path),
			printTypeAnnotationProperty(path, print),
		))
	} else {
		spacing := softline
		if settingsOf(options).BracketSpacing {
			spacing = line
		}
		var trailingSeparator Doc = emptyDoc
		if canHaveTrailingSeparator {
			trailingSeparator = printTrailingComma(options, "")
		}
		content = concatIn(path, "{",
			indentIn(path, append([]Doc{spacing}, parts...)),
			trailingSeparator,
			spacing,
			"}",
			printOptionalToken(path),
			printTypeAnnotationProperty(path, print),
		)
	}

	// If we inline the object as first argument of the parent, we don't want
	// to create another group so that the object breaks before the return
	// type
	if path.Match(
		func(value any, _ any, _ int, _ bool) bool {
			matched, _ := value.(Node)
			return matched.Is("ObjectPattern") && len(matched.List("decorators")) == 0
		},
		shouldHugTheOnlyParameterPredicate,
	) ||
		(isObjectType(current) &&
			(path.Match(
				nil,
				func(_ any, name any, _ int, _ bool) bool { return name == "typeAnnotation" },
				func(_ any, name any, _ int, _ bool) bool { return name == "typeAnnotation" },
				shouldHugTheOnlyParameterPredicate,
			) ||
				path.Match(
					nil,
					func(value any, name any, _ int, _ bool) bool {
						matched, _ := value.(Node)
						return matched.Is("FunctionTypeParam") && name == "typeAnnotation"
					},
					shouldHugTheOnlyParameterPredicate,
				))) ||
		// Assignment printing logic (printAssignment) is responsible
		// for adding a group if needed
		(!shouldBreak &&
			path.Match(
				func(value any, _ any, _ int, _ bool) bool {
					matched, _ := value.(Node)
					return matched.Is("ObjectPattern")
				},
				func(value any, _ any, _ int, _ bool) bool {
					matched, _ := value.(Node)
					return matched.Is("AssignmentExpression") || matched.Is("VariableDeclarator")
				},
			)) {
		return content
	}

	return groupWithIn(path, content, doc.GroupOptions{ShouldBreak: shouldBreak})
}

// shouldHugTheOnlyParameterPredicate adapts upstream's shouldHugTheOnlyParameter(node, name), which
// upstream passes to path.match directly, to a Match predicate.
func shouldHugTheOnlyParameterPredicate(value any, name any, _ int, _ bool) bool {
	matched, _ := value.(Node)
	key, _ := name.(string)
	return shouldHugTheOnlyParameter(matched, key)
}

// someNode is upstream's Array.prototype.some over a node list.
func someNode(nodes []Node, predicate func(Node) bool) bool {
	for _, each := range nodes {
		if predicate(each) {
			return true
		}
	}
	return false
}

// hasNewLineAfterOpeningBrace is upstream's hasNewLineAfterOpeningBrace. Upstream's development-only
// assertions on the brace's position are dropped.
func hasNewLineAfterOpeningBrace(node Node, firstProperty Node, options *Options) bool {
	text := options.OriginalText
	firstPropertyStart := locStart(firstProperty)

	openingBraceIndex := locStart(node)
	if isPrintingImportAttributes(node) {
		openingBraceIndex = lastIndexOfFrom(stripComments(options), "{", firstPropertyStart)
	}

	return hasNewlineInRange(text, openingBraceIndex, firstPropertyStart)
}
