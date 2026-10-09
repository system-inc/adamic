package css

// src/language-css/parse/parse-value.js.

import (
	"regexp"
	"strings"

	"github.com/system-inc/cohere/internal/format/css/values"
	"github.com/system-inc/cohere/internal/format/estree"
)

func isClosingParenthesis(node *estree.Node) bool {
	return node.Is("paren") && node.String("value") == ")"
}

// newParenGroup is upstream's { open, close, groups, type: "paren_group" } literal, keys in that order.
func newParenGroup(open *estree.Node) *estree.Node {
	return estree.New("paren_group", 0, 0, "open", open, "close", nil, "groups", []*estree.Node{})
}

// newCommaGroup is upstream's { groups: [], type: "comma_group" } literal.
func newCommaGroup() *estree.Node {
	return estree.New("comma_group", 0, 0, "groups", []*estree.Node{})
}

// pushGroup is group.groups.push(child).
func pushGroup(group *estree.Node, child *estree.Node) {
	group.Set("groups", append(group.List("groups"), child))
}

func parseValueNode(valueNode *estree.Node, walk valueRootWalk, options *parseOptions) *estree.Node {
	nodes := valueNode.List("nodes")
	parenGroup := newParenGroup(nil)
	parenGroupStack := []*estree.Node{parenGroup}
	rootParenGroup := parenGroup
	commaGroup := newCommaGroup()
	commaGroupStack := []*estree.Node{commaGroup}

	for i := 0; i < len(nodes); i++ {
		node := nodes[i]

		if options.parser == "scss" &&
			node.Is("number") &&
			node.String("unit") == ".." &&
			strings.HasSuffix(node.String("value"), ".") {
			// Work around postcss bug parsing `50...` as `50.` with unit `..`
			// Set the unit to `...` to "accidentally" have arbitrary arguments work in the same way that cases where the node already had a unit work.
			// For example, 50px... is parsed as `50` with unit `px...` already by postcss-values-parser.
			value := node.String("value")
			node.Set("value", value[:len(value)-1])
			node.Set("unit", "...")
		}

		if node.Is("func") && node.String("value") == "selector" {
			group := node.Child("group")
			selector := sliceJavaScript(
				getValueRoot(walk).String("text"),
				intProperty(group.Child("open"), "sourceIndex")+1,
				intProperty(group.Child("close"), "sourceIndex"),
			)
			parsedSelector := parseSelector(selector)
			parsedSelector.Set("sourceIndex", intProperty(group.Child("open"), "sourceIndex")+1)
			raws := rawsIn(parsedSelector)
			if raws == nil {
				raws = map[string]any{}
				parsedSelector.Set("raws", raws)
			}
			raws["selector"] = selector
			group.Set("groups", []*estree.Node{parsedSelector})
		}

		if node.Is("func") && node.String("value") == "url" {
			groups := node.Child("group").List("groups")

			// Create a view with any top-level comma groups flattened.
			var groupList []*estree.Node
			for _, group := range groups {
				if group.Is("comma_group") {
					groupList = append(groupList, group.List("groups")...)
				} else {
					groupList = append(groupList, group)
				}
			}

			var firstGroup *estree.Node
			if len(groupList) > 0 {
				firstGroup = groupList[0]
			}

			// Stringify if the value parser can't handle the content.
			if hasSCSSInterpolation(groupList) ||
				(!hasStringOrFunction(groupList) &&
					!isSCSSVariable(firstGroup, options)) {
				// The one place a groups list holds a string, so it is a []any here and not a
				// []*estree.Node.
				node.Child("group").Set("groups", []any{getFunctionArgumentsText(node, walk)})
			}
		}
		if node.Is("paren") && node.String("value") == "(" {
			parenGroup = newParenGroup(node)
			parenGroupStack = append(parenGroupStack, parenGroup)

			commaGroup = newCommaGroup()
			commaGroupStack = append(commaGroupStack, commaGroup)
		} else if isClosingParenthesis(node) {
			if len(commaGroup.List("groups")) > 0 {
				pushGroup(parenGroup, commaGroup)
			}
			parenGroup.Set("close", node)

			/* c8 ignore next 3 */
			if len(commaGroupStack) == 1 {
				panic(javaScriptError{message: "Unbalanced parenthesis"})
			}

			commaGroupStack = commaGroupStack[:len(commaGroupStack)-1]
			commaGroup = commaGroupStack[len(commaGroupStack)-1]
			pushGroup(commaGroup, parenGroup)

			parenGroupStack = parenGroupStack[:len(parenGroupStack)-1]
			parenGroup = parenGroupStack[len(parenGroupStack)-1]
		} else if node.Is("comma") {
			// Trialing comma
			if i == len(nodes)-3 &&
				nodes[i+1].Is("comment") &&
				isClosingParenthesis(nodes[i+2]) {
				continue
			}

			pushGroup(parenGroup, commaGroup)
			commaGroup = newCommaGroup()
			commaGroupStack[len(commaGroupStack)-1] = commaGroup
		} else {
			pushGroup(commaGroup, node)
		}
	}
	if len(commaGroup.List("groups")) > 0 {
		pushGroup(parenGroup, commaGroup)
	}

	return rootParenGroup
}

func flattenGroups(node *estree.Node) *estree.Node {
	groups := node.List("groups")

	if node.Is("paren_group") &&
		node.Child("open") == nil &&
		node.Child("close") == nil &&
		len(groups) == 1 {
		return flattenGroups(groups[0])
	}

	if node.Is("comma_group") && len(groups) == 1 {
		return flattenGroups(groups[0])
	}

	if node.Is("paren_group", "comma_group") {
		// { ...node, groups: node.groups.map(flattenGroups) }: a copy, keys in the same order.
		flattened := make([]*estree.Node, len(groups))
		for index, group := range groups {
			flattened[index] = flattenGroups(group)
		}
		copied := estree.New(node.Type(), node.Range[0], node.Range[1])
		for _, key := range node.Keys() {
			if key == "groups" {
				copied.Set(key, flattened)
			} else {
				copied.Set(key, node.Get(key))
			}
		}
		return copied
	}

	return node
}

// parseNestedValue walks the tree children first and turns every nodes list into a group.
//
// upstream iterates the node's keys as they stand; a key it adds on the way (group) is not visited,
// so the Go walk iterates the keys it started with. walk carries the root getValueRoot climbs to.
func parseNestedValue(node *estree.Node, walk valueRootWalk, options *parseOptions) *estree.Node {
	if node == nil {
		return node
	}
	for _, key := range node.Keys() {
		switch child := node.Get(key).(type) {
		case *estree.Node:
			parseNestedValue(child, walk, options)
		case []*estree.Node:
			for _, element := range child {
				parseNestedValue(element, walk, options)
			}
		}
		if key == "nodes" {
			if !(node.Is("atword") && len(node.List("nodes")) == 0) {
				node.Set("group", flattenGroups(parseValueNode(node, walk, options)))
			}
			node.Delete(key)
		}
	}
	return node
}

var selectorTypePattern = regexp.MustCompile(`^selector-`)

func parseValue(value string, options *parseOptions) *estree.Node {
	// The Less `~` + "`" inline JavaScript case (options.parser === "less") is not ported.

	result, err := values.Parse(value, values.Options{Loose: true})
	if err != nil {
		// Upstream's catch takes every throw, so every error lands here, the port's own included.
		return estree.New("value-unknown", 0, 0, "value", value)
	}

	result.Set("text", value)

	parsedResult := parseNestedValue(result, valueRootWalk{root: result}, options)

	addTypePrefix(parsedResult, "value-", selectorTypePattern)
	return parsedResult
}
