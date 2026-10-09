package javascript

import (
	"strings"

	"github.com/system-inc/cohere/internal/format/doc"
	"github.com/system-inc/cohere/internal/format/estree"
)

// print/jsx.js, first half: printJsxElementInternal, printJsxChildren and the separators they
// choose between. The rest of the file is print_jsx.go.
//
// Upstream compares docs by identity here (`children[i] === hardline`, `child === whitespace`), and
// the cleanup below depends on it: a hardline printed by a child is a different array from the
// `hardline` separator printJsxChildren pushes. isSameDoc is that `===`.

// isSameDoc is JavaScript's `===` between two docs: strings compare by value, arrays and objects by
// identity. A Concat is a slice, which Go cannot compare with ==, so two arrays are the same doc when
// their parts share a backing array and length, which is what a copy of the one `hardline` value does.
func isSameDoc(left Doc, right Doc) bool {
	switch typedLeft := left.(type) {
	case nil:
		return right == nil
	case doc.Text:
		typedRight, isText := right.(doc.Text)
		return isText && typedLeft == typedRight
	}
	leftParts, leftIsArray := doc.Parts(left)
	rightParts, rightIsArray := doc.Parts(right)
	if leftIsArray || rightIsArray {
		return leftIsArray && rightIsArray &&
			len(leftParts) > 0 && len(leftParts) == len(rightParts) && &leftParts[0] == &rightParts[0]
	}
	// Every other doc is a pointer or an empty struct; interfaces holding different dynamic types
	// compare unequal without panicking.
	return left == right
}

// isEmptyText is JavaScript's `doc === ""`. Unlike isEmptyString it is false for a missing doc,
// because upstream's out-of-range `children[i + 2]` is undefined, and undefined !== "".
func isEmptyText(document Doc) bool {
	text, isText := document.(doc.Text)
	return isText && text == ""
}

// isEmptyStringOrAnyLine is upstream's isEmptyStringOrAnyLine.
func isEmptyStringOrAnyLine(document Doc) bool {
	return isEmptyText(document) || isSameDoc(document, line) || isSameDoc(document, hardline) ||
		isSameDoc(document, softline)
}

// JSX expands children from the inside-out, instead of the outside-in.
// This is both to break children before attributes,
// and to ensure that when children break, their parents do as well.
//
// Any element that is written without any newlines and fits on a single line
// is left that way.
// Not only that, any user-written-line containing multiple JSX siblings
// should also be kept on one line if possible,
// so each user-written-line is wrapped in its own group.
//
// Elements that contain newlines or don't fit on a single line (recursively)
// are fully-split, using hardline and shouldBreak: true.
//
// To support that case properly, all leading and trailing spaces
// are stripped from the list of children, and replaced with a single hardline.
func printJsxElementInternal(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)

	if current.Is("JSXElement") && isEmptyJsxElement(current) {
		return concatIn(path, print("openingElement", nil), print("closingElement", nil))
	}

	var openingLines Doc
	if current.Is("JSXElement") {
		openingLines = print("openingElement", nil)
	} else {
		openingLines = print("openingFragment", nil)
	}
	var closingLines Doc
	if current.Is("JSXElement") {
		closingLines = print("closingElement", nil)
	} else {
		closingLines = print("closingFragment", nil)
	}

	if len(current.List("children")) == 1 &&
		current.List("children")[0].Is("JSXExpressionContainer") &&
		current.List("children")[0].Child("expression").Is("TemplateLiteral", "TaggedTemplateExpression") {
		parts := []any{openingLines}
		for _, printed := range printAll(path, print, "children") {
			parts = append(parts, printed)
		}
		parts = append(parts, closingLines)
		return concatIn(path, parts...)
	}

	// Convert `{" "}` to text nodes containing a space.
	// This makes it easy to turn them into `jsxWhitespace` which
	// can then print as either a space or `{" "}` when breaking.
	//
	// Upstream's replacement node has no range; nothing reads one, since printJsxChildren never prints
	// a JSXText, so the Go gives it an empty range at 0.
	convertedChildren := make([]Node, len(current.List("children")))
	for index, child := range current.List("children") {
		if isJsxWhitespaceExpression(child) {
			convertedChildren[index] = estree.New("JSXText", 0, 0, "value", " ", "raw", " ")
		} else {
			convertedChildren[index] = child
		}
	}
	current.Set("children", convertedChildren)

	containsTag := false
	containsText := false
	expressionCount := 0
	for _, child := range current.List("children") {
		if isJsxElement(child) {
			containsTag = true
		}
		if child.Is("JSXExpressionContainer") {
			expressionCount++
		}
	}
	containsMultipleExpressions := expressionCount > 1
	containsMultipleAttributes := current.Is("JSXElement") &&
		len(current.Child("openingElement").List("attributes")) > 1

	// Record any breaks. Should never go from true to false, only false to true.
	forcedBreak := doc.WillBreak(openingLines) ||
		containsTag ||
		containsMultipleAttributes ||
		containsMultipleExpressions

	// Upstream's isMdxBlock (`path.parent.rootMarker === "mdx"`) is never true for a TypeScript file,
	// so whitespace is always the ifBreak and the mdx return below is dropped.
	rawJsxWhitespace := `{" "}`
	if settingsOf(options).SingleQuote {
		rawJsxWhitespace = "{' '}"
	}
	whitespace := ifBreak(concatIn(path, rawJsxWhitespace, softline), " ")

	isFacebookTranslationTag := current.Child("openingElement").Child("name").String("name") == "fbt"

	children := printJsxChildren(
		path,
		options,
		print,
		whitespace,
		isFacebookTranslationTag,
	)

	for _, child := range current.List("children") {
		if isMeaningfulJsxText(child) {
			containsText = true
			break
		}
	}

	// at is JavaScript's children[index]: undefined (nil) out of range.
	at := func(index int) Doc {
		if index < 0 || index >= len(children) {
			return nil
		}
		return children[index]
	}
	// splice is children.splice(start, count).
	splice := func(start int, count int) {
		end := min(start+count, len(children))
		children = append(children[:start], children[end:]...)
	}

	// We can end up we multiple whitespace elements with empty string
	// content between them.
	// We need to remove empty whitespace and softlines before JSX whitespace
	// to get the correct output.
	for i := len(children) - 2; i >= 0; i-- {
		isPairOfEmptyStrings := isEmptyText(at(i)) && isEmptyText(at(i+1))
		isPairOfHardlines := isSameDoc(at(i), hardline) &&
			isEmptyText(at(i+1)) &&
			isSameDoc(at(i+2), hardline)
		isLineFollowedByJsxWhitespace := (isSameDoc(at(i), softline) || isSameDoc(at(i), hardline)) &&
			isEmptyText(at(i+1)) &&
			isSameDoc(at(i+2), whitespace)
		isJsxWhitespaceFollowedByLine := isSameDoc(at(i), whitespace) &&
			isEmptyText(at(i+1)) &&
			(isSameDoc(at(i+2), softline) || isSameDoc(at(i+2), hardline))
		isDoubleJsxWhitespace := isSameDoc(at(i), whitespace) &&
			isEmptyText(at(i+1)) &&
			isSameDoc(at(i+2), whitespace)
		isPairOfHardOrSoftLines := (isSameDoc(at(i), softline) &&
			isEmptyText(at(i+1)) &&
			isSameDoc(at(i+2), hardline)) ||
			(isSameDoc(at(i), hardline) &&
				isEmptyText(at(i+1)) &&
				isSameDoc(at(i+2), softline))

		if (isPairOfHardlines && containsText) ||
			isPairOfEmptyStrings ||
			isLineFollowedByJsxWhitespace ||
			isDoubleJsxWhitespace ||
			isPairOfHardOrSoftLines {
			splice(i, 2)
		} else if isJsxWhitespaceFollowedByLine {
			splice(i+1, 2)
		}
	}

	// Trim trailing lines (or empty strings)
	for len(children) > 0 && isEmptyStringOrAnyLine(children[len(children)-1]) {
		children = children[:len(children)-1]
	}

	// Trim leading lines (or empty strings)
	for len(children) > 1 &&
		isEmptyStringOrAnyLine(children[0]) &&
		isEmptyStringOrAnyLine(children[1]) {
		children = children[2:]
	}

	/*
	 * Tweak how we format children if outputting this element over multiple lines.
	 * Also detect whether we will force this element to output over multiple lines.
	 *
	 * Moreover, we need to ensure that we always have line-like doc at odd index, that is rule of fill().
	 * Assuming that parts.length is always odd, satisfying the above can be straightforwardly done by:
	 * - if we push line-like doc, we push empty string after it
	 * - if we push non-line-like doc, push [parts.pop(), doc] instead
	 */
	multilineChildren := []Doc{emptyDoc}
	// pushOntoLast is multilineChildren.push([multilineChildren.pop(), doc]).
	pushOntoLast := func(document any) {
		last := len(multilineChildren) - 1
		multilineChildren[last] = concatIn(path, multilineChildren[last], document)
	}
	for i, child := range children {
		// There are a number of situations where we need to ensure we display
		// whitespace as `{" "}` when outputting this element over multiple lines.
		if isSameDoc(child, whitespace) {
			if i == 1 && doc.IsEmptyDoc(children[i-1]) {
				if len(children) == 2 {
					// Solitary whitespace
					pushOntoLast(rawJsxWhitespace)
					continue
				}
				// Leading whitespace
				multilineChildren = append(multilineChildren, concatIn(path, rawJsxWhitespace, hardline), emptyDoc)
				continue
			}

			if i == len(children)-1 {
				// Trailing whitespace
				pushOntoLast(rawJsxWhitespace)
				continue
			}

			if isEmptyText(at(i-1)) && isSameDoc(at(i-2), hardline) {
				// Whitespace after line break
				pushOntoLast(rawJsxWhitespace)
				continue
			}
		}

		// Note that children always satisfy the rule of fill() content.
		// - printJsxChildren always returns valid fill() content
		// - we always remove even number (containing zero) of leading items from children.
		if i%2 == 0 {
			// non-line-like
			pushOntoLast(child)
		} else {
			// line-like
			multilineChildren = append(multilineChildren, child, emptyDoc)
		}

		if doc.WillBreak(child) {
			forcedBreak = true
		}
	}

	// If there is text we use `fill` to fit as much onto each line as possible.
	// When there is no text (just tags and expressions) we use `group`
	// to output each on a separate line.
	var content Doc
	if containsText {
		content = fill(multilineChildren)
	} else {
		content = groupWithIn(path, multilineChildren, doc.GroupOptions{ShouldBreak: true})
	}

	// Upstream wraps content in `cursor` when options.cursorNode, nodeBeforeCursor or nodeAfterCursor
	// is a JSXText child. cohere never tracks a cursor (see printing.PrintAstToDoc), so that is dropped.

	multiLineElem := groupIn(path, concatIn(path, openingLines,
		indentIn(path, concatIn(path, hardline, content)),
		hardline,
		closingLines,
	))

	if forcedBreak {
		return multiLineElem
	}

	flatParts := []any{openingLines}
	for _, child := range children {
		flatParts = append(flatParts, child)
	}
	flatParts = append(flatParts, closingLines)
	return conditionalGroup([]Doc{
		groupIn(path, concatIn(path, flatParts...)),
		multiLineElem,
	}, doc.GroupOptions{})
}

// JSX Children are strange, mostly for two reasons:
//  1. JSX reads newlines into string values, instead of skipping them like JS
//  2. up to one whitespace between elements within a line is significant,
//     but not between lines.
//
// Leading, trailing, and lone whitespace all need to
// turn themselves into the rather ugly `{' '}` when breaking.
//
// This function returns Doc array that satisfies rule of `fill()`.
//
// The text is split in bytes: every JSX whitespace character is ASCII, so the words are the same
// strings upstream's UTF-16 split produces.
func printJsxChildren(
	path *Path,
	options *Options,
	print PrintFunc,
	whitespace Doc,
	isFacebookTranslationTag bool,
) []Doc {
	var prevPart Doc = emptyDoc
	parts := []Doc{prevPart}
	// To ensure rule of `fill()`, we use `push()` and `pushLine()` instead of `parts.push()`.
	push := func(document Doc) {
		prevPart = document
		last := len(parts) - 1
		parts[last] = concatIn(path, parts[last], document)
	}
	pushLine := func(document Doc) {
		if isEmptyText(document) {
			return
		}
		prevPart = document
		parts = append(parts, document, emptyDoc)
	}
	// prevPartText is prevPart as the string upstream passes to the separators as `child`. Where they
	// are called with prevPart it is always the last word pushed, a string.
	prevPartText := func() string {
		text, _ := prevPart.(doc.Text)
		return string(text)
	}
	each(path, func(path *Path, _ int) {
		current, next := node(path), nextOf(path)
		if current.Is("JSXText") {
			text := getRaw(current)

			// Contains a non-whitespace character
			if isMeaningfulJsxText(current) {
				words := jsxWhitespace.split(text, true /* captureWhitespace */)

				// Starts with whitespace
				if words[0] == "" {
					words = words[1:]
					if strings.Contains(words[0], "\n") {
						pushLine(
							separatorWithWhitespace(
								isFacebookTranslationTag,
								words[1],
								current,
								next,
							),
						)
					} else {
						pushLine(whitespace)
					}
					words = words[1:]
				}

				// Upstream's `let endWhitespace;` stays undefined when there is no trailing whitespace,
				// and also when words.pop() runs on an already empty array.
				var endWhitespace string
				hasEndWhitespace := false
				// Ends with whitespace
				if len(words) > 0 && words[len(words)-1] == "" {
					words = words[:len(words)-1]
					if len(words) > 0 {
						endWhitespace = words[len(words)-1]
						hasEndWhitespace = true
						words = words[:len(words)-1]
					}
				}

				// This was whitespace only without a new line.
				if len(words) == 0 {
					return
				}

				for i, word := range words {
					if i%2 == 1 {
						pushLine(line)
					} else {
						push(doc.Text(word))
					}
				}

				if hasEndWhitespace {
					if strings.Contains(endWhitespace, "\n") {
						pushLine(
							separatorWithWhitespace(
								isFacebookTranslationTag,
								prevPartText(),
								current,
								next,
							),
						)
					} else {
						pushLine(whitespace)
					}
				} else {
					pushLine(
						separatorNoWhitespace(
							isFacebookTranslationTag,
							prevPartText(),
							current,
							next,
						),
					)
				}
			} else if strings.Contains(text, "\n") {
				// Keep (up to one) blank line between tags/expressions/text.
				// Note: We don't keep blank lines between text elements.
				if strings.Count(text, "\n") > 1 {
					pushLine(hardline)
				}
			} else {
				pushLine(whitespace)
			}
		} else {
			printedChild := print(nil, nil)
			push(printedChild)

			directlyFollowedByMeaningfulText := next != nil && isMeaningfulJsxText(next)
			if directlyFollowedByMeaningfulText {
				trimmed := jsxWhitespace.trim(getRaw(next))
				firstWord := jsxWhitespace.split(trimmed, false)[0]
				pushLine(
					separatorNoWhitespace(
						isFacebookTranslationTag,
						firstWord,
						current,
						next,
					),
				)
			} else {
				pushLine(hardline)
			}
		}
	}, "children")

	return parts
}

// separatorNoWhitespace is upstream's separatorNoWhitespace. `child.length` is UTF-16 code units.
func separatorNoWhitespace(
	isFacebookTranslationTag bool,
	child string,
	childNode Node,
	nextNode Node,
) Doc {
	if isFacebookTranslationTag {
		return emptyDoc
	}

	if (childNode.Is("JSXElement") && childNode.Child("closingElement") == nil) ||
		(nextNode.Is("JSXElement") && nextNode.Child("closingElement") == nil) {
		if utf16Length(child) == 1 {
			return softline
		}
		return hardline
	}

	return softline
}

// separatorWithWhitespace is upstream's separatorWithWhitespace. `child.length` is UTF-16 code units.
func separatorWithWhitespace(
	isFacebookTranslationTag bool,
	child string,
	childNode Node,
	nextNode Node,
) Doc {
	if isFacebookTranslationTag {
		return hardline
	}

	if utf16Length(child) == 1 {
		if (childNode.Is("JSXElement") && childNode.Child("closingElement") == nil) ||
			(nextNode.Is("JSXElement") && nextNode.Child("closingElement") == nil) {
			return hardline
		}
		return softline
	}

	return hardline
}
