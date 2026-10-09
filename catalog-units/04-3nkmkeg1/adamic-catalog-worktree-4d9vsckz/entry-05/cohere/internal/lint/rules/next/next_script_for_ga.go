package next

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/jsx"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNextScriptForGa = rule.Message{
	Id: "nextScriptForGa",
	Description: "This loads Google Analytics through a plain <script> tag, which the framework " +
		"cannot schedule: it has no loading strategy, so it competes with hydration for the main " +
		"thread, and nothing deduplicates it across client navigations. Use the `Script` component " +
		"from `next/script`, which picks a strategy and injects the tag once.",
}

// NextScriptForGa flags a plain <script> that loads Google Analytics.
//
//	valid:   import Script from 'next/script'; <Script src="https://www.googletagmanager.com/gtag/js?id=X" />
//	valid:   <script dangerouslySetInnerHTML={{}} />
//	invalid: <script src="https://www.google-analytics.com/analytics.js" />
//	invalid: <script dangerouslySetInnerHTML={{__html: `...www.googletagmanager.com/gtm.js...`}} />
//
// Ported from oxc's `next_script_for_ga.rs`, whose corpus is four pass cases and five fail cases
// carrying five diagnostics, one per input.
//
// # The rule is narrower than its name, and the corpus is what says so
//
// The name reads as "Google Analytics written inline", and a port built on that reading reports
// roughly twice what upstream does. What the rule actually matches is **two short substring lists,
// and they are not the same list**:
//
//	src            www.google-analytics.com/analytics.js   www.googletagmanager.com/gtag/js
//	inline __html  www.google-analytics.com/analytics.js   www.googletagmanager.com/gtm.js
//
// `gtag/js` is src-only and `gtm.js` is inline-only, which is not a symmetry a reader would invent.
// Measured against the release binary: `<script src=".../gtm.js?id=X" />` is silent and
// `{__html: ` + "`" + `...gtag/js...` + "`" + `}` is silent, each on the arm the other list would have caught.
//
// The consequence is visible inside the corpus rather than only at the edges. Two of the five fail
// inputs contain a **second** <script> that does not report, and one of those is an inline gtag
// bootstrap, exactly the code the rule's documentation shows as the thing to avoid. It is silent
// because a gtag snippet names neither inline substring. Both fixtures assert one finding, not two,
// and a port reading "inline Google Analytics" as the subject fails on the count.
//
// Neither list is a host or a prefix. The match is a plain case-sensitive `contains` over the whole
// value, so `http://www.google-analytics.com/analytics.js` reports (scheme is never examined) and
// `WWW.GOOGLE-ANALYTICS.COM/ANALYTICS.JS` is silent (the value is compared case-sensitively even
// though the attribute name is not). Both measured.
//
// # Case sensitivity runs opposite ways on the name and the value
//
// Upstream reaches both attributes through `has_jsx_prop_ignore_case`, so the attribute **names**
// are matched with `eq_ignore_ascii_case`, and it spells the second one lowercase as
// `"dangerouslysetinnerhtml"`, which only finds anything because of that leniency. Measured:
// `<script SRC=... />` reports and `<script DANGEROUSLYSETINNERHTML=... />` reports.
//
// The **tag** name is not lenient. Upstream compares it to the literal `"script"`, so `<SCRIPT
// src=".../analytics.js" />` is silent, measured. That asymmetry is the reason this rule uses
// `jsx.MatchIgnoringCase` where its `no-sync-scripts` and `no-unwanted-polyfillio` siblings use
// `jsx.MatchExactly`. The shelf's own comment says the matcher is a per-rule decision rather than a
// per-family one, and this rule is on the other side of that line from every script rule beside it.
//
// A namespaced attribute (`x:src`) is silent upstream, which `jsx.AttributeName` already declines.
//
// # The two branches, and the one that returns
//
// The `src` branch reports and **returns**, so an element whose src matches never also reports for
// its inline content. Measured: a script carrying both a matching src and a matching `__html`
// reports once. Every other exit from the src branch falls through, so a non-matching src, an
// expression-valued src, an empty src and a spread all still reach the inline test. All four
// measured.
//
// # The inline branch reproduces three upstream narrownesses that read as defects
//
// Each of these would look like a correctness improvement to repair, and each would be a divergence.
//
// **Only a template literal counts.** `{__html: "www.google-analytics.com/analytics.js"}` written
// with a plain string literal is silent, measured. Upstream destructures
// `Expression::TemplateLiteral` and returns otherwise, so a string, a call, and an identifier all
// decline. The corpus' fifth fail case is exactly this: its inline value is
// `this.createGoogleAnalyticsMarkup()`, and the finding it carries belongs to the *other* script on
// the line below.
//
// **Only `quasis[0]` is searched.** A URL sitting after an interpolation is invisible. Measured:
// “ `head${x}www.google-analytics.com/analytics.js` “ is **silent** while
// “ `www.google-analytics.com/analytics.js${x}tail` “ reports. This is upstream indexing the first
// chunk rather than joining them, and it is reproduced.
//
// **The raw text is searched, not the cooked value.** Measured:
// “ `www.google-analytics.com/analytics.js` “ is silent upstream, because `.raw` still holds
// the six characters of the escape. Our `Node.Text()` cooks that to a period and would report, so
// the raw bytes are recovered by slicing the source with the node's own range. There is no
// `RawText` accessor on the shim, and the one in typescript-go panics on a
// `NoSubstitutionTemplateLiteral`, so the slice is the available route rather than a preference.
//
// **Only a plain identifier key named `__html` counts**, and the **first** one wins. Measured:
// `{"__html": ...}`, `{["__html"]: ...}` and a bare `{...spread}` are all silent, and a duplicated
// `__html` takes the first value and stops rather than searching on. `ast.TryGetTextOfPropertyName`
// resolves the string and computed forms and is deliberately not used here for that reason, which
// is the same call the neighbouring `inline-script-id` makes about the same shelf function.
//
// **Parentheses are not skipped.** `dangerouslySetInnerHTML={(({__html: ...}))}` is silent upstream,
// measured, because this rule destructures `JSXExpression::ObjectExpression` directly with no
// `without_parentheses()` call. The neighbouring `inline-script-id` **does** call it and this one
// does not, so the two rules disagree about parentheses over the same attribute shape. Adding a
// `ast.SkipParentheses` here would read as tidying and would change a verdict.
//
// The finding points at the tag name, six characters of `script`, taken from the snapshot's caret.
// No fix: the repair is an import plus a component plus a strategy choice, which upstream does not
// attempt either.
var NextScriptForGa = rule.Rule{
	// No namespace prefix. The config writes `nextjs/next-script-for-ga` and the parity guard strips
	// the namespace on a `/` boundary, so a prefixed name would match no inventory entry and lint no
	// files while every fixture in this package stayed green.
	Name: "@next/next/next-script-for-ga",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		report := func(node *ast.Node) {
			tagName, attributes := jsx.ElementParts(node)
			// The tag is compared case-sensitively to the lowercase intrinsic, unlike the two
			// attribute names below. `<SCRIPT>` is silent upstream, measured.
			if !jsx.IsIntrinsicElementNamed(tagName, "script") {
				return
			}

			// The alternative async tag, which loads Google Analytics by URL.
			if source, found := jsx.StringAttributeValue(attributes, "src", jsx.MatchIgnoringCase); found {
				for _, marker := range googleAnalyticsSourceMarkers {
					if strings.Contains(source, marker) {
						ctx.ReportNode(tagName, messageNextScriptForGa)
						// Upstream returns here, so a matching src suppresses the inline test on
						// the same element rather than adding a second finding. Measured.
						return
					}
				}
			}

			// The inline bootstrap, which loads Google Analytics from a string of code.
			inlineHtml, found := firstTemplateHeadRawText(ctx, dangerouslySetInnerHtmlValue(attributes))
			if !found {
				return
			}
			for _, marker := range googleAnalyticsInlineMarkers {
				if strings.Contains(inlineHtml, marker) {
					ctx.ReportNode(tagName, messageNextScriptForGa)
					return
				}
			}
		}

		return rule.Listeners{
			ast.KindJsxOpeningElement:     report,
			ast.KindJsxSelfClosingElement: report,
		}
	},
}

// googleAnalyticsSourceMarkers is oxc's SUPPORTED_SRCS, matched as plain substrings of a `src`.
//
// `gtag/js` appears here and not in the inline list. The two lists differ by one entry each and the
// asymmetry is upstream's; see the doc comment above for the measurement.
var googleAnalyticsSourceMarkers = []string{
	"www.google-analytics.com/analytics.js",
	"www.googletagmanager.com/gtag/js",
}

// googleAnalyticsInlineMarkers is oxc's SUPPORTED_HTML_CONTENT_URLS, matched as plain substrings of
// an inline template literal's first chunk.
//
// `gtm.js` appears here and not in the source list.
var googleAnalyticsInlineMarkers = []string{
	"www.google-analytics.com/analytics.js",
	"www.googletagmanager.com/gtm.js",
}

// dangerouslySetInnerHtmlValue returns the expression assigned to a `__html` key inside the object
// literal an element's `dangerouslySetInnerHTML` attribute holds.
//
// Every step declines rather than continuing, matching upstream's chain of `let ... else` returns:
// the attribute must be a written attribute (a spread supplies nothing), its value must be an
// expression container, that expression must be an object literal written without parentheses, and
// the key must be a plain identifier spelled `__html`. The first such key wins; upstream's
// `find_map` stops there and a second `__html` is never read, measured against a duplicated key.
//
// Kept local rather than lifted to the jsx shelf. It answers a question about one attribute of one
// framework component, and the neighbouring `inline-script-id` reads the same attribute with a
// deliberately different parenthesis rule, so a shared helper would invite one of them to adopt the
// other's verdict.
func dangerouslySetInnerHtmlValue(attributes *ast.Node) *ast.Node {
	if attributes == nil || attributes.Kind != ast.KindJsxAttributes {
		return nil
	}
	properties := attributes.AsJsxAttributes().Properties
	if properties == nil {
		return nil
	}

	for _, property := range properties.Nodes {
		// oxc spells the target lowercase and finds it through the case-insensitive matcher, so the
		// leniency is load-bearing rather than incidental: an exact match against that spelling
		// would find nothing at all.
		attributeName, named := jsx.AttributeName(property)
		if !named || !jsx.MatchIgnoringCase(attributeName, "dangerouslySetInnerHTML") {
			continue
		}

		initializer := property.AsJsxAttribute().Initializer
		if initializer == nil || initializer.Kind != ast.KindJsxExpression {
			return nil
		}
		expression := initializer.AsJsxExpression().Expression
		// No SkipParentheses. Upstream destructures the object expression directly here, and a
		// parenthesized object is silent, measured. The sibling rule over this same attribute does
		// skip them, so this line is the disagreement rather than an oversight.
		if expression == nil || expression.Kind != ast.KindObjectLiteralExpression {
			return nil
		}

		objectProperties := expression.AsObjectLiteralExpression().Properties
		if objectProperties == nil {
			return nil
		}
		for _, objectProperty := range objectProperties.Nodes {
			// Only `ObjectProperty` with a `StaticIdentifier` key upstream. A spread, a method, an
			// accessor, a string key and a computed key all contribute nothing, all measured.
			// Shorthand is excluded too: upstream needs the property's *value* expression and a
			// shorthand's value is an identifier reference, which the template test below declines
			// anyway.
			if objectProperty.Kind != ast.KindPropertyAssignment {
				continue
			}
			name := objectProperty.Name()
			if name == nil || name.Kind != ast.KindIdentifier || name.Text() != "__html" {
				continue
			}
			return objectProperty.AsPropertyAssignment().Initializer
		}
		return nil
	}

	return nil
}

// firstTemplateHeadRawText returns the raw source text of a template literal's first chunk.
//
// Two shapes carry that chunk. A template with no interpolation parses to a
// `NoSubstitutionTemplateLiteral` whose whole body is the chunk; one with an interpolation parses to
// a `TemplateExpression` whose `Head` is it. Anything else, including a plain string literal and a
// call expression, declines, matching upstream's `Expression::TemplateLiteral` destructure.
//
// Raw rather than cooked, which is why this slices the file instead of calling `Text()`. oxc reads
// `quasis[0].value.raw`, so `.` stays six characters and does not match a period; our `Text()`
// returns the cooked value and would report on an input upstream is silent about, measured. The
// shim exposes no `RawText`, and typescript-go's own accessor panics on the no-substitution kind, so
// the node's range over the source text is the route. The delimiters are trimmed off the ends: a
// leading backtick always, and a trailing backtick or `${` depending on which shape it is.
func firstTemplateHeadRawText(ctx rule.Context, value *ast.Node) (string, bool) {
	if value == nil {
		return "", false
	}
	var chunk *ast.Node
	switch value.Kind {
	case ast.KindNoSubstitutionTemplateLiteral:
		chunk = value
	case ast.KindTemplateExpression:
		chunk = value.AsTemplateExpression().Head
	default:
		return "", false
	}
	if chunk == nil {
		return "", false
	}

	sourceText := ctx.SourceFile.Text()
	start, end := chunk.Pos(), chunk.End()
	if start < 0 || end > len(sourceText) || start >= end {
		return "", false
	}
	raw := sourceText[start:end]

	// The range starts before the token when leading trivia is attached, so the opening backtick is
	// found rather than assumed to be at index zero. Probed: the slice is `` `abc` `` with no space
	// before the value, `` \u0020`abc` `` with one, and `   /*c*/ `abc`` with a comment, so the
	// offset genuinely varies.
	//
	// Replacing this search with a constant zero is an EQUIVALENT mutation and survives the whole
	// suite, which is recorded here so the next reader does not chase it as a blind spot. Retaining
	// trivia only prepends text in front of the body, and every marker is a plain URL substring
	// containing no backtick, so no prefix can break a `Contains`. Shifting the offset the other way
	// instead, past the start of the body, fails fifteen lines, which is the control proving the
	// fixtures do reach this line.
	openingBacktick := strings.IndexByte(raw, '`')
	if openingBacktick < 0 {
		return "", false
	}
	raw = raw[openingBacktick+1:]

	// A no-substitution literal closes with a backtick and a head closes with `${`. Trimming the
	// last byte covers the first and trimming two covers the second.
	if value.Kind == ast.KindNoSubstitutionTemplateLiteral {
		raw = strings.TrimSuffix(raw, "`")
	} else {
		raw = strings.TrimSuffix(raw, "${")
	}
	return raw, true
}
