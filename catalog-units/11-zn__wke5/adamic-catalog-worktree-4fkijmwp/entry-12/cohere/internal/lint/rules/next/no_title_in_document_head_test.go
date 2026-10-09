package next

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// The three cases upstream ships, byte for byte out of the extractor's dump so no escape is retyped
// on the way in. They are thin, and two of the three pass for more reasons than they look like they
// test, which is why the invented cases below outnumber them by a wide margin.
const (
	upstreamPassNextHead = "import Head from \"next/head\";\n\n\t\t\t     class Test {\n\t\t\t      render() {\n\t\t\t        return (\n\t\t\t          <Head>\n\t\t\t            <title>My page title</title>\n\t\t\t          </Head>\n\t\t\t        );\n\t\t\t      }\n\t\t\t     }"

	upstreamPassEmptyHead = "import Document, { Html, Head } from \"next/document\";\n\n\t\t\t     class MyDocument extends Document {\n\t\t\t      render() {\n\t\t\t        return (\n\t\t\t          <Html>\n\t\t\t            <Head>\n\t\t\t            </Head>\n\t\t\t          </Html>\n\t\t\t        );\n\t\t\t      }\n\t\t\t     }\n\n\t\t\t     export default MyDocument;\n\t\t\t     "

	upstreamFail = "\n\t\t\t      import { Head } from \"next/document\";\n\n\t\t\t      class Test {\n\t\t\t        render() {\n\t\t\t          return (\n\t\t\t            <Head>\n\t\t\t              <title>My page title</title>\n\t\t\t            </Head>\n\t\t\t          );\n\t\t\t        }\n\t\t\t      }"
)

func TestNoTitleInDocumentHeadReports(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
		want   int
	}{
		{
			name:   "the upstream failure",
			source: upstreamFail,
			want:   1,
		},
		{
			// The alias is the case every upstream fixture omits and the one a port matching the
			// text `Head` gets wrong while staying green on the whole corpus. Measured reporting on
			// the release binary.
			name:   "an aliased named import",
			source: "import { Head as H } from \"next/document\";\nexport const C = () => <H><title>x</title></H>;",
			want:   1,
		},
		{
			// The rule never checks that the imported name is `Head`. Any first named specifier
			// arms it, so `Html` does too. Measured reporting; pinned so nobody helpfully adds a
			// name check that would look like a correctness improvement and would silence this.
			name:   "a named import that is not Head at all",
			source: "import { Html } from \"next/document\";\nexport const C = () => <Html><title>x</title></Html>;",
			want:   1,
		},
		{
			// Our AST gives a self-closing element its own kind rather than an opening element, so
			// a child scan reading only KindJsxOpeningElement is silent here while every upstream
			// fixture passes. Measured reporting.
			name:   "a self-closing title child",
			source: "import { Head } from \"next/document\";\nexport const C = () => <Head><title /></Head>;",
			want:   1,
		},
		{
			// Upstream reports once per matching child rather than once per element, so two titles
			// under one Head produce two findings at the same span. Measured: two diagnostics.
			name:   "two title children report twice",
			source: "import { Head } from \"next/document\";\nexport const C = () => <Head><title>a</title><title>b</title></Head>;",
			want:   2,
		},
		{
			// Only the FIRST named specifier of a declaration is watched, so two declarations each
			// contribute their own. Measured reporting.
			name:   "a second declaration brings its own first specifier",
			source: "import { Html } from \"next/document\";\nimport { Head } from \"next/document\";\nexport const C = () => <Head><title>x</title></Head>;",
			want:   1,
		},
		{
			// The first named specifier is the one watched even when a default precedes it in the
			// same declaration, because the default is a different specifier arm and is skipped
			// rather than counted. Measured reporting on `<Html>`, the first NAMED name.
			name:   "a default before the first named specifier does not consume the slot",
			source: "import Document, { Html, Head } from \"next/document\";\nexport const C = () => <Html><title>x</title></Html>;",
			want:   1,
		},
		{
			name:   "a single named specifier after a default",
			source: "import Document, { Head } from \"next/document\";\nexport const C = () => <Head><title>x</title></Head>;",
			want:   1,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoTitleInDocumentHead, "pages/_document.tsx", testCase.source)
			ids := make([]string, testCase.want)
			for index := range ids {
				ids[index] = messageNoTitleInDocumentHead.Id
			}
			rule_testing.ExpectFindings(t, result, ids...)
		})
	}
}

func TestNoTitleInDocumentHeadIsSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
	}{
		{
			// Upstream's pass 1. It is clean for two independent reasons, the wrong module and the
			// default specifier arm, so on its own it distinguishes neither. The two cases below it
			// separate them.
			name:   "the upstream next/head pass",
			source: upstreamPassNextHead,
		},
		{
			// Upstream's pass 2, and it is clean for two reasons as well: the Head is empty, and
			// the watched binding is `Html` rather than `Head` because only the first named
			// specifier arms the rule.
			name:   "the upstream empty head pass",
			source: upstreamPassEmptyHead,
		},
		{
			// The half of pass 1 that is about the specifier arm. A default import from the RIGHT
			// module is still silent, because the rule matches only the named arm. This is the one
			// fixture that separates the two reasons pass 1 is clean, and a port accepting the
			// default arm passes all three upstream cases without it. Measured silent.
			name:   "a default import from next/document",
			source: "import Head from \"next/document\";\nexport const C = () => <Head><title>x</title></Head>;",
		},
		{
			// The half of pass 1 that is about the module. A named import from the wrong module is
			// silent. Measured silent.
			name:   "a named import from next/head",
			source: "import { Head } from \"next/head\";\nexport const C = () => <Head><title>x</title></Head>;",
		},
		{
			// Only the first named specifier is watched. `Head` is second here, so nothing watches
			// it and a real violation goes unreported. This is upstream's sharpest defect and it is
			// reproduced rather than fixed; see the rule's doc comment. Measured silent.
			name:   "a title under a named import that is not the first one",
			source: "import { Html, Head } from \"next/document\";\nexport const C = () => <Head><title>x</title></Head>;",
		},
		{
			name:   "a title under the third of three named imports",
			source: "import { A, Head, C } from \"next/document\";\nexport const C2 = () => <Head><title>x</title></Head>;",
		},
		{
			// Only DIRECT children are scanned. A title one level down is silent. Measured silent.
			name:   "a title nested below a direct child",
			source: "import { Head } from \"next/document\";\nexport const C = () => <Head><div><title>x</title></div></Head>;",
		},
		{
			// A JSX expression container is not a JSX element child, so the title inside it is not
			// seen. Measured silent.
			name:   "a title inside an expression container",
			source: "import { Head } from \"next/document\";\nexport const C = (cond: boolean) => <Head>{cond ? <title>x</title> : null}</Head>;",
		},
		{
			// A namespace import is a third specifier arm and never matches, and the tag is a
			// member expression besides. Measured silent.
			name:   "a namespace import",
			source: "import * as Doc from \"next/document\";\nexport const C = () => <Doc.Head><title>x</title></Doc.Head>;",
		},
		{
			// The binding used as the OBJECT of a member-expression tag is not the tag itself, so
			// the reference's parent is not a JSX opening element. Measured silent.
			name:   "the binding as the object of a member expression tag",
			source: "import { Head } from \"next/document\";\nexport const C = () => <Head.Sub><title>x</title></Head.Sub>;",
		},
		{
			// A self-closing Head has no children to scan. Measured silent.
			name:   "a self-closing Head",
			source: "import { Head } from \"next/document\";\nexport const C = () => <Head />;",
		},
		{
			// The child name is compared exactly, so the capitalised component is a different thing
			// and never the intrinsic element. Measured silent.
			name:   "a capitalised Title child",
			source: "import { Head } from \"next/document\";\nexport const C = () => <Head><Title>x</Title></Head>;",
		},
		{
			// A reference outside JSX has no opening element parent. Measured silent.
			name:   "a non-JSX reference to the binding",
			source: "import { Head } from \"next/document\";\nconst held = Head;\nexport const C = () => <div><title>x</title></div>;",
		},
		{
			// A require destructure is not an import declaration and upstream's entry point is the
			// declaration kind alone. Measured silent.
			name:   "a require destructure rather than an import",
			source: "const { Head } = require(\"next/document\");\nexport const C = () => <Head><title>x</title></Head>;",
		},
		{
			// The substring trap. `next/document-x` is a different module and the comparison is
			// exact. No upstream fixture writes this.
			name:   "a module whose name extends the specifier",
			source: "import { Head } from \"next/documents\";\nexport const C = () => <Head><title>x</title></Head>;",
		},
		{
			// A side-effect import binds nothing, so there is no first named specifier.
			name:   "a side-effect import",
			source: "import \"next/document\";\nexport const C = () => <Head><title>x</title></Head>;",
		},
		{
			name:   "no next/document import at all",
			source: "export const C = () => <Head><title>x</title></Head>;",
		},
		{
			// A local component shadowing the import is a different binding. Our port resolves
			// references by name within the file rather than through a scope index, so this is the
			// case that separates the two; see the rule's doc comment on the stated divergence.
			// Measured silent upstream.
			name:   "a local declaration shadowing the import",
			source: "import { Head } from \"next/document\";\nfunction Inner() { const Head = (p: {children?: unknown}) => null; return <Head><title>x</title></Head>; }",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoTitleInDocumentHead, "pages/_document.tsx", testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// The finding points at the `Head` tag name identifier and nothing else, which is upstream's
// `jsx_opening_element.name.span()` rather than the `<title>` that eslint blames. A message-id
// assertion cannot see where a finding lands, and pointing at the title instead would satisfy every
// case above.
func TestNoTitleInDocumentHeadPointsAtTheTagName(t *testing.T) {
	t.Parallel()

	source := "import { Head as PageHead } from \"next/document\";\nexport const C = () => <PageHead><title>x</title></PageHead>;"

	result := rule_testing.RunTyped(t, NoTitleInDocumentHead, "pages/_document.tsx", source)
	rule_testing.ExpectFindings(t, result, messageNoTitleInDocumentHead.Id)

	diagnostic := result.Diagnostics[0]
	reported := source[diagnostic.Range.Pos():diagnostic.Range.End()]
	if reported != "PageHead" {
		t.Fatalf("want the tag name %q, got %q", "PageHead", reported)
	}
}

// The message is asserted whole rather than by substring, because a substring predicate is weaker
// than the property it guards and has gone green over a wrong message in this tree before.
func TestNoTitleInDocumentHeadMessageText(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTyped(t, NoTitleInDocumentHead, "pages/_document.tsx", upstreamFail)
	rule_testing.ExpectFindings(t, result, messageNoTitleInDocumentHead.Id)

	if !strings.HasPrefix(messageNoTitleInDocumentHead.Description, "A <title> here is emitted") {
		t.Fatalf("message description changed shape: %q", messageNoTitleInDocumentHead.Description)
	}
}
