package next

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// The upstream cases are copied byte for byte out of oxc's tester block through the extractor's
// dumper, never retyped, and a script compares the bytes on disk against the extractor output. The
// added cases are marked and each says what it pins and how it was established. Anything labelled
// "measured" was run against the release oxlint binary with `no-sync-scripts` enabled alongside as a
// control, because a probe config missing its `plugins` key makes every rule silent and a silent
// probe is indistinguishable from a real decline.

func TestNoUnwantedPolyfillioReports(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
		want   []string
	}{
		{
			// Upstream. No features parameter at all: the security arm reports on the domain alone.
			// This case would pass under eslint, which has no security arm.
			name:   "upstream fail 1",
			source: "export class Blah {\n            render() {\n                return (\n                    <div>\n                        <h1>Hello title</h1>\n                        <script src='https://polyfill.io/v3/polyfill.min.js'></script>\n                    </div>\n                );\n            }\n        }",
			want:   []string{messageNoUnwantedPolyfillioSecurity.Id},
		},
		{
			// Upstream. The security arm fires and returns before the feature check, so Promise is
			// never named even though it is in the shipped list.
			name:   "upstream fail 2",
			source: "import Script from 'next/script';\n        export function MyApp({ Component, pageProps }) {\n            return (\n                <div>\n                    <Component {...pageProps} />\n                    <Script src='https://cdn.polyfill.io/v2/polyfill.min.js?features=Promise' />\n                </div>\n            );\n        }",
			want:   []string{messageNoUnwantedPolyfillioSecurity.Id},
		},
		{
			// Upstream. Singular, so the message reads `is already shipped`.
			name:   "upstream fail 3",
			source: "import Script from 'next/script';\n        export function MyApp({ Component, pageProps }) {\n            return (\n                <div>\n                <Component {...pageProps} />\n                <Script src='https://polyfill-fastly.io/v3/polyfill.min.js?features=Array.prototype.copyWithin' />\n                </div>\n            );\n        }",
			want:   []string{messageNoUnwantedPolyfillioDuplicate.Id},
		},
		{
			// Upstream. The %2C decode case, plural, and the order follows the URL rather than
			// being sorted.
			name:   "upstream fail 4",
			source: "export function MyApp({ Component, pageProps }) {\n            return (\n                <div>\n                    <Component {...pageProps} />\n                    <script src='https://cdnjs.cloudflare.com/polyfill/v3/polyfill.min.js?features=Promise%2CObject.fromEntries' />\n                </div>\n            );\n        }",
			want:   []string{messageNoUnwantedPolyfillioDuplicate.Id},
		},
		{
			// Measured on the release binary, with no-sync-scripts alongside as a control. The
			// .net host is never exercised upstream, so a port that typoed this one prefix would
			// pass every imported case.
			name:   "polyfill-fastly.net host",
			source: "export const A = () => <script src='https://polyfill-fastly.net/v3/p.js?features=Promise' />;",
			want:   []string{messageNoUnwantedPolyfillioDuplicate.Id},
		},
		{
			// Measured. An unknown feature beside a known one still reports, and the message names
			// only the known one.
			name:   "unknown feature beside a shipped one",
			source: "export const A = () => <script src='https://cdnjs.cloudflare.com/polyfill/v3/p.js?features=Foo,Promise' />;",
			want:   []string{messageNoUnwantedPolyfillioDuplicate.Id},
		},
		{
			// The double at sign is a literal part of the feature name and the shape a hand copy of
			// the list corrupts first.
			name:   "at sign iterator feature",
			source: "export const A = () => <script src='https://cdnjs.cloudflare.com/polyfill/v3/p.js?features=Array.prototype.@@iterator' />;",
			want:   []string{messageNoUnwantedPolyfillioDuplicate.Id},
		},
		{
			// Kept as an id assertion; the text assertion that actually pins both spellings lives
			// in TestNoUnwantedPolyfillioNamesTheFeaturesItFound, because dropping one entry leaves
			// the id unchanged.
			name:   "both epsilon spellings",
			source: "export const A = () => <script src='https://cdnjs.cloudflare.com/polyfill/v3/p.js?features=Number.Epsilon,Number.EPSILON' />;",
			want:   []string{messageNoUnwantedPolyfillioDuplicate.Id},
		},
		{
			// Measured. A named import binds the tag name too, which the shelf's default-import
			// helper would have missed.
			name:   "named next/script import",
			source: "import {Script} from 'next/script';\nexport const A = () => <Script src='https://polyfill.io/v3/polyfill.min.js' />;",
			want:   []string{messageNoUnwantedPolyfillioSecurity.Id},
		},
		{
			// Measured. An aliased default import binds whatever local name is written.
			name:   "aliased default next/script import",
			source: "import Poly from 'next/script';\nexport const A = () => <Poly src='https://polyfill.io/v3/polyfill.min.js' />;",
			want:   []string{messageNoUnwantedPolyfillioSecurity.Id},
		},
		{
			// Measured. A namespace import binds the tag name as well, the third shape the
			// default-import helper would have dropped.
			name:   "namespace next/script import",
			source: "import * as S from 'next/script';\nexport const A = () => <S src='https://polyfill.io/v3/polyfill.min.js' />;",
			want:   []string{messageNoUnwantedPolyfillioSecurity.Id},
		},
		{
			// Entities in the URL. ESLint's parser decodes a JSX attribute string, so the installed
			// plugin (16.3.1) reads the host and copyWithin in both and reports the duplicate
			// (#5f61r5b). On polyfill-fastly.io rather than polyfill.io, whose security arm answers
			// first here and would hide whether the feature list was decoded.
			name:   "a host spelled with character references",
			source: "export const A = () => <script src='https:&#47;&#47;polyfill-fastly.io/v3/polyfill.min.js?features=Array.prototype.copyWithin' />;",
			want:   []string{messageNoUnwantedPolyfillioDuplicate.Id},
		},
		{
			name:   "a feature spelled with a character reference",
			source: "export const A = () => <script src='https://polyfill-fastly.io/v3/polyfill.min.js?features=Array.prototype.copy&#87;ithin' />;",
			want:   []string{messageNoUnwantedPolyfillioDuplicate.Id},
		},
		{
			// The lowercase intrinsic tag needs no import at all, which is the arm the corpus only
			// covers incidentally.
			name:   "intrinsic script needs no import",
			source: "export const A = () => <script src='https://polyfill.io/v3/polyfill.min.js' />;",
			want:   []string{messageNoUnwantedPolyfillioSecurity.Id},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUnwantedPolyfillio, "Component.tsx", testCase.source)
			rule_testing.ExpectFindings(t, result, testCase.want...)
		})
	}
}

func TestNoUnwantedPolyfillioIsSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
	}{
		{
			// Upstream. AbortController is not in the shipped feature list, and the surrounding
			// next/document import and Head subclass are decoration the rule never reads.
			name:   "upstream pass 1",
			source: "import {Head} from 'next/document';\n        export class Blah extends Head {\n            render() {\n                return (\n                    <div>\n                        <h1>Hello title</h1>\n                        <script src='https://cdnjs.cloudflare.com/polyfill/v3/polyfill.min.js?features=AbortController'></script>\n                    </div>\n                );\n            }\n        }",
		},
		{
			// Upstream. IntersectionObserver is likewise not shipped.
			name:   "upstream pass 2",
			source: "import {Head} from 'next/document';\n        export class Blah extends Head {\n            render() {\n                return (\n                    <div>\n                        <h1>Hello title</h1>\n                        <script src='https://cdnjs.cloudflare.com/polyfill/v3/polyfill.min.js?features=IntersectionObserver'></script>\n                    </div>\n                );\n            }\n        }",
		},
		{
			// Upstream. The only pass case using the next/script import path, and it passes for the
			// feature reason rather than the import reason, so it does not test the import gate.
			name:   "upstream pass 3",
			source: "import Script from 'next/script';\n        export function MyApp({ Component, pageProps }) {\n            return (\n                <div>\n                    <Component {...pageProps} />\n                    <Script src='https://cdnjs.cloudflare.com/polyfill/v3/polyfill.min.js?features=IntersectionObserver' />\n                </div>\n            );\n        }",
		},
		{
			// Measured. Case-sensitive whole-token equality: the lowercase spelling is not the
			// entry in the list and does not match it.
			name:   "feature name in the wrong case",
			source: "export const A = () => <script src='https://cdnjs.cloudflare.com/polyfill/v3/p.js?features=promise' />;",
		},
		{
			// Measured, and the divergence a port would most likely repair. Upstream replaces only
			// the uppercase escape, so both tokens here keep their encoding and match nothing. A
			// port using Go's percent decoder would report and be wrong.
			name:   "lowercase percent escape stays undecoded",
			source: "export const A = () => <script src='https://cdnjs.cloudflare.com/polyfill/v3/p.js?features=Promise%2cSet' />;",
		},
		{
			// Measured. Nothing upstream asserts a non-matching host is silent, so a port that
			// dropped the prefix guard entirely would pass every imported case.
			name:   "unrelated host",
			source: "export const A = () => <script src='https://example.com/x.js?features=Promise' />;",
		},
		{
			// Measured. The prefixes carry their scheme, so the plain http spelling of a
			// compromised host does not match.
			name:   "http rather than https",
			source: "export const A = () => <script src='http://polyfill.io/v3/polyfill.min.js' />;",
		},
		{
			// Measured. A protocol-relative URL fails the same scheme-attached prefix test.
			name:   "protocol relative url",
			source: "export const A = () => <script src='//cdn.polyfill.io/v2/polyfill.min.js' />;",
		},
		{
			// Measured. A path-only src matches no prefix and the query reader declines it too.
			name:   "path only src",
			source: "export const A = () => <script src='/v3/polyfill.min.js?features=Promise' />;",
		},
		{
			// Measured. The tag gate broken shut is invisible to the corpus, which writes no
			// non-script element carrying one of these URLs.
			name:   "a div is not a script",
			source: "export const A = () => <div src='https://polyfill.io/v3/polyfill.min.js' />;",
		},
		{
			// Measured. The intrinsic match is exact, so the uppercase spelling is a component
			// reference and matches no import either.
			name:   "uppercase SCRIPT tag",
			source: "export const A = () => <SCRIPT src='https://polyfill.io/v3/polyfill.min.js' />;",
		},
		{
			// Measured, and the negative the corpus never writes: the component arm must decline
			// when the file imports nothing from next/script.
			name:   "Script component with no import",
			source: "export const A = () => <Script src='https://polyfill.io/v3/polyfill.min.js' />;",
		},
		{
			// Measured. A member-expression tag is not the bare local name and never matches.
			name:   "member expression tag",
			source: "import Script from 'next/script';\nexport const A = () => <Script.Sub src='https://polyfill.io/v3/polyfill.min.js' />;",
		},
		{
			// An import of a different module must not bind the tag name.
			name:   "import from another module",
			source: "import Script from 'other/script';\nexport const A = () => <Script src='https://polyfill.io/v3/polyfill.min.js' />;",
		},
		{
			// Measured. A safe host with no query at all: the query reader answers nothing and the
			// feature arm has nothing to split.
			name:   "safe host with no query",
			source: "export const A = () => <script src='https://cdnjs.cloudflare.com/polyfill/v3/polyfill.min.js' />;",
		},
		{
			// An empty features value splits to one empty token, which is not in the list.
			name:   "empty features value",
			source: "export const A = () => <script src='https://cdnjs.cloudflare.com/polyfill/v3/p.js?features=' />;",
		},
		{
			// Measured. The query reader takes the first matching key rather than the last, so the
			// unshipped AbortController wins and the Promise in the second copy is never read.
			name:   "duplicate features key takes the first",
			source: "export const A = () => <script src='https://cdnjs.cloudflare.com/polyfill/v3/p.js?features=AbortController&features=Promise' />;",
		},
		{
			// A src that is not a plain string literal cannot be read, so the rule declines rather
			// than guessing at the value. This is the one real usage shape in Kirk's tree.
			name:   "computed src expression",
			source: "import Script from 'next/script';\nexport const A = (p) => <Script src={p.url} />;",
		},
		{
			// An element with no src at all exits at the attribute lookup.
			name:   "script with no src",
			source: "export const A = () => <script>{\"var a = 1;\"}</script>;",
		},
		{
			// A spread carries no readable attribute name, so the rule sees no src.
			name:   "spread attributes only",
			source: "export const A = (p) => <script {...p} />;",
		},
		{
			// A feature name that is a prefix of a shipped entry is not that entry: matching is on
			// the whole comma-separated token rather than by prefix.
			name:   "feature name is a prefix of a shipped one",
			source: "export const A = () => <script src='https://cdnjs.cloudflare.com/polyfill/v3/p.js?features=Array.prototype' />;",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUnwantedPolyfillio, "Component.tsx", testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// The finding underlines the `src` attribute rather than the element, which is what oxc's snapshot
// shows and what no message id assertion can see. Upstream's underline covers `src='...'` and stops
// before the closing slash, so both ends are asserted here by slicing the source with the finding's
// own range.
func TestNoUnwantedPolyfillioPointsAtTheSourceAttribute(t *testing.T) {
	t.Parallel()

	source := "export const A = () => <script src='https://polyfill.io/v3/polyfill.min.js' />;"
	result := rule_testing.Run(t, NoUnwantedPolyfillio, "Component.tsx", source)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("wanted one finding, got %d", len(result.Diagnostics))
	}
	reported := source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
	if reported != "src='https://polyfill.io/v3/polyfill.min.js'" {
		t.Fatalf("finding underlines %q", reported)
	}

	// The duplicate arm reports at its own call site, so asserting the security arm's span says
	// nothing about it. A sweep pointing only that second report at the element survived a suite
	// carrying this test for the first arm alone, which is why both are asserted rather than one
	// standing in for the other.
	duplicateSource := "export const A = () => <script src='https://polyfill-fastly.io/v3/p.js?features=Promise' />;"
	duplicateResult := rule_testing.Run(t, NoUnwantedPolyfillio, "Component.tsx", duplicateSource)
	if len(duplicateResult.Diagnostics) != 1 {
		t.Fatalf("wanted one finding on the duplicate arm, got %d", len(duplicateResult.Diagnostics))
	}
	duplicateReported := duplicateSource[duplicateResult.Diagnostics[0].Range.Pos():duplicateResult.Diagnostics[0].Range.End()]
	if duplicateReported != "src='https://polyfill-fastly.io/v3/p.js?features=Promise'" {
		t.Fatalf("duplicate finding underlines %q", duplicateReported)
	}
}

// The duplicate message names the features it found, in the order the URL wrote them, and agrees
// its verb with how many there are. Both are visible in upstream's snapshot and neither is visible
// to an assertion on the message id, so the rendered text is compared exactly rather than with a
// containment check: a containment check passes on a string carrying extra text.
func TestNoUnwantedPolyfillioNamesTheFeaturesItFound(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
		want   string
	}{
		{
			// Both epsilon spellings are in upstream's list deliberately, and the lowercase `es5`
			// alias is one of nine entries whose case a hand copy would normalise. A fixture
			// asserting only the message id cannot see either: dropping one entry still leaves one
			// finding carrying the same id. Both were live survivors in the sweep until the
			// assertion moved to the rendered text, which is the "fixtures assert the wrong layer"
			// case rather than a blind spot.
			name:   "both epsilon spellings are named",
			source: "export const A = () => <script src='https://cdnjs.cloudflare.com/polyfill/v3/p.js?features=Number.Epsilon,Number.EPSILON' />;",
			want: "This script asks a polyfill service for Number.Epsilon, Number.EPSILON, which " +
				"Next.js already ships, so the bytes are downloaded and parsed to redefine what are " +
				"already there. Drop Number.Epsilon, Number.EPSILON from the `features` list.",
		},
		{
			name:   "lowercase es alias is named",
			source: "export const A = () => <script src='https://cdnjs.cloudflare.com/polyfill/v3/p.js?features=es5,es6' />;",
			want: "This script asks a polyfill service for es5, es6, which Next.js already ships, so " +
				"the bytes are downloaded and parsed to redefine what are already there. Drop es5, " +
				"es6 from the `features` list.",
		},
		{
			name:   "singular agrees its verb",
			source: "export const A = () => <script src='https://polyfill-fastly.io/v3/p.js?features=Array.prototype.copyWithin' />;",
			want: "This script asks a polyfill service for Array.prototype.copyWithin, which Next.js " +
				"already ships, so the bytes are downloaded and parsed to redefine what is already " +
				"there. Drop Array.prototype.copyWithin from the `features` list.",
		},
		{
			// Order follows the URL rather than the feature list, which is upstream's behaviour and
			// is only visible when the two orders differ. Promise sits before Object.fromEntries in
			// the URL and after it in nothing in particular, so the assertion pins the URL order.
			name:   "plural names both in url order",
			source: "export const A = () => <script src='https://cdnjs.cloudflare.com/polyfill/v3/p.js?features=Promise%2CObject.fromEntries' />;",
			want: "This script asks a polyfill service for Promise, Object.fromEntries, which Next.js " +
				"already ships, so the bytes are downloaded and parsed to redefine what are already " +
				"there. Drop Promise, Object.fromEntries from the `features` list.",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUnwantedPolyfillio, "Component.tsx", testCase.source)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one finding, got %d", len(result.Diagnostics))
			}
			if got := result.Diagnostics[0].Message.Description; got != testCase.want {
				t.Fatalf("message was %q", got)
			}
		})
	}
}

// Both message ids are pinned against literal strings typed here rather than against the rule's own
// constants. Asserting `messageNoUnwantedPolyfillioDuplicate.Id` looks like equality and is not a
// guard: a mutation renaming the constant moves both sides of the comparison together, and that
// mutant survived a suite whose every other assertion used the constant.
func TestNoUnwantedPolyfillioUsesStableMessageIds(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
		want   string
	}{
		{
			name:   "security arm",
			source: "export const A = () => <script src='https://polyfill.io/v3/polyfill.min.js' />;",
			want:   "noUnwantedPolyfillioSecurity",
		},
		{
			name:   "duplicate arm",
			source: "export const A = () => <script src='https://polyfill-fastly.io/v3/p.js?features=Promise' />;",
			want:   "noUnwantedPolyfillioDuplicate",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUnwantedPolyfillio, "Component.tsx", testCase.source)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one finding, got %d", len(result.Diagnostics))
			}
			if got := result.Diagnostics[0].Message.Id; got != testCase.want {
				t.Fatalf("message id was %q", got)
			}
		})
	}
}

// The security arm's text is constant rather than rendered, so nothing else in this file reads it,
// and a mutation rewriting the sentence survived until this existed. Compared with equality rather
// than containment: a containment check passes on a string that has grown extra text.
func TestNoUnwantedPolyfillioSecurityMessageText(t *testing.T) {
	t.Parallel()

	source := "export const A = () => <script src='https://polyfill.io/v3/polyfill.min.js' />;"
	result := rule_testing.Run(t, NoUnwantedPolyfillio, "Component.tsx", source)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("wanted one finding, got %d", len(result.Diagnostics))
	}
	want := "This script loads from polyfill.io, whose domain was sold in 2024 and then used to " +
		"serve malicious code to the sites embedding it. Every visitor runs whatever that domain " +
		"returns, so this is a live supply chain risk rather than a stale dependency. Point it at " +
		"https://cdnjs.cloudflare.com/polyfill/ or drop the polyfill and use the browser feature " +
		"directly."
	if got := result.Diagnostics[0].Message.Description; got != want {
		t.Fatalf("security message was %q", got)
	}
}

// The security arm reports its own message and returns, so a URL that would also have qualified as
// a duplicate is never described as one. This is the distinction a port folding the two arms into
// one code path loses, and half the imported corpus still passes without it.
func TestNoUnwantedPolyfillioSecurityArmReturnsBeforeTheFeatureCheck(t *testing.T) {
	t.Parallel()

	source := "export const A = () => <script src='https://cdn.polyfill.io/v2/polyfill.min.js?features=Promise' />;"
	result := rule_testing.Run(t, NoUnwantedPolyfillio, "Component.tsx", source)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("wanted one finding, got %d", len(result.Diagnostics))
	}
	if result.Diagnostics[0].Message.Id != "noUnwantedPolyfillioSecurity" {
		t.Fatalf("reported %q", result.Diagnostics[0].Message.Id)
	}
	if strings.Contains(result.Diagnostics[0].Message.Description, "Promise") {
		t.Fatalf("the security message named a feature: %q", result.Diagnostics[0].Message.Description)
	}
}
