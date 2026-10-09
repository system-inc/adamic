package next

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// The seven passing cases from oxc's corpus, copied byte for byte out of the extractor's dump rather
// than transcribed, and pinned by file name because on this rule the file name is most of the
// judgment. Ten of the thirteen upstream cases vary only the path over near-identical source.
func TestNoBeforeInteractiveScriptOutsideDocumentIsSilentOnUpstreamPassCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		fileName string
		source   string
	}{
		{
			name:     "the document with the default binding named Script",
			fileName: "pages/_document.js",
			source:   "import Document, { Html, Main, NextScript } from 'next/document'\n                  import Script from 'next/script'\n\n                  class MyDocument extends Document {\n                    render() {\n                      return (\n                        <Html>\n                          <Head>\n                            <meta charSet=\"utf-8\" />\n                          </Head>\n                          <body>\n                            <Main />\n                            <NextScript />\n                            <Script\n                              id=\"scriptBeforeInteractive\"\n                              src=\"https://cdnjs.cloudflare.com/ajax/libs/lodash.js/4.17.20/lodash.min.js?a=scriptBeforeInteractive\"\n                              strategy=\"beforeInteractive\"\n                            ></Script>\n                          </body>\n                        </Html>\n                      )\n                    }\n                  }\n\n                  export default MyDocument\n\t\t\t",
		},
		{
			name:     "the document with the import renamed to ScriptComponent, which is what proves the tag is matched against the import rather than against the word Script",
			fileName: "pages/_document.tsx",
			source:   "import Document, { Html, Main, NextScript } from 'next/document'\n                  import ScriptComponent from 'next/script'\n\n                  class MyDocument extends Document {\n                    render() {\n                      return (\n                        <Html>\n                          <Head>\n                            <meta charSet=\"utf-8\" />\n                          </Head>\n                          <body>\n                            <Main />\n                            <NextScript />\n                            <ScriptComponent\n                              id=\"scriptBeforeInteractive\"\n                              src=\"https://cdnjs.cloudflare.com/ajax/libs/lodash.js/4.17.20/lodash.min.js?a=scriptBeforeInteractive\"\n                              strategy=\"beforeInteractive\"\n                            ></ScriptComponent>\n                          </body>\n                        </Html>\n                      )\n                    }\n                  }\n\n                  export default MyDocument\n\t\t\t",
		},
		{
			name:     "the document with no strategy attribute at all",
			fileName: "pages/_document.tsx",
			source:   "import Document, { Html, Main, NextScript } from 'next/document'\n                  import ScriptComponent from 'next/script'\n\n                  class MyDocument extends Document {\n                    render() {\n                      return (\n                        <Html>\n                          <Head>\n                            <meta charSet=\"utf-8\" />\n                          </Head>\n                          <body>\n                            <Main />\n                            <NextScript />\n                            <ScriptComponent\n                              id=\"scriptBeforeInteractive\"\n                              src=\"https://cdnjs.cloudflare.com/ajax/libs/lodash.js/4.17.20/lodash.min.js?a=scriptBeforeInteractive\"\n                            ></ScriptComponent>\n                          </body>\n                        </Html>\n                      )\n                    }\n                  }\n\n                  export default MyDocument\n\t\t\t",
		},
		{
			name:     "an app directory layout, exempt before the tag name is even read",
			fileName: "/Users/user_name/projects/project-name/app/layout.tsx", // Not a machine path: upstream's own test input, a file the rule only reads by name
			source:   "import Script from \"next/script\";\n\n                  export default function Index() {\n                    return (\n                      <html lang=\"en\">\n                        <body className={inter.className}>{children}</body>\n                        <Script\n                          src=\"https://cdnjs.cloudflare.com/ajax/libs/lodash.js/4.17.20/lodash.min.js?a=scriptBeforeInteractive\"\n                          strategy='beforeInteractive'\n                        />\n                      </html>\n                    );\n                  }\n\t\t\t",
		},
		{
			name:     "the same app directory layout written with Windows separators",
			fileName: "C:\\Users\\username\\projects\\project-name\\app\\layout.tsx",
			source:   "import Script from \"next/script\";\n\n                  export default function test() {\n                    return (\n                      <html lang=\"en\">\n                        <body className={inter.className}>{children}</body>\n                        <Script\n                          src=\"https://cdnjs.cloudflare.com/ajax/libs/lodash.js/4.17.20/lodash.min.js?a=scriptBeforeInteractive\"\n                          strategy='beforeInteractive'\n                        />\n                      </html>\n                    );\n                  }\n\t\t\t",
		},
		{
			name:     "a src app directory layout",
			fileName: "/Users/user_name/projects/project-name/src/app/layout.tsx", // Not a machine path: upstream's own test input, a file the rule only reads by name
			source:   "import Script from \"next/script\";\n\n                  export default function Index() {\n                    return (\n                      <html lang=\"en\">\n                        <body className={inter.className}>{children}</body>\n                        <Script\n                          src=\"https://cdnjs.cloudflare.com/ajax/libs/lodash.js/4.17.20/lodash.min.js?a=scriptBeforeInteractive\"\n                          strategy='beforeInteractive'\n                        />\n                      </html>\n                    );\n                  }\n\t\t\t",
		},
		{
			name:     "the same src app directory layout written with Windows separators",
			fileName: "C:\\Users\\username\\projects\\project-name\\src\\app\\layout.tsx",
			source:   "import Script from \"next/script\";\n\n                  export default function test() {\n                    return (\n                      <html lang=\"en\">\n                        <body className={inter.className}>{children}</body>\n                        <Script\n                          src=\"https://cdnjs.cloudflare.com/ajax/libs/lodash.js/4.17.20/lodash.min.js?a=scriptBeforeInteractive\"\n                          strategy='beforeInteractive'\n                        />\n                      </html>\n                    );\n                  }\n\t\t\t",
		}}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoBeforeInteractiveScriptOutsideDocument, testCase.fileName, testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// The six failing cases from oxc's corpus. The snapshot carries exactly six diagnostics against six
// inputs, so one finding each, with no discrepancy for the extractor to warn about.
func TestNoBeforeInteractiveScriptOutsideDocumentFiresOnUpstreamFailCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		fileName string
		source   string
	}{
		{
			name:     "a page that is not the document",
			fileName: "pages/index.js",
			source:   "import Head from \"next/head\";\n                  import Script from \"next/script\";\n\n                  export default function Index() {\n                    return (\n                      <Script\n                        id=\"scriptBeforeInteractive\"\n                        src=\"https://cdnjs.cloudflare.com/ajax/libs/lodash.js/4.17.20/lodash.min.js?a=scriptBeforeInteractive\"\n                        strategy=\"beforeInteractive\"\n                      ></Script>\n                    );\n                  }\n\t\t\t",
		},
		{
			name:     "a file under neither pages nor app, which still reports",
			fileName: "components/outside-known-dirs.js",
			source:   " import Head from \"next/head\";\n             import Script from \"next/script\";\n\n             export default function Index() {\n               return (\n                 <Script\n                   id=\"scriptBeforeInteractive\"\n                   src=\"https://cdnjs.cloudflare.com/ajax/libs/lodash.js/4.17.20/lodash.min.js?a=scriptBeforeInteractive\"\n                   strategy=\"beforeInteractive\"\n                 ></Script>\n               );\n             }\n\t\t\t",
		},
		{
			name:     "a pages file that is not the document, absolute POSIX path",
			fileName: "/Users/user_name/projects/project-name/pages/layout.tsx", // Not a machine path: upstream's own test input, a file the rule only reads by name
			source:   " import Script from \"next/script\";\n\n                  export default function Index() {\n                    return (\n                      <html lang=\"en\">\n                        <body className={inter.className}>{children}</body>\n                        <Script\n                          src=\"https://cdnjs.cloudflare.com/ajax/libs/lodash.js/4.17.20/lodash.min.js?a=scriptBeforeInteractive\"\n                          strategy='beforeInteractive'\n                        />\n                      </html>\n                    );\n                  }\n\t\t\t",
		},
		{
			name:     "the same pages file written with Windows separators",
			fileName: "C:\\Users\\username\\projects\\project-name\\pages\\layout.tsx",
			source:   " import Script from \"next/script\";\n\n                  export default function Index() {\n                    return (\n                      <html lang=\"en\">\n                        <body className={inter.className}>{children}</body>\n                        <Script\n                          src=\"https://cdnjs.cloudflare.com/ajax/libs/lodash.js/4.17.20/lodash.min.js?a=scriptBeforeInteractive\"\n                          strategy='beforeInteractive'\n                        />\n                      </html>\n                    );\n                  }\n\t\t\t",
		},
		{
			name:     "a src pages file that is not the document",
			fileName: "/Users/user_name/projects/project-name/src/pages/layout.tsx", // Not a machine path: upstream's own test input, a file the rule only reads by name
			source:   " import Script from \"next/script\";\n\n                  export default function Index() {\n                    return (\n                      <html lang=\"en\">\n                        <body className={inter.className}>{children}</body>\n                        <Script\n                          src=\"https://cdnjs.cloudflare.com/ajax/libs/lodash.js/4.17.20/lodash.min.js?a=scriptBeforeInteractive\"\n                          strategy='beforeInteractive'\n                        />\n                      </html>\n                    );\n                  }\n\t\t\t",
		},
		{
			name:     "the same src pages file written with Windows separators",
			fileName: "C:\\Users\\username\\projects\\project-name\\src\\pages\\layout.tsx",
			source:   " import Script from \"next/script\";\n\n                  export default function test() {\n                    return (\n                      <html lang=\"en\">\n                        <body className={inter.className}>{children}</body>\n                        <Script\n                          src=\"https://cdnjs.cloudflare.com/ajax/libs/lodash.js/4.17.20/lodash.min.js?a=scriptBeforeInteractive\"\n                          strategy='beforeInteractive'\n                        />\n                      </html>\n                    );\n                  }\n\t\t\t",
		}}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoBeforeInteractiveScriptOutsideDocument, testCase.fileName, testCase.source)
			rule_testing.ExpectFindings(t, result, "noBeforeInteractiveScriptOutsideDocument")
		})
	}
}

// The finding points at the `strategy` attribute node rather than at the element, which is a real
// disagreement between the two upstreams: oxc labels `strategy.span` and the eslint original reports
// the whole opening element. The snapshot's caret spans twenty eight characters under
// `strategy="beforeInteractive"`, so the span is asserted by slicing the source with the finding's
// own range and comparing the text. A message id assertion cannot see where a finding points, and
// this rule carries no fix, so the span is the only thing that can.
func TestNoBeforeInteractiveScriptOutsideDocumentPointsAtTheStrategyAttribute(t *testing.T) {
	t.Parallel()

	source := "import Script from \"next/script\";\nexport default function Index() {\n  return <Script src=\"/a.js\" strategy=\"beforeInteractive\" />;\n}\n"

	result := rule_testing.Run(t, NoBeforeInteractiveScriptOutsideDocument, "pages/index.tsx", source)
	rule_testing.ExpectFindings(t, result, "noBeforeInteractiveScriptOutsideDocument")

	diagnostic := result.Diagnostics[0]
	reported := source[diagnostic.Range.Pos():diagnostic.Range.End()]
	// Compared against a literal typed here rather than against the rule's own constant, so the
	// assertion cannot move together with the thing it guards.
	if reported != "strategy=\"beforeInteractive\"" {
		t.Fatalf("reported range: want %q, got %q", "strategy=\"beforeInteractive\"", reported)
	}
	if length := diagnostic.Range.End() - diagnostic.Range.Pos(); length != 28 {
		t.Fatalf("span length: want 28 as the snapshot's caret shows, got %d", length)
	}
}

// The rendered message is asserted whole rather than by substring, because a predicate weaker than
// the property it guards is not a guard: a `strings.Contains` on an interpolated value stays green
// while the text around the needle is wrong.
func TestNoBeforeInteractiveScriptOutsideDocumentRendersItsMessage(t *testing.T) {
	t.Parallel()

	source := "import Script from \"next/script\";\nexport default function Index() {\n  return <Script strategy=\"beforeInteractive\" />;\n}\n"

	result := rule_testing.Run(t, NoBeforeInteractiveScriptOutsideDocument, "pages/index.tsx", source)
	rule_testing.ExpectFindings(t, result, "noBeforeInteractiveScriptOutsideDocument")

	got := result.Diagnostics[0].Message
	if got.Id != "noBeforeInteractiveScriptOutsideDocument" {
		t.Fatalf("message id: want %q, got %q", "noBeforeInteractiveScriptOutsideDocument", got.Id)
	}
	// The description is built by concatenating string fragments, which is exactly where a missing
	// or doubled space hides, so both ends are asserted against literals typed here.
	const wantPrefix = "This loads a script with next/script's beforeInteractive strategy outside of pages/_document.js."
	if !strings.HasPrefix(got.Description, wantPrefix) {
		t.Fatalf("description: want the prefix %q, got %q", wantPrefix, got.Description)
	}
	if !strings.HasSuffix(got.Description, "or pick a strategy that does not promise to run first.") {
		t.Fatalf("description: unexpected ending in %q", got.Description)
	}
	if strings.Contains(got.Description, "  ") {
		t.Fatalf("description has a doubled space, which is how a concatenated string goes wrong: %q", got.Description)
	}
}

// Cases upstream does not write, each one measured against the release oxlint binary rather than
// reasoned about, because this corpus varies the path and leaves almost everything else untested.
//
// The largest hole it leaves is the strategy comparison itself. Upstream's only case without a
// `beforeInteractive` strategy sits in `pages/_document.tsx`, where the path gate has already
// exempted the file, so deleting the strategy comparison entirely would leave every upstream fixture
// green. The first four cases here close that.
func TestNoBeforeInteractiveScriptOutsideDocumentIsSilentOnCasesUpstreamDoesNotWrite(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		fileName string
		source   string
	}{
		{
			// Upstream: 0 findings. The single most important invented case: without it the whole
			// strategy comparison is dead code that no fixture can see.
			name:     "a different strategy in a position that would otherwise report",
			fileName: "pages/index.tsx",
			source:   "import Script from \"next/script\";\nexport default function Index() {\n  return <Script strategy=\"afterInteractive\" />;\n}\n",
		},
		{
			name:     "the third real strategy value, also in a reporting position",
			fileName: "pages/index.tsx",
			source:   "import Script from \"next/script\";\nexport default function Index() {\n  return <Script strategy=\"lazyOnload\" />;\n}\n",
		},
		{
			name:     "no strategy attribute at all, outside the document",
			fileName: "pages/index.tsx",
			source:   "import Script from \"next/script\";\nexport default function Index() {\n  return <Script src=\"/a.js\" />;\n}\n",
		},
		{
			name:     "no attributes at all, which upstream short circuits on before reading a name",
			fileName: "pages/index.tsx",
			source:   "import Script from \"next/script\";\nexport default function Index() {\n  return <Script />;\n}\n",
		},
		{
			// Only a string literal counts. Upstream: 0 findings.
			name:     "the strategy written as an expression container rather than a string literal",
			fileName: "pages/index.tsx",
			source:   "import Script from \"next/script\";\nexport default function Index() {\n  return <Script strategy={\"beforeInteractive\"} />;\n}\n",
		},
		{
			// A spread is a real KindJsxSpreadAttribute node in this AST, so it is skipped rather
			// than unseen, and skipping it is what upstream's find over JSXAttributeItem does.
			name:     "a spread supplying the strategy inline",
			fileName: "pages/index.tsx",
			source:   "import Script from \"next/script\";\nexport default function Index() {\n  return <Script {...{ strategy: \"beforeInteractive\" }} />;\n}\n",
		},
		{
			name:     "a spread supplying the strategy through a variable",
			fileName: "pages/index.tsx",
			source:   "import Script from \"next/script\";\nconst properties = { strategy: \"beforeInteractive\" };\nexport default function Index() {\n  return <Script {...properties} />;\n}\n",
		},
		{
			// Upstream's `find` stops at the FIRST attribute named strategy and reads its value,
			// so a second one is never consulted and a non-matching first value ends the judgment.
			// Added for a surviving mutant that continued the loop instead of returning: nothing in
			// the corpus writes a duplicate attribute, so every fixture was blind to it. Measured on
			// the release binary: this shape is 0 findings and the reversed order is 1.
			name:     "a non-matching strategy written before a matching one",
			fileName: "pages/index.tsx",
			source:   "import Script from \"next/script\";\nexport default function Index() {\n  return <Script strategy=\"afterInteractive\" strategy=\"beforeInteractive\" />;\n}\n",
		},
		{
			name:     "an expression container strategy written before a matching literal one",
			fileName: "pages/index.tsx",
			source:   "import Script from \"next/script\";\nexport default function Index() {\n  return <Script strategy={x} strategy=\"beforeInteractive\" />;\n}\n",
		},
		{
			// The attribute name is compared exactly, not case insensitively.
			name:     "a capitalized Strategy attribute",
			fileName: "pages/index.tsx",
			source:   "import Script from \"next/script\";\nexport default function Index() {\n  return <Script Strategy=\"beforeInteractive\" />;\n}\n",
		},
		{
			// The import is string matched rather than resolved, so a local component wearing the
			// name is silent whenever the real import bound a different one. This looks like a
			// defect and reproducing it is the port; `imports.LocalNameOfDefaultImport` would flip
			// it and its doc comment invites exactly that.
			name:     "a local component named Script while the real import is bound to something else",
			fileName: "pages/index.tsx",
			source:   "import S from \"next/script\";\nfunction Script() {\n  return null;\n}\nexport default function Index() {\n  return <Script strategy=\"beforeInteractive\" />;\n}\n",
		},
		{
			name:     "a component named Script with no next/script import in the file at all",
			fileName: "pages/index.tsx",
			source:   "function Script() {\n  return null;\n}\nexport default function Index() {\n  return <Script strategy=\"beforeInteractive\" />;\n}\n",
		},
		{
			// The specifier is compared exactly rather than by substring. Added for a surviving
			// mutant that relaxed the comparison to a Contains: nothing in the corpus writes a
			// neighbouring module name, so the whole fixture set was blind to it. Measured against
			// the release binary at 0 findings for both directions of the substring.
			name:     "a module whose name merely extends next/script",
			fileName: "pages/index.tsx",
			source:   "import Script from \"next/scripts\";\nexport default function Index() {\n  return <Script strategy=\"beforeInteractive\" />;\n}\n",
		},
		{
			name:     "a module whose name merely ends with next/script",
			fileName: "pages/index.tsx",
			source:   "import Script from \"ext/script\";\nexport default function Index() {\n  return <Script strategy=\"beforeInteractive\" />;\n}\n",
		},
		{
			// A side-effect import introduces no entry, so nothing is bound.
			name:     "a side-effect import of next/script binding no name",
			fileName: "pages/index.tsx",
			source:   "import \"next/script\";\nexport default function Index() {\n  return <Script strategy=\"beforeInteractive\" />;\n}\n",
		},
		{
			// get_identifier_name declines anything that is not a plain identifier, and reading text
			// off the property access this parses to would panic.
			name:     "a member expression tag",
			fileName: "pages/index.tsx",
			source:   "import Script from \"next/script\";\nexport default function Index() {\n  return <Script.Inner strategy=\"beforeInteractive\" />;\n}\n",
		},
		{
			// Only the FIRST next/script entry binds, matching upstream's find_map, so the second
			// import's name is not a script name at all.
			name:     "the second of two next/script imports, whose name never binds",
			fileName: "pages/index.tsx",
			source:   "import A from \"next/script\";\nimport B from \"next/script\";\nexport default function Index() {\n  return <B strategy=\"beforeInteractive\" />;\n}\n",
		},
		{
			// The LOCAL name is compared. Reading the imported name instead would invert this and
			// the reporting case below it.
			name:     "a renamed named import addressed by its imported name rather than its local one",
			fileName: "pages/index.tsx",
			source:   "import { Foo as Bar } from \"next/script\";\nexport default function Index() {\n  return <Foo strategy=\"beforeInteractive\" />;\n}\n",
		},
		{
			// The substring app gate, faithfully loose: this path contains `app/` inside `myapp/`
			// and is exempt for that reason alone. Measured: upstream reports nothing here.
			name:     "a directory whose name merely ends in the word app",
			fileName: "myapp/pages/index.tsx",
			source:   "import Script from \"next/script\";\nexport default function Index() {\n  return <Script strategy=\"beforeInteractive\" />;\n}\n",
		},
		{
			// Pins the document predicate in the direction where the union predicate would report.
			// `is_document_page` tests a prefix with no trailing dot, so this is exempt upstream,
			// and `nextjs.IsDocumentFile` requires the dot and would report it. Measured: 0.
			name:     "a file merely beginning with the word document, which upstream exempts for want of a trailing dot",
			fileName: "pages/_documentation.tsx",
			source:   "import Script from \"next/script\";\nexport default function Index() {\n  return <Script strategy=\"beforeInteractive\" />;\n}\n",
		},
		{
			name:     "the document directory form, which the split on pages accepts",
			fileName: "pages/_document/index.tsx",
			source:   "import Script from \"next/script\";\nexport default function Index() {\n  return <Script strategy=\"beforeInteractive\" />;\n}\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoBeforeInteractiveScriptOutsideDocument, testCase.fileName, testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// The reporting half of the invented cases, measured the same way.
func TestNoBeforeInteractiveScriptOutsideDocumentFiresOnCasesUpstreamDoesNotWrite(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		fileName string
		source   string
	}{
		{
			// A named import binds, which `imports.LocalNameOfDefaultImport` would not see at all.
			name:     "a named import from next/script used as the tag",
			fileName: "pages/index.tsx",
			source:   "import { Foo } from \"next/script\";\nexport default function Index() {\n  return <Foo strategy=\"beforeInteractive\" />;\n}\n",
		},
		{
			name:     "a namespace import from next/script used as the tag",
			fileName: "pages/index.tsx",
			source:   "import * as All from \"next/script\";\nexport default function Index() {\n  return <All strategy=\"beforeInteractive\" />;\n}\n",
		},
		{
			name:     "a renamed named import addressed by its local name",
			fileName: "pages/index.tsx",
			source:   "import { Foo as Bar } from \"next/script\";\nexport default function Index() {\n  return <Bar strategy=\"beforeInteractive\" />;\n}\n",
		},
		{
			// Upstream reads a whole-file module record, so textual order does not matter. Our scan
			// over the statements is what reproduces that; a listener setting a variable would not.
			name:     "the import written textually after the component that uses it",
			fileName: "pages/index.tsx",
			source:   "export default function Index() {\n  return <Script strategy=\"beforeInteractive\" />;\n}\nimport Script from \"next/script\";\n",
		},
		{
			// A spread alongside a real attribute: the spread is passed over and the literal still
			// decides, which is the same skip as the silent spread cases seen from the other side.
			name:     "a spread alongside a real strategy literal",
			fileName: "pages/index.tsx",
			source:   "import Script from \"next/script\";\nexport default function Index() {\n  return <Script {...properties} strategy=\"beforeInteractive\" />;\n}\n",
		},
		{
			// Pins the document predicate in the direction where the union predicate would silence.
			// A `_document.tsx` that is not the immediate child of a pages directory is not the
			// document to `is_document_page`, and `nextjs.IsDocumentFile` would exempt it.
			name:     "a document named file outside any pages directory, which the union predicate would exempt",
			fileName: "components/_document.tsx",
			source:   "import Script from \"next/script\";\nexport default function Index() {\n  return <Script strategy=\"beforeInteractive\" />;\n}\n",
		},
		{
			name:     "a document named file nested below pages rather than directly under it",
			fileName: "src/pages/user/_document.tsx",
			source:   "import Script from \"next/script\";\nexport default function Index() {\n  return <Script strategy=\"beforeInteractive\" />;\n}\n",
		},
		{
			// The mirror of the duplicate-attribute case above: the first strategy matches, so the
			// second is never read. Measured at 1 finding.
			name:     "a matching strategy written before a non-matching one",
			fileName: "pages/index.tsx",
			source:   "import Script from \"next/script\";\nexport default function Index() {\n  return <Script strategy=\"beforeInteractive\" strategy=\"afterInteractive\" />;\n}\n",
		},
		{
			// The paired form. Four of upstream's six fail cases are self closing and two are
			// paired, so both kinds are covered upstream, but our tree's one real instance writes
			// both in one file and that is the reason both are registered.
			name:     "the paired element form rather than the self closing one",
			fileName: "pages/index.tsx",
			source:   "import Script from \"next/script\";\nexport default function Index() {\n  return <Script strategy=\"beforeInteractive\"></Script>;\n}\n",
		},
		{
			// An entity in the value. ESLint's parser decodes a JSX attribute string, so the installed
			// plugin (16.3.1) reads beforeInteractive here and reports (#5f61r5b).
			name:     "a strategy spelled with a character reference",
			fileName: "pages/index.tsx",
			source:   "import Script from \"next/script\";\nexport default function Index() {\n  return <Script strategy=\"before&#73;nteractive\" />;\n}\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoBeforeInteractiveScriptOutsideDocument, testCase.fileName, testCase.source)
			rule_testing.ExpectFindings(t, result, "noBeforeInteractiveScriptOutsideDocument")
		})
	}
}

// Two elements in one file report twice, which is our own tree's real shape: the single file in
// these checkouts that trips this rule writes a self closing Script and a paired one, and upstream
// reports both. Measured on the binary at two findings.
func TestNoBeforeInteractiveScriptOutsideDocumentReportsEveryElement(t *testing.T) {
	t.Parallel()

	source := "import Script from \"next/script\";\nexport default function Index() {\n  return (\n    <>\n      <Script src=\"/a.js\" strategy=\"beforeInteractive\" />\n      <Script id=\"b\" strategy=\"beforeInteractive\"></Script>\n    </>\n  );\n}\n"

	result := rule_testing.Run(t, NoBeforeInteractiveScriptOutsideDocument, "pages/index.tsx", source)
	rule_testing.ExpectFindings(t, result,
		"noBeforeInteractiveScriptOutsideDocument",
		"noBeforeInteractiveScriptOutsideDocument")
}
