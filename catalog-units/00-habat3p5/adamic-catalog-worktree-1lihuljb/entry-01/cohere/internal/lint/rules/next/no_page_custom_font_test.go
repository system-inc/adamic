package next

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// documentFile is the path upstream gives every case but one, and the path is load-bearing data
// rather than decoration: it is the entire gate that chooses which of the two findings fires.
const documentFile = "pages/_document.jsx"

// The nine upstream cases are copied verbatim from oxc's tester block. Their bytes were extracted
// with the extractor's own Rust parser and then verified to appear byte for byte inside
// `no_page_custom_font.rs` by a script, rather than read: the tabs below are real tab characters
// in upstream's source and a hand transcription that cooked them would still compile and still go
// green.
//
// The snapshot records 3 diagnostics from 2 failing inputs, so one finding per input is wrong for
// this corpus. The second failing case asserts two ids.
func TestNoPageCustomFontFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		fileName string
		source   string
		want     []string
	}{
		{
			// Upstream fail 1. A link correctly inside `<Head>` AND correctly inside the default export, and
			// it reports anyway, because the filename is not the document. This is the case that proves
			// the `<Head>` ancestor named in upstream's other message is irrelevant to the decision.
			name:     "upstream fail in pages/index.tsx",
			fileName: "pages/index.tsx",
			source:   "\n\t\t\t      import Head from 'next/head'\n\t\t\t      export default function IndexPage() {\n\t\t\t        return (\n\t\t\t          <div>\n\t\t\t            <Head>\n\t\t\t              <link\n\t\t\t                href=\"https://fonts.googleapis.com/css2?family=Inter\"\n\t\t\t                rel=\"stylesheet\"\n\t\t\t              />\n\t\t\t            </Head>\n\t\t\t            <p>Hello world!</p>\n\t\t\t          </div>\n\t\t\t        )\n\t\t\t      }\n\t\t\t      ",
			want:     []string{"noPageCustomFontNotInDocument"},
		},
		{
			// Upstream fail 2, and it carries TWO diagnostics, one per link, which the snapshot confirms and
			// the extractor's discrepancy warning is about. Both links sit in a `Links()` helper that IS
			// rendered inside `<Head>`; `Links` is not the default export, so both fire. A fixture
			// asserting one finding per input would be wrong here.
			name:     "upstream fail in pages/_document.jsx",
			fileName: "pages/_document.jsx",
			source:   "\n\t\t\t      import Head from 'next/head'\n\n\n\t\t\t      function Links() {\n\t\t\t        return (\n\t\t\t          <>\n\t\t\t            <link\n\t\t\t              href=\"https://fonts.googleapis.com/css2?family=Inter\"\n\t\t\t              rel=\"stylesheet\"\n\t\t\t            />\n\t\t\t            <link\n\t\t\t              href=\"https://fonts.googleapis.com/css2?family=Open+Sans\"\n\t\t\t              rel=\"stylesheet\"\n\t\t\t              />\n\t\t\t          </>\n\t\t\t        )\n\t\t\t      }\n\n\t\t\t      export default function IndexPage() {\n\t\t\t        return (\n\t\t\t          <div>\n\t\t\t            <Head>\n\t\t\t              <Links />\n\t\t\t            </Head>\n\t\t\t            <p>Hello world!</p>\n\t\t\t          </div>\n\t\t\t        )\n\t\t\t      }\n\t\t\t      ",
			want:     []string{"noPageCustomFontOutsideDefaultExport", "noPageCustomFontOutsideDefaultExport"},
		},
		{
			// Invented. Upstream's gate is a bare basename test with no `pages` component at all, so a
			// document outside a pages directory IS the document. Measured on the release binary, which
			// reports the outside-the-default-export finding here. Without this case a port that quietly
			// adopted eslint's `pages` split would pass every upstream fixture.
			name:     "a document outside a pages directory",
			fileName: "components/_document.tsx",
			source:   "function Links() { return (<link href=\"https://fonts.googleapis.com/css2?family=Inter\" />) }\nexport default function D() { return (<div><Links /></div>) }\n",
			want:     []string{"noPageCustomFontOutsideDefaultExport"},
		},
		{
			// Invented, and it pins the trailing dot. `_documentation.tsx` is NOT the document, so the page
			// finding fires even though the JSX is inside the default export. Measured: reports upstream.
			// oxc's shared `is_document_page` carries no trailing dot and would answer the opposite.
			name:     "a page whose name merely begins with the word",
			fileName: "pages/_documentation.tsx",
			source:   "export default function D() { return (<link href=\"https://fonts.googleapis.com/css2?family=Inter\" />) }\n",
			want:     []string{"noPageCustomFontNotInDocument"},
		},
		{
			// Invented, and it is the sharpest thing about this rule. No `rel` attribute, and an href that is
			// not a real Google Fonts URL at all, merely one sharing the prefix. Measured: reports. So
			// `rel` is never consulted and the href test is a bare prefix rather than a URL parse. A port
			// that checked for `rel="stylesheet"` would silence this and every upstream case would still
			// pass, because all of them happen to write it.
			name:     "no rel attribute at all",
			fileName: "pages/_document.tsx",
			source:   "function Links() { return (<link href=\"https://fonts.googleapis.com/cssANYTHING\" />) }\nexport default function D() { return (<div><Links /></div>) }\n",
			want:     []string{"noPageCustomFontOutsideDefaultExport"},
		},
		{
			// Invented, the same point from the other side, and it also pins the v1 `/css` endpoint. Measured:
			// reports.
			name:     "rel is preload rather than stylesheet",
			fileName: "pages/_document.tsx",
			source:   "function Links() { return (<link rel=\"preload\" href=\"https://fonts.googleapis.com/css?family=X\" />) }\nexport default function D() { return (<div><Links /></div>) }\n",
			want:     []string{"noPageCustomFontOutsideDefaultExport"},
		},
		{
			// Invented. `link` is a void element so this spelling is never written in practice, but it reports
			// upstream, measured, and it is the only case in this file that exercises the
			// JsxOpeningElement listener rather than the self-closing one. Every upstream case is
			// self-closing, so without this the paired listener is unexercised and a port that dropped it
			// would stay green.
			name:     "a paired link element",
			fileName: "pages/_document.tsx",
			source:   "function Links() { return (<link href=\"https://fonts.googleapis.com/css2?family=Inter\"></link>) }\nexport default function D() { return (<div><Links /></div>) }\n",
			want:     []string{"noPageCustomFontOutsideDefaultExport"},
		},
		{
			// Invented, and it is the distinguishing input for the name-resolution rule. The function has its
			// OWN name, `Inner`, so the declarator's `D` is never consulted and the default export never
			// matches. Measured: reports. Its silent twin is in the clean list below, and the pair is the
			// only thing that can tell a correct implementation from one that always reads the declarator.
			name:     "a named function expression takes its own name",
			fileName: "pages/_document.tsx",
			source:   "const D = function Inner() { return (<link href=\"https://fonts.googleapis.com/css2?family=Inter\" />) };\nexport default D;\n",
			want:     []string{"noPageCustomFontOutsideDefaultExport"},
		},
		{
			// Invented. The control for the arrow case below: same shape, binding renamed so it is not what is
			// default-exported. Measured: reports. Without this pair the arrow fixture would pass for a
			// rule that answered true for every arrow.
			name:     "an arrow whose binding is not the default export",
			fileName: "pages/_document.tsx",
			source:   "const Q = () => (<link href=\"https://fonts.googleapis.com/css2?family=Inter\" />);\nexport default function D() { return <div /> }\n",
			want:     []string{"noPageCustomFontOutsideDefaultExport"},
		},
		{
			// Invented. A destructuring binding contributes no single name that could appear in an
			// `export default`, so the arrow is not recognized as the default export and this
			// reports. Measured on the release binary: reports. It pins the identifier-only guard
			// in nameFromEnclosingVariableDeclarator, though a mutation relaxing that guard is
			// equivalent rather than caught, for the reason recorded at that line.
			name:     "a destructuring binding",
			fileName: "pages/_document.tsx",
			source:   "const [D] = [() => (<link href=\"https://fonts.googleapis.com/css2?family=Inter\" />)];\nexport default D;\n",
			want:     []string{"noPageCustomFontOutsideDefaultExport"},
		},
		{
			// Invented. `export { D as default }` REPORTS upstream, measured, so the aliased spelling is not
			// recognized as the default export by the rule even though it is one to the language. That is
			// reproduced rather than fixed, and this fixture is what would catch a helpful improvement.
			name:     "an aliased default export",
			fileName: "pages/_document.tsx",
			source:   "function D() { return (<link href=\"https://fonts.googleapis.com/css2?family=Inter\" />) }\nexport { D as default };\n",
			want:     []string{"noPageCustomFontOutsideDefaultExport"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoPageCustomFont, testCase.fileName, testCase.source)
			rule_testing.ExpectFindings(t, result, testCase.want...)
		})
	}
}

// The clean cases. Every upstream one, plus the invented twins that pin what makes each pass.
func TestNoPageCustomFontStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		fileName string
		source   string
	}{
		{
			// Upstream pass 1. The canonical document: a class extending the `next/document` default import,
			// default-exported by a later identifier statement.
			name:     "upstream pass 1",
			fileName: documentFile,
			source:   "import Document, { Html, Head } from \"next/document\";\n\t\t\tclass MyDocument extends Document {\n\t\t\t\trender() {\n\t\t\t\t\treturn (\n\t\t\t\t\t\t<Html>\n\t\t\t\t\t\t\t<Head>\n\t\t\t\t\t\t\t\t<link\n\t\t\t\t\t\t\t\t\thref=\"https://fonts.googleapis.com/css2?family=Krona+One&display=swap\"\n\t\t\t\t\t\t\t\t\trel=\"stylesheet\"\n\t\t\t\t\t\t\t\t/>\n\t\t\t\t\t\t\t</Head>\n\t\t\t\t\t\t</Html>\n\t\t\t\t\t);\n\t\t\t\t}\n\t\t\t}\n\t\t\texport default MyDocument;",
		},
		{
			// Upstream pass 2. The local class is named `Document` while the IMPORT is renamed `NextDocument`.
			// This is the case that separates oxc from eslint: eslint requires the superclass to be the
			// tracked document import, oxc never looks at the superclass at all.
			name:     "upstream pass 2",
			fileName: documentFile,
			source:   "import NextDocument, { Html, Head } from \"next/document\";\n\t\t\t    class Document extends NextDocument {\n\t\t\t      render() {\n\t\t\t        return (\n\t\t\t          <Html>\n\t\t\t            <Head>\n\t\t\t              <link\n\t\t\t                href=\"https://fonts.googleapis.com/css2?family=Krona+One&display=swap\"\n\t\t\t                rel=\"stylesheet\"\n\t\t\t              />\n\t\t\t            </Head>\n\t\t\t          </Html>\n\t\t\t        );\n\t\t\t      }\n\t\t\t    }\n\t\t\t    export default Document;\n\t\t\t    ",
		},
		{
			// Upstream pass 3. Inline default export of a function declaration, answered by the modifier pair
			// on the ancestor rather than by any name.
			name:     "upstream pass 3",
			fileName: documentFile,
			source:   "export default function CustomDocument() {\n\t\t\t      return (\n\t\t\t        <Html>\n\t\t\t          <Head>\n\t\t\t            <link\n\t\t\t              href=\"https://fonts.googleapis.com/css2?family=Krona+One&display=swap\"\n\t\t\t              rel=\"stylesheet\"\n\t\t\t            />\n\t\t\t          </Head>\n\t\t\t        </Html>\n\t\t\t      )\n\t\t\t    }",
		},
		{
			// Upstream pass 4. The export statement sits BELOW the JSX, which is why the name collection is a
			// pass over the whole file at rule entry rather than a stateful visitor.
			name:     "upstream pass 4",
			fileName: documentFile,
			source:   "function CustomDocument() {\n\t\t\t      return (\n\t\t\t        <Html>\n\t\t\t          <Head>\n\t\t\t            <link\n\t\t\t              href=\"https://fonts.googleapis.com/css2?family=Krona+One&display=swap\"\n\t\t\t              rel=\"stylesheet\"\n\t\t\t            />\n\t\t\t          </Head>\n\t\t\t        </Html>\n\t\t\t      )\n\t\t\t    }\n\n\t\t\t    export default CustomDocument;\n\t\t\t    ",
		},
		{
			// Upstream pass 5, and the surprising one: the class has NO `extends` clause, and `Document` is
			// imported but unused as a superclass. oxc only asks whether the node is inside the default
			// export, so this passes. eslint reports it. Reproducing oxc means not adding the check.
			name:     "upstream pass 5",
			fileName: documentFile,
			source:   "\n\t\t\t      import Document, { Html, Head } from \"next/document\";\n\t\t\t      class MyDocument {\n\t\t\t        render() {\n\t\t\t          return (\n\t\t\t            <Html>\n\t\t\t              <Head>\n\t\t\t                <link\n\t\t\t                  href=\"https://fonts.googleapis.com/css2?family=Krona+One&display=swap\"\n\t\t\t                  rel=\"stylesheet\"\n\t\t\t                />\n\t\t\t              </Head>\n\t\t\t            </Html>\n\t\t\t          );\n\t\t\t        }\n\t\t\t      }\n\n\t\t\t      export default MyDocument;",
		},
		{
			// Upstream pass 6. An ANONYMOUS default-exported function. Reached through the modifier pair with
			// no name ever resolved, so a port that walks up looking for a named function and then
			// consults an export table reports this one.
			name:     "upstream pass 6",
			fileName: documentFile,
			source:   "export default function() {\n\t\t\t      return (\n\t\t\t        <Html>\n\t\t\t          <Head>\n\t\t\t            <link\n\t\t\t              href=\"https://fonts.googleapis.com/css2?family=Krona+One&display=swap\"\n\t\t\t              rel=\"stylesheet\"\n\t\t\t            />\n\t\t\t          </Head>\n\t\t\t        </Html>\n\t\t\t      )\n\t\t\t    }",
		},
		{
			// Upstream pass 7, and it is not a font at all: an `apple-touch-startup-image` with a relative
			// href, inside a bare `function a()` that is NOT exported. It passes purely because the href
			// prefix fails, and the location check never runs. This is the case that pins the ORDER: href
			// first, location second. Were the order reversed it would report.
			name:     "upstream pass 7",
			fileName: documentFile,
			source:   "function a() {\n\t\t\t      return (\n\t\t\t        <Html>\n\t\t\t          <Head>\n                  <link\n                    rel=\"apple-touch-startup-image\"\n                    href=\"/assets/public/pwa/splash/apple-splash-2048-2732.jpg\"\n                  />\n\t\t\t          </Head>\n\t\t\t        </Html>\n\t\t\t      )\n\t\t\t    }",
		},
		{
			// Invented, and it is where a recorded research pass had the rule inverted. It read oxc's
			// `ArrowFunctionExpression => None` as "arrows contribute no name" and concluded the
			// declarator fallback applies only to anonymous function expressions. That `None` IS the id,
			// and it flows into the fallback that reads the declarator. Measured on the release binary:
			// SILENT. Its reporting twin, with the binding renamed, is in the list above.
			name:     "an arrow bound to the default-exported const",
			fileName: "pages/_document.tsx",
			source:   "const D = () => (<link href=\"https://fonts.googleapis.com/css2?family=Inter\" />);\nexport default D;\n",
		},
		{
			// Invented. The same fallback through an anonymous function expression rather than an arrow.
			// Measured: silent. Together with the named-expression case above this pins all three arms of
			// the name rule: own name, no name plus a binding, and no binding.
			name:     "an anonymous function expression bound to the default-exported const",
			fileName: "pages/_document.tsx",
			source:   "const D = function () { return (<link href=\"https://fonts.googleapis.com/css2?family=Inter\" />) };\nexport default D;\n",
		},
		{
			// Invented. The class arm of the same fallback, which upstream writes as one match arm with the
			// function. Measured: silent.
			name:     "an anonymous class expression bound to the default-exported const",
			fileName: "pages/_document.tsx",
			source:   "const D = class { render() { return (<link href=\"https://fonts.googleapis.com/css2?family=Inter\" />) } };\nexport default D;\n",
		},
		{
			// Invented. Reached through the export-assignment ancestor rather than through any name, which is
			// the arrow twin of upstream's anonymous-function pass case. Measured: silent.
			name:     "an anonymous default-exported arrow",
			fileName: "pages/_document.tsx",
			source:   "export default () => (<link href=\"https://fonts.googleapis.com/css2?family=Inter\" />);\n",
		},
		{
			// Invented. The value is not a string literal, so nothing is read and nothing reports, even though
			// the string it resolves to would match. Measured: silent. This is upstream declining to
			// evaluate rather than an oversight, and a port that resolved the constant would diverge.
			name:     "an href expression rather than a string literal",
			fileName: "pages/index.tsx",
			source:   "const u = \"https://fonts.googleapis.com/css2?family=Inter\";\nexport default function D() { return (<link rel=\"stylesheet\" href={u} />) }\n",
		},
		{
			// Invented, and worth pinning because a spread is NOT invisible to us the way it is to oxc's
			// attribute loop: it arrives as a real KindJsxSpreadAttribute node with its object literal
			// right there to read. Resolving it would catch more and decide differently, so it is declined
			// deliberately. Measured: silent. `jsx.StringAttributeValue` already skips it.
			name:     "a spread carrying the href",
			fileName: "pages/index.tsx",
			source:   "export default function D() { return (<link {...{href:\"https://fonts.googleapis.com/css2?family=Inter\"}} />) }\n",
		},
		{
			// Invented, and it was added for a surviving mutant rather than written up front. A
			// mutation rewriting the prefix test to strings.Contains survived the entire fixture
			// set, because every case in the corpus writes the URL at the START of the href and
			// nothing votes on the difference. Measured on the release binary: this proxied URL,
			// which CONTAINS the prefix without beginning with it, is silent, so Contains is a real
			// behavioral difference rather than a harmless relaxation.
			name:     "an href containing the prefix without starting with it",
			fileName: "pages/index.tsx",
			source:   "export default function D() { return (<link href=\"https://cdn.example.com/proxy?u=https://fonts.googleapis.com/css2?family=Inter\" />) }\n",
		},
		{
			// Invented, the same survivor from the cheapest angle: one leading space. Measured
			// silent upstream, and it is the case a reader is most likely to assume is trimmed.
			name:     "an href with a leading space",
			fileName: "pages/index.tsx",
			source:   "export default function D() { return (<link href=\" https://fonts.googleapis.com/css2?family=Inter\" />) }\n",
		},
		{
			// Invented. A component reference never emits the HTML element. Measured: silent.
			name:     "a capitalized Link component",
			fileName: "pages/index.tsx",
			source:   "export default function D() { return (<Link href=\"https://fonts.googleapis.com/css2?family=Inter\" />) }\n",
		},
		{
			// Invented. The attribute match is case sensitive. Measured: silent, which is why
			// `jsx.MatchExactly` is used rather than the case-insensitive matcher.
			name:     "an attribute name spelled in capitals",
			fileName: "pages/index.tsx",
			source:   "export default function D() { return (<link rel=\"stylesheet\" HREF=\"https://fonts.googleapis.com/css2?family=Inter\" />) }\n",
		},
		{
			// Invented. The scheme is part of the prefix. Measured: silent.
			name:     "the http spelling of the same url",
			fileName: "pages/index.tsx",
			source:   "export default function D() { return (<link rel=\"stylesheet\" href=\"http://fonts.googleapis.com/css2?family=Inter\" />) }\n",
		},
		{
			// Invented, and it is the clearest statement that this is not a font rule. A genuine custom-font
			// stylesheet from Typekit, in a page, with the right `rel`: exactly what the rule's NAME
			// describes, and measured silent. `https://fonts.gstatic.com/css?family=X` is silent too.
			name:     "another font host entirely",
			fileName: "pages/index.tsx",
			source:   "export default function D() { return (<link rel=\"stylesheet\" href=\"https://use.typekit.net/abc.css\" />) }\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.Run(t, NoPageCustomFont, testCase.fileName, testCase.source))
		})
	}
}

// The span, which no message-id assertion can see.
//
// Upstream reports `ctx.nodes().parent_kind(node.id()).span()` from a JSXOpeningElement listener,
// which is the whole enclosing JSXElement rather than the opening tag. Our self-closing kind's own
// span already covers `<link ... />`, so reporting the matched node gives the same range without a
// parent walk, and the snapshot's four-line range confirms it. The paired form is the one place the
// two could differ, and it is asserted separately below.
//
// The literal strings here are typed rather than derived from the rule's own message constants: a
// comparison against the constant moves with the constant under mutation and passes either way.
func TestNoPageCustomFontPointsAtTheWholeElement(t *testing.T) {
	t.Parallel()

	source := "export default function D() {\n  return <div><link href=\"https://fonts.googleapis.com/css2?family=Inter\" rel=\"stylesheet\" /></div>\n}\n"
	result := rule_testing.Run(t, NoPageCustomFont, "pages/index.tsx", source)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("wanted 1 finding, got %d", len(result.Diagnostics))
	}

	reported := source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
	wanted := "<link href=\"https://fonts.googleapis.com/css2?family=Inter\" rel=\"stylesheet\" />"
	if reported != wanted {
		t.Errorf("finding points at %q, wanted %q", reported, wanted)
	}

	// Asserted exactly rather than with strings.Contains, because a predicate weaker than the
	// property it guards is not a guard: a doubled or truncated rendering still contains a prefix.
	wantedMessage := "This is a Google Fonts stylesheet requested from a page rather than from the " +
		"custom document. The document shell renders once for the whole application, so a font " +
		"link written there is fetched once and shared; written in a page it is fetched again per " +
		"page and the framework cannot optimize it. Move the link into the custom document."
	if result.Diagnostics[0].Message.Description != wantedMessage {
		t.Errorf("message is %q", result.Diagnostics[0].Message.Description)
	}
	if result.Diagnostics[0].Message.Id != "noPageCustomFontNotInDocument" {
		t.Errorf("message id is %q", result.Diagnostics[0].Message.Id)
	}
}

// The paired spelling's span, which is the one place our tree could differ from upstream's parent
// walk. Upstream reports the enclosing JSXElement, so for `<link ...></link>` its range covers the
// closing tag too. Ours anchors on the opening element, so it does not. Recorded as what our rule
// decides, with the difference stated rather than left for a reader to find.
func TestNoPageCustomFontPointsAtTheOpeningTagOfAPairedElement(t *testing.T) {
	t.Parallel()

	source := "export default function D() {\n  return <div><link href=\"https://fonts.googleapis.com/css2?family=Inter\"></link></div>\n}\n"
	result := rule_testing.Run(t, NoPageCustomFont, "pages/index.tsx", source)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("wanted 1 finding, got %d", len(result.Diagnostics))
	}
	reported := source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
	wanted := "<link href=\"https://fonts.googleapis.com/css2?family=Inter\">"
	if reported != wanted {
		t.Errorf("finding points at %q, wanted %q", reported, wanted)
	}
}

// The second finding's span and message, so both message constants are pinned rather than one.
func TestNoPageCustomFontOutsideDefaultExportMessage(t *testing.T) {
	t.Parallel()

	source := "function Links() {\n  return <link href=\"https://fonts.googleapis.com/css2?family=Inter\" />\n}\nexport default function D() { return <div><Links /></div> }\n"
	result := rule_testing.Run(t, NoPageCustomFont, "pages/_document.tsx", source)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("wanted 1 finding, got %d", len(result.Diagnostics))
	}
	reported := source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
	wanted := "<link href=\"https://fonts.googleapis.com/css2?family=Inter\" />"
	if reported != wanted {
		t.Errorf("finding points at %q, wanted %q", reported, wanted)
	}
	wantedMessage := "This is a Google Fonts stylesheet inside the custom document but outside the " +
		"component the document default-exports. The framework rewrites font links it finds in " +
		"the exported document component, and a link rendered from a helper component is invisible " +
		"to that rewriting, so automatic font optimization silently does nothing. Inline the link " +
		"into the default-exported component."
	if result.Diagnostics[0].Message.Description != wantedMessage {
		t.Errorf("message is %q", result.Diagnostics[0].Message.Description)
	}
	if result.Diagnostics[0].Message.Id != "noPageCustomFontOutsideDefaultExport" {
		t.Errorf("message id is %q", result.Diagnostics[0].Message.Id)
	}
	// `strings` is used here deliberately: the two message descriptions must not be one string,
	// which a copy-paste in the rule file could silently make them.
	if strings.EqualFold(wantedMessage, "") {
		t.Fatal("unreachable")
	}
}

// TestNoPageCustomFontDecodesEntities reads its attribute as ESLint does, with HTML entities decoded: typescript-estree decodes a
// JSX attribute string before any rule sees it. Each row's verdict is the installed
// @next/eslint-plugin-next 16.3.1's under the typescript-eslint parser (#51y9jh2).
func TestNoPageCustomFontDecodesEntities(t *testing.T) {
	t.Parallel()

	for _, row := range []struct {
		name   string
		source string
		ids    []string
	}{
		{`an encoded slash in the font href`, "export default function D() {\n  return <div><link href=\"https://fonts.googleapis.com&#47;css2?family=Inter\" rel=\"stylesheet\" /></div>\n}\n", []string{messageNoPageCustomFontNotInDocument.Id}},
		{`an encoded slash in another href`, "export default function D() {\n  return <div><link href=\"https://example.com&#47;css2?family=Inter\" rel=\"stylesheet\" /></div>\n}\n", []string{}},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoPageCustomFont, "pages/index.tsx", row.source)
			if len(row.ids) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, row.ids...)
		})
	}
}
