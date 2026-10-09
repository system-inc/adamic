package react

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// incompatibleLibraryFile is where the fixtures pretend to live.
//
// A `.tsx` extension because every fixture here contains JSX. Under `.ts` the parser reads `<div>`
// as a type assertion rather than an element, and the component gate these cases depend on would
// then be answered by a tree with no JSX in it.
const incompatibleLibraryFile = "/repository/source/IncompatibleLibrary.tsx"

// withSeededIncompatibleLibraryTable swaps the module table for the duration of one test.
//
// Unlike `configuration.go`'s equivalent seam, this one is a convenience rather than a necessity, and the
// difference is the whole finding of this port. That rule cannot report on its shipped table at all,
// so seeding is the only way to prove it is not inert. This rule's shipped table fires on real
// library names with no configuration, which `TestIncompatibleLibraryFiresOnTheRealDefaultTable`
// below asserts directly. Seeding exists here only so upstream's own three fixtures, which import a
// module registered by a test harness rather than a real package, can be run exactly as written.
func withSeededIncompatibleLibraryTable(t *testing.T, seeded map[string]map[string]incompatibleLibraryEntry) {
	t.Helper()
	previous := incompatibleLibraryTable
	incompatibleLibraryTable = func(moduleName string) (map[string]incompatibleLibraryEntry, bool) {
		entries, present := seeded[moduleName]
		return entries, present
	}
	t.Cleanup(func() { incompatibleLibraryTable = previous })
}

// upstreamIncompatibleLibraryProvider is the `ReactCompilerKnownIncompatibleTest` entry from
// upstream's snap type provider, as oxc records it at `tests/snapshot.rs:390`.
//
// Three of its four members are reproduced. `knownIncompatibleAliasing` is omitted: it exists to
// exercise an aliasing signature, a shape this port does not model, and its fixture lives only in
// oxc's compiler corpus rather than in the conformance fixtures this package is measured against.
func upstreamIncompatibleLibraryProvider() map[string]map[string]incompatibleLibraryEntry {
	return map[string]map[string]incompatibleLibraryEntry{
		"ReactCompilerKnownIncompatibleTest": {
			"useKnownIncompatible": {message: "useKnownIncompatible is known to be incompatible"},
			"knownIncompatible":    {message: "useKnownIncompatible is known to be incompatible"},
			"useKnownIncompatibleIndirect": {properties: map[string]string{
				"incompatible": "useKnownIncompatibleIndirect returns an incompatible() function that is known incompatible",
			}},
		},
	}
}

// TestIncompatibleLibraryFiresOnUpstreamFixtures runs React's three error fixtures verbatim.
//
// Each source is byte-for-byte the `.js` file from `internal/react_conformance/testdata/fixtures`,
// read and emitted as a Go literal by a script rather than typed, so no transcription step existed
// that could cook an escape. Verified by sha256 against the files on disk:
//
//	error.invalid-known-incompatible-function                 05a61c12eade9e86...
//	error.invalid-known-incompatible-hook                     c1c78a35d5b489d0...
//	error.invalid-known-incompatible-hook-return-property     6f3a1196f08d0db7...
//
// These are the complete set of `IncompatibleLibrary` error fixtures the conformance corpus ships.
// The dispatch for this port said four; three is what is present, and the fourth
// (`error.invalid-known-incompatible-aliasing-function`) exists only under oxc's compiler fixtures.
// Not parallel: it swaps the package variable incompatibleLibraryTable through
// withSeededIncompatibleLibraryTable, which the incompatible-library rule reads on every run, so a
// parallel test running that rule would race the swap and read the seeded table instead of the shipped
// one.
func TestIncompatibleLibraryFiresOnUpstreamFixtures(t *testing.T) {
	withSeededIncompatibleLibraryTable(t, upstreamIncompatibleLibraryProvider())

	cases := []struct {
		name   string
		source string
	}{
		// error.invalid-known-incompatible-function
		// A plain function import, called directly. The entry's own message fires.
		{"directFunctionCall", "import {knownIncompatible} from 'ReactCompilerKnownIncompatibleTest';\n\nfunction Component() {\n  const data = knownIncompatible();\n  return <div>Error</div>;\n}\n"},
		// error.invalid-known-incompatible-hook
		// The same shape under a hook-named import.
		{"directHookCall", "import {useKnownIncompatible} from 'ReactCompilerKnownIncompatibleTest';\n\nfunction Component() {\n  const data = useKnownIncompatible();\n  return <div>Error</div>;\n}\n"},
		// error.invalid-known-incompatible-hook-return-property
		// The one level of indirection: destructure a property off a hook result, then call it.
		// The hook itself is compatible; the property it hands back is not.
		{"hookReturnProperty", "import {useKnownIncompatibleIndirect} from 'ReactCompilerKnownIncompatibleTest';\n\nfunction Component() {\n  const {incompatible} = useKnownIncompatibleIndirect();\n  return <div>{incompatible()}</div>;\n}\n"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, IncompatibleLibrary, incompatibleLibraryFile, testCase.source)
			rule_testing.ExpectFindings(t, result, "incompatibleLibrary")
		})
	}
}

// TestIncompatibleLibraryFiresOnTheRealDefaultTable is the proof this rule is not configuration-bound.
//
// It runs WITHOUT the seeded table, so the rule consults the shipped
// `defaultIncompatibleLibraryTable` exactly as it does on Kirk's tree. All three entries report on
// their real library names with no configuration whatsoever, which is the finding that separates
// this rule from `react/config`: that one's default table is self-consistent and can never fire,
// while this one's carries a live message at every entry.
//
// Each of these three was measured against React 7.1.1 through the ESLint Linter interface before
// being written here, rather than derived from reading the table.
func TestIncompatibleLibraryFiresOnTheRealDefaultTable(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		source string
	}{
		{"tanstackTable", "import {useReactTable} from '@tanstack/react-table';\n\nfunction Component() {\n  const table = useReactTable();\n  return <div>{table}</div>;\n}\n"},
		{"tanstackVirtual", "import {useVirtualizer} from '@tanstack/react-virtual';\n\nfunction Component() {\n  const rows = useVirtualizer();\n  return <div>{rows}</div>;\n}\n"},
		// The nested entry. `useForm` itself is fine; the `watch` it returns is not.
		{"reactHookFormWatch", "import {useForm} from 'react-hook-form';\n\nfunction Component() {\n  const {watch} = useForm();\n  return <div>{watch()}</div>;\n}\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, IncompatibleLibrary, incompatibleLibraryFile, testCase.source)
			rule_testing.ExpectFindings(t, result, "incompatibleLibrary")
		})
	}
}

// TestIncompatibleLibraryFiresOnTheOxcCorpus runs oxc's single failing lint case verbatim.
//
// oxc does ship a lint rule for this, which corrects a standing belief that this family has none on
// that side. Its corpus is one pass and one fail, and the fail imports the *genuine*
// `@tanstack/react-table` rather than the synthetic test module, which is only a sensible corpus if
// the shipped table fires. That is independent confirmation of the finding above, arrived at from
// the other implementation.
//
// The leading newline is upstream's own, and `rule_testing.Run` does not trim, so this is byte-identical
// to the Rust literal.
func TestIncompatibleLibraryFiresOnTheOxcCorpus(t *testing.T) {
	t.Parallel()
	source := "\nimport {useReactTable} from '@tanstack/react-table';\nfunction Component({columns, data}) {\n  const table = useReactTable({columns, data});\n  return <div>{table.getRowModel().rows.length}</div>;\n}\n"
	result := rule_testing.Run(t, IncompatibleLibrary, incompatibleLibraryFile, source)
	rule_testing.ExpectFindings(t, result, "incompatibleLibrary")
}

// TestIncompatibleLibraryReadsThroughAnAlias pins that the IMPORTED name decides the lookup while
// the LOCAL name decides the match at the call site.
//
// No upstream fixture writes an alias, so this is written from measuring React 7.1.1 directly:
// `import {useReactTable as renamed}` reports, with the span on `renamed`. Keying the table by the
// local name would go silent here, and keying the call site by the imported name would go silent
// too. Only splitting the two answers both, and nothing in the imported corpus can see it.
func TestIncompatibleLibraryReadsThroughAnAlias(t *testing.T) {
	t.Parallel()
	source := "import {useReactTable as renamed} from '@tanstack/react-table';\n\nfunction Component() {\n  const table = renamed();\n  return <div>{table}</div>;\n}\n"
	result := rule_testing.Run(t, IncompatibleLibrary, incompatibleLibraryFile, source)
	rule_testing.ExpectFindings(t, result, "incompatibleLibrary")
	found := result.Diagnostics[0]
	if reported := source[found.Range.Pos():found.Range.End()]; reported != "renamed" {
		t.Errorf("reported span = %q, want %q", reported, "renamed")
	}
}

// TestIncompatibleLibraryFiresOnARenamedDestructure covers the nested entry through a rename.
//
// Measured at React 7.1.1: `const {watch: w} = useForm(); w()` reports with the span on `w`. The
// PROPERTY name is what looks the entry up and the LOCAL name is what the call spells, which is the
// same split as the alias case above but at a different node. Written because the single upstream
// nested fixture does not rename, so nothing imported exercises this.
func TestIncompatibleLibraryFiresOnARenamedDestructure(t *testing.T) {
	t.Parallel()
	source := "import {useForm} from 'react-hook-form';\n\nfunction Component() {\n  const {watch: w} = useForm();\n  return <div>{w()}</div>;\n}\n"
	result := rule_testing.Run(t, IncompatibleLibrary, incompatibleLibraryFile, source)
	rule_testing.ExpectFindings(t, result, "incompatibleLibrary")
	found := result.Diagnostics[0]
	if reported := source[found.Range.Pos():found.Range.End()]; reported != "w" {
		t.Errorf("reported span = %q, want %q", reported, "w")
	}
}

// TestIncompatibleLibraryFiresOnAMemberAccess covers `form.watch()` without a destructure.
//
// A separate code path from the destructure, and silent if only the destructure is matched. The
// span is the one case in this rule a reader would guess wrong, and this fixture was written
// asserting the property and CORRECTED by measurement rather than the other way round: React 7.1.1
// underlines the OBJECT, `form`, not `form.watch` and not `watch`. The wrong expectation failed
// here first, which is the fixture doing its job.
//
// The chained form is the same judgment with no intervening binding, and upstream's span there is
// the whole inner call `useForm()`.
func TestIncompatibleLibraryFiresOnAMemberAccess(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		source string
		span   string
	}{
		{"boundResult", "import {useForm} from 'react-hook-form';\n\nfunction Component() {\n  const form = useForm();\n  return <div>{form.watch()}</div>;\n}\n", "form"},
		{"chainedOffTheCall", "import {useForm} from 'react-hook-form';\n\nfunction Component() {\n  return <div>{useForm().watch()}</div>;\n}\n", "useForm()"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, IncompatibleLibrary, incompatibleLibraryFile, testCase.source)
			rule_testing.ExpectFindings(t, result, "incompatibleLibrary")
			found := result.Diagnostics[0]
			if reported := testCase.source[found.Range.Pos():found.Range.End()]; reported != testCase.span {
				t.Errorf("reported span = %q, want %q", reported, testCase.span)
			}
		})
	}
}

// TestIncompatibleLibraryDeclinesAReassignedResult records a second measured gap.
//
// React 7.1.1 reports on `const a = useForm(); const b = a; b.watch()`, because it tracks the value
// rather than the name. This port matches syntactically and is silent, which is the narrowness
// stated at the rule. Asserted with the reason at the line rather than omitted.
func TestIncompatibleLibraryDeclinesAReassignedResult(t *testing.T) {
	t.Parallel()
	source := "import {useForm} from 'react-hook-form';\n\nfunction Component() {\n  const a = useForm();\n  const b = a;\n  return <div>{b.watch()}</div>;\n}\n"
	result := rule_testing.Run(t, IncompatibleLibrary, incompatibleLibraryFile, source)
	rule_testing.ExpectClean(t, result)
}

// TestIncompatibleLibraryFiresInsideACustomHook pins the second half of the component gate.
//
// Measured: a hook-named function is a compilation root just as a component is, so the same call
// that is silent inside `helper()` reports inside `useThing()`. Without this case the gate could be
// narrowed to components alone and every imported fixture would stay green, since all three of them
// use a function named `Component`.
func TestIncompatibleLibraryFiresInsideACustomHook(t *testing.T) {
	t.Parallel()
	source := "import {useReactTable} from '@tanstack/react-table';\n\nfunction useThing() {\n  return useReactTable();\n}\n"
	result := rule_testing.Run(t, IncompatibleLibrary, incompatibleLibraryFile, source)
	rule_testing.ExpectFindings(t, result, "incompatibleLibrary")
}

// TestIncompatibleLibraryReportsOncePerFunction reproduces upstream's throw.
//
// Upstream raises the diagnostic by throwing, which aborts compilation of the enclosing function, so
// a second offending call never runs. Measured at React 7.1.1: both a repeated call and two
// *different* incompatible APIs in one component report exactly once, on the first in source order.
//
// The second case is the load-bearing one. A rule that deduplicated by message text rather than by
// enclosing function would pass the repeated-call case and report twice here.
func TestIncompatibleLibraryReportsOncePerFunction(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		source string
		span   string
	}{
		{"sameCallTwice", "import {useReactTable} from '@tanstack/react-table';\n\nfunction Component() {\n  const a = useReactTable();\n  const b = useReactTable();\n  return <div>{a}{b}</div>;\n}\n", "useReactTable"},
		{"twoDifferentLibraries", "import {useReactTable} from '@tanstack/react-table';\nimport {useVirtualizer} from '@tanstack/react-virtual';\n\nfunction Component() {\n  const a = useReactTable();\n  const b = useVirtualizer();\n  return <div>{a}{b}</div>;\n}\n", "useReactTable"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, IncompatibleLibrary, incompatibleLibraryFile, testCase.source)
			rule_testing.ExpectFindings(t, result, "incompatibleLibrary")
			found := result.Diagnostics[0]
			if reported := testCase.source[found.Range.Pos():found.Range.End()]; reported != testCase.span {
				t.Errorf("reported span = %q, want %q", reported, testCase.span)
			}
		})
	}
}

// TestIncompatibleLibraryReportsPerComponent pins that the abort is scoped to ONE function.
//
// Two sibling components each calling an incompatible API report once each, for two findings in the
// file. Written because the one-per-function bookkeeping could be implemented as a single file-wide
// flag, which would pass every case above while silencing the second component here.
func TestIncompatibleLibraryReportsPerComponent(t *testing.T) {
	t.Parallel()
	source := "import {useReactTable} from '@tanstack/react-table';\n\nfunction First() {\n  const a = useReactTable();\n  return <div>{a}</div>;\n}\n\nfunction Second() {\n  const b = useReactTable();\n  return <div>{b}</div>;\n}\n"
	result := rule_testing.Run(t, IncompatibleLibrary, incompatibleLibraryFile, source)
	rule_testing.ExpectFindings(t, result, "incompatibleLibrary", "incompatibleLibrary")
}

// TestIncompatibleLibraryStaysSilent covers everything the rule must decline.
func TestIncompatibleLibraryStaysSilent(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		source string
	}{
		// oxc's single passing case, verbatim, including its leading newline.
		{"oxcPassCase", "\nfunction Component(props) {\n  return <div>{props.text}</div>;\n}\n"},
		// A module the table answers nothing for. This is the case every import on Kirk's tree
		// takes, and it is why this rule finds nothing there.
		{"unconfiguredModule", "import {useAnything} from 'some-other-module';\n\nfunction Component() {\n  const x = useAnything();\n  return <div>{x}</div>;\n}\n"},
		// The right name from the wrong module. The table is keyed by module, so the name alone is
		// not enough, and a rule matching on callee name only would report here.
		{"rightNameWrongModule", "import {useReactTable} from 'some-other-module';\n\nfunction Component() {\n  const t = useReactTable();\n  return <div>{t}</div>;\n}\n"},
		// A property the configured module does not declare.
		{"propertyNotInTable", "import {somethingElse} from '@tanstack/react-table';\n\nfunction Component() {\n  const x = somethingElse();\n  return <div>{x}</div>;\n}\n"},
		// Imported and never called. The diagnostic is raised while inferring a CALL's effects, so
		// an unused import is silent. Measured.
		{"importedNotCalled", "import {useReactTable} from '@tanstack/react-table';\n\nfunction Component() {\n  return <div>hi</div>;\n}\n"},
		// Passed onward as a value rather than called. Also measured silent upstream.
		{"passedNotCalled", "import {useReactTable} from '@tanstack/react-table';\n\nfunction Component() {\n  return <div onClick={useReactTable}>hi</div>;\n}\n"},
		// The component gate, three ways. A lowercase function is not a compilation root, so the
		// identical call inside it is silent at upstream.
		{"lowercaseHelper", "import {useReactTable} from '@tanstack/react-table';\n\nfunction helper() {\n  return useReactTable();\n}\n"},
		{"topLevelCall", "import {useReactTable} from '@tanstack/react-table';\n\nconst table = useReactTable();\n"},
		// A class method is not a compilation root either. Measured silent at React 7.1.1 even
		// though the class name is capitalized and the method returns JSX.
		{"classMethod", "import {useReactTable} from '@tanstack/react-table';\n\nclass Widget {\n  render() {\n    const t = useReactTable();\n    return <div>{t}</div>;\n  }\n}\n"},
		// The nested entry's hook, called without ever touching the incompatible property. `useForm`
		// declares no message of its own, so this is silent. Measured.
		{"hookWithoutTheProperty", "import {useForm} from 'react-hook-form';\n\nfunction Component() {\n  const form = useForm();\n  return <div>{form}</div>;\n}\n"},
		// A different property off the same hook result.
		{"otherPropertyOfHookResult", "import {useForm} from 'react-hook-form';\n\nfunction Component() {\n  const {register} = useForm();\n  return <div>{register()}</div>;\n}\n"},
		// Destructured but never called.
		{"propertyDestructuredNotCalled", "import {useForm} from 'react-hook-form';\n\nfunction Component() {\n  const {watch} = useForm();\n  return <div>hi</div>;\n}\n"},
		// Read off the result but never called. Measured silent at React 7.1.1: the diagnostic is
		// raised while inferring a CALL's effects, so merely naming the property is fine.
		{"propertyReadNotCalled", "import {useForm} from 'react-hook-form';\n\nfunction Component() {\n  const form = useForm();\n  return <div>{form.watch}</div>;\n}\n"},
		// The module name is compared EXACTLY. Both of these are measured silent at React 7.1.1,
		// and both would report under a lowercasing or prefix-matching comparison. `configuration.go`'s
		// `isKnownReactModule` does lowercase, so this is the axis on which the two rules differ and
		// a reader copying that helper here would introduce a false positive.
		{"moduleNameCasing", "import {useReactTable} from '@TanStack/React-Table';\n\nfunction Component() {\n  const t = useReactTable();\n  return <div>{t}</div>;\n}\n"},
		{"moduleSubpath", "import {useReactTable} from '@tanstack/react-table/core';\n\nfunction Component() {\n  const t = useReactTable();\n  return <div>{t}</div>;\n}\n"},
		// A local alias defeats the rule at React too, measured, so this silence is upstream's
		// rather than a gap introduced here.
		{"localAliasVariable", "import {useReactTable} from '@tanstack/react-table';\n\nconst local = useReactTable;\n\nfunction Component() {\n  const t = local();\n  return <div>{t}</div>;\n}\n"},
		// A side-effect import binds nothing to call.
		{"sideEffectImport", "import '@tanstack/react-table';\n\nfunction Component() {\n  return <div>hi</div>;\n}\n"},
		// A default import from a configured module. The table declares no `default`, and measured
		// silent upstream.
		{"defaultImport", "import table from '@tanstack/react-table';\n\nfunction Component() {\n  const t = table();\n  return <div>{t}</div>;\n}\n"},
		// A module specifier that is not a string literal. Our parser recovers into an identifier
		// whose Text() reads back as a usable name, so this consults nothing and is silent.
		{"nonLiteralSpecifier", "import {useReactTable} from tanstack;\n\nfunction Component() {\n  const t = useReactTable();\n  return <div>{t}</div>;\n}\n"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, IncompatibleLibrary, incompatibleLibraryFile, testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestIncompatibleLibrarySurvivesAMalformedModuleSpecifier pins the kind guard as CRASH protection.
//
// This exists because a mutation deleting `ast.IsStringLiteralLike` SURVIVED the whole suite, and
// the reason is the fifth survivor category rather than a fixture blind spot: no `ExpectFindings`
// case can see a panic, so the sweep scores a guard that prevents one identically whether it matters
// or not.
//
// The first instinct was wrong and is worth recording. The `nonLiteralSpecifier` case above was
// written believing the guard was behavioral, on the reasoning that a recovered identifier's
// `Text()` reads back as a usable module name. It does, and both versions still reach silence there,
// because `tanstack` is not a name the table configures and no identifier can spell
// `@tanstack/react-table`. That is "both branches reach the same verdict by different routes".
//
// Probed over the specifier shapes our parser actually produces, which is what settled it:
//
//	import {u} from tanstack;                 KindIdentifier                 Text() ok
//	import {u} from 123;                      KindNumericLiteral             Text() ok
//	import {u} from `@tanstack/${x}`;         KindTemplateExpression         Text() PANICS
//	import {u} from a.b;                      KindPropertyAccessExpression   Text() PANICS
//	import {u} from ('@tanstack/react-table') KindParenthesizedExpression    Text() PANICS
//
// So three of six recovered shapes take the linter down without the guard, and the last one is the
// sharpest: the parenthesized form carries a specifier that WOULD match the table. Each is asserted
// clean here, and each would panic rather than fail if the guard were removed.
func TestIncompatibleLibrarySurvivesAMalformedModuleSpecifier(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		source string
	}{
		{"templateWithSubstitution", "import {useReactTable} from `@tanstack/${x}`;\n\nfunction Component() {\n  const t = useReactTable();\n  return <div>{t}</div>;\n}\n"},
		{"memberExpressionSpecifier", "import {useReactTable} from a.b;\n\nfunction Component() {\n  const t = useReactTable();\n  return <div>{t}</div>;\n}\n"},
		{"parenthesizedSpecifier", "import {useReactTable} from ('@tanstack/react-table');\n\nfunction Component() {\n  const t = useReactTable();\n  return <div>{t}</div>;\n}\n"},
		{"numericSpecifier", "import {useReactTable} from 123;\n\nfunction Component() {\n  const t = useReactTable();\n  return <div>{t}</div>;\n}\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, IncompatibleLibrary, incompatibleLibraryFile, testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestIncompatibleLibraryAcceptsATemplateSpecifier pins the other half of `IsStringLiteralLike`.
//
// A no-substitution template is a real module specifier and the predicate accepts it, so this
// REPORTS. Written because a guard narrowed to `ast.IsStringLiteral` would pass every case above
// while going silent here, and nothing in the imported corpus writes a template.
func TestIncompatibleLibraryAcceptsATemplateSpecifier(t *testing.T) {
	t.Parallel()
	source := "import {useReactTable} from `@tanstack/react-table`;\n\nfunction Component() {\n  const t = useReactTable();\n  return <div>{t}</div>;\n}\n"
	result := rule_testing.Run(t, IncompatibleLibrary, incompatibleLibraryFile, source)
	rule_testing.ExpectFindings(t, result, "incompatibleLibrary")
}

// TestIncompatibleLibraryDeclinesTheNamespaceForm records a measured gap rather than hiding it.
//
// React 7.1.1 DOES report on `table.useReactTable()` after `import * as table`, with the span on the
// object rather than the property. This port is silent there, because reproducing it means resolving
// a namespace member against the module's table, which is the member-access path the rule declines
// to build. Asserted as clean with the reason at the line rather than deleted: a case removed from a
// corpus is a divergence no later reader can find.
func TestIncompatibleLibraryDeclinesTheNamespaceForm(t *testing.T) {
	t.Parallel()
	source := "import * as table from '@tanstack/react-table';\n\nfunction Component() {\n  const t = table.useReactTable();\n  return <div>{t}</div>;\n}\n"
	result := rule_testing.Run(t, IncompatibleLibrary, incompatibleLibraryFile, source)
	rule_testing.ExpectClean(t, result)
}

// TestIncompatibleLibraryReportsOnTheCallee pins WHERE the finding points.
//
// `ExpectFindings` asserts ids and count and nothing else, so a rule pointing at the whole call
// expression, at the variable being assigned, or at the import would pass every case above while
// being wrong. Both upstreams agree on the callee: React's span is `useReactTable` and oxc's
// snapshot underlines the same thirteen characters.
//
// The expectation is a literal typed here rather than anything the rule computes, so it cannot move
// with the rule.
func TestIncompatibleLibraryReportsOnTheCallee(t *testing.T) {
	t.Parallel()
	source := "import {useReactTable} from '@tanstack/react-table';\n\nfunction Component() {\n  const table = useReactTable({columns: 1});\n  return <div>{table}</div>;\n}\n"
	result := rule_testing.Run(t, IncompatibleLibrary, incompatibleLibraryFile, source)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("want 1 diagnostic, got %d", len(result.Diagnostics))
	}
	found := result.Diagnostics[0]
	if reported := source[found.Range.Pos():found.Range.End()]; reported != "useReactTable" {
		t.Errorf("reported span = %q, want %q", reported, "useReactTable")
	}
}

// TestIncompatibleLibraryMessageCarriesTheEntryText asserts the per-entry half of the message.
//
// This rule is the interpolating kind the port brief warns about: the description is the shared
// explanation plus the table entry's own text, so a rule that dropped the entry text, or attached
// the wrong entry's, would satisfy every id assertion above. Both halves are asserted, and against
// literals typed here rather than against `messageIncompatibleLibrary`, so a mutation rewriting the
// constant moves the rule and the test in opposite directions instead of together.
//
// The two cases carry DIFFERENT entry text, which is what makes this able to see a rule that always
// appends the same message.
func TestIncompatibleLibraryMessageCarriesTheEntryText(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		source string
		want   string
	}{
		{"tanstackTable", "import {useReactTable} from '@tanstack/react-table';\n\nfunction Component() {\n  const t = useReactTable();\n  return <div>{t}</div>;\n}\n", "TanStack Table's `useReactTable()` API returns functions that cannot be memoized safely"},
		{"reactHookForm", "import {useForm} from 'react-hook-form';\n\nfunction Component() {\n  const {watch} = useForm();\n  return <div>{watch()}</div>;\n}\n", "React Hook Form's `useForm()` API returns a `watch()` function which cannot be memoized safely."},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, IncompatibleLibrary, incompatibleLibraryFile, testCase.source)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("want 1 diagnostic, got %d", len(result.Diagnostics))
			}
			found := result.Diagnostics[0]
			if found.Message.Id != "incompatibleLibrary" {
				t.Errorf("message id = %q, want %q", found.Message.Id, "incompatibleLibrary")
			}
			if !strings.HasPrefix(found.Message.Description, "This library's API returns functions that cannot be memoized without showing stale data.") {
				t.Errorf("message = %q, want it to open with the shared explanation", found.Message.Description)
			}
			if !strings.HasSuffix(found.Message.Description, testCase.want) {
				t.Errorf("message = %q, want it to end with %q", found.Message.Description, testCase.want)
			}
		})
	}
}

// TestDefaultIncompatibleLibraryTableMatchesUpstream pins the three entries and their messages.
//
// The messages are printed verbatim by both upstreams, so a reworded copy is a visible divergence
// rather than a cosmetic one. Asserted against literals rather than against the table itself, which
// is the same reason the message test above does not compare to the rule's own constant.
//
// The last two rows are the exactness of the module comparison, which is the axis this table differs
// from `configuration.go`'s `isKnownReactModule` on.
func TestDefaultIncompatibleLibraryTableMatchesUpstream(t *testing.T) {
	t.Parallel()
	cases := []struct {
		module   string
		member   string
		property string
		want     string
	}{
		{"@tanstack/react-table", "useReactTable", "", "TanStack Table's `useReactTable()` API returns functions that cannot be memoized safely"},
		{"@tanstack/react-virtual", "useVirtualizer", "", "TanStack Virtual's `useVirtualizer()` API returns functions that cannot be memoized safely"},
		{"react-hook-form", "useForm", "watch", "React Hook Form's `useForm()` API returns a `watch()` function which cannot be memoized safely."},
	}
	for _, testCase := range cases {
		entries, configured := defaultIncompatibleLibraryTable(testCase.module)
		if !configured {
			t.Fatalf("defaultIncompatibleLibraryTable(%q) reported no configuration", testCase.module)
		}
		entry, present := entries[testCase.member]
		if !present {
			t.Fatalf("module %q declares no entry for %q", testCase.module, testCase.member)
		}
		got := entry.message
		if testCase.property != "" {
			got = entry.properties[testCase.property]
		}
		if got != testCase.want {
			t.Errorf("%s.%s message = %q, want %q", testCase.module, testCase.member, got, testCase.want)
		}
	}

	// Exact comparison, both directions. A lowercasing version would configure the first and a
	// prefix-matching one would configure the second, and both were measured silent at React 7.1.1.
	for _, moduleName := range []string{"@TanStack/React-Table", "@tanstack/react-table/core", "react-hook-form-extra", ""} {
		if _, configured := defaultIncompatibleLibraryTable(moduleName); configured {
			t.Errorf("defaultIncompatibleLibraryTable(%q) reported a configuration, want none", moduleName)
		}
	}
}
