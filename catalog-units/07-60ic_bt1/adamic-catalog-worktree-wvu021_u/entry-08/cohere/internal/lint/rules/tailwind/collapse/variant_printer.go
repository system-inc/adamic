package tailwind

import (
	"strings"

	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
)

// PrintVariant writes a parsed variant back as Tailwind spells it, which is upstream's
// `printVariant` (candidate.ts), ported from the 4.3.3 bundle together with the arbitrary-value
// printer it calls.
//
// It is not the inverse of parsing. An arbitrary selector comes back with its `&:is(…)` wrapper
// taken off, its spaces as underscores and its combinators unspaced, and a `var(--a)` value as the
// `(--a)` shorthand, so two spellings of one variant print the same. That is the property
// enforce-consistent-class-order's `strict` order needs: it groups classes by printed variant, and
// `[&_p]` and `[&:is(p)]` must land in one group.
func PrintVariant(variant ParsedVariant) string {
	switch variant.Kind {
	case ParsedVariantKindStatic:
		return variant.Root
	case ParsedVariantKindArbitrary:
		return "[" + printArbitraryValue(unwrapIsSelector(variant.Selector)) + "]"
	}

	var printed strings.Builder
	if variant.Kind == ParsedVariantKindFunctional {
		printed.WriteString(variant.Root)
		// `@` is the container-query variant, written `@lg` and `@[30rem]` with no dash.
		dash := ""
		if variant.Root != "@" {
			dash = "-"
		}
		if variant.Value != nil {
			switch variant.Value.Kind {
			case ParsedValueKindArbitrary:
				printed.WriteString(dash + bracketed(variant.Value.Value))
			case ParsedValueKindNamed:
				printed.WriteString(dash + variant.Value.Value)
			}
		}
	}
	if variant.Kind == ParsedVariantKindCompound && variant.Variant != nil {
		printed.WriteString(variant.Root + "-" + PrintVariant(*variant.Variant))
	}
	if variant.Kind == ParsedVariantKindFunctional || variant.Kind == ParsedVariantKindCompound {
		printed.WriteString(printModifier(variant.Modifier))
	}
	return printed.String()
}

// printModifier is upstream's modifier printer: `/50`, `/[0.5]`, or `/(--a)` for a lone `var()`.
func printModifier(modifier *ParsedModifier) string {
	if modifier == nil {
		return ""
	}
	switch modifier.Kind {
	case ParsedModifierKindArbitrary:
		return "/" + bracketed(modifier.Value)
	case ParsedModifierKindNamed:
		return "/" + modifier.Value
	}
	return ""
}

// bracketed prints an arbitrary value in brackets, or as the `(--a)` shorthand when it is exactly one
// `var()`, which is how upstream writes it back.
func bracketed(value string) string {
	if isLoneVar(value) {
		return "(" + printArbitraryValue(value[len("var("):len(value)-1]) + ")"
	}
	return "[" + printArbitraryValue(value) + "]"
}

// isLoneVar reports whether a value parses to exactly one `var()` call.
func isLoneVar(value string) bool {
	nodes := ParseValue(value)
	return len(nodes) == 1 && nodes[0].Kind == ValueNodeKindFunction && nodes[0].Value == "var"
}

// unwrapIsSelector takes `&:is(…)` off a selector, the wrapper the parser adds, leaving what is
// inside. Any other selector is returned as it is.
func unwrapIsSelector(selector string) string {
	nodes := ParseValue(selector)
	if len(nodes) == 3 &&
		nodes[0].Kind == ValueNodeKindWord && nodes[0].Value == "&" &&
		nodes[1].Kind == ValueNodeKindSeparator && nodes[1].Value == ":" &&
		nodes[2].Kind == ValueNodeKindFunction && nodes[2].Value == "is" {
		return ValueToCss(nodes[2].Nodes)
	}
	return selector
}

// printNode is a value node the printer can point at, since it removes nodes by identity and edits
// them in place, as upstream's does on its own tree.
type printNode struct {
	kind  ValueNodeKind
	value string
	nodes []*printNode
}

func toPrintNodes(nodes []ValueNode) []*printNode {
	printed := make([]*printNode, 0, len(nodes))
	for _, node := range nodes {
		printed = append(printed, &printNode{kind: node.Kind, value: node.Value, nodes: toPrintNodes(node.Nodes)})
	}
	return printed
}

func printNodesCss(nodes []*printNode) string {
	var css strings.Builder
	for _, node := range nodes {
		switch node.kind {
		case ValueNodeKindWord, ValueNodeKindSeparator:
			css.WriteString(node.value)
		case ValueNodeKindFunction:
			css.WriteString(node.value + "(" + printNodesCss(node.nodes) + ")")
		}
	}
	return css.String()
}

// valueOperators are the words upstream unspaces between two operands.
var valueOperators = map[string]bool{"~": true, ">": true, "+": true, "-": true, "*": true, "/": true}

// printArbitraryValue is upstream's arbitrary-value printer: spaces around a lone operator and at
// either end of a node list go, a comma separator loses its spaces, a `--name(…)` call that follows
// an operand without a comma is wrapped in parentheses, and then every space becomes `_` and every
// `_` becomes `\_`, except inside a `url()` and in a `--custom-property` word.
func printArbitraryValue(value string) string {
	nodes := toPrintNodes(ParseValue(value))
	removed := map[*printNode]bool{}

	var visit func(siblings []*printNode) []*printNode
	visit = func(siblings []*printNode) []*printNode {
		for index := 0; index < len(siblings); index++ {
			node := siblings[index]
			at := func(offset int) *printNode {
				if position := index + offset; position >= 0 && position < len(siblings) {
					return siblings[position]
				}
				return nil
			}
			switch {
			case node.kind == ValueNodeKindWord && valueOperators[node.value]:
				before, after := at(-1), at(1)
				if !isSpaceSeparator(before) || !isSpaceSeparator(after) {
					break
				}
				if farBefore := at(-2); farBefore != nil && valueOperators[farBefore.value] {
					break
				}
				if farAfter := at(2); farAfter != nil && valueOperators[farAfter.value] {
					break
				}
				removed[before], removed[after] = true, true
			case node.kind == ValueNodeKindSeparator && node.value != "" && text.TrimWhitespace(node.value) == "":
				if index == 0 || index == len(siblings)-1 {
					removed[node] = true
				}
			case node.kind == ValueNodeKindSeparator && text.TrimWhitespace(node.value) == ",":
				node.value = ","
			case node.kind == ValueNodeKindFunction && strings.HasPrefix(node.value, "--"):
				if index <= 0 {
					break
				}
				if before := at(-1); before.kind == ValueNodeKindSeparator && before.value == "," {
					break
				}
				if farBefore := at(-2); farBefore != nil && !valueOperators[farBefore.value] {
					break
				}
				// Wrapped, and the wrapper is not entered, as upstream's ReplaceSkip.
				siblings[index] = &printNode{kind: ValueNodeKindFunction, nodes: []*printNode{node}}
				continue
			}
			if len(node.nodes) > 0 {
				node.nodes = visit(node.nodes)
			}
		}
		return siblings
	}
	nodes = visit(nodes)

	if len(removed) > 0 {
		nodes = withoutRemoved(nodes, removed)
	}
	escapeUnderscoresAndSpaces(nodes)
	return printNodesCss(nodes)
}

// isSpaceSeparator reports whether a node is a separator of exactly one space.
func isSpaceSeparator(node *printNode) bool {
	return node != nil && node.kind == ValueNodeKindSeparator && node.value == " "
}

// withoutRemoved drops the marked nodes at every depth, as upstream's second walk does. A removed
// node's children go with it.
func withoutRemoved(nodes []*printNode, removed map[*printNode]bool) []*printNode {
	kept := make([]*printNode, 0, len(nodes))
	for _, node := range nodes {
		if removed[node] {
			continue
		}
		if len(node.nodes) > 0 {
			node.nodes = withoutRemoved(node.nodes, removed)
		}
		kept = append(kept, node)
	}
	return kept
}

// escapeUnderscoresAndSpaces is upstream's escape pass: a `url()`'s arguments are left as written,
// a `var()` or `theme()` call's arguments are escaped one by one, a `--custom-property` word keeps its
// underscores, and everything else has `_` written `\_` and a space written `_`.
func escapeUnderscoresAndSpaces(nodes []*printNode) {
	for _, node := range nodes {
		switch node.kind {
		case ValueNodeKindFunction:
			name := node.value
			node.value = escapeUnderscoreAndSpace(name)
			if name == "url" || strings.HasSuffix(name, "_url") {
				continue
			}
			escapeUnderscoresAndSpaces(node.nodes)
		case ValueNodeKindSeparator:
			node.value = escapeUnderscoreAndSpace(node.value)
		case ValueNodeKindWord:
			if !strings.HasPrefix(node.value, "--") {
				node.value = escapeUnderscoreAndSpace(node.value)
			}
		}
	}
}

func escapeUnderscoreAndSpace(value string) string {
	return strings.ReplaceAll(strings.ReplaceAll(value, "_", `\_`), " ", "_")
}
