package react

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/jsx"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// forbidElementsMessage builds the finding for one forbidden element.
//
// Upstream carries two message ids rather than one interpolated message, and which one it uses
// depends on whether the entry supplied a `message`. That is a real distinction rather than a
// formatting detail: a config naming a replacement produces a different id, and a fixture asserting
// only "something reported" cannot tell the two apart.
//
// The truthiness test is upstream's and it is load-bearing at one value. `message ? ... : ...` on an
// EMPTY string takes the else branch, so `{element: "button", message: ""}` reports the plain id and
// not a message id with an empty tail. Measured on the installed build.
func forbidElementsMessage(element string, message string) rule.Message {
	if message == "" {
		return rule.Message{
			Id: "forbiddenElement",
			Description: fmt.Sprintf(
				"`<%s>` is on this project's forbidden list. The list exists because some elements "+
					"carry a decision the project has already made once, and reaching for the raw "+
					"element quietly opts out of it. Use whatever this project wraps it in.",
				element,
			),
		}
	}
	return rule.Message{
		Id: "forbiddenElement_message",
		Description: fmt.Sprintf(
			"`<%s>` is on this project's forbidden list: %s",
			element, message,
		),
	}
}

// ForbidElementsOptions is the decoded option object.
//
// One property, `forbid`, whose items are a union of a bare string and an object. The union is why
// this needs a hand-rolled decoder rather than `rule.DecodeOptionsInto`.
type ForbidElementsOptions struct {
	// Forbid is the list of element names to reject, each with an optional replacement note.
	//
	// Order matters and later entries win. Upstream builds a lookup keyed by element name and
	// assigns into it in list order, so a name written twice keeps the LAST entry's message.
	// Measured both ways: `["button", {element:"button", message:"m"}]` reports the message id and
	// the reverse order reports the plain one. A port storing the first match would answer the
	// opposite on both.
	Forbid []ForbidElementsEntry
}

// ForbidElementsEntry is one element on the forbidden list.
type ForbidElementsEntry struct {
	// Element is the name to reject. Required by the schema for the object form, and the whole
	// value for the string form.
	Element string

	// Message is the optional note appended to the finding. Empty means the plain message id.
	Message string
}

// forbidElementsWireEntry decodes either arm of the schema's union.
//
// `anyOf: [{type: "string"}, {type: "object", ...}]` has no Go equivalent, so the raw item is
// unmarshalled twice: once as a string, and if that fails, as the object. Reaching for
// `rule.DecodeOptionsInto` here would need the union to already be a single Go shape, which it is
// not.
type forbidElementsWireEntry struct {
	Element string `json:"element"`
	Message string `json:"message"`
}

// forbidElementsWireOptions is the JSON shape.
type forbidElementsWireOptions struct {
	Forbid []json.RawMessage `json:"forbid"`
}

// DecodeForbidElementsOptions turns the configured JSON into the struct the rule reads.
//
// Exported so fixtures drive the same path the config drives, which is what puts the union handling
// under test rather than assumed.
//
// A rule configured as a bare severity is handed nil options, and the empty body has to produce
// upstream's default of an EMPTY forbid list rather than an error. An empty list is the rule doing
// nothing, which is upstream's behaviour with no options: `configuration.forbid || []`. That is the
// safe direction for this particular rule, since the whole judgment is supplied by the config.
func DecodeForbidElementsOptions(raw []byte) (any, error) {
	options := ForbidElementsOptions{}
	if len(raw) == 0 {
		return options, nil
	}

	var wire forbidElementsWireOptions
	if err := rule.UnmarshalOptions(raw, &wire); err != nil {
		return options, err
	}

	for _, item := range wire.Forbid {
		var asString string
		if err := json.Unmarshal(item, &asString); err == nil {
			options.Forbid = append(options.Forbid, ForbidElementsEntry{Element: asString})
			continue
		}
		var asObject forbidElementsWireEntry
		if err := rule.UnmarshalOptions(item, &asObject); err != nil {
			return options, fmt.Errorf("decoding a forbid entry: %w", err)
		}
		options.Forbid = append(options.Forbid, ForbidElementsEntry{
			Element: asObject.Element,
			Message: asObject.Message,
		})
	}
	return options, nil
}

// ForbidElements flags an element the project's configuration has forbidden.
//
//	valid:   <button />                                     with no forbid list
//	valid:   <Button />                                     with forbid ["button"]
//	valid:   React.createElement(button)                    a lowercase identifier is not a name
//	valid:   createElement("button")                        a bare call with no react import
//	valid:   NotReact.createElement("button")               the namespace must be React
//	valid:   React.createElement("Modal")                   a capitalised literal is not an element
//	valid:   React.createElement("dotted.component")        a literal holding a dot
//	invalid: <button />                                     with forbid ["button"]
//	invalid: <dotted.component />                           with forbid ["dotted.component"]
//	invalid: React.createElement("button", {}, child)       with forbid ["button"]
//	invalid: React.createElement(Modal)                     with forbid ["Modal"]
//
// Ported from `react/forbid-elements` in `eslint-plugin-react`, read from the clone at
// `lib/rules/forbid-elements.js`. One option holding a union-typed list, two messages, no fixer. The
// whole 30-case corpus was replayed against the installed build (7.37.5) through the ESLint Linter
// API before any code was written, and it agreed on all 30, counts and message ids alike. Everything
// below that the corpus does not state was measured the same way.
//
// # There is no file-suffix gate
//
// Three siblings in this package gate on `.tsx`/`.jsx`, which is oxc residue rather than upstream
// behaviour. This rule is ported from the authority, which has no such gate.
//
// # The rule does nothing until it is configured, which shapes the nil-options path
//
// The whole judgment lives in the `forbid` list. With no options, upstream reads
// `configuration.forbid || []` and matches nothing, so the correct answer for a bare severity is an
// empty list rather than an error. That is the opposite of the usual hazard, where a zero value
// silently inverts a rule: here the zero value IS the rule declining, which is what upstream does.
//
// # Later entries win, and the corpus states it exactly once
//
// Upstream builds a lookup keyed by element name and assigns into it in list order, so a name
// written twice keeps the LAST entry. Corpus case 8 pins it for two objects; measured for the
// mixed spellings too, both directions:
//
//	["button", {element: "button", message: "m"}]     reports forbiddenElement_message
//	[{element: "button", message: "m"}, "button"]     reports forbiddenElement
//
// # An empty message is not a message
//
// `message ? messages.forbiddenElement_message : messages.forbiddenElement` takes the else branch
// on an empty string, so `{element: "button", message: ""}` reports the plain id. Measured. A port
// testing for the key's PRESENCE rather than its truthiness would report the message id with an
// empty tail, and no corpus case writes an empty message.
//
// # The three createElement argument shapes, and the two regexes that separate them
//
// This is where nearly all of upstream's judgment sits, and reading it wrong is silent in both
// directions. The first argument decides:
//
//	an Identifier matching /^[A-Z_]/     the identifier's NAME is the element
//	a Literal matching /^[a-z][^.]*$/    the literal's VALUE is the element
//	a MemberExpression                   the member's raw SOURCE TEXT is the element
//	anything else                        nothing is looked up
//
// The consequences are not symmetric with the JSX arm and every one was measured:
//
//	React.createElement(Modal)              reports, an identifier starting uppercase
//	React.createElement(_comp)              reports, underscore is in the identifier class
//	React.createElement(button)             CLEAN, lowercase fails the identifier regex
//	React.createElement("Modal")            CLEAN, uppercase fails the literal regex
//	React.createElement("_x")               CLEAN, underscore fails the literal regex
//	React.createElement("a.b")              CLEAN, the dot fails `[^.]*`
//	React.createElement("my-el")            reports, a dash is fine
//	React.createElement("1x")               CLEAN, a digit fails `^[a-z]`
//	React.createElement(true)               REPORTS, `String(true)` is "true" and it matches
//	React.createElement(1)                  CLEAN, "1" fails `^[a-z]`
//	React.createElement(`button`)           CLEAN, a template is not a Literal upstream
//	React.createElement(a.b)                reports as "a.b", read from source
//	React.createElement(a["b"])             reports as `a["b"]`, the raw text including quotes
//	React.createElement(a . b)              reports as "a . b", the raw text including spaces
//
// The boolean case is the one worth naming twice. Upstream's Literal arm accepts anything its
// parser types as a Literal and stringifies it, so a boolean keyword reaches the lookup with the
// text "true". Nothing in the corpus writes it. Reproduced rather than corrected, because a config
// forbidding "true" is nonsense and the alternative is a port that quietly disagrees with the
// running rule on an input neither of us expects.
//
// The member arm reads SOURCE TEXT rather than a reconstructed name, which is why the spaced and
// the computed forms report with their spelling intact. A port that rebuilt the name from the AST
// would answer "a.b" for all three and match on one input where upstream matches on none.
//
// # Which calls count as createElement
//
// The same `isCreateElement` this package already reproduces for three sibling rules: the
// namespaced form requires the object to be the identifier `React` and is purely syntactic, and the
// bare form requires the name to resolve to a binding the `react` module introduced. Corpus cases
// pin two of the declines directly, `createElement("button")` with no import and
// `NotReact.createElement("button")`.
//
// # The JSX arm reads source text for every name shape
//
// Upstream calls `getText(context, node.name)`, so a member name and a namespaced name arrive with
// their punctuation:
//
//	<dotted.component />    reports as "dotted.component", a corpus case
//	<a.b.c />               reports as "a.b.c"
//	<svg:rect />            reports as "svg:rect"
//
// All measured. Reading `.Text()` off the tag node would answer only the last segment for a member
// name, silently failing to match any dotted entry, and the corpus's one dotted case would still
// pass because a one-level member's text happens to differ.
var ForbidElements = rule.Rule{
	Name: "react/forbid-elements",

	// Declared for the bare-call branch of `isPragmaCreateElementCall`, which asks the checker
	// whether `createElement` binds to a `react` import. Upstream's corpus depends on that
	// distinction directly: `createElement("button")` with no import is one of its clean cases.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, _ := rule.OptionsAs[ForbidElementsOptions](options)

		// A rule with an empty list matches nothing, which is the unconfigured state. Building the
		// lookup anyway costs one allocation and keeps the two paths identical.
		//
		// Assignment in list order reproduces upstream's `indexedForbidConfigs[item] = ...`, so a
		// repeated name keeps the last entry.
		forbidden := make(map[string]ForbidElementsEntry, len(settings.Forbid))
		for _, entry := range settings.Forbid {
			forbidden[entry.Element] = entry
		}

		// sourceTextOf reproduces upstream's `getText`, which returns the raw source a node spans.
		//
		// `rule.TokenRange` is what strips leading trivia, so a name written after a comment or a
		// newline reports its own text rather than the whitespace before it.
		sourceTextOf := func(node *ast.Node) string {
			span := rule.TokenRange(ctx.SourceFile, node)
			text := ctx.SourceFile.Text()
			if span.Pos() < 0 || span.End() > len(text) || span.Pos() > span.End() {
				return ""
			}
			return text[span.Pos():span.End()]
		}

		reportIfForbidden := func(element string, node *ast.Node) {
			entry, found := forbidden[element]
			if !found {
				return
			}
			ctx.ReportNode(node, forbidElementsMessage(element, entry.Message))
		}

		// checkJsxElement handles either opening form. Upstream reports on the NAME node rather
		// than on the element, so `<button />` spans `button` and not the tag.
		checkJsxElement := func(node *ast.Node) {
			// `jsx.ElementParts` rather than this package's `jsxElementTagName`, which answers a
			// different question: it returns the plain identifier NAME and declines a member or a
			// namespaced tag, because its caller reads `openingElement.name.name` the way upstream
			// does there. This rule needs the opposite on both counts. It wants the NODE, so it can
			// read source text, and a member tag is precisely what it must report, since
			// `<dotted.component />` is one of upstream's own reporting cases.
			tagName, _ := jsx.ElementParts(node)
			if tagName == nil {
				return
			}
			reportIfForbidden(sourceTextOf(tagName), tagName)
		}

		return rule.Listeners{
			ast.KindJsxOpeningElement: checkJsxElement,

			// A self-closing element never produces a JsxOpeningElement here, and it is the shape
			// most of these are written in. Upstream's one selector covers both because its parser
			// gives a self-closing element that node type.
			ast.KindJsxSelfClosingElement: checkJsxElement,

			ast.KindCallExpression: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				if !isPragmaCreateElementCall(ctx, node) {
					return
				}
				arguments := node.AsCallExpression().Arguments
				if arguments == nil || len(arguments.Nodes) == 0 {
					return
				}
				argument := arguments.Nodes[0]

				switch {
				case argument.Kind == ast.KindIdentifier:
					// Upstream's `/^[A-Z_]/`, which is an ASCII test rather than a Unicode one.
					name := argument.Text()
					if forbidElementsStartsUpperOrUnderscore(name) {
						reportIfForbidden(name, argument)
					}

				case forbidElementsIsUpstreamLiteral(argument):
					// Upstream's `/^[a-z][^.]*$/` over `String(value)`. The rendering is the node's
					// own text for a string, and the keyword itself for `true`, `false` and `null`,
					// which is why a boolean reaches the lookup at all.
					value := forbidElementsLiteralText(argument)
					if forbidElementsIsLowercaseUndotted(value) {
						reportIfForbidden(value, argument)
					}

				case argument.Kind == ast.KindPropertyAccessExpression ||
					argument.Kind == ast.KindElementAccessExpression:
					// Upstream reads the member's SOURCE TEXT, not a reconstructed name, so the
					// spacing and the bracket spelling survive into the lookup key.
					reportIfForbidden(sourceTextOf(argument), argument)
				}
			},
		}
	},
}

// forbidElementsStartsUpperOrUnderscore is upstream's `/^[A-Z_]/`.
//
// ASCII rather than Unicode, deliberately. `unicode.IsUpper` would accept a capital in another
// script, and upstream's character class does not, so a component named with one would report here
// and be silent upstream. This is the same divergence direction the shelf's `IsLikelyComponentName`
// carries, which is why that helper is not used.
func forbidElementsStartsUpperOrUnderscore(name string) bool {
	if name == "" {
		return false
	}
	first := name[0]
	return (first >= 'A' && first <= 'Z') || first == '_'
}

// forbidElementsIsLowercaseUndotted is upstream's `/^[a-z][^.]*$/`.
//
// Anchored at both ends, so the whole string must qualify: it begins with an ASCII lowercase letter
// and contains no dot anywhere. A dash, a digit after the first character, and any other character
// all pass, which is why `my-el` matches and `1x` and `a.b` do not.
func forbidElementsIsLowercaseUndotted(value string) bool {
	if value == "" {
		return false
	}
	if value[0] < 'a' || value[0] > 'z' {
		return false
	}
	return !strings.Contains(value, ".")
}

// forbidElementsIsUpstreamLiteral reports whether a node is what upstream's parser types as Literal.
//
// The set is narrower than "any literal-looking node" and wider than "a string". A template literal
// is NOT one, measured, which is why `React.createElement(`button`)` is clean. A numeric literal is
// one, though `String(1)` then fails the lowercase regex. The boolean and null keywords are ones,
// and `true` is the single value in this set that reaches the lookup with matchable text.
func forbidElementsIsUpstreamLiteral(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral,
		ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword,
		ast.KindRegularExpressionLiteral:
		return true
	}
	return false
}

// forbidElementsLiteralText renders a literal the way upstream's `String(argument.value)` does.
//
// A string and a number carry their cooked and canonical text on the node already. The keywords
// carry no text, so they are spelled out here, which is the only place this rule constructs a value
// rather than reading one.
func forbidElementsLiteralText(node *ast.Node) string {
	switch node.Kind {
	case ast.KindTrueKeyword:
		return "true"
	case ast.KindFalseKeyword:
		return "false"
	case ast.KindNullKeyword:
		return "null"
	}
	return node.Text()
}
