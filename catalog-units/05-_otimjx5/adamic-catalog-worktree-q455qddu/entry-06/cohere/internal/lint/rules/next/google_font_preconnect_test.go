package next

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

func TestGoogleFontPreconnectReports(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
	}{
		{
			name:   "no rel at all",
			source: `export const Test = () => <div><link href="https://fonts.gstatic.com" /></div>;`,
		},
		{
			// A rel that is present but is not preconnect is the same finding upstream, and it is
			// the more common mistake of the two. oxc writes this as `is_none_or`.
			name:   "rel is present but wrong",
			source: `export const Test = () => <link rel="stylesheet" href="https://fonts.gstatic.com" />;`,
		},
		{
			// The attribute names are matched case-insensitively in this rule, unlike in
			// no-css-tags. Pinned so the two spellings cannot be quietly unified.
			name:   "uppercase attribute names",
			source: `export const Test = () => <link HREF="https://fonts.gstatic.com" />;`,
		},
		{
			name:   "a longer path on the font host",
			source: `export const Test = () => <link href="https://fonts.gstatic.com/s/font.woff2" />;`,
		},
		{
			// A rel the rule cannot read as a string is treated as not saying preconnect, which is
			// the stricter reading and is oxc's.
			name:   "rel is a computed expression",
			source: `export const Test = () => <link rel={kind} href="https://fonts.gstatic.com" />;`,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, GoogleFontPreconnect, "Component.tsx", testCase.source)
			rule_testing.ExpectFindings(t, result, messageGoogleFontPreconnect.Id)
		})
	}
}

func TestGoogleFontPreconnectIsSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
	}{
		{
			name:   "preconnect is present",
			source: `export const Test = () => <link rel="preconnect" href="https://fonts.gstatic.com" />;`,
		},
		{
			name:   "preconnect written in mixed case",
			source: `export const Test = () => <link REL="preconnect" href="https://fonts.gstatic.com" />;`,
		},
		{
			// oxc's own pass cases: a computed href cannot be read, so the rule declines rather
			// than guessing at the host.
			name:   "computed href",
			source: `export const Test = () => <link href={process.env.NEXT_PUBLIC_CANONICAL_URL} rel="canonical" />;`,
		},
		{
			name:   "href built from a URL object",
			source: `export const Test = () => <link href={new URL("../public/favicon.ico", import.meta.url).toString()} rel="icon" />;`,
		},
		{
			name:   "a different host",
			source: `export const Test = () => <link href="https://example.com/a.css" rel="stylesheet" />;`,
		},
		{
			// The stylesheet host is a different origin from the font host, and only the font host
			// is what this rule is about.
			name:   "the googleapis host rather than gstatic",
			source: `export const Test = () => <link href="https://fonts.googleapis.com/css?family=Krona+One" rel="stylesheet" />;`,
		},
		{
			name:   "no href",
			source: `export const Test = () => <link rel="preconnect" />;`,
		},
		{
			name:   "spread attributes",
			source: `export const Test = (props) => <link {...props} />;`,
		},
		{
			name:   "member expression name",
			source: `export const Test = () => <Foo.link href="https://fonts.gstatic.com" />;`,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, GoogleFontPreconnect, "Component.tsx", testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestGoogleFontPreconnectDecodesEntities reads its attribute as ESLint does, with HTML entities decoded: typescript-estree decodes a
// JSX attribute string before any rule sees it. Each row's verdict is the installed
// @next/eslint-plugin-next 16.3.1's under the typescript-eslint parser (#51y9jh2).
func TestGoogleFontPreconnectDecodesEntities(t *testing.T) {
	t.Parallel()

	for _, row := range []struct {
		name   string
		source string
		ids    []string
	}{
		{`encoded slashes in the font host`, `export const Test = () => <link href="https:&#47;&#47;fonts.gstatic.com" />;`, []string{messageGoogleFontPreconnect.Id}},
		{`an encoded letter in rel`, `export const Test = () => <link rel="pre&#99;onnect" href="https://fonts.gstatic.com" />;`, []string{}},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, GoogleFontPreconnect, "Component.tsx", row.source)
			if len(row.ids) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, row.ids...)
		})
	}
}
