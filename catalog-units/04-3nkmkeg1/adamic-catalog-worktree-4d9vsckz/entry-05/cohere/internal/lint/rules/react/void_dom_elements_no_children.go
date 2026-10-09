package react

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/jsx"
	utilsreact "github.com/system-inc/cohere/internal/lint/ecmascript/react"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// voidDomElements is the set of HTML elements that are written self-closing and hold no content.
//
// Sixteen names, upstream's `VOID_DOM_ELEMENTS` verbatim and in the same order. This is the rule's
// entire subject, so a typo in any one entry would make the rule silently narrower on exactly that
// element while every other fixture stayed green. Upstream's corpus exercises only `br` and `img`,
// so fourteen of the sixteen have no imported coverage and are pinned by a test written here.
//
// A map rather than a slice because the lookup is the rule's inner loop and a map states that the
// order carries no meaning. Built once at package initialisation rather than per file, which is the
// shape a rule that rebuilt a map per file was measured costing most of a lint run.
var voidDomElements = map[string]bool{
	"area": true, "base": true, "br": true, "col": true,
	"embed": true, "hr": true, "img": true, "input": true,
	"keygen": true, "link": true, "menuitem": true, "meta": true,
	"param": true, "source": true, "track": true, "wbr": true,
}

// childrenPropertyName and dangerPropertyName are already declared by `no_danger_with_children.go`
// in this package, and they carry the same two names this rule compares against. They are reused
// rather than redeclared: a second pair of constants spelling the same strings is exactly the drift
// the shelf exists to prevent, and a build error would catch a redeclaration anyway.

var messageVoidDomElementsNoChildren = rule.Message{
	Id: "voidDomElementsNoChildren",
	Description: "This element is a void HTML element, so the browser gives it no closing tag and " +
		"no content model. React throws at runtime rather than rendering anything when children " +
		"are passed to one, so the children written here never appear and the element they were " +
		"meant for never renders. Remove the children, or use an element that can contain them.",
}

// VoidDomElementsNoChildren flags a void HTML element that is passed children.
//
//	valid:   <div>Children</div>
//	valid:   <br />
//	valid:   React.createElement("div", {}, "Children")
//	invalid: <br>Children</br>
//	invalid: <br children="Children" />
//	invalid: React.createElement("br", {}, "Children")
//
// Ported from `react/void-dom-elements-no-children`, read against oxc's
// `void_dom_elements_no_children.rs`. The corpus is 12 passing and 10 failing inputs from that
// file's single tester block, and the snapshot carries 10 diagnostics, exactly one per failing
// input, so there was no per-input count to recover. The rule takes no options: upstream declares
// no configuration struct and no case in the corpus carries a second tuple element.
//
// # Children arrive four ways and the rule counts rather than reads
//
// Content between the tags, a `children` attribute, a `dangerouslySetInnerHTML` attribute, and the
// third and later arguments of a `createElement` call. The danger prop belongs in that list because
// it sets the element's content just as children do, which is why upstream compares against both
// names at both sites rather than only against `children`.
//
// The JSX half asks whether the children list is non-empty and never looks at what a child holds.
// That is narrower reading than it sounds and it cuts the other way from the sibling rule beside
// it: `<br>` written across two lines reports, because the formatting newline between the tags is a
// JsxText node, and `<br>{}</br>` reports because an empty expression container is a node too.
// `no-danger-with-children` filters exactly that shape out through `ast.IsWhitespaceOnlyJsxText`,
// and reaching for the same predicate here because the neighbouring rule has it would make this
// rule silent on inputs upstream reports. Measured on the release oxlint binary rather than
// reasoned about: `<br>\n</br>` and `<br>{}</br>` each report once, while `<br></br>` is silent.
//
// # The element list is compared exactly, and that is a judgment rather than an oversight
//
// `<BR>`, `<Br>` and `<Img>` are all silent, measured on the release binary. Upstream compares
// against a lowercase list with no folding, and it is right rather than merely faithful: JSX
// resolves a capitalized tag to a component binding in scope, so `<Img>` renders whatever that name
// is bound to, which is not the void HTML element and may legitimately take children. A port that
// folded case would report on component references.
//
// # A spread is seen and deliberately declined
//
// The natural reading of a rule that misses spreads is that spreads are invisible, and here that
// reading is wrong. A JSX spread is a `JsxSpreadAttribute` in the same properties list and a
// spread property is a `SpreadAssignment` in the same object literal, so both are perfectly visible
// to a listener that looks for them. Upstream looks and answers false, at both sites, matching
// `JSXAttributeItem::SpreadAttribute(_) => false` and `ObjectPropertyKind::SpreadProperty(_) =>
// false`. The decline is reproduced because reading a spread means deciding what an arbitrary
// expression evaluates to, and this rule declares no type checker precisely because it never has to
// ask. The sibling `no-danger-with-children` does resolve spreads and does declare one, which is
// the difference between the two rules rather than an inconsistency in either.
//
// So `<br {...{children: 1}} />` is silent here and on the release binary, and it is a real input
// rather than a hypothetical. Upstream's corpus does carry the case proving a spread does not stop
// the search either: `<img {...props} children='Foo' />` reports, because the written attribute
// beside the spread still answers.
//
// # Only a static identifier key answers in the props object
//
// Upstream matches `PropertyKey::StaticIdentifier` alone, so `{'children': 'Foo'}` and
// `{['children']: 'Foo'}` are both silent even though both name the prop unambiguously. That is
// narrower than the rest of this package, which reads keys through `ast.TryGetTextOfPropertyName`
// and therefore resolves a quoted key and a computed key holding a literal. The narrowness is
// upstream's and is reproduced here deliberately, because reaching for the house helper would widen
// the rule past upstream on two inputs no imported fixture covers. Established by running
// `oxlint --config` with `{"plugins":["react"],"rules":{"react/void-dom-elements-no-children":"error"}}`
// over both spellings and reading zero findings, with a known-failing control in the same run.
//
// # The callee test is the shared helper here, unlike the rule beside it
//
// This rule calls oxc's shared `is_create_element_call`, so a bare `createElement(...)`, a computed
// `React["createElement"](...)` and any object other than React all count, while
// `document.createElement` is excluded by name. `no-danger-with-children` writes a narrower test
// inline because upstream writes a narrower test inline *there*, so the two rules genuinely look at
// different sets of calls and `utilsreact.IsCreateElementCall` is the right helper for this one.
//
// One difference had to be handled rather than inherited. Our helper runs `ast.SkipParentheses`
// over the callee and upstream does not: `is_create_element_call` matches the callee directly
// against three expression variants, and a parenthesized expression is none of them, so it reaches
// the `_ => false` arm. Measured rather than assumed, because the brief makes measuring parens
// mandatory and guessing here would have shipped a divergence in either direction:
// `(React.createElement)('br', {}, 'Foo')` and `(createElement)('br', {}, 'Foo')` are both silent on
// the release binary while both unparenthesized forms report. The callee's kind is therefore
// checked here before the helper is consulted, which reproduces upstream's fall-through without
// restating which three spellings are accepted. The helper is left alone: other rules in this
// package depend on its current behavior, and this is a difference between callers rather than a
// defect in it.
var VoidDomElementsNoChildren = rule.Rule{
	// No namespace prefix. The config writes `react/void-dom-elements-no-children` and the parity
	// guard strips the namespace on a `/` boundary, so a rule named
	// `react-void-dom-elements-no-children` would match no inventory entry and lint no files while
	// passing every fixture in this file.
	Name: "react/void-dom-elements-no-children",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// reportElement decides about one JSX element, given its own children list.
		//
		// Shared by both element kinds because upstream anchors on `JSXElement` and reads the
		// opening element out of it, which makes one body serve both shapes in its tree. Ours
		// splits them, so the children list is passed in: a self-closing element has none.
		reportElement := func(node *ast.Node, children *ast.NodeList) {
			tagName, attributes := jsx.ElementParts(node)
			// A member-expression tag (`<Foo.br />`) or a namespaced one is not an identifier, and
			// upstream returns on both by destructuring `JSXElementName::Identifier`. The kind
			// check is load-bearing against a crash as well as a wrong verdict: a JSX member tag
			// parses to a PropertyAccessExpression here and `Node.Text()` panics on that kind.
			if tagName == nil || tagName.Kind != ast.KindIdentifier {
				return
			}
			elementName := tagName.Text()
			if !voidDomElements[elementName] {
				return
			}

			// Upstream's `!children.is_empty() || has_children_attribute_or_danger`, one
			// disjunction feeding one report, which is why an element carrying children two ways
			// reports once rather than twice.
			hasChildren := children != nil && len(children.Nodes) > 0
			if !hasChildren && !hasContentAttribute(attributes) {
				return
			}

			ctx.ReportNode(tagName, voidDomElementMessage(elementName))
		}

		return rule.Listeners{
			// Both element kinds, and neither is redundant. A self-closing element produces no
			// JsxOpeningElement in our tree, and half of upstream's failing corpus is
			// self-closing, so dropping the second arm would silence it while leaving the first
			// arm's fixtures green.
			ast.KindJsxElement: func(node *ast.Node) {
				element := node.AsJsxElement()
				reportElement(element.OpeningElement, element.Children)
			},

			ast.KindJsxSelfClosingElement: func(node *ast.Node) {
				reportElement(node, nil)
			},

			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()

				// Upstream's paren behavior, reproduced by checking the kind before consulting the
				// shared helper, which would otherwise skip parentheses the helper's other callers
				// want skipped. See the doc comment above for the measurement.
				callee := call.Expression
				if callee == nil {
					return
				}
				switch callee.Kind {
				case ast.KindIdentifier, ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
				default:
					return
				}
				if !utilsreact.IsCreateElementCall(node) {
					return
				}

				arguments := call.Arguments
				if arguments == nil || len(arguments.Nodes) == 0 {
					return
				}

				// The element name must be a string literal written at the call site. A component
				// reference, a template literal and a variable all decline, because upstream
				// matches `Argument::StringLiteral` and nothing else. This also answers the
				// parenthesized-argument case without special handling.
				elementNameNode := arguments.Nodes[0]
				if elementNameNode.Kind != ast.KindStringLiteral {
					return
				}
				elementName := elementNameNode.Text()
				if !voidDomElements[elementName] {
					return
				}

				// Fewer than two arguments cannot carry children either way, which is upstream's
				// own gate and the reason `React.createElement('img')` is a passing corpus case.
				if len(arguments.Nodes) < 2 {
					return
				}

				// The props argument must be an object literal written at the call site. A
				// variable, `null` and a string all decline here, before the third argument is
				// ever consulted, which is why `React.createElement('br', null, 'Foo')` is silent
				// upstream despite passing children positionally. Reproduced rather than improved
				// on, and measured on the release binary.
				properties := arguments.Nodes[1]
				if properties.Kind != ast.KindObjectLiteralExpression {
					return
				}

				// Three or more arguments means children by position. Exactly two means children
				// only if the props object names one of them.
				hasChildren := len(arguments.Nodes) > 2 || objectLiteralHasContentProperty(properties)
				if !hasChildren {
					return
				}

				ctx.ReportNode(elementNameNode, voidDomElementMessage(elementName))
			},
		}
	},
}

// voidDomElementMessage names the element in the finding.
//
// Upstream interpolates the tag through Rust's `{:?}` debug formatting, so its rendered text reads
// `Void DOM element <"br" /> cannot receive children.` with the element name in quotes. That is a
// formatting slip rather than a judgment, visible in the snapshot and in the release binary alike,
// and it is not reproduced: the quotes say nothing to a reader and the same message with a bare
// name is the one upstream's own documentation and its help line describe. This is a difference the
// differential harness can see, so it is stated here rather than left for the next reader to find.
func voidDomElementMessage(elementName string) rule.Message {
	return rule.Message{
		Id: messageVoidDomElementsNoChildren.Id,
		Description: "Void DOM element <" + elementName +
			" /> cannot receive children.",
	}
}

// hasContentAttribute reports whether a JSX element writes a prop that sets its content.
//
// Either `children` or `dangerouslySetInnerHTML`, which upstream tests together in one `any` over
// the attribute list. `jsx.AttributeName` supplies the guard: it declines a spread, which carries
// no name, and a namespaced name, which upstream also declines by destructuring
// `JSXAttributeName::Identifier`. Its body was read rather than its name trusted, and it makes
// exactly those two declines and no others.
func hasContentAttribute(attributes *ast.Node) bool {
	if attributes == nil || attributes.Kind != ast.KindJsxAttributes {
		return false
	}
	properties := attributes.AsJsxAttributes().Properties
	if properties == nil {
		return false
	}

	for _, property := range properties.Nodes {
		name, named := jsx.AttributeName(property)
		if named && isContentPropertyName(name) {
			return true
		}
	}
	return false
}

// objectLiteralHasContentProperty reports whether a props object names children or danger.
//
// Only a static identifier key answers, which is upstream's `PropertyKey::StaticIdentifier` arm.
// The key's text is read through the node's own `Name()` and guarded on `KindIdentifier` rather
// than through `ast.TryGetTextOfPropertyName`, and that is the deliberate narrowing described in
// the doc comment above: the house helper resolves a quoted key and a computed key holding a
// literal, both of which upstream is measurably silent on. A spread carries a nil name and declines
// through the same guard rather than through a kind list, so a member shape nobody has met yet
// declines instead of panicking.
func objectLiteralHasContentProperty(object *ast.Node) bool {
	properties := object.AsObjectLiteralExpression().Properties
	if properties == nil {
		return false
	}

	for _, property := range properties.Nodes {
		name := property.Name()
		if name == nil || name.Kind != ast.KindIdentifier {
			continue
		}
		if isContentPropertyName(name.Text()) {
			return true
		}
	}
	return false
}

// isContentPropertyName reports whether a prop name is one of the two that set an element's content.
//
// One function rather than four comparisons spread across two loops, so the pair of names is
// written once and the two sites cannot drift apart on one of them.
func isContentPropertyName(name string) bool {
	return name == childrenPropertyName || name == dangerPropertyName
}
