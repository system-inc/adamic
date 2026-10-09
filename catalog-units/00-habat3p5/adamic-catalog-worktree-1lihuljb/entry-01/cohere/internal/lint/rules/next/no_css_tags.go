package next

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/jsx"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoCssTags = rule.Message{
	Id: "noCssTags",
	Description: "This is a manual stylesheet <link>. Next builds and versions its own CSS, so a " +
		"hand-written link to a local stylesheet is not fingerprinted, not bundled, and will " +
		"serve a stale file from cache after a deploy. Import the stylesheet instead and let the " +
		"framework emit the tag.",
}

// NoCssTags flags a <link rel="stylesheet"> pointing at a local stylesheet.
//
//	valid:   <link rel="stylesheet" href="https://fonts.googleapis.com/css" />
//	valid:   <link {...props} />
//	valid:   <link rel="stylesheet" href={path} />
//	invalid: <link rel="stylesheet" href="/_next/static/css/styles.css" />
//
// Ported from `@next/next/no-css-tags`, read against oxc's `no_css_tags.rs`.
//
// # Why an absolute URL is exempt, and why a computed href is not reported
//
// The exemption is on the href rather than on the rel: a stylesheet fetched from another origin is
// not something the build could have fingerprinted, so the complaint does not apply to it. A
// protocol-relative or root-relative path is local and is reported.
//
// Both attributes must be plain string literals for the rule to decide. `href={path}` and a spread
// carry a value the rule would have to evaluate, and upstream declines rather than guessing, which
// is why the corpus contains a passing `<link {...props} />`.
var NoCssTags = rule.Rule{
	// No family prefix. The config writes `nextjs/no-css-tags` and matching strips the namespace on
	// a `/` boundary, so a prefixed name matches nothing and runs on no files while its tests pass.
	Name: "@next/next/no-css-tags",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		report := func(node *ast.Node) {
			tagName, attributes := jsx.ElementParts(node)
			if !jsx.IsIntrinsicElementNamed(tagName, "link") {
				return
			}

			relationship, hasRelationship := jsx.StringAttributeValue(attributes, "rel", jsx.MatchExactly)
			if !hasRelationship || relationship != "stylesheet" {
				return
			}

			reference, hasReference := jsx.StringAttributeValue(attributes, "href", jsx.MatchExactly)
			if !hasReference {
				return
			}
			if strings.HasPrefix(reference, "https://") || strings.HasPrefix(reference, "http://") {
				return
			}

			ctx.ReportNode(node, messageNoCssTags)
		}

		return rule.Listeners{
			ast.KindJsxOpeningElement:     report,
			ast.KindJsxSelfClosingElement: report,
		}
	},
}
