package react

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/ecmascript/jsx"
	"github.com/system-inc/cohere/internal/lint/rule"
)

const (
	buttonHasTypeMissingId   = "missingType"
	buttonHasTypeComplexId   = "complexType"
	buttonHasTypeInvalidId   = "invalidValue"
	buttonHasTypeForbiddenId = "forbiddenValue"
)

var (
	messageButtonMissingType = rule.Message{
		Id: buttonHasTypeMissingId,
		Description: "This `button` has no `type`. HTML defaults a button inside a form to " +
			"`type=\"submit\"`, so a button written for anything else submits the form and reloads " +
			"the page the first time somebody clicks it. Write the type you meant, which is " +
			"usually `type=\"button\"`.",
	}
	messageButtonComplexType = rule.Message{
		Id: buttonHasTypeComplexId,
		Description: "This `button` computes its `type`, so no reader and no tool can tell whether " +
			"it submits. The rule reads a static string or a ternary between static strings, and " +
			"anything else could evaluate to a value that silently submits a form. Write the type " +
			"out, or branch between written-out types.",
	}
)

// buttonHasTypeValueMessage renders the two value messages, which share their text upstream.
//
// `meta.messages` gives `invalidValue` and `forbiddenValue` byte-identical strings and upstream
// reports them as separate ids anyway, because they are separate judgments: the first says the
// value is not one of the three HTML allows, and the second says it is allowed by HTML and turned
// off by this project's configuration. Reproduced as two ids with one text rather than collapsed,
// because a config that switches `reset` off wants to be able to tell "you wrote a typo" from "you
// wrote the thing we banned".
//
// The text is upstream's, verb for verb, with the reason appended after it. The leading sentence is
// asserted by fixture against a literal so a later edit to the wording fails loudly.
func buttonHasTypeValueMessage(id string, value string) rule.Message {
	return rule.Message{
		Id: id,
		Description: fmt.Sprintf(
			"%q is an invalid value for button type attribute. A button type has to be one of "+
				"`submit`, `reset` or `button`, and a browser silently treats anything else as "+
				"`submit`.", value),
	}
}

// ButtonHasTypeOptions is the decoded option object.
//
// Three booleans, one per HTML button type, matching `meta.schema` exactly. All three default to
// TRUE, which is the opposite of Go's zero value, and that inversion is the whole reason the decoder
// below is hand-written; see DecodeButtonHasTypeOptions.
type ButtonHasTypeOptions struct {
	// Button allows `type="button"`. Default true.
	Button bool

	// Submit allows `type="submit"`. Default true.
	Submit bool

	// Reset allows `type="reset"`. Default true.
	Reset bool
}

// DefaultButtonHasTypeOptions is what an unconfigured rule uses.
//
// Upstream's `optionDefaults`, which `Object.assign` merges under whatever the config supplies.
func DefaultButtonHasTypeOptions() ButtonHasTypeOptions {
	return ButtonHasTypeOptions{Button: true, Submit: true, Reset: true}
}

// allows reports whether a written-out type is permitted, and whether it is a type at all.
//
// The two returns are upstream's two branches in `checkValue` and they mean different things. The
// second says the value is not one of the three names in the configuration object at all, which is
// `invalidValue`; the first says it is one of them and is switched off, which is `forbiddenValue`.
// Upstream expresses both against the same object, so a name outside the three can never be
// "forbidden" no matter how the rule is configured.
func (o ButtonHasTypeOptions) allows(value string) (permitted bool, known bool) {
	switch value {
	case "button":
		return o.Button, true
	case "submit":
		return o.Submit, true
	case "reset":
		return o.Reset, true
	}
	return false, false
}

// buttonHasTypeWireOptions is the JSON shape, which is not the struct shape.
//
// Pointers so an absent key is distinguishable from an explicit `false`. Every default here is TRUE,
// so a plain bool field would make `{}` and `{"button": false}` decode identically and silently
// invert the rule, and every fixture built from a struct rather than routed through this decoder
// would pass anyway. That is the default-true trap the port brief names, and this is the shape that
// avoids it.
type buttonHasTypeWireOptions struct {
	Button *bool `json:"button"`
	Submit *bool `json:"submit"`
	Reset  *bool `json:"reset"`
}

// DecodeButtonHasTypeOptions decodes the option object over upstream's defaults.
//
// Hand-written rather than `rule.DecodeOptionsInto` for two reasons, either of which alone would be
// enough. The generic helper errors on EMPTY input, which is what a bare `"error"` configuration
// hands a rule. And its zero value is all-false, which for this rule means every written-out type is
// forbidden, so a rule that fell back to the zero value would report `forbiddenValue` on correct
// code rather than going quiet, which is the loud direction of a silent inversion but still wrong.
func DecodeButtonHasTypeOptions(raw []byte) (any, error) {
	options := DefaultButtonHasTypeOptions()
	if len(raw) == 0 {
		return options, nil
	}

	var wire buttonHasTypeWireOptions
	if err := rule.UnmarshalOptions(raw, &wire); err != nil {
		return options, err
	}
	if wire.Button != nil {
		options.Button = *wire.Button
	}
	if wire.Submit != nil {
		options.Submit = *wire.Submit
	}
	if wire.Reset != nil {
		options.Reset = *wire.Reset
	}
	return options, nil
}

// ButtonHasType flags a `button` element whose type is missing, computed, or not a real type.
//
//	valid:   <button type="button" />
//	valid:   <button type={condition ? "button" : "submit"} />
//	valid:   React.createElement("button", { type: "button" })
//	invalid: <button />
//	invalid: <button type="foo" />
//	invalid: <button type={someExpression} />
//	invalid: React.createElement("button")
//
// Ported from `react/button-has-type` in `eslint-plugin-react`, the implementation that defined it.
// Its 30 clean and 28 reporting cases were extracted mechanically from the clone by stubbing its
// `RuleTester` and serializing the captured object, then every one of the 58 was run against the
// installed build, 7.37.5, through the ESLint Linter API with the TypeScript parser. The two
// authorities agreed on all 58.
//
// # Four message ids, and two of them carry identical text
//
// `invalidValue` and `forbiddenValue` have byte-identical strings in `meta.messages` and are still
// two ids, because they answer different questions; see `buttonHasTypeValueMessage`. The other two
// are `missingType`, for no `type` at all, and `complexType`, for a type this rule declines to
// evaluate.
//
// # complexType points at the EXPRESSION, and everything else points at the element
//
// Upstream's `reportComplex(expression)` takes the offending expression node while every other
// report takes the element or the call. That is not cosmetic: on
// `<button type={condition ? "button" : foo}/>` the finding lands on `foo`, six characters wide, and
// on the element it would be the whole tag. Both spans are asserted by fixture, because
// `ExpectFindings` cannot see where a finding points and a rule anchoring everything on the element
// passes a complete fixture pair while being wrong about the one id whose whole purpose is to point
// at the thing it cannot read.
//
// # A ternary recurses, and reports once per bad branch
//
// `checkExpression` calls itself on the consequent and the alternate, so a nested ternary is walked
// to whatever depth it is written at, and `<button type={condition ? "foo" : "bar"}/>` reports
// TWICE. Measured on the installed build. A rule that stopped at one finding per element would pass
// every corpus case, because the corpus writes no ternary with two bad branches.
//
// # What counts as a static value, measured rather than assumed
//
// Upstream's `Literal` arm stringifies whatever the value is, so several things that read as
// "computed" are actually static and land on `invalidValue` with the stringified value in the
// message. Every row below was measured on the installed build:
//
//	{null} {true} {false} {42} {1.50} {0x10} {1e3} {1_0} {1n} {/re/}   invalidValue
//	  rendering as "null" "true" "false" "42" "1.5" "16" "1000" "10" "1" "/re/"
//	{undefined}                                                        complexType, an Identifier
//	{-1}                                                               complexType, a unary expression
//	{'bu' + 'tton'}                                                    complexType
//	{``}                                                               invalidValue, empty string
//	{`bu${''}tton`}                                                    complexType, it interpolates
//	<button type/>                                                     invalidValue rendering "true"
//
// Our parser hands back the same canonical rendering upstream computes: `.Text()` on a
// NumericLiteral is already `1.5`, `16`, `1000`, `10`, probed against all six forms. The one
// difference is a BigIntLiteral, whose `.Text()` keeps the `n` suffix where upstream renders `1`, so
// that suffix is stripped at the site.
var ButtonHasType = rule.Rule{
	Name:             "react/button-has-type",
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, ok := rule.OptionsAs[ButtonHasTypeOptions](options)
		if !ok {
			// Nil options reach here when the config layer turns a decoder error into nil for a
			// non-required rule. The zero value would forbid all three types and report on correct
			// code, so the defaults are restored explicitly rather than inherited from Go.
			settings = DefaultButtonHasTypeOptions()
		}

		reportMissing := func(node *ast.Node) {
			ctx.ReportNode(node, messageButtonMissingType)
		}

		checkValue := func(node *ast.Node, value string) {
			permitted, known := settings.allows(value)
			if !known {
				ctx.ReportNode(node, buttonHasTypeValueMessage(buttonHasTypeInvalidId, value))
				return
			}
			if !permitted {
				ctx.ReportNode(node, buttonHasTypeValueMessage(buttonHasTypeForbiddenId, value))
			}
		}

		// checkExpression mirrors upstream's recursive walk. Declared before assignment because it
		// calls itself through the conditional arm.
		var checkExpression func(node *ast.Node, expression *ast.Node)
		checkExpression = func(node *ast.Node, expression *ast.Node) {
			if expression == nil {
				// Upstream reaches its default arm with `undefined` when an object property has no
				// value, which is the `{type}` shorthand. Its `switch` on `expression.type` would
				// throw there, so the shorthand is handled at the call site and this is defensive.
				return
			}

			// Upstream's parser produces no parenthesis node at all, so `(b ? "x" : "y")` reaches
			// its ConditionalExpression arm directly. Ours preserves the parentheses, and without
			// this skip a parenthesized branch falls to the default arm and reports complexType
			// where upstream recurses. Measured: `<button type={a ? (b ? "button" : "foo") : "reset"}/>`
			// reports invalidValue upstream and reported complexType here until this line existed.
			// Skipping reproduces upstream rather than improving on it, and `SkipParentheses`
			// dereferences its argument, which is why the nil check above it is not optional.
			expression = ast.SkipParentheses(expression)
			if expression == nil {
				return
			}

			if value, static := buttonHasTypeLiteralValue(expression); static {
				checkValue(node, value)
				return
			}

			switch expression.Kind {
			case ast.KindNoSubstitutionTemplateLiteral:
				// A template with no interpolations is upstream's `expressions.length === 0` arm,
				// which reads `quasis[0].value.raw`. Our parser gives this its own kind, so an
				// interpolating template is a TemplateExpression and never reaches here.
				checkValue(node, expression.Text())

			case ast.KindConditionalExpression:
				conditional := expression.AsConditionalExpression()
				checkExpression(node, conditional.WhenTrue)
				checkExpression(node, conditional.WhenFalse)

			default:
				// Deliberately quieter than upstream, which is type-blind and stops here. When the
				// checker proves every value this expression can take is a written-out type the
				// config permits, nothing is computed in the sense the message means, so there is
				// nothing to report. See buttonHasTypeProvenPermitted.
				if buttonHasTypeProvenPermitted(ctx, expression, settings) {
					return
				}
				// The complex report anchors on the EXPRESSION rather than on the element, which is
				// upstream's `reportComplex(expression)` and is the one place these two nodes differ.
				ctx.ReportNode(expression, messageButtonComplexType)
			}
		}

		return rule.Listeners{
			ast.KindJsxElement: func(node *ast.Node) {
				opening := node.AsJsxElement().OpeningElement
				if opening == nil {
					return
				}
				buttonHasTypeCheckElement(node, opening, reportMissing, checkValue, checkExpression)
			},

			ast.KindJsxSelfClosingElement: func(node *ast.Node) {
				// Upstream listens on `JSXElement`, whose `openingElement` covers both the paired
				// and the self-closing spelling because its parser produces one node for both. Ours
				// produces two kinds, so both are listened for, and the reported node is the element
				// itself in each case, which is what makes the spans match. Measured: a self-closing
				// `<button/>` reports on the whole tag, exactly as the paired form does.
				buttonHasTypeCheckElement(node, node, reportMissing, checkValue, checkExpression)
			},

			ast.KindCallExpression: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				if !isPragmaCreateElementCall(ctx, node) {
					return
				}
				arguments := node.AsCallExpression().Arguments
				if arguments == nil || len(arguments.Nodes) < 1 {
					return
				}
				first := arguments.Nodes[0]
				if first.Kind != ast.KindStringLiteral || first.Text() != "button" {
					return
				}

				if len(arguments.Nodes) < 2 || arguments.Nodes[1].Kind != ast.KindObjectLiteralExpression {
					// No properties object at all, including an explicit `null`, is a missing type.
					reportMissing(node)
					return
				}

				typeProperty := buttonHasTypeFindTypeProperty(arguments.Nodes[1])
				if typeProperty == nil {
					reportMissing(node)
					return
				}

				checkExpression(node, buttonHasTypePropertyValue(typeProperty))
			},
		}
	},
}

// buttonHasTypeCheckElement is the shared body of the two JSX listeners.
//
// `reported` is the node a finding anchors on, which is the whole element, and `opening` is where
// the attributes live. They are the same node for a self-closing element and differ for a paired
// one, which is why they are separate parameters rather than one derived from the other.
//
// The three report paths arrive as closures rather than through the context, because this helper is
// shared by two listeners and threading `ctx` here would let a later reader report from a place the
// rule's own body does not name. Each closure is the same one the rule built.
func buttonHasTypeCheckElement(
	reported *ast.Node,
	opening *ast.Node,
	reportMissing func(node *ast.Node),
	checkValue func(node *ast.Node, value string),
	checkExpression func(node *ast.Node, expression *ast.Node),
) {
	tagName, attributes := jsx.ElementParts(opening)
	if !jsx.IsIntrinsicElementNamed(tagName, "button") {
		return
	}

	typeAttribute := buttonHasTypeFindAttribute(attributes)
	if typeAttribute == nil {
		// A spread does not exempt. `<button {...properties}/>` reports, measured on the installed
		// build, because upstream's `getProp` looks for a NAMED attribute and a spread carries none.
		reportMissing(reported)
		return
	}

	initializer := typeAttribute.AsJsxAttribute().Initializer
	if initializer == nil {
		// `<button type/>`, an attribute with no value. Upstream's `getLiteralPropValue` renders
		// this as the boolean true, so the message reads `"true" is an invalid value`. Measured.
		checkValue(reported, "true")
		return
	}

	if initializer.Kind == ast.KindJsxExpression {
		checkExpression(reported, initializer.AsJsxExpression().Expression)
		return
	}

	// A plain string attribute, `type="button"`.
	checkValue(reported, initializer.Text())
}

// buttonHasTypeFindAttribute returns the `type` attribute node, or nil.
//
// Upstream's `getProp` matches the attribute name exactly and case-sensitively, so `TYPE="button"`
// does not count and the element reports as missing a type. Measured on the installed build.
func buttonHasTypeFindAttribute(attributes *ast.Node) *ast.Node {
	if attributes == nil || attributes.Kind != ast.KindJsxAttributes {
		return nil
	}
	properties := attributes.AsJsxAttributes().Properties
	if properties == nil {
		return nil
	}
	for _, property := range properties.Nodes {
		name, named := jsx.AttributeName(property)
		if named && jsx.MatchExactly(name, "type") {
			return property
		}
	}
	return nil
}

// buttonHasTypeFindTypeProperty returns the object member keyed `type`, or nil.
//
// Upstream requires `'key' in prop && prop.key && 'name' in prop.key && prop.key.name === 'type'`,
// which is an Identifier key and nothing else. So `{'type': 'foo'}` and `{['type']: 'foo'}` are both
// INVISIBLE and the call reports as missing a type rather than as an invalid one, and a spread
// contributes no key at all. All three measured on the installed build.
func buttonHasTypeFindTypeProperty(object *ast.Node) *ast.Node {
	properties := object.AsObjectLiteralExpression().Properties
	if properties == nil {
		return nil
	}
	for _, property := range properties.Nodes {
		name := property.Name()
		if name != nil && name.Kind == ast.KindIdentifier && name.Text() == "type" {
			return property
		}
	}
	return nil
}

// buttonHasTypePropertyValue returns the value expression of an object member.
//
// A shorthand `{type}` has no value node, and upstream passes `undefined` into `checkExpression`,
// whose `switch` falls through to the default arm and reports complex ON THE PROPERTY. So the
// shorthand itself is returned rather than nil, which is what puts the finding in the same place.
// Measured: `React.createElement('button', {type})` reports complexType spanning `type`.
func buttonHasTypePropertyValue(property *ast.Node) *ast.Node {
	switch property.Kind {
	case ast.KindPropertyAssignment:
		return property.AsPropertyAssignment().Initializer
	case ast.KindShorthandPropertyAssignment:
		return property
	}
	return property
}

// buttonHasTypeProvenPermitted reports whether the checker proves an expression can only be one of
// the types the configuration permits.
//
// This is where cohere is deliberately quieter than upstream. `complexType` exists because a
// computed value "could evaluate to a value that silently submits a form", and when the expression's
// type is a union whose every constituent is a string literal naming a permitted type, it cannot.
// The real site is `Button.tsx:222` in ahra, `<button type={type}>` with `type = 'button'`
// destructured from a prop typed `'button' | 'submit' | 'reset'`; ESLint reports it, and the finding
// is false.
//
// Every constituent must be a string literal, so `undefined` (an optional prop with no default, where
// HTML's `submit` default applies), `string`, `any`, an error type and a type parameter all keep
// reporting. Every literal must be PERMITTED rather than merely known, so a config that switches
// `reset` off still reports a union that contains it.
func buttonHasTypeProvenPermitted(ctx rule.Context, expression *ast.Node, settings ButtonHasTypeOptions) bool {
	if ctx.TypeChecker == nil {
		return false
	}
	// A shorthand `{type}` arrives as the property node itself (see buttonHasTypePropertyValue), and
	// the checker types it as the binding it names, so it needs no unwrapping. Measured: unwrapping to
	// its name changed no fixture.
	expressionType := ctx.TypeChecker.GetTypeAtLocation(expression)
	if expressionType == nil {
		return false
	}
	for constituent := range type_checking.UnionTypePartsSeq(expressionType) {
		if !type_checking.IsTypeFlagSet(constituent, checker.TypeFlagsStringLiteral) {
			return false
		}
		// A StringLiteral constituent always carries a string value; were it ever not to, the empty
		// string is not a known type and the proof fails closed below. `allows` answers not-permitted
		// for every unknown value, so one return decides both.
		value, _ := constituent.AsLiteralType().Value().(string)
		if permitted, _ := settings.allows(value); !permitted {
			return false
		}
	}
	return true
}

// buttonHasTypeLiteralValue renders a node upstream would treat as a `Literal`, if it is one.
//
// Upstream's first `checkExpression` arm accepts anything its parser types as `Literal` and calls
// `String(value)` on it, which is why several shapes that read as computed are static here. The set
// and its rendering were measured against the installed build rather than inferred; the table is in
// the rule's doc comment.
//
// The numeric rendering is our parser's own `.Text()`, which is already the canonical form of the
// double: probed over `1.50`, `0x10`, `1e3`, `1_0`, `0.0` and every one matched what upstream
// renders. A BigInt is the single exception, keeping its `n` suffix where upstream drops it, so that
// one is trimmed.
//
// A keyword carries no text at all, which is why the three keyword arms return written-out strings
// rather than calling `.Text()`. Calling it there panics.
func buttonHasTypeLiteralValue(expression *ast.Node) (string, bool) {
	switch expression.Kind {
	case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindRegularExpressionLiteral:
		return expression.Text(), true
	case ast.KindBigIntLiteral:
		return strings.TrimSuffix(expression.Text(), "n"), true
	case ast.KindTrueKeyword:
		return "true", true
	case ast.KindFalseKeyword:
		return "false", true
	case ast.KindNullKeyword:
		return "null", true
	}
	return "", false
}
