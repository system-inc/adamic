package yaml

import (
	"fmt"

	"github.com/system-inc/cohere/internal/format/yaml/unist"
)

// src/language-yaml/visitor-keys.evaluate.js and get-visitor-keys.js.
//
// generateReferenceSharedVisitorKeys only makes types with equal keys share one array, which nothing
// observes, so it is not ported.

var commentsKeys = []string{
	"indicatorComment",
	"leadingComments",
	"middleComments",
	"trailingComment",
	"endComments",
}

var tagAndAnchor = []string{"anchor", "tag"}

// visitorKeys is each type's own keys, then anchor and tag (not for a tag or an anchor), then the
// comment keys.
var visitorKeys = func() map[string][]string {
	own := map[string][]string{
		"root":             {"children"},
		"document":         {"head", "body", "children"},
		"documentHead":     {"children"},
		"documentBody":     {"children"},
		"directive":        {},
		"alias":            {},
		"blockLiteral":     {},
		"blockFolded":      {"children"},
		"plain":            {"children"},
		"quoteSingle":      {},
		"quoteDouble":      {},
		"mapping":          {"children"},
		"mappingItem":      {"key", "value", "children"},
		"mappingKey":       {"content", "children"},
		"mappingValue":     {"content", "children"},
		"sequence":         {"children"},
		"sequenceItem":     {"content", "children"},
		"flowMapping":      {"children"},
		"flowMappingItem":  {"key", "value", "children"},
		"flowSequence":     {"children"},
		"flowSequenceItem": {"content", "children"},
		"comment":          {},
		"tag":              {},
		"anchor":           {},
	}
	keys := make(map[string][]string, len(own))
	for nodeType, typeKeys := range own {
		combined := append([]string{}, typeKeys...)
		if nodeType != "tag" && nodeType != "anchor" {
			combined = append(combined, tagAndAnchor...)
		}
		keys[nodeType] = append(combined, commentsKeys...)
	}
	return keys
}()

// getVisitorKeys is createGetVisitorKeys(visitorKeys): the keys for node.type, and an error for a type
// it does not know.
func getVisitorKeys(node *unist.Node) []string {
	keys, known := visitorKeys[node.NodeType]
	if !known {
		panic(fmt.Sprintf("Unexpected node type %q in YAML", node.NodeType))
	}
	return keys
}
