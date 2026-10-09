package css

// src/language-css/parse/utilities.js: addTypePrefix and addMissingType.

import (
	"regexp"
	"strings"

	"github.com/system-inc/cohere/internal/format/estree"
)

// addTypePrefix prefixes every type in the tree that does not already start with prefix (and does
// not match skipPrefix), children first.
//
// upstream deletes node.parent on the way; the Go trees have no parent links. A node whose type is ""
// is one whose JavaScript type was undefined, which upstream leaves alone (typeof is not "string").
func addTypePrefix(value any, prefix string, skipPrefix *regexp.Regexp) any {
	switch node := value.(type) {
	case *estree.Node:
		if node == nil {
			return value
		}
		for _, key := range node.Keys() {
			addTypePrefix(node.Get(key), prefix, skipPrefix)
		}
		if node.Type() != "" &&
			!strings.HasPrefix(node.Type(), prefix) &&
			(skipPrefix == nil || !skipPrefix.MatchString(node.Type())) {
			node.SetType(prefix + node.Type())
		}
	case []*estree.Node:
		for _, child := range node {
			addTypePrefix(child, prefix, skipPrefix)
		}
	case map[string]any:
		// A plain object (raws, spaces, source) is walked too, and a "type" string in it would be
		// prefixed the same way; none of the libraries put one there.
		for key, child := range node {
			addTypePrefix(child, prefix, skipPrefix)
			if text, isString := child.(string); key == "type" && isString &&
				!strings.HasPrefix(text, prefix) &&
				(skipPrefix == nil || !skipPrefix.MatchString(text)) {
				node[key] = prefix + text
			}
		}
	case []any:
		for _, child := range node {
			addTypePrefix(child, prefix, skipPrefix)
		}
	}
	return value
}

// addMissingType gives a node that has a truthy value and no type the type "unknown", children first.
func addMissingType(value any) any {
	switch node := value.(type) {
	case *estree.Node:
		if node == nil {
			return value
		}
		for _, key := range node.Keys() {
			addMissingType(node.Get(key))
		}
		if node.Truthy("value") && node.Type() == "" {
			node.SetType("unknown")
		}
	case []*estree.Node:
		for _, child := range node {
			addMissingType(child)
		}
	}
	return value
}

// The helpers below are the glue's reads of the JavaScript objects the libraries hand back: property
// access with typeof checks, which upstream writes inline.

// sourceOf is node.source when it is an object.
func sourceOf(node *estree.Node) map[string]any {
	source, _ := node.Get("source").(map[string]any)
	return source
}

// ensureSource is `node.source ??= {}`.
func ensureSource(node *estree.Node) map[string]any {
	source := sourceOf(node)
	if source == nil {
		source = map[string]any{}
		node.Set("source", source)
	}
	return source
}

// rawsIn is node.raws when it is an object, or nil (which reads as an empty object).
func rawsIn(node *estree.Node) map[string]any {
	raws, _ := node.Get("raws").(map[string]any)
	return raws
}

// mapIn is object[key] when it is an object.
func mapIn(object map[string]any, key string) map[string]any {
	value, _ := object[key].(map[string]any)
	return value
}

// numberIn is object[key] when `typeof object[key] === "number"`. The libraries' offsets, lines and
// columns are ints; a float64 is accepted the same way.
func numberIn(object map[string]any, key string) (int, bool) {
	return toNumber(object[key])
}

// numberProperty is node[key] when it is a number.
func numberProperty(node *estree.Node, key string) (int, bool) {
	return toNumber(node.Get(key))
}

// intProperty is node[key] as a number, 0 when it is not one.
func intProperty(node *estree.Node, key string) int {
	number, _ := numberProperty(node, key)
	return number
}

func toNumber(value any) (int, bool) {
	switch number := value.(type) {
	case int:
		return number, true
	case float64:
		// NaN is a number to typeof. postcss-values-parser writes one as a paren's end column, which
		// loc.js never reads (a paren's end comes from its sourceIndex); 0 keeps Go's conversion defined.
		if number != number {
			return 0, true
		}
		return int(number), true
	}
	return 0, false
}

// hasStringType is `typeof node.type === "string"`: "" is a type upstream left undefined.
func hasStringType(node *estree.Node) bool {
	return node != nil && node.Type() != ""
}

// objectChildren are the objects upstream's walks visit under one property: the node, or the nodes in
// the array, skipping holes and strings (a url() argument the value parser cannot handle is a string
// among the groups). Plain objects (raws, spaces) are left out: upstream walks into them too, but they
// hold no type, no source and no sourceIndex, so nothing happens there.
func objectChildren(value any) []*estree.Node {
	switch typed := value.(type) {
	case *estree.Node:
		if typed != nil {
			return []*estree.Node{typed}
		}
	case []*estree.Node:
		children := make([]*estree.Node, 0, len(typed))
		for _, child := range typed {
			if child != nil {
				children = append(children, child)
			}
		}
		return children
	case []any:
		children := make([]*estree.Node, 0, len(typed))
		for _, child := range typed {
			if node, isNode := child.(*estree.Node); isNode && node != nil {
				children = append(children, node)
			}
		}
		return children
	}
	return nil
}

// isEmptyArray is `Array.isArray(value) && value.length === 0`.
func isEmptyArray(value any) bool {
	switch typed := value.(type) {
	case []*estree.Node:
		return len(typed) == 0
	case []any:
		return len(typed) == 0
	}
	return false
}

// firstTruthyString is `a || b || ""` over values that are strings when truthy.
func firstTruthyString(values ...any) string {
	for _, value := range values {
		if text, isString := value.(string); isString && text != "" {
			return text
		}
	}
	return ""
}

// sliceJavaScript is text.slice(start, end): negative indexes count from the end, and both clamp.
func sliceJavaScript(text string, start int, end int) string {
	clamp := func(index int) int {
		if index < 0 {
			index += len(text)
		}
		return max(0, min(index, len(text)))
	}
	start, end = clamp(start), clamp(end)
	if start >= end {
		return ""
	}
	return text[start:end]
}
