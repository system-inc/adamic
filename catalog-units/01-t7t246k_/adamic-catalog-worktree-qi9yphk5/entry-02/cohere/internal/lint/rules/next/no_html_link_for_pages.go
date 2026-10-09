package next

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/jsx"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoHtmlLinkForPages = rule.Message{
	Id: "noHtmlLinkForPages",
}

// noHtmlLinkForPagesMessage is the finding for one route, naming the href as upstream names it,
// normalized: `/about` reads `/about/`.
func noHtmlLinkForPagesMessage(hrefPath string) rule.Message {
	return rule.Message{
		Id: messageNoHtmlLinkForPages.Id,
		Description: "This is a raw <a> element navigating to `" + hrefPath + "`, a route this project's " +
			"pages or app directory defines. Use `Link` from `next/link`, which navigates on the client " +
			"and prefetches the destination. A plain <a> does a full document load instead, so the " +
			"application is torn down and rebuilt and every piece of in-memory state goes with it.",
	}
}

// NoHtmlLinkForPages flags a raw <a> whose href is a route the project's pages or app directory
// defines.
//
//	valid:   <Link href="/about">About</Link>
//	valid:   <a href="https://example.com">External</a>
//	valid:   <a href="/about" target="_blank">About</a>
//	valid:   <a href="/nothing">No such page</a>
//	invalid: <a href="/about">About</a>      with pages/about.tsx
//
// Ported from @next/eslint-plugin-next 16.3.1, the rule its name and its ESLint twin answer to. The
// route model is in no_html_link_for_pages_routes.go; this file is the visitor.
//
// # It reads the disk, and that is the rule
//
// The rule looks for `pages/` and `src/pages/` (or the configured pages directories) and `app/` and
// `src/app/` under the project root, builds a route pattern for every page file it finds, and reports
// an anchor only where its href matches one. With none of those directories present it reports
// nothing, as upstream returns an empty visitor. So `<a href="about">`, `./about` and `/nothing` are
// all silent: a relative href and an href naming no page are not routes this project defines.
//
// An earlier port here was oxlint's, which decides from the href's shape alone and reports every
// href that looks internal. That is a different rule under the same name, and the ESLint twin runs
// @next's, so the two engines disagreed on every relative href and on every project with no pages
// directory (#cn8sthd). Ruled by @system_cohere_lint_sets on #d21war2: port @next's route model.
//
// The root is the project root. Upstream reads `settings.next.rootDir` and falls back to the working
// directory; cohere reads no Next settings (the loader refuses `settings.next`), so it is the root
// cohere checks, which is where every consumer starts ESLint.
//
// # Declared reads, and how it is cached
//
// The page files are read through ctx.Program.FS(), declared ReadsOtherFiles, since creating or
// deleting a page changes findings in files that did not change. The model is built once per run and
// root, not per file. Its routes are the rule's program fingerprint (noHtmlLinkForPagesFingerprint),
// so the findings cache replays a file's verdict while its bytes and the routes hold, and walks every
// file again once a page is added, removed or renamed (#s9k38p3).
//
// # Where upstream crashes, this reads on
//
// Upstream's visitor throws a TypeError on a bare `href` and on a bare `target`, because it reads
// `.value.value` of an attribute with no value, and ESLint then reports nothing for the whole file.
// Here a bare href has no route to match, so the anchor is silent, and a bare target is not
// `_blank`, so it does not exempt. Two more are in the route model: a page name that is not a valid
// regular expression, and a file named `pages`.
var NoHtmlLinkForPages = rule.Rule{
	// No family prefix. The config writes `nextjs/no-html-link-for-pages` and matching strips the
	// namespace on a `/` boundary, so a prefixed name matches nothing and runs on no files.
	Name: "@next/next/no-html-link-for-pages",
	// The root is the run's directory, and the routes come off the disk.
	ProgramReads: rule.ReadsCompilerOptions | rule.ReadsOtherFiles,
	// A file's verdict reads its own bytes and the route model, so the model is the whole fingerprint.
	ProgramFingerprint: noHtmlLinkForPagesFingerprint,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, _ := rule.OptionsAs[NoHtmlLinkForPagesOptions](options)
		// A harness that builds its Context by hand has no program and so no disk to read, which is
		// upstream's no-directory case: an empty visitor.
		if ctx.Program == nil {
			return nil
		}
		model := noHtmlLinkForPagesRouteModelFor(ctx.Program, ctx.Program.GetCurrentDirectory(), settings)
		if !model.found {
			return nil
		}

		report := func(node *ast.Node) {
			tagName, attributes := jsx.ElementParts(node)
			if !jsx.IsIntrinsicElementNamed(tagName, "a") {
				return
			}

			// Upstream's order: target, then href, then download. Each finds the first attribute
			// by that exact name, as `attributes.find` does, and reads only a plain string value,
			// so `target={"_blank"}` does not exempt and `href={"/about"}` is not read.
			if target, found := jsx.StringAttributeValue(attributes, "target", jsx.MatchExactly); found && target == "_blank" {
				return
			}
			href, found := jsx.StringAttributeValue(attributes, "href", jsx.MatchExactly)
			if !found {
				return
			}
			if jsx.HasAttributeNamed(attributes, "download", jsx.MatchExactly) {
				return
			}

			// An empty href normalizes to nothing, and an href that was only a query or a fragment
			// to the empty path, and every route begins with `/`, so neither can match.
			hrefPath, defined := normalizeURL(href)
			if !defined || hrefPath == "" {
				return
			}
			// Outgoing links are ignored.
			if strings.HasPrefix(hrefPath, "http://") || strings.HasPrefix(hrefPath, "https://") || strings.HasPrefix(hrefPath, "//") {
				return
			}

			// Upstream normalizes hrefPath a second time before each test. A normalized path ends in
			// `/` and has no query, fragment or trailing index.html left, so that second pass returns
			// it unchanged and is not repeated here.
			for _, route := range model.routes {
				if route.Test(hrefPath) {
					ctx.ReportNode(node, noHtmlLinkForPagesMessage(hrefPath))
				}
			}
		}

		return rule.Listeners{
			ast.KindJsxOpeningElement:     report,
			ast.KindJsxSelfClosingElement: report,
		}
	},
}
