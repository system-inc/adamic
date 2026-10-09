package doc

import (
	"fmt"
	"strings"

	"github.com/system-inc/cohere/internal/format/arena"
)

/*
 * The document utilities printers call while building docs, src/document/utilities/index.js and
 * traverse-doc.js in the fork. Statement for statement, including the parts that read as accidents:
 *
 *   - findInDoc stops descending once it has an answer but still pops what is already stacked, so the
 *     callback is never called again after the first result, which is all that is observable.
 *   - mapDoc caches by identity, so a subtree shared by two parents (a conditional group's states share
 *     their parts) maps to one shared result. Go docs that are slices (Concat) have no identity and are
 *     mapped afresh each time, which is safe because every node that can carry identity inside them
 *     (a group, an indent) is a pointer and is cached. A *Sequence is mapped afresh too, exactly as the
 *     Concat of its parts would be.
 *   - cleanDoc concatenates adjacent strings and flattens arrays one level, after its children are
 *     already clean, the way upstream's mapDoc order makes it.
 */

// TraverseDoc is upstream's traverseDoc. onEnter returning false skips the doc's children; onExit may
// be nil.
func TraverseDoc(root Doc, onEnter func(Doc) bool, onExit func(Doc), shouldTraverseConditionalGroups bool) {
	stack := []Doc{root}
	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if _, isExit := current.(traverseExit); isExit {
			leaving := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			onExit(leaving)
			continue
		}
		if onExit != nil {
			stack = append(stack, current, traverseExit{})
		}
		if onEnter != nil && !onEnter(current) {
			continue
		}

		switch document := current.(type) {
		case *Fill:
			for index := len(document.Parts) - 1; index >= 0; index-- {
				stack = append(stack, document.Parts[index])
			}
		case *IfBreak:
			stack = append(stack, document.FlatContents, document.BreakContents)
		case *Group:
			if shouldTraverseConditionalGroups && document.ExpandedStates != nil {
				for index := len(document.ExpandedStates) - 1; index >= 0; index-- {
					stack = append(stack, document.ExpandedStates[index])
				}
			} else {
				stack = append(stack, document.Contents)
			}
		case *Align:
			stack = append(stack, document.Contents)
		case *Indent:
			stack = append(stack, document.Contents)
		case *IndentIfBreak:
			stack = append(stack, document.Contents)
		case *Label:
			stack = append(stack, document.Contents)
		case *LineSuffix:
			stack = append(stack, document.Contents)
		case Text, trimDoc, lineSuffixBoundaryDoc, *Line, breakParentDoc, nil:
		default:
			parts, isArray := Parts(current)
			if !isArray {
				panic(fmt.Sprintf("invalid doc %T", current))
			}
			for index := len(parts) - 1; index >= 0; index-- {
				stack = append(stack, parts[index])
			}
		}
	}
}

// FindInDoc is upstream's findInDoc: the first non-nil answer fn gives, in traversal order.
func FindInDoc[T any](root Doc, fn func(Doc) (T, bool), defaultValue T) T {
	result := defaultValue
	found := false
	TraverseDoc(root, func(document Doc) bool {
		if found {
			return false
		}
		if value, ok := fn(document); ok {
			found = true
			result = value
		}
		return true
	}, nil, false)
	return result
}

// WillBreak is upstream's willBreak: a broken group, a hard line, or a break-parent anywhere inside.
func WillBreak(document Doc) bool {
	return FindInDoc(document, func(current Doc) (bool, bool) {
		switch typed := current.(type) {
		case *Group:
			if typed.Break {
				return true, true
			}
		case *Line:
			if typed.Hard {
				return true, true
			}
		case breakParentDoc:
			return true, true
		}
		return false, false
	}, false)
}

// CanBreak is upstream's canBreak: any line at all inside.
func CanBreak(document Doc) bool {
	return FindInDoc(document, func(current Doc) (bool, bool) {
		if _, isLine := current.(*Line); isLine {
			return true, true
		}
		return false, false
	}, false)
}

// MapDoc is upstream's mapDoc: children first, then fn on the rebuilt node.
func MapDoc(root Doc, fn func(Doc) Doc) Doc {
	if text, isText := root.(Text); isText {
		return fn(text)
	}
	mapped := map[Doc]Doc{}
	var rec func(Doc) Doc
	rec = func(document Doc) Doc {
		cacheable := isPointerDoc(document)
		if cacheable {
			if result, present := mapped[document]; present {
				return result
			}
		}
		result := mapOne(document, rec, fn)
		if cacheable {
			mapped[document] = result
		}
		return result
	}
	return rec(root)
}

func isPointerDoc(document Doc) bool {
	switch document.(type) {
	case *Group, *Fill, *IfBreak, *Indent, *Align, *IndentIfBreak, *LineSuffix, *Label, *Line:
		return true
	}
	return false
}

// mapOne maps a node's children and hands fn the node rebuilt over them. Upstream always rebuilds; here a
// node none of whose children changed is handed over as it is, which holds the same docs, so mapping a
// tree fn leaves alone copies nothing (#r89mksm). No caller's fn writes to the node it is given, and the
// one identity upstream's mapped tree is read for, a conditional group's states (sameStates), is still
// rebuilt every time. The printer writes break marks onto groups in place, but a mark follows from the
// group's contents alone, so a group the mapped doc shares with its source is marked as a copy would be.
func mapOne(document Doc, rec func(Doc) Doc, fn func(Doc) Doc) Doc {
	if original, isArray := Parts(document); isArray {
		parts := mapParts(original, rec)
		if sameParts(parts, original) {
			// The doc itself: putting a slice back in an interface allocates.
			return fn(document)
		}
		return fn(Concat(parts))
	}
	switch typed := document.(type) {
	case *Fill:
		parts := mapParts(typed.Parts, rec)
		if sameParts(parts, typed.Parts) {
			return fn(typed)
		}
		return fn(&Fill{Parts: parts})
	case *IfBreak:
		breakContents, flatContents := rec(typed.BreakContents), rec(typed.FlatContents)
		if sameDoc(breakContents, typed.BreakContents) && sameDoc(flatContents, typed.FlatContents) {
			return fn(typed)
		}
		return fn(&IfBreak{BreakContents: breakContents, FlatContents: flatContents, GroupID: typed.GroupID})
	case *Group:
		if typed.ExpandedStates == nil {
			contents := rec(typed.Contents)
			if sameDoc(contents, typed.Contents) {
				return fn(typed)
			}
			copied := *typed
			copied.Contents = contents
			return fn(&copied)
		}
		copied := *typed
		states := make([]Doc, len(typed.ExpandedStates))
		for index, state := range typed.ExpandedStates {
			states[index] = rec(state)
		}
		copied.ExpandedStates = states
		copied.Contents = states[0]
		return fn(&copied)
	case *Align:
		contents := rec(typed.Contents)
		if sameDoc(contents, typed.Contents) {
			return fn(typed)
		}
		copied := *typed
		copied.Contents = contents
		return fn(&copied)
	case *Indent:
		contents := rec(typed.Contents)
		if sameDoc(contents, typed.Contents) {
			return fn(typed)
		}
		return fn(&Indent{Contents: contents})
	case *IndentIfBreak:
		contents := rec(typed.Contents)
		if sameDoc(contents, typed.Contents) {
			return fn(typed)
		}
		copied := *typed
		copied.Contents = contents
		return fn(&copied)
	case *Label:
		contents := rec(typed.Contents)
		if sameDoc(contents, typed.Contents) {
			return fn(typed)
		}
		return fn(&Label{Label: typed.Label, Contents: contents})
	case *LineSuffix:
		contents := rec(typed.Contents)
		if sameDoc(contents, typed.Contents) {
			return fn(typed)
		}
		return fn(&LineSuffix{Contents: contents})
	case Text, trimDoc, lineSuffixBoundaryDoc, *Line, breakParentDoc, nil:
		return fn(document)
	}
	panic(fmt.Sprintf("invalid doc %T", document))
}

// mapParts maps each part, returning parts itself when every part maps to itself and a new slice
// otherwise.
func mapParts(parts []Doc, rec func(Doc) Doc) []Doc {
	var mapped []Doc
	for index, part := range parts {
		result := rec(part)
		if mapped == nil && !sameDoc(result, part) {
			mapped = make([]Doc, len(parts))
			copy(mapped, parts[:index])
		}
		if mapped != nil {
			mapped[index] = result
		}
	}
	if mapped == nil {
		return parts
	}
	return mapped
}

// sameParts is whether mapParts handed parts back: the same array at the same length.
func sameParts(left []Doc, right []Doc) bool {
	return len(left) == len(right) && (len(left) == 0 || &left[0] == &right[0])
}

// sameDoc is whether two docs are the same doc: the same node, the same text, or for an array, which Go
// cannot compare as a Concat, the same parts.
func sameDoc(left Doc, right Doc) bool {
	leftParts, leftIsArray := Parts(left)
	rightParts, rightIsArray := Parts(right)
	if leftIsArray || rightIsArray {
		return leftIsArray && rightIsArray && (leftParts == nil) == (rightParts == nil) && sameParts(leftParts, rightParts)
	}
	return left == right
}

// isFalsy is JavaScript truthiness for a doc in the places upstream writes `!doc.contents`.
func isFalsy(document Doc) bool {
	if document == nil {
		return true
	}
	text, isText := document.(Text)
	return isText && text == ""
}

// CleanDoc is upstream's cleanDoc, mapDoc(doc, cleanDocFn), walked by docCleaner rather than MapDoc: the
// same nodes, cleaned in the same order, with fewer copies (#r89mksm).
func CleanDoc(document Doc) Doc {
	if text, isText := document.(Text); isText {
		return cleanDocFn(text)
	}
	return cleanDocInto(document, nil)
}

// cleanDocInto is CleanDoc with the concats it rebuilds cut from slab, which may be nil to allocate each.
func cleanDocInto(document Doc, slab *arena.Slab[Doc]) Doc {
	if text, isText := document.(Text); isText {
		return cleanDocFn(text)
	}
	cleaner := &docCleaner{slab: slab}
	cleaner.rec = cleaner.clean
	return cleaner.clean(document)
}

// docCleaner is CleanDoc's walk. It is MapDoc with cleanDocFn, with one difference: a concat's cleaned
// children are gathered in buffer, one stretch per concat on the walk's path, and the concat is built
// once, at its cleaned length, or handed back as it is when cleaning leaves it unchanged. MapDoc builds
// the mapped concat, then cleanDocFn builds the cleaned one by appending.
type docCleaner struct {
	// mapped is MapDoc's cache by identity, made at the first node that can carry one.
	mapped map[Doc]Doc
	buffer []Doc
	rec    func(Doc) Doc
	// slab is where the concats cleaning rebuilds come from, nil to allocate each (#v6ksqg3).
	slab *arena.Slab[Doc]
}

func (cleaner *docCleaner) clean(document Doc) Doc {
	cacheable := isPointerDoc(document)
	if cacheable {
		if result, present := cleaner.mapped[document]; present {
			return result
		}
	}
	var result Doc
	if parts, isArray := Parts(document); isArray {
		result = cleaner.cleanConcat(document, parts)
	} else {
		result = mapOne(document, cleaner.rec, cleanDocFn)
	}
	if cacheable {
		if cleaner.mapped == nil {
			cleaner.mapped = map[Doc]Doc{}
		}
		cleaner.mapped[document] = result
	}
	return result
}

// cleanConcat is mapOne and cleanDocFn for a concat. Each child's own walk leaves buffer as it found it,
// so the concat's children are buffer[start:] once they are all cleaned.
func (cleaner *docCleaner) cleanConcat(document Doc, parts []Doc) Doc {
	start := len(cleaner.buffer)
	changed := false
	for _, part := range parts {
		cleaned := cleaner.clean(part)
		if !sameDoc(cleaned, part) {
			changed = true
		}
		cleaner.buffer = append(cleaner.buffer, cleaned)
	}
	children := cleaner.buffer[start:]
	var result Doc
	if !changed && isCleanConcat(parts) {
		result = document
	} else {
		result = cleanParts(children, cleaner.slab)
	}
	clear(children)
	cleaner.buffer = cleaner.buffer[:start]
	return result
}

func cleanDocFn(document Doc) Doc {
	if parts, isArray := Parts(document); isArray {
		if isCleanConcat(parts) {
			// The doc itself: putting a slice back in an interface allocates.
			return document
		}
		return cleanParts(parts, nil)
	}
	switch typed := document.(type) {
	case *Fill:
		allEmpty := true
		for _, part := range typed.Parts {
			if text, isText := part.(Text); !isText || text != "" {
				allEmpty = false
				break
			}
		}
		if allEmpty {
			return Text("")
		}
		if len(typed.Parts) == 1 {
			return typed.Parts[0]
		}
	case *Group:
		if isFalsy(typed.Contents) && typed.ID == nil && !typed.Break && typed.ExpandedStates == nil {
			return Text("")
		}
		if inner, isGroup := typed.Contents.(*Group); isGroup && inner.ID == typed.ID && inner.Break == typed.Break &&
			sameStates(inner.ExpandedStates, typed.ExpandedStates) {
			return inner
		}
	case *Align:
		if isFalsy(typed.Contents) {
			return Text("")
		}
	case *Indent:
		if isFalsy(typed.Contents) {
			return Text("")
		}
	case *IndentIfBreak:
		if isFalsy(typed.Contents) {
			return Text("")
		}
	case *LineSuffix:
		if isFalsy(typed.Contents) {
			return Text("")
		}
	case *IfBreak:
		if isFalsy(typed.FlatContents) && isFalsy(typed.BreakContents) {
			return Text("")
		}
	}
	return document
}

// cleanParts is cleanDocFn for a concat's parts: falsy parts dropped, a nested concat flattened one level,
// and a text joined onto a text before it. The result is built at its final length.
//
// The result comes from slab, which may be nil to allocate it.
func cleanParts(parts []Doc, slab *arena.Slab[Doc]) Doc {
	size := 0
	for _, part := range parts {
		if nested, isArray := Parts(part); isArray && len(nested) > 0 {
			size += len(nested)
		} else {
			size++
		}
	}
	cleaned := Concat(slab.Make(size))
	for _, part := range parts {
		if isFalsy(part) {
			continue
		}
		current, rest := part, []Doc(nil)
		if nested, isArray := Parts(part); isArray {
			if len(nested) == 0 {
				// `[first, ...rest] = []` makes first undefined, which upstream then pushes.
				cleaned = append(cleaned, nil)
				continue
			}
			current, rest = nested[0], nested[1:]
		}
		currentText, currentIsText := current.(Text)
		if len(cleaned) > 0 {
			if previous, previousIsText := cleaned[len(cleaned)-1].(Text); currentIsText && previousIsText {
				cleaned[len(cleaned)-1] = previous + currentText
				cleaned = append(cleaned, rest...)
				continue
			}
		}
		cleaned = append(cleaned, current)
		cleaned = append(cleaned, rest...)
	}
	if len(cleaned) == 0 {
		return Text("")
	}
	if len(cleaned) == 1 {
		return cleaned[0]
	}
	return cleaned
}

// isCleanConcat is whether cleanDocFn would rebuild parts equal to themselves: two or more, none falsy,
// none a concat to flatten, and no two texts side by side to join. Then the concat is returned as it is
// rather than copied (#r89mksm).
func isCleanConcat(parts []Doc) bool {
	if len(parts) < 2 {
		return false
	}
	previousIsText := false
	for _, part := range parts {
		if isFalsy(part) {
			return false
		}
		if _, isArray := Parts(part); isArray {
			return false
		}
		_, isText := part.(Text)
		if isText && previousIsText {
			return false
		}
		previousIsText = isText
	}
	return true
}

// sameStates is upstream's `a.expandedStates === b.expandedStates`: identity of the array, which in
// Go is two nil slices, or two slices sharing a backing array and length.
func sameStates(left []Doc, right []Doc) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return len(left) == len(right) && &left[0] == &right[0]
}

// ReplaceEndOfLine is upstream's replaceEndOfLine: every newline in every string becomes replacement,
// literalline by default.
func ReplaceEndOfLine(document Doc, replacement Doc) Doc {
	if replacement == nil {
		replacement = Literalline
	}
	return MapDoc(document, func(current Doc) Doc {
		text, isText := current.(Text)
		if !isText {
			return current
		}
		lines := strings.Split(string(text), "\n")
		parts := make([]Doc, len(lines))
		for index, line := range lines {
			parts[index] = Text(line)
		}
		return Join(replacement, parts)
	})
}

// RemoveLines is upstream's removeLines: soft lines vanish, lines become spaces, if-breaks go flat.
func RemoveLines(document Doc) Doc {
	return MapDoc(document, func(current Doc) Doc {
		switch typed := current.(type) {
		case *Line:
			if !typed.Hard {
				if typed.Soft {
					return Text("")
				}
				return Text(" ")
			}
		case *IfBreak:
			return typed.FlatContents
		}
		return current
	})
}

// IsEmptyDoc is upstream's isEmptyDoc: whether cleanDoc(doc) would be "".
func IsEmptyDoc(document Doc) bool {
	isEmpty := true
	TraverseDoc(document, func(current Doc) bool {
		switch typed := current.(type) {
		case Text:
			if typed == "" {
				return true
			}
			isEmpty = false
			return false
		case trimDoc, lineSuffixBoundaryDoc, *Line, breakParentDoc:
			isEmpty = false
			return false
		}
		return true
	}, nil, false)
	return isEmpty
}

// StripTrailingHardline is upstream's stripTrailingHardline.
func StripTrailingHardline(document Doc) Doc {
	return stripTrailingHardlineFromDoc(CleanDoc(document))
}

// StripTrailingHardlineWith is StripTrailingHardline with the concats cleaning rebuilds cut from slab, which
// may be nil to allocate each (#v6ksqg3). The caller owns the slab's lifetime: the stripped doc reads it
// until the doc is laid out.
func StripTrailingHardlineWith(document Doc, slab *arena.Slab[Doc]) Doc {
	return stripTrailingHardlineFromDoc(cleanDocInto(document, slab))
}

func stripTrailingHardlineFromParts(parts []Doc) []Doc {
	parts = append([]Doc(nil), parts...)
	for len(parts) >= 2 {
		_, isLine := parts[len(parts)-2].(*Line)
		_, isBreakParent := parts[len(parts)-1].(breakParentDoc)
		if !isLine || !isBreakParent {
			break
		}
		parts = parts[:len(parts)-2]
	}
	if len(parts) > 0 {
		parts[len(parts)-1] = stripTrailingHardlineFromDoc(parts[len(parts)-1])
	}
	return parts
}

func stripTrailingHardlineFromDoc(document Doc) Doc {
	if parts, isArray := Parts(document); isArray {
		return Concat(stripTrailingHardlineFromParts(parts))
	}
	switch typed := document.(type) {
	case *Indent:
		return &Indent{Contents: stripTrailingHardlineFromDoc(typed.Contents)}
	case *IndentIfBreak:
		copied := *typed
		copied.Contents = stripTrailingHardlineFromDoc(typed.Contents)
		return &copied
	case *Group:
		copied := *typed
		copied.Contents = stripTrailingHardlineFromDoc(typed.Contents)
		return &copied
	case *LineSuffix:
		return &LineSuffix{Contents: stripTrailingHardlineFromDoc(typed.Contents)}
	case *Label:
		return &Label{Label: typed.Label, Contents: stripTrailingHardlineFromDoc(typed.Contents)}
	case *IfBreak:
		return &IfBreak{
			BreakContents: stripTrailingHardlineFromDoc(typed.BreakContents),
			FlatContents:  stripTrailingHardlineFromDoc(typed.FlatContents),
			GroupID:       typed.GroupID,
		}
	case *Fill:
		return &Fill{Parts: stripTrailingHardlineFromParts(typed.Parts)}
	case Text:
		// trim-newlines' trimNewlinesEnd: trailing \r and \n only.
		return Text(strings.TrimRight(string(typed), "\r\n"))
	}
	return document
}
