package javascript

import (
	"github.com/system-inc/cohere/internal/format/doc"
	"github.com/system-inc/cohere/internal/format/printing"
)

// print/array.js, print/tuple.js and print/array-type.js.
//
// Upstream's `node.inexact` is set only on Flow's TupleTypeAnnotation, so it is always false here and
// the `...` it prints is dropped.

/*
- `ArrayExpression`
- `ArrayPattern`
- `TSTupleType`(TypeScript)
*/

// printArray is upstream's printArray.
func printArray(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)
	var parts []any

	elementsProperty := "elements"
	if isTupleType(current) {
		elementsProperty = "elementTypes"
	}
	elements := current.List(elementsProperty)
	if len(elements) == 0 {
		parts = append(parts, groupIn(path, concatIn(path, "[", printDanglingCommentsInList(path, options, nil), "]")))
	} else {
		lastElem := elements[len(elements)-1]
		canHaveTrailingComma := !lastElem.Is("RestElement")

		// JavaScript allows you to have empty elements in an array which
		// changes its length based on the number of commas. The algorithm
		// is that if the last argument is null, we need to force insert
		// a comma to ensure JavaScript recognizes it.
		//   [,].length === 1
		//   [1,].length === 1
		//   [1,,].length === 2
		//
		// Note that getLast returns null if the array is empty, but
		// we already check for an empty array just above so we are safe
		needsForcedTrailingComma := lastElem == nil

		groupID := newGroupID("array")

		shouldBreak := (!settingsOf(options).inJestEach &&
			len(elements) > 1 &&
			everyElement(elements, func(element Node, index int) bool {
				if !isArrayExpression(element) && !isObjectExpression(element) {
					return false
				}

				elementType := element.Type()
				if index+1 < len(elements) {
					nextElement := elements[index+1]
					if nextElement != nil && elementType != nextElement.Type() {
						return false
					}
				}

				itemsKey := "properties"
				if isArrayExpression(element) {
					itemsKey = "elements"
				}

				return len(element.List(itemsKey)) > 1
			})) ||
			hasComment(current, commentDangling|commentLine, nil)

		shouldUseConciseFormatting := isConciselyPrintedArray(current, options)

		var trailingComma Doc
		switch {
		case !canHaveTrailingComma:
			trailingComma = emptyDoc
		case needsForcedTrailingComma:
			trailingComma = doc.Text(",")
		case !shouldPrintTrailingComma(options, "es5"):
			trailingComma = emptyDoc
		case shouldUseConciseFormatting:
			trailingComma = ifBreakWithGroup(",", "", groupID)
		default:
			trailingComma = ifBreak(",", "")
		}

		var elementsDoc Doc
		if shouldUseConciseFormatting {
			elementsDoc = printArrayElementsConcisely(path, options, print, trailingComma)
		} else {
			elementsDoc = concatIn(path, printArrayElements(path, options, print, elementsProperty),
				trailingComma,
			)
		}

		parts = append(parts, groupWithIn(path, concatIn(path, "[",
			indentIn(path, concatIn(path, softline,
				elementsDoc,
				printing.PrintDanglingComments(path, options, printing.DanglingOptions[Node]{}),
			)),
			softline,
			"]",
		),
			doc.GroupOptions{ShouldBreak: shouldBreak, ID: groupID},
		))
	}

	parts = append(parts,
		printOptionalToken(path),
		printTypeAnnotationProperty(path, print),
	)

	return concatIn(path, parts...)
}

// everyElement is upstream's Array.prototype.every over a node list, with the index.
func everyElement(elements []Node, predicate func(element Node, index int) bool) bool {
	for index, element := range elements {
		if !predicate(element, index) {
			return false
		}
	}
	return true
}

// isConciselyPrintedArray is upstream's isConciselyPrintedArray.
func isConciselyPrintedArray(node Node, options *Options) bool {
	return isArrayExpression(node) &&
		len(node.List("elements")) > 0 &&
		everyElement(node.List("elements"), func(element Node, _ int) bool {
			return element != nil &&
				(isNumericLiteral(element) ||
					(isSignedNumericLiteral(element) && !hasAnyComment(element.Child("argument")))) &&
				!hasComment(
					element,
					commentTrailing|commentLine,
					func(comment Node) bool {
						return !hasNewlineBackwards(options.OriginalText, locStart(comment))
					},
				)
		})
}

// isLineAfterElementEmpty is upstream's isLineAfterElementEmpty.
func isLineAfterElementEmpty(path *Path, options *Options) bool {
	current := node(path)
	text := options.OriginalText

	currentIdx := locEnd(current)
	if currentIdx == locStart(current) {
		return false
	}

	length := len(text)
	for currentIdx < length {
		if text[currentIdx] == ',' {
			break
		}

		currentIdx = printing.SkipInlineComment(text, printing.SkipTrailingComment(text, currentIdx+1))
	}

	return isNextLineEmpty(text, currentIdx)
}

// printArrayElements is upstream's printArrayElements, without the Flow-only `inexact` parameter.
func printArrayElements(path *Path, options *Options, print PrintFunc, elementsProperty string) Doc {
	var parts []Doc

	each(path, func(path *Path, _ int) {
		current := node(path)
		if current != nil {
			parts = append(parts, groupIn(path, print(nil, nil)))
		} else {
			parts = append(parts, emptyDoc)
		}

		if !path.IsLast() {
			var separator Doc = emptyDoc
			if current != nil && isLineAfterElementEmpty(path, options) {
				separator = softline
			}
			parts = append(parts, concatIn(path, ",", line, separator))
		}
	}, elementsProperty)

	return doc.Concat(parts)
}

// printArrayElementsConcisely is upstream's printArrayElementsConcisely.
func printArrayElementsConcisely(path *Path, options *Options, print PrintFunc, trailingComma Doc) Doc {
	var parts []Doc

	each(path, func(path *Path, _ int) {
		isLast := path.IsLast()
		var separator Doc = doc.Text(",")
		if isLast {
			separator = trailingComma
		}
		parts = append(parts, concatIn(path, print(nil, nil), separator))

		if !isLast {
			switch {
			case isLineAfterElementEmpty(path, options):
				parts = append(parts, concatIn(path, hardline, hardline))
			case hasComment(nextOf(path), commentLeading|commentLine, nil):
				parts = append(parts, hardline)
			default:
				parts = append(parts, line)
			}
		}
	}, "elements")

	return fill(parts)
}

/*
- `TSNamedTupleMember`(TypeScript)
*/

// printNamedTupleMember is upstream's printNamedTupleMember, print/tuple.js. `variance` is
// `TupleTypeLabeledElement` (Flow) only and is dropped.
func printNamedTupleMember(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)

	optional := ""
	if current.Bool("optional") {
		optional = "?"
	}
	return concatIn(path, print("label", nil),
		optional,
		": ",
		print("elementType", nil),
	)
}

/*
- `TSArrayType`
*/

// printArrayType is upstream's printArrayType, print/array-type.js.
func printArrayType(print PrintFunc) Doc {
	return concat(print("elementType", nil), "[]")
}
