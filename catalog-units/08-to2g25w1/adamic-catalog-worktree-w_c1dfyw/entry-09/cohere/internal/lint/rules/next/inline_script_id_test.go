package next

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// The eight passing cases from oxc's corpus, copied byte for byte out of the extractor's dump
// rather than transcribed. Two of them turn on spreads, which is the part of this rule a port is
// most likely to get wrong, and one of those passes for a reason its shape does not show: the
// opaque-spread case also carries an id, so it cannot isolate the abandonment it is really
// testing. The invented case below supplies that half.
func TestInlineScriptIdIsSilentOnUpstreamPassCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
	}{
		{
			name:   "an id attribute alongside children",
			source: "import Script from 'next/script';\n\n\t\t\t      export default function TestPage() {\n\t\t\t        return (\n\t\t\t          <Script id=\"test-script\">\n\t\t\t            {`console.log('Hello world');`}\n\t\t\t          </Script>\n\t\t\t        )\n\t\t\t      }",
		},
		{
			name:   "an id attribute alongside dangerouslySetInnerHTML, self closing",
			source: "import Script from 'next/script';\n\n\t\t\t      export default function TestPage() {\n\t\t\t        return (\n\t\t\t          <Script\n\t\t\t            id=\"test-script\"\n\t\t\t            dangerouslySetInnerHTML={{\n\t\t\t              __html: `console.log('Hello world');`\n\t\t\t            }}\n\t\t\t          />\n\t\t\t        )\n\t\t\t      }",
		},
		{
			name:   "an external script with a src and nothing inline",
			source: "import Script from 'next/script';\n\n\t\t\t      export default function TestPage() {\n\t\t\t        return (\n\t\t\t          <Script src=\"https://example.com\" />\n\t\t\t        )\n\t\t\t      }",
		},
		{
			name:   "the renamed default import MyScript with an id and children",
			source: "import MyScript from 'next/script';\n\n\t\t\t      export default function TestPage() {\n\t\t\t        return (\n\t\t\t          <MyScript id=\"test-script\">\n\t\t\t            {`console.log('Hello world');`}\n\t\t\t          </MyScript>\n\t\t\t        )\n\t\t\t      }",
		},
		{
			name:   "the renamed default import MyScript with an id and dangerouslySetInnerHTML",
			source: "import MyScript from 'next/script';\n\n\t\t\t      export default function TestPage() {\n\t\t\t        return (\n\t\t\t          <MyScript\n\t\t\t            id=\"test-script\"\n\t\t\t            dangerouslySetInnerHTML={{\n\t\t\t              __html: `console.log('Hello world');`\n\t\t\t            }}\n\t\t\t          />\n\t\t\t        )\n\t\t\t      }",
		},
		{
			name:   "an object spread that does not supply id, next to a real id written as an expression container",
			source: "import Script from 'next/script';\n\n\t\t\t      export default function TestPage() {\n\t\t\t        return (\n\t\t\t          <Script {...{ strategy: \"lazyOnload\" }} id={\"test-script\"}>\n\t\t\t            {`console.log('Hello world');`}\n\t\t\t          </Script>\n\t\t\t        )\n\t\t\t      }",
		},
		{
			name:   "an object spread that supplies the id entirely, which is the surprising case",
			source: "import Script from 'next/script';\n\n\t\t\t      export default function TestPage() {\n\t\t\t        return (\n\t\t\t          <Script {...{ strategy: \"lazyOnload\", id: \"test-script\" }}>\n\t\t\t            {`console.log('Hello world');`}\n\t\t\t          </Script>\n\t\t\t        )\n\t\t\t      }",
		},
		{
			name:   "an opaque spread, which abandons the element before the id it also carries is ever consulted",
			source: "import Script from 'next/script';\n\t\t\t      const spread = { strategy: \"lazyOnload\" }\n\t\t\t      export default function TestPage() {\n\t\t\t        return (\n\t\t\t          <Script {...spread} id={\"test-script\"}>\n\t\t\t            {`console.log('Hello world');`}\n\t\t\t          </Script>\n\t\t\t        )\n\t\t\t      }",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, InlineScriptId, "pages/index.tsx", testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// The four failing cases from oxc's corpus. The snapshot carries exactly four diagnostics against
// four inputs, so one finding each, and the extractor reports no discrepancy. The four are the two
// by two of children against dangerouslySetInnerHTML crossed with the plain and renamed imports.
func TestInlineScriptIdFiresOnUpstreamFailCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
	}{
		{
			name:   "children and no id",
			source: "import Script from 'next/script';\n\n\t\t\t        export default function TestPage() {\n\t\t\t          return (\n\t\t\t            <Script>\n\t\t\t              {`console.log('Hello world');`}\n\t\t\t            </Script>\n\t\t\t          )\n\t\t\t        }",
		},
		{
			name:   "dangerouslySetInnerHTML and no id, self closing",
			source: "import Script from 'next/script';\n\n\t\t\t        export default function TestPage() {\n\t\t\t          return (\n\t\t\t            <Script\n\t\t\t              dangerouslySetInnerHTML={{\n\t\t\t                __html: `console.log('Hello world');`\n\t\t\t              }}\n\t\t\t            />\n\t\t\t          )\n\t\t\t        }",
		},
		{
			name:   "the renamed default import MyScript with children and no id",
			source: "import MyScript from 'next/script';\n\n\t\t\t        export default function TestPage() {\n\t\t\t          return (\n\t\t\t            <MyScript>\n\t\t\t              {`console.log('Hello world');`}\n\t\t\t            </MyScript>\n\t\t\t          )\n\t\t\t        }",
		},
		{
			name:   "the renamed default import MyScript with dangerouslySetInnerHTML and no id, self closing",
			source: "import MyScript from 'next/script';\n\n\t\t\t        export default function TestPage() {\n\t\t\t          return (\n\t\t\t            <MyScript\n\t\t\t              dangerouslySetInnerHTML={{\n\t\t\t                __html: `console.log('Hello world');`\n\t\t\t              }}\n\t\t\t            />\n\t\t\t          )\n\t\t\t        }",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, InlineScriptId, "pages/index.tsx", testCase.source)
			rule_testing.ExpectFindings(t, result, "inlineScriptId")
		})
	}
}

// Cases upstream does not write, each measured against the release oxlint binary before it was
// written down. The corpus tests the import gate only broken open: all twelve of its cases import
// from next/script, so a port that skipped resolution entirely and matched the bare word Script
// would pass every one of them. The first six here are that missing half.
func TestInlineScriptIdIsSilentOnCasesUpstreamDoesNotWrite(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
	}{
		{
			name:   "a local component named Script with no import at all, which is the whole negative half the corpus never writes",
			source: "function Script(props) { return null }\nexport default function TestPage() {\n  return (<Script>{`console.log('x');`}</Script>)\n}\n",
		},
		{
			name:   "a default import from a different module, guarding the exact specifier compare",
			source: "import Script from 'other/script';\nexport default function TestPage() {\n  return (<Script>{`console.log('x');`}</Script>)\n}\n",
		},
		{
			name:   "a module whose name merely extends the specifier, which a substring compare would bind",
			source: "import Script from 'next/scripts';\nexport default function TestPage() {\n  return (<Script>{`console.log('x');`}</Script>)\n}\n",
		},
		{
			name:   "a named import, which does not bind here even though the sibling rule's upstream binds it",
			source: "import { X } from 'next/script';\nexport default function TestPage() {\n  return (<X>{`console.log('x');`}</X>)\n}\n",
		},
		{
			name:   "a namespace import, silent for the same reason as the named one",
			source: "import * as Script from 'next/script';\nexport default function TestPage() {\n  return (<Script>{`console.log('x');`}</Script>)\n}\n",
		},
		{
			name:   "a member expression tag, where the kind guard is load bearing because Text panics on a property access",
			source: "import Script from 'next/script';\nexport default function TestPage() {\n  return (<Script.Inner>{`console.log('x');`}</Script.Inner>)\n}\n",
		},
		{
			name:   "an opaque spread with children and no id, the half of vercel/next.js#34030 the corpus cannot isolate",
			source: "import Script from 'next/script';\nconst spread = { strategy: \"lazyOnload\" }\nexport default function TestPage() {\n  return (<Script {...spread}>{`console.log('x');`}</Script>)\n}\n",
		},
		{
			name:   "an opaque spread abandoning the element before dangerouslySetInnerHTML is ever consulted",
			source: "import Script from 'next/script';\nconst spread = { strategy: \"lazyOnload\" }\nexport default function TestPage() {\n  return (<Script {...spread} dangerouslySetInnerHTML={{__html:\"x\"}} />)\n}\n",
		},
		{
			name:   "a parenthesized object spread supplying the id, which upstream reaches through without_parentheses",
			source: "import Script from 'next/script';\nexport default function TestPage() {\n  return (<Script {...({ id: \"a\" })}>{`console.log('x');`}</Script>)\n}\n",
		},
		{
			name:   "a shorthand property in a spread supplying the id, whose key is a static identifier",
			source: "import Script from 'next/script';\nconst id = \"a\";\nexport default function TestPage() {\n  return (<Script {...{ id }}>{`console.log('x');`}</Script>)\n}\n",
		},
		{
			name:   "an element with zero children between its tags, which is not inline",
			source: "import Script from 'next/script';\nexport default function TestPage() {\n  return (<Script></Script>)\n}\n",
		},
		{
			name:   "an external script that is self closing, carrying neither trigger",
			source: "import Script from 'next/script';\nexport default function TestPage() {\n  return (<Script src=\"https://example.com\" />)\n}\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, InlineScriptId, "pages/index.tsx", testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// The reporting half of the same set. Four of these pin a shape a reader would expect to be
// exempt and upstream reports: a src attribute, whitespace-only children, and the two object keys
// that spell id without being a static identifier. Each was run against the release binary rather
// than reasoned about, because each reads as though it should fall the other way.
func TestInlineScriptIdFiresOnCasesUpstreamDoesNotWrite(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
	}{
		{
			name:   "an aliased default import used under its alias, which no string match on the word Script would catch",
			source: "import S from 'next/script';\nexport default function TestPage() {\n  return (<S>{`console.log('x');`}</S>)\n}\n",
		},
		{
			name:   "a default import written alongside a named one, where the default half still binds",
			source: "import Script, { X } from 'next/script';\nexport default function TestPage() {\n  return (<Script>{`console.log('x');`}</Script>)\n}\n",
		},
		{
			name:   "a src attribute, which exempts nothing despite what the rule's own documentation suggests",
			source: "import Script from 'next/script';\nexport default function TestPage() {\n  return (<Script src=\"https://example.com\">{`console.log('x');`}</Script>)\n}\n",
		},
		{
			name:   "children that are whitespace only, which ast.IsWhitespaceOnlyJsxText would wrongly silence",
			source: "import Script from 'next/script';\nexport default function TestPage() {\n  return (<Script>\n  </Script>)\n}\n",
		},
		{
			name:   "a string key in a spread, which upstream does not read because it is not a static identifier",
			source: "import Script from 'next/script';\nexport default function TestPage() {\n  return (<Script {...{ \"id\": \"a\" }}>{`console.log('x');`}</Script>)\n}\n",
		},
		{
			name:   "a computed key in a spread, silent for the same reason as the string key",
			source: "import Script from 'next/script';\nexport default function TestPage() {\n  return (<Script {...{ [\"id\"]: \"a\" }}>{`console.log('x');`}</Script>)\n}\n",
		},
		{
			name:   "a nested spread inside the spread object, which contributes no key at all",
			source: "import Script from 'next/script';\nconst o = { id: \"a\" };\nexport default function TestPage() {\n  return (<Script {...{ ...o }}>{`console.log('x');`}</Script>)\n}\n",
		},
		{
			name:   "a namespaced attribute name, which is not a JsxAttributeName Identifier upstream",
			source: "import Script from 'next/script';\nexport default function TestPage() {\n  return (<Script x:id=\"a\">{`console.log('x');`}</Script>)\n}\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, InlineScriptId, "pages/index.tsx", testCase.source)
			rule_testing.ExpectFindings(t, result, "inlineScriptId")
		})
	}
}

// The finding points at the tag name, not at the element and not at the opening element. The
// snapshot's caret spans the six characters of `Script` and the eight of `MyScript`, so the tag name
// is the only span consistent with both, and the eslint original reports the whole JsxElement, which
// makes this the fixture that records which of the two upstreams was ported. Message-id assertions
// stay green over either choice.
//
// The expected text is a literal typed here rather than a reference to the rule's own message
// constant. Comparing a finding against the constant it was reported with is an equality that both
// sides of a mutation move together, so it looks correct and guards nothing.
func TestInlineScriptIdPointsAtTheTagName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		source       string
		wantReported string
	}{
		{
			name:         "a paired element reports on the tag name and not on the element or its children",
			source:       "import Script from 'next/script';\nexport default function TestPage() {\n  return (<Script>{`console.log('x');`}</Script>)\n}\n",
			wantReported: "Script",
		},
		{
			name:         "a self closing element reports on the tag name and not on the attribute that armed it",
			source:       "import Script from 'next/script';\nexport default function TestPage() {\n  return (<Script dangerouslySetInnerHTML={{__html:\"x\"}} />)\n}\n",
			wantReported: "Script",
		},
		{
			name:         "the renamed import reports its own longer tag name, which a hardcoded span would get wrong",
			source:       "import MyScript from 'next/script';\nexport default function TestPage() {\n  return (<MyScript>{`console.log('x');`}</MyScript>)\n}\n",
			wantReported: "MyScript",
		},
		{
			name:         "an aliased import reports its short alias, pinning the span against the import rather than the word Script",
			source:       "import S from 'next/script';\nexport default function TestPage() {\n  return (<S>{`console.log('x');`}</S>)\n}\n",
			wantReported: "S",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, InlineScriptId, "pages/index.tsx", testCase.source)
			rule_testing.ExpectFindings(t, result, "inlineScriptId")

			diagnostic := result.Diagnostics[0]
			reported := testCase.source[diagnostic.Range.Pos():diagnostic.Range.End()]
			if reported != testCase.wantReported {
				t.Fatalf("reported span: want %q, got %q", testCase.wantReported, reported)
			}
		})
	}
}

// The message text asserted exactly rather than by substring. A predicate weaker than the property
// it guards is not a guard: a rule rendering a doubled word or a stray placeholder still contains
// any needle short enough to look right.
func TestInlineScriptIdRendersItsMessageExactly(t *testing.T) {
	t.Parallel()

	source := "import Script from 'next/script';\nexport default function TestPage() {\n  return (<Script>{`console.log('x');`}</Script>)\n}\n"
	result := rule_testing.Run(t, InlineScriptId, "pages/index.tsx", source)
	rule_testing.ExpectFindings(t, result, "inlineScriptId")

	wantMessage := "This renders a next/script component with inline content but no id attribute. " +
		"The framework deduplicates inline scripts by that id, so without one the same script is " +
		"re-injected on every navigation and any one-time setup it performs runs again, which shows " +
		"up as doubled analytics events or global state clobbered after a client transition. Give " +
		"the script a stable id attribute."
	if got := result.Diagnostics[0].Message.Description; got != wantMessage {
		t.Fatalf("message: want %q, got %q", wantMessage, got)
	}
}

// Two next/script default imports in one file both arm the rule, so two inline elements report
// twice. oxc anchors on each ImportDefaultSpecifier separately and walks that symbol's references;
// the eslint original keeps one module-scoped variable and would track only the second import, so
// this fixture records which upstream was ported. It is also the case that a single-name
// implementation passes every other fixture in this file while getting wrong.
func TestInlineScriptIdArmsOnEveryDefaultImportRatherThanTheLastOne(t *testing.T) {
	t.Parallel()

	source := "import Script from 'next/script';\nimport Other from 'next/script';\nexport default function TestPage() {\n  return (<div><Script>{`a`}</Script><Other>{`b`}</Other></div>)\n}\n"
	result := rule_testing.Run(t, InlineScriptId, "pages/index.tsx", source)
	rule_testing.ExpectFindings(t, result, "inlineScriptId", "inlineScriptId")

	firstReported := source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
	secondReported := source[result.Diagnostics[1].Range.Pos():result.Diagnostics[1].Range.End()]
	if firstReported != "Script" || secondReported != "Other" {
		t.Fatalf("reported spans: want \"Script\" and \"Other\", got %q and %q", firstReported, secondReported)
	}
}

// An import written textually after the JSX still arms the rule, probed against the release binary.
// oxc reads a whole-file module record that does not care about position; our walk is single pass
// and ordered, so a listener on KindImportDeclaration setting a variable would answer differently.
// This fixture is what fails if the scan is ever moved out of Run into a listener.
func TestInlineScriptIdSeesAnImportWrittenAfterTheJsx(t *testing.T) {
	t.Parallel()

	source := "export default function TestPage() {\n  return (<Script>{`console.log('x');`}</Script>)\n}\nimport Script from 'next/script';\n"
	result := rule_testing.Run(t, InlineScriptId, "pages/index.tsx", source)
	rule_testing.ExpectFindings(t, result, "inlineScriptId")
}

// A parenthesized object spread that does NOT supply an id, which is the only shape that separates
// reading through the parentheses from treating them as opaque.
//
// The obvious fixture for `without_parentheses` is a parenthesized spread that supplies the id, and
// it is written above, but it cannot see this: without the skip the spread is not an object literal,
// so the element is abandoned as opaque and stays silent for the wrong reason. Both paths agree on
// silence and a mutant dropping the skip survives. Here they disagree, because reading the object
// finds no id and lets the children arm the report while abandoning would not.
//
// Our parser keeps the parenthesis as a real KindParenthesizedExpression node where oxc's has
// already dropped it, probed rather than assumed, so the skip is load bearing here in a way the Rust
// does not make obvious.
func TestInlineScriptIdReadsThroughParenthesesRatherThanAbandoningTheElement(t *testing.T) {
	t.Parallel()

	source := "import Script from 'next/script';\nexport default function TestPage() {\n  return (<Script {...({ strategy: \"lazyOnload\" })}>{`console.log('x');`}</Script>)\n}\n"
	result := rule_testing.Run(t, InlineScriptId, "pages/index.tsx", source)
	rule_testing.ExpectFindings(t, result, "inlineScriptId")
}
