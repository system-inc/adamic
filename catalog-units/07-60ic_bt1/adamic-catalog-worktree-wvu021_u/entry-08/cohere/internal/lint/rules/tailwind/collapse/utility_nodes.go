// The tree operations the `@utility` evaluator needs: cloning a definition per candidate, removing
// the declarations that dropped, substituting resolved nodes into a parsed value, and the
// registration-time normalization of `--value()` argument lists.
//
// Split out of utility.go because these are mechanics rather than semantics. The rules that decide
// a reading live there; what is here is the tree surgery those rules ask for, and keeping them
// apart is what lets utility.go read as the same sequence of branches upstream's `createCssUtility`
// is.
//
// Ported from `src/utilities.ts` and `src/ast.ts` at Tailwind 4.3.3.
package tailwind

import (
	"math"
	"regexp"
	"strconv"
	"strings"
)

// cloneNodes deep-copies a node list.
//
// Evaluation rewrites declaration values in place and removes declarations, so a definition
// consumed once would answer differently on the second class that used it. Upstream clones for the
// same reason, per candidate, in `createCssUtility`'s returned function.
//
// The Context map is copied rather than shared. Nothing in the evaluator writes it, but a shared
// map across every clone of a definition is the kind of aliasing that is correct until one caller
// is added, and a `@utility` body can carry a context node through a nested at-rule.
func cloneNodes(nodes []*Node) []*Node {
	if nodes == nil {
		return nil
	}
	clones := make([]*Node, len(nodes))
	for index, node := range nodes {
		clones[index] = cloneNode(node)
	}
	return clones
}

func cloneNode(node *Node) *Node {
	if node == nil {
		return nil
	}
	clone := *node
	if node.Context != nil {
		context := make(map[string]string, len(node.Context))
		for key, value := range node.Context {
			context[key] = value
		}
		clone.Context = context
	}
	clone.Nodes = cloneNodes(node.Nodes)
	return &clone
}

// removeNodes returns the tree with every node in the set deleted, descending into containers.
//
// A nil or empty set returns the tree unchanged and allocates nothing, which is the common path:
// most candidates resolve every declaration they were written with.
//
// Containers whose children all dropped are kept rather than pruned, matching upstream, which
// splices the declaration out of `parent.nodes` and leaves the parent standing.
//
// This is fidelity rather than a correctness requirement, and it is worth saying which: PropertySort
// counts declarations and an emptied rule holds none, so pruning would produce the same reading.
// Mutating this to prune changed no reading anywhere in the corpus, which is the honest state of it.
// It is written the upstream way because the tree is the compiled body, and a consumer that ever
// reads structure rather than only the reading would see a tree the engine does not produce.
func removeNodes(nodes []*Node, removed map[*Node]bool) []*Node {
	if len(removed) == 0 {
		return nodes
	}
	kept := make([]*Node, 0, len(nodes))
	for _, node := range nodes {
		if removed[node] {
			continue
		}
		if len(node.Nodes) > 0 {
			node.Nodes = removeNodes(node.Nodes, removed)
		}
		kept = append(kept, node)
	}
	return kept
}

// unionNodeSets returns the union of two node sets, without mutating either.
//
// Either may be nil, which is the usual case: a body where nothing dropped and a ratio resolved has
// only the splice set, and a body where something dropped and no ratio resolved has only the drop
// set.
func unionNodeSets(left, right map[*Node]bool) map[*Node]bool {
	if len(left) == 0 {
		return right
	}
	if len(right) == 0 {
		return left
	}
	union := make(map[*Node]bool, len(left)+len(right))
	for node := range left {
		union[node] = true
	}
	for node := range right {
		union[node] = true
	}
	return union
}

// replaceValueNode substitutes a resolved node list in place of a `--value()` or `--modifier()`
// call.
//
// Upstream returns `WalkAction.ReplaceSkip(resolved.nodes)`, which splices the resolved nodes into
// the parent's list at the call's position and does not descend into them. A ValueNode holds no
// parent pointer, so the splice is expressed as turning the function node into a container that
// reprints as exactly its children: an empty name with a synthetic wrapper would print parentheses
// the engine does not, so the node becomes a word whose text is the reprinted replacement instead.
//
// Reprinting rather than restructuring is safe here because ValueToCss is the exact inverse of
// ParseValue for every string these resolutions produce: theme references, numbers, and bracketed
// values, none of which contain the unmatched `)` that is the only input that does not round-trip.
func replaceValueNode(node *ValueNode, resolved []ValueNode) {
	node.Kind = ValueNodeKindWord
	node.Value = ValueToCss(resolved)
	node.Nodes = nil
}

// The three rewrites `createCssUtility` applies to a `--value()` argument before it is ever
// resolved. Compiled once at package scope rather than per definition; each runs over every
// argument of every value function in every `@utility` block.
var (
	// escapedStarPattern is the `\*` a formatter may insert, which normalizes back to `*`.
	escapedStarPattern = regexp.MustCompile(`\\\*`)
	// spacedNestedKeyPattern joins `--text --line-height` into `--text-*--line-height`.
	//
	// Ported with upstream's lazy quantifiers intact. `/--(.*?)\s--(.*?)/g` matches the shortest
	// possible run on both sides, so the second group captures nothing at all and the replacement
	// is effectively an insertion of `-*` before the second `--`. Writing the greedy version here
	// would look like a correction and would change what a three-part argument produces.
	spacedNestedKeyPattern = regexp.MustCompile(`--(.*?)\s--(.*?)`)
	// whitespacePattern removes all whitespace, turning `--value([ *])` into `--value([*])`.
	whitespacePattern = regexp.MustCompile(`\s+`)
	// repeatedStarSuffixPattern collapses `-*-*` into a single `-*`.
	repeatedStarSuffixPattern = regexp.MustCompile(`(-\*){2,}`)
)

// normalizeUtilityDefinition rewrites every `--value()` and `--modifier()` argument list in a
// definition into the spelling the resolver expects.
//
// This is `createCssUtility`'s pre-processing walk, minus the suggestion tracking. It exists
// because the argument grammar as authored is not the grammar as resolved: `--value(--spacing)`
// names the `--spacing` namespace and must become `--value(--spacing-*)` before
// resolveThemeArgument's suffix test can see it, and `--value(--text --line-height)` is the nested
// form written with a space.
//
// Skipping it would not fail loudly. Every argument in the corpus that this changes is one where
// the unnormalized spelling resolves nothing, so a port without it drops declarations that should
// survive and reports a count too low, on exactly the classes whose count this component exists to
// get right.
//
// Run once per definition at registration, matching upstream, and it mutates the definition's own
// nodes. That is safe because Compile clones before evaluating; the definition itself is only ever
// read after this.
func normalizeUtilityDefinition(definition *UtilityDefinition) {
	if definition == nil {
		return
	}
	normalizeValueFunctionArguments(definition.Nodes)
}

// normalizeValueFunctionArguments walks a body and normalizes the argument list of every
// `--value()` and `--modifier()` it finds.
func normalizeValueFunctionArguments(nodes []*Node) {
	for _, node := range nodes {
		if len(node.Nodes) > 0 {
			normalizeValueFunctionArguments(node.Nodes)
		}
		if node.Kind != KindDeclaration || !node.ValuePresent || node.Value == "" {
			continue
		}
		// Upstream's early bail. It is a substring test on the raw declaration text rather than a
		// structural one, so a declaration mentioning neither function is never parsed at all.
		if !strings.Contains(node.Value, "--value(") && !strings.Contains(node.Value, "--modifier(") {
			continue
		}

		valueAst := ParseValue(node.Value)
		normalizeValueFunctionNodes(valueAst)
		node.Value = ValueToCss(valueAst)
	}
}

// normalizeValueFunctionNodes rewrites the argument lists of `--value()` and `--modifier()` nodes
// in a parsed declaration value.
func normalizeValueFunctionNodes(nodes []ValueNode) {
	for index := range nodes {
		node := &nodes[index]
		if node.Kind != ValueNodeKindFunction {
			continue
		}
		if node.Value != "--value" && node.Value != "--modifier" {
			// Descend, because the call is normally inside a `calc()`.
			normalizeValueFunctionNodes(node.Nodes)
			continue
		}

		// The argument list is split on top-level commas of the reprinted text rather than on the
		// node list, matching upstream. The two differ: `--value(--text-*--line-height)` is several
		// nodes and one argument, and splitting nodes would treat the separator run between
		// `--text` and `--line-height` as a boundary.
		arguments := segment(ValueToCss(node.Nodes), ',')
		for argumentIndex, argument := range arguments {
			arguments[argumentIndex] = normalizeValueFunctionArgument(argument)
		}
		node.Nodes = ParseValue(strings.Join(arguments, ","))
	}
}

// normalizeValueFunctionArgument applies the four rewrites to one argument, in upstream's order.
//
// Order decides the result. Whitespace removal runs after the nested-key join, so
// `--text --line-height` becomes `--text-*--line-height` rather than `--text--line-height`; running
// it first would erase the space the join matches on and produce an argument that names no
// namespace at all.
func normalizeValueFunctionArgument(argument string) string {
	argument = escapedStarPattern.ReplaceAllString(argument, "*")
	argument = spacedNestedKeyPattern.ReplaceAllString(argument, "--$1-*--$2")
	argument = whitespacePattern.ReplaceAllString(argument, "")
	argument = repeatedStarSuffixPattern.ReplaceAllString(argument, "-*")

	// A bare `--foo` argument gains the `-*` suffix that marks it as a namespace. The two guards
	// are upstream's: an argument containing `(` is a call such as `--default(4)` and is left
	// alone, and one already carrying `-*` anywhere is already in namespace form.
	if strings.HasPrefix(argument, "--") && !strings.Contains(argument, "(") && !strings.Contains(argument, "-*") {
		argument += "-*"
	}
	return argument
}

// modFloat is JavaScript's `%` on two numbers.
//
// `math.Mod` is the same operation: IEEE 754 remainder truncated toward zero, sign following the
// dividend. Named rather than called inline so the equivalence is stated where a reader of
// isMultipleOf will look for it.
func modFloat(value, divisor float64) float64 {
	return math.Mod(value, divisor)
}

// formatJavaScriptNumber is `String(Number(value))` for the values isMultipleOf reaches.
//
// Go's shortest round-trip formatting picks the same digits ECMAScript's Number::toString does for
// every finite value that prints without an exponent, and every value that reaches here is a
// non-negative multiple of 0.25 parsed from a class name, so it is far below the 1e21 threshold
// where JavaScript switches to exponential notation. A value at or above that threshold is rejected
// explicitly rather than compared, because Go would print `1e+21` where JavaScript prints `1e21`
// and the two would disagree on a string neither accepts anyway.
func formatJavaScriptNumber(value float64) string {
	if value >= 1e21 {
		return ""
	}
	return strconv.FormatFloat(value, 'f', -1, 64)
}
