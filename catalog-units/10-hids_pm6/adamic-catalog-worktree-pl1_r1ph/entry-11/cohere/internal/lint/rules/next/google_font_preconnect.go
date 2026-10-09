package next

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/jsx"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageGoogleFontPreconnect = rule.Message{
	Id: "googleFontPreconnect",
	Description: "This <link> points at fonts.gstatic.com without `rel=\"preconnect\"`. The " +
		"browser only opens the connection to the font host once it has parsed the stylesheet " +
		"that references it, so the DNS lookup, TCP handshake and TLS negotiation all happen on " +
		"the critical path and the text waits on them. Add rel=\"preconnect\".",
}

// GoogleFontPreconnect flags a <link> to the Google font host without rel="preconnect".
//
//	valid:   <link rel="preconnect" href="https://fonts.gstatic.com" />
//	valid:   <link rel="canonical" href={someUrl} />
//	invalid: <link href="https://fonts.gstatic.com" />
//	invalid: <link rel="stylesheet" href="https://fonts.gstatic.com" />
//
// Ported from `@next/next/google-font-preconnect`, read against oxc's `google_font_preconnect.rs`.
//
// # Two things here differ from the neighbouring rules, and both are upstream's
//
// The attribute names are matched case-insensitively, through what oxc spells
// `has_jsx_prop_ignore_case`. `no-css-tags` and `no-sync-scripts` compare exactly. That difference
// is preserved per rule rather than unified, because unifying it would be us deciding for upstream
// which of its rules should catch `<link HREF=...>`.
//
// A missing `rel` and a `rel` that is present but is not `preconnect` are the same finding. oxc
// writes this as `is_none_or`, so `<link rel="stylesheet" href="https://fonts.gstatic.com">`
// reports exactly as a bare href does. A port that only checked for an absent `rel` would pass the
// more common of the two mistakes.
var GoogleFontPreconnect = rule.Rule{
	// No family prefix. The config writes `nextjs/google-font-preconnect` and matching strips the
	// namespace on a `/` boundary, so a prefixed name matches nothing and runs on no files.
	Name: "@next/next/google-font-preconnect",
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
			if !strings.HasPrefix(reference, "https://fonts.gstatic.com") {
				return
			}

			// A `rel` that cannot be read as a string is treated as not saying `preconnect`, which
			// is oxc's `is_none_or` and is the stricter reading: the rule reports unless it can see
			// the fix in the source.
			relationship, hasRelationship := jsx.StringAttributeValue(attributes, "rel", jsx.MatchIgnoringCase)
			if hasRelationship && relationship == "preconnect" {
				return
			}

			ctx.ReportNode(node, messageGoogleFontPreconnect)
		}

		return rule.Listeners{
			ast.KindJsxOpeningElement:     report,
			ast.KindJsxSelfClosingElement: report,
		}
	},
}
