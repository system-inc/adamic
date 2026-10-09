package typescript

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const unsafeDeclarationMergingFile = "/repository/source/Merged.ts"

// TestNoUnsafeDeclarationMergingFires carries oxc's three failing inputs verbatim, each of which the
// snapshot records as producing exactly one diagnostic.
//
// One is the count worth noticing rather than the obvious two. Both listeners see the same pair, and
// only the one visiting the LATER declaration reports, because the earlier declaration is the one
// the symbol records first and the opposite-kind test is asked against it. The rule note explains
// why that asymmetry is upstream's decision rather than an artifact.
func TestNoUnsafeDeclarationMergingFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"an interface then a class", "\n            interface Foo {}\n            class Foo {}\n                  "},
		{"a class then an interface", "\n                     class Foo {}\n                     interface Foo {}\n                           "},
		{"an interface and a class inside declare global", "\n                     declare global {\n                       interface Foo {}\n                       class Foo {}\n                     }\n                           "},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoUnsafeDeclarationMerging,
				unsafeDeclarationMergingFile, testCase.sourceText), "unsafeMerging")
		})
	}
}

// TestNoUnsafeDeclarationMergingStaysSilent carries oxc's eight passing inputs verbatim.
//
// Three of them are about scope rather than about kinds, and they are the ones that would be easy to
// invent wrongly: a class returned from a function body, a class inside an immediately invoked
// function, and an interface inside `declare global` beside a class outside it. Nothing in this rule
// implements scoping; typescript-go's binder gives each of those its own symbol, and the
// opposite-kind test then finds a single declaration of the same kind as the node. These fixtures are
// what pins that, since the rule reads as though it would merge them.
func TestNoUnsafeDeclarationMergingStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"an interface implemented by a differently named class", "\n            interface Foo {}\n            class Bar implements Foo {}\n                "},
		{"two namespaces", "\n                     namespace Foo {}\n                     namespace Foo {}\n                         "},
		{"an enum and a namespace", "\n                     enum Foo {}\n                     namespace Foo {}\n                         "},
		{"a namespace and a differently named function", "\n                     namespace Fooo {}\n                     function Foo() {}\n                         "},
		{"an anonymous class expression", "\n                     const Foo = class {};\n                         "},
		{"a class returned from a function body", "\n                     interface Foo {\n                       props: string;\n                     }\n\n                     function bar() {\n                       return class Foo {};\n                     }\n                         "},
		{"a class inside an immediately invoked function", "\n                     interface Foo {\n                       props: string;\n                     }\n\n                     (function bar() {\n                       class Foo {}\n                     })();\n                         "},
		{"an interface in declare global and a class outside it", "\n                     declare global {\n                       interface Foo {}\n                     }\n\n                     class Foo {}\n                         "},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoUnsafeDeclarationMerging,
				unsafeDeclarationMergingFile, testCase.sourceText))
		})
	}
}

// TestNoUnsafeDeclarationMergingCountsFindingsPerDeclaration pins the asymmetry the corpus cannot
// show, because every upstream failing input holds exactly one class and one interface.
//
// Both verdicts were measured on the release oxlint binary before being written here, per the rule
// that a hypothesis about upstream is not evidence about upstream. They are the two directions of
// the same mechanism: only the declarations that are NOT first report, and only when the first one
// is of the opposite kind. So three interfaces after a class produce two findings, and two
// interfaces before a class produce one.
func TestNoUnsafeDeclarationMergingCountsFindingsPerDeclaration(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		want       []string
	}{
		{
			"two interfaces after a class, both reporting",
			"class Foo {}\ninterface Foo {}\ninterface Foo {}\nexport { Foo };\n",
			[]string{"unsafeMerging", "unsafeMerging"},
		},
		{
			"two interfaces before a class, only the class reporting",
			"interface Foo {}\ninterface Foo {}\nclass Foo {}\nexport { Foo };\n",
			[]string{"unsafeMerging"},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoUnsafeDeclarationMerging,
				unsafeDeclarationMergingFile, testCase.sourceText), testCase.want...)
		})
	}
}

// TestNoUnsafeDeclarationMergingGoesSilentWhenAThirdKindIsFirst is upstream's real limitation,
// reproduced deliberately and pinned so a later reader cannot quietly correct it.
//
// A namespace or an import written above the pair leaves the class and the interface genuinely
// merged and genuinely unsafe, and upstream reports nothing, because the declaration it consults is
// the first one and that one is neither kind. Both were measured on the release binary at zero
// findings. typescript-eslint reports on both, since it scans every definition rather than the
// first; the rule note carries that comparison.
//
// If this test ever fails, the port has been improved rather than broken, and the question is
// whether the differential harness wants the improvement, not whether the rule is wrong.
func TestNoUnsafeDeclarationMergingGoesSilentWhenAThirdKindIsFirst(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a namespace declared first", "namespace Foo { export const value = 1; }\ninterface Foo {}\nclass Foo {}\nexport { Foo };\n"},
		{"an import declared first", "import { Foo } from './Other';\ninterface Foo {}\nclass Foo {}\nexport { Foo };\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoUnsafeDeclarationMerging,
				unsafeDeclarationMergingFile, testCase.sourceText))
		})
	}
}

// TestNoUnsafeDeclarationMergingStaysSilentOnUnpairedKinds covers the pairings the corpus leaves to
// inference, each measured on the release binary at zero findings rather than reasoned from the
// rule's shape.
//
// The rule reads as "two declarations of one name", and it is not: it is "a class and an interface",
// specifically. An interface beside a function, a namespace beside a class, and an interface beside
// an interface are all merges and all silent.
func TestNoUnsafeDeclarationMergingStaysSilentOnUnpairedKinds(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"an interface and a function", "interface Foo {}\nfunction Foo() {}\nexport { Foo };\n"},
		{"a namespace and a class", "namespace Foo { export const value = 1; }\nclass Foo {}\nexport { Foo };\n"},
		{"a class and a namespace", "class Foo {}\nnamespace Foo { export const value = 1; }\nexport { Foo };\n"},
		{"two interfaces", "interface Foo {}\ninterface Foo {}\nexport { Foo };\n"},
		// Different scopes, which is the case the rule does not implement and gets from the binder.
		// A block-scoped class has its own symbol carrying one declaration, so the opposite-kind
		// test sees a class where it needed an interface.
		{"a class in a nested block", "interface Foo {}\n{\n    class Foo {}\n}\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoUnsafeDeclarationMerging,
				unsafeDeclarationMergingFile, testCase.sourceText))
		})
	}
}

// TestNoUnsafeDeclarationMergingPointsAtTheName asserts where the finding lands and what it says,
// neither of which ExpectFindings can see.
//
// The literals here are typed out rather than read off the rule's own constants, deliberately. A
// comparison against `messageUnsafeDeclarationMerging.Id` is an equality that looks correct and
// guards nothing, because a mutation to the constant moves both sides of it at once. These strings
// are the independent copy that makes the assertion mean something.
func TestNoUnsafeDeclarationMergingPointsAtTheName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		sourceText   string
		wantReported []string
	}{
		// Both candidate spans hold the same text, so this half only proves the finding lands on a
		// name rather than on a whole declaration or a keyword. Which of the two names it is, is
		// decided by offset in the test below.
		{"an interface then a class", "interface Foo {}\nclass Foo {}\nexport { Foo };\n", []string{"Foo"}},
		{"a class then an interface", "class Bar {}\ninterface Bar {}\nexport { Bar };\n", []string{"Bar"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUnsafeDeclarationMerging, unsafeDeclarationMergingFile, testCase.sourceText)
			if len(result.Diagnostics) != len(testCase.wantReported) {
				t.Fatalf("got %d findings, want %d", len(result.Diagnostics), len(testCase.wantReported))
			}
			for index, diagnostic := range result.Diagnostics {
				reported := testCase.sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]
				if reported != testCase.wantReported[index] {
					t.Errorf("finding %d covers %q, want %q", index, reported, testCase.wantReported[index])
				}
				if diagnostic.Message.Id != "unsafeMerging" {
					t.Errorf("finding %d has id %q, want %q", index, diagnostic.Message.Id, "unsafeMerging")
				}
			}
		})
	}
}

// TestNoUnsafeDeclarationMergingPointsAtTheFirstDeclaration is the offset half of the span
// assertion, and it is the whole reason the span was measured rather than assumed.
//
// The two candidate spans hold the SAME TEXT, `Foo` either way, so the text comparison above passes
// over whichever node the rule picked and cannot tell them apart. Only the position separates them.
//
// Upstream attaches two labels and sorts them, so its primary is the declaration written FIRST,
// while the listener arm that fires is always visiting the SECOND. Reporting the visited node is the
// natural way to write this rule and puts every finding one declaration too late. Pinned on the
// release binary at 1:7 for `class Bar {}; interface Bar {}` and 1:11 for
// `interface Foo {}; class Foo {}`, which is the first declaration in both directions.
//
// Both offsets below are computed from the source by hand rather than read back from the rule, so
// the assertion has an independent copy of the answer.
func TestNoUnsafeDeclarationMergingPointsAtTheFirstDeclaration(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		sourceText   string
		wantPosition int
		wrongAt      int
	}{
		// `interface Foo` puts its name at 10; `class Foo` on the next line puts its name at 23.
		{"the interface written first", "interface Foo {}\nclass Foo {}\nexport { Foo };\n", 10, 23},
		// `class Bar` puts its name at 6; `interface Bar` on the next line puts its name at 23.
		{"the class written first", "class Bar {}\ninterface Bar {}\nexport { Bar };\n", 6, 23},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUnsafeDeclarationMerging, unsafeDeclarationMergingFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("got %d findings, want 1", len(result.Diagnostics))
			}
			if got := result.Diagnostics[0].Range.Pos(); got != testCase.wantPosition {
				t.Errorf("finding starts at %d, want %d (the second declaration's name is at %d)",
					got, testCase.wantPosition, testCase.wrongAt)
			}
		})
	}
}

// TestNoUnsafeDeclarationMergingNeedsTheTypedHarness states the dependency as an executable claim.
//
// The rule resolves through the checker, and the plain harness hands it nil. Without this, a later
// change dropping NeedsTypeChecker would leave every Fires case failing and every StaysSilent case
// passing vacuously, and only the Fires half announces itself. Measured rather than assumed:
// `node.Symbol()` is nil for all of these inputs under the untyped harness, because typescript-go
// populates symbol tables when the program is built.
func TestNoUnsafeDeclarationMergingNeedsTheTypedHarness(t *testing.T) {
	t.Parallel()

	if !NoUnsafeDeclarationMerging.NeedsTypeChecker {
		t.Fatal("the rule resolves through the checker and must declare NeedsTypeChecker")
	}
	const sourceText = "interface Foo {}\nclass Foo {}\nexport { Foo };\n"
	if result := rule_testing.Run(t, NoUnsafeDeclarationMerging, unsafeDeclarationMergingFile, sourceText); len(result.Diagnostics) != 0 {
		t.Fatalf("the untyped harness produced %d findings, so the nil-checker guard is not being taken", len(result.Diagnostics))
	}
	if result := rule_testing.RunTyped(t, NoUnsafeDeclarationMerging, unsafeDeclarationMergingFile, sourceText); len(result.Diagnostics) != 1 {
		t.Fatalf("the typed harness produced %d findings, want 1", len(result.Diagnostics))
	}
}

// TestNoUnsafeDeclarationMergingDescriptionExplainsTheHazard guards the sentence a reader sees.
//
// The literal is typed here rather than read from the rule, for the same reason the id assertion
// above is: comparing a message to its own constant is an equality that moves on both sides at once
// and catches nothing.
func TestNoUnsafeDeclarationMergingDescriptionExplainsTheHazard(t *testing.T) {
	t.Parallel()

	const want = "A class and an interface sharing a name are merged into one type, and the members " +
		"the interface contributes are never initialized by the class constructor. TypeScript does " +
		"not check them, so reading one compiles and returns undefined at runtime. Give the " +
		"interface a different name, or declare the members on the class."
	const sourceText = "interface Foo {}\nclass Foo {}\nexport { Foo };\n"
	result := rule_testing.RunTyped(t, NoUnsafeDeclarationMerging, unsafeDeclarationMergingFile, sourceText)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("got %d findings, want 1", len(result.Diagnostics))
	}
	if got := result.Diagnostics[0].Message.Description; got != want {
		t.Errorf("description is\n%q\nwant\n%q", got, want)
	}
}

// TestNoUnsafeDeclarationMergingHandlesAnAnonymousClass covers the one declaration in this rule's
// two kinds that can have no name at all.
//
// `export default class {}` parses as a ClassDeclaration whose Name() is nil, which is the only
// input that reaches the rule's name guard. Probed rather than assumed: an interface always carries
// a name, and a class expression is a different kind, so this single syntax is the whole of it. A
// mutant neutralizing that guard survived the entire fixture set until these cases existed.
//
// The second case is the pair that makes the first one mean something. `export default class Foo {}`
// is a named default export, it merges with the interface exactly as a plain class does, and it
// reports. So the guard has to decline the anonymous form specifically rather than decline default
// exports as a family.
func TestNoUnsafeDeclarationMergingHandlesAnAnonymousClass(t *testing.T) {
	t.Parallel()

	t.Run("an anonymous default export beside an interface", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoUnsafeDeclarationMerging,
			unsafeDeclarationMergingFile, "interface Foo {}\nexport default class {}\n"))
	})
	t.Run("an anonymous default export alone", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoUnsafeDeclarationMerging,
			unsafeDeclarationMergingFile, "export default class {}\n"))
	})
	t.Run("a named default export beside an interface", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoUnsafeDeclarationMerging,
			unsafeDeclarationMergingFile, "interface Foo {}\nexport default class Foo {}\n"), "unsafeMerging")
	})
}

// TestNoUnsafeDeclarationMergingSeesThroughAnExport is the family the imported corpus cannot reach,
// and the one that made this port more than a transcription.
//
// Upstream writes no `export` in any of its eleven cases, so nothing in the corpus exercises what an
// export does to symbol resolution. It does something substantial: typescript-go hands an exported
// declaration its own symbol carrying only itself and puts the merged list on that declaration's
// LOCAL symbol, while the non-exported partner sees both through the ordinary lookup. A rule reading
// only GetSymbolAtLocation therefore goes silent on three of the six export arrangements below,
// every one of which upstream reports.
//
// All six were pinned on the release oxlint binary at exactly one finding before being written here,
// which is what makes this a fidelity test rather than an opinion about what the rule should do.
func TestNoUnsafeDeclarationMergingSeesThroughAnExport(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// The three that were silent before the local-symbol fallback existed. In each, the merged
		// partner is the exported declaration and it is written second.
		{"an interface then an exported class", "interface Foo {}\nexport class Foo {}\n"},
		{"a class then an exported interface", "class Foo {}\nexport interface Foo {}\n"},
		{"an interface then a named default export", "interface Foo {}\nexport default class Foo {}\n"},
		// The three that already worked, kept so a later change cannot fix one direction by
		// breaking the other.
		{"an exported class then an interface", "export class Foo {}\ninterface Foo {}\n"},
		{"an exported interface then a class", "export interface Foo {}\nclass Foo {}\n"},
		{"both exported", "export interface Foo {}\nexport class Foo {}\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoUnsafeDeclarationMerging,
				unsafeDeclarationMergingFile, testCase.sourceText), "unsafeMerging")
		})
	}
}
