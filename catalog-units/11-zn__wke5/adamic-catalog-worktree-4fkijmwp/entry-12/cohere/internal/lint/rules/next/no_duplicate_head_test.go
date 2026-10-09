package next

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// The four cases oxc ships, copied verbatim through the extractor's own dump and re-quoted by
// strconv so nothing on the path from the Rust source to this file could cook an escape. Both fail
// cases import `Head` as a *default* from `next/head` while both pass cases import it *named* from
// `next/document`, so a port implementing only one binding form passes one half and fails the other.
const upstreamPass1 = "import Document, { Html, Head, Main, NextScript } from 'next/document'\n\n\t\t\t      class MyDocument extends Document {\n\t\t\t        static async getInitialProps(ctx) {\n\t\t\t          //...\n\t\t\t        }\n\n\t\t\t        render() {\n\t\t\t          return (\n\t\t\t            <Html>\n\t\t\t              <Head/>\n\t\t\t            </Html>\n\t\t\t          )\n\t\t\t        }\n\t\t\t      }\n\n\t\t\t      export default MyDocument\n\t\t\t    "

const upstreamPass2 = "import Document, { Html, Head, Main, NextScript } from 'next/document'\n\n\t\t\t      class MyDocument extends Document {\n\t\t\t        render() {\n\t\t\t          return (\n\t\t\t            <Html>\n\t\t\t              <Head>\n\t\t\t                <meta charSet=\"utf-8\" />\n\t\t\t                <link\n\t\t\t                  href=\"https://fonts.googleapis.com/css2?family=Sarabun:ital,wght@0,400;0,700;1,400;1,700&display=swap\"\n\t\t\t                  rel=\"stylesheet\"\n\t\t\t                />\n\t\t\t              </Head>\n\t\t\t            </Html>\n\t\t\t          )\n\t\t\t        }\n\t\t\t      }\n\n\t\t\t      export default MyDocument\n\t\t\t    "

const upstreamFail1 = "\n\t\t\t      import Document, { Html, Main, NextScript } from 'next/document'\n\t\t\t      import Head from 'next/head'\n\n\t\t\t      class MyDocument extends Document {\n\t\t\t        render() {\n\t\t\t          return (\n\t\t\t            <Html>\n\t\t\t              <Head />\n\t\t\t              <Head />\n\t\t\t              <Head />\n\t\t\t            </Html>\n\t\t\t          )\n\t\t\t        }\n\t\t\t      }\n\n\t\t\t      export default MyDocument\n\t\t\t      "

const upstreamFail2 = "\n\t\t\t      import Document, { Html, Main, NextScript } from 'next/document'\n\t\t\t      import Head from 'next/head'\n\n\t\t\t      class MyDocument extends Document {\n\t\t\t        render() {\n\t\t\t          return (\n\t\t\t            <Html>\n\t\t\t              <Head>\n\t\t\t                <meta charSet=\"utf-8\" />\n\t\t\t                <link\n\t\t\t                  href=\"https://fonts.googleapis.com/css2?family=Sarabun:ital,wght@0,400;0,700;1,400;1,700&display=swap\"\n\t\t\t                  rel=\"stylesheet\"\n\t\t\t                />\n\t\t\t              </Head>\n\t\t\t              <body>\n\t\t\t                <Main />\n\t\t\t                <NextScript />\n\t\t\t              </body>\n\t\t\t              <Head>\n\t\t\t                <script\n\t\t\t                  dangerouslySetInnerHTML={{\n\t\t\t                    __html: '',\n\t\t\t                  }}\n\t\t\t                />\n\t\t\t              </Head>\n\t\t\t            </Html>\n\t\t\t          )\n\t\t\t        }\n\t\t\t      }\n\n\t\t\t      export default MyDocument\n\t\t\t      "

func TestNoDuplicateHeadReports(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		fileName string
		source   string
	}{
		{
			// Upstream fail 1. Three copies, self-closing, default import from `next/head`.
			name:     "three self closing copies",
			fileName: "pages/_document.tsx",
			source:   upstreamFail1,
		},
		{
			// Upstream fail 2. Two copies with children, separated by a `body` element, so the two
			// occurrences are not adjacent siblings.
			name:     "two copies with children separated",
			fileName: "pages/_document.tsx",
			source:   upstreamFail2,
		},
		{
			// Not upstream. The module specifier is never consulted, so a component named `Head`
			// imported from anywhere at all arms the rule. Measured reporting on the release
			// binary. This is upstream firing outside the scope its own message describes and it is
			// reproduced deliberately; see the rule's doc comment.
			name:     "imported from an unrelated module",
			fileName: "components/Shell.tsx",
			source:   "import { Head } from './my-own-widgets'\nconst a = <div><Head/><Head/></div>\n",
		},
		{
			// Not upstream, and the axis the recorded research had backwards. oxc matches the
			// specifier's LOCAL name, so an import whose *imported* name is `Header` but which
			// binds locally to `Head` reports. Measured reporting.
			name:     "aliased so the local name is Head",
			fileName: "components/Shell.tsx",
			source:   "import { Header as Head } from 'next/document'\nconst a = <div><Head/><Head/></div>\n",
		},
		{
			// Not upstream. The count spans the whole file rather than one returned tree, so two
			// unrelated components each rendering one `Head` report. ESLint is silent here because
			// it filters the direct children of a single return statement. Measured reporting.
			name:     "two separate components in one file",
			fileName: "components/Shell.tsx",
			source:   "import { Head } from 'next/document'\nconst a = <div><Head/></div>\nconst b = <div><Head/></div>\n",
		},
		{
			// Not upstream. Any depth, not just direct children of the returned element. This is
			// the sharpest disagreement with ESLint, which is silent. Measured reporting.
			name:     "nested at different depths",
			fileName: "components/Shell.tsx",
			source:   "import { Head } from 'next/document'\nconst a = <div><span><Head/></span><Head/></div>\n",
		},
		{
			// Not upstream. A namespace import binds a local `Head` and arms the rule the same way
			// a default or named one does, because oxc's specifier list is flat across all three.
			// Measured reporting.
			name:     "namespace import",
			fileName: "components/Shell.tsx",
			source:   "import * as Head from 'next/head'\nconst a = <div><Head/><Head/></div>\n",
		},
		{
			// Not upstream. A non-JSX value reference is skipped rather than counted, so this
			// reports for the two tags alone rather than for three references. The count is pinned
			// by the span and count assertions below rather than here.
			name:     "a value reference alongside two tags",
			fileName: "components/Shell.tsx",
			source:   "import { Head } from 'next/document'\nconst x = Head\nconst a = <div><Head/><Head/></div>\n",
		},
		{
			// Not upstream, and it exists because a mutant reading only `Declarations[0]` survived
			// every other fixture here. Declaration merging gives one symbol two declarations, and
			// the checker lists them in source order, so writing the interface FIRST puts the
			// import at index 1 and a rule reading only the first declaration goes silent on an
			// input the release binary reports. Measured both ways: the binary reports, and the
			// probe confirmed the ordering flips with the source order.
			name:     "declaration merged with an interface written first",
			fileName: "components/Shell.tsx",
			source:   "interface Head { x: number }\nimport { Head } from 'next/document'\nconst a = <div><Head/><Head/></div>\n",
		},
		{
			// Not upstream. The corpus ships three-self-closing and two-with-children; this is the
			// missing diagonal, two self-closing, which is the smallest input that reports at all.
			name:     "exactly two self closing",
			fileName: "components/Shell.tsx",
			source:   "import { Head } from 'next/document'\nconst a = <div><Head/><Head/></div>\n",
		},
		{
			// Not upstream. The other missing diagonal, three with children, which also pins that a
			// closing tag is not counted a second time: six tags, three occurrences, one finding.
			name:     "three copies with children",
			fileName: "components/Shell.tsx",
			source:   "import { Head } from 'next/document'\nconst a = <div><Head>a</Head><Head>b</Head><Head>c</Head></div>\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoDuplicateHead, testCase.fileName, testCase.source)
			rule_testing.ExpectFindings(t, result, messageNoDuplicateHead.Id)
		})
	}
}

func TestNoDuplicateHeadIsSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		fileName string
		source   string
	}{
		{
			// Upstream pass 1. Exactly one `Head`, self-closing.
			name:     "one self closing copy",
			fileName: "pages/_document.tsx",
			source:   upstreamPass1,
		},
		{
			// Upstream pass 2. Exactly one `Head` carrying children.
			name:     "one copy with children",
			fileName: "pages/_document.tsx",
			source:   upstreamPass2,
		},
		{
			// Not upstream. The other half of the alias axis, and the one that catches a port
			// matching the imported name instead of the local one: the imported name is `Head` but
			// nothing local is called `Head`, so no specifier matches. Measured silent.
			name:     "aliased away from Head",
			fileName: "components/Shell.tsx",
			source:   "import { Head as PageHead } from 'next/document'\nconst a = <div><PageHead/><PageHead/></div>\n",
		},
		{
			// Not upstream. Not an import, so upstream's `flags.is_import()` guard declines it.
			// Measured silent. A syntactic name-matching port reports here.
			name:     "locally declared rather than imported",
			fileName: "components/Shell.tsx",
			source:   "const Head = () => null\nconst a = <div><Head/><Head/></div>\n",
		},
		{
			// Not upstream. The usages resolve to a function-scoped `Head` rather than to the
			// import, so identity declines them. Measured silent, and this is what makes symbol
			// identity load-bearing rather than an optimization over text matching.
			name:     "shadowed inside a function",
			fileName: "components/Shell.tsx",
			source:   "import { Head } from 'next/document'\nfunction f() { const Head = () => null; return <div><Head/><Head/></div> }\n",
		},
		{
			// Not upstream. One root usage plus two shadowed ones is one counted occurrence, not
			// three. Measured silent, and it is the input that separates identity from a root-scope
			// text count.
			name:     "one root usage and two shadowed",
			fileName: "components/Shell.tsx",
			source:   "import { Head } from 'next/document'\nconst a = <div><Head/></div>\nfunction f() { const Head = () => null; return <div><Head/><Head/></div> }\n",
		},
		{
			// Not upstream. A member-expression tag never counts, because the reference's parent is
			// the property access rather than the opening element. Measured silent, and reading
			// text off that node kind panics, so the guard is doubly load-bearing.
			name:     "member expression tags",
			fileName: "components/Shell.tsx",
			source:   "import { Head } from 'next/document'\nconst a = <div><Head.Sub/><Head.Sub/></div>\n",
		},
		{
			// Not upstream. One plain tag plus one member tag is one occurrence. Measured silent,
			// which pins that the member form contributes nothing rather than merely being unable
			// to start the count.
			name:     "one plain tag and one member tag",
			fileName: "components/Shell.tsx",
			source:   "import { Head } from 'next/document'\nconst a = <div><Head/><Head.Sub/></div>\n",
		},
		{
			// Not upstream. A lowercase intrinsic is a different tag and resolves to no import.
			// Measured silent.
			name:     "lowercase intrinsic head elements",
			fileName: "components/Shell.tsx",
			source:   "import { Head } from 'next/document'\nconst a = <div><head/><head/></div>\n",
		},
		{
			// Not upstream. A type-only import binds no value. Measured silent.
			name:     "type only import",
			fileName: "components/Shell.tsx",
			source:   "import type { Head } from 'next/document'\nconst a = <div><Head/><Head/></div>\n",
		},
		{
			// Not upstream. A single occupied `Head` with children is one occurrence even though it
			// writes both an opening and a closing tag. Measured silent, and it is the case that
			// would report if closing tags were counted.
			name:     "one copy with children is one occurrence",
			fileName: "components/Shell.tsx",
			source:   "import { Head } from 'next/document'\nconst a = <div><Head>x</Head></div>\n",
		},
		{
			// Not upstream. No import at all, so the rule registers no listener.
			name:     "no import present",
			fileName: "components/Shell.tsx",
			source:   "const a = <div><Head/><Head/></div>\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoDuplicateHead, testCase.fileName, testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestNoDuplicateHeadReportsOncePerFile pins the count, which is the question the corpus cannot
// answer and the one a reader is most likely to get wrong.
//
// oxc accumulates labels and calls `ctx.diagnostic` once, so N copies produce ONE finding carrying N
// underlines. The two other readings are both natural designs: one finding per extra copy, which is
// what `jsx-no-duplicate-props` in this tree does for a structurally similar question, and one per
// pair. All three agree on a two-copy input, which is why the corpus cannot separate them, and they
// disagree at three: one, two, and three respectively. Measured on the release oxlint binary at
// three copies and again at four, both producing a single diagnostic.
func TestNoDuplicateHeadReportsOncePerFile(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
	}{
		{
			name:   "two copies",
			source: "import { Head } from 'next/document'\nconst a = <div><Head/><Head/></div>\n",
		},
		{
			name:   "three copies",
			source: "import { Head } from 'next/document'\nconst a = <div><Head/><Head/><Head/></div>\n",
		},
		{
			name:   "four copies",
			source: "import { Head } from 'next/document'\nconst a = <div><Head/><Head/><Head/><Head/></div>\n",
		},
		{
			name:   "five copies across three components",
			source: "import { Head } from 'next/document'\nconst a = <div><Head/><Head/></div>\nconst b = <div><Head/></div>\nconst c = <div><Head/><Head/></div>\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoDuplicateHead, "components/Shell.tsx", testCase.source)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("expected exactly one finding however many copies, got %d", len(result.Diagnostics))
			}
		})
	}
}

// TestNoDuplicateHeadPointsAtTheFirstOccurrence pins where the finding lands.
//
// `ExpectFindings` asserts ids and count and nothing else, so a rule pointing at the last occurrence
// rather than the first, or at the whole element rather than its tag name, passes every fixture
// above. oxc's diagnostic renders at the first `<Head`: the snapshot for the three-copy fail prints
// `9:19`, and the underline is four bytes wide, the tag name alone.
func TestNoDuplicateHeadPointsAtTheFirstOccurrence(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
		want   string
		// offset is where the reported text is expected to start, so a rule pointing at a later
		// occurrence of the same four bytes is caught rather than passing on the text alone.
		offset int
	}{
		{
			name:   "three self closing copies",
			source: "import { Head } from 'next/document'\nconst a = <div><Head/><Head/><Head/></div>\n",
			want:   "Head",
			offset: 53,
		},
		{
			name:   "two copies with children",
			source: "import { Head } from 'next/document'\nconst a = <div><Head>x</Head><Head>y</Head></div>\n",
			want:   "Head",
			offset: 53,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoDuplicateHead, "components/Shell.tsx", testCase.source)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("expected one finding, got %d", len(result.Diagnostics))
			}
			diagnostic := result.Diagnostics[0]
			reported := testCase.source[diagnostic.Range.Pos():diagnostic.Range.End()]
			if reported != testCase.want {
				t.Errorf("reported text = %q, want %q", reported, testCase.want)
			}
			if diagnostic.Range.Pos() != testCase.offset {
				t.Errorf("reported at offset %d, want %d (the FIRST occurrence)", diagnostic.Range.Pos(), testCase.offset)
			}
		})
	}
}

// TestNoDuplicateHeadMessageText asserts the rendered description exactly.
//
// A `strings.Contains` predicate over an interpolated message is weaker than the property it
// guards, so equality is asserted here. This message interpolates nothing, which is itself worth
// pinning: a later edit adding a name into it would need this assertion updated deliberately.
func TestNoDuplicateHeadMessageText(t *testing.T) {
	t.Parallel()

	source := "import { Head } from 'next/document'\nconst a = <div><Head/><Head/></div>\n"
	result := rule_testing.RunTyped(t, NoDuplicateHead, "components/Shell.tsx", source)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("expected one finding, got %d", len(result.Diagnostics))
	}
	if got := result.Diagnostics[0].Message.Description; got != messageNoDuplicateHead.Description {
		t.Errorf("description = %q, want %q", got, messageNoDuplicateHead.Description)
	}
	if !strings.HasPrefix(messageNoDuplicateHead.Description, "This file renders the imported `Head` component more than once.") {
		t.Errorf("description no longer opens by saying what is wrong: %q", messageNoDuplicateHead.Description)
	}
}

// TestNoDuplicateHeadNeedsTheTypedHarness pins the checker declaration.
//
// The plain harness hands the rule a nil checker, so every tag resolves to no symbol and the rule
// goes completely silent: every StaysSilent case above would pass vacuously and every Fires case
// would fail looking like a rule defect. A later revert of `NeedsTypeChecker` fails here loudly
// instead.
func TestNoDuplicateHeadNeedsTheTypedHarness(t *testing.T) {
	t.Parallel()

	if !NoDuplicateHead.NeedsTypeChecker {
		t.Fatal("rule must declare NeedsTypeChecker; symbol identity is what separates a shadow from the import")
	}
}
