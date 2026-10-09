package javascript

import (
	"github.com/system-inc/cohere/internal/format/doc"
	"github.com/system-inc/cohere/internal/format/printing"
)

// print/class.js. The Flow-only branches (InterfaceExtends heritage, the `variance` sigil on a class
// property) are dropped: only the Flow parsers produce them.

var getHeritageGroupId = createGroupIdMapper("heritageGroup")

// isInterface is upstream's isInterface.
func isInterface(node Node) bool {
	return node.Is("TSInterfaceDeclaration", "DeclareInterface", "InterfaceDeclaration", "InterfaceTypeAnnotation")
}

/*
- `ClassDeclaration`
- `ClassExpression`
- `DeclareClass`(flow)
- `DeclareInterface`(flow)
- `InterfaceDeclaration`(flow)
- `InterfaceTypeAnnotation`(flow)
- `RecordDeclaration`(flow)
- `TSInterfaceDeclaration`(TypeScript)
*/
// printClass is upstream's printClass.
func printClass(path *Path, options *Options, print PrintFunc) Doc {
	printed := printClassWithoutDecorators(path, options, print)

	current := node(path)
	if current.Is("ClassExpression") && len(current.List("decorators")) > 0 {
		decoratorsDoc := printDecorators(path, options, print)
		needsParens := needsParentheses(path, options)
		if needsParens {
			return concatIn(path, indentIn(path, concatIn(path, softline, decoratorsDoc, printed)), softline)
		}
		return concatIn(path, decoratorsDoc, printed)
	}

	return printed
}

// printClassWithoutDecorators is upstream's printClassWithoutDecorators.
func printClassWithoutDecorators(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)
	isPrintingInterface := isInterface(current)
	isPrintingRecord := current.Is("RecordDeclaration")

	keyword := "class"
	if isPrintingInterface {
		keyword = "interface"
	} else if isPrintingRecord {
		keyword = "record"
	}

	parts := []any{printDeclareToken(path), printAbstractToken(path), keyword}

	// Keep old behaviour of extends in same line
	// If there is only on extends and there are not comments
	groupMode := shouldPrintClassInGroupMode(path)

	var partsGroup []any
	var extendsParts []any

	if !current.Is("InterfaceTypeAnnotation") {
		if current.Child("id") != nil {
			partsGroup = append(partsGroup, " ")
		}

		for _, property := range []string{"id", "typeParameters"} {
			if current.Child(property) != nil {
				// Upstream destructures { leading, trailing } from path.call; the Go returns the pair.
				comments := call(path, func(path *Path) [2]Doc {
					leading, trailing := printing.PrintCommentsSeparately(path, options, nil)
					return [2]Doc{leading, trailing}
				}, property)
				partsGroup = append(partsGroup, comments[0], print(property, nil), indentIn(path, comments[1]))
			}
		}
	}

	if current.Child("superClass") != nil {
		printed := concatIn(path, printSuperClass(path, options, print),
			print("superTypeArguments", nil),
		)
		printedWithComments := call(path, func(path *Path) Doc {
			return concatIn(path, "extends ", printing.PrintComments(path, printed, options, nil))
		}, "superClass")
		if groupMode {
			extendsParts = append(extendsParts, line, groupIn(path, printedWithComments))
		} else {
			extendsParts = append(extendsParts, " ", printedWithComments)
		}
	} else {
		extendsParts = append(extendsParts, printHeritageClauses(path, options, print, "extends"))
	}

	extendsParts = append(extendsParts,
		printHeritageClauses(path, options, print, "mixins"),
		printHeritageClauses(path, options, print, "implements"),
	)

	var heritageGroupId *doc.GroupID
	if groupMode {
		heritageGroupId = getHeritageGroupId(options, current)
		parts = append(parts,
			groupWithIn(path, concatIn(path, append(partsGroup, indentIn(path, concatIn(path, extendsParts...)))...), doc.GroupOptions{ID: heritageGroupId}),
		)
	} else {
		parts = append(parts, partsGroup...)
		parts = append(parts, extendsParts...)
	}

	/*
	  To improve visual separation between class head and body https://github.com/prettier/prettier/issues/10018
	  we introduced https://github.com/prettier/prettier/pull/10085
	  However, users complaint.
	  We decide to defer to solve the inconsistency to a major release (V4)
	  Meanwhile, we are not going to put the `{` of interface body on a new line
	  https://github.com/prettier/prettier/issues/18115
	*/
	if !isPrintingInterface && groupMode && isNonEmptyClassBody(current.Child("body")) {
		parts = append(parts, ifBreakWithGroup(hardline, " ", heritageGroupId))
	} else {
		parts = append(parts, " ")
	}

	parts = append(parts, print("body", nil))

	return concatIn(path, parts...)
}

// hasMultipleHeritage is upstream's hasMultipleHeritage.
func hasMultipleHeritage(node Node) bool {
	count := 0
	if node.Child("superClass") != nil {
		count = 1
	}
	for _, listName := range []string{"extends", "mixins", "implements"} {
		count += len(node.List(listName))
		if count > 1 {
			return true
		}
	}
	return count > 1
}

// shouldPrintClassInGroupModeWithoutCache is upstream's shouldPrintClassInGroupModeWithoutCache.
func shouldPrintClassInGroupModeWithoutCache(path *Path) bool {
	current := node(path)
	if hasComment(current.Child("id"), commentTrailing, nil) ||
		hasComment(current.Child("typeParameters"), commentTrailing, nil) ||
		hasAnyComment(current.Child("superClass")) ||
		hasMultipleHeritage(current) {
		return true
	}

	if current.Child("superClass") != nil {
		if parentOf(path).Is("AssignmentExpression") {
			return false
		}

		return current.Child("superTypeArguments") == nil &&
			isMemberExpression(stripChainElementWrappers(current.Child("superClass")))
	}

	// upstream's node.extends?.[0] ?? node.mixins?.[0] ?? node.implements?.[0]
	var heritage Node
	for _, listName := range []string{"extends", "mixins", "implements"} {
		if list := current.List(listName); len(list) > 0 && list[0] != nil {
			heritage = list[0]
			break
		}
	}

	if heritage == nil {
		return false
	}

	// Upstream's first alternative, an `InterfaceExtends` with a `QualifiedTypeIdentifier` id, is
	// Flow-only and dropped.
	groupMode := (heritage.Is("TSClassImplements") ||
		heritage.Is("TSInterfaceHeritage")) &&
		isMemberExpression(heritage.Child("expression")) &&
		heritage.Child("typeArguments") == nil

	return groupMode
}

// shouldPrintClassInGroupMode is upstream's shouldPrintClassInGroupMode. Upstream memoizes it in a
// WeakMap keyed on the node; the answer reads only the node and its parent, which never change during
// a format, so the Go computes it each time.
func shouldPrintClassInGroupMode(path *Path) bool {
	return shouldPrintClassInGroupModeWithoutCache(path)
}

// printHeritageClauses is upstream's printHeritageClauses.
func printHeritageClauses(path *Path, options *Options, print PrintFunc, listName string) Doc {
	current := node(path)
	if len(current.List(listName)) == 0 {
		return emptyDoc
	}

	printedLeadingComments := printing.PrintDanglingComments(path, options, printing.DanglingOptions[Node]{
		Marker: listName,
	})

	heritageClausesDoc := join(concatIn(path, ",", line), printAll(path, print, listName))

	// Make it print like `superClass`
	if !hasMultipleHeritage(current) {
		printed := concatIn(path, listName+" ",
			printedLeadingComments,
			heritageClausesDoc,
		)
		if shouldPrintClassInGroupMode(path) {
			return concatIn(path, line, groupIn(path, printed))
		}
		return concatIn(path, " ", printed)
	}

	// upstream's `printedLeadingComments && hardline`
	var leadingCommentsBreak Doc = printedLeadingComments
	if !isEmptyString(printedLeadingComments) {
		leadingCommentsBreak = hardline
	}

	return concatIn(path, line,
		printedLeadingComments,
		leadingCommentsBreak,
		listName,
		groupIn(path, indentIn(path, concatIn(path, line, heritageClausesDoc))),
	)
}

// printSuperClass is upstream's printSuperClass.
func printSuperClass(path *Path, options *Options, print PrintFunc) Doc {
	printed := print("superClass", nil)
	parent := parentOf(path)
	if parent.Is("AssignmentExpression") {
		return groupIn(path, ifBreak(concatIn(path, "(", indentIn(path, concatIn(path, softline, printed)), softline, ")"), printed))
	}
	return printed
}

// printClassMethod is upstream's printClassMethod.
func printClassMethod(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)
	var parts []any

	if len(current.List("decorators")) > 0 {
		parts = append(parts, printClassMemberDecorators(path, options, print))
	}

	parts = append(parts, printTypeScriptAccessibilityToken(current))

	if current.Truthy("static") {
		parts = append(parts, "static ")
	}

	parts = append(parts, printAbstractToken(path))

	if current.Truthy("override") {
		parts = append(parts, "override ")
	}

	parts = append(parts, printMethod(path, options, print))

	return concatIn(path, parts...)
}

/*
- `ClassProperty`
- `PropertyDefinition`
- `ClassPrivateProperty`
- `ClassAccessorProperty`
- `AccessorProperty`
- `TSAbstractAccessorProperty` (TypeScript)
- `TSAbstractPropertyDefinition` (TypeScript)
*/
// printClassProperty is upstream's printClassProperty. Upstream's `node.variance` (Flow's +/- sigil) is
// dropped with the Flow parsers that produce it.
func printClassProperty(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)
	var parts []any

	if len(current.List("decorators")) > 0 {
		parts = append(parts, printClassMemberDecorators(path, options, print))
	}

	parts = append(parts, printDeclareToken(path), printTypeScriptAccessibilityToken(current))

	if current.Truthy("static") {
		parts = append(parts, "static ")
	}

	parts = append(parts, printAbstractToken(path))

	if current.Truthy("override") {
		parts = append(parts, "override ")
	}
	if current.Truthy("readonly") {
		parts = append(parts, "readonly ")
	}
	if current.Is("ClassAccessorProperty", "AccessorProperty", "TSAbstractAccessorProperty") {
		parts = append(parts, "accessor ")
	}
	parts = append(parts,
		printKey(path, options, print),
		printOptionalToken(path),
		printDefiniteToken(path),
		printTypeAnnotationProperty(path, print),
	)

	isAbstractProperty := current.Is("TSAbstractPropertyDefinition", "TSAbstractAccessorProperty")

	// upstream's `isAbstractProperty ? undefined : "value"`; printAssignment takes "" for absent.
	rightPropertyName := "value"
	if isAbstractProperty {
		rightPropertyName = ""
	}

	return concatIn(path, printAssignment(
		path,
		options,
		print,
		concatIn(path, parts...),
		" =",
		rightPropertyName,
	),
		printSemicolon(options),
	)
}
