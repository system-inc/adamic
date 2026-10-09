package next

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

func TestNoSyncScriptsReports(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
	}{
		{
			name:   "src with no async or defer",
			source: `export const C = () => <div><script src="https://blah.com"></script></div>;`,
		},
		{
			name:   "self closing form",
			source: `export const C = () => <script src="/a.js" />;`,
		},
		{
			// Another attribute being present must not read as async or defer being present.
			name:   "src with an unrelated attribute",
			source: `export const C = () => <script src="/a.js" type="text/javascript" />;`,
		},
		{
			// oxc's own fail case, and the one a value-reading port gets backwards. There is no
			// string literal to read here, so a rule that asked for the value of `src` would find
			// none and silently exempt a script that does block the parser. Presence is the right
			// question: the attribute is written, so the fetch happens.
			name:   "src is a computed expression",
			source: `export const C = (props) => <script src={props.src}></script>;`,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoSyncScripts, "Component.tsx", testCase.source)
			rule_testing.ExpectFindings(t, result, messageNoSyncScripts.Id)
		})
	}
}

func TestNoSyncScriptsIsSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
	}{
		{
			// The idiomatic spelling of a boolean attribute has no initializer at all. A port that
			// read the value rather than the presence would find nothing here and report.
			name:   "bare async",
			source: `export const C = () => <script src="/a.js" async />;`,
		},
		{
			name:   "bare defer",
			source: `export const C = () => <script src="/a.js" defer />;`,
		},
		{
			name:   "async with an explicit value",
			source: `export const C = () => <script src="/a.js" async={true} />;`,
		},
		{
			// An inline script has no fetch to defer, so async and defer would do nothing on it.
			name:   "inline script with no src",
			source: `export const C = () => <script>{"var a = 1;"}</script>;`,
		},
		{
			// oxc's paired pass case. A spread carries no readable name, so the rule cannot see a
			// `src` and must not assume one, which is what keeps the computed-src fail case above
			// from generalising into reporting every spread.
			name:   "spread attributes",
			source: `export const C = (props) => <script {...props}></script>;`,
		},
		{
			name:   "member expression name",
			source: `export const C = () => <Foo.script src="/a.js" />;`,
		},
		{
			name: "the next/script component",
			source: `import Script from 'next/script';
export const C = () => <Script src="/a.js" />;`,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoSyncScripts, "Component.tsx", testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}
