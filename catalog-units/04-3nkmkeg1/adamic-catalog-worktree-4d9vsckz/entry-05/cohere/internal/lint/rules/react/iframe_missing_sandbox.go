package react

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/ecmascript/jsx"
	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// iframeMissingSandboxAllowedValues is upstream's ALLOWED_VALUES, copied in order.
//
// The empty string is a member, and it is not padding: upstream splits the attribute on a single
// space and trims each piece, so `sandbox=" allow-forms "` yields an empty first and last token,
// and both are accepted only because the empty string is in this list. Corpus case
// `<iframe sandbox=" allow-forms __unknown__ allow-popups __unknown__  "/>` reports exactly twice
// rather than four times for that reason, and it is the one case in the corpus that pins it.
//
// Two entries look redundant and are not. `allow-downloads-without-user-activation` was the
// original spelling and `allow-downloads` replaced it, and upstream carries both, so a project
// written against either browser generation stays clean.
var iframeMissingSandboxAllowedValues = map[string]bool{
	"": true,
	"allow-downloads-without-user-activation": true,
	"allow-downloads":                         true,
	"allow-forms":                             true,
	"allow-modals":                            true,
	"allow-orientation-lock":                  true,
	"allow-pointer-lock":                      true,
	"allow-popups":                            true,
	"allow-popups-to-escape-sandbox":          true,
	"allow-presentation":                      true,
	"allow-same-origin":                       true,
	"allow-scripts":                           true,
	"allow-storage-access-by-user-activation": true,
	"allow-top-navigation":                    true,
	"allow-top-navigation-by-user-activation": true,
}

var (
	messageIframeMissingSandboxAttributeMissing = rule.Message{
		Id: "attributeMissing",
		Description: "This `iframe` has no `sandbox` attribute, so the page it embeds runs with the " +
			"full privileges of a top-level document: it can run scripts against its own origin, " +
			"submit forms, open popups, and navigate the page that framed it. Third-party embeds " +
			"are the common case and the one that matters, because the framing page has no " +
			"control over what the other origin ships tomorrow. Add `sandbox` with the narrowest " +
			"set of `allow-` tokens the embed actually needs; a bare `sandbox` denies everything.",
	}
	messageIframeMissingSandboxInvalidCombination = rule.Message{
		Id: "invalidCombination",
		Description: "This `sandbox` lists both `allow-scripts` and `allow-same-origin` on a frame " +
			"whose document shares this page's origin, which undoes the sandbox rather than " +
			"narrowing it. With both set the framed document runs script as this page, so it can " +
			"reach into the parent and remove the `sandbox` attribute from its own frame, escaping " +
			"for every load after the first. Drop one of the two: keep `allow-scripts` for " +
			"untrusted content, or `allow-same-origin` for content that needs the origin but no " +
			"script.",
	}
)

// iframeMissingSandboxInvalidValue builds the per-finding message naming the token it rejected.
//
// The Id is fixed and only the Description moves, which is what keeps `ExpectFindings` able to
// count these while the rendered text still names the offending token the way upstream's
// `{{ value }}` interpolation does.
func iframeMissingSandboxInvalidValue(value string) rule.Message {
	return rule.Message{
		Id: "invalidValue",
		Description: fmt.Sprintf(
			"`%s` is not a sandbox token, so the browser ignores it and the permission the author "+
				"meant to grant is simply absent. A misspelled token fails silently in the "+
				"direction that looks safe and is not: the embed loses a capability it needs and "+
				"the author reads the missing behaviour as a bug in the embed. Check it against "+
				"the `allow-` list for the sandbox attribute.",
			value,
		),
	}
}

// IframeMissingSandbox flags an `iframe` that embeds another document without sandboxing it.
//
//	valid:   <iframe sandbox="" />
//	valid:   <iframe src="foo.htm" sandbox></iframe>
//	valid:   <iframe sandbox="allow-forms allow-modals"></iframe>
//	valid:   <iframe sandbox={""} />
//	valid:   <div sandbox="__unknown__" />
//	valid:   React.createElement("iframe", { sandbox: "" })
//	valid:   React.createElement("iframe", { src: "foo.htm", sandbox: true })
//	invalid: <iframe></iframe>
//	invalid: <iframe sandbox="__unknown__"></iframe>
//	invalid: <iframe sandbox="allow-scripts allow-same-origin"></iframe>
//	invalid: React.createElement("iframe")
//	invalid: React.createElement("iframe", {})
//
// Ported from `react/iframe-missing-sandbox` in `eslint-plugin-react`, read from the clone at
// `lib/rules/iframe-missing-sandbox.js`. No options (`schema: []`), three messages, no fixer. The
// whole 45-case corpus was replayed against the installed build (7.37.5) through the ESLint Linter
// API before any code was written, and it agreed on all 45, counts and message ids alike. Every
// behaviour below that the corpus does not state was measured the same way, on inputs written for
// the question.
//
// # There is no file-suffix gate
//
// Three siblings in this package gate on `.tsx`/`.jsx`, which is oxc residue rather than upstream
// behaviour: this rule was ported from the authority, which has no such gate anywhere. A fixture
// writes the same reporting source under `.tsx` and `.ts` and asserts both report. `.jsx` and `.js`
// are not testable here: `rule_testing`'s tsconfig pins `"include": ["**/*.ts", "**/*.tsx"]`, so a
// JavaScript-suffixed fixture is not in the program at all and the typed harness fails building the
// type graph. That is a limit of the harness rather than of the rule, and `.ts` is the extension
// that separates the gated behaviour from the ungated one anyway.
//
// # The two arms and why the createElement one is narrower than the shelf helper
//
// `internal/utilities/react.IsCreateElementCall` accepts a bare `createElement(...)` with no
// import, accepts `Preact.createElement`, and accepts a computed `React['createElement']`.
// Upstream's `isCreateElement` accepts none of those three. Measured on the installed build, all
// three are silent, and the shelf helper's own doc comment says to check before reaching for it,
// which is the check. This rule uses this package's `isPragmaCreateElementCall` instead, already
// ported against the same authority for `checked-requires-onchange-or-readonly` and shared by four
// rules. That is where the checker declaration comes from: the bare-identifier branch resolves the
// name to a binding the `react` module introduced.
//
// # `sandbox` is matched exactly, and case matters
//
// Upstream compares `attribute.name.name === 'sandbox'`. `<iframe SANDBOX="" />` reports
// `attributeMissing`, measured, so the attribute match is `jsx.MatchExactly` rather than the
// case-insensitive matcher some siblings use. The tag name is matched the same way: `<IFRAME>` is a
// component reference to JSX, not the HTML element, and is silent.
//
// # Presence and readability are two different questions, which is why the shelf value reader is
// not used here
//
// `jsx.StringAttributeValue` returns `("", false)` both for an absent attribute and for one present
// with a non-literal value, and this rule has to tell those apart: `<iframe sandbox={x} />` is
// CLEAN upstream because the attribute is present, while a missing one reports. So the attribute
// loop is written here, setting a found flag on the name alone and validating only when the value
// is a plain string literal.
//
// Three value shapes are present-but-unvalidated and all three are measured clean:
//
//	<iframe sandbox />              no initializer at all
//	<iframe sandbox={""} />         an expression container, even holding a literal
//	<iframe sandbox={`allow-forms`} />   a template literal
//
// The expression-container one is upstream's own corpus case and it is the surprising one: the
// value IS a literal, and upstream still declines it because its `attribute.value.type` is
// `JSXExpressionContainer` rather than `Literal`. Reproduced rather than improved on.
//
// # A spread does not count as supplying the attribute
//
// `<iframe {...props} />` reports `attributeMissing` and `React.createElement("iframe", {...p})`
// does too, both measured. Upstream's loops test the member kind and a spread never matches, so a
// spread that plainly could supply `sandbox` is treated as absent. This is upstream reporting on
// code it cannot see into, and it is reproduced: the alternative silences the rule on any element
// that spreads.
//
// # Which object keys count
//
// The createElement arm reads `x.key.name`, which is defined only on an identifier key, so a
// quoted key declines and the whole call reports as missing:
//
//	React.createElement("iframe", {'sandbox': ''})      reports attributeMissing
//	React.createElement("iframe", {['sandbox']: ''})    reports attributeMissing
//	React.createElement("iframe", {sandbox})           CLEAN
//
// All three measured. The shorthand is the sharp one: it satisfies the name test, so the attribute
// counts as found, and its value is an identifier rather than a literal, so nothing is validated.
// Present and unreadable, exactly like `sandbox={x}` on the JSX side.
//
// An accessor or a method key counts too, and this cost a defect on the way here. The first draft
// filtered on the member kind before the name test, accepting only a property assignment and a
// shorthand, which reads as a faithful spelling of upstream's `x.type === 'Property'`. It is not:
// in ESTree a getter, a setter and a shorthand method are ALL `Property`, differing only in `kind`,
// and every one of them carries `key.name`. Parsed with espree to confirm, and measured on the
// installed build:
//
//	React.createElement("iframe", { get sandbox() { return ""; } })   CLEAN
//	React.createElement("iframe", { set sandbox(v) {} })              CLEAN
//	React.createElement("iframe", { sandbox() { return ""; } })       CLEAN
//
// Our parser gives those three `KindGetAccessor`, `KindSetAccessor` and `KindMethodDeclaration`, so
// the kind filter excluded them and the rule reported `attributeMissing` on all three where
// upstream is silent. No imported fixture could see it, because the corpus writes no accessor
// anywhere, and the mutation that exposed it read as a subsumed guard: deleting the filter changed
// nothing the fixtures asserted. The filter is gone, the name test does the whole job, and three
// fixtures pin it.
//
// # Only the second argument is read
//
//	React.createElement("iframe", {a:1}, {sandbox: ""})    reports, measured
//
// Children are children.
//
// # Truthiness gates validation, and a non-string passes through it
//
// Upstream validates only when `value.value` is truthy AND a string. So `{ sandbox: true }` is
// clean (truthy, not a string, corpus case), and so are `{ sandbox: null }`, `{ sandbox: false }`
// and `{ sandbox: 0 }` (falsy). `{ sandbox: 5 }` is clean on both counts. All measured. Here the
// same gate is a single test that the value is a string literal, plus the empty-string case, which
// is the one shape where truthiness and stringness disagree: `sandbox=""` is a string and falsy, so
// upstream skips validation for it, and skipping validation on the empty string reaches the same
// verdict as validating it, since `""` is in the allowed list. Both routes are clean and the corpus
// asserts it twice.
//
// # The dangerous pair is silent only on a provably cross-origin frame, which is where cohere leaves upstream
//
// Upstream reports `allow-scripts allow-same-origin` on every iframe. The escape it warns about
// needs both tokens AND a framed document in the framing page's own origin, since only then can
// the frame's script reach the parent and strip its own `sandbox`. A cross-origin document gets
// its own origin back from `allow-same-origin`, which is what a YouTube player needs for its own
// storage, and it can reach nothing of ours whatever its tokens. Three real embeds in phi web,
// `HomePagePodcastSection.tsx:46`, `PodcastEpisodePlayer.tsx:71` and `YouTubeEmbed.tsx:23`, all
// `youtube-nocookie.com`, were findings upstream would raise and none of them is a defect.
//
// So the combination is silent only when the frame is PROVABLY cross-origin, and reports on
// everything else, same-origin or unknown, by @system_cohere_lint's ruling of 2026-10-04 (#3rxx2y9):
// a dynamic `src` that resolves same-origin is exactly the case the warning exists for, so silence on
// an unknown origin is the wrong default for a security rule. Provably cross-origin is a `src` whose
// text fixes an origin other than the page's: an `http:` or `https:` URL or a protocol-relative
// `//host` with its host complete, a `data:` URL (an opaque origin matches nothing), or another
// scheme no page is served from. The text is the literal's, a template's head, the initializer of a
// `const` it names, or every string and template literal constituent of its type
// (`iframeMissingSandboxFrame`). A relative `src`, an absent one (`about:blank` inherits the parent),
// an `about:`, `blob:` or `javascript:` URL, and any `srcDoc`, which takes the parent's origin and
// wins over `src`, are same-origin. A plain `string`, a template whose head stops before the host is
// complete, a bare `src` and a spread after the last `src` are unknown. Both report. The
// invalid-token and missing-attribute reports are untouched and remain exactly upstream's.
var IframeMissingSandbox = rule.Rule{
	Name: "react/iframe-missing-sandbox",

	// Declared for the bare-call branch of `isPragmaCreateElementCall`, which asks the checker
	// whether `createElement` binds to a `react` import. The namespaced `React.createElement`
	// branch is purely syntactic and needs nothing.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		// validateSandboxValue splits a sandbox attribute's text and reports each rejected token,
		// then the invalid pairing. Upstream reports every unknown token separately and the
		// combination once, in that order, which is what the two-finding corpus case pins.
		//
		// `reportsPair` is asked only once both dangerous tokens are present, because answering it can
		// cost a checker query and almost no sandbox carries the pair.
		validateSandboxValue := func(reported *ast.Node, value string, reportsPair func() bool) {
			allowScripts := false
			allowSameOrigin := false
			// Upstream splits on a single space, not on runs of whitespace, so a double space
			// yields an empty token. That token is accepted only because the empty string is in
			// the allowed list, which is why the list carries it.
			for _, token := range strings.Split(value, " ") {
				token = text.TrimWhitespace(token)
				if !iframeMissingSandboxAllowedValues[token] {
					ctx.ReportNode(reported, iframeMissingSandboxInvalidValue(token))
				}
				if token == "allow-scripts" {
					allowScripts = true
				}
				if token == "allow-same-origin" {
					allowSameOrigin = true
				}
			}
			if allowScripts && allowSameOrigin && reportsPair() {
				ctx.ReportNode(reported, messageIframeMissingSandboxInvalidCombination)
			}
		}

		// checkJsxElement runs the attribute loop for either opening form.
		//
		// `reported` is the node a finding anchors on. Upstream reports on the JSXOpeningElement,
		// which is the tag and its attributes rather than the whole element, so a paired
		// `<iframe></iframe>` points at the opening tag alone.
		checkJsxElement := func(node *ast.Node) {
			tagName, attributes := jsx.ElementParts(node)
			if !jsx.IsIntrinsicElementNamed(tagName, "iframe") {
				return
			}

			sandboxFound := false
			if attributes != nil && attributes.Kind == ast.KindJsxAttributes {
				properties := attributes.AsJsxAttributes().Properties
				if properties != nil {
					reportsPair := func() bool {
						var frame iframeMissingSandboxFrame
						for _, property := range properties.Nodes {
							attributeName, named := jsx.AttributeName(property)
							if !named {
								// Only a spread reaches here, and it may supply `src` or `srcDoc`.
								frame.opaque()
								continue
							}
							switch attributeName {
							case "src":
								frame.source(iframeMissingSandboxJsxValue(property.AsJsxAttribute().Initializer))
							case "srcDoc", "srcdoc":
								frame.sourceDocument = true
							}
						}
						return !frame.provablyCrossOrigin(ctx)
					}
					for _, property := range properties.Nodes {
						attributeName, named := jsx.AttributeName(property)
						if !named || !jsx.MatchExactly(attributeName, "sandbox") {
							continue
						}
						sandboxFound = true
						// Upstream requires the value to be a `Literal` with a truthy value, so an
						// expression container declines even when it holds a literal, and a bare
						// `sandbox` with no initializer declines too. Both are present, so neither
						// clears `sandboxFound`.
						initializer := property.AsJsxAttribute().Initializer
						if initializer != nil && initializer.Kind == ast.KindStringLiteral {
							validateSandboxValue(node, initializer.Text(), reportsPair)
						}
					}
				}
			}

			if !sandboxFound {
				ctx.ReportNode(node, messageIframeMissingSandboxAttributeMissing)
			}
		}

		return rule.Listeners{
			ast.KindJsxOpeningElement: checkJsxElement,

			// A self-closing `<iframe />` never produces a JsxOpeningElement, and it is the shape
			// most iframes are written in. Upstream's selector `JSXOpeningElement[name.name=...]`
			// matches both because its parser gives a self-closing element that node type; ours
			// does not, so both kinds are listened for.
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
				// Upstream tests `tag.type === 'Literal' && tag.value === 'iframe'`, so a template
				// literal and an identifier both decline. Measured silent for both.
				first := arguments.Nodes[0]
				if first.Kind != ast.KindStringLiteral || first.Text() != "iframe" {
					return
				}

				sandboxFound := false
				// Only the second argument is read; a third holding `sandbox` changes nothing.
				if len(arguments.Nodes) > 1 && arguments.Nodes[1].Kind == ast.KindObjectLiteralExpression {
					properties := arguments.Nodes[1].AsObjectLiteralExpression().Properties
					if properties != nil {
						reportsPair := func() bool {
							var frame iframeMissingSandboxFrame
							for _, property := range properties.Nodes {
								// A quoted key supplies `src` at runtime as surely as a bare one does,
								// so both are read here, unlike the `sandbox` test below, which keeps
								// upstream's identifier-only reading. A computed key or a spread could
								// be either, and is opaque.
								name := propertyIdentifierName(property)
								if keyNode := property.Name(); name == "" && keyNode != nil && keyNode.Kind == ast.KindStringLiteral {
									name = keyNode.Text()
								}
								switch {
								case name == "src" && property.Kind == ast.KindPropertyAssignment:
									frame.source(property.AsPropertyAssignment().Initializer)
								case name == "src" && property.Kind == ast.KindShorthandPropertyAssignment:
									// The checker types a shorthand as the binding it names.
									frame.source(property)
								case name == "srcDoc" || name == "srcdoc":
									frame.sourceDocument = true
								default:
									// A getter or method named `src` computes its value, which is as
									// unreadable as a spread, and so is any key this cannot name.
									if name == "" || name == "src" {
										frame.opaque()
									}
								}
							}
							return !frame.provablyCrossOrigin(ctx)
						}
						for _, property := range properties.Nodes {
							// Upstream requires `x.type === 'Property'` and reads `x.key.name`,
							// and the two tests together are exactly what
							// `propertyIdentifierName` answers: a spread is a SpreadElement with
							// no key, and a quoted or computed key leaves `key.name` undefined.
							// Both come back as the empty string here, so the name test alone
							// declines all three. No kind filter is written above it, and that is
							// deliberate rather than an omission; see the accessor note below.
							if propertyIdentifierName(property) != "sandbox" {
								continue
							}
							sandboxFound = true
							// Only a plain `key: value` member carries an initializer to read.
							// A shorthand, a getter, a setter and a method all count as SUPPLYING
							// the attribute and have no literal value to validate, which is why
							// presence is set above this rather than inside it.
							if property.Kind == ast.KindPropertyAssignment {
								initializer := property.AsPropertyAssignment().Initializer
								if initializer != nil && initializer.Kind == ast.KindStringLiteral {
									validateSandboxValue(node, initializer.Text(), reportsPair)
								}
							}
						}
					}
				}

				if !sandboxFound {
					ctx.ReportNode(node, messageIframeMissingSandboxAttributeMissing)
				}
			},
		}
	},
}

// iframeMissingSandboxFrame gathers, in member order, what an iframe's props say about where its
// document comes from, so the dangerous pair is silent only when the frame provably loads another
// origin.
//
// Order matters because React lets a later prop replace an earlier one: `src` written after a spread
// is the `src` that renders, and `src` written before one may not be.
type iframeMissingSandboxFrame struct {
	// sourceDocument is set by any `srcDoc`. An `srcdoc` document always takes the framing page's
	// origin, and it wins over `src` when both are present, so its presence alone settles the
	// question.
	sourceDocument bool

	// sourceSeen and sourceValue describe the last `src`. A nil value means present and unreadable,
	// as in a bare `<iframe src />` or `src={}`.
	sourceSeen  bool
	sourceValue *ast.Node

	// opaqueAfterSource is set by a spread, or any member this cannot name, that follows the last
	// `src`, or that appears when there is no `src` at all. Either way the rendered `src` is unknown.
	opaqueAfterSource bool
}

func (frame *iframeMissingSandboxFrame) source(value *ast.Node) {
	frame.sourceSeen = true
	frame.sourceValue = value
	frame.opaqueAfterSource = false
}

func (frame *iframeMissingSandboxFrame) opaque() {
	frame.opaqueAfterSource = true
}

// provablyCrossOrigin answers whether the framed document provably loads an origin other than the
// framing page's. Only a yes keeps the pair silent, so every shape this cannot read reports: a
// `srcDoc` and an absent `src` (`about:blank`) are the page's own origin, and a bare `src` or a
// spread after the last one leaves the rendered `src` unknown.
func (frame *iframeMissingSandboxFrame) provablyCrossOrigin(ctx rule.Context) bool {
	if frame.sourceDocument || frame.opaqueAfterSource || !frame.sourceSeen {
		return false
	}
	return iframeMissingSandboxValueIsCrossOrigin(ctx, frame.sourceValue, true)
}

// iframeMissingSandboxJsxValue unwraps a JSX attribute's initializer to the expression it holds, or
// nil when there is nothing to read.
func iframeMissingSandboxJsxValue(initializer *ast.Node) *ast.Node {
	if initializer == nil {
		return nil
	}
	if initializer.Kind == ast.KindJsxExpression {
		return initializer.AsJsxExpression().Expression
	}
	return initializer
}

// iframeMissingSandboxValueIsCrossOrigin decides a `src` value from its text when it has one, from a
// `const` binding's initializer when it names one, and from its type otherwise.
//
// A template expression is decided from its head alone, which is how all three real embeds in phi
// web are written: `https://www.youtube-nocookie.com/embed/${id}?rel=0` fixes its origin before the
// first substitution. PodcastEpisodePlayer passes that template through a `const`, so a const's
// initializer is read once, never a chain of them. Anything else asks the checker, and every string
// literal or template literal constituent must be cross-origin; a plain `string` proves nothing, and
// one constituent that is not provably cross-origin is enough to report, since the value can be it.
func iframeMissingSandboxValueIsCrossOrigin(ctx rule.Context, value *ast.Node, followConst bool) bool {
	if value == nil {
		return false
	}
	value = ast.SkipParentheses(value)
	switch value.Kind {
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		return iframeMissingSandboxClassifyUrl(value.Text(), true) == iframeMissingSandboxOriginCross
	case ast.KindTemplateExpression:
		return iframeMissingSandboxClassifyUrl(value.AsTemplateExpression().Head.Text(), false) == iframeMissingSandboxOriginCross
	}
	if ctx.TypeChecker == nil {
		return false
	}
	if followConst && ast.IsIdentifier(value) {
		if symbol := ctx.TypeChecker.GetSymbolAtLocation(value); symbol != nil && symbol.ValueDeclaration != nil &&
			ast.IsVariableDeclaration(symbol.ValueDeclaration) &&
			ast.GetCombinedNodeFlags(symbol.ValueDeclaration)&ast.NodeFlagsConst != 0 {
			if initializer := symbol.ValueDeclaration.Initializer(); initializer != nil &&
				iframeMissingSandboxValueIsCrossOrigin(ctx, initializer, false) {
				return true
			}
		}
	}
	valueType := ctx.TypeChecker.GetTypeAtLocation(value)
	if valueType == nil {
		return false
	}
	parts := type_checking.UnionTypeParts(valueType)
	if len(parts) == 0 {
		return false
	}
	for _, constituent := range parts {
		switch {
		case type_checking.IsTypeFlagSet(constituent, checker.TypeFlagsStringLiteral):
			text, _ := constituent.AsLiteralType().Value().(string)
			if iframeMissingSandboxClassifyUrl(text, true) != iframeMissingSandboxOriginCross {
				return false
			}
		case type_checking.IsTypeFlagSet(constituent, checker.TypeFlagsTemplateLiteral):
			if iframeMissingSandboxClassifyUrl(constituent.AsTemplateLiteralType().Texts()[0], false) != iframeMissingSandboxOriginCross {
				return false
			}
		default:
			return false
		}
	}
	return true
}

// iframeMissingSandboxOrigin is what URL text proves about the framed document's origin.
type iframeMissingSandboxOrigin int

const (
	iframeMissingSandboxOriginUnknown iframeMissingSandboxOrigin = iota
	iframeMissingSandboxOriginSame
	iframeMissingSandboxOriginCross
)

// iframeMissingSandboxClassifyUrl decides what URL text proves about the framed document's origin.
// `complete` is false for a prefix, the head of a template, whose substitution may still change what
// the text means.
//
// Same-origin: a relative reference, which resolves against the page, and the three schemes whose
// documents inherit their creator's origin, `about:` (`about:blank`, `about:srcdoc`), `blob:` and
// `javascript:`. Cross-origin: an `http:` or `https:` URL, or a protocol-relative `//host`, whose host
// is complete, read as another origin than the page's since the rule cannot know the page's own host;
// a `data:` URL, whose document gets an opaque origin that matches nothing; and every other scheme,
// since no page is served from one. Unknown: a prefix that stops before any of that is settled, such
// as `https://${host}` or a lone `/` that may become `//host`.
//
// The URL parser drops tabs and newlines anywhere, strips leading spaces and controls, reads a
// backslash as a slash, and ignores how many slashes follow a special scheme, so `\\host` and
// `/\nhost` are protocol-relative too. The text is normalised the same way first.
func iframeMissingSandboxClassifyUrl(text string, complete bool) iframeMissingSandboxOrigin {
	text = strings.TrimLeftFunc(text, func(character rune) bool { return character <= ' ' })
	text = strings.NewReplacer("\t", "", "\n", "", "\r", "", `\`, "/").Replace(strings.ToLower(text))
	if rest, isProtocolRelative := strings.CutPrefix(text, "//"); isProtocolRelative {
		return iframeMissingSandboxHostOrigin(rest, complete)
	}
	end := strings.IndexAny(text, ":/?#")
	if end == -1 {
		// A bare reference like `embed.html` is relative. A prefix with no delimiter yet is not decided:
		// `${scheme}` may still follow it.
		if complete {
			return iframeMissingSandboxOriginSame
		}
		return iframeMissingSandboxOriginUnknown
	}
	if text[end] == ':' && iframeMissingSandboxIsScheme(text[:end]) {
		switch scheme := text[:end]; scheme {
		case "about", "blob", "javascript":
			return iframeMissingSandboxOriginSame
		case "http", "https":
			return iframeMissingSandboxHostOrigin(strings.TrimLeft(text[end+1:], "/"), complete)
		}
		return iframeMissingSandboxOriginCross
	}
	if end == 0 && text == "/" && !complete {
		// A lone slash becomes protocol-relative if the substitution after it starts with another.
		return iframeMissingSandboxOriginUnknown
	}
	return iframeMissingSandboxOriginSame
}

// iframeMissingSandboxHostOrigin decides the text after an absolute URL's scheme or a protocol-relative
// `//`: cross-origin once its host is complete, which a delimiter proves in a prefix, and unknown before.
func iframeMissingSandboxHostOrigin(rest string, complete bool) iframeMissingSandboxOrigin {
	end := strings.IndexAny(rest, "/?#")
	host := rest
	if end != -1 {
		host = rest[:end]
	}
	if host == "" || (end == -1 && !complete) {
		return iframeMissingSandboxOriginUnknown
	}
	return iframeMissingSandboxOriginCross
}

// iframeMissingSandboxIsScheme reports whether text is a valid URL scheme: a letter, then letters,
// digits, `+`, `-` or `.`. Text before a colon that fails this is a relative path, not a scheme.
func iframeMissingSandboxIsScheme(text string) bool {
	if text == "" || text[0] < 'a' || text[0] > 'z' {
		return false
	}
	for _, character := range text[1:] {
		isLetter := character >= 'a' && character <= 'z'
		isDigit := character >= '0' && character <= '9'
		if !isLetter && !isDigit && character != '+' && character != '-' && character != '.' {
			return false
		}
	}
	return true
}
