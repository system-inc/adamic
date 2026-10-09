package next

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

func TestNoScriptComponentInHeadReports(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
		count  int
	}{
		{
			// Verbatim from oxc's tester block, bytes unchanged.
			name:   "upstream fail: Script inside an imported Head",
			source: "\n\t\t\timport Head from \"next/head\";\n\t\t\timport Script from \"next/script\";\n\n\t\t\texport default function Index() {\n\t\t\t    return (\n\t\t\t        <Head>\n\t\t\t            <Script></Script>\n\t\t\t        </Head>\n\t\t\t    );\n\t\t    }\n        ",
			count:  1,
		},
		{
			// The only form real code writes, and upstream has no fixture for it. Our parser gives a
			// self closing element its own kind, so a port scanning only KindJsxElement children
			// passes both upstream cases and reports on nothing anybody writes. Measured reporting
			// on the release binary.
			name:   "self closing Script child",
			source: "import Head from \"next/head\";\nimport Script from \"next/script\";\nexport default function Index() { return (<Head><Script src=\"/a.js\" /></Head>); }\n",
			count:  1,
		},
		{
			// The entire justification for resolving the head side rather than matching the text
			// `Head`. Upstream reports; a name matching port is silent and passes every imported
			// case.
			name:   "aliased default import of next/head",
			source: "import H from \"next/head\";\nimport Script from \"next/script\";\nexport default function Index() { return (<H><Script></Script></H>); }\n",
			count:  1,
		},
		{
			// Upstream never resolves `next/script`, so this reports even though no framework script
			// is involved. It looks like a defect and it is reproduced, because the alternative is
			// a silent behavior change no imported fixture can see. Measured reporting on the
			// release binary.
			name:   "a local component named Script",
			source: "import Head from \"next/head\";\nconst Script = ({children}) => children;\nexport default function Index() { return (<Head><Script></Script></Head>); }\n",
			count:  1,
		},
		{
			// The diagnostic is inside the child loop with no break, so the count is per matching
			// child rather than per Head. Measured as two on the release binary.
			name:   "two Script children report twice",
			source: "import Head from \"next/head\";\nimport Script from \"next/script\";\nexport default function Index() { return (<Head><Script></Script><Script></Script></Head>); }\n",
			count:  2,
		},
		{
			// Nested heads each ask the question separately, and only the one that directly contains
			// the script answers yes. The release binary points at the inner one, column fifty
			// against forty four for the outer.
			name:   "the inner Head is the one blamed",
			source: "import Head from \"next/head\";\nimport Script from \"next/script\";\nexport default function Index() { return (<Head><Head><Script></Script></Head></Head>); }\n",
			count:  1,
		},
		{
			// The shadow is scoped, so the outer use still binds to the import. A rule that gave up on
			// the whole file after seeing any local `Head` would be silent here.
			name:   "a shadow in another function does not disarm the import",
			source: "import Head from \"next/head\";\nimport Script from \"next/script\";\nfunction Other() { const Head = ({children}) => children; return (<Head><b></b></Head>); }\nexport default function Index() { return (<Head><Script></Script></Head>); }\n",
			count:  1,
		},
		{
			// Children are a mixed list, so a walk that stopped at the first non element child would
			// miss this.
			name:   "text and expression siblings do not hide the script",
			source: "import Head from \"next/head\";\nimport Script from \"next/script\";\nexport default function Index() { return (<Head>text<Script></Script>{\"x\"}</Head>); }\n",
			count:  1,
		},
		{
			// The default binding is read out of a clause that also carries named bindings, which is
			// the shape `BindingsOf` exists to keep separate.
			name:   "default and named specifiers on the same import",
			source: "import Head, { Foo } from \"next/head\";\nimport Script from \"next/script\";\nexport default function Index() { return (<Head><Script></Script></Head>); }\n",
			count:  1,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoScriptComponentInHead, "Component.tsx", testCase.source)
			expected := make([]string, testCase.count)
			for index := range expected {
				expected[index] = messageNoScriptComponentInHead.Id
			}
			rule_testing.ExpectFindings(t, result, expected...)
		})
	}
}

func TestNoScriptComponentInHeadIsSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
	}{
		{
			// Verbatim from oxc's tester block, bytes unchanged.
			name:   "upstream pass: a locally defined Head",
			source: "\n            import Script from \"next/script\";\n\t\t\tconst Head = ({children}) => children\n\n\t\t\texport default function Index() {\n\t\t\t    return (\n\t\t\t        <Head>\n\t\t\t            <Script></Script>\n\t\t\t        </Head>\n\t\t\t    );\n\t\t\t}\n\t\t",
		},
		{
			// The specifier gate read in the shut direction. Upstream tests it only broken open, so
			// nothing in the corpus catches a port that stopped comparing the string.
			name:   "Head imported from another module",
			source: "import Head from \"other/head\";\nimport Script from \"next/script\";\nexport default function Index() { return (<Head><Script></Script></Head>); }\n",
		},
		{
			// Direct children only, one level. Pinned so a helpful deep walk fails a test instead of
			// quietly widening the rule. Measured silent on the release binary.
			name:   "Script nested one level deeper",
			source: "import Head from \"next/head\";\nimport Script from \"next/script\";\nexport default function Index() { return (<Head><div><Script></Script></div></Head>); }\n",
		},
		{
			// The other direction of the unresolved script side: this is a real framework script
			// inside a real framework head and upstream says nothing, because the tag does not spell
			// `Script`. Reproduced rather than fixed, for the reason in the rule doc.
			name:   "next/script under a different local name",
			source: "import Head from \"next/head\";\nimport S from \"next/script\";\nexport default function Index() { return (<Head><S></S></Head>); }\n",
		},
		{
			// Textually identical to the reporting form and silent upstream. This is the case that
			// makes the checker load bearing rather than an improvement.
			name:   "a local Head shadows the import",
			source: "import Head from \"next/head\";\nimport Script from \"next/script\";\nexport default function Index() {\n  const Head = ({children}) => children;\n  return (<Head><Script></Script></Head>);\n}\n",
		},
		{
			// Only a default specifier arms the rule. A namespace import binds one name for the whole
			// module and the head component is reached through a member expression, which is not an
			// identifier tag.
			name:   "namespace import of next/head",
			source: "import * as H from \"next/head\";\nimport Script from \"next/script\";\nexport default function Index() { return (<H.default><Script></Script></H.default>); }\n",
		},
		{
			// The module has no such named export, and upstream reads the default specifier alone.
			name:   "named import from next/head",
			source: "import { Head } from \"next/head\";\nimport Script from \"next/script\";\nexport default function Index() { return (<Head><Script></Script></Head>); }\n",
		},
		{
			// A second default import alongside a real `next/head` one, and the script sits under the
			// wrong one. This is the input that separates node identity from declaration kind: with
			// only one import in the file the clause set is empty and a kind test never gets to be
			// wrong, so the "Head imported from another module" case above cannot see the
			// difference. Measured silent on the release binary, and a kind comparison survived the
			// whole fixture set until this was written.
			name:   "a second import used where the head is not",
			source: "import Head from \"next/head\";\nimport Other from \"other/thing\";\nimport Script from \"next/script\";\nexport default function Index() { return (<Other><Script></Script></Other>); }\n",
		},
		{
			// A dotted head tag. Upstream matches an `IdentifierReference` element name and this is
			// not one, so it is silent, and measured silent on the release binary. Written because
			// removing the identifier guard on the head tag survived the whole fixture set: a
			// property access still resolves through the checker to the import's clause, so the
			// guard is the only thing declining it. Reading text off that kind also panics, so the
			// guard is load bearing twice over.
			name:   "a dotted head tag",
			source: "import Head from \"next/head\";\nimport Script from \"next/script\";\nexport default function Index() { return (<Head.Inner><Script></Script></Head.Inner>); }\n",
		},
		{
			// The empty reference set. Upstream walks references and finds none; here no JSX tag
			// resolves to the clause.
			name:   "next/head imported but never used",
			source: "import Head from \"next/head\";\nexport default function Index() { return (<div></div>); }\n",
		},
		{
			// A dotted tag parses as a property access rather than an identifier, so its text is never
			// compared. Measured silent on the release binary, and reading text off that kind panics.
			name:   "a member expression tag named Script",
			source: "import Head from \"next/head\";\nimport Foo from \"foo\";\nexport default function Index() { return (<Head><Foo.Script></Foo.Script></Head>); }\n",
		},
		{
			// The comparison is case sensitive against the component spelling, so plain HTML markup
			// in the head is left to no-sync-scripts.
			name:   "a lowercase intrinsic script inside Head",
			source: "import Head from \"next/head\";\nexport default function Index() { return (<Head><script src=\"/a.js\"></script></Head>); }\n",
		},
		{
			// The outer listener is on KindJsxElement alone, which a self closing element never is.
			name:   "a self closing Head has no children to search",
			source: "import Head from \"next/head\";\nimport Script from \"next/script\";\nexport default function Index() { return (<div><Head /><Script></Script></div>); }\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoScriptComponentInHead, "Component.tsx", testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestNoScriptComponentInHeadPointsAtTheHeadName asserts where the finding lands and what it says.
//
// Message ids cannot see either. The snapshot puts the caret under `Head` for four columns, which is
// the opening element's name node rather than the element and rather than the `<Script>` that caused
// the finding, and a port pointing at any of the three satisfies every id assertion above.
func TestNoScriptComponentInHeadPointsAtTheHeadName(t *testing.T) {
	t.Parallel()

	source := "import Head from \"next/head\";\nimport Script from \"next/script\";\nexport default function Index() { return (<Head><Script></Script></Head>); }\n"
	result := rule_testing.RunTyped(t, NoScriptComponentInHead, "Component.tsx", source)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("expected one finding, got %d", len(result.Diagnostics))
	}

	diagnostic := result.Diagnostics[0]
	reported := source[diagnostic.Range.Pos():diagnostic.Range.End()]
	if reported != "Head" {
		t.Errorf("finding points at %q, want %q", reported, "Head")
	}

	// Equality against a literal rather than against the message value the rule reports with. A
	// predicate weaker than the property it guards is not a guard, and comparing the finding to
	// `messageNoScriptComponentInHead` is exactly that: both sides move together, so rewriting the
	// id or the description survived a mutation sweep with every assertion still green. The literal
	// is the only thing that does not move.
	wantId := "noScriptComponentInHead"
	wantDescription := "This puts a <Script> inside <Head>. The head component only forwards plain " +
		"markup into the document head, so the script loader never mounts and the loading strategy " +
		"is dropped: nothing decides when the script runs and nothing tracks whether it already " +
		"did. Move the <Script> outside of <Head> and let it sit in the page body."
	if diagnostic.Message.Id != wantId {
		t.Errorf("message id is %q, want %q", diagnostic.Message.Id, wantId)
	}
	if diagnostic.Message.Description != wantDescription {
		t.Errorf("message text is %q, want %q", diagnostic.Message.Description, wantDescription)
	}
}

// TestNoScriptComponentInHeadBlamesTheInnerHead pins which of two nested heads is reported.
//
// Both resolve to the same import, so both reach the child scan and only the inner one finds a
// script. A rule reporting the outer would pass the count assertion in the table above while
// pointing a reader at the wrong element. The release binary reports at column fifty, which is the
// inner `Head`; the outer sits at forty four.
func TestNoScriptComponentInHeadBlamesTheInnerHead(t *testing.T) {
	t.Parallel()

	source := "import Head from \"next/head\";\nimport Script from \"next/script\";\nexport default function Index() { return (<Head><Head><Script></Script></Head></Head>); }\n"
	result := rule_testing.RunTyped(t, NoScriptComponentInHead, "Component.tsx", source)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("expected one finding, got %d", len(result.Diagnostics))
	}

	position := result.Diagnostics[0].Range.Pos()
	outer := strings.Index(source, "<Head>") + len("<")
	inner := strings.Index(source[outer:], "<Head>") + outer + len("<")
	if position == outer {
		t.Errorf("finding blames the outer Head at %d, want the inner one at %d", position, inner)
	}
	if position != inner {
		t.Errorf("finding is at %d, want the inner Head at %d", position, inner)
	}
}

// TestNoScriptComponentInHeadRequiresTheTypedHarness fails loudly if the checker declaration is
// reverted.
//
// The plain harness hands a rule a nil checker, so this rule would go completely silent: every
// silent fixture above would pass vacuously and every reporting one would fail in a way that reads
// as a rule bug rather than as a harness mismatch. Asserting the silence directly names the cause.
func TestNoScriptComponentInHeadRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	if !NoScriptComponentInHead.NeedsTypeChecker {
		t.Fatal("the rule resolves tags through the checker and must declare NeedsTypeChecker")
	}

	source := "import Head from \"next/head\";\nimport Script from \"next/script\";\nexport default function Index() { return (<Head><Script></Script></Head>); }\n"
	result := rule_testing.Run(t, NoScriptComponentInHead, "Component.tsx", source)
	if len(result.Diagnostics) != 0 {
		t.Fatalf("untyped harness produced %d findings, so the rule is no longer resolving", len(result.Diagnostics))
	}
}
