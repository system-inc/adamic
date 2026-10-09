package react

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/ecmascript/comments"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var (
	messageJsxCurlyBracePresenceUnnecessaryCurly = rule.Message{
		Id: "UnnecessaryCurly",
		Description: "These curly braces wrap a plain string and do nothing. A JSX expression " +
			"container exists to let a value cross from JavaScript into the markup; around a " +
			"literal that the markup could hold directly it is a pair of braces the reader has " +
			"to look past. Write the literal on its own.",
	}
	messageJsxCurlyBracePresenceMissingCurly = rule.Message{
		Id: "MissingCurly",
		Description: "This literal is written as bare markup where the surrounding code wraps its " +
			"literals in an expression container. Written both ways in one file, the reader has " +
			"to decide each time whether the difference means something. Wrap it.",
	}
)

// jsxCurlyBracePresenceSetting is one of the three values each half of the option surface takes.
type jsxCurlyBracePresenceSetting string

const (
	jsxCurlyBracePresenceAlways jsxCurlyBracePresenceSetting = "always"
	jsxCurlyBracePresenceNever  jsxCurlyBracePresenceSetting = "never"
	jsxCurlyBracePresenceIgnore jsxCurlyBracePresenceSetting = "ignore"
)

// JsxCurlyBracePresenceOptions is the decoded option object.
//
// Upstream accepts either an object with three keys or a bare string that sets two of them, and the
// defaults are NOT symmetric: `props` and `children` default to `never` while `propElementValues`
// defaults to `ignore`. The string form sets `props` and `children` to the named value and forces
// `propElementValues` to `ignore` rather than to the same value, which is why the two forms are not
// interchangeable. Measured: `["always"]` leaves `<App horror=<div /> />` clean, while
// `[{props: "always", children: "always", propElementValues: "always"}]` reports it.
type JsxCurlyBracePresenceOptions struct {
	Props             jsxCurlyBracePresenceSetting
	Children          jsxCurlyBracePresenceSetting
	PropElementValues jsxCurlyBracePresenceSetting
}

// DefaultJsxCurlyBracePresenceOptions is the configuration a bare `"error"` gets.
func DefaultJsxCurlyBracePresenceOptions() JsxCurlyBracePresenceOptions {
	return JsxCurlyBracePresenceOptions{
		Props:             jsxCurlyBracePresenceNever,
		Children:          jsxCurlyBracePresenceNever,
		PropElementValues: jsxCurlyBracePresenceIgnore,
	}
}

// DecodeJsxCurlyBracePresenceOptions decodes either spelling of the option surface.
//
// Hand-written rather than `rule.DecodeOptionsInto` for two reasons the port brief names. The
// generic helper errors on EMPTY input, which is what a bare `"error"` configuration hands a rule,
// and it cannot express a wire shape that is sometimes a string and sometimes an object. The
// defaults here are also not Go zero values, so a struct built by hand rather than routed through
// this function would silently configure the rule as `ignore` on every axis and report nothing.
func DecodeJsxCurlyBracePresenceOptions(raw []byte) (any, error) {
	settings := DefaultJsxCurlyBracePresenceOptions()
	if len(raw) == 0 {
		return settings, nil
	}

	// The bare string form. Upstream reads it as both halves at once, and pins the third axis to
	// `ignore` rather than carrying the value across.
	var shorthand string
	if err := json.Unmarshal(raw, &shorthand); err == nil {
		value := jsxCurlyBracePresenceSetting(shorthand)
		if !jsxCurlyBracePresenceIsValidSetting(value) {
			return settings, nil
		}
		return JsxCurlyBracePresenceOptions{
			Props:             value,
			Children:          value,
			PropElementValues: jsxCurlyBracePresenceIgnore,
		}, nil
	}

	// The object form. Each key is optional and an absent one keeps its default, so the wire
	// struct uses pointers: an absent `props` and an explicit `"props": "never"` must not be
	// distinguishable in the result but must be distinguishable while decoding, since the default
	// for `propElementValues` differs from the other two.
	var wire struct {
		Props             *string `json:"props"`
		Children          *string `json:"children"`
		PropElementValues *string `json:"propElementValues"`
	}
	if err := rule.UnmarshalOptions(raw, &wire); err != nil {
		return settings, err
	}
	for _, field := range []struct {
		value  *string
		target *jsxCurlyBracePresenceSetting
	}{
		{wire.Props, &settings.Props},
		{wire.Children, &settings.Children},
		{wire.PropElementValues, &settings.PropElementValues},
	} {
		if field.value == nil {
			continue
		}
		candidate := jsxCurlyBracePresenceSetting(*field.value)
		if jsxCurlyBracePresenceIsValidSetting(candidate) {
			*field.target = candidate
		}
	}
	return settings, nil
}

func jsxCurlyBracePresenceIsValidSetting(value jsxCurlyBracePresenceSetting) bool {
	switch value {
	case jsxCurlyBracePresenceAlways, jsxCurlyBracePresenceNever, jsxCurlyBracePresenceIgnore:
		return true
	}
	return false
}

// JsxCurlyBracePresence flags curly braces around a JSX literal that does not need them, and in the
// other direction a bare literal where the configuration asks for braces.
//
//	valid:   <App>foo</App>
//	valid:   <App>{' '}</App>
//	valid:   <App prop={'it\'s'} />
//	invalid: <App>{'foo'}</App>
//	invalid: <App prop={'bar'} />
//	invalid: <App prop='bar' />        with `{props: 'always'}`
//
// Ported from `react/jsx-curly-brace-presence` in `eslint-plugin-react`, the implementation that
// defined it. Its 89 clean and 48 reporting cases were extracted from the clone by parsing the
// corpus file with `@babel/parser` and walking its AST, rather than by transcription, so no escape
// could be cooked on the way in. All 134 expressible cases were then replayed against the installed
// build, 7.37.5, through the ESLint Linter API with the TypeScript parser: 133 agreed, and the one
// that did not is discussed under "one corpus case does not apply here" below.
//
// # Escaping is the whole difficulty, and upstream's refusals are reproduced rather than re-derived
//
// Removing the braces from `{'foo'}` is a text substitution, so every case where the literal cannot
// survive unquoted has to be declined. Upstream's `needToEscapeCharacterForJSX` declines three
// classes, and the discrimination is between the RAW source text and the COOKED value, which our
// parser exposes as different things: `Text()` is cooked, and the raw text is the source slice.
//
//	a backslash in the raw       `{'it\'s'}` cooked to `it's` would need requoting
//	a character reference        `{'&middot;'}` unquoted would render as the character
//	`{`, `}`, `<` or `>`         only in a CHILD, where they would end the text node
//
// The last one is why `<App p={'a{b'} />` reports while `<App>{'a{b'}</App>` does not, measured
// both ways against the installed build.
//
// # The quote check is asymmetric between a string and a template, and that asymmetry is upstream's
//
// For a string literal, upstream's own condition `(!containsQuoteCharacters(value) || typeof value
// === 'string')` is always true, because the arm only runs when the value IS a string. So the quote
// test is dead for strings and a quote in the value never declines: `<App>{"it's"}</App>` reports.
// For a TEMPLATE literal the same helper is called without that escape hatch, so any quote in the
// cooked text declines: `<App>{`it's`}</App>` is clean. Both directions measured. This reads like a
// defect in upstream and it is reproduced rather than corrected, because the corpus asserts it:
// case I35 requires `<App>{"foo 'bar'"}</App>` to report.
//
// # The attribute fixer chooses its quote character from the raw text
//
// A repaired attribute is wrapped in double quotes, unless the raw source between the quotes already
// contains a double quote, in which case the raw literal is written back verbatim so its own quoting
// survives. That is what makes `<Foo help={'... "P30D" ...'} />` fix to a single-quoted attribute
// rather than to a broken double-quoted one, which is corpus case I47.
//
// # Upstream oscillates on one option combination, and the port reproduces one pass of it
//
// With `propElementValues: "never"` and `props` at its default `never`, `<App horror={<div />} />`
// fixes to `<App horror=<div /> />`, which the missing-curly arm then fixes straight back. Driven
// through the installed build one pass at a time, it alternates forever, and ESLint's own
// `cohereAndFix` gives up after ten rounds with a circular-fixes warning. Corpus cases I44 and I45
// assert one pass in each direction and this port produces exactly that; the loop is a property of
// the option combination rather than of either fixer, so nothing here can resolve it.
//
// # One corpus case does not apply here
//
// Case I12 and case I13 hold byte-identical source under identical options and assert different
// finding counts, separated only by the parser under test: I12 carries `features:
// ["no-default", "no-ts-new", "no-babel-new"]`, which selects old parsers that model the last two
// children differently. Replayed against the TypeScript parser the installed rule reports four
// findings, which is what I13 asserts, so I13 is the row that describes this tree and I12 is
// recorded here rather than imported. This is the version-gate shape the port brief warns about,
// and it was visible only because the extractor kept the `features` field beside each case.
var JsxCurlyBracePresence = rule.Rule{
	Name: "react/jsx-curly-brace-presence",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, ok := rule.OptionsAs[JsxCurlyBracePresenceOptions](options)
		if !ok {
			settings = DefaultJsxCurlyBracePresenceOptions()
		}

		return rule.Listeners{
			ast.KindJsxExpression: func(node *ast.Node) {
				// Upstream reaches this through two selectors that can both match one node. The
				// element-valued-prop selector runs first and short-circuits, matching the order
				// ESLint visits them in, which matters because both would otherwise report.
				if jsxCurlyBracePresenceIsElementValuedProp(node) {
					if settings.PropElementValues == jsxCurlyBracePresenceNever {
						jsxCurlyBracePresenceReportUnnecessary(ctx, node)
					}
					return
				}
				if jsxCurlyBracePresenceShouldCheckUnnecessary(ctx, node, settings) {
					jsxCurlyBracePresenceLintUnnecessary(ctx, node, settings)
				}
			},

			ast.KindJsxText: func(node *ast.Node) {
				if jsxCurlyBracePresenceShouldCheckMissing(node, settings) {
					jsxCurlyBracePresenceReportMissing(ctx, node)
				}
			},

			// A string-valued attribute, `prop="bar"`. Upstream's third selector names `Literal`,
			// which in its parser covers this position; here the node is a plain StringLiteral
			// whose parent is the attribute.
			ast.KindStringLiteral: func(node *ast.Node) {
				if node.Parent == nil || node.Parent.Kind != ast.KindJsxAttribute {
					return
				}
				if jsxCurlyBracePresenceShouldCheckMissing(node, settings) {
					jsxCurlyBracePresenceReportMissing(ctx, node)
				}
			},

			// An element written as a bare attribute value, `<App horror=<div /> />`. Upstream's
			// selector is `JSXAttribute > JSXElement`. Our parser accepts this in a `.tsx` file,
			// which was measured rather than assumed: the corpus marks these cases `no-ts` because
			// older TypeScript parsers rejected them, and this one does not.
			ast.KindJsxSelfClosingElement: func(node *ast.Node) {
				jsxCurlyBracePresenceCheckBareElementAttribute(ctx, node, settings)
			},
			ast.KindJsxElement: func(node *ast.Node) {
				jsxCurlyBracePresenceCheckBareElementAttribute(ctx, node, settings)
			},
		}
	},
}

// jsxCurlyBracePresenceCheckBareElementAttribute handles `<App horror=<div /> />`.
func jsxCurlyBracePresenceCheckBareElementAttribute(
	ctx rule.Context,
	node *ast.Node,
	settings JsxCurlyBracePresenceOptions,
) {
	if node.Parent == nil || node.Parent.Kind != ast.KindJsxAttribute {
		return
	}
	if settings.PropElementValues == jsxCurlyBracePresenceIgnore {
		return
	}
	if settings.PropElementValues == jsxCurlyBracePresenceAlways {
		jsxCurlyBracePresenceReportMissing(ctx, node)
	}
}

// jsxCurlyBracePresenceIsElementValuedProp matches `JSXAttribute > JSXExpressionContainer >
// JSXElement`, the shape upstream gives its own listener.
func jsxCurlyBracePresenceIsElementValuedProp(container *ast.Node) bool {
	if container.Parent == nil || container.Parent.Kind != ast.KindJsxAttribute {
		return false
	}
	expression := jsxCurlyBracePresenceUnwrapParentheses(container.AsJsxExpression().Expression)
	return expression != nil && jsxCurlyBracePresenceIsJsx(expression)
}

// jsxCurlyBracePresenceIsJsx reproduces `jsxUtil.isJSX`: an element or a fragment, nothing else.
//
// Upstream names `JSXElement` and `JSXFragment`. Our parser splits the first into a paired
// `KindJsxElement` and a `KindJsxSelfClosingElement`, and both are the same thing to this rule.
func jsxCurlyBracePresenceIsJsx(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindJsxElement, ast.KindJsxSelfClosingElement, ast.KindJsxFragment:
		return true
	}
	return false
}

// jsxCurlyBracePresenceShouldCheckUnnecessary reproduces `shouldCheckForUnnecessaryCurly`.
func jsxCurlyBracePresenceShouldCheckUnnecessary(
	ctx rule.Context,
	container *ast.Node,
	settings JsxCurlyBracePresenceOptions,
) bool {
	parent := container.Parent
	if parent == nil {
		return false
	}

	// A prop holding anything other than a string or a template is left alone, because
	// `<App prop1={<CustomEl />} />` is not a candidate for unwrapping. The element-valued case is
	// handled by its own listener above and never reaches here.
	if parent.Kind == ast.KindJsxAttribute {
		expression := jsxCurlyBracePresenceUnwrapParentheses(container.AsJsxExpression().Expression)
		if expression != nil && !jsxCurlyBracePresenceIsStringOrTemplate(expression) {
			return false
		}
	}

	if jsxCurlyBracePresenceIsJsx(parent) {
		children := jsxCurlyBracePresenceChildrenOf(parent)
		if jsxCurlyBracePresenceHasAdjacentContainers(container, children) {
			return false
		}
		// This guard survives mutation ALONE and is not dead, which took two sites to establish.
		// A whitespace-only container beside an element is also declined further down, by the
		// child arm's edge-whitespace test, so neutralizing either one changes no output. Both
		// were neutralized together and all four shapes went from clean to reporting, which is
		// what shows the pair is load-bearing rather than one member of it. Kept because it is
		// upstream's, and because the two tests are not equivalent: this one asks about the
		// container's SIBLINGS and the later one about the value, so a change to either would
		// leave the other covering a different set.
		if jsxCurlyBracePresenceContainsWhitespaceExpression(container) &&
			jsxCurlyBracePresenceHasAdjacentJsx(container, children) {
			return false
		}
		// Subsumed the same way the guard above is, by the same later test, and kept for the same
		// reason. Neutralized alone it changes nothing; neutralized together with the child arm's
		// edge-whitespace decline, a lone whitespace container starts reporting. Upstream carries
		// both and so does this.
		if len(children) == 1 && jsxCurlyBracePresenceContainsWhitespaceExpression(container) {
			return false
		}
	}

	return jsxCurlyBracePresenceConditionsSatisfied(parent, settings, jsxCurlyBracePresenceNever)
}

// jsxCurlyBracePresenceShouldCheckMissing reproduces `shouldCheckForMissingCurly`.
func jsxCurlyBracePresenceShouldCheckMissing(node *ast.Node, settings JsxCurlyBracePresenceOptions) bool {
	if jsxCurlyBracePresenceIsJsx(node) {
		return settings.PropElementValues != jsxCurlyBracePresenceIgnore
	}

	raw := jsxCurlyBracePresenceRawTextOf(node)
	if jsxCurlyBracePresenceIsLineBreak(raw) || jsxCurlyBracePresenceContainsOnlyCharacterReferences(raw) {
		return false
	}

	parent := node.Parent
	if parent == nil {
		return false
	}
	children := jsxCurlyBracePresenceChildrenOf(parent)
	if len(children) == 1 && jsxCurlyBracePresenceContainsWhitespaceExpression(children[0]) {
		return false
	}

	return jsxCurlyBracePresenceConditionsSatisfied(parent, settings, jsxCurlyBracePresenceAlways)
}

// jsxCurlyBracePresenceConditionsSatisfied reproduces `areRuleConditionsSatisfied`.
//
// The two halves of the option surface are keyed on WHERE the node sits: inside an attribute the
// `props` setting decides, inside an element or a fragment the `children` setting does.
func jsxCurlyBracePresenceConditionsSatisfied(
	parent *ast.Node,
	settings JsxCurlyBracePresenceOptions,
	wanted jsxCurlyBracePresenceSetting,
) bool {
	if parent.Kind == ast.KindJsxAttribute {
		return settings.Props == wanted
	}
	if jsxCurlyBracePresenceIsJsx(parent) {
		return settings.Children == wanted
	}
	return false
}

// jsxCurlyBracePresenceLintUnnecessary reproduces `lintUnnecessaryCurly`.
func jsxCurlyBracePresenceLintUnnecessary(
	ctx rule.Context,
	container *ast.Node,
	settings JsxCurlyBracePresenceOptions,
) {
	expression := jsxCurlyBracePresenceUnwrapParentheses(container.AsJsxExpression().Expression)
	if expression == nil {
		return
	}

	// Braces holding a comment are load-bearing: `{/* why */}` has nowhere else to put the comment
	// and `{/* why */ 'foo'}` would lose it. Upstream asks `sourceCode.getCommentsInside`, which
	// sees a comment on either side of the expression; a leading-trivia check would miss a trailing
	// one, measured on `{'foo' /* trailing */}`, which upstream leaves clean.
	if jsxCurlyBracePresenceHasCommentInside(ctx, container) {
		return
	}

	parent := container.Parent
	inAttribute := parent != nil && parent.Kind == ast.KindJsxAttribute

	switch {
	case expression.Kind == ast.KindStringLiteral:
		value := expression.Text()
		raw := jsxCurlyBracePresenceRawTextOf(expression)

		// Whitespace is the one thing a container can express that bare markup cannot, so a
		// whitespace-only value keeps its braces. In an attribute the test is only that; in a child
		// any leading or trailing whitespace at all is enough, because JSX would trim it.
		//
		// The attribute test is upstream's `isWhiteSpaceLiteral`, whose `node.value &&` makes an
		// EMPTY string not whitespace, so `label={''}` reports and repairs to `label=""`. Testing
		// the value against `/^\s*$/` alone, which an empty string satisfies, silenced it: ESLint
		// reported two sites in www-phi-health that this did not (#t5dwy1t).
		if inAttribute {
			if jsxCurlyBracePresenceIsWhitespaceLiteral(expression) {
				return
			}
		} else if jsxCurlyBracePresenceHasEdgeWhitespace(value) {
			return
		}
		if strings.Contains(value, "/*") {
			return
		}
		if jsxCurlyBracePresenceNeedsEscaping(raw, inAttribute) {
			return
		}
		// Upstream's quote test on this arm is `!containsQuoteCharacters(value) || typeof value
		// === 'string'`, and the second half is true whenever this arm runs, so the test never
		// declines anything. Reproduced as its absence rather than written out and negated, with
		// the reasoning recorded here: corpus case I35 asserts that `{"foo 'bar'"}` reports, so the
		// dead test is not merely inert, it is required to be inert.
		jsxCurlyBracePresenceReportUnnecessary(ctx, container)

	case expression.Kind == ast.KindNoSubstitutionTemplateLiteral:
		// A template with substitutions is a different node kind and never reaches here, which is
		// how `expression.expressions.length === 0` is expressed.
		raw := jsxCurlyBracePresenceRawTextOf(expression)
		inner := jsxCurlyBracePresenceTemplateInner(raw)
		if strings.Contains(inner, "\n") {
			return
		}
		if jsxCurlyBracePresenceHasEdgeWhitespace(inner) {
			return
		}
		if jsxCurlyBracePresenceNeedsEscaping(inner, inAttribute) {
			return
		}
		// Unlike the string arm above, this one's quote test is live: upstream calls
		// `containsQuoteCharacters` on the COOKED text with no escape hatch, so any quote declines.
		// Measured both ways against the installed build: `{"it's"}` reports and `` {`it's`} ``
		// does not.
		if strings.ContainsAny(expression.Text(), "'\"") {
			return
		}
		jsxCurlyBracePresenceReportUnnecessary(ctx, container)

	case jsxCurlyBracePresenceIsJsx(expression):
		jsxCurlyBracePresenceReportUnnecessary(ctx, container)
	}
}

// jsxCurlyBracePresenceUnwrapParentheses removes any parentheses around an expression.
//
// This has no counterpart upstream and it exists for a parser difference. typescript-eslint's tree
// has no node for a parenthesis, so `{('bar')}` reaches upstream's rule as the bare literal and
// reports; ours keeps a `KindParenthesizedExpression` and the literal arm never matched, so all
// three parenthesized shapes went SILENT here while upstream repairs them. No imported fixture can
// see this, because upstream's parser deleted the node before its corpus was written.
//
// The loop rather than a single step, because `(('bar'))` nests, and `ast.SkipParentheses` is not
// used because it dereferences its argument and a container's expression is optional.
//
// The repair is unaffected by the unwrapping: it is built from the CONTAINER's span and the
// unwrapped expression's value, so `<App>{('bar')}</App>` becomes `<App>bar</App>` and the
// parentheses go with the braces, which is what upstream writes. Measured against the installed
// build on all three shapes.
func jsxCurlyBracePresenceUnwrapParentheses(expression *ast.Node) *ast.Node {
	for expression != nil && expression.Kind == ast.KindParenthesizedExpression {
		expression = expression.AsParenthesizedExpression().Expression
	}
	return expression
}

// jsxCurlyBracePresenceIsStringOrTemplate answers the parent-attribute guard in
// `shouldCheckForUnnecessaryCurly`.
func jsxCurlyBracePresenceIsStringOrTemplate(expression *ast.Node) bool {
	switch expression.Kind {
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral, ast.KindTemplateExpression:
		return true
	}
	return false
}

// jsxCurlyBracePresenceSpanOf returns the range a finding points at and a repair writes over.
//
// `rule.TokenRange` is right for every kind here except JSX TEXT, and wrong for that one in a way
// that corrupts the repair rather than merely misplacing the finding. It skips leading trivia, and
// the newline and indentation in front of a text child is not trivia, it is part of the text: the
// range would start after the indent while the replacement text still carries it, so applying the
// fix leaves the original indentation AND writes it again, adding a blank line to the file for
// every child it repairs. Four of upstream's own repair vectors caught this once the repairs were
// asserted; none of the message-id fixtures over the same inputs could see it.
func jsxCurlyBracePresenceSpanOf(ctx rule.Context, node *ast.Node) core.TextRange {
	if node.Kind == ast.KindJsxText {
		return node.Loc
	}
	return rule.TokenRange(ctx.SourceFile, node)
}

// jsxCurlyBracePresenceReportUnnecessary reports the container and offers to unwrap it.
func jsxCurlyBracePresenceReportUnnecessary(ctx rule.Context, container *ast.Node) {
	span := jsxCurlyBracePresenceSpanOf(ctx, container)
	replacement, ok := jsxCurlyBracePresenceUnwrappedText(ctx, container)
	if !ok {
		ctx.ReportNode(container, messageJsxCurlyBracePresenceUnnecessaryCurly)
		return
	}
	ctx.ReportNodeWithFixes(container, messageJsxCurlyBracePresenceUnnecessaryCurly,
		rule.ReplaceRange(span, replacement))
}

// jsxCurlyBracePresenceUnwrappedText builds the text that replaces `{...}`.
//
// Every branch writes a slice of the ORIGINAL source or a value derived from it, and the span is
// the container's own token range, so nothing outside the braces is inside the edit. That
// enumeration is the check the port brief asks for on any fixer that constructs a replacement: what
// lives in this span is the two braces and the expression between them, and the expression is
// copied rather than rebuilt except where a quote character has to change.
func jsxCurlyBracePresenceUnwrappedText(ctx rule.Context, container *ast.Node) (string, bool) {
	expression := jsxCurlyBracePresenceUnwrapParentheses(container.AsJsxExpression().Expression)
	if expression == nil {
		return "", false
	}

	// An element or fragment value is written back byte for byte.
	if jsxCurlyBracePresenceIsJsx(expression) {
		return jsxCurlyBracePresenceRawTextOf(expression), true
	}

	parent := container.Parent
	if parent != nil && parent.Kind == ast.KindJsxAttribute {
		// An attribute value has to end up inside quotes, so its quoting is the thing being
		// decided. Upstream writes the raw literal back UNCHANGED when the text between its quotes
		// already contains a double quote, which preserves a single-quoted literal that could not
		// be double-quoted without escaping; otherwise it wraps the inner text in double quotes.
		raw := jsxCurlyBracePresenceRawTextOf(expression)
		inner := jsxCurlyBracePresenceStripDelimiters(raw)
		if expression.Kind != ast.KindNoSubstitutionTemplateLiteral && strings.Contains(inner, "\"") {
			return raw, true
		}
		return "\"" + inner + "\"", true
	}

	// A child is written as its COOKED value, which is what the markup would have held. A template
	// uses the same cooked text; upstream reads `quasis[0].value.cooked` here and `.raw` in the
	// attribute branch above, and the difference is deliberate on its side.
	return expression.Text(), true
}

// jsxCurlyBracePresenceReportMissing reports a bare literal and offers to wrap it.
func jsxCurlyBracePresenceReportMissing(ctx rule.Context, node *ast.Node) {
	span := jsxCurlyBracePresenceSpanOf(ctx, node)
	replacement, ok := jsxCurlyBracePresenceWrappedText(ctx, node)
	if !ok {
		// Upstream returns null from its fixer for three shapes it can see no single right answer
		// for, and a declined repair is reproduced as a declined repair.
		ctx.ReportNode(node, messageJsxCurlyBracePresenceMissingCurly)
		return
	}
	ctx.ReportNodeWithFixes(node, messageJsxCurlyBracePresenceMissingCurly,
		rule.ReplaceRange(span, replacement))
}

// jsxCurlyBracePresenceWrappedText builds the text that replaces a bare literal, or declines.
//
// The three declines are upstream's, at the line where it returns null: a literal that is only
// character references, because the repair could equally write the reference or the character it
// names; an ATTRIBUTE literal containing a line terminator; and a child that is a line break with
// nothing else on it.
func jsxCurlyBracePresenceWrappedText(ctx rule.Context, node *ast.Node) (string, bool) {
	if jsxCurlyBracePresenceIsJsx(node) {
		return "{" + jsxCurlyBracePresenceRawTextOf(node) + "}", true
	}

	raw := jsxCurlyBracePresenceRawTextOf(node)
	inAttribute := node.Parent != nil && node.Parent.Kind == ast.KindJsxAttribute

	// Upstream's three declines, two of which are unreachable HERE and are kept anyway because
	// they are upstream's and because reachability is a property of the callers rather than of
	// this function. Measured by neutralizing each and looking for an input whose output moves:
	//
	//	only character references   an ATTRIBUTE's raw text carries its own quotes, so it is never
	//	                            only references; a CHILD that is, is declined a level earlier
	//	                            by shouldCheckForMissingCurly and never reports
	//	a line terminator in an     the only one that fires, on `<App p="a\nb" />`
	//	  attribute
	//	a child that is a line      declined a level earlier by the same caller, and an attribute
	//	  break                     line break is taken by the branch above it first
	//
	// Both unreachable arms survive mutation on their own, which is what an unreachable branch
	// does. Deleting them would make this function disagree with the original for no gain, and the
	// verdict expires if a caller changes, which is why it is written down rather than acted on.
	if jsxCurlyBracePresenceContainsOnlyCharacterReferences(raw) ||
		(inAttribute && jsxCurlyBracePresenceContainsLineTerminator(raw)) ||
		jsxCurlyBracePresenceIsLineBreak(raw) {
		return "", false
	}

	if inAttribute {
		inner := jsxCurlyBracePresenceStripDelimiters(raw)
		return "{\"" + jsxCurlyBracePresenceEscapeDoubleQuotes(
			jsxCurlyBracePresenceEscapeBackslashes(inner)) + "\"}", true
	}
	return jsxCurlyBracePresenceWrapChildText(raw), true
}

// jsxCurlyBracePresenceWrapChildText reproduces `wrapWithCurlyBraces`.
//
// A single-line child becomes one container holding a JSON-quoted string. A multi-line one is
// rewritten line by line so the indentation survives, and a line holding a character reference is
// split around it so the reference stays bare markup rather than being quoted into a literal.
func jsxCurlyBracePresenceWrapChildText(raw string) string {
	if !jsxCurlyBracePresenceContainsLineTerminator(raw) {
		return "{" + jsxCurlyBracePresenceQuote(raw) + "}"
	}

	lines := strings.Split(raw, "\n")
	for index, line := range lines {
		// Not `strings.TrimSpace`, whose character set is one rune wider than JavaScript's; see
		// the helper. That rune is reachable and changes what the repair writes, and swapping the
		// two spellings is caught by a fixture.
		//
		// The GUARD itself, as opposed to the character set it uses, is equivalent to the
		// `firstCharacter < 0` test below: both ask whether the line holds a non-whitespace
		// character, now through the same set, so no line can take one path and not the other.
		// Kept because it is upstream's and because the two would diverge again the moment either
		// character set moved, which is exactly the bug the fixture above pins.
		if jsxCurlyBracePresenceTrimLikeJavaScript(line) == "" {
			continue
		}
		// And here the `\s` set, because upstream's is `line.search(/[^\s]/)`.
		firstCharacter := strings.IndexFunc(line, func(character rune) bool {
			return !jsxCurlyBracePresenceIsWhitespaceRune(character)
		})
		if firstCharacter < 0 {
			continue
		}
		leading := line[:firstCharacter]
		text := line[firstCharacter:]
		if jsxCurlyBracePresenceContainsCharacterReference(text) {
			lines[index] = leading + jsxCurlyBracePresenceWrapAroundReferences(text)
			continue
		}
		lines[index] = leading + "{" + jsxCurlyBracePresenceQuote(text) + "}"
	}
	return strings.Join(lines, "\n")
}

// jsxCurlyBracePresenceTrimLikeJavaScript reproduces `String.prototype.trim`.
//
// Its character set is exactly `\s`, so this is a thin wrapper and NOT `strings.TrimSpace`. The
// difference is one rune and it changes the output: Go's `unicode.IsSpace` counts U+0085 and
// JavaScript counts it in neither `trim` nor `\s`, so `TrimSpace` calls a line holding only that
// character blank and skips it, while upstream treats the character as the line's content and
// writes an empty container around it. Measured against the installed build, which produces
// `{""}` for that line where the Go-native spelling produced the bare character.
//
// I had this backwards once. Reading `"\u0085".trim()` in a terminal shows nothing and reads as an
// empty string; the comparison `=== ""` is what actually answers the question, and it is false.
func jsxCurlyBracePresenceTrimLikeJavaScript(text string) string {
	return strings.TrimFunc(text, jsxCurlyBracePresenceIsWhitespaceRune)
}

// jsxCurlyBracePresenceWrapAroundReferences reproduces `wrapNonHTMLEntities`.
//
// The text is split on every character reference, each surviving run is wrapped on its own, and the
// references are put back where they were. Upstream does this with a placeholder token and a
// reduce; the same result is reached here by interleaving the split pieces with the matches, which
// avoids the placeholder being confusable with real text.
func jsxCurlyBracePresenceWrapAroundReferences(text string) string {
	references := jsxCurlyBracePresenceFindCharacterReferences(text)
	if len(references) == 0 {
		return "{" + jsxCurlyBracePresenceQuote(text) + "}"
	}

	var builder strings.Builder
	previousEnd := 0
	for _, reference := range references {
		piece := text[previousEnd:reference[0]]
		if piece != "" {
			builder.WriteString("{" + jsxCurlyBracePresenceQuote(piece) + "}")
		}
		builder.WriteString(text[reference[0]:reference[1]])
		previousEnd = reference[1]
	}
	if tail := text[previousEnd:]; tail != "" {
		builder.WriteString("{" + jsxCurlyBracePresenceQuote(tail) + "}")
	}
	return builder.String()
}

// jsxCurlyBracePresenceQuote renders a string the way `JSON.stringify` renders one.
//
// Two encoders were wrong here before this one. `strconv.Quote` escapes every non-ASCII rune as a
// numeric escape where JavaScript emits the character itself, which would rewrite a repaired line
// of Japanese into escapes. Plain `json.Marshal` is right about that and wrong about three ASCII
// characters: it escapes the less-than, greater-than and ampersand for safety when the result is
// embedded in a document, so a repaired `&middot;` came out as a numeric escape. An encoder with
// that behaviour switched off matches JavaScript on both counts.
//
// The trailing newline is the encoder's, not the value's: `json.Encoder` writes one after each
// value and it has to come back off.
func jsxCurlyBracePresenceQuote(text string) string {
	var builder strings.Builder
	encoder := json.NewEncoder(&builder)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(text); err != nil {
		return strconv.Quote(text)
	}
	return strings.TrimSuffix(builder.String(), "\n")
}

// jsxCurlyBracePresenceChildrenOf returns the children of an element or a fragment.
func jsxCurlyBracePresenceChildrenOf(node *ast.Node) []*ast.Node {
	var list *ast.NodeList
	switch node.Kind {
	case ast.KindJsxElement:
		list = node.AsJsxElement().Children
	case ast.KindJsxFragment:
		list = node.AsJsxFragment().Children
	}
	if list == nil {
		return nil
	}
	return list.Nodes
}

// jsxCurlyBracePresenceHasAdjacentContainers reproduces `hasAdjacentJsxExpressionContainers`.
//
// Two containers written side by side are how a JSX author writes two values with nothing between
// them, and unwrapping either would join them into one text node. So the pair is left alone.
func jsxCurlyBracePresenceHasAdjacentContainers(node *ast.Node, children []*ast.Node) bool {
	for _, sibling := range jsxCurlyBracePresenceAdjacentSiblings(node, children) {
		if sibling.Kind == ast.KindJsxExpression {
			return true
		}
	}
	return false
}

// jsxCurlyBracePresenceHasAdjacentJsx reproduces `hasAdjacentJsx`.
func jsxCurlyBracePresenceHasAdjacentJsx(node *ast.Node, children []*ast.Node) bool {
	for _, sibling := range jsxCurlyBracePresenceAdjacentSiblings(node, children) {
		if sibling.Kind == ast.KindJsxExpression || sibling.Kind == ast.KindJsxElement {
			return true
		}
	}
	return false
}

// jsxCurlyBracePresenceAdjacentSiblings reproduces `getAdjacentSiblings`, filtering first.
//
// Upstream filters whitespace-only literals out of the list before asking, so a container separated
// from its neighbour only by indentation still counts as adjacent. Its loop over the interior runs
// `for (let i = 1; i < children.length - 1; i++)`, which deliberately excludes both ends, and the
// two ends are then handled by the pair of checks after it. Reproducing the bounds exactly matters:
// a node that is BOTH first and last has no siblings under this definition even when the list has
// other members, and widening the loop would change which inputs report.
func jsxCurlyBracePresenceAdjacentSiblings(node *ast.Node, children []*ast.Node) []*ast.Node {
	filtered := make([]*ast.Node, 0, len(children))
	for _, child := range children {
		if jsxCurlyBracePresenceIsWhitespaceLiteral(child) {
			continue
		}
		filtered = append(filtered, child)
	}

	for index := 1; index < len(filtered)-1; index++ {
		if filtered[index] == node {
			return []*ast.Node{filtered[index-1], filtered[index+1]}
		}
	}
	if len(filtered) > 1 && filtered[0] == node {
		return []*ast.Node{filtered[1]}
	}
	if last := len(filtered) - 1; last >= 1 && filtered[last] == node {
		return []*ast.Node{filtered[last-1]}
	}
	return nil
}

// jsxCurlyBracePresenceIsWhitespaceLiteral reproduces `isWhiteSpaceLiteral`.
//
// Two details, and getting either wrong changes which inputs report.
//
// It tests `node.type === 'Literal'`, which does NOT include a JSXText node. The indentation
// between two children is a JSXText, so it is NOT filtered out of the sibling list and a container
// separated from its neighbour by a newline is therefore NOT adjacent to it. Reading the function's
// name rather than its condition costs findings in one direction only: corpus case I13 lays four
// containers out one per line and expects all four to report, and treating the indentation as
// filterable makes the last two read as adjacent and silences them. The two children lists were
// compared element for element, ours against the installed build's, and they are identical.
//
// And it tests `node.value &&`, so an EMPTY string is falsy and is not whitespace. That is
// preserved: an empty string literal in a container is not a whitespace literal and does report.
func jsxCurlyBracePresenceIsWhitespaceLiteral(node *ast.Node) bool {
	if node.Kind != ast.KindStringLiteral {
		return false
	}
	value := node.Text()
	if value == "" {
		return false
	}
	return jsxCurlyBracePresenceIsAllWhitespace(value)
}

// jsxCurlyBracePresenceLiteralValue reads the cooked value of a text or string child.
//
// The `Text` FIELD is read for JSX text rather than `Node.Text()`, which panics on some kinds; this
// is the same accessor `jsx-no-useless-fragment` uses beside it for the same reason.
func jsxCurlyBracePresenceLiteralValue(node *ast.Node) (string, bool) {
	switch node.Kind {
	case ast.KindJsxText:
		return node.AsJsxText().Text, true
	case ast.KindStringLiteral:
		return node.Text(), true
	}
	return "", false
}

// jsxCurlyBracePresenceContainsWhitespaceExpression reproduces `containsWhitespaceExpression`.
func jsxCurlyBracePresenceContainsWhitespaceExpression(child *ast.Node) bool {
	if child.Kind != ast.KindJsxExpression {
		return false
	}
	expression := jsxCurlyBracePresenceUnwrapParentheses(child.AsJsxExpression().Expression)
	if expression == nil {
		return false
	}
	value, ok := jsxCurlyBracePresenceLiteralValue(expression)
	if !ok || value == "" {
		return false
	}
	return jsxCurlyBracePresenceIsAllWhitespace(value)
}

// jsxCurlyBracePresenceHasCommentInside reports whether a comment sits inside the braces.
//
// `comments.ForFile` is the shelf's cached per-file scan, so asking it once per container costs one
// scan for the whole file rather than one per node.
func jsxCurlyBracePresenceHasCommentInside(ctx rule.Context, container *ast.Node) bool {
	start := scanner.SkipTrivia(ctx.SourceFile.Text(), container.Pos())
	end := container.End()
	for _, comment := range comments.ForFile(ctx) {
		if comment.Range.Pos() >= start && comment.Range.End() <= end {
			return true
		}
	}
	return false
}

// jsxCurlyBracePresenceRawTextOf returns a node's text as it is written in the source.
//
// `Node.Text()` is the COOKED value for a literal, so `'it\'s'` comes back as `it's` with the
// backslash gone and the quotes stripped. Every escaping decision in this rule is about the raw
// form, so the source is sliced instead.
//
// `Pos()` includes leading trivia, which has to be skipped for an EXPRESSION and must not be
// skipped for JSX TEXT. Whitespace between two JSX children is text content rather than trivia,
// and `scanner.SkipTrivia` consumes all of it: handed the indentation between two elements it
// returns the node's own end, so the raw text comes back empty, the line-break test that should
// have declined it sees no line break, and the rule reports a finding on every indent in the file.
// Measured on a three-level nesting from upstream's corpus, which went from clean to eight
// findings on that one difference.
func jsxCurlyBracePresenceRawTextOf(node *ast.Node) string {
	sourceFile := ast.GetSourceFileOfNode(node)
	if sourceFile == nil {
		return ""
	}
	text := sourceFile.Text()
	start := node.Pos()
	if node.Kind != ast.KindJsxText {
		start = scanner.SkipTrivia(text, start)
	}
	end := node.End()
	if start < 0 || end > len(text) || start > end {
		return ""
	}
	return text[start:end]
}

// jsxCurlyBracePresenceStripDelimiters removes the opening and closing quote or backtick.
func jsxCurlyBracePresenceStripDelimiters(raw string) string {
	if len(raw) < 2 {
		return raw
	}
	return raw[1 : len(raw)-1]
}

// jsxCurlyBracePresenceTemplateInner returns a template literal's raw text without its backticks.
func jsxCurlyBracePresenceTemplateInner(raw string) string {
	return jsxCurlyBracePresenceStripDelimiters(raw)
}

// jsxCurlyBracePresenceNeedsEscaping reproduces `needToEscapeCharacterForJSX`.
//
// A backslash or a character reference declines anywhere, because unquoting either would change
// what the text means. The third class, the four characters that delimit JSX itself, declines only
// in a CHILD: inside an attribute's quotes they are ordinary. Measured both ways.
func jsxCurlyBracePresenceNeedsEscaping(raw string, inAttribute bool) bool {
	if strings.Contains(raw, `\`) {
		return true
	}
	if jsxCurlyBracePresenceContainsCharacterReference(raw) {
		return true
	}
	if !inAttribute && strings.ContainsAny(raw, "{}<>") {
		return true
	}
	return false
}

// jsxCurlyBracePresenceEscapeDoubleQuotes reproduces `escapeDoubleQuotes`.
//
// Upstream unescapes any already-escaped double quote first and then escapes every double quote,
// which collapses a doubled escape rather than deepening it. The two steps are not interchangeable
// and the order is upstream's.
func jsxCurlyBracePresenceEscapeDoubleQuotes(text string) string {
	return strings.ReplaceAll(strings.ReplaceAll(text, `\"`, `"`), `"`, `\"`)
}

// jsxCurlyBracePresenceEscapeBackslashes reproduces `escapeBackslashes`.
func jsxCurlyBracePresenceEscapeBackslashes(text string) string {
	return strings.ReplaceAll(text, `\`, `\\`)
}

// jsxCurlyBracePresenceCharacterReferencePattern matches what upstream's `HTML_ENTITY_REGEX`
// matches: an ampersand, one or more letters, digits or a hash, and a semicolon.
//
// Written as a scan rather than as a compiled expression because it runs on every literal the rule
// looks at and the grammar is three character classes deep. The pattern is deliberately as loose as
// upstream's: it accepts `&1;` and `&###;`, neither of which names a real character, and narrowing
// it would decline fewer inputs than upstream declines.
func jsxCurlyBracePresenceFindCharacterReferences(text string) [][2]int {
	var found [][2]int
	for index := 0; index < len(text); index++ {
		if text[index] != '&' {
			continue
		}
		cursor := index + 1
		for cursor < len(text) && jsxCurlyBracePresenceIsReferenceBodyByte(text[cursor]) {
			cursor++
		}
		if cursor > index+1 && cursor < len(text) && text[cursor] == ';' {
			found = append(found, [2]int{index, cursor + 1})
			index = cursor
		}
	}
	return found
}

func jsxCurlyBracePresenceIsReferenceBodyByte(character byte) bool {
	switch {
	case character >= 'A' && character <= 'Z':
		return true
	case character >= 'a' && character <= 'z':
		return true
	case character >= '0' && character <= '9':
		return true
	case character == '#':
		return true
	}
	return false
}

// jsxCurlyBracePresenceContainsCharacterReference reproduces `containsHTMLEntity`.
func jsxCurlyBracePresenceContainsCharacterReference(text string) bool {
	return len(jsxCurlyBracePresenceFindCharacterReferences(text)) > 0
}

// jsxCurlyBracePresenceContainsOnlyCharacterReferences reproduces `containsOnlyHtmlEntities`.
//
// Upstream removes every reference and trims what is left, so a literal that is references and
// whitespace is treated as references only. The trim is JavaScript's `String.prototype.trim`, whose
// character set is the same one this file's whitespace predicate uses.
func jsxCurlyBracePresenceContainsOnlyCharacterReferences(text string) bool {
	references := jsxCurlyBracePresenceFindCharacterReferences(text)
	if len(references) == 0 {
		return strings.TrimFunc(text, jsxCurlyBracePresenceIsWhitespaceRune) == "" && text != ""
	}
	var remaining strings.Builder
	previousEnd := 0
	for _, reference := range references {
		remaining.WriteString(text[previousEnd:reference[0]])
		previousEnd = reference[1]
	}
	remaining.WriteString(text[previousEnd:])
	return strings.TrimFunc(remaining.String(), jsxCurlyBracePresenceIsWhitespaceRune) == ""
}

// jsxCurlyBracePresenceContainsLineTerminator reproduces `containsLineTerminators`.
//
// The set is JavaScript's four line terminators, which is narrower than its whitespace class: a tab
// is whitespace and is not a terminator.
func jsxCurlyBracePresenceContainsLineTerminator(text string) bool {
	return strings.ContainsAny(text, "\n\r\u2028\u2029")
}

// jsxCurlyBracePresenceIsLineBreak reproduces `isLineBreak`: a terminator and nothing else.
func jsxCurlyBracePresenceIsLineBreak(text string) bool {
	return jsxCurlyBracePresenceContainsLineTerminator(text) &&
		strings.TrimFunc(text, jsxCurlyBracePresenceIsWhitespaceRune) == ""
}

// jsxCurlyBracePresenceIsAllWhitespace reproduces `jsxUtil.isWhiteSpaces`, which is `/^\s*$/`.
//
// Note that an EMPTY string satisfies this, matching the regular expression, and callers that need
// to exclude it do so themselves the way upstream's `node.value &&` guard does.
func jsxCurlyBracePresenceIsAllWhitespace(text string) bool {
	return strings.TrimFunc(text, jsxCurlyBracePresenceIsWhitespaceRune) == ""
}

// jsxCurlyBracePresenceHasEdgeWhitespace reproduces `isStringWithTrailingWhiteSpaces`, `/^\s|\s$/`.
func jsxCurlyBracePresenceHasEdgeWhitespace(text string) bool {
	if text == "" {
		return false
	}
	runes := []rune(text)
	return jsxCurlyBracePresenceIsWhitespaceRune(runes[0]) ||
		jsxCurlyBracePresenceIsWhitespaceRune(runes[len(runes)-1])
}

// jsxCurlyBracePresenceIsWhitespaceRune is the character set JavaScript's `\s` matches.
//
// This package already has `isJavaScriptWhitespace` in `self_closing_comp.go` and it is NOT what
// this rule needs: that one deliberately excludes the non-breaking space, because the rule it
// serves uses a negative lookahead that spares it. JavaScript's `\s` does match U+00A0, and so does
// the regular expression this rule reproduces. Measured against the installed build: a container
// holding only a non-breaking space is left alone exactly as one holding only a space is, so
// borrowing the neighbour's predicate would have made this rule report an input upstream declines.
func jsxCurlyBracePresenceIsWhitespaceRune(character rune) bool {
	switch character {
	case '\t', '\n', '\v', '\f', '\r', ' ',
		'\u00a0',                     // no-break space
		'\u1680',                     // ogham space mark
		'\u2000', '\u2001', '\u2002', // en quad, em quad, en space
		'\u2003', '\u2004', '\u2005', // em space, three-per-em, four-per-em
		'\u2006', '\u2007', '\u2008', // six-per-em, figure space, punctuation space
		'\u2009', '\u200a', // thin space, hair space
		'\u2028', '\u2029', // line separator, paragraph separator
		'\u202f', // narrow no-break space
		'\u205f', // medium mathematical space
		'\u3000', // ideographic space
		'\ufeff': // byte order mark
		return true
	}
	return false
}
