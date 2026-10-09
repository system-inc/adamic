package next

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// Upstream ships one tester block: two passing cases and five failing ones, and the extractor
// counted five diagnostics against five failing inputs, so exactly one finding per input and no
// discrepancy to recover. Every source string below that is marked as upstream was pulled out of
// the Rust with the extractor and then checked byte for byte against the file, rather than
// transcribed, because the corpus is the only thing in this port that did not come from a belief
// about the rule.
//
// Upstream's two identical failing cases are carried once. They are byte-identical source at
// byte-identical paths, checked by hashing rather than by reading, so the second is a duplicated
// fixture rather than a distinction a port could miss.

// Every case here reports exactly once.
func TestNoHeadImportInDocumentReports(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		fileName string
		source   string
	}{
		{
			// Upstream. The plain document.
			name:     "the document as javascript",
			fileName: "pages/_document.js",
			source:   "\n\t\t\t      import Document, { Html, Main, NextScript } from 'next/document'\n\t\t\t      import Head from 'next/head'\n\n\t\t\t      class MyDocument extends Document {\n\t\t\t        render() {\n\t\t\t          return (\n\t\t\t            <Html>\n\t\t\t              <Head />\n\t\t\t              <body>\n\t\t\t                <Main />\n\t\t\t                <NextScript />\n\t\t\t              </body>\n\t\t\t            </Html>\n\t\t\t          )\n\t\t\t        }\n\t\t\t      }\n\n\t\t\t      export default MyDocument\n\t\t\t      ",
		},
		{
			// Upstream. The same document in TypeScript.
			name:     "the document as typescript",
			fileName: "pages/_document.tsx",
			source:   "\n\t\t\t      import Document, { Html, Main, NextScript } from 'next/document'\n\t\t\t      import Head from 'next/head'\n\n\t\t\t      class MyDocument extends Document {\n\t\t\t        render() {\n\t\t\t          return (\n\t\t\t            <Html>\n\t\t\t              <Head />\n\t\t\t              <body>\n\t\t\t                <Main />\n\t\t\t                <NextScript />\n\t\t\t              </body>\n\t\t\t            </Html>\n\t\t\t          )\n\t\t\t        }\n\t\t\t      }\n\n\t\t\t      export default MyDocument\n\t\t\t      ",
		},
		{
			// Upstream. The page-suffix convention, and the only upstream fixture where the trailing dot in `_document.` does real work: a port comparing the basename for equality passes every other case and fails this one.
			name:     "the page suffix convention",
			fileName: "pages/_document.page.tsx",
			source:   "\n\t\t\t      import Document, { Html, Main, NextScript } from 'next/document'\n\t\t\t      import Head from 'next/head'\n\n\t\t\t      class MyDocument extends Document {\n\t\t\t        render() {\n\t\t\t          return (\n\t\t\t            <Html>\n\t\t\t              <Head />\n\t\t\t              <body>\n\t\t\t                <Main />\n\t\t\t                <NextScript />\n\t\t\t              </body>\n\t\t\t            </Html>\n\t\t\t          )\n\t\t\t        }\n\t\t\t      }\n\n\t\t\t      export default MyDocument\n\t\t\t      ",
		},
		{
			// Upstream. The directory form, which only this rule's own gate spelling accepts. Upstream lists it twice, byte-identically, and the duplicate is dropped here.
			name:     "the directory form",
			fileName: "pages/_document/index.tsx",
			source:   "\n\t\t\t      import Document, { Html, Main, NextScript } from 'next/document'\n\t\t\t      import Head from 'next/head'\n\n\t\t\t      class MyDocument extends Document {\n\t\t\t        render() {\n\t\t\t          return (\n\t\t\t            <Html>\n\t\t\t              <Head />\n\t\t\t              <body>\n\t\t\t                <Main />\n\t\t\t                <NextScript />\n\t\t\t              </body>\n\t\t\t            </Html>\n\t\t\t          )\n\t\t\t        }\n\t\t\t      }\n\n\t\t\t      export default MyDocument\n\t\t\t      ",
		},
		{
			// Ours. Upstream writes the directory form only as `index.tsx`, so nothing there pins
			// that the gate reads a prefix of the basename rather than that one name.
			name:     "the directory form with another extension",
			fileName: "pages/_document/index.jsx",
			source:   "import Head from 'next/head';\n",
		},
		{
			// Ours, and counter-intuitive. Upstream's directory arm asks
			// `file_name.starts_with("index")` rather than testing equality, so a file merely
			// beginning with the word is the document. ESLint requires `name === 'index'` and
			// exempts this. Measured reporting on the release oxlint binary before it was
			// written down, because the tidy answer is the wrong one.
			name:     "the directory form with a compound basename",
			fileName: "pages/_document/index.helper.tsx",
			source:   "import Head from 'next/head';\n",
		},
		{
			// Ours, and the sharpest thing in this rule: upstream spells the document check two
			// ways inside one function. The basename arm carries a trailing dot
			// (`_document.`) and the directory arm does not (`_document`), so a directory named
			// `_documentation` reads as the document while a file named `_documentation.tsx`
			// does not. One rule disagreeing with itself. Reproduced rather than tidied, and
			// measured on the release binary rather than modelled.
			name:     "a longer directory name sharing the prefix",
			fileName: "pages/_documentation/index.tsx",
			source:   "import Head from 'next/head';\n",
		},
		{
			// Ours, and a deliberate divergence from ESLint that a later reader will want to
			// "fix". ESLint bails unless the path contains `pages`; oxc looks at the basename
			// alone and has no such requirement, so any `_document.tsx` anywhere is treated as
			// the document. oxc is the port target. Measured reporting on the binary.
			name:     "a document outside any pages directory",
			fileName: "source/components/_document.tsx",
			source:   "import Head from 'next/head';\n",
		},
		{
			// Ours. The rule reads the module specifier and never looks at the import clause, so
			// an import binding nothing still reports. Nothing upstream pins this and it is the
			// case a port that reasoned from the rule's name would filter out.
			name:     "a bare side effect import",
			fileName: "pages/_document.tsx",
			source:   "import 'next/head';\n",
		},
		{
			// Ours, and the one most likely to be "improved" away. A type-only import is erased
			// at compile time and arguably harmless, and upstream reports it anyway because it
			// has no import-kind check at all. Measured on the binary rather than reasoned:
			// reporting is upstream's answer, so it is ours.
			name:     "a type only import",
			fileName: "pages/_document.tsx",
			source:   "import type Head from 'next/head';\n",
		},
		{
			// Ours. The local name is irrelevant; every upstream fixture happens to bind `Head`,
			// so a port keying on the binding name would pass all five.
			name:     "a renamed default binding",
			fileName: "pages/_document.tsx",
			source:   "import Whatever from 'next/head';\n",
		},
		{
			// Ours, and the one shape that is string-literal-like without being a string literal.
			// A no-substitution template parses to KindNoSubstitutionTemplateLiteral, which
			// ast.IsStringLiteralLike accepts and whose Text() is the module name, so it reports.
			//
			// Upstream has no answer to compare against: oxc's parser rejects a template specifier
			// outright with "Unexpected token" and the rule is never asked, measured on the release
			// binary. So this is our parser's recovery being reported on rather than a divergence
			// from a decision upstream made, and it is asserted here so the behaviour is pinned
			// rather than incidental. It was written into the silent table first, on the assumption
			// that an unparseable-upstream input should be silent here too, and the fixture caught
			// that: the guard admits it and the rule reports.
			name:     "a template literal specifier",
			fileName: "pages/_document.tsx",
			source:   "import Head from `next/head`;\n",
		},
		{
			// Ours. Another clause shape upstream never writes and still reports on.
			name:     "a namespace import",
			fileName: "pages/_document.tsx",
			source:   "import * as Head from 'next/head';\n",
		},
		{
			// Ours. The quote style is not part of the specifier's text, and upstream's only
			// double-quoted `next/head` sits in a passing case where the path closes the gate,
			// so nothing there proves the match survives the other quote.
			name:     "the double quoted specifier",
			fileName: "pages/_document.tsx",
			source:   "import Head from \"next/head\";\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoHeadImportInDocument, testCase.fileName, testCase.source)
			rule_testing.ExpectFindings(t, result, messageNoHeadImportInDocument.Id)
		})
	}
}

// The rule reports per import declaration with no deduplication, which no upstream fixture pins
// because not one of them writes the module twice. Measured on the release binary, which
// reports at both column 1 and column 34 for a file holding two.
func TestNoHeadImportInDocumentReportsOncePerImport(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		fileName string
		source   string
		findings int
	}{
		{
			name:     "two imports of the module",
			fileName: "pages/_document.tsx",
			source:   "import Head from 'next/head';\nimport Other from 'next/head';\n",
			findings: 2,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoHeadImportInDocument, testCase.fileName, testCase.source)
			expected := make([]string, testCase.findings)
			for index := range expected {
				expected[index] = messageNoHeadImportInDocument.Id
			}
			rule_testing.ExpectFindings(t, result, expected...)
		})
	}
}

// No case here reports.
func TestNoHeadImportInDocumentIsSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		fileName string
		source   string
	}{
		{
			// Upstream. The gate is open and the file imports `Head`, but from `next/document`,
			// which is the correct source and the thing the rule pushes people toward. This is
			// the case that catches a port matching the binding name instead of the specifier.
			name:     "the document importing head from next document",
			fileName: "pages/_document.tsx",
			source:   "import Document, { Html, Head, Main, NextScript } from 'next/document'\n\n\t\t\t      class MyDocument extends Document {\n\t\t\t        static async getInitialProps(ctx) {\n\t\t\t          //...\n\t\t\t        }\n\n\t\t\t        render() {\n\t\t\t          return (\n\t\t\t            <Html>\n\t\t\t              <Head>\n\t\t\t              </Head>\n\t\t\t            </Html>\n\t\t\t          )\n\t\t\t        }\n\t\t\t      }\n\n\t\t\t      export default MyDocument\n\t\t\t    ",
		},
		{
			// Upstream. `next/head` on an ordinary page is exactly right, so the whole rule is a
			// question about the file. This is upstream's only case closing the gate.
			name:     "an ordinary page importing next head",
			fileName: "pages/index.tsx",
			source:   "import Head from \"next/head\";\n\n\t\t\t      export default function IndexPage() {\n\t\t\t        return (\n\t\t\t          <Head>\n\t\t\t            <title>My page title</title>\n\t\t\t            <meta name=\"viewport\" content=\"initial-scale=1.0, width=device-width\" />\n\t\t\t          </Head>\n\t\t\t        );\n\t\t\t      }\n\t\t\t    ",
		},
		{
			// Ours. The trailing dot on the basename arm, which is the one piece of upstream
			// precision worth keeping: ESLint tests `startsWith('_document')` with no dot and
			// has a real false positive here. Note the contrast with the directory case in the
			// reporting table above, where upstream's own second spelling has no dot.
			name:     "a longer file name sharing the prefix",
			fileName: "pages/_documentation.tsx",
			source:   "import Head from 'next/head';\n",
		},
		{
			// Ours. The parent matches and the basename does not begin with `index`, so both
			// halves of the directory arm are load-bearing. Nothing upstream tests the second.
			name:     "a sibling of the document directory",
			fileName: "pages/_document/helpers.tsx",
			source:   "import Head from 'next/head';\n",
		},
		{
			// Ours. A prefix test rather than a contains, and the leading underscore is part of
			// the prefix.
			name:     "a plural sharing the word",
			fileName: "pages/documents.tsx",
			source:   "import Head from 'next/head';\n",
		},
		{
			// Ours. Same distinction from the other side: `strings.Contains` would fire here.
			name:     "a name merely containing the word",
			fileName: "pages/my_document.tsx",
			source:   "import Head from 'next/head';\n",
		},
		{
			// Ours. `starts_with("_document.")` requires the dot, so a file named exactly
			// `_document` is not the document upstream. Counter-intuitive and reproduced.
			name:     "the document with no extension",
			fileName: "pages/_document",
			source:   "import Head from 'next/head';\n",
		},
		{
			// Ours. An equality test rather than a prefix, measured silent on the release binary.
			// A port reaching for HasPrefix passes every upstream fixture and fires here.
			name:     "a deep path into the module",
			fileName: "pages/_document.tsx",
			source:   "import Head from 'next/head/dist/index';\n",
		},
		{
			// Ours. The same equality, from the direction a reader is likeliest to think
			// equivalent. Measured silent.
			name:     "a trailing slash on the specifier",
			fileName: "pages/_document.tsx",
			source:   "import Head from 'next/head/';\n",
		},
		{
			// Ours. A contains test would fire here.
			name:     "a scoped package sharing the name",
			fileName: "pages/_document.tsx",
			source:   "import Head from '@next/head';\n",
		},
		{
			// Ours. So would a suffix test.
			name:     "a relative path sharing the name",
			fileName: "pages/_document.tsx",
			source:   "import Head from './next/head';\n",
		},
		{
			// Ours, and a narrowness reproduced on purpose. Upstream registers the import
			// declaration alone, so a require is invisible to it. `imports.SourceVisitors` on our
			// shelf would catch this and the dynamic form below, and its own doc comment draws
			// the line: a rule guarding a boundary wants all three shapes, a rule asking about
			// the syntax that brings a module in wants one. Measured silent on the binary.
			name:     "a require of the module",
			fileName: "pages/_document.tsx",
			source:   "const Head = require('next/head');\n",
		},
		{
			// Ours. The other shape SourceVisitors would have added. Measured silent.
			name:     "a dynamic import of the module",
			fileName: "pages/_document.tsx",
			source:   "const Head = import('next/head');\n",
		},
		{
			// Ours, and it exists because a mutation deleting the string-literal guard survived
			// every other case here. Our parser recovers from this into an ImportDeclaration whose
			// ModuleSpecifier is a KindBinaryExpression, and Text() on that is not the specifier,
			// so a rule reading it unguarded compares against something that is not a module name.
			// oxc never faces this: its parser rejects the input outright with "Unexpected token"
			// and the rule is never asked, which is why the corpus is silent and why the guard is
			// ours to keep rather than upstream's to have written.
			name:     "an unquoted specifier the parser recovers from",
			fileName: "pages/_document.tsx",
			source:   "import Head from next/head;\n",
		},
		{
			// The other recovered shape, whose specifier parses to a bare KindIdentifier.
			name:     "an empty specifier the parser recovers from",
			fileName: "pages/_document.tsx",
			source:   "import Head from ;\n",
		},
		{
			// Ours, silent for a plainer reason than the two above: it is a different node kind
			// entirely, not an import declaration. Measured silent.
			name:     "a re-export of the module",
			fileName: "pages/_document.tsx",
			source:   "export { default } from 'next/head';\n",
		},
		{
			// Ours, and the shortest form of upstream's first passing case. Named `Head`, imported
			// into the document, and correct.
			name:     "the sibling module the rule points people toward",
			fileName: "pages/_document.tsx",
			source:   "import Document, { Html, Head, Main, NextScript } from 'next/document';\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoHeadImportInDocument, testCase.fileName, testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// ExpectFindings asserts message ids and a count and nothing else, so a rule pointing at the wrong
// node passes every table above. Upstream reports the whole import declaration:
// `ctx.diagnostic(...(import_decl.span))`. Measured on the release oxlint binary against a file
// whose import sits under two comment lines, the span is offset 24 length 28, which starts at the
// `import` keyword and excludes the leading trivia.
//
// That measurement is the reason this rule reports the declaration node directly rather than
// wrapping it in `imports.SpecifierNode`. That helper exists because a node's Pos() includes its
// leading trivia, which once anchored findings at a comment above the import and put them out of
// reach of any suppression a reader could write. `ctx.ReportNode` routes through TokenRange and
// already strips it, so wrapping here would move the finding off upstream's span onto the
// specifier for no gain.
func TestNoHeadImportInDocumentPointsAtTheImportDeclaration(t *testing.T) {
	t.Parallel()

	source := "// a comment\n// another\nimport Head from 'next/head';\n"

	result := rule_testing.Run(t, NoHeadImportInDocument, "pages/_document.tsx", source)
	rule_testing.ExpectFindings(t, result, messageNoHeadImportInDocument.Id)

	reported := source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
	if want := "import Head from 'next/head';"; reported != want {
		t.Fatalf("pointed at %q, want %q", reported, want)
	}
}

// The two findings on one file must point at their own declarations rather than both at the first,
// which a count assertion cannot see. Upstream reports at column 1 and column 34 for the single-line
// form; here they are on separate lines.
func TestNoHeadImportInDocumentPointsAtEachImportSeparately(t *testing.T) {
	t.Parallel()

	source := "import Head from 'next/head';\nimport Other from 'next/head';\n"

	result := rule_testing.Run(t, NoHeadImportInDocument, "pages/_document.tsx", source)
	rule_testing.ExpectFindings(t, result, messageNoHeadImportInDocument.Id, messageNoHeadImportInDocument.Id)

	want := []string{"import Head from 'next/head';", "import Other from 'next/head';"}
	for index, expected := range want {
		range_ := result.Diagnostics[index].Range
		if reported := source[range_.Pos():range_.End()]; reported != expected {
			t.Fatalf("finding %d pointed at %q, want %q", index, reported, expected)
		}
	}
}

// The rendered message is asserted against a literal rather than against the symbol the rule
// reports, and the difference is the whole point. Comparing the result to
// `messageNoHeadImportInDocument.Description` compares the symbol to itself, so a mutation rewriting
// the description moves both sides together and the assertion cannot fail. Measured: that mutant
// SURVIVED against the symbol form of this test and is caught by the literal form below.
//
// Equality rather than containment, because a predicate weaker than the property it guards is not a
// guard.
func TestNoHeadImportInDocumentMessage(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoHeadImportInDocument, "pages/_document.tsx",
		"import Head from 'next/head';\n")
	rule_testing.ExpectFindings(t, result, messageNoHeadImportInDocument.Id)

	if got := result.Diagnostics[0].Message.Id; got != "noHeadImportInDocument" {
		t.Fatalf("reported under the id %q", got)
	}

	want := "This is the custom document and it imports `next/head`. That component collects " +
		"what individual pages contribute to the head and merges it during the page render, " +
		"which is a phase the document has already finished by the time it runs, so what it " +
		"gathers here is silently dropped or duplicated into the shell. The document has its own " +
		"`Head`, exported from `next/document`, and that is the one that writes the shell."
	if got := result.Diagnostics[0].Message.Description; got != want {
		t.Fatalf("rendered %q, want %q", got, want)
	}
}
