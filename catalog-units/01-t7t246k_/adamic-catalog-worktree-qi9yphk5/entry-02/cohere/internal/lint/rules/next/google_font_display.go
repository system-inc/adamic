package next

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/jsx"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageGoogleFontDisplayMissing = rule.Message{
	Id: "googleFontDisplayMissing",
	Description: "This Google Fonts <link> has no `display` parameter, so the browser uses the " +
		"default and hides the text while the font downloads. On a slow connection that is a " +
		"blank block where the words should be. Add `&display=optional` or `&display=swap` to " +
		"the href so the text is painted in a fallback face first.",
}

var messageGoogleFontDisplayNotRecommended = rule.Message{
	Id: "googleFontDisplayNotRecommended",
	Description: "This Google Fonts <link> sets `display` to auto, block, or fallback, each of " +
		"which blocks or flashes the text while the font loads. Use `optional`, which paints a " +
		"fallback immediately and only swaps if the font arrives in time, or `swap` if the exact " +
		"face matters more than the reflow.",
}

// GoogleFontDisplay flags a Google Fonts <link> with a missing or text-blocking `display`.
//
//	valid:   <link href="https://fonts.googleapis.com/css2?family=Krona+One&display=optional" />
//	valid:   <link href="https://fonts.googleapis.com/css?family=Krona+One&display=swap" />
//	invalid: <link href="https://fonts.googleapis.com/css2?family=Krona+One" />
//	invalid: <link href="https://fonts.googleapis.com/css2?family=Krona+One&display=block" />
//
// Ported from `@next/next/google-font-display`, read against oxc's `google_font_display.rs`.
//
// # Two findings, not one, and the query parse is deliberately loose
//
// A missing `display` and a present-but-discouraged one are separate messages upstream, because
// they are separate fixes: one adds a parameter, the other changes its value. Collapsing them into
// a single message would be a smaller rule that says less at the point it fires.
//
// `auto`, `block` and `fallback` are the reported values. `optional` and `swap` pass. Any other
// value passes too, which is upstream's choice: an unrecognised value is not something this rule
// claims to know about.
//
// The query is read by splitting on `?` and `&` rather than by parsing a URL, matching oxc's
// `find_url_query_value`. Order does not matter, so `?display=fallback&family=Krona+One` is caught,
// and that case is in the upstream corpus precisely because a port that assumed `family` came first
// would miss it.
var GoogleFontDisplay = rule.Rule{
	// No family prefix. The config writes `nextjs/google-font-display` and matching strips the
	// namespace on a `/` boundary, so a prefixed name matches nothing and runs on no files.
	Name: "@next/next/google-font-display",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		report := func(node *ast.Node) {
			tagName, attributes := jsx.ElementParts(node)
			if !jsx.IsIntrinsicElementNamed(tagName, "link") {
				return
			}

			reference, hasReference := jsx.StringAttributeValue(attributes, "href", jsx.MatchIgnoringCase)
			if !hasReference {
				return
			}
			if !strings.HasPrefix(reference, "https://fonts.googleapis.com/css") {
				return
			}

			display, hasDisplay := urlQueryValue(reference, "display")
			if !hasDisplay {
				ctx.ReportNode(node, messageGoogleFontDisplayMissing)
				return
			}

			switch display {
			case "auto", "block", "fallback":
				ctx.ReportNode(node, messageGoogleFontDisplayNotRecommended)
			}
		}

		return rule.Listeners{
			ast.KindJsxOpeningElement:     report,
			ast.KindJsxSelfClosingElement: report,
		}
	},
}

// urlQueryValue reads one query parameter out of an absolute http(s) URL.
//
// Deliberately not a URL parse, matching oxc's `find_url_query_value`: it splits on `?` then `&`,
// takes the first `key=value` whose key matches exactly, and does no percent-decoding. A font href
// is written as a literal in source, so the strict reading would only differ on inputs that are not
// really Google Fonts URLs, and the loose one is what upstream's behaviour is defined by.
//
// A non-absolute URL answers nothing at all, which is upstream's first guard and matters here: it
// is what keeps a relative href from being read as a query string.
func urlQueryValue(url string, key string) (string, bool) {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return "", false
	}

	_, query, found := strings.Cut(url, "?")
	if !found {
		return "", false
	}

	for _, pair := range strings.Split(query, "&") {
		name, value, hasValue := strings.Cut(pair, "=")
		if !hasValue {
			continue
		}
		if name == key {
			return value, true
		}
	}

	return "", false
}
