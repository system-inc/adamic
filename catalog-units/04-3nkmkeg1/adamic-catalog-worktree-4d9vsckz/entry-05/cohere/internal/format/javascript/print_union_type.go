package javascript

import "github.com/system-inc/cohere/internal/format/printing"

// print/union-type.js and print/intersection-type.js.

// printUnionType is upstream's printUnionType, for `TSUnionType` and `UnionTypeAnnotation`.
func printUnionType(path *Path, options *Options, print PrintFunc, args *printArguments) Doc {
	current := node(path)

	// {
	//   a: string
	// } | null | void
	// should be inlined and not be printed in the multi-line variant
	if shouldHugUnionType(current) {
		return join(" | ", printAll(path, print, "types"))
	}

	// single-line variation
	// A | B | C

	// multi-line variation
	// | A
	// | B
	// | C

	// We want to align the children but without its comment, so it looks like
	// | child1
	// // comment
	// | child2
	var printed Doc = groupIn(path, mapPath(path, func(path *Path, _ int) Doc {
		var bar Doc
		if path.IsFirst() {
			bar = ifBreak("| ", "")
		} else {
			bar = concatIn(path, line, "| ")
		}
		typeDoc := print(nil, nil)
		if hasComment(node(path), commentLeading, nil) {
			return concatIn(path, bar, align(2, printing.PrintComments(path, typeDoc, options, nil)))
		}

		return concatIn(path, bar, printing.PrintComments(path, align(2, typeDoc), options, nil))
	}, "types"),
	)

	if shouldUnionTypePrintOwnComments(path) {
		printed = printing.PrintComments(path, printed, options, nil)
	}

	if needsParentheses(path, options) {
		return groupIn(path, concatIn(path, indentIn(path, concatIn(path, softline, printed)), softline))
	}

	if isMultipleTupleTypeElement(path) {
		return groupIn(path, concatIn(path, indentIn(path, concatIn(path, ifBreak(concatIn(path, "(", softline), ""), printed)),
			softline,
			ifBreak(")", ""),
		))
	}

	// Already indent in parent
	if (args != nil && args.assignmentLayout == "break-after-operator") ||
		!shouldIndentUnionType(path) {
		return printed
	}

	return groupIn(path, indentIn(path, concatIn(path, softline, printed)))
}

// shouldIndentUnionType is upstream's shouldIndentUnionType. The Flow branches (a FunctionTypeParam's
// typeAnnotation, and the FunctionTypeParam / FunctionTypeAnnotation / ObjectTypeProperty match) never
// match a TypeScript tree and are dropped.
func shouldIndentUnionType(path *Path) bool {
	key, parent := keyOf(path), parentOf(path)
	if (key == "typeAnnotation" && parent.Is("TSTypeAssertion")) ||
		(key == "elementTypes" && isTupleType(parent)) ||
		((key == "trueType" || key == "falseType") &&
			isConditionalType(parent)) ||
		(key == "params" && isTypeParameterInstantiation(parent)) {
		return false
	}

	return true
}

// printIntersectionType is upstream's printIntersectionType, for `TSIntersectionType` and
// `IntersectionTypeAnnotation`.
func printIntersectionType(path *Path, options *Options, print PrintFunc) Doc {
	wasIndented := false
	return groupIn(path, mapPath(path, func(path *Path, index int) Doc {
		printed := print(nil, nil)
		if path.IsFirst() {
			return printed
		}

		currentIsObjectType := isObjectType(node(path))
		previousIsObjectType := isObjectType(previousOf(path))

		// If both are objects, don't indent
		if previousIsObjectType && currentIsObjectType {
			if wasIndented {
				return concatIn(path, " & ", indentIn(path, printed))
			}
			return concatIn(path, " & ", printed)
		}

		if
		// If no object is involved, go to the next line if it breaks
		(!previousIsObjectType && !currentIsObjectType) ||
			hasLeadingOwnLineComment(options.OriginalText, node(path)) {
			// experimentalOperatorPosition is always "end", so upstream's "start" branch is dropped.
			return indentIn(path, concatIn(path, " &", line, printed))
		}

		// If you go from object to non-object or vis-versa, then inline it
		if index > 1 {
			wasIndented = true
		}

		if index > 1 {
			return concatIn(path, " & ", indentIn(path, printed))
		}
		return concatIn(path, " & ", printed)
	}, "types"),
	)
}
