package react

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoDangerWithChildren = rule.Message{
	Id: "noDangerWithChildren",
	Description: "This element sets `dangerouslySetInnerHTML` and also passes children. React " +
		"renders one or the other, not both, so the children written here are silently " +
		"discarded and React warns at runtime. Keep whichever one is meant to render and " +
		"remove the other.",
}

// dangerPropertyName is the prop that replaces an element's content with raw markup.
//
// Written once and compared at four sites. A misspelling at any one of them would make the rule
// quietly narrower rather than fail, since every comparison is against a name that simply never
// matches.
const dangerPropertyName = "dangerouslySetInnerHTML"

// childrenPropertyName is the prop React fills from what is nested between the tags.
const childrenPropertyName = "children"

// NoDangerWithChildren flags an element that sets `dangerouslySetInnerHTML` and also has children.
//
//	valid:   <div>Children</div>
//	valid:   <div dangerouslySetInnerHTML={{ __html: "HTML" }} />
//	invalid: <div dangerouslySetInnerHTML={{ __html: "HTML" }}>Children</div>
//	invalid: React.createElement("div", { dangerouslySetInnerHTML: { __html: "HTML" } }, "Children")
//
// Ported from `react/no-danger-with-children`, read against oxc's `no_danger_with_children.rs`.
// The corpus is 17 passing and 14 failing inputs from that file's single tester block, and the
// snapshot carries 14 diagnostics, exactly one per failing input.
//
// # Children arrive three ways and all three are the corpus
//
// A `children` attribute (`<div children="Children" />`), content nested between the tags
// (`<div>Children</div>`), and the arguments after the props object in a `createElement` call.
// Upstream tests all three against both an intrinsic element and a component, which is why the
// fail list has near-duplicate `div` and `Hello` pairs: the rule deliberately does not care
// whether the tag is a host element, and dropping either half of each pair would leave that
// indifference untested.
//
// # Why this needs the type checker, when the sibling rule beside it does not
//
// Because a spread is not a dead end here. `no-children-prop` reads attributes that are written
// out, so a spread carries nothing for it to read and it declines. This rule instead follows the
// spread to the variable it names and asks what that object holds, which is `find_var_in_scope`
// upstream, built on `ctx.scoping().find_binding`. That is name resolution rather than a scope
// flag, so by the split in the porting brief it is the kind of question that costs a checker, and
// `GetSymbolAtLocation` is what answers it here. Five of the fourteen failing inputs and three of
// the seventeen passing ones turn on this, so a version without the checker would not be a
// slightly narrower port, it would get eight of thirty one corpus cases wrong.
//
// Measured rather than assumed, on the corpus itself. Every spread and every identifier in the
// props position resolves to a declaration node, and the three that must not resolve fail for the
// right reason rather than by accident:
//
//	const { a, b, ...props } = otherProps   resolves to a BindingElement, not a VariableDeclaration
//	<Hello {...undefined}>                  resolves to a symbol carrying no declarations
//	React.createElement("Hello", undefined) same
//
// Upstream declines the first by destructuring `AstKind::VariableDeclarator` and the other two by
// `find_binding` finding nothing, so all three stay clean here for the same reasons they do there
// rather than because this port happens to stop early.
//
// # The callee test is written inline, and that is deliberate
//
// `utilsreact.IsCreateElementCall` is on the shelf and is the wrong helper for this rule. Its own
// doc comment says so, naming this rule: oxc destructures a `StaticMemberExpression` here and
// returns early on anything else, so upstream never sees a bare `createElement(...)`, never
// rejects `document.createElement`, and never accepts `React["createElement"]`. Calling the shared
// helper would make this rule report where upstream is silent, and no imported fixture could catch
// it because the corpus writes every call as `React.createElement`. The inline test below is
// upstream's, and `TestNoDangerWithChildrenMatchesUpstreamsNarrowerCalleeTest` pins each of the
// three differences so a later tidy-up toward the shelf fails loudly.
//
// # Where a whitespace child counts and where it does not
//
// `<Hello dangerouslySetInnerHTML={...}> </Hello>` reports and the same element written across two
// lines does not. That reads like a bug and it is upstream's judgment, reproduced: a single space
// between the tags is real content that React would render, while a newline and the indentation
// after it is how the source was formatted and renders nothing. Upstream draws the line with
// `is_line_break`, which asks that the text both contain a newline and be entirely whitespace.
//
// Our parser has already made that exact distinction. `ast.IsWhitespaceOnlyJsxText` was measured
// against all three discriminating inputs and agrees with `is_line_break` on every one: `" "` is
// false, `"\n"` is true, and `"\n  x\n"` is false. So the predicate is asked rather than rebuilt.
//
// Only the *first* child is tested, which is upstream's `jsx.children[0]` and not a simplification
// of it. An element whose first child is a formatting newline and whose second is real text is
// therefore clean, which is a narrowness this port carries deliberately rather than improves on.
var NoDangerWithChildren = rule.Rule{
	// No namespace prefix. The config writes `react/no-danger-with-children` and the parity guard
	// strips the namespace on a `/` boundary, so a rule named `react-no-danger-with-children`
	// would match no inventory entry and lint no files while passing every fixture in this file.
	Name:             "react/no-danger-with-children",
	NeedsTypeChecker: true,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			// JsxElement rather than the opening element, because the finding spans the whole
			// element and because the children are on this node. A self-closing element parses to
			// JsxSelfClosingElement and is handled below: it has no children between tags, so it
			// can only be caught through a written or spread `children` prop.
			ast.KindJsxElement: func(node *ast.Node) {
				element := node.AsJsxElement()
				attributes := element.OpeningElement.AsJsxOpeningElement().Attributes
				if !hasElementChildren(ctx, attributes, element.Children) {
					return
				}
				if hasJsxProperty(ctx, attributes, dangerPropertyName) {
					ctx.ReportNode(node, messageNoDangerWithChildren)
				}
			},

			ast.KindJsxSelfClosingElement: func(node *ast.Node) {
				attributes := node.AsJsxSelfClosingElement().Attributes
				if !hasElementChildren(ctx, attributes, nil) {
					return
				}
				if hasJsxProperty(ctx, attributes, dangerPropertyName) {
					ctx.ReportNode(node, messageNoDangerWithChildren)
				}
			},

			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				arguments := call.Arguments
				// A call with zero or one argument cannot be a createElement call carrying props,
				// which is upstream's own first gate and its stated reason.
				if arguments == nil || len(arguments.Nodes) <= 1 {
					return
				}
				if !isStaticCreateElementCallee(call.Expression) {
					return
				}

				properties := arguments.Nodes[1]

				// Three or more arguments means the arguments after the props are the children, so
				// there are children without having to read the props object at all. Exactly two
				// means children can only be inside the props.
				hasChildren := len(arguments.Nodes) > 2 ||
					objectExpressionHasProperty(ctx, properties, childrenPropertyName)
				if !hasChildren {
					return
				}

				if objectExpressionHasProperty(ctx, properties, dangerPropertyName) {
					ctx.ReportNode(node, messageNoDangerWithChildren)
				}
			},
		}
	},
}

// hasElementChildren reports whether a JSX element passes children either way.
//
// Either as a `children` prop, written out or reached through a spread, or as content nested
// between the tags. The two are one question because upstream asks them as one: an element is
// only clean when it passes children neither way.
func hasElementChildren(ctx rule.Context, attributes *ast.Node, children *ast.NodeList) bool {
	if hasJsxProperty(ctx, attributes, childrenPropertyName) {
		return true
	}
	if children == nil || len(children.Nodes) == 0 {
		return false
	}
	// Upstream reads `children[0]` alone rather than searching, so an element whose first child is
	// a formatting newline is clean even when a later child is real content. Reproduced, and the
	// narrowness is stated in the doc comment above so the next reader does not helpfully widen it.
	return !ast.IsWhitespaceOnlyJsxText(children.Nodes[0])
}

// hasJsxProperty reports whether a JSX element passes a named prop, following spreads.
//
// This is upstream's `has_jsx_prop`. A written attribute answers by name; a spread answers by
// resolving the identifier it spreads and asking whether that object carries the name. A spread of
// anything other than a plain identifier, such as `{...{children: 1}}`, is declined here exactly as
// upstream declines it through `get_identifier_reference`, so an object literal spread inline is
// invisible to both tools rather than exempted by either.
func hasJsxProperty(ctx rule.Context, attributes *ast.Node, propertyName string) bool {
	if attributes == nil || attributes.Kind != ast.KindJsxAttributes {
		return false
	}
	properties := attributes.AsJsxAttributes().Properties
	if properties == nil {
		return false
	}

	for _, property := range properties.Nodes {
		switch property.Kind {
		case ast.KindJsxAttribute:
			// A namespaced name (`<svg xlink:children="x" />`) is a JsxNamespacedName rather
			// than an Identifier, and it is declined here to match upstream's destructuring of
			// `JSXAttributeName::Identifier`.
			//
			// The kind check cannot change a verdict and is kept anyway, which is worth stating so
			// nobody adds a fixture claiming to guard it. A namespaced name's `Text()` renders the
			// full `namespace:name`, measured rather than assumed, so it comes back
			// `"xlink:children"` and never equals a bare prop name. A mutant dropping the kind
			// check therefore survives every fixture, and correctly: JSX grammar makes the colon
			// mandatory, so no input can distinguish the two versions. It stays because it says
			// what the code means and because `Text()` on an unguarded kind is a live panic risk
			// elsewhere in this tree.
			name := property.AsJsxAttribute().Name()
			if name != nil && name.Kind == ast.KindIdentifier && name.Text() == propertyName {
				return true
			}

		case ast.KindJsxSpreadAttribute:
			expression := property.AsJsxSpreadAttribute().Expression
			if expression == nil || expression.Kind != ast.KindIdentifier {
				continue
			}
			if variableHasProperty(ctx, expression, propertyName, nil) {
				return true
			}
		}
	}

	return false
}

// objectExpressionHasProperty reports whether an expression in the props position carries a name.
//
// Two shapes answer and everything else is false, which is upstream's match arm exactly: an object
// literal written at the call site is read directly, and an identifier is resolved to the variable
// it names and read there. A call, a member expression, or `undefined` all fall through to false,
// and `React.createElement("Hello", undefined, "Children")` is a passing corpus case that lands
// here.
func objectExpressionHasProperty(ctx rule.Context, expression *ast.Node, propertyName string) bool {
	if expression == nil {
		return false
	}
	switch expression.Kind {
	case ast.KindObjectLiteralExpression:
		// Upstream's `is_object_with_prop_name`, which reads only own properties and does not
		// follow a spread. That asymmetry with the variable case below is upstream's and is
		// reproduced: `React.createElement("div", { ...other })` is not searched, while
		// `const props = { ...other }; React.createElement("div", props)` is.
		return objectLiteralHasOwnProperty(expression, propertyName)

	case ast.KindIdentifier:
		return variableHasProperty(ctx, expression, propertyName, nil)
	}
	return false
}

// variableHasProperty resolves an identifier to its variable and asks what that object holds.
//
// Upstream's `does_object_var_have_prop_name`. Only a variable declared with an object literal
// initializer answers: a parameter, a destructuring binding element, an import, and a variable
// initialized from a call all decline, because none of them names an object this rule can read
// without evaluating something.
//
// The spread inside the object recurses, so `const a = {children: 1}; const b = {...a}` answers for
// `b`. `seen` carries the declarations already visited on this path, which is upstream's cycle
// guard generalized. Upstream compares only against the immediately preceding symbol, catching the
// direct `const props = {...props}` self-reference that its corpus ships and nothing longer; a set
// costs the same and terminates on a mutual cycle too, which is a divergence in the direction of
// not hanging rather than in the direction of a different verdict.
func variableHasProperty(
	ctx rule.Context,
	identifier *ast.Node,
	propertyName string,
	seen map[*ast.Node]bool,
) bool {
	declaration := resolvedDeclaration(ctx, identifier)
	if declaration == nil || declaration.Kind != ast.KindVariableDeclaration {
		return false
	}
	if seen[declaration] {
		return false
	}

	initializer := declaration.AsVariableDeclaration().Initializer
	if initializer == nil || initializer.Kind != ast.KindObjectLiteralExpression {
		return false
	}

	if seen == nil {
		seen = map[*ast.Node]bool{}
	}
	seen[declaration] = true

	for _, property := range initializer.AsObjectLiteralExpression().Properties.Nodes {
		switch property.Kind {
		case ast.KindSpreadAssignment:
			spread := property.AsSpreadAssignment().Expression
			if spread == nil || spread.Kind != ast.KindIdentifier {
				continue
			}
			if variableHasProperty(ctx, spread, propertyName, seen) {
				return true
			}

		default:
			if name, named := staticPropertyName(property); named && name == propertyName {
				return true
			}
		}
	}

	return false
}

// objectLiteralHasOwnProperty reports whether an object literal writes a named property directly.
//
// A spread is skipped rather than followed, which is upstream's `is_object_with_prop_name` matching
// `ObjectPropertyKind::ObjectProperty` and letting `SpreadProperty` fall through.
func objectLiteralHasOwnProperty(object *ast.Node, propertyName string) bool {
	properties := object.AsObjectLiteralExpression().Properties
	if properties == nil {
		return false
	}
	for _, property := range properties.Nodes {
		// Subsumed rather than load-bearing, and measured: a SpreadAssignment's `Name()` is nil, so
		// `staticPropertyName`'s nil guard already declines it and a mutant removing this skip
		// survives every fixture because no input can distinguish the two. It is kept because it
		// states upstream's partition at the place upstream draws it, matching
		// `ObjectPropertyKind::ObjectProperty` and letting `SpreadProperty` fall through.
		if property.Kind == ast.KindSpreadAssignment {
			continue
		}
		if name, named := staticPropertyName(property); named && name == propertyName {
			return true
		}
	}
	return false
}

// staticPropertyName reads the name an object member writes, when it names one without evaluation.
//
// This is upstream's `key.static_name()` at our AST's spelling, and it is asked through
// `ast.TryGetTextOfPropertyName` rather than by reading the key's text directly. That matters for
// two reasons rather than one. It checks the kind before reading, and `Node.Text()` panics outright
// on a `ComputedPropertyName` rather than returning empty, so a version that read first and
// filtered afterwards would crash the linter on any `{[k]: v}` in the tree. And it resolves a
// computed key holding a literal, so `{["children"]: 1}` answers `children` here exactly as
// `static_name` answers it upstream, while `{[children]: 1}` declines in both.
//
// A member with no name at all, which is a spread, answers false through the nil guard rather than
// through a kind list, so a member kind nobody has thought of yet declines instead of panicking.
func staticPropertyName(property *ast.Node) (string, bool) {
	name := property.Name()
	if name == nil {
		return "", false
	}
	return ast.TryGetTextOfPropertyName(name)
}

// resolvedDeclaration reads the declaration node an identifier binds to.
//
// This replaces upstream's `find_var_in_scope`, which walks oxc's scope tree through
// `ctx.scoping().find_binding`. We have no such index and we have the checker instead, so the same
// question is asked of it. An identifier the checker cannot resolve, and a symbol carrying no
// declarations, both answer nil: `{...undefined}` is the corpus case that lands there and it is a
// passing case, so answering nil is the behavior rather than a fallback.
func resolvedDeclaration(ctx rule.Context, identifier *ast.Node) *ast.Node {
	symbol := ctx.TypeChecker.GetSymbolAtLocation(identifier)
	if symbol == nil || len(symbol.Declarations) == 0 {
		return nil
	}
	return symbol.Declarations[0]
}

// isStaticCreateElementCallee reports whether a callee is a dotted member named `createElement`.
//
// Deliberately narrower than `utilsreact.IsCreateElementCall`, and the narrowness is upstream's.
// oxc destructures `Expression::StaticMemberExpression` and then compares `callee.property.name`,
// so three things follow that the shared helper does the other way: a bare `createElement(...)` is
// not a member expression and declines, `document.createElement(...)` is not special-cased and
// reports, and `React["createElement"](...)` is an element access rather than a static member and
// declines. All three are pinned by fixtures, because none of them appears in the imported corpus.
func isStaticCreateElementCallee(callee *ast.Node) bool {
	if callee == nil || callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	name := callee.AsPropertyAccessExpression().Name()
	// A private name (`React.#createElement`) is a PrivateIdentifier rather than an Identifier and
	// carries a leading `#` in its text, so the kind is checked rather than only the text.
	return name != nil && name.Kind == ast.KindIdentifier && name.Text() == "createElement"
}
