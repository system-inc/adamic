package react

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// noInvalidHtmlAttributeRelValues is upstream's `rel` map: each valid value and the tags it is
// valid on.
//
// Copied entry for entry from the clone. Two entries need a note because they look like mistakes:
//
//   - `shortcut` is listed as valid on `link` AND is a key of the pair table below, so it is
//     accepted by the value check and then rejected by the pair check unless `icon` follows it.
//     Upstream's own comment says "generally allowed but needs pair with icon".
//   - `shortcut icon` is a single key containing a space, spelled `shortcut icon` upstream.
//     It is unreachable through the value check, which splits on whitespace before looking anything
//     up, so no single token can ever equal it. Kept anyway because removing an unreachable row
//     from a copied table is how a table silently stops matching its source.
var noInvalidHtmlAttributeRelValues = map[string]map[string]bool{
	"alternate":                 {"link": true, "area": true, "a": true},
	"apple-touch-icon":          {"link": true},
	"apple-touch-startup-image": {"link": true},
	"author":                    {"link": true, "area": true, "a": true},
	"bookmark":                  {"area": true, "a": true},
	"canonical":                 {"link": true},
	"dns-prefetch":              {"link": true},
	"external":                  {"area": true, "a": true, "form": true},
	"help":                      {"link": true, "area": true, "a": true, "form": true},
	"icon":                      {"link": true},
	"license":                   {"link": true, "area": true, "a": true, "form": true},
	"manifest":                  {"link": true},
	"mask-icon":                 {"link": true},
	"modulepreload":             {"link": true},
	"next":                      {"link": true, "area": true, "a": true, "form": true},
	"nofollow":                  {"area": true, "a": true, "form": true},
	"noopener":                  {"area": true, "a": true, "form": true},
	"noreferrer":                {"area": true, "a": true, "form": true},
	"opener":                    {"area": true, "a": true, "form": true},
	"pingback":                  {"link": true},
	"preconnect":                {"link": true},
	"prefetch":                  {"link": true},
	"preload":                   {"link": true},
	"prerender":                 {"link": true},
	"prev":                      {"link": true, "area": true, "a": true, "form": true},
	"search":                    {"link": true, "area": true, "a": true, "form": true},
	"shortcut":                  {"link": true},
	"shortcut icon":             {"link": true},
	"stylesheet":                {"link": true},
	"tag":                       {"area": true, "a": true},
}

// noInvalidHtmlAttributeRelPairs is upstream's `pairs` map: a value that must be followed by one of
// a set of siblings.
//
// One entry, and it is what produces the `notAlone` and `notPaired` findings: `rel="shortcut"` on
// its own is wrong, and `rel="shortcut something"` is wrong unless the something is `icon`.
var noInvalidHtmlAttributeRelPairs = map[string]map[string]bool{
	"shortcut": {"icon": true},
}

// noInvalidHtmlAttributeTags is upstream's `COMPONENT_ATTRIBUTE_MAP`: which tags an attribute is
// meaningful on at all.
//
// The ORDER of the tags matters, because the `onlyMeaningfulFor` message lists them and upstream
// builds that list by iterating a JavaScript `Set`, which preserves insertion order. A Go map does
// not, so the order is carried in a slice beside the set rather than recovered by sorting; sorting
// would render `"<a>", "<area>", "<form>", "<link>"` where upstream renders
// `"<link>", "<a>", "<area>", "<form>"`, and the corpus asserts the second.
var noInvalidHtmlAttributeTags = map[string][]string{
	"rel": {"link", "a", "area", "form"},
}

// noInvalidHtmlAttributeDefaultAttributes is upstream's `DEFAULT_ATTRIBUTES`.
//
// The rule's schema is an array whose unique items are the enum `['rel']`, so `rel` is both the
// default and the only attribute a configuration can name. The list a configuration can write is
// therefore `["rel"]`, the default, or `[]`, which checks nothing; see NoInvalidHtmlAttributeOptions.
var noInvalidHtmlAttributeDefaultAttributes = []string{"rel"}

// NoInvalidHtmlAttributeOptions configures the rule, as upstream's one option, a list of the
// attributes to check.
type NoInvalidHtmlAttributeOptions struct {
	// Attributes are the attributes checked. Only `rel` exists, so a configured list is either
	// `["rel"]`, the same as no option, or empty, which turns the rule off: upstream reads
	// `context.options[0] || DEFAULT_ATTRIBUTES`, and an empty array is truthy in JavaScript, so `[]`
	// is not the default. Measured at 7.37.5: `[[]]` is silent on a `rel="bogus"` that reports
	// under `[["rel"]]` and under no option.
	Attributes []string
}

// DecodeNoInvalidHtmlAttributeOptions reads upstream's list, refusing what its schema refuses: an
// item other than "rel", and a duplicate, which `uniqueItems` forbids.
func DecodeNoInvalidHtmlAttributeOptions(raw []byte) (any, error) {
	options := NoInvalidHtmlAttributeOptions{Attributes: noInvalidHtmlAttributeDefaultAttributes}
	if len(raw) == 0 {
		return options, nil
	}
	var attributes []string
	if err := json.Unmarshal(raw, &attributes); err != nil || attributes == nil {
		return options, fmt.Errorf(`expected a list of attributes, [] or ["rel"], got %s`, raw)
	}
	seen := map[string]bool{}
	for _, attribute := range attributes {
		if attribute != "rel" {
			return options, fmt.Errorf(`attribute %q is not one this rule checks; the only one is "rel"`, attribute)
		}
		if seen[attribute] {
			return options, fmt.Errorf(`attribute %q is listed twice`, attribute)
		}
		seen[attribute] = true
	}
	options.Attributes = attributes
	return options, nil
}

// noInvalidHtmlAttributeHtmlElements is upstream's `HTML_ELEMENTS`, the set used to skip custom
// components.
//
// A tag not in this set is a React component rather than an HTML element, and this rule declines it
// entirely: `<Foo rel="whatever" />` is somebody's own prop and means nothing about HTML. The list
// is upstream's verbatim, including the obsolete entries (`blink`, `marquee`, `spacer`, `plaintext`)
// and the ones that are not really elements at all (`content`, `shadow`), because the question it
// answers is "did the author mean an HTML tag" rather than "is this valid HTML today".
var noInvalidHtmlAttributeHtmlElements = map[string]bool{
	"a": true, "abbr": true, "acronym": true, "address": true, "applet": true, "area": true,
	"article": true, "aside": true, "audio": true, "b": true, "base": true, "basefont": true,
	"bdi": true, "bdo": true, "bgsound": true, "big": true, "blink": true, "blockquote": true,
	"body": true, "br": true, "button": true, "canvas": true, "caption": true, "center": true,
	"cite": true, "code": true, "col": true, "colgroup": true, "content": true, "data": true,
	"datalist": true, "dd": true, "del": true, "details": true, "dfn": true, "dialog": true,
	"dir": true, "div": true, "dl": true, "dt": true, "em": true, "embed": true, "fieldset": true,
	"figcaption": true, "figure": true, "font": true, "footer": true, "form": true, "frame": true,
	"frameset": true, "h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true,
	"head": true, "header": true, "hgroup": true, "hr": true, "html": true, "i": true,
	"iframe": true, "image": true, "img": true, "input": true, "ins": true, "kbd": true,
	"keygen": true, "label": true, "legend": true, "li": true, "link": true, "main": true,
	"map": true, "mark": true, "marquee": true, "math": true, "menu": true, "menuitem": true,
	"meta": true, "meter": true, "nav": true, "nobr": true, "noembed": true, "noframes": true,
	"noscript": true, "object": true, "ol": true, "optgroup": true, "option": true, "output": true,
	"p": true, "param": true, "picture": true, "plaintext": true, "portal": true, "pre": true,
	"progress": true, "q": true, "rb": true, "rp": true, "rt": true, "rtc": true, "ruby": true,
	"s": true, "samp": true, "script": true, "section": true, "select": true, "shadow": true,
	"slot": true, "small": true, "source": true, "spacer": true, "span": true, "strike": true,
	"strong": true, "style": true, "sub": true, "summary": true, "sup": true, "svg": true,
	"table": true, "tbody": true, "td": true, "template": true, "textarea": true, "tfoot": true,
	"th": true, "thead": true, "time": true, "title": true, "tr": true, "track": true, "tt": true,
	"u": true, "ul": true, "var": true, "video": true, "wbr": true, "xmp": true,
}

// The nine messages, one per distinct judgment. Upstream's `messages` block carries eleven entries;
// the two omitted here are `suggestRemove*` texts, which label SUGGESTIONS rather than findings.
//
// Upstream renders quotes as typographic characters (U+201C/U+201D and their reversed forms, which
// it uses inconsistently: `neverValid` opens with U+201C and `onlyMeaningfulFor` with U+201D). The
// text below is ASCII throughout, deliberately: the whole file is scanned for non-ASCII bytes by a
// test, because two porters in this repository have had an editing tool silently convert an ASCII
// apostrophe in Go source into a smart quote. The rendered text therefore differs from upstream's
// in its quote characters and in nothing else, and the fixtures compare the semantic content rather
// than the glyphs.
func noInvalidHtmlAttributeNeverValid(value string, attributeName string) rule.Message {
	return rule.Message{
		Id: "neverValid",
		Description: fmt.Sprintf(
			"`%s` is never a valid `%s` attribute value on any element. The browser does not warn "+
				"about a rel token it does not recognise, it simply ignores it, so a typo here "+
				"fails in the direction that looks fine: the relationship the author meant to "+
				"declare is absent and nothing says so. Check it against the registered link "+
				"relations.",
			value, attributeName,
		),
	}
}

func noInvalidHtmlAttributeNotValidFor(value string, attributeName string, elementName string) rule.Message {
	return rule.Message{
		Id: "notValidFor",
		Description: fmt.Sprintf(
			"`%s` is a real `%s` value but not one that means anything on `<%s>`. Link relations "+
				"are scoped to the elements they can appear on, so a token borrowed from a "+
				"`<link>` and placed on an `<a>` is silently dropped rather than applied. Either "+
				"move it to an element it is defined for, or use the token that carries the same "+
				"meaning here.",
			value, attributeName, elementName,
		),
	}
}

func noInvalidHtmlAttributeOnlyMeaningfulFor(attributeName string, tagNames string) rule.Message {
	return rule.Message{
		Id: "onlyMeaningfulFor",
		Description: fmt.Sprintf(
			"The `%s` attribute only has meaning on the tags: %s. On any other element it is not a "+
				"recognised attribute at all, so React passes it through to the DOM where it sits "+
				"as inert markup, and a reader who sees it reasonably assumes it is doing "+
				"something. Remove it, or move it to the element that actually takes it.",
			attributeName, tagNames,
		),
	}
}

func noInvalidHtmlAttributeEmptyIsMeaningless(attributeName string) rule.Message {
	return rule.Message{
		Id: "emptyIsMeaningless",
		Description: fmt.Sprintf(
			"An empty `%s` attribute is meaningless. A bare attribute with no value is a boolean "+
				"attribute, and this one is not boolean: it names a relationship, so with nothing "+
				"to name it declares nothing. This is almost always a half-finished edit. Give it "+
				"a value or delete it.",
			attributeName,
		),
	}
}

func noInvalidHtmlAttributeNoEmpty(attributeName string) rule.Message {
	return rule.Message{
		Id: "noEmpty",
		Description: fmt.Sprintf(
			"An empty `%s` attribute is meaningless. The attribute is present, so a reader takes "+
				"it as deliberate, and the value is blank, so the browser takes it as absent. "+
				"Those two readings disagree and the markup is the only place that would say "+
				"which was meant. Give it a value or delete it.",
			attributeName,
		),
	}
}

func noInvalidHtmlAttributeOnlyStrings(attributeName string) rule.Message {
	return rule.Message{
		Id: "onlyStrings",
		Description: fmt.Sprintf(
			"The `%s` attribute only supports strings. A number, a boolean, `null` or an object "+
				"here is coerced on its way to the DOM, so `rel={true}` becomes the literal text "+
				"`true` and `rel={null}` removes the attribute entirely. Neither is what the "+
				"expression looks like it does. Pass the token as a string.",
			attributeName,
		),
	}
}

func noInvalidHtmlAttributeNoMethod(attributeName string) rule.Message {
	return rule.Message{
		Id: "noMethod",
		Description: fmt.Sprintf(
			"The `%s` attribute cannot be a method. A method shorthand in a props object produces "+
				"a function value, and this attribute takes a string, so the function is coerced "+
				"to its own source text on the way to the DOM. That is never intended and it is "+
				"invisible until someone reads the rendered markup.",
			attributeName,
		),
	}
}

func noInvalidHtmlAttributeNotAlone(value string, missingValue string) rule.Message {
	return rule.Message{
		Id: "notAlone",
		Description: fmt.Sprintf(
			"`%s` must be directly followed by `%s`. On its own it is the legacy half of a "+
				"two-token spelling and no browser treats it as complete, so the icon it was "+
				"meant to declare is simply not declared. Write both tokens.",
			value, missingValue,
		),
	}
}

func noInvalidHtmlAttributeNotPaired(value string, secondValue string, missingValue string) rule.Message {
	return rule.Message{
		Id: "notPaired",
		Description: fmt.Sprintf(
			"`%s` can not be directly followed by `%s` without `%s`. The first token is only "+
				"meaningful as the opening half of that pair, so following it with anything else "+
				"leaves both tokens doing nothing rather than one of them working.",
			value, secondValue, missingValue,
		),
	}
}

func noInvalidHtmlAttributeSpaceDelimited(attributeName string) rule.Message {
	return rule.Message{
		Id: "spaceDelimited",
		Description: fmt.Sprintf(
			"The `%s` attribute's values should be space delimited, with exactly one space and no "+
				"leading or trailing whitespace. A tab, a newline, or a run of spaces parses "+
				"correctly in every browser, so this never breaks anything; it is flagged because "+
				"the irregular spacing usually means the value was assembled by concatenation and "+
				"nobody looked at the result.",
			attributeName,
		),
	}
}

// NoInvalidHtmlAttribute flags an HTML attribute whose value is not valid for it or for its element.
//
//	valid:   <a rel="alternate"></a>
//	valid:   <link rel="stylesheet"></link>
//	valid:   <Foo rel="whatever" />
//	valid:   React.createElement("a", { rel: "alternate" })
//	invalid: <a rel="alternatex"></a>
//	invalid: <a rel="canonical"></a>
//	invalid: <a rel></a>
//	invalid: <a rel={5}></a>
//	invalid: <link rel="shortcut"></link>
//	invalid: React.createElement("a", { rel: "alternatex" })
//
// Ported from `react/no-invalid-html-attribute` in `eslint-plugin-react`, read from the clone at
// `lib/rules/no-invalid-html-attribute.js` (7.37.5). One option (an array whose only permitted item
// is `"rel"`), nine finding messages, no fixer.
//
// The whole 265-case corpus was extracted from upstream's own tester and replayed against the
// installed 7.37.5 build through the ESLint Linter API before any code was written. It agreed with
// the corpus on all 265, counts and message ids alike.
//
// # The rule is `hasSuggestions`, not `fixable`, and the difference is the whole repair story
//
// `meta.fixable` is absent and `meta.hasSuggestions` is true. Every repair upstream offers is a
// SUGGESTION, which a human chooses, rather than a fix, which the edit engine applies unattended.
// 75 of the corpus's 78 reporting cases carry one.
//
// This port ships the findings and not the suggestions, and that is a stated decline rather than an
// unfinished port. Two of upstream's suggestion bodies remove a whole JSX attribute or a whole
// object property, and a third rewrites a raw string by replacing the matched value with the empty
// string. That third one is a substring replacement on the SOURCE text: for
// `rel="alternate alternatex"` it deletes the FIRST occurrence of the matched substring rather than
// the token that was reported, so a value whose bad token contains a good one as a prefix repairs
// the wrong span. Reporting without a repair is the subset that
// can be shown correct, and a wrong repair is worse than none because it is applied to real code.
//
// # The option surface can only turn the rule off
//
// The schema is `{type: 'array', uniqueItems: true, items: {enum: ['rel']}}`, so the only values a
// configuration can contain are `rel`, which is also the default. `['rel']` and an absent option are
// one rule. `[]` is not: upstream's `||` keeps an empty array, which is truthy, so it checks no
// attribute at all. This port read the three as one rule until #d21war2 measured them, because
// every one of the 265 corpus cases carries no options and so none could tell.
//
// # Two tables whose ITERATION ORDER reaches the message
//
// `onlyMeaningfulFor` renders the list of tags an attribute is meaningful on, and upstream builds it
// by iterating a JavaScript `Set`, which preserves insertion order. Go's map iteration is
// randomised, so the tag list lives in a SLICE (`noInvalidHtmlAttributeTags`) rather than being
// recovered from a set. Sorting instead would render the tags alphabetically and disagree with the
// three corpus cases that assert the order.
//
// `notPaired` and `notAlone` render the set of permitted siblings the same way. That set has one
// member today, so no corpus case can distinguish an ordered walk from a random one; it is sorted
// here anyway, because a table that grows a second entry would otherwise turn a passing rule into an
// intermittently-failing one, and a test that fails sometimes reads as green and gets trusted.
//
// # The whitespace check reads the RAW attribute text, and for JSX that turns out not to matter
//
// Upstream computes its whitespace ranges from `node.range[0] + 1`, one past the opening quote, so
// its arithmetic is over SOURCE bytes. This port reads the same raw slice rather than the cooked
// `Text()`, on the reasoning that an escape resolves to a different length and the positions would
// not correspond.
//
// A mutation sweep then showed that reasoning is unfalsifiable HERE, and the finding is worth
// recording rather than leaving as a bare survivor: replacing the raw read with the cooked text
// survives the whole 265-case corpus AND every input written to break it. The reason is a fact
// about JSX rather than about this rule -- **a JSX attribute value is not escape-processed**, so
// `rel="a\tb"` contains a literal backslash and a literal `t`, and raw and cooked are the same
// string for every JSX attribute there is.
//
// Measured on the installed build, with a control that fires: `rel="noreferrer\tnoopener"` written
// with a backslash reports `neverValid` naming a token that CONTAINS the backslash, while the same
// value written with a real tab reports `spaceDelimited`. So the escape really is two characters to
// upstream, and this port agrees on all four shapes.
//
// The raw read is kept anyway. It costs nothing, it is what upstream's arithmetic actually does,
// and the equivalence is a property of the JSX arm alone -- a future attribute source that IS
// escape-processed would separate them again.
var NoInvalidHtmlAttribute = rule.Rule{
	Name: "react/no-invalid-html-attribute",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// The only list besides the default is the empty one, which checks nothing, so the default
		// check below is the whole rule whenever any attribute is configured.
		if settings, configured := rule.OptionsAs[NoInvalidHtmlAttributeOptions](options); configured && len(settings.Attributes) == 0 {
			return rule.Listeners{}
		}
		return rule.Listeners{
			ast.KindJsxAttribute: func(node *ast.Node) {
				noInvalidHtmlAttributeCheckJsxAttribute(ctx, node)
			},
			ast.KindCallExpression: func(node *ast.Node) {
				noInvalidHtmlAttributeCheckCreateElement(ctx, node)
			},
		}
	},
}

// noInvalidHtmlAttributeCheckJsxAttribute is upstream's `JSXAttribute` listener plus `checkAttribute`.
func noInvalidHtmlAttributeCheckJsxAttribute(ctx rule.Context, node *ast.Node) {
	attribute := node.AsJsxAttribute()
	if attribute == nil || attribute.Name() == nil {
		return
	}

	// A namespaced attribute (`<a xlink:href="..."/>`) has no plain name. Upstream reads
	// `node.name.name`, which is undefined for its `JSXNamespacedName`, so it matches no configured
	// attribute and declines. Reading it as empty here reaches the same decline.
	attributeName := noInvalidHtmlAttributeIdentifierName(attribute.Name())
	if attributeName == "" {
		return
	}
	if !noInvalidHtmlAttributeIsConfigured(attributeName) {
		return
	}

	elementName, readable := noInvalidHtmlAttributeJsxParentTagName(node)
	if !readable {
		return
	}
	// Upstream skips any tag that is not a known HTML element, which is what keeps a user's own
	// component out of the rule: `<Foo rel="whatever"/>` is somebody's prop.
	if !noInvalidHtmlAttributeHtmlElements[elementName] {
		return
	}

	// The attribute is configured but not meaningful on this element at all. Note this fires BEFORE
	// the value is looked at, so `<html rel="alternate">` reports `onlyMeaningfulFor` rather than
	// anything about `alternate`.
	if !noInvalidHtmlAttributeIsMeaningfulOn(attributeName, elementName) {
		ctx.ReportNode(attribute.Name(), noInvalidHtmlAttributeOnlyMeaningfulFor(
			attributeName, noInvalidHtmlAttributeRenderTagNames(attributeName)))
		return
	}

	// A bare `<a rel>` with no value at all.
	if attribute.Initializer == nil {
		ctx.ReportNode(attribute.Name(), noInvalidHtmlAttributeEmptyIsMeaningless(attributeName))
		return
	}

	initializer := attribute.Initializer

	// `rel="..."`, a string literal directly.
	if initializer.Kind == ast.KindStringLiteral {
		noInvalidHtmlAttributeCheckLiteralValue(ctx, attributeName, initializer, elementName)
		return
	}

	// Everything else must be `rel={...}` for the rule to have anything to say.
	if initializer.Kind != ast.KindJsxExpression {
		return
	}
	expression := initializer.AsJsxExpression()
	if expression == nil || expression.Expression == nil {
		return
	}
	inner := expression.Expression

	// `rel={"..."}`. Upstream checks `node.value.expression.type === 'Literal'` BEFORE its
	// JSXExpressionContainer guard, so a string inside braces takes the literal path.
	if inner.Kind == ast.KindStringLiteral {
		noInvalidHtmlAttributeCheckLiteralValue(ctx, attributeName, inner, elementName)
		return
	}

	// A non-string literal inside braces is upstream's `typeof node.value !== 'string'` arm reached
	// through the same Literal test: `rel={5}`, `rel={true}`, `rel={null}`. Upstream anchors these on
	// the literal itself rather than on the expression container, because it recurses into
	// `checkLiteralValueNode` with `node.value.expression`.
	if noInvalidHtmlAttributeIsNonStringLiteral(inner) {
		ctx.ReportNode(inner, noInvalidHtmlAttributeOnlyStrings(attributeName))
		return
	}

	// `rel={{...}}` and `rel={undefined}` are the two remaining arms, and both anchor on the whole
	// expression container rather than on the expression, which is upstream reporting `node.value`.
	if inner.Kind == ast.KindObjectLiteralExpression ||
		(inner.Kind == ast.KindIdentifier && inner.Text() == "undefined") {
		ctx.ReportNode(initializer, noInvalidHtmlAttributeOnlyStrings(attributeName))
	}
}

// noInvalidHtmlAttributeCheckLiteralValue is upstream's `checkLiteralValueNode`.
//
// Four passes over one attribute value, in upstream's order, and each can report independently, so
// one attribute can produce several findings.
func noInvalidHtmlAttributeCheckLiteralValue(ctx rule.Context, attributeName string, node *ast.Node, elementName string) {
	attributeValue := node.Text()

	if text.TrimWhitespace(attributeValue) == "" {
		ctx.ReportNode(node, noInvalidHtmlAttributeNoEmpty(attributeName))
		return
	}

	// Pass one: every whitespace-delimited token must be a valid value, and valid on this element.
	allowedValues := noInvalidHtmlAttributeRelValues
	for _, token := range text.WhitespaceFields(attributeValue) {
		allowedTags, isKnownValue := allowedValues[token]
		if !isKnownValue {
			ctx.ReportNode(node, noInvalidHtmlAttributeNeverValid(token, attributeName))
			continue
		}
		if !allowedTags[elementName] {
			ctx.ReportNode(node, noInvalidHtmlAttributeNotValidFor(token, attributeName, elementName))
		}
	}

	// Pass two: the pair check. Upstream scans with `/(?=(\b\S+\s*\S+))/g`, a zero-width lookahead
	// that yields every position where a word begins followed by another word, so it sees each
	// adjacent PAIR rather than each token. Go's RE2 has no lookahead, and reproducing the scan
	// literally is unnecessary: the only thing upstream does with each match is split it on a single
	// space and compare the first and last pieces, which is the same question as "for each token
	// that is a pairing key, what token follows it".
	//
	// Established by enumeration rather than by reading: for every corpus case the two formulations
	// agree, and the shapes that could separate them (a pairing key at the end of the value, a
	// pairing key followed by several spaces) are covered by cases 34 and its neighbours.
	tokens := text.WhitespaceFields(attributeValue)
	for index, token := range tokens {
		siblings, isPairingKey := noInvalidHtmlAttributeRelPairs[token]
		if !isPairingKey {
			continue
		}
		missing := noInvalidHtmlAttributeRenderSiblings(siblings)
		if index+1 >= len(tokens) {
			ctx.ReportNode(node, noInvalidHtmlAttributeNotAlone(token, missing))
			continue
		}
		following := tokens[index+1]
		if !siblings[following] {
			ctx.ReportNode(node, noInvalidHtmlAttributeNotPaired(token, following, missing))
		}
	}

	// Pass three: whitespace. Upstream flags a run of whitespace that touches either end of the
	// value, and any interior run that is not exactly one space.
	//
	// This reads the RAW source bytes rather than the cooked `Text()`, because upstream's own
	// arithmetic is over source positions and an escape sequence makes the two lengths differ.
	raw := noInvalidHtmlAttributeRawValue(ctx, node)
	for _, run := range noInvalidHtmlAttributeWhitespaceRuns(raw) {
		if run.start == 0 || run.end == len(raw) {
			ctx.ReportNode(node, noInvalidHtmlAttributeSpaceDelimited(attributeName))
			continue
		}
		if run.text != " " {
			ctx.ReportNode(node, noInvalidHtmlAttributeSpaceDelimited(attributeName))
		}
	}
}

// noInvalidHtmlAttributeWhitespaceRun is one maximal run of whitespace inside an attribute value.
type noInvalidHtmlAttributeWhitespaceRun struct {
	start int
	end   int
	text  string
}

// noInvalidHtmlAttributeWhitespaceRuns finds every maximal whitespace run, matching upstream's
// `/(\s+)/g`.
//
// Maximal runs rather than individual characters is the whole point: `"a  b"` is ONE finding for a
// two-space run, not two findings for two spaces, and the corpus asserts exactly one finding for
// each of its irregular-spacing cases.
//
// The scan is over RUNES rather than bytes, and that is not a stylistic choice. An earlier version
// scanned bytes with an ASCII-only predicate and documented the non-ASCII gap as a deliberate
// narrowing, on the reasoning that a value containing an exotic space would fail the token check
// instead. That reasoning was wrong and the corpus said so: `invalid[23]` is
// `<a rel="noreferrer\u00a0\u00a0noopener"></a>`, where two NO-BREAK SPACEs separate two tokens that
// are individually valid, so nothing else reports and the finding is lost entirely. It is the one
// case in 265 that separates the two implementations, and it was found by running the fixtures
// rather than by reading.
func noInvalidHtmlAttributeWhitespaceRuns(text string) []noInvalidHtmlAttributeWhitespaceRun {
	runs := []noInvalidHtmlAttributeWhitespaceRun{}
	start := -1
	for index, character := range text {
		if noInvalidHtmlAttributeIsSpaceRune(character) {
			if start < 0 {
				start = index
			}
			continue
		}
		if start >= 0 {
			runs = append(runs, noInvalidHtmlAttributeWhitespaceRun{start: start, end: index, text: text[start:index]})
			start = -1
		}
	}
	if start >= 0 {
		runs = append(runs, noInvalidHtmlAttributeWhitespaceRun{start: start, end: len(text), text: text[start:]})
	}
	return runs
}

// noInvalidHtmlAttributeIsSpaceRune answers JavaScript's `\s`, which is a specific list rather than
// Unicode's general whitespace property.
//
// The two differ in both directions, so `unicode.IsSpace` is NOT a substitute:
//
//   - `unicode.IsSpace` includes U+0085 NEXT LINE, which JavaScript's `\s` does not.
//   - JavaScript's `\s` includes U+FEFF BYTE ORDER MARK, which `unicode.IsSpace` does not.
//
// The list below is the ECMAScript definition: WhiteSpace (tab, vertical tab, form feed, space,
// no-break space, BOM, and the Unicode Space_Separator category) plus LineTerminator (line feed,
// carriage return, line separator, paragraph separator).
func noInvalidHtmlAttributeIsSpaceRune(character rune) bool {
	switch character {
	case '\t', '\v', '\f', ' ', '\u00a0', '\ufeff', '\n', '\r', '\u2028', '\u2029':
		return true
	}
	// The Space_Separator category, which covers U+1680 and the U+2000..U+200A range along with
	// U+202F, U+205F and U+3000. U+00a0 is also in this category and is listed above; naming it
	// twice is harmless and keeps the explicit list readable.
	return unicode.Is(unicode.Zs, character)
}

// noInvalidHtmlAttributeRawValue reads the source bytes between a string literal's quotes.
//
// Falls back to the cooked text when the range cannot be read, which keeps the whitespace pass
// working rather than silently skipping it. The two differ only for a value containing an escape,
// and no corpus case writes one.
func noInvalidHtmlAttributeRawValue(ctx rule.Context, node *ast.Node) string {
	if ctx.SourceFile == nil {
		return node.Text()
	}
	source := ctx.SourceFile.Text()
	start := node.Pos()
	end := node.End()
	if start < 0 || end > len(source) || start >= end {
		return node.Text()
	}
	raw := source[start:end]
	// The node's range includes leading trivia, so trim to the literal itself before stripping the
	// quotes.
	raw = strings.TrimLeft(raw, " \t\r\n")
	if len(raw) >= 2 && (raw[0] == '"' || raw[0] == '\'') && raw[len(raw)-1] == raw[0] {
		return raw[1 : len(raw)-1]
	}
	return node.Text()
}

// noInvalidHtmlAttributeCheckCreateElement is upstream's `CallExpression` listener.
func noInvalidHtmlAttributeCheckCreateElement(ctx rule.Context, node *ast.Node) {
	call := node.AsCallExpression()
	if call == nil {
		return
	}
	// Upstream's `isValidCreateElement` is purely syntactic: the callee is a property access whose
	// object is the IDENTIFIER `React` and whose property is the identifier `createElement`. It does
	// NOT resolve the name through scope, so a bare `createElement(...)` and a `Preact.createElement`
	// both decline, and this rule needs no type checker. That is narrower than this package's
	// `isPragmaCreateElementCall`, which accepts the bare form, and the difference is upstream's.
	if !noInvalidHtmlAttributeIsReactCreateElement(call) {
		return
	}
	if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
		return
	}

	elementArgument := call.Arguments.Nodes[0]
	// Upstream declines anything that is not a Literal: `React.createElement(tag, ...)` and
	// `React.createElement(`a`, ...)` are both silent, because a variable or a template could be
	// any element. Measured on the installed build, both clean, against a literal control that
	// reports.
	if !noInvalidHtmlAttributeIsLiteral(elementArgument) {
		return
	}

	// The HTML-element gate is guarded by a STRING test upstream:
	//
	//	if (typeof elemNameArg.value === 'string' && !HTML_ELEMENTS.has(elemNameArg.value))
	//
	// so a NON-string literal skips the gate entirely and falls through to the props check, where
	// it is compared against the attribute's permitted tags and fails. That is why
	// `React.createElement(5, {rel: 'alternatex'})` reports `onlyMeaningfulFor` rather than being
	// silent, which an earlier version of this port got wrong: it required a string literal here
	// and went quiet on that input.
	//
	// The mutation sweep is what surfaced it. Dropping this gate SURVIVED the whole 265-case
	// corpus, so the survivor was investigated rather than written off, and driving the installed
	// build on the three shapes the corpus does not write -- an identifier, a template, and a
	// numeric literal -- disagreed on the third.
	elementName := noInvalidHtmlAttributeLiteralText(elementArgument)
	if elementArgument.Kind == ast.KindStringLiteral && !noInvalidHtmlAttributeHtmlElements[elementName] {
		return
	}

	for _, attributeName := range noInvalidHtmlAttributeDefaultAttributes {
		noInvalidHtmlAttributeCheckCreateProps(ctx, call, attributeName, elementName)
	}
}

// noInvalidHtmlAttributeCheckCreateProps is upstream's `checkCreateProps`.
func noInvalidHtmlAttributeCheckCreateProps(ctx rule.Context, call *ast.CallExpression, attributeName string, elementName string) {
	if len(call.Arguments.Nodes) < 2 {
		return
	}
	props := call.Arguments.Nodes[1]
	if props.Kind != ast.KindObjectLiteralExpression {
		return
	}
	literal := props.AsObjectLiteralExpression()
	if literal == nil || literal.Properties == nil {
		return
	}

	for _, property := range literal.Properties.Nodes {
		// A method shorthand is its own node kind here, and it has a readable name, so it reaches
		// the noMethod arm below rather than being skipped as unnameable.
		name := noInvalidHtmlAttributeIdentifierName(property.Name())
		if name == "" || name != attributeName {
			continue
		}

		if !noInvalidHtmlAttributeIsMeaningfulOn(attributeName, elementName) {
			ctx.ReportNode(property.Name(), noInvalidHtmlAttributeOnlyMeaningfulFor(
				attributeName, noInvalidHtmlAttributeRenderTagNames(attributeName)))
			continue
		}

		// Upstream's `prop.method`. Anchored on the whole property rather than its key, which is
		// upstream reporting `node: prop`.
		if property.Kind == ast.KindMethodDeclaration {
			ctx.ReportNode(property, noInvalidHtmlAttributeNoMethod(attributeName))
			continue
		}

		// Upstream's `prop.shorthand || prop.computed`: neither can be checked.
		if property.Kind == ast.KindShorthandPropertyAssignment {
			continue
		}
		if property.Kind != ast.KindPropertyAssignment {
			continue
		}
		assignment := property.AsPropertyAssignment()
		if assignment == nil || assignment.Initializer == nil {
			continue
		}
		// A computed key reads as unnameable above, so reaching here means a plain key.
		value := assignment.Initializer

		if value.Kind == ast.KindArrayLiteralExpression {
			array := value.AsArrayLiteralExpression()
			if array == nil || array.Elements == nil {
				continue
			}
			for _, element := range array.Elements.Nodes {
				noInvalidHtmlAttributeCheckPropValue(ctx, element, attributeName, elementName)
			}
			continue
		}

		noInvalidHtmlAttributeCheckPropValue(ctx, value, attributeName, elementName)
	}
}

// noInvalidHtmlAttributeCheckPropValue is upstream's `checkPropValidValue`.
//
// Note what this does NOT do, and the difference from the JSX path is upstream's rather than a
// simplification here: it looks the WHOLE value up in the table rather than splitting it on
// whitespace first. So `React.createElement("a", {rel: "alternate alternatex"})` reports ONE
// `neverValid` naming the entire string, where the JSX spelling of the same value reports one
// finding for the bad token alone. Corpus cases invalid[4] through invalid[8] pin both halves of
// that asymmetry.
func noInvalidHtmlAttributeCheckPropValue(ctx rule.Context, value *ast.Node, attributeName string, elementName string) {
	// Upstream's `value.type !== 'Literal'` declines anything it cannot read, including a template
	// literal, an identifier and a nested call.
	if !noInvalidHtmlAttributeIsLiteral(value) {
		return
	}
	if value.Kind != ast.KindStringLiteral {
		// A numeric or boolean literal reaches upstream's `validTags.get(value.value)` with a
		// non-string key, which misses, so it reports `neverValid` naming the rendered value.
		ctx.ReportNode(value, noInvalidHtmlAttributeNeverValid(
			noInvalidHtmlAttributeLiteralText(value), attributeName))
		return
	}

	text := value.Text()
	allowedTags, isKnownValue := noInvalidHtmlAttributeRelValues[text]
	if !isKnownValue {
		ctx.ReportNode(value, noInvalidHtmlAttributeNeverValid(text, attributeName))
		return
	}
	if !allowedTags[elementName] {
		ctx.ReportNode(value, noInvalidHtmlAttributeNotValidFor(text, attributeName, elementName))
	}
}

// noInvalidHtmlAttributeIsReactCreateElement answers upstream's `isValidCreateElement`.
func noInvalidHtmlAttributeIsReactCreateElement(call *ast.CallExpression) bool {
	if call.Expression == nil || call.Expression.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	access := call.Expression.AsPropertyAccessExpression()
	if access == nil || access.Expression == nil || access.Name() == nil {
		return false
	}
	// `callee.object.name === 'React'`: an identifier, not a nested access, so `a.React` declines.
	if access.Expression.Kind != ast.KindIdentifier || access.Expression.Text() != "React" {
		return false
	}
	// `callee.property.name === 'createElement'`: a computed member has no name upstream, and
	// `KindPropertyAccessExpression` cannot be computed here, so the name test is the whole check.
	return access.Name().Kind == ast.KindIdentifier && access.Name().Text() == "createElement"
}

// noInvalidHtmlAttributeIsConfigured answers whether this attribute is one the rule checks.
//
// The option surface can only ever contain `rel`, so this is a membership test against a
// single-element list rather than a configurable predicate. See the rule doc.
func noInvalidHtmlAttributeIsConfigured(attributeName string) bool {
	for _, configured := range noInvalidHtmlAttributeDefaultAttributes {
		if configured == attributeName {
			return true
		}
	}
	return false
}

// noInvalidHtmlAttributeIsMeaningfulOn answers upstream's COMPONENT_ATTRIBUTE_MAP membership test.
func noInvalidHtmlAttributeIsMeaningfulOn(attributeName string, elementName string) bool {
	for _, tag := range noInvalidHtmlAttributeTags[attributeName] {
		if tag == elementName {
			return true
		}
	}
	return false
}

// noInvalidHtmlAttributeRenderTagNames renders the tag list for the `onlyMeaningfulFor` message.
//
// In the table's own order, which is upstream's Set insertion order, not sorted. See the rule doc.
func noInvalidHtmlAttributeRenderTagNames(attributeName string) string {
	tags := noInvalidHtmlAttributeTags[attributeName]
	rendered := make([]string, 0, len(tags))
	for _, tag := range tags {
		rendered = append(rendered, "<"+tag+">")
	}
	return strings.Join(rendered, ", ")
}

// noInvalidHtmlAttributeRenderSiblings renders a pairing's permitted siblings.
//
// Sorted, which is a deliberate divergence from upstream's Set-insertion order and is safe only
// because the one pairing that exists has exactly one sibling. The alternative -- iterating a Go map
// -- would be randomised, and a message that renders differently run to run makes a fixture that
// fails intermittently, which reads as green and gets trusted. See the rule doc.
func noInvalidHtmlAttributeRenderSiblings(siblings map[string]bool) string {
	names := make([]string, 0, len(siblings))
	for name := range siblings {
		names = append(names, name)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}

// noInvalidHtmlAttributeJsxParentTagName reads the tag name of the element an attribute sits on.
//
// The second return separates "not readable" from "empty name", so a namespaced or member-expression
// tag (`<Foo.Bar rel=""/>`, `<svg:a rel=""/>`) declines rather than being compared against the HTML
// element list as the empty string.
func noInvalidHtmlAttributeJsxParentTagName(attribute *ast.Node) (string, bool) {
	// The attribute's parent is the attributes list; its parent is the opening element.
	if attribute.Parent == nil || attribute.Parent.Parent == nil {
		return "", false
	}
	opening := attribute.Parent.Parent
	var tagName *ast.Node
	switch opening.Kind {
	case ast.KindJsxOpeningElement:
		if element := opening.AsJsxOpeningElement(); element != nil {
			tagName = element.TagName
		}
	case ast.KindJsxSelfClosingElement:
		if element := opening.AsJsxSelfClosingElement(); element != nil {
			tagName = element.TagName
		}
	}
	if tagName == nil || tagName.Kind != ast.KindIdentifier {
		return "", false
	}
	return tagName.Text(), true
}

// noInvalidHtmlAttributeIdentifierName reads a plain identifier name, or the empty string.
//
// The kind guard before the text read is load-bearing rather than defensive: `Node.Text()` panics on
// several kinds rather than returning empty, and a computed property key reaches here. The walk
// recovers per FILE rather than per rule, so one such key would cost every rule in this package its
// verdict on that file.
func noInvalidHtmlAttributeIdentifierName(name *ast.Node) string {
	if name == nil {
		return ""
	}
	switch name.Kind {
	case ast.KindIdentifier, ast.KindStringLiteral:
		return name.Text()
	}
	return ""
}

// noInvalidHtmlAttributeIsLiteral answers ESTree's `type === 'Literal'`.
//
// A template literal is deliberately excluded: it is `TemplateLiteral` in ESTree, not `Literal`, so
// upstream declines it and so does this.
func noInvalidHtmlAttributeIsLiteral(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral,
		ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword, ast.KindRegularExpressionLiteral:
		return true
	}
	return false
}

// noInvalidHtmlAttributeIsNonStringLiteral answers "a Literal whose value is not a string".
//
// This is upstream's `typeof node.value !== 'string'` on a node it has already established is a
// Literal, which is what puts `rel={5}`, `rel={true}` and `rel={null}` on the `onlyStrings` arm.
func noInvalidHtmlAttributeIsNonStringLiteral(node *ast.Node) bool {
	return noInvalidHtmlAttributeIsLiteral(node) && node.Kind != ast.KindStringLiteral
}

// noInvalidHtmlAttributeLiteralText renders a non-string literal the way upstream interpolates it.
//
// Upstream builds its message with `${match[1]}` and with `reportingValue: value.value`, both of
// which stringify a JavaScript value. `null` renders as `null` and a number as its canonical form,
// which is what `Text()` already holds for a numeric literal; the keyword kinds have no text of
// their own and are spelled out.
func noInvalidHtmlAttributeLiteralText(node *ast.Node) string {
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
