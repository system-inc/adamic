// The CSS AST that Tailwind's own compiler produces, and the traversals that read it.
//
// Ported from `src/ast.ts` and `src/walk.ts` at Tailwind 4.3.3, read at the pinned tag rather than
// from the minified bundle, so a disagreement is a finding rather than version skew.
//
// # What this is for
//
// cohere never emits CSS. It asks one question of a compiled utility: what is its `{order, count}`
// reading, the pair `enforce-consistent-class-order` sorts on. That reading is a traversal over the
// declarations a candidate compiles to, and this file is the tree that traversal walks plus the
// traversal itself. The CSS parser (#ce8thd1) produces these nodes; theme resolution and the
// `@utility` evaluator consume them.
//
// # What was deliberately not ported, and why
//
// `ast.ts` is 1,582 lines. This is a few hundred, and the gap is not abbreviation — it is every
// part of the file that exists to turn a tree back into text, which cohere never does:
//
//   - `toCss` (~200 lines) and all printing. Nothing here renders. The reading of a candidate is
//     computed from the tree, and the CSS text is never materialized.
//   - `optimizeAst` (~430 lines). It exists so printing is a 1-to-1 transformation: hoisting
//     at-rules, deduplicating `@property`, synthesizing `color-mix` fallbacks, pruning unused
//     keyframes. All of it runs after the reading is taken and none of it changes the reading.
//   - `handleNesting` (~650 lines) and the `&`-flattening selector machinery. Nesting is resolved
//     for output; the traversal below descends into nested rules directly and does not need them
//     flattened first.
//   - `@property` emission, `propertyFallbacksRoot`/`Universal`, and the `Polyfills` bitfield.
//     See the note on at-root below: `@property` bodies are structurally invisible to the reading,
//     so the machinery that generates them has no reader here.
//   - Source maps: the `src`/`dst` `SourceLocation` fields on every node. cohere reports on the
//     class literal in the user's source, not on generated CSS, so there is no position to map back
//     to.
//   - `cloneAstNode`, `cssContext`, and `WalkAction.Replace`/`ReplaceSkip`/`ReplaceStop`. These are
//     the mutation half of the API. Every traversal here is read-only, and a replace-capable walker
//     whose replace path has no caller is untested surface that reads as supported. Walk deletes
//     during traversal are therefore not expressible, which is the intended constraint.
//
// The node kinds are not abbreviated, and that is deliberate. All six upstream kinds are here even
// though only four carry data the reading uses, because the two structural wrappers change the
// answer by existing — see PropertySort.
package tailwind

// NodeKind identifies which of the six CSS AST node shapes a Node holds.
//
// Upstream this is a discriminated union on a `kind` string field. Go has no sum type, so Node is
// one struct with a Kind tag and the fields of every variant; the alternative, an interface per
// kind, costs an allocation and a type switch at every step of a traversal that runs over every
// compiled declaration of every class literal in the repo.
type NodeKind string

const (
	// KindRule is a style rule: a selector and a body. `.flex { display: flex }`.
	KindRule NodeKind = "rule"
	// KindAtRule is an at-rule: a name, params, and a body. `@media (width >= 40rem) { ... }`.
	KindAtRule NodeKind = "at-rule"
	// KindDeclaration is a single property/value pair. `display: flex`.
	KindDeclaration NodeKind = "declaration"
	// KindComment is a comment. Carried so a parsed tree round-trips structurally; the reading
	// ignores it.
	KindComment NodeKind = "comment"
	// KindContext is a scoping wrapper carrying key/value pairs down to its subtree. It has no CSS
	// representation; it exists so the compiler can know, at an arbitrary depth, that it is inside
	// `@theme` or inside `@keyframes`.
	KindContext NodeKind = "context"
	// KindAtRoot is a wrapper marking its children for hoisting to the top level of the output.
	// Like KindContext it has no CSS representation of its own.
	KindAtRoot NodeKind = "at-root"
)

// Node is one node of the CSS AST.
//
// Which fields are meaningful depends on Kind:
//
//	KindRule         Selector, Nodes
//	KindAtRule       Name, Params, Nodes
//	KindDeclaration  Property, Value, ValuePresent, Important
//	KindComment      Value
//	KindContext      Context, Nodes
//	KindAtRoot       Nodes
//
// Reading a field outside its kind's set is a bug in the caller, not a defined zero value.
type Node struct {
	Kind NodeKind

	// Selector is the selector of a KindRule. `.flex`, `:where(& > :not(:last-child))`.
	Selector string

	// Name is the at-keyword of a KindAtRule, including its leading `@`. `@media`, `@property`.
	Name string
	// Params is everything between a KindAtRule's name and its body.
	Params string

	// Property is the property name of a KindDeclaration, including custom properties.
	// `display`, `--tw-sort`.
	Property string
	// Value is the value of a KindDeclaration, or the text of a KindComment.
	Value string
	// ValuePresent distinguishes a declaration whose value is the empty string from one whose
	// value is absent.
	//
	// This is not a Go nil-versus-empty nicety; it decides the count. Upstream the field is
	// `string | undefined` and PropertySort skips a declaration when it is `undefined` while
	// counting one whose value is `''`, with the comment that `--tw-foo:;` is valid CSS. Collapsing
	// both onto Go's empty string would undercount every such declaration, and count is the second
	// half of the sort key.
	ValuePresent bool
	// Important marks a declaration written with `!important`.
	Important bool

	// Context is the key/value scope a KindContext carries to its subtree. Upstream the values are
	// `string | boolean`; every consumer this port reaches treats them as opaque flags compared for
	// presence, so they are strings here and a boolean flag is its own key.
	Context map[string]string

	// Nodes is the body of a container kind: KindRule, KindAtRule, KindContext, KindAtRoot. It is
	// nil for KindDeclaration and KindComment.
	Nodes []*Node
}

// StyleRule builds a KindRule node.
func StyleRule(selector string, nodes ...*Node) *Node {
	return &Node{Kind: KindRule, Selector: selector, Nodes: nodes}
}

// AtRule builds a KindAtRule node. name includes its leading `@`.
func AtRule(name, params string, nodes ...*Node) *Node {
	return &Node{Kind: KindAtRule, Name: name, Params: params, Nodes: nodes}
}

// Declaration builds a KindDeclaration node with a value present.
//
// To build one whose value is absent — the `undefined` case that PropertySort skips — set
// ValuePresent to false on the result rather than passing a sentinel string, since the empty string
// is itself a meaningful value.
func Declaration(property, value string) *Node {
	return &Node{Kind: KindDeclaration, Property: property, Value: value, ValuePresent: true}
}

// Comment builds a KindComment node.
func Comment(value string) *Node {
	return &Node{Kind: KindComment, Value: value}
}

// Context builds a KindContext node.
func Context(context map[string]string, nodes ...*Node) *Node {
	return &Node{Kind: KindContext, Context: context, Nodes: nodes}
}

// AtRoot builds a KindAtRoot node.
func AtRoot(nodes ...*Node) *Node {
	return &Node{Kind: KindAtRoot, Nodes: nodes}
}

// IsContainer reports whether the node's Nodes field is meaningful for its kind.
//
// Note that this is broader than the set of kinds a reading traversal descends into. KindContext
// and KindAtRoot hold children and are not descended into by PropertySort. See its doc comment.
func (node *Node) IsContainer() bool {
	switch node.Kind {
	case KindRule, KindAtRule, KindContext, KindAtRoot:
		return true
	default:
		return false
	}
}
