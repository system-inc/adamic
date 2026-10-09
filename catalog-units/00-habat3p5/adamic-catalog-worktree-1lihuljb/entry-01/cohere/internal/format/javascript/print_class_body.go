package javascript

import (
	"github.com/system-inc/cohere/internal/format/doc"
	"github.com/system-inc/cohere/internal/format/printing"
)

// print/class-body.js. Flow's ObjectTypeAnnotation and RecordDeclarationBody never reach a TypeScript
// tree, so their branches are dropped: the exact `{| |}` braces, the comma separators, the inexact
// `...`, and the FunctionTypeParam hug test.

/*
- `ClassBody`
- `TSInterfaceBody` (TypeScript)
- `TSTypeLiteral` (TypeScript)
- `ObjectTypeAnnotation` (Flow)
- `RecordDeclarationBody` (Flow)
*/
// printClassBody is upstream's printClassBody.
func printClassBody(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)
	var parts []any
	isObjectType := !isClassBody(path)
	separator := hardline
	if isObjectType {
		separator = line
	}
	hasDanglingComments := hasComment(current, commentDangling, nil)

	openingBrace, closingBrace := "{", "}"
	isEmpty := true
	var firstMember Node

	iterateClassMembersPath(path, func(path *Path) {
		member := node(path)
		next := nextOf(path)
		isLast := path.IsLast()
		if firstMember == nil {
			firstMember = member
		}
		isEmpty = false
		parts = append(parts, print(nil, nil))

		if !isObjectType &&
			(shouldPrintSemicolonAfterClassProperty(member, next, options) ||
				shouldPrintSemicolonAfterInterfaceProperty(member, next, options)) {
			parts = append(parts, ";")
		}

		if !isLast {
			parts = append(parts, separator)

			if isNextLineEmptyAfter(member, options) {
				parts = append(parts, hardline)
			}
		}
	})

	if hasDanglingComments {
		parts = append(parts, printing.PrintDanglingComments(path, options, printing.DanglingOptions[Node]{}))
	}

	if isObjectType {
		// objectWrap is always "preserve".
		shouldBreak := hasComment(current, commentDangling|commentLine, nil) ||
			(firstMember != nil &&
				hasNewlineInRange(
					originalText(options),
					locStart(current),
					locStart(firstMember),
				))

		var content Doc
		if len(parts) == 0 {
			content = doc.Text(openingBrace + closingBrace)
		} else {
			spacing := line
			if !settingsOf(options).BracketSpacing || (isEmpty && !shouldBreak) {
				spacing = softline
			}
			content = concatIn(path, openingBrace,
				indentIn(path, concatIn(path, append([]any{spacing}, parts...)...)),
				spacing,
				closingBrace,
			)
		}

		// If we inline the object as first argument of the parent, we don't want
		// to create another group so that the object breaks before the return
		// type
		if path.Match(
			nil,
			func(_ any, name any, _ int, _ bool) bool { return name == "typeAnnotation" },
			func(_ any, name any, _ int, _ bool) bool { return name == "typeAnnotation" },
			shouldHugTheOnlyParameterPredicate,
		) {
			return content
		}

		return groupWithIn(path, content, doc.GroupOptions{ShouldBreak: shouldBreak})
	}

	var body Doc = emptyDoc
	if len(parts) > 0 {
		body = concatIn(path, indentIn(path, concatIn(path, hardline, concatIn(path, parts...))), hardline)
	}
	return concatIn(path, openingBrace,
		body,
		closingBrace,
	)
}

// isClassBody is upstream's isClassBody. The ObjectTypeAnnotation branch, Flow's interface and declare
// class bodies, is dropped.
func isClassBody(path *Path) bool {
	current := node(path)

	return current.Is("ClassBody", "TSInterfaceBody", "RecordDeclarationBody")
}

// printClassMemberSemicolon is upstream's printClassMemberSemicolon. Upstream prints a bare ";" when
// the parent is Flow's ObjectTypeAnnotation; that branch is dropped.
func printClassMemberSemicolon(path *Path, options *Options) Doc {
	parent := parentOf(path)

	if callParent(path, isClassBody, 0) {
		return printSemicolon(options)
	}

	if parent.Is("TSTypeLiteral") {
		if path.IsLast() {
			if settingsOf(options).Semi {
				return ifBreak(";", nil)
			}
			return emptyDoc
		}

		if settingsOf(options).Semi ||
			shouldPrintSemicolonAfterInterfaceProperty(node(path), nextOf(path), options) {
			return doc.Text(";")
		}

		return ifBreak("", ";")
	}

	return emptyDoc
}

// isClassProperty is upstream's isClassProperty.
func isClassProperty(node Node) bool {
	return node.Is("ClassProperty", "PropertyDefinition", "ClassPrivateProperty", "ClassAccessorProperty",
		"AccessorProperty", "TSAbstractPropertyDefinition", "TSAbstractAccessorProperty")
}

// isKeywordProperty is upstream's isKeywordProperty.
func isKeywordProperty(node Node) bool {
	if node.Truthy("computed") || node.Child("typeAnnotation") != nil {
		return false
	}

	key := node.Child("key")
	name := key.String("name")
	return key.Is("Identifier") &&
		(name == "static" || name == "get" || name == "set")
}

// shouldPrintSemicolonAfterClassProperty is upstream's shouldPrintSemicolonAfterClassProperty. Upstream
// takes { node, next }; the Go takes the two nodes.
func shouldPrintSemicolonAfterClassProperty(node Node, nextNode Node, options *Options) bool {
	if settingsOf(options).Semi || !isClassProperty(node) {
		return false
	}

	if node.Child("value") == nil && isKeywordProperty(node) {
		return true
	}

	if nextNode == nil {
		return false
	}

	if nextNode.Truthy("static") ||
		nextNode.Truthy("accessibility") || // TypeScript
		nextNode.Truthy("readonly") { // TypeScript
		return false
	}

	if !nextNode.Truthy("computed") {
		name := nextNode.Child("key").String("name")
		if name == "in" || name == "instanceof" {
			return true
		}
	}

	// Flow variance sigil +/- requires semi if there's no
	// "declare" or "static" keyword before it.
	if isClassProperty(nextNode) &&
		!nextNode.Truthy("static") &&
		nextNode.Truthy("variance") &&
		!nextNode.Truthy("declare") {
		return true
	}

	switch nextNode.Type() {
	case "ClassProperty", "PropertyDefinition", "TSAbstractPropertyDefinition":
		return nextNode.Truthy("computed")
	case "MethodDefinition", "TSAbstractMethodDefinition", "ClassMethod", "ClassPrivateMethod":
		// Babel
		value := nextNode.Child("value")
		var isAsync bool
		if value != nil {
			isAsync = value.Truthy("async")
		} else {
			isAsync = nextNode.Truthy("async")
		}
		kind := nextNode.String("kind")
		if isAsync || kind == "get" || kind == "set" {
			return false
		}

		var isGenerator bool
		if value != nil {
			isGenerator = value.Truthy("generator")
		} else {
			isGenerator = nextNode.Truthy("generator")
		}
		if nextNode.Truthy("computed") || isGenerator {
			return true
		}

		return false

	case "TSIndexSignature":
		return true
	}

	return false
}

// isInterfaceProperty is upstream's isInterfaceProperty.
func isInterfaceProperty(node Node) bool {
	return node.Is("TSPropertySignature")
}

// shouldPrintSemicolonAfterInterfaceProperty is upstream's shouldPrintSemicolonAfterInterfaceProperty.
// Upstream takes { node, next }; the Go takes the two nodes.
func shouldPrintSemicolonAfterInterfaceProperty(node Node, nextNode Node, options *Options) bool {
	if settingsOf(options).Semi || !isInterfaceProperty(node) {
		return false
	}

	if isKeywordProperty(node) {
		return true
	}

	if nextNode == nil {
		return false
	}

	switch nextNode.Type() {
	case "TSCallSignatureDeclaration":
		if node.Is("TSPropertySignature") && node.Child("typeAnnotation") != nil {
			return false
		}

		return true
	}

	return false
}
