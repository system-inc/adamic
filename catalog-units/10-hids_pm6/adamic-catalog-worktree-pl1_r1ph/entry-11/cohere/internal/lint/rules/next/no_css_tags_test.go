package next

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

func TestNoCssTagsReports(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
	}{
		{
			name:   "local stylesheet",
			source: `export const C = () => <div><link href="/_next/static/css/styles.css" rel="stylesheet" /></div>;`,
		},
		{
			// Attribute order is not part of the decision. Written out because a port that walked
			// to the first attribute rather than searching by name would pass the case above.
			name:   "rel written before href",
			source: `export const C = () => <link rel="stylesheet" href="/a.css" />;`,
		},
		{
			name:   "root-relative path with no leading slash",
			source: `export const C = () => <link rel="stylesheet" href="styles.css" />;`,
		},
		{
			// Protocol-relative is still not an http(s) prefix, so it is local by the rule's own
			// test. Pinned because it is the boundary case of the prefix check.
			name:   "protocol relative",
			source: `export const C = () => <link rel="stylesheet" href="//example.com/a.css" />;`,
		},
		{
			name:   "paired form",
			source: `export const C = () => <link rel="stylesheet" href="/a.css"></link>;`,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoCssTags, "Component.tsx", testCase.source)
			rule_testing.ExpectFindings(t, result, messageNoCssTags.Id)
		})
	}
}

func TestNoCssTagsIsSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
	}{
		{
			name:   "no link at all",
			source: `export const C = () => <div><h1>Hello title</h1></div>;`,
		},
		{
			name:   "absolute https stylesheet",
			source: `export const C = () => <link href="https://fonts.googleapis.com/css?family=Open+Sans" rel="stylesheet" />;`,
		},
		{
			name:   "absolute http stylesheet",
			source: `export const C = () => <link href="http://example.com/a.css" rel="stylesheet" />;`,
		},
		{
			// oxc's own passing case, and the one a porter is least likely to invent. A spread
			// carries no readable name or value, so the rule must decline rather than assume.
			name:   "spread attributes",
			source: `export const C = (props) => <link {...props} />;`,
		},
		{
			name:   "computed href",
			source: `export const C = () => <link rel="stylesheet" href={path} />;`,
		},
		{
			name:   "computed rel",
			source: `export const C = () => <link rel={kind} href="/a.css" />;`,
		},
		{
			name:   "a link that is not a stylesheet",
			source: `export const C = () => <link rel="icon" href="/favicon.ico" />;`,
		},
		{
			name:   "href absent",
			source: `export const C = () => <link rel="stylesheet" />;`,
		},
		{
			name:   "member expression name",
			source: `export const C = () => <Foo.link rel="stylesheet" href="/a.css" />;`,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoCssTags, "Component.tsx", testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestNoCssTagsDecodesEntities reads its attribute as ESLint does, with HTML entities decoded: typescript-estree decodes a
// JSX attribute string before any rule sees it. Each row's verdict is the installed
// @next/eslint-plugin-next 16.3.1's under the typescript-eslint parser (#51y9jh2).
func TestNoCssTagsDecodesEntities(t *testing.T) {
	t.Parallel()

	for _, row := range []struct {
		name   string
		source string
		ids    []string
	}{
		{`an encoded letter in rel`, `export const C = () => <link rel="style&#115;heet" href="/a.css" />;`, []string{messageNoCssTags.Id}},
		{`encoded slashes in an absolute href`, `export const C = () => <link rel="stylesheet" href="https:&#47;&#47;example.com/a.css" />;`, []string{}},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoCssTags, "Component.tsx", row.source)
			if len(row.ids) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, row.ids...)
		})
	}
}
