package css

// src/language-css/visitor-keys.evaluate.js, read through src/language-css/get-visitor-keys.js.

import "github.com/system-inc/cohere/internal/format/estree"

var cssVisitorKeys = map[string][]string{
	"css-root":                 {"frontMatter", "nodes"},
	"css-comment":              {},
	"css-rule":                 {"selector", "nodes"},
	"css-decl":                 {"value", "selector", "nodes"},
	"css-atrule":               {"selector", "params", "value", "nodes"},
	"media-query-list":         {"nodes"},
	"media-query":              {"nodes"},
	"media-type":               {},
	"media-feature-expression": {"nodes"},
	"media-feature":            {},
	"media-colon":              {},
	"media-value":              {},
	"media-keyword":            {},
	"media-url":                {},
	"media-unknown":            {},
	"selector-root":            {"nodes"},
	"selector-selector":        {"nodes"},
	"selector-comment":         {},
	"selector-string":          {},
	"selector-tag":             {},
	"selector-id":              {},
	"selector-class":           {},
	"selector-attribute":       {},
	"selector-combinator":      {"nodes"},
	"selector-universal":       {},
	"selector-pseudo":          {"nodes"},
	"selector-nesting":         {},
	"selector-unknown":         {},
	"value-value":              {"group"},
	"value-root":               {"group"},
	"value-comment":            {},
	"value-comma_group":        {"groups"},
	"value-paren_group":        {"open", "groups", "close"},
	"value-func":               {"group"},
	"value-paren":              {},
	"value-number":             {},
	"value-operator":           {},
	"value-word":               {},
	"value-colon":              {},
	"value-comma":              {},
	"value-string":             {},
	"value-atword":             {},
	"value-unicode-range":      {},
	"value-unknown":            {},
}

// visitorKeys is getVisitorKeys(node). Upstream throws "Missing visitor keys" for a type it does not
// know; this returns nil and the second result false, so the caller decides. ("front-matter" is not in
// the list: upstream reaches it through embed's getVisitorKeys, ["frontMatter"] on css-root.)
func visitorKeys(node *estree.Node) []string {
	keys, _ := lookupVisitorKeys(node)
	return keys
}

// lookupVisitorKeys is visitorKeys with whether the type is known.
func lookupVisitorKeys(node *estree.Node) ([]string, bool) {
	keys, known := cssVisitorKeys[node.Type()]
	return keys, known
}
