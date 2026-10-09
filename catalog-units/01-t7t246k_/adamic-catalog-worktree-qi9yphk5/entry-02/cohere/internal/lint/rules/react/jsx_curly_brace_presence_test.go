package react

import (
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// jsxCurlyBracePresenceRun routes a fixture through the rule's own decoder.
//
// The options arrive as the JSON text a configuration file would hold, rather than as a Go struct,
// because two of the three defaults are not zero values and one of the two accepted wire shapes is
// a bare string. A fixture that built `JsxCurlyBracePresenceOptions` directly would exercise
// neither, and both are exactly the lines with no upstream counterpart.
func jsxCurlyBracePresenceRun(t *testing.T, source string, optionsJson string) rule_testing.Result {
	t.Helper()

	decoded, err := DecodeJsxCurlyBracePresenceOptions([]byte(optionsJson))
	if err != nil {
		t.Fatalf("decoding options %q: %v", optionsJson, err)
	}
	return rule_testing.RunWithOptions(t, JsxCurlyBracePresence, "component.tsx", source, decoded)
}

// Clean cases, imported verbatim from upstream's corpus.
func TestJsxCurlyBracePresenceStaysSilent(t *testing.T) {
	t.Parallel()

	for index, testCase := range []struct {
		source  string
		options string
	}{
		{source: "<App {...props}>foo</App>", options: ""},
		{source: "<>foo</>", options: ""},
		{source: "<App {...props}>foo</App>", options: "{\"props\":\"never\"}"},
		{source: "<App>{' '}</App>", options: ""},
		{source: "<App>{' '}\n</App>", options: ""},
		{source: "<App>{'     '}</App>", options: ""},
		{source: "<App>{'     '}\n</App>", options: ""},
		{source: "<App>{' '}</App>", options: "{\"children\":\"never\"}"},
		{source: "<App>{'    '}</App>", options: "{\"children\":\"never\"}"},
		{source: "<App>{' '}</App>", options: "{\"children\":\"always\"}"},
		{source: "<App>{'        '}</App>", options: "{\"children\":\"always\"}"},
		{source: "<App {...props}>foo</App>", options: "{\"props\":\"always\"}"},
		{source: "<App>{`Hello ${word} World`}</App>", options: "{\"children\":\"never\"}"},
		{source: "\n        <React.Fragment>\n          foo{' '}\n          <span>bar</span>\n        </React.Fragment>\n      ", options: "{\"children\":\"never\"}"},
		{source: "\n        <>\n          foo{' '}\n          <span>bar</span>\n        </>\n      ", options: "{\"children\":\"never\"}"},
		{source: "<App>{`Hello \\n World`}</App>", options: "{\"children\":\"never\"}"},
		{source: "<App>{`Hello ${word} World`}{`foo`}</App>", options: "{\"children\":\"never\"}"},
		{source: "<App prop={`foo ${word} bar`}>foo</App>", options: "{\"props\":\"never\"}"},
		{source: "<App prop={`foo ${word} bar`} />", options: "{\"props\":\"never\"}"},
		{source: "<App>{<myApp></myApp>}</App>", options: "{\"children\":\"always\"}"},
		{source: "<App>{[]}</App>", options: ""},
		{source: "<App>foo</App>", options: ""},
		{source: "<App>{\"foo\"}{<Component>bar</Component>}</App>", options: ""},
		{source: "<App prop='bar'>foo</App>", options: ""},
		{source: "<App prop={true}>foo</App>", options: ""},
		{source: "<App prop>foo</App>", options: ""},
		{source: "<App prop='bar'>{'foo \\n bar'}</App>", options: ""},
		{source: "<App prop={ ' ' }/>", options: ""},
		{source: "<MyComponent prop='bar'>foo</MyComponent>", options: "{\"props\":\"never\"}"},
		{source: "<MyComponent prop=\"bar\">foo</MyComponent>", options: "{\"props\":\"never\"}"},
		{source: "<MyComponent>foo</MyComponent>", options: "{\"children\":\"never\"}"},
		{source: "<MyComponent>{<App/>}{\"123\"}</MyComponent>", options: "{\"children\":\"never\"}"},
		{source: "<App>{\"foo 'bar' \\\"foo\\\" bar\"}</App>", options: "{\"children\":\"never\"}"},
		{source: "<MyComponent prop={'bar'}>foo</MyComponent>", options: "{\"props\":\"always\"}"},
		{source: "<MyComponent>{'foo'}</MyComponent>", options: "{\"children\":\"always\"}"},
		{source: "<MyComponent prop={\"bar\"}>foo</MyComponent>", options: "{\"props\":\"always\"}"},
		{source: "<MyComponent>{\"foo\"}</MyComponent>", options: "{\"children\":\"always\"}"},
		{source: "<MyComponent>{'foo'}</MyComponent>", options: "{\"children\":\"ignore\"}"},
		{source: "<MyComponent prop={'bar'}>foo</MyComponent>", options: "{\"props\":\"ignore\"}"},
		{source: "<MyComponent>foo</MyComponent>", options: "{\"children\":\"ignore\"}"},
		{source: "<MyComponent prop='bar'>foo</MyComponent>", options: "{\"props\":\"ignore\"}"},
		{source: "<MyComponent prop=\"bar\">foo</MyComponent>", options: "{\"props\":\"ignore\"}"},
		{source: "<MyComponent prop='bar'>{'foo'}</MyComponent>", options: "{\"children\":\"always\",\"props\":\"never\"}"},
		{source: "<MyComponent prop={'bar'}>foo</MyComponent>", options: "{\"children\":\"never\",\"props\":\"always\"}"},
		{source: "<MyComponent prop={'bar'}>{'foo'}</MyComponent>", options: "\"always\""},
		{source: "<MyComponent prop={\"bar\"}>{\"foo\"}</MyComponent>", options: "\"always\""},
		{source: "<MyComponent prop={\"bar\"} attr={'foo'} />", options: "\"always\""},
		{source: "<MyComponent prop=\"bar\" attr='foo' />", options: "\"never\""},
		{source: "<MyComponent prop='bar'>foo</MyComponent>", options: "\"never\""},
		{source: "<MyComponent prop={`bar ${word} foo`}>{`foo ${word}`}</MyComponent>", options: "\"never\""},
		{source: "<MyComponent>{\"div { margin-top: 0; }\"}</MyComponent>", options: "\"never\""},
		{source: "<MyComponent>{\"<Foo />\"}</MyComponent>", options: "\"never\""},
		{source: "<MyComponent prop={\"Hello \\u1026 world\"}>bar</MyComponent>", options: "\"never\""},
		{source: "<MyComponent>{\"Hello \\u1026 world\"}</MyComponent>", options: "\"never\""},
		{source: "<MyComponent prop={\"Hello &middot; world\"}>bar</MyComponent>", options: "\"never\""},
		{source: "<MyComponent>{\"Hello &middot; world\"}</MyComponent>", options: "\"never\""},
		{source: "<MyComponent>{\"Hello \\n world\"}</MyComponent>", options: "\"never\""},
		{source: "<MyComponent>{\"space after \"}</MyComponent>", options: "\"never\""},
		{source: "<MyComponent>{\" space before\"}</MyComponent>", options: "\"never\""},
		{source: "<MyComponent>{`space after `}</MyComponent>", options: "\"never\""},
		{source: "<MyComponent>{` space before`}</MyComponent>", options: "\"never\""},
		{source: "<a a={\"start\\/n\\/nend\"}/>", options: "\"never\""},
		{source: "\n        <App prop={`\n          a\n          b\n        `} />\n      ", options: "\"never\""},
		{source: "\n        <App prop={`\n          a\n          b\n        `} />\n      ", options: "\"always\""},
		{source: "\n        <App>\n          {`\n            a\n            b\n          `}\n        </App>\n      ", options: "\"never\""},
		{source: "\n        <App>{`\n          a\n          b\n        `}</App>\n      ", options: "\"always\""},
		{source: "\n        <MyComponent>\n          %\n        </MyComponent>\n      ", options: "{\"children\":\"never\"}"},
		{source: "\n        <MyComponent>\n          { 'space after ' }\n          <b>foo</b>\n          { ' space before' }\n        </MyComponent>\n      ", options: "{\"children\":\"never\"}"},
		{source: "\n        <MyComponent>\n          { `space after ` }\n          <b>foo</b>\n          { ` space before` }\n        </MyComponent>\n      ", options: "{\"children\":\"never\"}"},
		{source: "\n        <MyComponent>\n          foo\n          <div>bar</div>\n        </MyComponent>\n      ", options: "{\"children\":\"never\"}"},
		{source: "\n        <MyComponent p={<Foo>Bar</Foo>}>\n        </MyComponent>\n      ", options: ""},
		{source: "\n        <MyComponent>\n          <div>\n            <p>\n              <span>\n                {\"foo\"}\n              </span>\n            </p>\n          </div>\n        </MyComponent>\n      ", options: "{\"children\":\"always\"}"},
		{source: "\n        <App>\n          <Component />&nbsp;\n          &nbsp;\n        </App>\n      ", options: "{\"children\":\"always\"}"},
		{source: "\n        const Component2 = () => {\n          return <span>/*</span>;\n        };\n      ", options: ""},
		{source: "\n        const Component2 = () => {\n          return <span>/*</span>;\n        };\n      ", options: "{\"props\":\"never\",\"children\":\"never\"}"},
		{source: "\n        import React from \"react\";\n\n        const Component = () => {\n          return <span>{\"/*\"}</span>;\n        };\n      ", options: "{\"props\":\"never\",\"children\":\"never\"}"},
		{source: "<App>{/* comment */}</App>", options: ""},
		{source: "<App>{/* comment */ <Foo />}</App>", options: ""},
		{source: "<App>{/* comment */ 'foo'}</App>", options: ""},
		{source: "<App prop={/* comment */ 'foo'} />", options: ""},
		{source: "\n          <App>\n            {\n              // comment\n              <Foo />\n            }\n          </App>\n        ", options: ""},
		{source: "<App horror=<div /> />", options: ""},
		{source: "<App horror={<div />} />", options: ""},
		{source: "<App horror=<div /> />", options: "{\"propElementValues\":\"ignore\"}"},
		{source: "<App horror={<div />} />", options: "{\"propElementValues\":\"ignore\"}"},
		{source: "\n        <script>{`window.foo = \"bar\"`}</script>\n      ", options: ""},
		{source: "\n        <CollapsibleTitle\n          extra={<span className=\"activity-type\">{activity.type}</span>}\n        />\n      ", options: "\"never\""},
		{source: "<App label={`${label}`} />", options: "\"never\""},
		{source: "<App>{`${label}`}</App>", options: "\"never\""},
	} {
		result := jsxCurlyBracePresenceRun(t, testCase.source, testCase.options)
		if len(result.Diagnostics) != 0 {
			t.Errorf("case %d: expected no findings for %q with options %q, got %v",
				index, testCase.source, testCase.options, result.MessageIds())
		}
	}
}

// Reporting cases, imported verbatim from upstream's corpus, with the repair each asserts.
func TestJsxCurlyBracePresenceFires(t *testing.T) {
	t.Parallel()

	for index, testCase := range []struct {
		source  string
		options string
		wantIds []string
		wantFix string
	}{
		{
			source:  "<App prop={`foo`} />",
			options: "{\"props\":\"never\"}",
			wantIds: []string{"UnnecessaryCurly"},
			wantFix: "<App prop=\"foo\" />",
		},
		{
			source:  "<App>{<myApp></myApp>}</App>",
			options: "{\"children\":\"never\"}",
			wantIds: []string{"UnnecessaryCurly"},
			wantFix: "<App><myApp></myApp></App>",
		},
		{
			source:  "<App>{<myApp></myApp>}</App>",
			options: "",
			wantIds: []string{"UnnecessaryCurly"},
			wantFix: "<App><myApp></myApp></App>",
		},
		{
			source:  "<App prop={`foo`}>foo</App>",
			options: "{\"props\":\"never\"}",
			wantIds: []string{"UnnecessaryCurly"},
			wantFix: "<App prop=\"foo\">foo</App>",
		},
		{
			source:  "<App>{`foo`}</App>",
			options: "{\"children\":\"never\"}",
			wantIds: []string{"UnnecessaryCurly"},
			wantFix: "<App>foo</App>",
		},
		{
			source:  "<>{`foo`}</>",
			options: "{\"children\":\"never\"}",
			wantIds: []string{"UnnecessaryCurly"},
			wantFix: "<>foo</>",
		},
		{
			source:  "<MyComponent>{'foo'}</MyComponent>",
			options: "",
			wantIds: []string{"UnnecessaryCurly"},
			wantFix: "<MyComponent>foo</MyComponent>",
		},
		{
			source:  "<MyComponent prop={'bar'}>foo</MyComponent>",
			options: "",
			wantIds: []string{"UnnecessaryCurly"},
			wantFix: "<MyComponent prop=\"bar\">foo</MyComponent>",
		},
		{
			source:  "<MyComponent>{'foo'}</MyComponent>",
			options: "{\"children\":\"never\"}",
			wantIds: []string{"UnnecessaryCurly"},
			wantFix: "<MyComponent>foo</MyComponent>",
		},
		{
			source:  "<MyComponent prop={'bar'}>foo</MyComponent>",
			options: "{\"props\":\"never\"}",
			wantIds: []string{"UnnecessaryCurly"},
			wantFix: "<MyComponent prop=\"bar\">foo</MyComponent>",
		},
		{
			source:  "\n        <MyComponent>\n          {'%'}\n        </MyComponent>\n      ",
			options: "{\"children\":\"never\"}",
			wantIds: []string{"UnnecessaryCurly"},
			wantFix: "\n        <MyComponent>\n          %\n        </MyComponent>\n      ",
		},
		{
			source:  "\n        <MyComponent>\n          {'foo'}\n          <div>\n            {'bar'}\n          </div>\n          {'baz'}\n        </MyComponent>\n      ",
			options: "{\"children\":\"never\"}",
			wantIds: []string{"UnnecessaryCurly", "UnnecessaryCurly", "UnnecessaryCurly"},
			wantFix: "\n        <MyComponent>\n          foo\n          <div>\n            bar\n          </div>\n          baz\n        </MyComponent>\n      ",
		},
		{
			source:  "\n        <MyComponent>\n          {'foo'}\n          <div>\n            {'bar'}\n          </div>\n          {'baz'}\n          {'some-complicated-exp'}\n        </MyComponent>\n      ",
			options: "{\"children\":\"never\"}",
			wantIds: []string{"UnnecessaryCurly", "UnnecessaryCurly", "UnnecessaryCurly", "UnnecessaryCurly"},
			wantFix: "\n        <MyComponent>\n          foo\n          <div>\n            bar\n          </div>\n          baz\n          some-complicated-exp\n        </MyComponent>\n      ",
		},
		{
			source:  "<MyComponent prop='bar'>foo</MyComponent>",
			options: "{\"props\":\"always\"}",
			wantIds: []string{"MissingCurly"},
			wantFix: "<MyComponent prop={\"bar\"}>foo</MyComponent>",
		},
		{
			source:  "<MyComponent prop=\"foo 'bar'\">foo</MyComponent>",
			options: "{\"props\":\"always\"}",
			wantIds: []string{"MissingCurly"},
			wantFix: "<MyComponent prop={\"foo 'bar'\"}>foo</MyComponent>",
		},
		{
			source:  "<MyComponent prop='foo \"bar\"'>foo</MyComponent>",
			options: "{\"props\":\"always\"}",
			wantIds: []string{"MissingCurly"},
			wantFix: "<MyComponent prop={\"foo \\\"bar\\\"\"}>foo</MyComponent>",
		},
		{
			source:  "<MyComponent>foo bar </MyComponent>",
			options: "{\"children\":\"always\"}",
			wantIds: []string{"MissingCurly"},
			wantFix: "<MyComponent>{\"foo bar \"}</MyComponent>",
		},
		{
			source:  "<MyComponent prop=\"foo 'bar' \\n \">foo</MyComponent>",
			options: "{\"props\":\"always\"}",
			wantIds: []string{"MissingCurly"},
			wantFix: "<MyComponent prop={\"foo 'bar' \\\\n \"}>foo</MyComponent>",
		},
		{
			source:  "<MyComponent>foo bar \\r </MyComponent>",
			options: "{\"children\":\"always\"}",
			wantIds: []string{"MissingCurly"},
			wantFix: "<MyComponent>{\"foo bar \\\\r \"}</MyComponent>",
		},
		{
			source:  "<MyComponent>foo bar 'foo'</MyComponent>",
			options: "{\"children\":\"always\"}",
			wantIds: []string{"MissingCurly"},
			wantFix: "<MyComponent>{\"foo bar 'foo'\"}</MyComponent>",
		},
		{
			source:  "<MyComponent>foo bar \"foo\"</MyComponent>",
			options: "{\"children\":\"always\"}",
			wantIds: []string{"MissingCurly"},
			wantFix: "<MyComponent>{\"foo bar \\\"foo\\\"\"}</MyComponent>",
		},
		{
			source:  "<MyComponent>foo bar <App/></MyComponent>",
			options: "{\"children\":\"always\"}",
			wantIds: []string{"MissingCurly"},
			wantFix: "<MyComponent>{\"foo bar \"}<App/></MyComponent>",
		},
		{
			source:  "<MyComponent>foo \\n bar</MyComponent>",
			options: "{\"children\":\"always\"}",
			wantIds: []string{"MissingCurly"},
			wantFix: "<MyComponent>{\"foo \\\\n bar\"}</MyComponent>",
		},
		{
			source:  "<MyComponent>foo \\u1234 bar</MyComponent>",
			options: "{\"children\":\"always\"}",
			wantIds: []string{"MissingCurly"},
			wantFix: "<MyComponent>{\"foo \\\\u1234 bar\"}</MyComponent>",
		},
		{
			source:  "<MyComponent prop='foo \\u1234 bar' />",
			options: "{\"props\":\"always\"}",
			wantIds: []string{"MissingCurly"},
			wantFix: "<MyComponent prop={\"foo \\\\u1234 bar\"} />",
		},
		{
			source:  "<MyComponent prop={'bar'}>{'foo'}</MyComponent>",
			options: "\"never\"",
			wantIds: []string{"UnnecessaryCurly", "UnnecessaryCurly"},
			wantFix: "<MyComponent prop=\"bar\">foo</MyComponent>",
		},
		{
			source:  "<MyComponent prop='bar'>foo</MyComponent>",
			options: "\"always\"",
			wantIds: []string{"MissingCurly", "MissingCurly"},
			wantFix: "<MyComponent prop={\"bar\"}>{\"foo\"}</MyComponent>",
		},
		{
			source:  "<App prop={'foo'} attr={\" foo \"} />",
			options: "{\"props\":\"never\"}",
			wantIds: []string{"UnnecessaryCurly", "UnnecessaryCurly"},
			wantFix: "<App prop=\"foo\" attr=\" foo \" />",
		},
		{
			source:  "<App prop='foo' attr=\"bar\" />",
			options: "{\"props\":\"always\"}",
			wantIds: []string{"MissingCurly", "MissingCurly"},
			wantFix: "<App prop={\"foo\"} attr={\"bar\"} />",
		},
		{
			source:  "<App prop='foo' attr={\"bar\"} />",
			options: "{\"props\":\"always\"}",
			wantIds: []string{"MissingCurly"},
			wantFix: "<App prop={\"foo\"} attr={\"bar\"} />",
		},
		{
			source:  "<App prop={'foo'} attr='bar' />",
			options: "{\"props\":\"always\"}",
			wantIds: []string{"MissingCurly"},
			wantFix: "<App prop={'foo'} attr={\"bar\"} />",
		},
		{
			source:  "<App prop='foo &middot; bar' />",
			options: "{\"props\":\"always\"}",
			wantIds: []string{"MissingCurly"},
			wantFix: "<App prop={\"foo &middot; bar\"} />",
		},
		{
			source:  "<App>foo &middot; bar</App>",
			options: "{\"children\":\"always\"}",
			wantIds: []string{"MissingCurly"},
			wantFix: "<App>{\"foo &middot; bar\"}</App>",
		},
		{
			source:  "<App>{'foo \"bar\"'}</App>",
			options: "{\"children\":\"never\"}",
			wantIds: []string{"UnnecessaryCurly"},
			wantFix: "<App>foo \"bar\"</App>",
		},
		{
			source:  "<App>{\"foo 'bar'\"}</App>",
			options: "{\"children\":\"never\"}",
			wantIds: []string{"UnnecessaryCurly"},
			wantFix: "<App>foo 'bar'</App>",
		},
		{
			source:  "\n        <App prop=\"    \n           a     \n             b      c\n                d\n        \">\n          a\n              b     c   \n                 d      \n        </App>\n      ",
			options: "\"always\"",
			wantIds: []string{"MissingCurly", "MissingCurly"},
			wantFix: "\n        <App prop=\"    \n           a     \n             b      c\n                d\n        \">\n          {\"a\"}\n              {\"b     c   \"}\n                 {\"d      \"}\n        </App>\n      ",
		},
		{
			source:  "\n        <App prop='    \n           a     \n             b      c\n                d\n        '>\n          a\n              b     c   \n                 d      \n        </App>\n      ",
			options: "\"always\"",
			wantIds: []string{"MissingCurly", "MissingCurly"},
			wantFix: "\n        <App prop='    \n           a     \n             b      c\n                d\n        '>\n          {\"a\"}\n              {\"b     c   \"}\n                 {\"d      \"}\n        </App>\n      ",
		},
		{
			source:  "\n        <App>\n          foo bar\n          <div>foo bar foo</div>\n          <span>\n            foo bar <i>foo bar</i>\n            <strong>\n              foo bar\n            </strong>\n          </span>\n        </App>\n      ",
			options: "{\"children\":\"always\"}",
			wantIds: []string{"MissingCurly", "MissingCurly", "MissingCurly", "MissingCurly", "MissingCurly"},
			wantFix: "\n        <App>\n          {\"foo bar\"}\n          <div>{\"foo bar foo\"}</div>\n          <span>\n            {\"foo bar \"}<i>{\"foo bar\"}</i>\n            <strong>\n              {\"foo bar\"}\n            </strong>\n          </span>\n        </App>\n      ",
		},
		{
			source:  "\n        <App>\n          &lt;Component&gt;\n          &nbsp;<Component />&nbsp;\n          &nbsp;\n        </App>\n      ",
			options: "{\"children\":\"always\"}",
			wantIds: []string{"MissingCurly"},
			wantFix: "\n        <App>\n          &lt;{\"Component\"}&gt;\n          &nbsp;<Component />&nbsp;\n          &nbsp;\n        </App>\n      ",
		},
		{
			source:  "\n        <Box mb={'1rem'} />\n      ",
			options: "{\"props\":\"never\"}",
			wantIds: []string{"UnnecessaryCurly"},
			wantFix: "\n        <Box mb=\"1rem\" />\n      ",
		},
		{
			source:  "\n        <Box mb={'1rem {}'} />\n      ",
			options: "\"never\"",
			wantIds: []string{"UnnecessaryCurly"},
			wantFix: "\n        <Box mb=\"1rem {}\" />\n      ",
		},
		{
			source:  "<MyComponent prop={\"{ style: true }\"}>bar</MyComponent>",
			options: "\"never\"",
			wantIds: []string{"UnnecessaryCurly"},
			wantFix: "<MyComponent prop=\"{ style: true }\">bar</MyComponent>",
		},
		{
			source:  "<MyComponent prop={\"< style: true >\"}>foo</MyComponent>",
			options: "\"never\"",
			wantIds: []string{"UnnecessaryCurly"},
			wantFix: "<MyComponent prop=\"< style: true >\">foo</MyComponent>",
		},
		{
			source:  "<App horror=<div /> />",
			options: "{\"props\":\"always\",\"children\":\"always\",\"propElementValues\":\"always\"}",
			wantIds: []string{"MissingCurly"},
			wantFix: "<App horror={<div />} />",
		},
		{
			source:  "<App horror={<div />} />",
			options: "{\"props\":\"never\",\"children\":\"never\",\"propElementValues\":\"never\"}",
			wantIds: []string{"UnnecessaryCurly"},
			wantFix: "<App horror=<div /> />",
		},
		{
			source:  "<Foo bar={\"'\"} />",
			options: "{\"props\":\"never\",\"children\":\"never\",\"propElementValues\":\"never\"}",
			wantIds: []string{"UnnecessaryCurly"},
			wantFix: "<Foo bar=\"'\" />",
		},
		{
			source:  "\n        <Foo help={'The maximum time range for searches. (i.e. \"P30D\" for 30 days, \"PT24H\" for 24 hours)'} />\n      ",
			options: "\"never\"",
			wantIds: []string{"UnnecessaryCurly"},
			wantFix: "\n        <Foo help='The maximum time range for searches. (i.e. \"P30D\" for 30 days, \"PT24H\" for 24 hours)' />\n      ",
		},
	} {
		result := jsxCurlyBracePresenceRun(t, testCase.source, testCase.options)
		got := result.MessageIds()
		if len(got) != len(testCase.wantIds) {
			t.Errorf("case %d: expected %v for %q, got %v", index, testCase.wantIds, testCase.source, got)
			continue
		}
		for position, want := range testCase.wantIds {
			if got[position] != want {
				t.Errorf("case %d: finding %d expected %q, got %q (source %q)",
					index, position, want, got[position], testCase.source)
			}
		}
	}
}

// Repairs, imported verbatim from the `output` each reporting case asserts upstream.
//
// This is the half a message-id fixture structurally cannot see. `ExpectFixedSource` replays
// every proposed fix into the source and compares the whole rewritten file, so a fix that
// repairs the right span with the wrong text fails here and passes everywhere else.
//
// A case whose upstream `output` equals its input is excluded, because upstream writes that
// when the fixer DECLINES; those are asserted separately by name below.
func TestJsxCurlyBracePresenceRepairsSource(t *testing.T) {
	t.Parallel()

	for index, testCase := range []struct {
		source  string
		options string
		want    string
	}{
		{
			source:  "<App prop={`foo`} />",
			options: "{\"props\":\"never\"}",
			want:    "<App prop=\"foo\" />",
		},
		{
			source:  "<App>{<myApp></myApp>}</App>",
			options: "{\"children\":\"never\"}",
			want:    "<App><myApp></myApp></App>",
		},
		{
			source:  "<App>{<myApp></myApp>}</App>",
			options: "",
			want:    "<App><myApp></myApp></App>",
		},
		{
			source:  "<App prop={`foo`}>foo</App>",
			options: "{\"props\":\"never\"}",
			want:    "<App prop=\"foo\">foo</App>",
		},
		{
			source:  "<App>{`foo`}</App>",
			options: "{\"children\":\"never\"}",
			want:    "<App>foo</App>",
		},
		{
			source:  "<>{`foo`}</>",
			options: "{\"children\":\"never\"}",
			want:    "<>foo</>",
		},
		{
			source:  "<MyComponent>{'foo'}</MyComponent>",
			options: "",
			want:    "<MyComponent>foo</MyComponent>",
		},
		{
			source:  "<MyComponent prop={'bar'}>foo</MyComponent>",
			options: "",
			want:    "<MyComponent prop=\"bar\">foo</MyComponent>",
		},
		{
			source:  "<MyComponent>{'foo'}</MyComponent>",
			options: "{\"children\":\"never\"}",
			want:    "<MyComponent>foo</MyComponent>",
		},
		{
			source:  "<MyComponent prop={'bar'}>foo</MyComponent>",
			options: "{\"props\":\"never\"}",
			want:    "<MyComponent prop=\"bar\">foo</MyComponent>",
		},
		{
			source:  "\n        <MyComponent>\n          {'%'}\n        </MyComponent>\n      ",
			options: "{\"children\":\"never\"}",
			want:    "\n        <MyComponent>\n          %\n        </MyComponent>\n      ",
		},
		{
			source:  "\n        <MyComponent>\n          {'foo'}\n          <div>\n            {'bar'}\n          </div>\n          {'baz'}\n        </MyComponent>\n      ",
			options: "{\"children\":\"never\"}",
			want:    "\n        <MyComponent>\n          foo\n          <div>\n            bar\n          </div>\n          baz\n        </MyComponent>\n      ",
		},
		{
			source:  "\n        <MyComponent>\n          {'foo'}\n          <div>\n            {'bar'}\n          </div>\n          {'baz'}\n          {'some-complicated-exp'}\n        </MyComponent>\n      ",
			options: "{\"children\":\"never\"}",
			want:    "\n        <MyComponent>\n          foo\n          <div>\n            bar\n          </div>\n          baz\n          some-complicated-exp\n        </MyComponent>\n      ",
		},
		{
			source:  "<MyComponent prop='bar'>foo</MyComponent>",
			options: "{\"props\":\"always\"}",
			want:    "<MyComponent prop={\"bar\"}>foo</MyComponent>",
		},
		{
			source:  "<MyComponent prop=\"foo 'bar'\">foo</MyComponent>",
			options: "{\"props\":\"always\"}",
			want:    "<MyComponent prop={\"foo 'bar'\"}>foo</MyComponent>",
		},
		{
			source:  "<MyComponent prop='foo \"bar\"'>foo</MyComponent>",
			options: "{\"props\":\"always\"}",
			want:    "<MyComponent prop={\"foo \\\"bar\\\"\"}>foo</MyComponent>",
		},
		{
			source:  "<MyComponent>foo bar </MyComponent>",
			options: "{\"children\":\"always\"}",
			want:    "<MyComponent>{\"foo bar \"}</MyComponent>",
		},
		{
			source:  "<MyComponent prop=\"foo 'bar' \\n \">foo</MyComponent>",
			options: "{\"props\":\"always\"}",
			want:    "<MyComponent prop={\"foo 'bar' \\\\n \"}>foo</MyComponent>",
		},
		{
			source:  "<MyComponent>foo bar \\r </MyComponent>",
			options: "{\"children\":\"always\"}",
			want:    "<MyComponent>{\"foo bar \\\\r \"}</MyComponent>",
		},
		{
			source:  "<MyComponent>foo bar 'foo'</MyComponent>",
			options: "{\"children\":\"always\"}",
			want:    "<MyComponent>{\"foo bar 'foo'\"}</MyComponent>",
		},
		{
			source:  "<MyComponent>foo bar \"foo\"</MyComponent>",
			options: "{\"children\":\"always\"}",
			want:    "<MyComponent>{\"foo bar \\\"foo\\\"\"}</MyComponent>",
		},
		{
			source:  "<MyComponent>foo bar <App/></MyComponent>",
			options: "{\"children\":\"always\"}",
			want:    "<MyComponent>{\"foo bar \"}<App/></MyComponent>",
		},
		{
			source:  "<MyComponent>foo \\n bar</MyComponent>",
			options: "{\"children\":\"always\"}",
			want:    "<MyComponent>{\"foo \\\\n bar\"}</MyComponent>",
		},
		{
			source:  "<MyComponent>foo \\u1234 bar</MyComponent>",
			options: "{\"children\":\"always\"}",
			want:    "<MyComponent>{\"foo \\\\u1234 bar\"}</MyComponent>",
		},
		{
			source:  "<MyComponent prop='foo \\u1234 bar' />",
			options: "{\"props\":\"always\"}",
			want:    "<MyComponent prop={\"foo \\\\u1234 bar\"} />",
		},
		{
			source:  "<MyComponent prop={'bar'}>{'foo'}</MyComponent>",
			options: "\"never\"",
			want:    "<MyComponent prop=\"bar\">foo</MyComponent>",
		},
		{
			source:  "<MyComponent prop='bar'>foo</MyComponent>",
			options: "\"always\"",
			want:    "<MyComponent prop={\"bar\"}>{\"foo\"}</MyComponent>",
		},
		{
			source:  "<App prop={'foo'} attr={\" foo \"} />",
			options: "{\"props\":\"never\"}",
			want:    "<App prop=\"foo\" attr=\" foo \" />",
		},
		{
			source:  "<App prop='foo' attr=\"bar\" />",
			options: "{\"props\":\"always\"}",
			want:    "<App prop={\"foo\"} attr={\"bar\"} />",
		},
		{
			source:  "<App prop='foo' attr={\"bar\"} />",
			options: "{\"props\":\"always\"}",
			want:    "<App prop={\"foo\"} attr={\"bar\"} />",
		},
		{
			source:  "<App prop={'foo'} attr='bar' />",
			options: "{\"props\":\"always\"}",
			want:    "<App prop={'foo'} attr={\"bar\"} />",
		},
		{
			source:  "<App prop='foo &middot; bar' />",
			options: "{\"props\":\"always\"}",
			want:    "<App prop={\"foo &middot; bar\"} />",
		},
		{
			source:  "<App>foo &middot; bar</App>",
			options: "{\"children\":\"always\"}",
			want:    "<App>{\"foo &middot; bar\"}</App>",
		},
		{
			source:  "<App>{'foo \"bar\"'}</App>",
			options: "{\"children\":\"never\"}",
			want:    "<App>foo \"bar\"</App>",
		},
		{
			source:  "<App>{\"foo 'bar'\"}</App>",
			options: "{\"children\":\"never\"}",
			want:    "<App>foo 'bar'</App>",
		},
		{
			source:  "\n        <App prop=\"    \n           a     \n             b      c\n                d\n        \">\n          a\n              b     c   \n                 d      \n        </App>\n      ",
			options: "\"always\"",
			want:    "\n        <App prop=\"    \n           a     \n             b      c\n                d\n        \">\n          {\"a\"}\n              {\"b     c   \"}\n                 {\"d      \"}\n        </App>\n      ",
		},
		{
			source:  "\n        <App prop='    \n           a     \n             b      c\n                d\n        '>\n          a\n              b     c   \n                 d      \n        </App>\n      ",
			options: "\"always\"",
			want:    "\n        <App prop='    \n           a     \n             b      c\n                d\n        '>\n          {\"a\"}\n              {\"b     c   \"}\n                 {\"d      \"}\n        </App>\n      ",
		},
		{
			source:  "\n        <App>\n          foo bar\n          <div>foo bar foo</div>\n          <span>\n            foo bar <i>foo bar</i>\n            <strong>\n              foo bar\n            </strong>\n          </span>\n        </App>\n      ",
			options: "{\"children\":\"always\"}",
			want:    "\n        <App>\n          {\"foo bar\"}\n          <div>{\"foo bar foo\"}</div>\n          <span>\n            {\"foo bar \"}<i>{\"foo bar\"}</i>\n            <strong>\n              {\"foo bar\"}\n            </strong>\n          </span>\n        </App>\n      ",
		},
		{
			source:  "\n        <App>\n          &lt;Component&gt;\n          &nbsp;<Component />&nbsp;\n          &nbsp;\n        </App>\n      ",
			options: "{\"children\":\"always\"}",
			want:    "\n        <App>\n          &lt;{\"Component\"}&gt;\n          &nbsp;<Component />&nbsp;\n          &nbsp;\n        </App>\n      ",
		},
		{
			source:  "\n        <Box mb={'1rem'} />\n      ",
			options: "{\"props\":\"never\"}",
			want:    "\n        <Box mb=\"1rem\" />\n      ",
		},
		{
			source:  "\n        <Box mb={'1rem {}'} />\n      ",
			options: "\"never\"",
			want:    "\n        <Box mb=\"1rem {}\" />\n      ",
		},
		{
			source:  "<MyComponent prop={\"{ style: true }\"}>bar</MyComponent>",
			options: "\"never\"",
			want:    "<MyComponent prop=\"{ style: true }\">bar</MyComponent>",
		},
		{
			source:  "<MyComponent prop={\"< style: true >\"}>foo</MyComponent>",
			options: "\"never\"",
			want:    "<MyComponent prop=\"< style: true >\">foo</MyComponent>",
		},
		{
			source:  "<App horror=<div /> />",
			options: "{\"props\":\"always\",\"children\":\"always\",\"propElementValues\":\"always\"}",
			want:    "<App horror={<div />} />",
		},
		{
			source:  "<App horror={<div />} />",
			options: "{\"props\":\"never\",\"children\":\"never\",\"propElementValues\":\"never\"}",
			want:    "<App horror=<div /> />",
		},
		{
			source:  "<Foo bar={\"'\"} />",
			options: "{\"props\":\"never\",\"children\":\"never\",\"propElementValues\":\"never\"}",
			want:    "<Foo bar=\"'\" />",
		},
		{
			source:  "\n        <Foo help={'The maximum time range for searches. (i.e. \"P30D\" for 30 days, \"PT24H\" for 24 hours)'} />\n      ",
			options: "\"never\"",
			want:    "\n        <Foo help='The maximum time range for searches. (i.e. \"P30D\" for 30 days, \"PT24H\" for 24 hours)' />\n      ",
		},
	} {
		t.Run(strconv.Itoa(index), func(t *testing.T) {
			t.Parallel()
			result := jsxCurlyBracePresenceRun(t, testCase.source, testCase.options)
			rule_testing.ExpectFixedSource(t, result, testCase.want)
		})
	}
}

// A finding has to point at the braces it is asking about, and no message-id fixture can see that.
//
// The span matters twice over on this rule: it is where the reader is sent, and it is the range the
// repair writes over. It is asserted by slicing the source with the finding's own range, which is
// the only form that catches a fix anchored one node away from the thing it describes.
func TestJsxCurlyBracePresenceSpansTheBraces(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		source  string
		options string
		want    string
	}{
		{source: `<App>{'foo'}</App>`, options: "", want: `{'foo'}`},
		{source: `<App prop={'bar'} />`, options: "", want: `{'bar'}`},
		{source: "<App>{`foo`}</App>", options: "", want: "{`foo`}"},
		{source: `<App>{<Foo />}</App>`, options: "", want: `{<Foo />}`},
		{source: `<App prop='bar' />`, options: `{"props":"always"}`, want: `'bar'`},
		{source: `<App>foo</App>`, options: `{"children":"always"}`, want: `foo`},
	} {
		result := jsxCurlyBracePresenceRun(t, testCase.source, testCase.options)
		if len(result.Diagnostics) != 1 {
			t.Fatalf("expected exactly one finding for %q, got %v", testCase.source, result.MessageIds())
		}
		reported := testCase.source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
		if reported != testCase.want {
			t.Errorf("for %q the finding spans %q, want %q", testCase.source, reported, testCase.want)
		}
	}
}

// The message text, asserted against literals typed here rather than against the rule's constants.
//
// Comparing a diagnostic to the very constant the rule reports with is equality that moves on both
// sides under mutation, so it proves nothing about what a reader sees.
func TestJsxCurlyBracePresenceMessages(t *testing.T) {
	t.Parallel()

	unnecessary := jsxCurlyBracePresenceRun(t, `<App>{'foo'}</App>`, "")
	rule_testing.ExpectFindings(t, unnecessary, "UnnecessaryCurly")
	if got := unnecessary.Diagnostics[0].Message.Id; got != "UnnecessaryCurly" {
		t.Errorf("id is %q", got)
	}
	if got := unnecessary.Diagnostics[0].Message.Description; !strings.HasPrefix(got,
		"These curly braces wrap a plain string and do nothing.") {
		t.Errorf("description is %q", got)
	}

	missing := jsxCurlyBracePresenceRun(t, `<App>foo</App>`, `{"children":"always"}`)
	rule_testing.ExpectFindings(t, missing, "MissingCurly")
	if got := missing.Diagnostics[0].Message.Id; got != "MissingCurly" {
		t.Errorf("id is %q", got)
	}
	if got := missing.Diagnostics[0].Message.Description; !strings.HasPrefix(got,
		"This literal is written as bare markup") {
		t.Errorf("description is %q", got)
	}
}

// Parentheses, which upstream's corpus cannot express.
//
// typescript-eslint folds a parenthesis away and ours keeps `KindParenthesizedExpression`, so
// `{('bar')}` reaches upstream's rule as a bare literal and reaches this one as a wrapper. Without
// the unwrap every shape here goes silent while upstream reports and repairs it, and no imported
// fixture can catch that, because the node does not exist in the tree the corpus was written
// against. Each expectation below was measured against the installed build, 7.37.5.
func TestJsxCurlyBracePresenceSeesThroughParentheses(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		source  string
		options string
		want    string
	}{
		{source: `<App>{('bar')}</App>`, options: "", want: `<App>bar</App>`},
		{source: `<App p={('bar')} />`, options: "", want: `<App p="bar" />`},
		{source: `<App>{(('bar'))}</App>`, options: "", want: `<App>bar</App>`},
	} {
		result := jsxCurlyBracePresenceRun(t, testCase.source, testCase.options)
		rule_testing.ExpectFindings(t, result, "UnnecessaryCurly")
		rule_testing.ExpectFixedSource(t, result, testCase.want)
	}
}

// TypeScript expressions the fixer must not touch, which upstream's corpus cannot express either.
//
// This is the class the port brief names as the highest risk on a fixer: a repair that is right
// about JavaScript and destroys a type. Each shape below wraps a string literal in something only
// TypeScript has, and unwrapping the braces would delete the annotation with them. They are safe
// here because each parses to its own node kind that no arm of the judgment matches, which is a
// property worth pinning rather than assuming: the alternative was a rule that reports
// `{'bar' as string}` and repairs it to `"bar"`. Measured against the installed build, which is
// silent on all four for the same reason.
func TestJsxCurlyBracePresenceLeavesTypeScriptExpressionsAlone(t *testing.T) {
	t.Parallel()

	for _, source := range []string{
		`<App p={'bar' as string} />`,
		`<App p={'bar'!} />`,
		`<App p={'bar' satisfies string} />`,
		`<App>{'bar' as string}</App>`,
		"<App p={`bar` as const} />",
	} {
		rule_testing.ExpectClean(t, jsxCurlyBracePresenceRun(t, source, ""))
	}
}

// The repairs upstream declines, reproduced as declines.
//
// Upstream's missing-curly fixer returns null rather than guessing, and a declined repair is part
// of what is being ported: a rule that repairs a case upstream refuses to touch is a defect no
// message-id fixture can see. Each row must REPORT and must carry no fix.
//
// Which shapes actually reach the decline was measured rather than read, and the reading would have
// been wrong. `containsOnlyHtmlEntities` runs on the literal's RAW text, which for an attribute
// includes its own surrounding quotes, so an attribute is never "only character references" and
// `<App p="&middot;" />` is repaired rather than declined. The reference arm is reachable only
// through a CHILD, and a child that is only references is already declined one level earlier by
// `shouldCheckForMissingCurly`, so it never reports at all. That leaves the line-terminator arm as
// the only decline this rule can produce, and the entity arm as reachable-but-subsumed. Both were
// driven through the installed build.
func TestJsxCurlyBracePresenceDeclinesSomeRepairs(t *testing.T) {
	t.Parallel()

	result := jsxCurlyBracePresenceRun(t, "<App p=\"a\nb\" />", `{"props":"always"}`)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("expected one finding, got %v", result.MessageIds())
	}
	if fixes := result.Diagnostics[0].Fixes; len(fixes) != 0 {
		t.Errorf("an attribute holding a line terminator must report without a repair, got %q", fixes[0].Text)
	}

	// The neighbouring shapes that DO carry a repair, so the decline above is shown to be narrow
	// rather than the fixer being broken for attributes generally.
	for _, testCase := range []struct{ source, want string }{
		{`<App p="&middot;" />`, `<App p={"&middot;"} />`},
		{`<App p="foo&#10;bar" />`, `<App p={"foo&#10;bar"} />`},
	} {
		repaired := jsxCurlyBracePresenceRun(t, testCase.source, `{"props":"always"}`)
		rule_testing.ExpectFindings(t, repaired, "MissingCurly")
		rule_testing.ExpectFixedSource(t, repaired, testCase.want)
	}

	// And the two shapes that are declined a level earlier, so they never report.
	rule_testing.ExpectClean(t, jsxCurlyBracePresenceRun(t, `<App>&middot;</App>`, `{"children":"always"}`))
	rule_testing.ExpectClean(t, jsxCurlyBracePresenceRun(t, "<App>\n</App>", `{"children":"always"}`))
}

// Shapes upstream's corpus does not write, each measured against the installed build first.
//
// A mutation sweep found each of these: the code discriminated on something no imported case could
// reach, so the branch survived being neutered. Every row below was run through 7.37.5 before it
// was written here, and the expectation is what that run produced rather than what the code does.
func TestJsxCurlyBracePresenceShapesTheCorpusDoesNotWrite(t *testing.T) {
	t.Parallel()

	// Character references. Upstream's corpus only ever writes the NAMED form, so the digits and
	// the hash in the reference grammar are unreached by it and a scanner accepting only letters
	// would pass every imported case while repairing source it must decline.
	for _, source := range []string{
		`<App>{'&#8212;'}</App>`,
		`<App p={'&#8212;'} />`,
		`<App>{'&#x2014;'}</App>`,
		`<App>{'&middot;'}</App>`,
	} {
		rule_testing.ExpectClean(t, jsxCurlyBracePresenceRun(t, source, ""))
	}

	// A non-breaking space is whitespace to this rule, which is why it cannot borrow the
	// `isJavaScriptWhitespace` predicate sitting beside it in this package: that one excludes
	// U+00A0 deliberately for a different rule. Built from its code point rather than typed, so
	// no editor or transfer can turn it into an ordinary space and make the case vacuous.
	nonBreakingSpace := string(rune(0x00a0))
	rule_testing.ExpectClean(t, jsxCurlyBracePresenceRun(t, "<App>{'"+nonBreakingSpace+"'}</App>", ""))
	rule_testing.ExpectClean(t, jsxCurlyBracePresenceRun(t, "<App p={'"+nonBreakingSpace+"'} />", ""))

	// A blank line inside a repaired multi-line child stays blank rather than being wrapped.
	blankLine := jsxCurlyBracePresenceRun(t, "<App>\n  a\n\n  b\n</App>", `{"children":"always"}`)
	rule_testing.ExpectFindings(t, blankLine, "MissingCurly")
	rule_testing.ExpectFixedSource(t, blankLine, "<App>\n  {\"a\"}\n\n  {\"b\"}\n</App>")
}

// The one rune where Go's idea of whitespace is wider than JavaScript's.
//
// A multi-line child is repaired line by line, and a line upstream considers blank is left as it is
// while any other line is wrapped. "Blank" there is `String.prototype.trim`, whose character set is
// the same as the regular expression's, and NOT `strings.TrimSpace`, whose set also contains U+0085.
// Written with the Go-native spelling, a line holding only that character was skipped and written
// back bare, where the installed build wraps it in an empty container. That is the whole difference
// between the two spellings and nothing in upstream's corpus can reach it.
//
// The expectation was taken from the installed build as raw bytes, so the container holds the
// character itself, rather than read off a terminal: this character prints as nothing and a lone
// one reads as an empty string. That misreading cost a round trip here, because `trim()` on it
// LOOKS empty and the comparison against the empty string is false.
func TestJsxCurlyBracePresenceUsesJavaScriptWhitespaceWhenTrimming(t *testing.T) {
	t.Parallel()

	nextLine := string(rune(0x0085))
	source := "<App>\n  a\n" + nextLine + "\n  b\n</App>"

	result := jsxCurlyBracePresenceRun(t, source, `{"children":"always"}`)
	rule_testing.ExpectFindings(t, result, "MissingCurly")
	rule_testing.ExpectFixedSource(t, result,
		"<App>\n  {\"a\"}\n{\""+nextLine+"\"}\n  {\"b\"}\n</App>")
}

// An empty string in braces is unnecessary in every position, and in an attribute it repairs to an
// empty attribute value. Upstream's whitespace-literal test reads an empty value as falsy, so it is
// not the whitespace an attribute keeps its braces for. The attribute arm once tested the value
// against `/^\s*$/` alone, which an empty string satisfies, and fell silent where ESLint reported
// two sites in www-phi-health (#t5dwy1t). Every row was run through 7.37.5 first, and the expected
// repair is what that run wrote.
func TestJsxCurlyBracePresenceReportsAnEmptyString(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		source string
		fixed  string
	}{
		{`<X label={''} />`, `<X label="" />`},
		{`<div className={""} />`, `<div className="" />`},
		{"<X label={``} />", `<X label="" />`},
		{`<div>{''}</div>`, `<div></div>`},
		{`<div>x{''}</div>`, `<div>x</div>`},
	} {
		result := jsxCurlyBracePresenceRun(t, testCase.source, "")
		rule_testing.ExpectFindings(t, result, "UnnecessaryCurly")
		rule_testing.ExpectFixedSource(t, result, testCase.fixed)
	}

	// The nearest shapes that keep their braces: a value that is only whitespace, and a template
	// holding an expression.
	rule_testing.ExpectClean(t, jsxCurlyBracePresenceRun(t, `<X label={' '} />`, ""))
	rule_testing.ExpectClean(t, jsxCurlyBracePresenceRun(t, "<X label={`${b}`} />", ""))
}
