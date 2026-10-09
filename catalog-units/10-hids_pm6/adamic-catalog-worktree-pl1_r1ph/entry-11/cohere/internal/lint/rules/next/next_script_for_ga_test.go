package next

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// The imported corpus is four pass cases and five fail cases carrying five diagnostics, one per
// input. Every string below is upstream's, copied through the extractor's own parser and re-quoted
// mechanically rather than transcribed, because three of the fail cases contain template literals
// whose escapes a hand copy or a heredoc would cook.
//
// Two of the fail inputs contain a second <script> that does NOT report, which is the corpus'
// sharpest content and the thing a reading of the rule name gets backwards. Those are asserted with
// a single finding each rather than two.

func TestNextScriptForGaReports(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
	}{
		{
			// Two scripts. Only the first reports: its `src` names the gtag host. The second is an
			// inline gtag bootstrap, which contains neither of the two inline substrings, so it is
			// silent. A port that treated "inline Google Analytics code" as the subject would
			// report twice here and the snapshot says once.
			name:   "a gtag src beside a silent inline gtag snippet",
			source: "\n\t\t\t        export class Blah extends Head {\n\t\t\t          render() {\n\t\t\t            return (\n\t\t\t              <div>\n\t\t\t                <h1>Hello title</h1>\n\t\t\t                <script async src='https://www.googletagmanager.com/gtag/js?id=${GA_TRACKING_ID}' />\n\t\t\t                <script\n\t\t\t                  dangerouslySetInnerHTML={{\n\t\t\t                    __html: `\n\t\t\t                      window.dataLayer = window.dataLayer || [];\n\t\t\t                      function gtag(){dataLayer.push(arguments);}\n\t\t\t                      gtag('js', new Date());\n\t\t\t                      gtag('config', '${GA_TRACKING_ID}', {\n\t\t\t                        page_path: window.location.pathname,\n\t\t\t                      });\n\t\t\t                  `,\n\t\t\t                }}/>\n\t\t\t              </div>\n\t\t\t            );\n\t\t\t          }\n\t\t\t      }",
		},
		{
			// The Google Tag Manager inline bootstrap, which is the only thing `gtm.js` matches.
			// `gtm.js` is an inline-only substring: it is not in the src list, measured.
			name:   "an inline tag manager bootstrap",
			source: "\n\t\t\t        export class Blah extends Head {\n\t\t\t          render() {\n\t\t\t            return (\n\t\t\t              <div>\n\t\t\t                <h1>Hello title</h1> qqq\n\t\t\t                {/* Google Tag Manager - Global base code */}\n\t\t\t                <script\n\t\t\t                dangerouslySetInnerHTML={{\n\t\t\t                  __html: `\n\t\t\t                    (function(w,d,s,l,i){w[l]=w[l]||[];w[l].push({'gtm.start':\n\t\t\t                    new Date().getTime(),event:'gtm.js'});var f=d.getElementsByTagName(s)[0],\n\t\t\t                    j=d.createElement(s),dl=l!='dataLayer'?'&l='+l:'';j.async=true;j.src=\n\t\t\t                    'https://www.googletagmanager.com/gtm.js?id='+i+dl;f.parentNode.insertBefore(j,f);\n\t\t\t                    })(window,document,'script','dataLayer', '${GTM_ID}');\n\t\t\t                  `,\n\t\t\t                }}/>\n\t\t\t              </div>\n\t\t\t            );\n\t\t\t          }\n\t\t\t      }",
		},
		{
			// The classic analytics.js inline loader. `analytics.js` is the one substring that
			// appears in both lists.
			name:   "an inline analytics.js loader",
			source: "\n\t\t\t        export class Blah extends Head {\n\t\t\t          render() {\n\t\t\t            return (\n\t\t\t              <div>\n\t\t\t                <h1>Hello title</h1>\n\t\t\t                <script dangerouslySetInnerHTML={{\n\t\t\t                    __html: `\n\t\t\t                      (function(i,s,o,g,r,a,m){i['GoogleAnalyticsObject']=r;i[r]=i[r]||function(){\n\t\t\t                        (i[r].q=i[r].q||[]).push(arguments)},i[r].l=1*new Date();a=s.createElement(o),\n\t\t\t                        m=s.getElementsByTagName(o)[0];a.async=1;a.src=g;m.parentNode.insertBefore(a,m)\n\t\t\t                        })(window,document,'script','https://www.google-analytics.com/analytics.js','ga');\n\n\t\t\t                        ga('create', 'UA-XXXXX-Y', 'auto');\n\t\t\t                        ga('send', 'pageview');\n\t\t\t                    `,\n\t\t\t                  }}/>\n\t\t\t              </div>\n\t\t\t            );\n\t\t\t          }\n\t\t\t      }",
		},
		{
			// Two scripts again. The inline one is a bare `ga(...)` snippet naming no URL and is
			// silent; the reporting one is the `src` on the second element.
			name:   "a silent inline ga snippet beside an analytics.js src",
			source: "\n\t\t\t        export class Blah extends Head {\n\t\t\t          render() {\n\t\t\t            return (\n\t\t\t              <div>\n\t\t\t                <h1>Hello title</h1>\n\t\t\t                <script dangerouslySetInnerHTML={{\n\t\t\t                    __html: `\n\t\t\t                        window.ga=window.ga||function(){(ga.q=ga.q||[]).push(arguments)};ga.l=+new Date;\n\t\t\t                        ga('create', 'UA-XXXXX-Y', 'auto');\n\t\t\t                        ga('send', 'pageview');\n\t\t\t                    `,\n\t\t\t                  }}/>\n\t\t\t                <script async src='https://www.google-analytics.com/analytics.js'></script>\n\t\t\t              </div>\n\t\t\t            );\n\t\t\t          }\n\t\t\t      }",
		},
		{
			// The inline value here is a call expression rather than a template literal, so the
			// inline branch declines it outright. The finding is the `src` on the second element.
			name:   "a call expression inline value beside an analytics.js src",
			source: "\n\t\t\t        export class Blah extends Head {\n\t\t\t          createGoogleAnalyticsMarkup() {\n\t\t\t            return {\n\t\t\t              __html: `\n\t\t\t                window.dataLayer = window.dataLayer || [];\n\t\t\t                function gtag(){dataLayer.push(arguments);}\n\t\t\t                gtag('js', new Date());\n\t\t\t                gtag('config', 'UA-148481588-2');`,\n\t\t\t            };\n\t\t\t          }\n\n\t\t\t          render() {\n\t\t\t            return (\n\t\t\t              <div>\n\t\t\t                <h1>Hello title</h1>\n\t\t\t                <script dangerouslySetInnerHTML={this.createGoogleAnalyticsMarkup()} />\n\t\t\t                <script async src='https://www.google-analytics.com/analytics.js'></script>\n\t\t\t              </div>\n\t\t\t            );\n\t\t\t          }\n\t\t\t      }",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NextScriptForGa, "pages/index.tsx", testCase.source)
			rule_testing.ExpectFindings(t, result, "nextScriptForGa")
		})
	}
}

func TestNextScriptForGaIsSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
	}{
		{
			// The rule's whole point: the same two scripts written as next/script components. The
			// tag name is compared to the literal "script", so a capitalised component never
			// matches whatever its src or children say.
			name:   "next/script components carrying gtag src and inline code",
			source: "import Script from 'next/script'\n\n\t\t\t      export class Blah extends Head {\n\t\t\t        render() {\n\t\t\t          return (\n\t\t\t            <div>\n\t\t\t              <h1>Hello title</h1>\n\t\t\t              <Script\n\t\t\t                src=\"https://www.googletagmanager.com/gtag/js?id=GA_MEASUREMENT_ID\"\n\t\t\t                strategy=\"lazyOnload\"\n\t\t\t              />\n\t\t\t              <Script id=\"google-analytics\">\n\t\t\t                {`\n\t\t\t                  window.dataLayer = window.dataLayer || [];\n\t\t\t                  function gtag(){window.dataLayer.push(arguments);}\n\t\t\t                  gtag('js', new Date());\n\n\t\t\t                  gtag('config', 'GA_MEASUREMENT_ID');\n\t\t\t                `}\n\t\t\t              </Script>\n\t\t\t            </div>\n\t\t\t          );\n\t\t\t        }\n\t\t\t    }",
		},
		{
			name:   "a next/script component carrying the analytics.js bootstrap",
			source: "import Script from 'next/script'\n\n\t\t\t      export class Blah extends Head {\n\t\t\t        render() {\n\t\t\t          return (\n\t\t\t            <div>\n\t\t\t              <h1>Hello title</h1>\n\t\t\t              <Script id=\"google-analytics\">\n\t\t\t                {`(function(i,s,o,g,r,a,m){i['GoogleAnalyticsObject']=r;i[r]=i[r]||function(){\n\t\t\t                    (i[r].q=i[r].q||[]).push(arguments)},i[r].l=1*new Date();a=s.createElement(o),\n\t\t\t                    m=s.getElementsByTagName(o)[0];a.async=1;a.src=g;m.parentNode.insertBefore(a,m)\n\t\t\t                    })(window,document,'script','https://www.google-analytics.com/analytics.js','ga');\n\n\t\t\t                    ga('create', 'UA-XXXXX-Y', 'auto');\n\t\t\t                    ga('send', 'pageview');\n\t\t\t                })`}\n\t\t\t              </Script>\n\t\t\t            </div>\n\t\t\t          );\n\t\t\t        }\n\t\t\t    }",
		},
		{
			name:   "a next/script component carrying a bare ga snippet",
			source: "import Script from 'next/script'\n\n\t\t\t        export class Blah extends Head {\n\t\t\t        render() {\n\t\t\t            return (\n\t\t\t            <div>\n\t\t\t                <h1>Hello title</h1>\n\t\t\t                <Script id=\"google-analytics\">\n\t\t\t                    {`window.ga=window.ga||function(){(ga.q=ga.q||[]).push(arguments)};ga.l=+new Date;\n\t\t\t                    ga('create', 'UA-XXXXX-Y', 'auto');\n\t\t\t                    ga('send', 'pageview');\n\t\t\t                    })`}\n\t\t\t                </Script>\n\t\t\t            </div>\n\t\t\t            );\n\t\t\t        }\n\t\t\t    }",
		},
		{
			// An empty object literal has no __html property, so the find returns nothing. This is
			// the corpus' only pass case on a real <script> element and it guards the property
			// search rather than the tag test.
			name:   "an intrinsic script whose dangerouslySetInnerHTML is empty",
			source: "export class Blah extends Head {\n\t\t\t          render() {\n\t\t\t            return (\n\t\t\t              <div>\n\t\t\t                <h1>Hello title</h1>\n\t\t\t                <script dangerouslySetInnerHTML={{}} />\n\t\t\t              </div>\n\t\t\t            );\n\t\t\t          }\n\t\t\t      }",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NextScriptForGa, "pages/index.tsx", testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// The corpus tests every gate broken open and almost none broken shut, so everything below was
// measured against the release oxlint binary before being written, with `no-sync-scripts` enabled
// alongside as a control so that a silent probe could be told apart from a config that never loaded
// the plugin. Each case names the verdict it pins.

func TestNextScriptForGaReportsTheMeasuredEdgeCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
	}{
		{
			// The attribute NAME is matched case-insensitively even though the tag name and the
			// attribute VALUE are not. Upstream reaches both attributes through
			// `has_jsx_prop_ignore_case`.
			name:   "an uppercase src attribute name",
			source: "export const A = <script SRC=\"https://www.google-analytics.com/analytics.js\" />;\n",
		},
		{
			// Upstream spells its own target lowercase, so the leniency is what makes the inline
			// branch reachable at all rather than a nicety.
			name:   "an uppercase dangerouslySetInnerHTML attribute name",
			source: "export const A = <script DANGEROUSLYSETINNERHTML={{__html: `www.googletagmanager.com/gtm.js`}} />;\n",
		},
		{
			// Not a prefix and not a host: a plain substring search that never looks at the scheme.
			name:   "an http scheme rather than https",
			source: "export const A = <script src=\"http://www.google-analytics.com/analytics.js\" />;\n",
		},
		{
			// Only the first template chunk is searched, so a URL BEFORE the interpolation is
			// found. Its mirror image below is silent, and the pair is what pins the index.
			name:   "a url before an interpolation",
			source: "export const A = <script dangerouslySetInnerHTML={{__html: `www.google-analytics.com/analytics.js${x}tail`}} />;\n",
		},
		{
			// A src that matches nothing falls through to the inline test rather than ending the
			// element, which is the half of the branching a `return` in the wrong place loses.
			name:   "a non matching src falling through to a matching inline value",
			source: "export const A = <script src=\"/local.js\" dangerouslySetInnerHTML={{__html: `www.googletagmanager.com/gtm.js`}} />;\n",
		},
		{
			// A spread carries no readable name, so the src lookup finds nothing and the written
			// attribute still decides.
			name:   "a spread beside a matching inline value",
			source: "export const A = <script {...rest} dangerouslySetInnerHTML={{__html: `www.googletagmanager.com/gtm.js`}} />;\n",
		},
		{
			// A paired element rather than a self-closing one. Both JSX forms have to be listened
			// for, and three of the five corpus fail cases are self-closing.
			name:   "a paired script element",
			source: "export const A = <script dangerouslySetInnerHTML={{__html: `www.google-analytics.com/analytics.js`}}></script>;\n",
		},
		{
			// The `__html` key sitting after an unrelated one still wins; the search is a find
			// rather than a look at the first property.
			name:   "an __html key written second",
			source: "export const A = <script dangerouslySetInnerHTML={{other: 1, __html: `www.google-analytics.com/analytics.js`}} />;\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NextScriptForGa, "pages/index.tsx", testCase.source)
			rule_testing.ExpectFindings(t, result, "nextScriptForGa")
		})
	}
}

func TestNextScriptForGaIsSilentOnTheMeasuredEdgeCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
	}{
		{
			// The two substring lists are NOT the same list. `gtm.js` is inline-only, so a src
			// naming it is silent even though the identical string reports inline.
			name:   "the tag manager gtm.js path in a src",
			source: "export const A = <script src=\"https://www.googletagmanager.com/gtm.js?id=X\" />;\n",
		},
		{
			// And the mirror: `gtag/js` is src-only, so it is silent inline.
			name:   "the gtag js path inside an inline value",
			source: "export const A = <script dangerouslySetInnerHTML={{__html: `www.googletagmanager.com/gtag/js`}} />;\n",
		},
		{
			// The tag name is compared to the literal lowercase spelling, unlike both attribute
			// names on the same element.
			name:   "an uppercase script tag",
			source: "export const A = <SCRIPT src=\"https://www.google-analytics.com/analytics.js\" />;\n",
		},
		{
			// The VALUE is compared case-sensitively, which is the opposite of how its attribute
			// name was found two lines earlier in the same rule.
			name:   "an uppercase src value",
			source: "export const A = <script src=\"WWW.GOOGLE-ANALYTICS.COM/ANALYTICS.JS\" />;\n",
		},
		{
			// Only a string literal answers, so a src the linter would have to evaluate is not
			// guessed at.
			name:   "an expression valued src",
			source: "export const A = <script src={someUrl} />;\n",
		},
		{
			// A namespaced attribute name is not a JSXAttributeName::Identifier upstream, and
			// jsx.AttributeName already declines it.
			name:   "a namespaced src attribute",
			source: "export const A = <script x:src=\"https://www.google-analytics.com/analytics.js\" />;\n",
		},
		{
			// The intrinsic tag has to be `script`. Any other element carrying the same inline code
			// is out of scope.
			name:   "a div carrying the inline loader",
			source: "export const A = <div dangerouslySetInnerHTML={{__html: `www.google-analytics.com/analytics.js`}} />;\n",
		},
		{
			// Only a TemplateLiteral counts. A plain string holding the identical text declines,
			// which is the narrowness the corpus' fifth fail case exercises from the other side
			// with a call expression.
			name:   "a string literal inline value",
			source: "export const A = <script dangerouslySetInnerHTML={{__html: \"www.google-analytics.com/analytics.js\"}} />;\n",
		},
		{
			// Only quasis[0]. The mirror of the reporting case above: a URL AFTER an interpolation
			// sits in a later chunk that upstream never reads.
			name:   "a url after an interpolation",
			source: "export const A = <script dangerouslySetInnerHTML={{__html: `head${x}www.google-analytics.com/analytics.js`}} />;\n",
		},
		{
			// The RAW text is searched, not the cooked value. A unicode escape for the period
			// survives as six characters upstream and matches nothing. Reading Node.Text() here
			// would cook it to a period and report, which is the single most invisible way this
			// port could have diverged.
			name:   "an escaped period inside the inline url",
			source: "export const A = <script dangerouslySetInnerHTML={{__html: `www.google-analytics.com/analytics\\u002ejs`}} />;\n",
		},
		{
			// A string key is not a StaticIdentifier upstream, so it contributes nothing.
			name:   "a string __html key",
			source: "export const A = <script dangerouslySetInnerHTML={{\"__html\": `www.google-analytics.com/analytics.js`}} />;\n",
		},
		{
			// Nor is a computed one. ast.TryGetTextOfPropertyName resolves both of these shapes and
			// is deliberately not used, because using it would silence nothing and instead START
			// reporting two inputs upstream declines.
			name:   "a computed __html key",
			source: "export const A = <script dangerouslySetInnerHTML={{[\"__html\"]: `www.google-analytics.com/analytics.js`}} />;\n",
		},
		{
			// This rule does not call without_parentheses where its inline-script-id neighbour
			// does. Adding ast.SkipParentheses here reads as tidying and flips this verdict.
			name:   "a parenthesized object literal",
			source: "export const A = <script dangerouslySetInnerHTML={(({__html: `www.google-analytics.com/analytics.js`}))} />;\n",
		},
		{
			// A spread inside the object supplies no readable key.
			name:   "an object holding only a spread",
			source: "export const A = <script dangerouslySetInnerHTML={{...markup}} />;\n",
		},
		{
			// An expression container holding something other than an object literal declines
			// before any key is looked for.
			name:   "an identifier rather than an object literal",
			source: "export const A = <script dangerouslySetInnerHTML={markup} />;\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NextScriptForGa, "pages/index.tsx", testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// A matching src ends the element rather than adding a second finding, which is upstream's `return`
// inside the src branch. Without it this input reports twice and no fixture above can see it,
// because every other case carries at most one matching arm.
func TestNextScriptForGaReportsOnceWhenBothArmsMatch(t *testing.T) {
	t.Parallel()

	source := "export const A = <script src=\"https://www.google-analytics.com/analytics.js\" dangerouslySetInnerHTML={{__html: `www.googletagmanager.com/gtm.js`}} />;\n"
	result := rule_testing.Run(t, NextScriptForGa, "pages/index.tsx", source)
	rule_testing.ExpectFindings(t, result, "nextScriptForGa")
}

// The first `__html` wins and the search stops, matching upstream's find_map. A rule that kept
// looking would find the second value and report on an input that is silent upstream, measured
// against a duplicated key.
func TestNextScriptForGaReadsOnlyTheFirstHtmlKey(t *testing.T) {
	t.Parallel()

	source := "export const A = <script dangerouslySetInnerHTML={{__html: `harmless`, __html: `www.google-analytics.com/analytics.js`}} />;\n"
	result := rule_testing.Run(t, NextScriptForGa, "pages/index.tsx", source)
	rule_testing.ExpectClean(t, result)
}

// The finding points at the tag name rather than the whole element or the attribute, taken from the
// snapshot's caret, which underlines the six characters of `script`. Message ids cannot see where a
// finding points, so this is the only fixture that records the choice.
func TestNextScriptForGaPointsAtTheTagName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
	}{
		{
			name:   "the src arm",
			source: "export const A = <script src=\"https://www.google-analytics.com/analytics.js\" />;\n",
		},
		{
			name:   "the inline arm",
			source: "export const A = <script dangerouslySetInnerHTML={{__html: `www.googletagmanager.com/gtm.js`}} />;\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NextScriptForGa, "pages/index.tsx", testCase.source)
			rule_testing.ExpectFindings(t, result, "nextScriptForGa")

			diagnostic := result.Diagnostics[0]
			reported := testCase.source[diagnostic.Range.Pos():diagnostic.Range.End()]
			if reported != "script" {
				t.Fatalf("reported span: want %q, got %q", "script", reported)
			}
		})
	}
}

// The id and the description are asserted against literals typed here rather than against the
// rule's own message constant, because a comparison to the constant moves with any mutation of it
// and proves nothing. This rule interpolates nothing, so the description is fixed text and equality
// is the right predicate.
func TestNextScriptForGaMessage(t *testing.T) {
	t.Parallel()

	source := "export const A = <script src=\"https://www.google-analytics.com/analytics.js\" />;\n"
	result := rule_testing.Run(t, NextScriptForGa, "pages/index.tsx", source)
	rule_testing.ExpectFindings(t, result, "nextScriptForGa")

	wantDescription := "This loads Google Analytics through a plain <script> tag, which the " +
		"framework cannot schedule: it has no loading strategy, so it competes with hydration for " +
		"the main thread, and nothing deduplicates it across client navigations. Use the `Script` " +
		"component from `next/script`, which picks a strategy and injects the tag once."
	if messageNextScriptForGa.Description != wantDescription {
		t.Fatalf("description: want %q, got %q", wantDescription, messageNextScriptForGa.Description)
	}
	if messageNextScriptForGa.Id != "nextScriptForGa" {
		t.Fatalf("id: want %q, got %q", "nextScriptForGa", messageNextScriptForGa.Id)
	}
}

// The property-kind guard is crash protection rather than a behavioural filter, which is why a
// mutation widening it to accept shorthand survives every fixture above: no verdict can differ,
// because a shorthand's value is an identifier reference and the template test declines it anyway.
// What it prevents is a panic. `AsPropertyAssignment()` is an unchecked interface conversion and it
// dies on a `ShorthandPropertyAssignment` with "ast.nodeData is *ast.ShorthandPropertyAssignment,
// not *ast.PropertyAssignment", probed directly. A linter panic takes the whole run down rather than
// costing one finding, and no ExpectFindings fixture can see one, so this asserts on the shapes
// themselves. Upstream is silent on all of them, measured.
func TestNextScriptForGaSurvivesPropertyShapesItCannotConvert(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
	}{
		{
			name:   "a shorthand property",
			source: "export const A = <script dangerouslySetInnerHTML={{__html}} />;\n",
		},
		{
			name:   "a shorthand beside the real key",
			source: "export const A = <script dangerouslySetInnerHTML={{other, __html: `www.google-analytics.com/analytics.js`}} />;\n",
		},
		{
			name:   "a method",
			source: "export const A = <script dangerouslySetInnerHTML={{__html() { return 1; }}} />;\n",
		},
		{
			name:   "a getter",
			source: "export const A = <script dangerouslySetInnerHTML={{get __html() { return 1; }}} />;\n",
		},
		{
			name:   "a spread beside the real key",
			source: "export const A = <script dangerouslySetInnerHTML={{...rest, __html: `www.google-analytics.com/analytics.js`}} />;\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			// The assertion that matters is that Run returns at all. The verdicts are recorded
			// alongside because upstream is silent on the two that supply no readable value and
			// reports on the two where a real __html key survives an unconvertible property in
			// front of it. That last pair was measured after this fixture failed on a wrong
			// prediction of mine: I expected a leading shorthand to end the search, and both
			// upstream and this port keep looking, because find_map skips rather than stops.
			result := rule_testing.Run(t, NextScriptForGa, "pages/index.tsx", testCase.source)
			if testCase.name == "a spread beside the real key" ||
				testCase.name == "a shorthand beside the real key" {
				rule_testing.ExpectFindings(t, result, "nextScriptForGa")
				return
			}
			rule_testing.ExpectClean(t, result)
		})
	}
}

// The attribute search stops at the FIRST dangerouslySetInnerHTML and answers from it, rather than
// searching on for one that yields a usable value. Upstream's `has_jsx_prop_ignore_case` is a
// `find`, so a second copy of the attribute is never consulted whatever it holds.
//
// This needs its own fixture because a duplicated attribute is the only input on which stopping and
// continuing disagree, and nothing else in this file writes one. A mutation that continued the loop
// survived the entire suite before this was added. All three shapes were measured silent upstream,
// including the third, where the two spellings differ only in case and the case-insensitive matcher
// makes the first one the match.
func TestNextScriptForGaAnswersFromTheFirstDangerouslySetInnerHtml(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
	}{
		{
			name:   "an empty object in front of a matching one",
			source: "export const A = <script dangerouslySetInnerHTML={{}} dangerouslySetInnerHTML={{__html: `www.google-analytics.com/analytics.js`}} />;\n",
		},
		{
			name:   "a non object in front of a matching one",
			source: "export const A = <script dangerouslySetInnerHTML={markup} dangerouslySetInnerHTML={{__html: `www.google-analytics.com/analytics.js`}} />;\n",
		},
		{
			name:   "a differently cased spelling in front of a matching one",
			source: "export const A = <script dangerouslysetinnerhtml={{}} dangerouslySetInnerHTML={{__html: `www.google-analytics.com/analytics.js`}} />;\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NextScriptForGa, "pages/index.tsx", testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestNextScriptForGaDecodesEntities reads its attribute as ESLint does, with HTML entities decoded: typescript-estree decodes a
// JSX attribute string before any rule sees it. Each row's verdict is the installed
// @next/eslint-plugin-next 16.3.1's under the typescript-eslint parser (#51y9jh2).
func TestNextScriptForGaDecodesEntities(t *testing.T) {
	t.Parallel()

	for _, row := range []struct {
		name   string
		source string
		ids    []string
	}{
		{`an encoded slash in the analytics src`, `export const A = <script src="https://www.google-analytics.com&#47;analytics.js" />;`, []string{messageNextScriptForGa.Id}},
		{`an encoded slash in another src`, `export const A = <script src="https://www.example.com&#47;analytics.js" />;`, []string{}},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NextScriptForGa, "pages/index.tsx", row.source)
			if len(row.ids) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, row.ids...)
		})
	}
}
