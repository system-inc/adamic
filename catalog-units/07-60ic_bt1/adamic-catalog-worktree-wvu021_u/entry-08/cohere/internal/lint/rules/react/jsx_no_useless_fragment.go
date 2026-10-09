package react

import (
	"encoding/json"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var (
	messageJsxNoUselessFragmentChildOfHtmlElement = rule.Message{
		Id: "ChildOfHtmlElement",
		Description: "This fragment is the child of an HTML element, which already accepts any " +
			"number of children. The fragment groups nothing that the element was not going to " +
			"accept anyway, so it only adds a layer for a reader to see through.",
	}
)

// messageJsxNoUselessFragmentNeedsMoreChildren describes the first arm's finding by what the
// fragment actually holds.
//
// The arm covers three shapes and one sentence used to describe all of them as "wraps one thing".
// Two ahra findings, `AddressesPage.tsx` and `NotificationsPage.tsx`, wrap only a commented-out
// block of dead JSX: there is no thing, and the repair is to delete the comment and return `null`
// rather than to "write the child on its own". An empty fragment has no child either.
func messageJsxNoUselessFragmentNeedsMoreChildren(ctx rule.Context, children []*ast.Node) rule.Message {
	const why = " A fragment exists to give several siblings a single parent without emitting an element"
	description := "This fragment wraps one thing, so it does nothing." + why +
		"; around a single child it is pure noise in the tree and in the source. Write the child on its own."

	kept := jsxNoUselessFragmentNonPaddingChildren(children)
	switch {
	case len(kept) == 0:
		description = "This fragment is empty, so it renders nothing." + why +
			", and here there is nothing to group. Delete it; where a value is still required, write `null`."
	case len(kept) == 1 && kept[0].Kind == ast.KindJsxExpression && kept[0].AsJsxExpression().Expression == nil:
		// `{/* ... */}` parses as an expression container with no expression. Braces holding
		// nothing at all, `{}`, are the same node with no comment, and render nothing the same way.
		if jsxNoUselessFragmentHoldsComment(ctx, kept[0]) {
			description = "This fragment holds only a comment, which renders nothing, so the fragment " +
				"renders nothing either." + why + ", and here there is nothing to group. Delete it, and the " +
				"comment with it if that is dead code; where a value is still required, write `null`."
		} else {
			description = "This fragment is empty, so it renders nothing." + why +
				", and here there is nothing to group. Delete it; where a value is still required, write `null`."
		}
	}
	return rule.Message{Id: "NeedsMoreChildren", Description: description}
}

// jsxNoUselessFragmentHoldsComment reports whether an expression container's source carries a
// comment, read from the file text because a comment is trivia and has no node.
func jsxNoUselessFragmentHoldsComment(ctx rule.Context, container *ast.Node) bool {
	text := ctx.SourceFile.Text()[container.Pos():container.End()]
	return strings.Contains(text, "/*") || strings.Contains(text, "//")
}

// JsxNoUselessFragmentOptions is the decoded option object.
//
// One key, matching `meta.schema` exactly. It defaults to false, which is Go's zero value, so no
// inversion is needed.
type JsxNoUselessFragmentOptions struct {
	// AllowExpressions keeps a fragment whose single child is an expression container.
	//
	// Upstream's reasoning is that `<Foo content={<>{value}</>} />` is a real use: the fragment
	// gives the expression a React element wrapper the prop may require. It silences ONLY the
	// `NeedsMoreChildren` arm; a fragment inside an HTML element still reports the other one.
	// Measured on the installed build.
	AllowExpressions bool `json:"allowExpressions"`
}

// DecodeJsxNoUselessFragmentOptions decodes the option object.
//
// Hand-written rather than `rule.DecodeOptionsInto` because the generic helper errors on EMPTY
// input, which is what a bare `"error"` configuration hands a rule. The one default is the zero
// value, so nothing else needs translating.
func DecodeJsxNoUselessFragmentOptions(raw []byte) (any, error) {
	var options JsxNoUselessFragmentOptions
	if len(raw) == 0 {
		return options, nil
	}
	// Lenient on purpose: eslint-plugin-react 7.37.5's schema leaves this object open (no
	// `additionalProperties`), so ESLint loads an extra key and so must this.
	if err := json.Unmarshal(raw, &options); err != nil {
		return options, err
	}
	return options, nil
}

// JsxNoUselessFragment flags a fragment that groups nothing.
//
//	valid:   <><Foo /><Bar /></>
//	valid:   <>foo</>
//	valid:   <Fragment key={item.id}>{item.name}</Fragment>
//	invalid: <><Foo /></>
//	invalid: <div><>{child}</></div>
//	invalid: <></>
//
// Ported from `react/jsx-no-useless-fragment` in `eslint-plugin-react`, the implementation that
// defined it. Its 14 clean and 16 reporting cases were extracted mechanically from the clone by
// stubbing its `RuleTester`, then all 30 were run against the installed build, 7.37.5, through the
// ESLint Linter API with the TypeScript parser, driving the FIXER as well as the judgment.
//
// # Two arms, and one input can trip both
//
// `NeedsMoreChildren` is about the fragment's own contents and `ChildOfHtmlElement` is about where
// it sits, so `<div><>{child}</></div>` reports twice, contents first.
//
// # The fixer replaces the fragment with its children's TEXT, and that is safe for TypeScript
//
// Upstream slices the source between the opener's end and the closer's start and writes it back over
// the whole node. That is the shape the port brief warns about, a fixer building a replacement span
// rather than deleting one, so the span's contents were enumerated before this was written. What
// lives inside is only the fragment's own delimiters; every child is copied byte for byte, which is
// why a type assertion, a non-null assertion, a `satisfies`, a generic call and a typed arrow all
// survive intact. Each of the five was driven through the installed fixer and then through this one.
//
// # But the fixer DOES delete attributes, and this port declines that case
//
// `isKeyedElement` skips a fragment carrying `key`, and nothing skips any other attribute. Measured
// on the installed build:
//
//	<div><Fragment id={1}>{x}</Fragment></div>          fixes to <div>{x}</div>, the id is GONE
//	<div><Fragment {...rest}>{x}</Fragment></div>       fixes to <div>{x}</div>, the spread is GONE
//	<div><Fragment key={1} id={2}>{x}</Fragment></div>  not reported at all, key exempts it
//
// A spread can carry `key`, so the second row can change behavior rather than spelling, and neither
// row tells the reader anything was removed. This port REPORTS both and withholds the repair, which
// is the shape `no-arrow-function-lifecycle` already uses here for the same reason: a rule that
// reports without fixing is useful, and a fixer that quietly deletes what somebody wrote is not.
// `TestJsxNoUselessFragmentDeclinesToDropAttributes` pins it, and the guard is mutation-tested.
//
// # Where upstream itself declines
//
// `canFix` refuses three reachable shapes and each is reproduced: a fragment with no JSX parent that
// is empty or holds real text or an expression, and any fragment inside a component element, because
// that component may require its children be a React element. Its fourth, a fragment missing its
// opening or closing token, is an old-parser artifact our parser cannot produce.
var JsxNoUselessFragment = rule.Rule{
	Name: "react/jsx-no-useless-fragment",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, ok := rule.OptionsAs[JsxNoUselessFragmentOptions](options)
		if !ok {
			settings = JsxNoUselessFragmentOptions{}
		}

		check := func(node *ast.Node) {
			if jsxNoUselessFragmentIsKeyed(node) {
				return
			}

			children := jsxNoUselessFragmentChildren(node)
			if jsxNoUselessFragmentHasLessThanTwoChildren(children) &&
				!jsxNoUselessFragmentIsOnlyTextAndNotChild(node, children) &&
				!(settings.AllowExpressions && jsxNoUselessFragmentIsSingleExpression(children)) {
				jsxNoUselessFragmentReport(ctx, node, messageJsxNoUselessFragmentNeedsMoreChildren(ctx, children))
			}

			if jsxNoUselessFragmentIsChildOfHtmlElement(node) {
				jsxNoUselessFragmentReport(ctx, node, messageJsxNoUselessFragmentChildOfHtmlElement)
			}
		}

		return rule.Listeners{
			ast.KindJsxFragment: check,

			ast.KindJsxElement: func(node *ast.Node) {
				opening := node.AsJsxElement().OpeningElement
				if opening == nil {
					return
				}
				// Not `jsx.ElementParts`, and the reason is measured rather than stylistic: that
				// helper answers a JsxOpeningElement or a JsxSelfClosingElement, and this holds a
				// JsxElement. Handed one, it returns a nil tag name, so calling it here would make
				// the rule silently stop recognizing `<Fragment>` entirely. The self-closing form
				// this file must also handle has its own listener below, which is the shape
				// `ElementParts` exists to prevent a rule from missing.
				if jsxNoUselessFragmentIsFragmentName(opening.AsJsxOpeningElement().TagName) {
					check(node)
				}
			},

			ast.KindJsxSelfClosingElement: func(node *ast.Node) {
				// `<Fragment />` is a fragment with no children at all, which upstream reaches
				// through its JSXElement listener because its parser models a self-closing element
				// the same way. Measured: it reports both arms inside a `div`.
				if jsxNoUselessFragmentIsFragmentName(node.AsJsxSelfClosingElement().TagName) {
					check(node)
				}
			},
		}
	},
}

// jsxNoUselessFragmentReport reports with a fix, or without one when the repair is not safe.
func jsxNoUselessFragmentReport(ctx rule.Context, node *ast.Node, message rule.Message) {
	fix, ok := jsxNoUselessFragmentFix(ctx, node)
	if !ok {
		ctx.ReportNode(node, message)
		return
	}
	ctx.ReportNodeWithFixes(node, message, fix)
}

// jsxNoUselessFragmentIsFragmentName matches `<Fragment>` and `<React.Fragment>`.
//
// Upstream's `jsxUtil.isFragment` compares against the fragment pragma, which is `Fragment` unless
// `settings.react.fragment` changes it, and against the react pragma for the namespaced form. cohere
// reads no React settings, so both are fixed. A different namespace does not count, which is
// the same discrimination `React.PureComponent` makes elsewhere in this package.
func jsxNoUselessFragmentIsFragmentName(tagName *ast.Node) bool {
	if tagName == nil {
		return false
	}
	switch tagName.Kind {
	case ast.KindIdentifier:
		return tagName.Text() == "Fragment"

	case ast.KindPropertyAccessExpression:
		access := tagName.AsPropertyAccessExpression()
		object := access.Expression
		if object == nil || object.Kind != ast.KindIdentifier || object.Text() != reactPragmaName {
			return false
		}
		name := access.Name()
		return name != nil && name.Kind == ast.KindIdentifier && name.Text() == "Fragment"
	}
	return false
}

// jsxNoUselessFragmentIsKeyed reports whether the element carries a `key` attribute.
//
// Upstream's `isKeyedElement` applies only to a JSXElement, never to a bare `<>` fragment, which
// cannot carry attributes anyway. A `key` means the fragment is a list item and removing it would
// break reconciliation, so the whole check is skipped rather than merely the fix.
func jsxNoUselessFragmentIsKeyed(node *ast.Node) bool {
	attributes := jsxNoUselessFragmentAttributes(node)
	if attributes == nil {
		return false
	}
	for _, attribute := range attributes.Nodes {
		if attribute.Kind != ast.KindJsxAttribute {
			continue
		}
		name := attribute.AsJsxAttribute().Name()
		if name != nil && name.Kind == ast.KindIdentifier && name.Text() == "key" {
			return true
		}
	}
	return false
}

// jsxNoUselessFragmentAttributes returns the attribute list of a fragment written as an element.
//
// A `<>` fragment has none, and returns nil.
func jsxNoUselessFragmentAttributes(node *ast.Node) *ast.NodeList {
	switch node.Kind {
	case ast.KindJsxElement:
		opening := node.AsJsxElement().OpeningElement
		if opening == nil {
			return nil
		}
		attributes := opening.AsJsxOpeningElement().Attributes
		if attributes == nil || attributes.Kind != ast.KindJsxAttributes {
			return nil
		}
		return attributes.AsJsxAttributes().Properties

	case ast.KindJsxSelfClosingElement:
		attributes := node.AsJsxSelfClosingElement().Attributes
		if attributes == nil || attributes.Kind != ast.KindJsxAttributes {
			return nil
		}
		return attributes.AsJsxAttributes().Properties
	}
	return nil
}

// jsxNoUselessFragmentChildren returns the children of either fragment spelling.
func jsxNoUselessFragmentChildren(node *ast.Node) []*ast.Node {
	var list *ast.NodeList
	switch node.Kind {
	case ast.KindJsxFragment:
		list = node.AsJsxFragment().Children
	case ast.KindJsxElement:
		list = node.AsJsxElement().Children
	case ast.KindJsxSelfClosingElement:
		return nil
	}
	if list == nil {
		return nil
	}
	return list.Nodes
}

// jsxNoUselessFragmentIsPaddingSpaces reports whether a child is whitespace React would strip.
//
// Upstream requires the text be whitespace AND contain a newline, which is exactly the text React's
// runtime removes: a line break with indentation around it is layout, while a single space between
// two elements is a rendered space and counts as a real child.
//
// `ast.Node.Text()` PANICS on a JsxText node, and the walk recovers per FILE rather than per rule,
// so one such call costs every rule in this package every finding in that file while the run still
// prints a plausible summary. The field is read directly instead, which is what `self_closing_comp`
// already does for the same reason. This was written as `Text()` first and the corpus caught it as
// a panic rather than as a wrong verdict; `TestJsxNoUselessFragmentSurvivesShapesThatWouldPanic`
// carries the multi-line cases that reach it.
func jsxNoUselessFragmentIsPaddingSpaces(child *ast.Node) bool {
	if child.Kind != ast.KindJsxText {
		return false
	}
	return text.TrimWhitespace(child.AsJsxText().Text) == "" && strings.Contains(child.AsJsxText().Text, "\n")
}

// jsxNoUselessFragmentNonPaddingChildren drops the whitespace React would strip.
func jsxNoUselessFragmentNonPaddingChildren(children []*ast.Node) []*ast.Node {
	kept := make([]*ast.Node, 0, len(children))
	for _, child := range children {
		if !jsxNoUselessFragmentIsPaddingSpaces(child) {
			kept = append(kept, child)
		}
	}
	return kept
}

// jsxNoUselessFragmentHasLessThanTwoChildren answers upstream's first arm.
//
// Fewer than two real children, EXCEPT when that one child is a call expression in braces. Upstream
// spells the exception as `return !containsCallExpression(nonPaddingChildren[0])`, on the reasoning
// that `<>{items.map(...)}</>` produces a list whose elements need the fragment's key context.
//
// Upstream's function has NO return at all when there are two or more children, so it answers
// `undefined`, which is falsy and behaves as false. Reproduced as a plain false, and the two are the
// same decision.
func jsxNoUselessFragmentHasLessThanTwoChildren(children []*ast.Node) bool {
	kept := jsxNoUselessFragmentNonPaddingChildren(children)
	if len(kept) >= 2 {
		return false
	}
	if len(kept) == 0 {
		return true
	}
	return !jsxNoUselessFragmentContainsCallExpression(kept[0])
}

// jsxNoUselessFragmentContainsCallExpression reports whether a child is `{someCall()}`.
//
// # An OPTIONAL call is not one, and the difference is in the two parsers rather than in the rule
//
// Upstream's `containsCallExpression` is `node.expression.type === 'CallExpression'`, and estree
// wraps a whole optional chain in a `ChainExpression`. So for `{items?.map(...)}` upstream sees a
// ChainExpression, answers false, and REPORTS the fragment. Our parser has no wrapper: `?.` is a
// `QuestionDotToken` hanging off the call, and the kind is `KindCallExpression` either way, so a
// bare kind test answers true and silently exempts the fragment.
//
// Measured against the installed 7.37.5 build, driving the same two sources through both:
//
//	<>{items.map(f)}</>    both silent      the exemption upstream intends
//	<>{items?.map(f)}</>   upstream REPORTS, we were silent
//
// Found on the ahra tree, where it was the whole of a two-finding disagreement:
// FinanceStatementPayeeGroupsBody.tsx:35 and FinanceStatementTransactionsBody.tsx:29 both wrap a
// single `properties.x.data?.groups.map(...)` in a fragment.
//
// This is the same parser difference `logicalAssignmentIsOptionalChain` records in
// `internal/rules/core/logical_assignment_operators.go`, reached from a different rule: any port
// that compares an estree node TYPE against our node KIND has to ask about `QuestionDotToken`
// separately, because the wrapper upstream tests for does not exist here.
func jsxNoUselessFragmentContainsCallExpression(child *ast.Node) bool {
	if child.Kind != ast.KindJsxExpression {
		return false
	}
	expression := child.AsJsxExpression().Expression
	if expression == nil || expression.Kind != ast.KindCallExpression {
		return false
	}
	// An optional call is upstream's ChainExpression, which its type test rejects. Only the
	// call's OWN `?.` matters: `a?.b.c()` is a ChainExpression upstream too, so the whole spine
	// is walked rather than just the outermost node.
	return !jsxNoUselessFragmentIsOptionalChain(expression)
}

// jsxNoUselessFragmentIsOptionalChain reports whether any link in an access spine carries `?.`.
//
// The spine is walked rather than the outermost node tested, because estree's ChainExpression wraps
// the ENTIRE chain: `a?.b.c()` and `a.b?.c()` are both a ChainExpression to upstream, so both must
// answer true here for the two rules to agree.
func jsxNoUselessFragmentIsOptionalChain(node *ast.Node) bool {
	for current := node; current != nil; {
		switch current.Kind {
		case ast.KindCallExpression:
			call := current.AsCallExpression()
			if call.QuestionDotToken != nil {
				return true
			}
			current = call.Expression
		case ast.KindPropertyAccessExpression:
			access := current.AsPropertyAccessExpression()
			if access.QuestionDotToken != nil {
				return true
			}
			current = access.Expression
		case ast.KindElementAccessExpression:
			access := current.AsElementAccessExpression()
			if access.QuestionDotToken != nil {
				return true
			}
			current = access.Expression
		case ast.KindNonNullExpression:
			current = current.AsNonNullExpression().Expression
		case ast.KindParenthesizedExpression:
			current = current.AsParenthesizedExpression().Expression
		default:
			return false
		}
	}
	return false
}

// jsxNoUselessFragmentIsOnlyTextAndNotChild answers upstream's exemption for a text-only fragment.
//
// `<>foo</>` written outside any JSX is kept, because `content={<>ee eeee</>}` is a real use: the
// prop wants an element and a bare string is not one. Upstream tests the RAW children list rather
// than the padding-filtered one, so a text-only fragment spanning lines does not qualify.
func jsxNoUselessFragmentIsOnlyTextAndNotChild(node *ast.Node, children []*ast.Node) bool {
	if len(children) != 1 || children[0].Kind != ast.KindJsxText {
		return false
	}
	parent := node.Parent
	return parent == nil || (parent.Kind != ast.KindJsxElement && parent.Kind != ast.KindJsxFragment)
}

// jsxNoUselessFragmentIsSingleExpression answers the `allowExpressions` exemption.
func jsxNoUselessFragmentIsSingleExpression(children []*ast.Node) bool {
	kept := jsxNoUselessFragmentNonPaddingChildren(children)
	return len(kept) == 1 && kept[0].Kind == ast.KindJsxExpression
}

// jsxNoUselessFragmentIsChildOfHtmlElement answers upstream's second arm.
//
// The parent must be a JSX element whose tag is a bare lowercase identifier, which is upstream's
// `/^[a-z]+$/`. That pattern accepts letters only, so `<my-tag>` and `<h1>` are NOT html elements to
// this rule, and a fragment inside either reports only the first arm. Both measured.
func jsxNoUselessFragmentIsChildOfHtmlElement(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil || parent.Kind != ast.KindJsxElement {
		return false
	}
	opening := parent.AsJsxElement().OpeningElement
	if opening == nil {
		return false
	}
	// Not `jsx.ElementParts`: the subject here is the PARENT, which the test above has already
	// narrowed to a JsxElement, and that helper answers nil for a JsxElement. The self-closing case
	// it guards against cannot arise on this line either, measured rather than argued: a
	// self-closing element has no children, so nothing nests inside one. Probed over five nesting
	// shapes and a fragment's parent came back JsxElement, JsxFragment, JsxExpression or a
	// declaration, never JsxSelfClosingElement. Upstream tests `parent.type === 'JSXElement'` the
	// same way.
	tagName := opening.AsJsxOpeningElement().TagName
	if tagName == nil || tagName.Kind != ast.KindIdentifier {
		return false
	}
	return jsxNoUselessFragmentIsAllLowercaseLetters(tagName.Text())
}

// jsxNoUselessFragmentIsAllLowercaseLetters reproduces upstream's `/^[a-z]+$/`.
//
// Letters only, so a digit, a dash, an uppercase letter or any byte outside ASCII fails. `<h1>` is
// therefore NOT an html element as far as this rule is concerned, which is upstream's behavior
// rather than an oversight here, and a Cyrillic tag is not one either.
//
// Both ends of the byte comparison are load-bearing and they are killed by different inputs, which
// is worth knowing because the obvious fixture kills only one. An uppercase letter is below `a`, so
// the lower bound catches it and a mutant dropping the UPPER bound survives an uppercase fixture. A
// tag outside ASCII is what separates them: every byte of it is above `z`.
//
// The `+` in upstream's pattern requires at least one character, and the empty-string guard that
// would express it is DELETED as unreachable. Probed over six shapes including three the parser only
// reaches through error recovery: `<>` and `<  >` both parse as fragments rather than as elements
// with an empty name, and the only callers are this file's `IsChildOfHtmlElement`, which reads a
// JsxOpeningElement's TagName. If a caller is ever added that can pass an empty string, this needs
// re-arguing, because an empty name would answer true here where upstream answers false.
func jsxNoUselessFragmentIsAllLowercaseLetters(name string) bool {
	for index := 0; index < len(name); index++ {
		if name[index] < 'a' || name[index] > 'z' {
			return false
		}
	}
	return true
}

// jsxNoUselessFragmentIsChildOfComponentElement answers upstream's `canFix` exclusion.
//
// A fragment inside `<Foo>` is not repaired, because `Foo` may require its children be a single
// React element. A fragment inside another fragment is excluded from this test, so it stays fixable.
func jsxNoUselessFragmentIsChildOfComponentElement(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil || parent.Kind != ast.KindJsxElement {
		return false
	}
	if jsxNoUselessFragmentIsChildOfHtmlElement(node) {
		return false
	}
	opening := parent.AsJsxElement().OpeningElement
	if opening == nil {
		return false
	}
	// Not `jsx.ElementParts`, for the same measured reason as `IsChildOfHtmlElement` above: the
	// subject is the parent, already narrowed to a JsxElement, and a fragment's parent is never a
	// JsxSelfClosingElement because such an element has no children to nest into.
	return !jsxNoUselessFragmentIsFragmentName(opening.AsJsxOpeningElement().TagName)
}

// jsxNoUselessFragmentCanFix answers whether the repair is safe, upstream's rules plus one of ours.
//
// Upstream refuses three reachable shapes. A fragment with no JSX parent is refused when it is empty
// (nothing to put in its place) or when any child is real text or an expression container (the
// replacement would leave a bare string or expression where an element was expected). And any
// fragment inside a component element is refused, because that component may require an element
// child.
//
// Its fourth refusal, a JSXFragment missing its opening or closing token, is an artifact of a
// TypeScript parser version that no longer ships; our parser cannot produce a JsxFragment without
// both, so there is nothing to reproduce and no fixture could reach it. That is a stated absence
// rather than a silent omission.
//
// The FIFTH refusal is this port's own and is not upstream's: a fragment written as an element
// carrying any attribute other than `key`. See the rule's doc comment for the measurement.
func jsxNoUselessFragmentCanFix(node *ast.Node) bool {
	if jsxNoUselessFragmentHasNonKeyAttributes(node) {
		return false
	}

	parent := node.Parent
	parentIsJsx := parent != nil &&
		(parent.Kind == ast.KindJsxElement || parent.Kind == ast.KindJsxFragment)
	if !parentIsJsx {
		children := jsxNoUselessFragmentChildren(node)
		if len(children) == 0 {
			return false
		}
		for _, child := range children {
			if jsxNoUselessFragmentIsNonspaceTextOrCurly(child) {
				return false
			}
		}
	}

	return !jsxNoUselessFragmentIsChildOfComponentElement(node)
}

// jsxNoUselessFragmentHasNonKeyAttributes reports whether anything but `key` is written.
//
// This is the guard behind this port's one added decline. A `key` never reaches here, because a
// keyed fragment is skipped before any report; anything else, including a spread, means the fixer
// would delete source the reader wrote.
func jsxNoUselessFragmentHasNonKeyAttributes(node *ast.Node) bool {
	attributes := jsxNoUselessFragmentAttributes(node)
	if attributes == nil {
		return false
	}
	for _, attribute := range attributes.Nodes {
		if attribute.Kind == ast.KindJsxSpreadAttribute {
			return true
		}
		if attribute.Kind != ast.KindJsxAttribute {
			continue
		}
		name := attribute.AsJsxAttribute().Name()
		if name == nil || name.Kind != ast.KindIdentifier || name.Text() != "key" {
			return true
		}
	}
	return false
}

// jsxNoUselessFragmentIsNonspaceTextOrCurly reproduces `isNonspaceJSXTextOrJSXCurly`.
func jsxNoUselessFragmentIsNonspaceTextOrCurly(child *ast.Node) bool {
	if child.Kind == ast.KindJsxExpression {
		return true
	}
	// The field rather than `Text()`, which panics on this kind; see IsPaddingSpaces above.
	return child.Kind == ast.KindJsxText && text.TrimWhitespace(child.AsJsxText().Text) != ""
}

// jsxNoUselessFragmentFix builds the replacement, or reports that none is safe.
//
// The replacement is the source text BETWEEN the opener and the closer, trimmed the way React trims
// it, written over the whole node. Copying the text verbatim is what makes this safe for TypeScript:
// a type assertion, a non-null assertion, a `satisfies`, a generic call and a typed arrow inside the
// children all survive byte for byte, each driven through both implementations.
func jsxNoUselessFragmentFix(ctx rule.Context, node *ast.Node) (rule.Fix, bool) {
	if !jsxNoUselessFragmentCanFix(node) {
		return rule.Fix{}, false
	}

	span := rule.TokenRange(ctx.SourceFile, node)
	openerEnd, closerStart, ok := jsxNoUselessFragmentInnerBounds(node)
	if !ok {
		return rule.Fix{}, false
	}

	text := ""
	if closerStart > openerEnd {
		text = jsxNoUselessFragmentTrimLikeReact(ctx.SourceFile.Text()[openerEnd:closerStart])
	}
	return rule.ReplaceRange(span, text), true
}

// jsxNoUselessFragmentInnerBounds returns the source offsets the children occupy.
//
// A self-closing element has no children and no closer, so upstream writes the empty string; that is
// expressed here as an empty range rather than as a special case at the call site.
func jsxNoUselessFragmentInnerBounds(node *ast.Node) (openerEnd int, closerStart int, ok bool) {
	switch node.Kind {
	case ast.KindJsxFragment:
		fragment := node.AsJsxFragment()
		if fragment.OpeningFragment == nil || fragment.ClosingFragment == nil {
			return 0, 0, false
		}
		return fragment.OpeningFragment.End(), fragment.ClosingFragment.Pos(), true

	case ast.KindJsxElement:
		element := node.AsJsxElement()
		if element.OpeningElement == nil || element.ClosingElement == nil {
			return 0, 0, false
		}
		return element.OpeningElement.End(), element.ClosingElement.Pos(), true

	case ast.KindJsxSelfClosingElement:
		// No children, so the replacement is empty. An equal pair produces that.
		return node.End(), node.End(), true
	}
	return 0, 0, false
}

// jsxNoUselessFragmentTrimLikeReact reproduces upstream's `trimLikeReact`.
//
// Leading whitespace is removed only when it contains a newline, and trailing whitespace likewise.
// That is React's own rule: a line break with indentation is layout and disappears, while a plain
// space between two elements is rendered and stays. Copying it exactly is what keeps the repaired
// source rendering the same thing.
func jsxNoUselessFragmentTrimLikeReact(text string) string {
	leading := 0
	for leading < len(text) && jsxNoUselessFragmentIsSpace(text[leading]) {
		leading++
	}
	trailing := len(text)
	for trailing > leading && jsxNoUselessFragmentIsSpace(text[trailing-1]) {
		trailing--
	}

	start := 0
	if strings.Contains(text[:leading], "\n") {
		start = leading
	}
	end := len(text)
	if strings.Contains(text[trailing:], "\n") {
		end = trailing
	}
	if start > end {
		return ""
	}
	return text[start:end]
}

// jsxNoUselessFragmentIsSpace matches the characters JavaScript's `\s` matches in practice here.
func jsxNoUselessFragmentIsSpace(character byte) bool {
	switch character {
	case ' ', '\t', '\n', '\r', '\v', '\f':
		return true
	}
	return false
}
