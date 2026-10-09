package react

import (
	"encoding/json"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/jsx"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// stylePropObjectMessage is the single finding this rule produces.
//
// One message id, no interpolation, so the text is a constant and there is nothing to render.
var stylePropObjectMessage = rule.Message{
	Id:          "stylePropNotObject",
	Description: "Style prop value must be an object",
}

// StylePropObjectOptions is the decoded option object.
type StylePropObjectOptions struct {
	// Allow names components whose style prop is not examined.
	//
	// Upstream matches the name exactly against a set. An empty or absent list allows nothing,
	// which is the unconfigured state and means every element is examined.
	Allow []string
}

// stylePropObjectWireOptions is the JSON shape.
type stylePropObjectWireOptions struct {
	Allow []string `json:"allow"`
}

// DecodeStylePropObjectOptions turns the configured JSON into the struct the rule reads.
//
// Exported so fixtures drive the same path the config drives. A rule configured as a bare severity
// is handed nil options, and the empty body has to produce an EMPTY allow list rather than an
// error. An empty list is the fully-enforcing state here, which is the opposite direction from the
// two sibling forbid rules where the empty list means the rule declines everything: upstream reads
// `context.options[0].allow || []` and an empty set exempts nothing.
func DecodeStylePropObjectOptions(raw []byte) (any, error) {
	options := StylePropObjectOptions{}
	if len(raw) == 0 {
		return options, nil
	}
	// Lenient on purpose: eslint-plugin-react 7.37.5's schema leaves this object open (no
	// `additionalProperties`), so ESLint loads an extra key and so must this.
	var wire stylePropObjectWireOptions
	if err := json.Unmarshal(raw, &wire); err != nil {
		return options, err
	}
	options.Allow = wire.Allow
	return options, nil
}

// stylePropObjectUnwrapParentheses removes every layer of parentheses from an expression.
//
// Our parser keeps `KindParenthesizedExpression` and upstream's folds it away, so without this the
// rule would go silent on inputs upstream reports. Measured on the installed build, in both the
// value position and through an identifier's initializer:
//
//	<div style={('x')} />              reports
//	<div style={(('x'))} />            reports
//	const s = (('x')); <div style={s}> reports
//	React.createElement("div", {style: ('x')})  reports
//
// A loop rather than a single step, because the nesting is unbounded. Written with its own nil
// check rather than `ast.SkipParentheses`, which dereferences its argument: every caller here is
// handed an optional expression, and an empty container is one of the shapes the parser produces.
func stylePropObjectUnwrapParentheses(expression *ast.Node) *ast.Node {
	for expression != nil && expression.Kind == ast.KindParenthesizedExpression {
		expression = expression.AsParenthesizedExpression().Expression
	}
	return expression
}

// stylePropObjectIsNonNullLiteral reports whether an expression is a literal other than `null`.
//
// This is upstream's `expression.type === 'Literal' && expression.value !== null`, and the set of
// nodes its parser calls a Literal is not obvious from here. Each of these was measured against the
// installed build rather than reasoned about:
//
//	'x'      reports    a string
//	5        reports    a number
//	1n       reports    a bigint
//	true     reports    a boolean, either spelling
//	false    reports
//	/x/      reports    a regular expression IS a Literal upstream
//	null     CLEAN      a Literal whose value is null, which the second test excludes
//	`x`      CLEAN      a template is a TemplateLiteral, a different node type
//	-5       CLEAN      a unary expression wrapping a literal, not a literal
//	[]       CLEAN      an array expression
//	{}       CLEAN      an object expression, which is the whole point of the rule
//
// The regular-expression and bigint answers are the two a port is most likely to get wrong, since
// neither is written anywhere in upstream's corpus and neither reads as a "literal" in the informal
// sense the rule's name suggests.
func stylePropObjectIsNonNullLiteral(expression *ast.Node) bool {
	if expression == nil {
		return false
	}
	switch expression.Kind {
	case ast.KindStringLiteral,
		ast.KindNumericLiteral,
		ast.KindBigIntLiteral,
		ast.KindRegularExpressionLiteral,
		ast.KindTrueKeyword,
		ast.KindFalseKeyword:
		return true
	}
	// `null` is deliberately absent. Upstream's node type covers it and the value test then
	// excludes it, so the two-part test collapses here into naming the kinds that pass. A
	// `KindNullKeyword` arm would invert three of upstream's clean cases.
	return false
}

// StylePropObject flags a style prop whose value is a literal rather than an object.
//
//	valid:   <div style={{ color: "red" }} />
//	valid:   <div style={null} />              a null literal is excluded
//	valid:   <div style={undefined} />         undefined is an identifier, not a literal
//	valid:   <div style />                     no value to read
//	valid:   const s = {}; <div style={s} />   resolved through the declaration
//	invalid: <div style="color: red" />
//	invalid: <div style={true} />
//	invalid: const s = "x"; <div style={s} />
//	invalid: React.createElement("div", { style: "x" })
//
// Ported from `react/style-prop-object` in `eslint-plugin-react`. One option, one message, no
// fixer. All 35 corpus cases were replayed against the installed build, version 7.37.5, through the
// ESLint Linter API before any code was written, and it agreed on all 35. Everything below that the
// corpus does not state was measured the same way.
//
// # The identifier arm resolves through the checker, and it takes the FIRST declaration
//
// Upstream asks scope analysis for the variable and reads `defs[0].node.init`, so a name declared
// twice is judged by the declaration written FIRST. Measured on the installed build:
//
//	var s = 'x'; var s = {};   REPORTS, judged by the string
//	var s = {}; var s = 'x';   CLEAN, judged by the object
//
// This is the one shape where the brief's standing advice to loop over declarations would be wrong.
// The question here is "what single thing is this name", not "does any declaration match", so a
// loop would be strictly wider than upstream and would report the second case too. Probed in our
// tree: `symbol.Declarations` is in source order for this shape and index zero is the first
// declaration, so taking index zero reproduces upstream on both orderings. Both are fixtures.
//
// A declaration that is not a variable declaration, such as an import clause, has no initializer to
// read and is therefore clean, which is upstream's answer through a different route. Probed.
//
// # Parentheses cost findings here rather than adding them
//
// Our parser keeps parenthesis nodes and upstream's folds them away, so `<div style={('x')} />` is
// a parenthesized expression here and a plain string literal there. Without unwrapping, this port
// would be silent on four shapes the installed build reports. All four measured. This is the
// direction of the hazard that no imported fixture can catch, because upstream's corpus cannot
// express a shape its own parser deletes.
//
// # The allow list only skips, it does not gate
//
// A member or namespaced tag has no plain name for the allow list to match, so the lookup simply
// misses and the element is still examined. That is the opposite of the two sibling forbid rules
// ported beside this one, where an unreadable tag name ends the check. Measured: `<a.b style="x" />`
// and `<svg:a style="x" />` both REPORT, and naming `a.b` in the allow list does not exempt it.
//
// # There is no file-suffix gate
//
// Three siblings in this package gate on `.tsx` and `.jsx`, which is oxc residue rather than
// upstream behaviour. The createElement arm needs no JSX at all, which a fixture uses to pin the
// absence of a gate on a plain `.ts` file.
var StylePropObject = rule.Rule{
	Name: "react/style-prop-object",

	// Declared for two reasons. The identifier arm resolves a name to its declaration through
	// `GetSymbolAtLocation`, and `isPragmaCreateElementCall` asks the checker whether a bare
	// `createElement` binds to a react import.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, _ := rule.OptionsAs[StylePropObjectOptions](options)
		allowed := make(map[string]bool, len(settings.Allow))
		for _, name := range settings.Allow {
			allowed[name] = true
		}

		// checkIdentifier reproduces upstream's `checkIdentifiers`: resolve the name, read the
		// first declaration's initializer, and report when that initializer is a non-null literal.
		// A name that resolves to nothing, to a declaration with no initializer, or to a
		// declaration that is not a variable, is clean.
		//
		// `shorthand` is the property node when this identifier is a shorthand property's name,
		// and it is not an optimization. `GetSymbolAtLocation` on such a name resolves to the
		// PROPERTY's own symbol, whose single declaration is the shorthand assignment itself, so
		// the variable declaration is never reached and the rule goes silent on an input upstream
		// reports. Probed directly: the plain accessor answers a `KindShorthandPropertyAssignment`
		// declaration and `GetShorthandAssignmentValueSymbol` answers the `KindVariableDeclaration`.
		// The failure is silent in the costly direction, and no imported fixture caught it because
		// upstream's corpus writes no shorthand.
		checkIdentifier := func(node *ast.Node, shorthand *ast.Node) {
			if ctx.TypeChecker == nil {
				return
			}
			var symbol *ast.Symbol
			if shorthand != nil {
				symbol = ctx.TypeChecker.GetShorthandAssignmentValueSymbol(shorthand)
			} else {
				symbol = ctx.TypeChecker.GetSymbolAtLocation(node)
			}
			if symbol == nil || len(symbol.Declarations) == 0 {
				return
			}
			// Index zero rather than a loop, deliberately, and the doc comment above says why:
			// upstream reads one declaration and a loop would be wider than the rule.
			declaration := symbol.Declarations[0]
			if declaration.Kind != ast.KindVariableDeclaration {
				return
			}
			initializer := stylePropObjectUnwrapParentheses(declaration.AsVariableDeclaration().Initializer)
			if stylePropObjectIsNonNullLiteral(initializer) {
				ctx.ReportNode(node, stylePropObjectMessage)
			}
		}

		return rule.Listeners{
			ast.KindJsxAttribute: func(node *ast.Node) {
				attributeName, named := jsx.AttributeName(node)
				if !named || attributeName != "style" {
					return
				}
				attribute := node.AsJsxAttribute()
				// Upstream's `!node.value` guard. A bare `<div style />` has nothing to judge.
				if attribute.Initializer == nil {
					return
				}

				// The allow list is consulted only to SKIP. A tag with no plain name simply fails
				// to match and the element is still examined, which is upstream's behaviour and
				// the opposite of the sibling forbid rules.
				attributes := node.Parent
				if attributes != nil && attributes.Kind == ast.KindJsxAttributes {
					// Upstream reads the name only when the parent is an opening element, which
					// its tree gives self-closing elements too. `jsx.ElementParts` covers both
					// forms here.
					if tagName, _ := jsx.ElementParts(attributes.Parent); tagName != nil &&
						tagName.Kind == ast.KindIdentifier && allowed[tagName.Text()] {
						return
					}
				}

				// A value that is not an expression container is a plain string attribute, which
				// upstream reports directly. The finding is on the whole attribute in that branch.
				if attribute.Initializer.Kind != ast.KindJsxExpression {
					ctx.ReportNode(node, stylePropObjectMessage)
					return
				}

				expression := stylePropObjectUnwrapParentheses(attribute.Initializer.AsJsxExpression().Expression)
				// An EMPTY container, `<div style={} />`, reports. Upstream reaches its literal
				// test with an undefined expression, which fails, and then falls into the report
				// branch because the container test already passed. Measured; no corpus case
				// writes it. The nil here is that same undefined.
				if stylePropObjectIsNonNullLiteral(expression) || expression == nil {
					ctx.ReportNode(node, stylePropObjectMessage)
					return
				}
				if expression.Kind == ast.KindIdentifier {
					// Not a shorthand: a JSX value position has no property node.
					checkIdentifier(expression, nil)
				}
			},

			ast.KindCallExpression: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				if !isPragmaCreateElementCall(ctx, node) {
					return
				}
				arguments := node.AsCallExpression().Arguments
				// Upstream requires more than one argument, so a one-argument call is clean.
				if arguments == nil || len(arguments.Nodes) < 2 {
					return
				}

				// The allow list matches the first argument's NAME, so it applies to an identifier
				// component and not to a string element. Measured: allowing "div" does not exempt
				// `React.createElement("div", ...)`, because a string literal has no name.
				first := arguments.Nodes[0]
				if first.Kind == ast.KindIdentifier && allowed[first.Text()] {
					return
				}

				properties := arguments.Nodes[1]
				if properties.Kind != ast.KindObjectLiteralExpression {
					return
				}
				for _, property := range properties.AsObjectLiteralExpression().Properties.Nodes {
					// Upstream requires a non-computed key named `style`. A computed key is
					// excluded even when its text is "style", and a spread carries no key at all.
					// Both measured.
					var value *ast.Node
					var shorthand *ast.Node
					switch property.Kind {
					case ast.KindPropertyAssignment:
						assignment := property.AsPropertyAssignment()
						name := assignment.Name()
						if name == nil || name.Kind != ast.KindIdentifier || name.Text() != "style" {
							continue
						}
						value = assignment.Initializer

					case ast.KindShorthandPropertyAssignment:
						// `{ style }` carries its own name as the value. Upstream's `style.value`
						// is the shorthand's identifier, so it takes the identifier arm. Measured:
						// `const style = 'x'; React.createElement("div", { style })` reports.
						name := property.AsShorthandPropertyAssignment().Name()
						if name == nil || name.Kind != ast.KindIdentifier || name.Text() != "style" {
							continue
						}
						value = name
						shorthand = property

					default:
						continue
					}

					value = stylePropObjectUnwrapParentheses(value)
					if value == nil {
						continue
					}
					if value.Kind == ast.KindIdentifier {
						checkIdentifier(value, shorthand)
						return
					}
					if stylePropObjectIsNonNullLiteral(value) {
						// Upstream reports on the VALUE here, not on the property and not on the
						// call, which is a different anchor from the JSX arm above. A span fixture
						// pins both.
						ctx.ReportNode(value, stylePropObjectMessage)
					}
					return
				}
			},
		}
	},
}
