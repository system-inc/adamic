package next

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// The four upstream cases are copied verbatim from oxc's tester block and their bytes are verified
// against the extractor's own output rather than by reading them. Each one carries the file path
// upstream gave it, because for this rule the path is load-bearing data rather than decoration: the
// only clean case upstream ships for the gate is a reporting body under a non-document name.
func TestNoStyledJsxInDocumentReports(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		fileName string
		source   string
	}{
		{
			// Upstream's only reporting case. One diagnostic, matching the snapshot.
			name:     "styled jsx inside a document",
			fileName: "pages/_document.jsx",
			source:   "\n\t\t\t            import Document, { Html, Head, Main, NextScript } from 'next/document'\n\n\t\t\t            export class MyDocument extends Document {\n\t\t\t              static async getInitialProps(ctx) {\n\t\t\t                const initialProps = await Document.getInitialProps(ctx)\n\t\t\t                return { ...initialProps }\n\t\t\t              }\n\n\t\t\t              render() {\n\t\t\t                return (\n\t\t\t                  <Html>\n\t\t\t                    <Head />\n\t\t\t                    <style jsx>{\"                    body{                      color:red;                    }                    \"}</style>\n\t\t\t                    <body>\n\t\t\t                      <Main />\n\t\t\t                      <NextScript />\n\t\t\t                    </body>\n\t\t\t                  </Html>\n\t\t\t                )\n\t\t\t              }\n\t\t\t            }",
		}, {
			// Invented. Upstream's gate is a basename test with no `pages` component at all, so a
			// document outside a pages directory reports. Pinned against the release binary, which
			// reports here. Without this case a port that quietly adopted eslint's `pages` split
			// would pass every upstream fixture.
			name:     "a document outside a pages directory",
			fileName: "components/_document.tsx",
			source:   "export const C = () => <style jsx>{\"body{color:red}\"}</style>;",
		},
		{
			// Invented, and this one is the divergence rather than a confirmation. Upstream is
			// SILENT on the directory form: this rule's own gate tests the basename, and
			// `index.tsx` does not begin with `_document.`. We report, because the shared
			// `nextjs.IsDocumentFile` is the union of the family's four spellings and one sibling
			// accepts this path. Measured on the release binary, which reports nothing here.
			// Recorded as reporting because that is what our rule decides, and the reason is at the
			// helper and in this rule's doc comment.
			name:     "the document directory form, wider than this rule upstream",
			fileName: "pages/_document/index.tsx",
			source:   "export const C = () => <style jsx>{\"body{color:red}\"}</style>;",
		},
		{
			// Invented. The attribute is matched by presence, not by value, so every value shape
			// reports. All four pinned against the release binary.
			name:     "a jsx attribute carrying a value",
			fileName: "pages/_document.tsx",
			source:   "export const C = () => <style jsx=\"true\">{\"x\"}</style>;",
		},
		{
			// Invented. styled-jsx's own `global` modifier, which upstream reports because `jsx` is
			// still present. Pinned.
			name:     "jsx global",
			fileName: "pages/_document.tsx",
			source:   "export const C = () => <style jsx global>{\"x\"}</style>;",
		},
		{
			// Invented, and it is the reason this rule registers two kinds where oxc registers one.
			// A self-closing element is a distinct AST kind here and is folded into
			// `JSXOpeningElement` upstream, so a single listener would make this silent. Upstream
			// reports it; pinned.
			name:     "a self closing styled jsx element",
			fileName: "pages/_document.tsx",
			source:   "export const C = () => <style jsx />;",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoStyledJsxInDocument, testCase.fileName, testCase.source)
			rule_testing.ExpectFindings(t, result, messageNoStyledJsxInDocument.Id)
		})
	}
}

func TestNoStyledJsxInDocumentIsSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		fileName string
		source   string
	}{
		{
			// Upstream's baseline: a document with no <style> at all.
			name:     "a clean document with no style element",
			fileName: "pages/_document.tsx",
			source:   "import Document, { Html, Head, Main, NextScript } from 'next/document'\n\n\t\t\t        export class MyDocument extends Document {\n\t\t\t          static async getInitialProps(ctx) {\n\t\t\t            const initialProps = await Document.getInitialProps(ctx)\n\t\t\t            return { ...initialProps }\n\t\t\t          }\n\n\t\t\t          render() {\n\t\t\t            return (\n\t\t\t              <Html>\n\t\t\t                <Head />\n\t\t\t                <body>\n\t\t\t                  <Main />\n\t\t\t                  <NextScript />\n\t\t\t                </body>\n\t\t\t              </Html>\n\t\t\t            )\n\t\t\t          }\n\t\t\t        }",
		},
		{
			// The valuable clean case, and it pins two distinct exemptions at once. The first
			// <style> has an expression child and no attributes: plain CSS in a document is legal
			// and only styled-jsx is not. The second carries a spread as its only attribute, which
			// is the shape a hand-rolled loop gets wrong by reading a name off a node that has
			// none.
			name:     "plain css and a spread only style in a document",
			fileName: "pages/_document.tsx",
			source:   "import Document, { Html, Head, Main, NextScript } from 'next/document'\n\n\t\t\t        export class MyDocument extends Document {\n\t\t\t          static async getInitialProps(ctx) {\n\t\t\t            const initialProps = await Document.getInitialProps(ctx)\n\t\t\t            return { ...initialProps }\n\t\t\t          }\n\n\t\t\t          render() {\n\t\t\t            return (\n\t\t\t              <Html>\n\t\t\t                <Head />\n\t\t\t                <style>{\"                  body{                    color:red;                  }                \"}</style>\n\t\t\t                <style {...{nonce: '123' }}></style>\n\t\t\t                <body>\n\t\t\t                  <Main />\n\t\t\t                  <NextScript />\n\t\t\t                </body>\n\t\t\t              </Html>\n\t\t\t            )\n\t\t\t          }\n\t\t\t        }",
		},
		{
			// Byte-identical offending code to the reporting case, silent only because the file is
			// not a document. This is the corpus's only test of the gate, and it can only catch a
			// gate broken open. A gate broken shut is caught by nothing upstream ships, which is
			// why the invented cases below exist.
			name:     "styled jsx outside the document",
			fileName: "pages/index.jsx",
			source:   "\n\t\t\t          export default function Page() {\n\t\t\t            return (\n\t\t\t              <>\n\t\t\t                <p>Hello world</p>\n\t\t\t                <style jsx>{`\n\t\t\t                  p {\n\t\t\t                    color: orange;\n\t\t\t                  }\n\t\t\t                `}</style>\n\t\t\t              </>\n\t\t\t            )\n\t\t\t          }\n\t\t\t          ",
		}, {
			// Invented, and it pins the trailing dot in `_document.`. Upstream's basename test
			// requires it, so a file merely beginning with the word is not the document. That dot
			// is the one piece of precision oxc has over eslint here, whose parsed-name test reads
			// this as the document and fires. Pinned silent against the release binary.
			name:     "a file whose name merely begins with the word document",
			fileName: "pages/_documentation.tsx",
			source:   "export const C = () => <style jsx>{\"body{color:red}\"}</style>;",
		},
		{
			// Invented. The spread is a real `KindJsxSpreadAttribute` node here and its object is
			// readable, so declining it is a decision rather than a limitation. Upstream is silent
			// and so are we. This is the strongest form of upstream's own `nonce` spread case.
			name:     "a spread supplying the jsx attribute",
			fileName: "pages/_document.tsx",
			source:   "export const C = () => <style {...{jsx: true}}></style>;",
		},
		{
			// Invented. The name comparison is bytewise upstream. Pinned silent.
			name:     "a jsx attribute in the wrong case",
			fileName: "pages/_document.tsx",
			source:   "export const C = () => <style JSX>{\"x\"}</style>;",
		},
		{
			// Invented. `global` alone is not styled-jsx to this rule. Pinned silent.
			name:     "global without jsx",
			fileName: "pages/_document.tsx",
			source:   "export const C = () => <style global>{\"x\"}</style>;",
		},
		{
			// Invented. The tag must be the intrinsic `style`, so a component of that name is not
			// it. Pinned silent, and it is the case that would break if the tag test were dropped.
			name:     "a capitalized style component",
			fileName: "pages/_document.tsx",
			source:   "export const C = () => <Style jsx>{\"x\"}</Style>;",
		},
		{
			// Invented, and it is the case that would panic rather than merely misreport if the tag
			// were read with `Text()` before its kind was checked: a dotted JSX tag parses to a
			// `KindPropertyAccessExpression`, which panics there. Pinned silent upstream.
			name:     "a member expression tag named style",
			fileName: "pages/_document.tsx",
			source:   "export const C = () => <style.x jsx>{\"x\"}</style.x>;",
		},
		{
			// Invented. A `jsx` attribute on some other intrinsic is not styled-jsx. Pinned silent.
			name:     "a jsx attribute on a different element",
			fileName: "pages/_document.tsx",
			source:   "export const C = () => <div jsx>{\"x\"}</div>;",
		},
		{
			// Invented. An attribute merely containing the name does not match. Pinned silent.
			name:     "an attribute whose name contains jsx",
			fileName: "pages/_document.tsx",
			source:   "export const C = () => <style data-jsx>{\"x\"}</style>;",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoStyledJsxInDocument, testCase.fileName, testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// The message id says which rule fired and nothing about where it pointed or what it said, so both
// are asserted here. Upstream's span is the opening element alone rather than the whole element:
// measured on the release binary at offset 17 length 11 over `export const a = <style jsx>{"x"}
// </style>;`, which is exactly `<style jsx>`.
func TestNoStyledJsxInDocumentPointsAtTheOpeningElement(t *testing.T) {
	t.Parallel()

	source := "export const a = <style jsx>{\"x\"}</style>;\n"
	result := rule_testing.Run(t, NoStyledJsxInDocument, "pages/_document.tsx", source)
	rule_testing.ExpectFindings(t, result, messageNoStyledJsxInDocument.Id)

	reported := source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
	if reported != "<style jsx>" {
		t.Fatalf("expected the finding to cover the opening element, got %q", reported)
	}
}

// A self-closing element reports over its whole self, since it has no separate opening node.
func TestNoStyledJsxInDocumentPointsAtASelfClosingElement(t *testing.T) {
	t.Parallel()

	source := "export const a = <style jsx />;\n"
	result := rule_testing.Run(t, NoStyledJsxInDocument, "pages/_document.tsx", source)
	rule_testing.ExpectFindings(t, result, messageNoStyledJsxInDocument.Id)

	reported := source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
	if reported != "<style jsx />" {
		t.Fatalf("expected the finding to cover the element, got %q", reported)
	}
}

// The rendered text is asserted rather than a substring of it, because a predicate weaker than the
// property it guards is not a guard.
func TestNoStyledJsxInDocumentRendersItsMessage(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoStyledJsxInDocument, "pages/_document.tsx",
		"export const a = <style jsx>{\"x\"}</style>;\n")
	rule_testing.ExpectFindings(t, result, messageNoStyledJsxInDocument.Id)

	if got := result.Diagnostics[0].Message.Description; got != messageNoStyledJsxInDocument.Description {
		t.Fatalf("expected the message description verbatim, got %q", got)
	}
	if !strings.HasPrefix(messageNoStyledJsxInDocument.Description, "This is styled-jsx inside the custom document.") {
		t.Fatalf("the message should say what is wrong rather than restate the rule name")
	}
}
