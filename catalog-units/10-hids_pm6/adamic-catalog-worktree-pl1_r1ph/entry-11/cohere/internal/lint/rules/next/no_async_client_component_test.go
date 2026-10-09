package next

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// asyncClientComponentFile is where the fixtures pretend to live.
//
// The name carries no meaning to this rule. Unlike most of its neighbours in this package, which
// gate on whether a path is the custom document or sits under a pages directory, this rule reads
// the file's directive prologue and never looks at the path.
const asyncClientComponentFile = "/repository/app/Component.tsx"

// The corpus is oxc's, extracted with the tool rather than transcribed.
//
// Every case below is verbatim from
// `oxc/crates/oxc_linter/src/rules/nextjs/no_async_client_component.rs`: 5 pass, 5 fail, and the
// snapshot records 5 diagnostics from those 5 inputs, so exactly one finding per failing input is
// measured rather than assumed. The strings were emitted from the extractor's own Rust parse
// through JSON encoding, so no escape was ever typed on the way here.
func TestNoAsyncClientComponentFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		why        string
	}{
		{"fail[0]", "\n                  \"use client\"\n\n                  export default async function MyComponent() {\n                    return <></>\n                  }\n                  ",
			"the direct shape, returning markup"},
		{"fail[1]", "\n                  \"use client\"\n\n                  export default async function MyFunction() {\n                    return ''\n                  }\n                  ",
			"the direct shape returning a string, which is why a component predicate would be wrong here"},
		{"fail[2]", "\n                  \"use client\"\n\n                  async function MyComponent() {\n                    return <></>\n                  }\n\n                  export default MyComponent\n                  ",
			"the indirect function shape"},
		{"fail[3]", "\n                  \"use client\"\n\n                  async function MyFunction() {\n                    return ''\n                  }\n\n                  export default MyFunction\n                  ",
			"the indirect function shape returning a string"},
		{"fail[4]", "\n                  \"use client\"\n\n                  const MyFunction = async () => {\n                    return '123'\n                  }\n\n                  export default MyFunction\n                  ",
			"the indirect async arrow shape"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.RunTyped(t, NoAsyncClientComponent, asyncClientComponentFile,
					testCase.sourceText), "noAsyncClientComponent")
		})
	}
}

// The clean cases are the whole discrimination, and each turns off exactly one condition.
//
// Read as a set they say what the rule is: a directive, a capitalized name, and an async binding
// that is the default export. Remove any one and the finding goes away.
func TestNoAsyncClientComponentStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		why        string
	}{
		{"pass[0]", "\n                export default async function MyComponent() {\n                  return <></>\n                }\n                ",
			"no directive at all, so the async default export is a server component and fine"},
		{"pass[1]", "\n                \"use client\"\n\n                export default async function myFunction() {\n                  return ''\n                }\n                ",
			"lowercase name, which is the entire component signal this rule has"},
		{"pass[2]", "\n                async function MyComponent() {\n                  return <></>\n                }\n\n                export default MyComponent\n                ",
			"indirect export with no directive"},
		{"pass[3]", "\n                \"use client\"\n\n                async function myFunction() {\n                  return ''\n                }\n\n                export default myFunction\n                ",
			"lowercase name reached indirectly"},
		{"pass[4]", "\n                \"use client\"\n\n                const myFunction = () => {\n                  return ''\n                }\n\n                export default myFunction\n                ",
			"not async, and lowercase, so it fails both tests"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoAsyncClientComponent,
				asyncClientComponentFile, testCase.sourceText))
		})
	}
}

// Cases upstream does not cover, each measured against the release oxlint binary rather than
// reasoned about. The corpus writes one quoting style, one directive, and no parenthesized forms,
// so every discrimination below is invisible to it.

// The prologue's position is load-bearing and the shim predicate does not know it.
//
// `ast.IsPrologueDirective` is purely local: it answers true for the string statement in both files
// below, because it asks only whether a statement is an expression statement wrapping a string
// literal. Upstream is silent on both, since a directive prologue ends at the first statement that
// is not a directive. Pinned on the release binary. A port reaching for that predicate alone,
// which is what the recorded research recommended building on, reports both.
func TestNoAsyncClientComponentRequiresTheProloguePosition(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"after an import", "import * as React from \"react\"\n\"use client\"\nexport default async function MyComponent() { return null }\n"},
		{"after a statement", "const z = 1\n\"use client\"\nexport default async function MyComponent() { return null }\n"},
		{"inside a function body", "function f() { \"use client\" }\nexport default async function MyComponent() { return null }\n"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoAsyncClientComponent,
				asyncClientComponentFile, testCase.sourceText))
		})
	}
}

// Quoting is transparent and a second directive above does not disarm the rule.
//
// The comparison is against the directive's parsed value, so single and double quotes both arm it.
// This is not cosmetic: every one of the 874 client files in the `ahra` tree writes single quotes
// with a trailing comment, so a port comparing raw source text would be inert on all of them while
// the imported corpus, which writes double quotes, stayed green.
func TestNoAsyncClientComponentReadsTheDirectiveValue(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"single quotes", "'use client'\nexport default async function MyComponent() { return null }\n"},
		{"single quotes with a trailing comment", "'use client'; // uses client-only features\nexport default async function MyComponent() { return null }\n"},
		{"a use strict directive above it", "\"use strict\"\n'use client'\nexport default async function MyComponent() { return null }\n"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoAsyncClientComponent,
				asyncClientComponentFile, testCase.sourceText), "noAsyncClientComponent")
		})
	}
}

// A backtick-quoted string is not a directive per the specification, and the parser agrees.
func TestNoAsyncClientComponentDeclinesATemplateLiteral(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoAsyncClientComponent,
		asyncClientComponentFile, "`use client`\nexport default async function MyComponent() { return null }\n"))
}

// Parenthesized forms are silent upstream, and reproducing that is a decision.
//
// oxc destructures an arrow initializer and an exported identifier directly with no parenthesis
// skipping, so both of these are silent on the release binary. The corpus writes no parenthesized
// form at all, so guessing either way would have cost nothing at fixture time and shipped a
// divergence. Measured rather than assumed, in both directions.
func TestNoAsyncClientComponentDeclinesParenthesizedForms(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a parenthesized async arrow initializer", "\"use client\"\nconst MyThing = (async () => { return null })\nexport default MyThing\n"},
		{"a parenthesized exported identifier", "\"use client\"\nasync function MyComponent() { return null }\nexport default (MyComponent)\n"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoAsyncClientComponent,
				asyncClientComponentFile, testCase.sourceText))
		})
	}
}

// Capitalization is Unicode-aware, which is where oxc and the `@next` original disagree.
//
// oxc tests `char::is_uppercase` and the original tests `/[A-Z]/`. Pinned on the release binary:
// the Cyrillic capital reports and its lowercase does not. `react.IsLikelyComponentName` decodes
// the first rune and asks `unicode.IsUpper`, which matches oxc exactly here. The helper's doc
// records an open ASCII question against oxc's *component* path, which uses `is_ascii_uppercase`;
// this rule is not on that path, so the helper is simply correct for it. That is the dispatch's
// warning about the helper checked and found not to apply to this rule.
func TestNoAsyncClientComponentAcceptsNonAsciiUppercase(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoAsyncClientComponent,
		asyncClientComponentFile, "\"use client\"\nexport default async function \u0424oo() { return null }\n"), "noAsyncClientComponent")
}

func TestNoAsyncClientComponentDeclinesNonAsciiLowercase(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoAsyncClientComponent,
		asyncClientComponentFile, "\"use client\"\nexport default async function \u0444oo() { return null }\n"))
}

// Shapes the corpus never writes, each silent upstream for a different reason.
func TestNoAsyncClientComponentDeclinesNearMisses(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		why        string
	}{
		{"an async function expression rather than an arrow", "\"use client\"\nconst MyThing = async function () { return null }\nexport default MyThing\n",
			"upstream binds an ArrowFunctionExpression specifically, so this is a kind check rather than an oversight"},
		{"an anonymous default export", "\"use client\"\nexport default async function () { return null }\n",
			"there is no name to capitalize-test, and upstream reads the check through the function id"},
		{"a named export with no default", "\"use client\"\nexport async function MyComponent() { return null }\n",
			"every reportable shape upstream requires the default export"},
		{"an imported binding re-exported", "\"use client\"\nimport MyComponent from \"./other\"\nexport default MyComponent\n",
			"the name resolves to an import rather than a function or a variable declaration"},
		{"a leading underscore", "\"use client\"\nexport default async function _MyComponent() { return null }\n",
			"the first rune is not uppercase, so it fails the only component signal the rule has"},
		{"a destructured binding", "\"use client\"\ndeclare const source: { MyThing: () => Promise<null> }\nconst { MyThing } = source\nexport default MyThing\n",
			"upstream requires a binding identifier, and a pattern binds names without naming the initializer"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoAsyncClientComponent,
				asyncClientComponentFile, testCase.sourceText))
		})
	}
}

// Shapes the corpus never writes that DO report, so the silence above is a discrimination rather
// than a narrow rule.
func TestNoAsyncClientComponentReportsShapesTheCorpusOmits(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		why        string
	}{
		{"a let binding rather than const", "\"use client\"\nlet MyThing = async () => { return null }\nexport default MyThing\n",
			"upstream reads the declarator, not the declaration kind"},
		{"a var binding", "\"use client\"\nvar MyThing = async () => { return null }\nexport default MyThing\n",
			"same, and var is the one whose hoisting might have suggested otherwise"},
		{"a declaration written after the export", "\"use client\"\nexport default MyThing\nconst MyThing = async () => { return null }\n",
			"upstream resolves the symbol rather than scanning statements before the export, so order does not matter"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoAsyncClientComponent,
				asyncClientComponentFile, testCase.sourceText), "noAsyncClientComponent")
		})
	}
}

// Shadowing, which is the input separating real symbol resolution from a name-matching scan.
//
// The `@next` original finds the declaration with `node.body.find(...)` over top-level statements
// by name. That scan finds the synchronous top-level `MyComponent` here and so happens to agree,
// but the two are different rules and no imported fixture votes on it. Silent on the release
// binary, and silent here because the checker resolves the export to the top-level declaration.
func TestNoAsyncClientComponentResolvesThroughShadowing(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoAsyncClientComponent,
		asyncClientComponentFile, "\"use client\"\nfunction MyComponent() { return null }\n{ async function MyComponent() { return null } }\nexport default MyComponent\n"))
}

// Declaration merging, where the standing advice in this tree points the wrong way and where the
// rule turns out to be order-sensitive.
//
// The advice is to loop over `symbol.Declarations` rather than index zero, because merging is said
// to put declarations in source order. Measured in typescript-go, it does not: an interface written
// above the function still leaves the function at index 0, because a value declaration sorts first
// regardless of where it was written. The input that advice exists to catch does not arise for this
// shape, which is why a mutant replacing the loop with an index-zero read survived a fixture
// written specifically to kill it.
//
// All four verdicts below are pinned on the release oxlint binary, and they split on which
// declaration appears FIRST IN SOURCE. oxc reaches one node through `symbol_declaration(symbol_id)`
// and its binder records the first declaration it walks past, so a merged name whose first
// declaration is a type or a namespace never reaches the async check at all. A loop reports all
// four; an index-zero read reports three and is right twice by accident.
func TestNoAsyncClientComponentReportsWhenTheAsyncDeclarationIsFirst(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"an interface below the function", "\"use client\"\nasync function MyComponent() { return null }\ninterface MyComponent { field: string }\nexport default MyComponent\n"},
		{"a namespace below the arrow", "\"use client\"\nlet MyThing = async () => { return null }\nnamespace MyThing { export const x = 1 }\nexport default MyThing\n"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoAsyncClientComponent,
				asyncClientComponentFile, testCase.sourceText), "noAsyncClientComponent")
		})
	}
}

func TestNoAsyncClientComponentDeclinesWhenTheMergedDeclarationIsFirst(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"an interface above the function", "\"use client\"\ninterface MyComponent { field: string }\nasync function MyComponent() { return null }\nexport default MyComponent\n"},
		{"a namespace above the function", "\"use client\"\nnamespace MyComponent { export const x = 1 }\nasync function MyComponent() { return null }\nexport default MyComponent\n"},
		{"a namespace above the arrow", "\"use client\"\nnamespace MyThing { export const x = 1 }\nlet MyThing = async () => { return null }\nexport default MyThing\n"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoAsyncClientComponent,
				asyncClientComponentFile, testCase.sourceText))
		})
	}
}

// A synchronous arrow with a CAPITALIZED name, which the corpus never writes.
//
// Upstream's only sync-arrow clean case is `pass[4]`, and its name is lowercase, so it exits on the
// capitalization test and never votes on whether the initializer is async. A mutant forcing the
// async check always-true survived every imported fixture on exactly that gap. This is the input
// that separates the two conditions.
func TestNoAsyncClientComponentDeclinesASynchronousCapitalizedArrow(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoAsyncClientComponent,
		asyncClientComponentFile, "\"use client\"\nconst MyThing = () => { return null }\nexport default MyThing\n"))
}

// A destructured name reaches the variable arm as a binding element rather than a declaration.
//
// Kept as a fixture even though the rule has no kind guard for it, because the reason it is silent
// lives in the parse shape rather than in the rule, and a later reader adding that guard back
// should see this case already passing without it.
func TestNoAsyncClientComponentDeclinesADestructuredArrayBinding(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoAsyncClientComponent,
		asyncClientComponentFile, "\"use client\"\ndeclare const source: Array<() => Promise<null>>\nconst [MyThing] = source\nexport default MyThing\n"))
}

// A declared binding with no initializer at all, which reaches the arm and must not dereference.
func TestNoAsyncClientComponentDeclinesABindingWithNoInitializer(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoAsyncClientComponent,
		asyncClientComponentFile, "\"use client\"\nlet MyThing: unknown\nexport default MyThing\n"))
}

// A lowercase name on an ASYNC arrow, which the corpus does not write.
//
// Upstream's `pass[4]` is the nearest case and it turns off two conditions at once, being both
// lowercase and synchronous, so it exits on the async test and never votes on capitalization for
// the arrow path. A mutant neutralizing the capitalization check in that arm survived every other
// fixture in this file on exactly that gap. Silent on the release binary.
func TestNoAsyncClientComponentDeclinesALowercaseAsyncArrow(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoAsyncClientComponent,
		asyncClientComponentFile, "\"use client\"\nconst myThing = async () => { return null }\nexport default myThing\n"))
}

// The typed harness is required, and this fails loudly if the declaration is ever reverted.
//
// Under the plain harness the checker is nil, the guard at the top of the listener returns, and
// every silence fixture above would pass vacuously while every firing one failed. That is the more
// dangerous half: a vacuous green does not announce itself.
func TestNoAsyncClientComponentNeedsTheTypedHarness(t *testing.T) {
	t.Parallel()

	if !NoAsyncClientComponent.NeedsTypeChecker {
		t.Fatal("the rule resolves an exported identifier through the checker and must declare it")
	}
	rule_testing.ExpectClean(t, rule_testing.Run(t, NoAsyncClientComponent, asyncClientComponentFile,
		"\"use client\"\nexport default async function MyComponent() { return null }\n"))
}

// The span and the rendered message, which the identifier assertions above cannot see.
//
// Upstream labels the identifier and the `@next` original labels the whole declaration statement,
// with the same message on both, so every message-id fixture in this file stays green over the
// wrong choice. The snapshot pins oxc's narrower span at column 49 of the direct shape and column
// 25 of the arrow shape, which is the identifier in each.
//
// The expected text is a literal typed here rather than a reference to the rule's own message
// constant. Compared against the constant, both sides move together under mutation and the
// assertion can never fail.
func TestNoAsyncClientComponentPointsAtTheIdentifier(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		want       string
	}{
		{"the direct shape", "\"use client\"\nexport default async function MyComponent() { return null }\n", "MyComponent"},
		{"the indirect function shape", "\"use client\"\nasync function MyComponent() { return null }\nexport default MyComponent\n", "MyComponent"},
		{"the indirect arrow shape", "\"use client\"\nconst MyThing = async () => { return null }\nexport default MyThing\n", "MyThing"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoAsyncClientComponent, asyncClientComponentFile,
				testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("want 1 finding, got %d", len(result.Diagnostics))
			}
			reported := testCase.sourceText[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if reported != testCase.want {
				t.Errorf("finding points at %q, want %q", reported, testCase.want)
			}
			wantDescription := "A client component declared as an async function is rendered on " +
				"the client, where React cannot await it the way the server can. The promise it " +
				"returns is not an element, so rendering fails or hydration diverges from what the " +
				"server produced. Move the awaiting into an effect or an event handler and keep " +
				"the component itself synchronous."
			if result.Diagnostics[0].Message.Description != wantDescription {
				t.Errorf("description is %q", result.Diagnostics[0].Message.Description)
			}
			if result.Diagnostics[0].Message.Id != "noAsyncClientComponent" {
				t.Errorf("id is %q", result.Diagnostics[0].Message.Id)
			}
		})
	}
}
