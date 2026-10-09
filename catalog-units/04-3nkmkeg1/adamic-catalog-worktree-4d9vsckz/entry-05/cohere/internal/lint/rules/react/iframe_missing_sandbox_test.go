package react

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// iframeMissingSandboxFile is where the fixtures pretend to live.
//
// A .tsx extension because most cases hold JSX. The rule has no suffix gate, which the suffix cases
// below pin by writing the same reporting source to four extensions.
const iframeMissingSandboxFile = "/repository/source/IframeMissingSandbox.tsx"

// The corpus is upstream's, extracted mechanically rather than retyped.
//
// Every case in the two tables below comes from
// /tmp/lint-sources/eslint-plugin-react/tests/lib/rules/iframe-missing-sandbox.js, read by
// evaluating the two arrays in the tester with a stubbed RuleTester and emitting Go raw strings
// from the decoded values, so no escape sequence was typed on the way here. Upstream carries 33
// valid and 12 invalid cases, and the invalid ones name 13 findings between them because one case
// reports twice.
//
// All 45 were replayed against the installed build, eslint-plugin-react 7.37.5, through the ESLint
// Linter API before any Go was written. The corpus and the running rule agreed on every one.

// TestIframeMissingSandboxFires runs the twelve failing cases from upstream.
func TestIframeMissingSandboxFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"invalid 0 paired tag with no attributes", `<iframe></iframe>;`, []string{"attributeMissing"}},
		{"invalid 1 self closing with no attributes", `<iframe/>;`, []string{"attributeMissing"}},
		{"invalid 2 createElement with no props argument", `React.createElement("iframe");`, []string{"attributeMissing"}},
		{"invalid 3 createElement with an empty props object", `React.createElement("iframe", {});`, []string{"attributeMissing"}},
		{"invalid 4 createElement with a null props argument", `React.createElement("iframe", null);`, []string{"attributeMissing"}},
		{"invalid 5 an unknown token alone", `<iframe sandbox="__unknown__"></iframe>`, []string{"invalidValue"}},
		{"invalid 6 an unknown token alone through createElement", `React.createElement("iframe", { sandbox: "__unknown__" })`, []string{"invalidValue"}},
		{"invalid 7 an unknown token after a known one", `<iframe sandbox="allow-popups __unknown__"/>`, []string{"invalidValue"}},
		{"invalid 8 an unknown token before a known one", `<iframe sandbox="__unknown__ allow-popups"/>`, []string{"invalidValue"}},
		// The one case in the corpus that reports twice, and the one that pins the empty string's
		// membership in the allowed list: the leading space and the trailing double space split
		// into three empty tokens, none of which reports, while the two unknown ones do.
		{"invalid 9 two unknown tokens amid padding", `<iframe sandbox=" allow-forms __unknown__ allow-popups __unknown__  "/>`, []string{"invalidValue", "invalidValue"}},
		{"invalid 10 scripts before same origin", `<iframe sandbox="allow-scripts allow-same-origin"></iframe>;`, []string{"invalidCombination"}},
		{"invalid 11 same origin before scripts", `<iframe sandbox="allow-same-origin allow-scripts"/>;`, []string{"invalidCombination"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, IframeMissingSandbox, iframeMissingSandboxFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestIframeMissingSandboxStaysSilent runs the thirty three passing cases from upstream.
func TestIframeMissingSandboxStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"valid 0 a different element entirely", `<div sandbox="__unknown__" />;`},
		{"valid 1 an empty sandbox", `<iframe sandbox="" />;`},
		// Present but unreadable: the value IS a literal and upstream still declines it, because
		// the attribute's value node is a JSXExpressionContainer rather than a Literal. Present is
		// what matters, so nothing reports.
		{"valid 2 an empty string inside an expression container", `<iframe sandbox={""} />`},
		{"valid 3 an empty sandbox through createElement", `React.createElement("iframe", { sandbox: "" });`},
		{"valid 4 a bare sandbox with no initializer", `<iframe src="foo.htm" sandbox></iframe>`},
		// Truthy and not a string, so validation is skipped and presence alone clears it.
		{"valid 5 a boolean sandbox through createElement", `React.createElement("iframe", { src: "foo.htm", sandbox: true })`},
		{"valid 6 the attribute written twice", `<iframe src="foo.htm" sandbox sandbox></iframe>`},
		{"valid 7 allow forms", `<iframe sandbox="allow-forms"></iframe>`},
		{"valid 8 allow modals", `<iframe sandbox="allow-modals"></iframe>`},
		{"valid 9 allow orientation lock", `<iframe sandbox="allow-orientation-lock"></iframe>`},
		{"valid 10 allow pointer lock", `<iframe sandbox="allow-pointer-lock"></iframe>`},
		{"valid 11 allow popups", `<iframe sandbox="allow-popups"></iframe>`},
		{"valid 12 allow popups to escape sandbox", `<iframe sandbox="allow-popups-to-escape-sandbox"></iframe>`},
		{"valid 13 allow presentation", `<iframe sandbox="allow-presentation"></iframe>`},
		{"valid 14 allow same origin alone", `<iframe sandbox="allow-same-origin"></iframe>`},
		{"valid 15 allow scripts alone", `<iframe sandbox="allow-scripts"></iframe>`},
		{"valid 16 allow top navigation", `<iframe sandbox="allow-top-navigation"></iframe>`},
		{"valid 17 allow top navigation by user activation", `<iframe sandbox="allow-top-navigation-by-user-activation"></iframe>`},
		{"valid 18 two known tokens", `<iframe sandbox="allow-forms allow-modals"></iframe>`},
		{"valid 19 five known tokens including same origin without scripts", `<iframe sandbox="allow-popups allow-popups-to-escape-sandbox allow-pointer-lock allow-same-origin allow-top-navigation"></iframe>`},
		{"valid 20 allow forms through createElement", `React.createElement("iframe", { sandbox: "allow-forms" })`},
		{"valid 21 allow modals through createElement", `React.createElement("iframe", { sandbox: "allow-modals" })`},
		{"valid 22 allow orientation lock through createElement", `React.createElement("iframe", { sandbox: "allow-orientation-lock" })`},
		{"valid 23 allow pointer lock through createElement", `React.createElement("iframe", { sandbox: "allow-pointer-lock" })`},
		{"valid 24 allow popups through createElement", `React.createElement("iframe", { sandbox: "allow-popups" })`},
		{"valid 25 allow popups to escape sandbox through createElement", `React.createElement("iframe", { sandbox: "allow-popups-to-escape-sandbox" })`},
		{"valid 26 allow presentation through createElement", `React.createElement("iframe", { sandbox: "allow-presentation" })`},
		{"valid 27 allow same origin through createElement", `React.createElement("iframe", { sandbox: "allow-same-origin" })`},
		{"valid 28 allow scripts through createElement", `React.createElement("iframe", { sandbox: "allow-scripts" })`},
		{"valid 29 allow top navigation through createElement", `React.createElement("iframe", { sandbox: "allow-top-navigation" })`},
		{"valid 30 allow top navigation by user activation through createElement", `React.createElement("iframe", { sandbox: "allow-top-navigation-by-user-activation" })`},
		{"valid 31 two known tokens through createElement", `React.createElement("iframe", { sandbox: "allow-forms allow-modals" })`},
		{"valid 32 five known tokens through createElement", `React.createElement("iframe", { sandbox: "allow-popups allow-popups-to-escape-sandbox allow-pointer-lock allow-same-origin allow-top-navigation" })`},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, IframeMissingSandbox, iframeMissingSandboxFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestIframeMissingSandboxHasNoFileSuffixGate pins that the rule reads a plain `.ts` file.
//
// Three siblings in this package gate on `.tsx`/`.jsx`, inherited from oxc, which costs them every
// finding in a `.ts` file. Upstream has no such gate anywhere, so a later reader adding one here
// fails this rather than silently blinding the rule on most of the tree. `.ts` is the extension
// that separates the two behaviours, and it is the one the tree is mostly made of.
//
// `.jsx` and `.js` are deliberately absent, and their absence is a fact about the harness rather
// than about the rule. `rule_testing`'s tsconfig pins `"include": ["**/*.ts", "**/*.tsx"]`
// (internal/rule_testing/program.go:30), so a JavaScript-suffixed fixture is not in the program at
// all and `RunTyped` fails while building the type graph with TS18003, no inputs found. Measured by
// writing all four and reading which two failed and how. A fixture for those suffixes would be
// unreachable through the harness rather than covered, which reads as coverage and is not.
func TestIframeMissingSandboxHasNoFileSuffixGate(t *testing.T) {
	t.Parallel()

	for _, suffix := range []string{".tsx", ".ts"} {
		t.Run(suffix, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(
				t,
				IframeMissingSandbox,
				"/repository/source/SuffixProbe"+suffix,
				`React.createElement("iframe");`,
			)
			rule_testing.ExpectFindings(t, result, "attributeMissing")
		})
	}
}

// TestIframeMissingSandboxAnchorsOnTheOpeningTag asserts where each finding points.
//
// The corpus asserts message ids only, so nothing in it can see a rule that reports the right
// judgment in the wrong place. Upstream reports on the JSXOpeningElement rather than on the whole
// element, so a paired tag points at the opening tag alone and its closing tag is outside the span.
// The createElement arm reports on the whole call.
func TestIframeMissingSandboxAnchorsOnTheOpeningTag(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantTexts  []string
	}{
		{"a paired tag points at the opening tag only", `<iframe></iframe>;`, []string{`<iframe>`}},
		{"a self closing tag points at the whole tag", `<iframe/>;`, []string{`<iframe/>`}},
		// The invalid-value finding is reported on the ELEMENT rather than on the attribute, which
		// is upstream's choice: `report(context, ..., { node })` where `node` is the opening
		// element passed down from the listener, not the attribute the token came from.
		{"an invalid token points at the element, not the attribute", `<iframe sandbox="__unknown__"/>;`, []string{`<iframe sandbox="__unknown__"/>`}},
		{"the combination points at the element", `<iframe sandbox="allow-scripts allow-same-origin"/>;`, []string{`<iframe sandbox="allow-scripts allow-same-origin"/>`}},
		{"createElement points at the whole call", `React.createElement("iframe");`, []string{`React.createElement("iframe")`}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			// RunTyped writes strings.TrimSpace(source)+"\n" to disk, so a span sliced from the Go
			// literal is off by however much leading whitespace the literal carries. These carry
			// none, and the transform is applied anyway so the slice is against the bytes the
			// harness actually wrote.
			onDisk := strings.TrimSpace(testCase.sourceText) + "\n"
			result := rule_testing.RunTyped(t, IframeMissingSandbox, iframeMissingSandboxFile, testCase.sourceText)
			if len(result.Diagnostics) != len(testCase.wantTexts) {
				t.Fatalf("want %d findings, got %d", len(testCase.wantTexts), len(result.Diagnostics))
			}
			for index, want := range testCase.wantTexts {
				diagnostic := result.Diagnostics[index]
				got := onDisk[diagnostic.Range.Pos():diagnostic.Range.End()]
				if got != want {
					t.Errorf("finding %d spans %q, want %q", index, got, want)
				}
			}
		})
	}
}

// TestIframeMissingSandboxRendersTheRejectedToken asserts the interpolated message text.
//
// The Id is fixed across every invalid-value finding, so `ExpectFindings` cannot see anything the
// format string does. A mutation moving only the per-finding text reads as a surviving mutant with
// no visible cause unless the rendered text is asserted, and it is asserted against a literal typed
// here rather than against the rule's own constant, so both sides cannot move together.
func TestIframeMissingSandboxRendersTheRejectedToken(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTyped(
		t,
		IframeMissingSandbox,
		iframeMissingSandboxFile,
		`<iframe sandbox="allow-popups __unknown__"/>;`,
	)
	rule_testing.ExpectFindings(t, result, "invalidValue")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("want 1 finding, got %d", len(result.Diagnostics))
	}
	got := result.Diagnostics[0].Message.Description
	if !strings.HasPrefix(got, "`__unknown__` is not a sandbox token,") {
		t.Errorf("message does not name the rejected token: %q", got)
	}
	// The token that was ACCEPTED must not appear, which is what separates naming the offender from
	// naming the whole attribute.
	if strings.Contains(got, "allow-popups") {
		t.Errorf("message names an accepted token: %q", got)
	}
}

// TestIframeMissingSandboxSeparatesPresentFromReadable covers the distinction the shelf value
// reader cannot express.
//
// `jsx.StringAttributeValue` answers `("", false)` both for an absent attribute and for a present
// one whose value it cannot read, and this rule reports on exactly one of those two. Every case
// here was measured against the installed build.
func TestIframeMissingSandboxSeparatesPresentFromReadable(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"an expression value is present and unreadable", `<iframe sandbox={x} />;`, nil},
		{"a template literal is present and unreadable", "<iframe sandbox={`allow-forms`} />;", nil},
		// A spread plainly could supply the attribute and upstream reports anyway, because its
		// loop tests the member kind and a spread never matches. Reproduced rather than improved
		// on: the alternative silences the rule on any element that spreads.
		{"a spread does not count as supplying it", `<iframe {...props} />;`, []string{"attributeMissing"}},
		{"the attribute name is matched case sensitively", `<iframe SANDBOX="" />;`, []string{"attributeMissing"}},
		// A capitalised tag is a component reference to JSX, not the HTML element.
		{"an uppercase tag is a component, not an iframe", `<IFRAME></IFRAME>;`, nil},
		{"a member tag is a component", `<Foo.iframe />;`, nil},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, IframeMissingSandbox, iframeMissingSandboxFile, testCase.sourceText)
			if len(testCase.wantIds) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestIframeMissingSandboxSplitsOnSpacesOnly covers what counts as a token separator.
//
// Upstream calls `attribute.split(' ')` on a single space and then trims each piece, which is not
// the same as splitting on whitespace, and the corpus cannot see the difference: every case in it
// separates tokens with plain spaces. A mutant swapping the split for `strings.Fields` survived the
// whole corpus and all five of the tables above.
//
// Each case here was measured against the installed build before it was written, and the two
// directions are what makes this a real discrimination rather than a preference:
//
//   - a tab-joined pair is ONE token upstream, and it is not in the allowed list, so it reports
//     `invalidValue` with the tab inside the rendered value. A whitespace split would see two known
//     tokens and stay silent, losing the finding.
//   - `allow-scripts` tab `allow-same-origin` reports `invalidValue` rather than
//     `invalidCombination`, because neither name was ever recognised. A whitespace split would
//     report the combination, which is a different judgment on the same input.
//
// The trim is what keeps a leading tab clean, and it is why `"\tallow-forms"` is silent while
// `"allow-forms\tallow-modals"` is not: trimming acts on the ends of a token, not on its middle.
func TestIframeMissingSandboxSplitsOnSpacesOnly(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"a tab joins two names into one unknown token", "<iframe sandbox=\"allow-forms\tallow-modals\" />;", []string{"invalidValue"}},
		{"a tab between the dangerous pair hides the pairing", "<iframe sandbox=\"allow-scripts\tallow-same-origin\" />;", []string{"invalidValue"}},
		{"a leading tab is trimmed off a single token", "<iframe sandbox=\"\tallow-forms\" />;", nil},
		{"a double space yields an empty token, which is allowed", `<iframe sandbox="allow-forms  allow-modals" />;`, nil},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, IframeMissingSandbox, iframeMissingSandboxFile, testCase.sourceText)
			if len(testCase.wantIds) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}

	// The rendered text carries the tab, which is what pins that the whole joined string was the
	// token rather than either half of it.
	result := rule_testing.RunTyped(
		t,
		IframeMissingSandbox,
		iframeMissingSandboxFile,
		"<iframe sandbox=\"allow-forms\tallow-modals\" />;",
	)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("want 1 finding, got %d", len(result.Diagnostics))
	}
	if !strings.HasPrefix(result.Diagnostics[0].Message.Description, "`allow-forms\tallow-modals` is not a sandbox token,") {
		t.Errorf("message does not carry the joined token: %q", result.Diagnostics[0].Message.Description)
	}
}

// TestIframeMissingSandboxReadsOnlyIdentifierKeys covers the createElement arm's key handling.
//
// Upstream reads `x.key.name`, defined only on an identifier key, so a quoted or computed key does
// not count as supplying `sandbox` and the whole call reports as missing. A shorthand DOES count,
// and its value is an identifier rather than a literal, so it is present and unvalidated. All
// measured against the installed build; none of these shapes is in the corpus.
func TestIframeMissingSandboxReadsOnlyIdentifierKeys(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"a quoted key does not count", `React.createElement("iframe", {'sandbox': ''});`, []string{"attributeMissing"}},
		{"a computed key does not count", `React.createElement("iframe", {['sandbox']: ''});`, []string{"attributeMissing"}},
		{"a shorthand counts and is not validated", `React.createElement("iframe", {sandbox});`, nil},
		// The three shapes a kind filter would wrongly exclude. In ESTree a getter, a setter and a
		// shorthand method are all `Property` carrying `key.name`, so upstream finds the attribute
		// and stays silent on all three, measured. Our parser gives them three distinct kinds, so a
		// port that filters on kind before reading the name reports where upstream does not. It
		// did, until a surviving mutant on that filter sent a probe through the parse shapes.
		{"a getter counts as supplying it", `React.createElement("iframe", { get sandbox() { return ""; } });`, nil},
		{"a setter counts as supplying it", `React.createElement("iframe", { set sandbox(v) {} });`, nil},
		{"a method counts as supplying it", `React.createElement("iframe", { sandbox() { return ""; } });`, nil},
		{"a spread in the props object does not count", `React.createElement("iframe", {...p});`, []string{"attributeMissing"}},
		{"a falsy value is present and skipped", `React.createElement("iframe", {sandbox: null});`, nil},
		{"a numeric value is present and skipped", `React.createElement("iframe", {sandbox: 5});`, nil},
		// Only the second argument is read, so a third holding `sandbox` changes nothing.
		{"the third argument is children", `React.createElement("iframe", {a:1}, {sandbox:''});`, []string{"attributeMissing"}},
		{"a non literal tag declines", `React.createElement(Foo);`, nil},
		{"a template literal tag declines", "React.createElement(`iframe`);", nil},
		{"an uppercase tag string declines", `React.createElement("IFRAME");`, nil},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, IframeMissingSandbox, iframeMissingSandboxFile, testCase.sourceText)
			if len(testCase.wantIds) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestIframeMissingSandboxMatchesUpstreamCreateElementResolution covers which callees count.
//
// The shelf's `react.IsCreateElementCall` accepts a bare `createElement` with no import, accepts
// `Preact.createElement`, and accepts a computed `React['createElement']`. Upstream's
// `isCreateElement` accepts none of the three, and all three are measured silent on the installed
// build. This pins the narrower reading, so a later reader swapping in the shelf helper because its
// name matches fails here rather than starting to report where upstream is silent.
func TestIframeMissingSandboxMatchesUpstreamCreateElementResolution(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"a bare call with no import declines", `createElement('iframe');`, nil},
		{"the preact namespace declines", `Preact.createElement('iframe');`, nil},
		{"a computed member declines", `React['createElement']('iframe');`, nil},
		{"document createElement declines", `document.createElement('iframe');`, nil},
		// The namespaced branch is purely syntactic upstream, so a local `React` that is plainly
		// not React still reports. Measured.
		{"a local React binding still reports", "const React = {};\nReact.createElement('iframe');", []string{"attributeMissing"}},
		{"a destructured react import reports", "import {createElement} from 'react';\ncreateElement('iframe');", []string{"attributeMissing"}},
		{"a destructured require reports", "const {createElement} = require('react');\ncreateElement('iframe');", []string{"attributeMissing"}},
		{"an alias off the pragma reports", "import React from 'react';\nconst createElement = React.createElement;\ncreateElement('iframe');", []string{"attributeMissing"}},
		{"a destructured preact import declines", "import {createElement} from 'preact';\ncreateElement('iframe');", nil},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, IframeMissingSandbox, iframeMissingSandboxFile, testCase.sourceText)
			if len(testCase.wantIds) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestIframeMissingSandboxRequiresTheTypedHarness pins the checker declaration.
//
// The bare-identifier branch resolves through the checker, so the plain harness hands the rule a
// nil checker and the createElement arm guards and goes silent. A later revert of
// `NeedsTypeChecker` fails here rather than producing a vacuous green, which is the failure mode
// that does not announce itself. The JSX arm needs no checker and keeps reporting either way, which
// is what makes this a discrimination rather than a blanket silence.
func TestIframeMissingSandboxRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	untyped := rule_testing.Run(
		t,
		IframeMissingSandbox,
		iframeMissingSandboxFile,
		"import {createElement} from 'react';\ncreateElement('iframe');",
	)
	rule_testing.ExpectClean(t, untyped)

	typed := rule_testing.RunTyped(
		t,
		IframeMissingSandbox,
		iframeMissingSandboxFile,
		"import {createElement} from 'react';\ncreateElement('iframe');",
	)
	rule_testing.ExpectFindings(t, typed, "attributeMissing")

	// The JSX arm is unaffected by the checker, so it reports under the plain harness too.
	jsxUntyped := rule_testing.Run(t, IframeMissingSandbox, iframeMissingSandboxFile, `<iframe/>;`)
	rule_testing.ExpectFindings(t, jsxUntyped, "attributeMissing")
}

// TestIframeMissingSandboxIsSilentOnThePairOnlyOnACrossOriginFrame covers where cohere leaves upstream.
//
// Upstream reports `allow-scripts allow-same-origin` on every iframe. The escape needs the framed
// document to share the page's origin, so a provably cross-origin embed is a false positive, and the
// first three silent cases are the real phi web sites that showed it, written as they are in the tree
// with the pair added. An unknown origin reports, by the ruling of 2026-10-04 (#3rxx2y9): a dynamic
// `src` that resolves same-origin is the case the warning exists for. Every silent case here reports
// under upstream's reading, which is what makes the table discriminate: deleting the origin check
// fails every one of them.
func TestIframeMissingSandboxIsSilentOnThePairOnlyOnACrossOriginFrame(t *testing.T) {
	t.Parallel()

	const pair = `sandbox="allow-scripts allow-same-origin"`
	cases := []struct {
		name       string
		sourceText string
		reports    bool
	}{
		// Silent: the frame provably loads another origin.
		{"HomePagePodcastSection.tsx:46, a template whose head fixes the origin", "declare const featuredEpisodeYouTubeVideoId: string;\n<iframe src={`https://www.youtube-nocookie.com/embed/${featuredEpisodeYouTubeVideoId}?rel=0`} " + pair + " />;", false},
		{"PodcastEpisodePlayer.tsx:71, the same template through a const typed string", "declare const properties: { videoId: string };\nconst embedSource = `https://www.youtube-nocookie.com/embed/${properties.videoId}?rel=0&enablejsapi=1`;\n<iframe src={embedSource} " + pair + " />;", false},
		{"YouTubeEmbed.tsx:23, a template off a props member", "declare const properties: { videoId: string };\n<iframe src={`https://www.youtube-nocookie.com/embed/${properties.videoId}?rel=0`} " + pair + " />;", false},
		{"an https literal", `<iframe src="https://player.vimeo.com/video/1" ` + pair + ` />;`, false},
		{"an uppercase scheme", `<iframe src="HTTPS://player.vimeo.com/video/1" ` + pair + ` />;`, false},
		{"a protocol relative literal", `<iframe src="//player.vimeo.com/video/1" ` + pair + ` />;`, false},
		{"a protocol relative literal behind leading space", `<iframe src="  //player.vimeo.com/video/1" ` + pair + ` />;`, false},
		{"backslashes read as slashes", `<iframe src="\\player.vimeo.com/video/1" ` + pair + ` />;`, false},
		{"a data url has an opaque origin", `<iframe src="data:text/html,hello" ` + pair + ` />;`, false},
		{"a plain string proves nothing", "declare const url: string;\n<iframe src={url} " + pair + " />;", true},
		{"a template with an empty head proves nothing", "declare const base: string;\n<iframe src={`${base}/embed`} " + pair + " />;", true},
		{"a lone slash head may become protocol relative", "declare const path: string;\n<iframe src={`/${path}`} " + pair + " />;", true},
		{"a scheme-shaped head with no delimiter yet", "declare const rest: string;\n<iframe src={`about${rest}`} " + pair + " />;", true},
		{"a union of cross-origin literals", "declare const url: 'https://a.example/' | 'https://b.example/';\n<iframe src={url} " + pair + " />;", false},
		{"a cross-origin template literal type", "declare const url: `https://www.youtube-nocookie.com/embed/${string}`;\n<iframe src={url} " + pair + " />;", false},
		{"a spread after a relative src may replace it", "declare const props: {};\n<iframe src=\"/embed.html\" {...props} " + pair + " />;", true},
		{"a spread with no src may supply one", "declare const props: {};\n<iframe {...props} " + pair + " />;", true},
		{"a bare src has no readable value", `<iframe src ` + pair + ` />;`, true},
		{"createElement with an https src", `React.createElement("iframe", { src: "https://player.vimeo.com/video/1", sandbox: "allow-scripts allow-same-origin" });`, false},
		// The shorthand has to be read as `src`, or the frame looks src-less, which is `about:blank`
		// and reports.
		{"createElement with a shorthand typed cross-origin", "declare const src: 'https://a.example/';\nReact.createElement(\"iframe\", { src, sandbox: \"allow-scripts allow-same-origin\" });", false},
		{"createElement with a spread after src", "declare const props: {};\nReact.createElement(\"iframe\", { src: \"/embed.html\", ...props, sandbox: \"allow-scripts allow-same-origin\" });", true},

		// Reports: the frame shares the page's origin, or nothing proves it does not.
		{"a relative path", `<iframe src="/embed/player.html" ` + pair + ` />;`, true},
		{"a bare relative reference", `<iframe src="player.html" ` + pair + ` />;`, true},
		// A scheme starts with a letter, so text before this colon is a relative path, not a scheme.
		{"a colon after text that is no scheme", `<iframe src="1x:player.html" ` + pair + ` />;`, true},
		{"an underscore is no scheme character either", `<iframe src="a_b:player.html" ` + pair + ` />;`, true},
		{"a literal inside an expression container", `<iframe src={"/embed/player.html"} ` + pair + ` />;`, true},
		{"an empty src", `<iframe src="" ` + pair + ` />;`, true},
		{"about blank inherits the parent", `<iframe src="about:blank" ` + pair + ` />;`, true},
		{"about blank in capitals", `<iframe src="ABOUT:blank" ` + pair + ` />;`, true},
		{"a blob url inherits its creator", `<iframe src="blob:https://phi.health/1" ` + pair + ` />;`, true},
		{"a template whose head is a relative path", "declare const id: string;\n<iframe src={`/embed/${id}`} " + pair + " />;", true},
		{"srcDoc wins over a cross-origin src", `<iframe src="https://player.vimeo.com/video/1" srcDoc="<p>hi</p>" ` + pair + ` />;`, true},
		{"a src after a spread is the one that renders", "declare const props: {};\n<iframe {...props} src=\"/embed.html\" " + pair + " />;", true},
		{"one same-origin constituent is enough", "declare const url: '/embed.html' | 'https://a.example/';\n<iframe src={url} " + pair + " />;", true},
		{"a same-origin template literal type", "declare const url: `/embed/${string}`;\n<iframe src={url} " + pair + " />;", true},
		{"createElement with a relative src", `React.createElement("iframe", { src: "/embed.html", sandbox: "allow-scripts allow-same-origin" });`, true},
		{"createElement with a quoted src key", `React.createElement("iframe", { 'src': "/embed.html", sandbox: "allow-scripts allow-same-origin" });`, true},
		{"createElement with a shorthand typed same-origin", "declare const src: '/embed.html';\nReact.createElement(\"iframe\", { src, sandbox: \"allow-scripts allow-same-origin\" });", true},
		{"createElement with srcDoc", `React.createElement("iframe", { src: "https://a.example/", srcDoc: "<p>hi</p>", sandbox: "allow-scripts allow-same-origin" });`, true},
		{"an https template whose head stops before the host is complete", "declare const host: string;\n<iframe src={`https://${host}/embed`} " + pair + " />;", true},
		{"an https template whose head stops inside the host", "declare const domain: string;\n<iframe src={`https://cdn.${domain}/embed`} " + pair + " />;", true},
		{"a const naming a relative path", "const embedSource = '/embed/player.html';\n<iframe src={embedSource} " + pair + " />;", true},
		{"a let holding a cross-origin template, which a later write can change", "declare const id: string;\nlet embedSource = `https://www.youtube-nocookie.com/embed/${id}`;\nembedSource = embedSource + '';\n<iframe src={embedSource} " + pair + " />;", true},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, IframeMissingSandbox, iframeMissingSandboxFile, testCase.sourceText)
			if !testCase.reports {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, "invalidCombination")
		})
	}
}
