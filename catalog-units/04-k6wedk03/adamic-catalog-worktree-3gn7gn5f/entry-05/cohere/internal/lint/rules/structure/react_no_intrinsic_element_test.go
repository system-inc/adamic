package structure

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const anchorFile = "/repository/source/components/Page.tsx"
const linkImplementationFile = "/repository/source/components/navigation/Link.tsx"
const horizontalRuleImplementationFile = "/repository/source/components/layout/HorizontalRule.tsx"

func TestReactNoAnchorElementFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantCount  int
	}{
		{
			"an anchor with children",
			"export function Page() {\n    return <a href=\"/about\">About</a>;\n}\n",
			1,
		},
		// The self-closing form reaches a different node kind entirely, so a rule listening only to
		// JsxOpeningElement is silent on it. Same class of miss as no-var's loop forms.
		{
			"a self-closing anchor",
			"export function Page() {\n    return <a href=\"/about\" />;\n}\n",
			1,
		},
		{
			"two anchors report twice",
			"export function Page() {\n    return <div><a href=\"/a\">A</a><a href=\"/b\">B</a></div>;\n}\n",
			2,
		},
		{
			"an anchor nested inside other elements",
			"export function Page() {\n    return <div><span><a href=\"/a\">A</a></span></div>;\n}\n",
			1,
		},
		{
			"an anchor inside a fragment",
			"export function Page() {\n    return <><a href=\"/a\">A</a></>;\n}\n",
			1,
		},
		// A file named Link.tsx that is not the Link component's own implementation is not exempt.
		// The exemption is a path, not a base name, and this is the boundary that separates them.
		{
			"an anchor in a differently placed Link.tsx",
			"export function Link() {\n    return <a href=\"/a\">A</a>;\n}\n",
			1,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			fileName := anchorFile
			if testCase.name == "an anchor in a differently placed Link.tsx" {
				fileName = "/repository/source/widgets/Link.tsx"
			}
			result := rule_testing.Run(t, ReactElementNoAnchor, fileName, testCase.sourceText)
			ids := result.MessageIds()
			if len(ids) != testCase.wantCount {
				t.Fatalf("expected %d findings, got %d: %v", testCase.wantCount, len(ids), ids)
			}
			for _, id := range ids {
				if id != "noAnchorElement" {
					t.Fatalf("unexpected message id %q", id)
				}
			}
		})
	}
}

func TestReactNoAnchorElementStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		fileName   string
		sourceText string
	}{
		// The shape the rule asks for.
		{
			"the Link component",
			anchorFile,
			"declare function Link(properties: { href: string; children?: unknown }): unknown;\n" +
				"export function Page() {\n    return <Link href=\"/about\">About</Link>;\n}\n",
		},
		// The exemption, and the reason the rule can be satisfied at all: Link itself has to render
		// an anchor. Without this the rule forbids the thing it recommends.
		{
			"an anchor inside the Link implementation",
			linkImplementationFile,
			"export function Link() {\n    return <a href=\"/about\">About</a>;\n}\n",
		},
		{
			"a self-closing anchor inside the Link implementation",
			linkImplementationFile,
			"export function Link() {\n    return <a href=\"/about\" />;\n}\n",
		},
		// A capitalized name is a component reference, not the HTML tag. This is the boundary a
		// case-insensitive comparison would cross.
		{
			"a component named A",
			anchorFile,
			"declare function A(properties: { href: string }): unknown;\n" +
				"export function Page() {\n    return <A href=\"/about\" />;\n}\n",
		},
		// A qualified name never emits an anchor, whatever its last segment says.
		{
			"a member-expression element ending in a",
			anchorFile,
			"declare const Namespace: { a: (properties: { href: string }) => unknown };\n" +
				"export function Page() {\n    return <Namespace.a href=\"/about\" />;\n}\n",
		},
		{
			"other intrinsic elements",
			anchorFile,
			"export function Page() {\n    return <div><span>text</span><abbr>x</abbr></div>;\n}\n",
		},
		// The identifier `a` outside JSX is not an element and must not be read as one. This is the
		// trap that produced 3,081 spurious findings on the Tailwind port.
		{
			"a variable named a",
			anchorFile,
			"export function Page() {\n    const a = 1;\n    return <div>{a}</div>;\n}\n",
		},
		{
			"a file with no JSX at all",
			"/repository/source/api/Thing.ts",
			"export const value = 1;\n",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, ReactElementNoAnchor, testCase.fileName, testCase.sourceText))
		})
	}
}

func TestReactNoHorizontalRuleElementFires(t *testing.T) {
	t.Parallel()

	// hr is written self-closing essentially always, so this is the shape that matters most and it
	// is the one a JsxOpeningElement-only listener misses entirely.
	rule_testing.ExpectFindings(t, rule_testing.Run(t, ReactElementNoHorizontalRule, anchorFile,
		"export function Page() {\n    return <div><hr /></div>;\n}\n"), "noHrElement")

	rule_testing.ExpectFindings(t, rule_testing.Run(t, ReactElementNoHorizontalRule, anchorFile,
		"export function Page() {\n    return <div><hr /><hr /></div>;\n}\n"), "noHrElement", "noHrElement")
}

func TestReactNoHorizontalRuleElementStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		fileName   string
		sourceText string
	}{
		{
			"the HorizontalRule component",
			anchorFile,
			"declare function HorizontalRule(): unknown;\n" +
				"export function Page() {\n    return <HorizontalRule />;\n}\n",
		},
		{
			"an hr inside the HorizontalRule implementation",
			horizontalRuleImplementationFile,
			"export function HorizontalRule() {\n    return <hr />;\n}\n",
		},
		// The two rules must not answer for each other. An anchor is not this rule's business, and
		// this fails if the shared helper is handed the wrong tag name.
		{
			"an anchor is not this rule's business",
			anchorFile,
			"export function Page() {\n    return <a href=\"/a\">A</a>;\n}\n",
		},
		{
			"an hr inside the Link implementation is still reported by the other rule only",
			linkImplementationFile,
			"export function Link() {\n    return <a href=\"/a\">A</a>;\n}\n",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, ReactElementNoHorizontalRule, testCase.fileName, testCase.sourceText))
		})
	}
}
