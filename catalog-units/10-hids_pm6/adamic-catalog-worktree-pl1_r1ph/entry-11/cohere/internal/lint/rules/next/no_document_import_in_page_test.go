package next

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// The seven passing cases from oxc's corpus, copied byte for byte and pinned by file name. Every
// one of them carries the same import; the corpus varies only the path, because the path is the
// whole judgment. Each was additionally confirmed against the release oxlint binary.
func TestNoDocumentImportInPageIsSilentOnUpstreamPassCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		fileName string
		source   string
	}{
		{
			name:     "the document with a default binding",
			fileName: "pages/_document.js",
			source:   "import Document from \"next/document\"\n\n            export default class MyDocument extends Document {\n              render() {\n                return (\n                  <Html>\n                  </Html>\n                );\n              }\n            }\n            ",
		},
		{
			name:     "the same document reindented, upstream ships both",
			fileName: "pages/_document.js",
			source:   "import Document from \"next/document\"\n\n                export default class MyDocument extends Document {\n                render() {\n                    return (\n                    <Html>\n                    </Html>\n                    );\n                }\n                }\n            ",
		},
		{
			name:     "a renamed binding in the document",
			fileName: "pages/_document.tsx",
			source:   "import NextDocument from \"next/document\"\n\n                export default class MyDocument extends NextDocument {\n                  render() {\n                    return (\n                      <Html>\n                      </Html>\n                    );\n                  }\n                }\n            ",
		},
		{
			name:     "the page suffix convention, a prefix test rather than an equality test",
			fileName: "pages/_document.page.tsx",
			source:   "import Document from \"next/document\"\n\n            export default class MyDocument extends Document {\n              render() {\n                return (\n                  <Html>\n                  </Html>\n                );\n              }\n            }\n            ",
		},
		{
			name:     "the directory form",
			fileName: "pages/_document/index.js",
			source:   "import NDocument from \"next/document\"\n\n            export default class Document extends NDocument {\n              render() {\n                return (\n                  <Html>\n                  </Html>\n                );\n              }\n            }\n            ",
		},
		{
			name:     "the directory form in TypeScript",
			fileName: "pages/_document/index.tsx",
			source:   "import NDocument from \"next/document\"\n\n            export default class Document extends NDocument {\n              render() {\n                return (\n                  <Html>\n                  </Html>\n                );\n              }\n            }\n            ",
		},
		{
			name:     "a directory merely beginning with the word pages, upstream's own probe of the last split segment",
			fileName: "pagesapp/src/pages/_document.js",
			source:   "import Document from \"next/document\"\n\n                export default class MyDocument extends Document {\n                  render() {\n                    return (\n                      <Html>\n                      </Html>\n                    );\n                  }\n                }\n            ",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoDocumentImportInPage, testCase.fileName, testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// The four failing cases from oxc's corpus, copied byte for byte. The last one is the case that
// decides which path predicate this rule may use: a file literally named `_document.tsx` that
// upstream reports, because it is not the immediate child of a `pages` directory. Confirmed
// reporting against the release oxlint binary rather than taken from the snapshot alone.
func TestNoDocumentImportInPageReportsOnUpstreamFailCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		fileName string
		source   string
	}{
		{
			name:     "outside any pages directory",
			fileName: "components/test.js",
			source:   "import Document from \"next/document\"\n\n                export const Test = () => <p>Test</p>\n            ",
		},
		{
			name:     "an ordinary page",
			fileName: "pages/test.js",
			source:   "import Document from \"next/document\"\n\n                export const Test = () => <p>Test</p>\n            ",
		},
		{
			name:     "a page nested below pages",
			fileName: "src/pages/user/test.tsx",
			source:   "import Document from \"next/document\"\n\n                export const Test = () => <p>Test</p>\n            ",
		},
		{
			name:     "a file named _document but not the immediate child of pages",
			fileName: "src/pages/user/_document.tsx",
			source:   "import Document from \"next/document\"\n\n                export const Test = () => <p>Test</p>\n            ",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoDocumentImportInPage, testCase.fileName, testCase.source)
			rule_testing.ExpectFindings(t, result, messageNoDocumentImportInPage.Id)
		})
	}
}

// Cases upstream does not cover, written from reading oxc's body and confirmed against the release
// oxlint binary. The corpus varies only the file path, so the entire specifier surface and every
// import shape other than a default binding is unpinned by it, and a port that guessed either would
// pass all eleven imported fixtures.
func TestNoDocumentImportInPageReportsOnShapesTheCorpusNeverPins(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		fileName string
		source   string
	}{

		{
			// The rule reads the specifier and never the clause. Measured reporting on the release
			// binary; the corpus contains no named import at all.
			name:     "a named import, which the rule name does not suggest",
			fileName: "components/Thing.tsx",
			source:   "import { Html, Main } from \"next/document\";\n",
		},
		{
			// Measured reporting on the release binary. Unpinned upstream.
			name:     "a namespace import",
			fileName: "components/Thing.tsx",
			source:   "import * as NextDocument from \"next/document\";\n",
		},
		{
			// No binding exists to name, and it still reports, which is the clearest evidence that
			// the clause is never examined. Measured on the release binary.
			name:     "a side effect import with no binding",
			fileName: "components/Thing.tsx",
			source:   "import \"next/document\";\n",
		},
		{
			// The case a porter is most likely to filter out as harmless. Upstream has no import kind
			// check so it reports, measured on the release binary, and this reproduces that.
			name:     "a type only import",
			fileName: "components/Thing.tsx",
			source:   "import type Document from \"next/document\";\n",
		},
		{
			// Measured reporting on the release binary.
			name:     "a type modifier on a named specifier",
			fileName: "components/Thing.tsx",
			source:   "import { type Html } from \"next/document\";\n",
		},
		{
			// The quote style is not part of the specifier's value. Measured on the release binary.
			name:     "single quotes around the specifier",
			fileName: "components/Thing.tsx",
			source:   "import Document from 'next/document';\n",
		},
		{
			// A page is not the document, which is the whole point of the rule's name.
			name:     "an ordinary page under pages",
			fileName: "pages/About.tsx",
			source:   "import Document from \"next/document\";\n",
		},
		{
			// Named like the document and still reported, because the shared helper requires a pages
			// segment. Measured reporting on the release binary, and the case where our other
			// predicate, nextjs.IsDocumentFile, would answer the opposite.
			name:     "the document convention outside a pages directory",
			fileName: "components/_document.tsx",
			source:   "import Document from \"next/document\";\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoDocumentImportInPage, testCase.fileName, testCase.source)
			rule_testing.ExpectFindings(t, result, messageNoDocumentImportInPage.Id)
		})
	}
}

// Nothing deduplicates per file, so a file importing the module twice reports twice. Split out
// because ExpectFindings takes one id per finding and the table above passes exactly one.
func TestNoDocumentImportInPageReportsOncePerDeclaration(t *testing.T) {
	t.Parallel()

	source := "import Document from \"next/document\";\n" +
		"import { Html } from \"next/document\";\n"
	result := rule_testing.Run(t, NoDocumentImportInPage, "components/Thing.tsx", source)
	rule_testing.ExpectFindings(t, result,
		messageNoDocumentImportInPage.Id, messageNoDocumentImportInPage.Id)
}

// The silences, each measured on the release binary rather than reasoned about. Several of these
// are shapes the rule's own name suggests it should catch, and reproducing the narrowness is the
// port rather than a gap in it.
func TestNoDocumentImportInPageIsSilentOnShapesTheCorpusNeverPins(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		fileName string
		source   string
	}{

		{
			// An equality test rather than a prefix test. Measured silent on the release binary, and
			// the case a port reaching for HasPrefix would flag.
			name:     "a deeper path under the same package",
			fileName: "components/Thing.tsx",
			source:   "import { Html } from \"next/document/foo\";\n",
		},
		{
			// Measured silent on the release binary. A prefix test would flag it.
			name:     "a plural near miss",
			fileName: "components/Thing.tsx",
			source:   "import D from \"next/documents\";\n",
		},
		{
			// Measured silent. A Contains test would flag this and the two below.
			name:     "a scoped package of the same name",
			fileName: "components/Thing.tsx",
			source:   "import D from \"@next/document\";\n",
		},
		{
			// Measured silent on the release binary.
			name:     "a relative path ending in the same text",
			fileName: "components/Thing.tsx",
			source:   "import D from \"./next/document\";\n",
		},
		{
			// Measured silent on the release binary.
			name:     "a package whose name merely contains the words",
			fileName: "components/Thing.tsx",
			source:   "import D from \"nextdocument\";\n",
		},
		{
			// Upstream registers the import declaration alone, so this is invisible to it. Measured
			// silent on the release binary. Reproduced deliberately: imports.SourceVisitors on our
			// shelf would catch it and would be a divergence, and its own doc comment draws this line.
			name:     "a dynamic import of the module",
			fileName: "components/Thing.tsx",
			source:   "const pending = import(\"next/document\");\n",
		},
		{
			// Same reason as the dynamic import. Measured silent on the release binary.
			name:     "a require of the module",
			fileName: "components/Thing.js",
			source:   "const Document = require(\"next/document\");\n",
		},
		{
			// Not an import declaration at all, so upstream never sees it. Measured silent on the
			// release binary, and the shape that surprised this port most.
			name:     "a re-export from the module",
			fileName: "components/Thing.tsx",
			source:   "export { Html } from \"next/document\";\n",
		},
		{
			// Measured silent on the release binary.
			name:     "a star re-export from the module",
			fileName: "components/Thing.tsx",
			source:   "export * from \"next/document\";\n",
		},
		{
			// The ordinary correct way for a page to contribute to the head.
			name:     "a different next module",
			fileName: "components/Thing.tsx",
			source:   "import Head from \"next/head\";\n",
		},
		{
			// Upstream tests both separators unconditionally and no upstream fixture exercises the
			// backslash branch. The path is not normalized before the test, matching upstream.
			name:     "the module imported inside the document itself, in the windows shape",
			fileName: "pages\\_document.tsx",
			source:   "import Document from \"next/document\";\n",
		},
		{
			// Exempt upstream, because the shared helper's prefix test carries no trailing dot.
			// Measured silent on the release binary. Our other predicate, nextjs.IsDocumentFile, keeps
			// the dot and would report here, which is why this rule does not use it.
			name:     "a longer name sharing the document prefix",
			fileName: "pages/_documentation.tsx",
			source:   "import Document from \"next/document\";\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoDocumentImportInPage, testCase.fileName, testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// ExpectFindings asserts the message id and the count and nothing else, so a rule pointing at the
// wrong node passes a complete fixture pair. Upstream reports `import_decl.span`, measured on the
// release binary at offset 20 length 37 for a file whose import sits under a comment, which is the
// whole declaration starting at the `import` keyword and excluding the leading trivia.
func TestNoDocumentImportInPagePointsAtTheWholeDeclaration(t *testing.T) {
	t.Parallel()

	source := "// a comment above\n" +
		"\n" +
		"import Document from \"next/document\";\n" +
		"\n" +
		"export const Test = 1;\n"

	result := rule_testing.Run(t, NoDocumentImportInPage, "components/Thing.tsx", source)
	rule_testing.ExpectFindings(t, result, messageNoDocumentImportInPage.Id)

	reported := source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
	if want := "import Document from \"next/document\";"; reported != want {
		t.Fatalf("finding points at %q, want %q", reported, want)
	}
}

// The message the finding carries, asserted as a whole rather than by Contains, because a
// predicate weaker than the property it guards is not a guard. This rule interpolates nothing into
// its text, so there is no rendering step that could go wrong, and asserting the identity is what
// there is to assert: it pins that the finding carries THIS message rather than a sibling rule's,
// which the id alone would also satisfy if two rules shared one.
func TestNoDocumentImportInPageCarriesItsOwnMessage(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoDocumentImportInPage, "components/Thing.tsx",
		"import Document from \"next/document\";\n")
	rule_testing.ExpectFindings(t, result, messageNoDocumentImportInPage.Id)

	if got := result.Diagnostics[0].Message; got != messageNoDocumentImportInPage {
		t.Fatalf("message is %+v, want %+v", got, messageNoDocumentImportInPage)
	}
	if got := result.Diagnostics[0].RuleName; got != "@next/next/no-document-import-in-page" {
		t.Fatalf("rule name is %q", got)
	}
}
