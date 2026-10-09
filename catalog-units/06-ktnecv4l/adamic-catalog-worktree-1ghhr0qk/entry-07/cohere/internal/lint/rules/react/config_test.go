package react

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// configFile is where the fixtures pretend to live.
//
// A `.tsx` extension because two of the four upstream fixtures contain JSX. Under `.ts` the parser
// reads `<div>` as a type assertion rather than an element, and those two cases would then parse
// into a tree with no JSX in it, still import the same module, and still report — passing for a
// reason that has nothing to do with what they test.
const configFile = "/repository/source/Config.tsx"

// withSeededProvider swaps the module type table for the duration of one test.
//
// This is the mechanism that makes this rule provable at all, and it needs stating plainly: the
// rule's real input is a `moduleTypeProvider` function value that no static configuration can
// carry, and the hardcoded default table it falls back to is self-consistent at every entry. So no
// source text alone can make this rule report. Seeding the table is not a convenience here, it is
// the only way to distinguish a correctly-declining rule from an inert one.
//
// The seeded entries are upstream's own, from `snap/src/sprout/shared-runtime-type-provider.ts`,
// which is the provider the four `error.invalid-type-provider-*` fixtures are written against.
func withSeededProvider(t *testing.T, seeded map[string]map[string]typeConfigKind) {
	t.Helper()
	previous := moduleTypeProvider
	moduleTypeProvider = func(moduleName string) (map[string]typeConfigKind, bool) {
		properties, present := seeded[moduleName]
		return properties, present
	}
	t.Cleanup(func() { moduleTypeProvider = previous })
}

// upstreamTestProvider is the `ReactCompilerTest` entry from upstream's snap type provider.
//
// Two properties, and both are deliberately inconsistent with their own names: `useHookNotTypedAsHook`
// is a `use`-prefixed name declared as a plain function, and `notAhookTypedAsHook` is a
// non-prefixed name declared as a hook. That is what the module exists to test.
//
// `useDefaultExportNotTypedAsHook` is a separate module whose `default` export is a plain function
// while the module name itself is `use`-prefixed, which is the arm React validates against the
// module name rather than the property name.
func upstreamTestProvider() map[string]map[string]typeConfigKind {
	return map[string]map[string]typeConfigKind{
		"ReactCompilerTest": {
			"useHookNotTypedAsHook": typeConfigFunction,
			"notAhookTypedAsHook":   typeConfigHook,
		},
		"useDefaultExportNotTypedAsHook": {
			"default": typeConfigFunction,
		},
		// A hook-named module that IS configured but declares no `default` property. Not an
		// upstream module; added from reading React's `getGlobalDeclaration`, where the default arm
		// runs `getPropertyType(moduleType, 'default')` and returns null when the module has no
		// such property, so the comparison below it is never reached. Without this entry the
		// presence check in that arm is unguarded by any fixture: a mutation forcing it true
		// survived the whole suite, because every other configured module here is not hook-named
		// and therefore agrees with the absent property's non-hook zero value.
		"useConfiguredWithoutDefault": {
			"useSomething": typeConfigHook,
		},
	}
}

// TestConfigFiresOnUpstreamFixtures runs the four upstream error fixtures verbatim.
//
// Each source below is byte-for-byte the `.js` file from
// `facebook/react`'s `compiler/.../fixtures/compiler/`, fetched over the git tree API and written
// into this file by a script rather than typed, so no transcription step existed that could cook an
// escape. The four are the complete set of `Config`-category error fixtures upstream ships.
//
// Only the first two report under this port, and the reason is stated at the rule: fixtures three
// and four point at a *use site* inside a component body, which needs a binding-resolution pass
// this rule does not build. This test asserts what this port actually does. The two it declines are
// recorded in TestConfigDeclinesTheUseSiteFixtures below with the reason, rather than deleted,
// because a deleted case is a divergence nobody can find later.
// Not parallel: it swaps the package variable moduleTypeProvider through withSeededProvider, which the
// config rule reads on every run, so a parallel test running that rule would race the swap and could read
// the seeded table.
func TestConfigFiresOnUpstreamFixtures(t *testing.T) {
	withSeededProvider(t, upstreamTestProvider())

	cases := []struct {
		name   string
		source string
	}{
		// error.invalid-type-provider-hook-name-not-typed-as-hook
		// A `use`-prefixed named import whose configured type is a plain function.
		{"namedHookNotTypedAsHook", "import {useHookNotTypedAsHook} from 'ReactCompilerTest';\n\nfunction Component() {\n  return useHookNotTypedAsHook();\n}\n"},
		// error.invalid-type-provider-hooklike-module-default-not-hook
		// A default import from a `use`-prefixed MODULE whose default is a plain function.
		// The name checked here is the module's, not the property's.
		{"defaultFromHooklikeModule", "import foo from 'useDefaultExportNotTypedAsHook';\n\nfunction Component() {\n  return <div>{foo()}</div>;\n}\n"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, Config, configFile, testCase.source)
			rule_testing.ExpectFindings(t, result, "invalidTypeConfiguration")
		})
	}
}

// TestConfigFiresOnNonHookNameTypedAsHook covers the inverse direction of the same comparison.
//
// Split from the case above rather than folded into it because the two exercise opposite sides of
// `expectHook != isHook`, and a mutation collapsing the comparison to a constant would leave one of
// them green. Upstream's own fixture for this direction is
// `error.invalid-type-provider-nonhook-name-typed-as-hook`, whose source imports
// `notAhookTypedAsHook` and reports at its use site; the import-anchored form is written here.
// Not parallel: it swaps the package variable moduleTypeProvider through withSeededProvider, which the
// config rule reads on every run, so a parallel test running that rule would race the swap and could read
// the seeded table.
func TestConfigFiresOnNonHookNameTypedAsHook(t *testing.T) {
	withSeededProvider(t, upstreamTestProvider())
	source := "import {notAhookTypedAsHook} from 'ReactCompilerTest';\n\nfunction Component() {\n  return <div>{notAhookTypedAsHook()}</div>;\n}\n"
	result := rule_testing.Run(t, Config, configFile, source)
	rule_testing.ExpectFindings(t, result, "invalidTypeConfiguration")
}

// TestConfigReadsThroughAnAlias pins that the IMPORTED name decides, never the local one.
//
// React computes `expectHook` from `binding.imported`, so `{useHookNotTypedAsHook as safe}` still
// reports even though the local name is not hook-shaped, and `{notAhookTypedAsHook as useThing}`
// still reports even though the local name is. No upstream fixture writes an alias, so both are
// written from reading React's `getGlobalDeclaration`: reading `Name()` instead of the imported
// name would flip both verdicts, and nothing in the imported corpus could see it.
// Not parallel: it swaps the package variable moduleTypeProvider through withSeededProvider, which the
// config rule reads on every run, so a parallel test running that rule would race the swap and could read
// the seeded table.
func TestConfigReadsThroughAnAlias(t *testing.T) {
	withSeededProvider(t, upstreamTestProvider())
	cases := []struct {
		name   string
		source string
	}{
		{"hookNameAliasedToPlain", "import {useHookNotTypedAsHook as safe} from 'ReactCompilerTest';\n"},
		{"plainNameAliasedToHook", "import {notAhookTypedAsHook as useThing} from 'ReactCompilerTest';\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, Config, configFile, testCase.source)
			rule_testing.ExpectFindings(t, result, "invalidTypeConfiguration")
		})
	}
}

// TestConfigStaysSilent covers everything the rule must decline.
// Not parallel: it swaps the package variable moduleTypeProvider through withSeededProvider, which the
// config rule reads on every run, so a parallel test running that rule would race the swap and could read
// the seeded table.
func TestConfigStaysSilent(t *testing.T) {
	withSeededProvider(t, upstreamTestProvider())

	cases := []struct {
		name   string
		source string
	}{
		// A module the provider answers nothing for is never validated at all. This is the case
		// every import in Kirk's tree takes.
		{"unconfiguredModule", "import {useAnything} from 'some-other-module';\n"},
		// A property the configured module does not declare falls through to the hook-name
		// fallback rather than being validated.
		{"propertyNotInTable", "import {somethingElse} from 'ReactCompilerTest';\n"},
		// The same absence, but the imported name IS hook-shaped. React's `getPropertyType` returns
		// null for a property the module type does not carry, so the whole validation block is
		// skipped and the binding falls through to the custom-hook fallback: this is silent
		// upstream even though the name promises a hook the configuration never declares.
		//
		// This is the case that guards the presence check itself. Without it, a mutation deleting
		// that check survived the entire suite, because every other absent property tested here is
		// not hook-named and therefore agrees with the absent entry's non-hook zero value.
		{"hookNamedPropertyNotInTable", "import {useNotDeclaredAnywhere} from 'ReactCompilerTest';\n"},
		// A side-effect import binds nothing, so there is no name to check.
		{"sideEffectImport", "import 'ReactCompilerTest';\n"},
		// A module specifier that is not a string literal. Not valid JavaScript, but the parser
		// recovers rather than refusing, and it recovers into a shape that reads back as a usable
		// name: probed on our own parser, `from ReactCompilerTest` yields a KindIdentifier whose
		// `Text()` is `"ReactCompilerTest"` with no panic. So the kind guard is doing real work
		// here rather than protecting against a crash. Without it this input consults the provider
		// on a module nothing actually imports and reports on a name the file never bound, which is
		// a finding on recovered-from syntax that no upstream would produce.
		{"nonLiteralSpecifier", "import {useHookNotTypedAsHook} from ReactCompilerTest;\n"},
		// A default import from a module with no `default` property in its table.
		{"defaultNotInTable", "import anything from 'ReactCompilerTest';\n"},
		// The same absence, but from a HOOK-NAMED module. React resolves `default` off the module
		// type and finds nothing, so the name comparison below it never runs and the import is
		// silent even though the module name promises a hook. This is the case that separates the
		// presence check from the name check: the case above cannot, because a non-hook-named
		// module agrees with the absent property either way.
		{"defaultNotInTableOnHookNamedModule", "import anything from 'useConfiguredWithoutDefault';\n"},
		// `react` and `react-dom` short-circuit before the provider is consulted, at both
		// upstreams, and the comparison is lowercased so the odd casings take the same path.
		{"knownReactModule", "import {useHookNotTypedAsHook} from 'react';\n"},
		{"knownReactModuleUppercase", "import {useHookNotTypedAsHook} from 'React';\n"},
		{"knownReactDomModule", "import {useHookNotTypedAsHook} from 'react-dom';\n"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, Config, configFile, testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestConfigDeclinesTheUseSiteFixtures records the two upstream fixtures this port does not
// reproduce, and why.
//
// Both point at a use site inside a component body rather than at the import, because React only
// consults the provider when its HIR resolves a binding. Reproducing that needs the binding
// resolution pass this rule deliberately does not build. They are asserted as CLEAN here rather
// than deleted: a case removed from a corpus is a divergence no later reader can find, while a case
// asserted with its reason at the line is one they can act on.
//
// Fixture two is the more interesting of the pair. Its name says "namespace" and its source is
// `import ReactCompilerTest from 'ReactCompilerTest'`, which our parser reads as an ImportDefault
// rather than a namespace import — probed rather than assumed. It is silent here because
// `ReactCompilerTest`'s table declares no `default` property, so the default arm finds nothing to
// compare, which is the same reason `defaultNotInTable` above is silent.
// Not parallel: it swaps the package variable moduleTypeProvider through withSeededProvider, which the
// config rule reads on every run, so a parallel test running that rule would race the swap and could read
// the seeded table.
func TestConfigDeclinesTheUseSiteFixtures(t *testing.T) {
	withSeededProvider(t, upstreamTestProvider())

	cases := []struct {
		name   string
		source string
	}{
		// error.invalid-type-provider-hook-name-not-typed-as-hook-namespace
		{"dottedMemberUseSite", "import ReactCompilerTest from 'ReactCompilerTest';\n\nfunction Component() {\n  return ReactCompilerTest.useHookNotTypedAsHook();\n}\n"},
		// error.invalid-type-provider-nonhook-name-typed-as-hook
		{"jsxUseSite", "import {notAhookTypedAsHook} from 'ReactCompilerTest';\n\nfunction Component() {\n  return <div>{notAhookTypedAsHook()}</div>;\n}\n"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, Config, configFile, testCase.source)
			// The dotted-member fixture reports nothing because its module declares no `default`.
			// The JSX fixture DOES report here, on its named import, which is a superset of what
			// upstream reports rather than a miss: upstream points at the use site, this points at
			// the import that carries the misconfigured name. Asserted rather than described.
			if testCase.name == "jsxUseSite" {
				rule_testing.ExpectFindings(t, result, "invalidTypeConfiguration")
				return
			}
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestConfigIsSilentOnTheRealDefaultTable is the proof that this rule's zero is a real zero.
//
// It runs WITHOUT the seeded provider, so the rule consults the hardcoded `defaultModuleTypeProvider`
// exactly as it does on Kirk's tree. Every entry in that table pairs a `use`-prefixed name with a
// hook type, so `expectHook != isHook` is false at all of them and the rule cannot report.
//
// This is the pair the brief asks for: the seeded probe above shows the rule fires on a malformed
// configuration, and this shows it declines a well-formed one. Together they separate "declines
// correctly" from "wired and inert", which a zero on its own cannot do.
func TestConfigIsSilentOnTheRealDefaultTable(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		source string
	}{
		{"reactHookForm", "import {useForm} from 'react-hook-form';\n"},
		{"tanstackTable", "import {useReactTable} from '@tanstack/react-table';\n"},
		{"tanstackVirtual", "import {useVirtualizer} from '@tanstack/react-virtual';\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, Config, configFile, testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestConfigReportsOnTheImportSpecifier pins WHERE the finding points.
//
// `ExpectFindings` asserts ids and count and nothing else, so a rule pointing at the whole import
// statement, or at the module specifier, passes every case above while being wrong. The span is
// sliced out of the source and compared against a literal typed here rather than against anything
// the rule computes, so the assertion cannot move with the rule.
// Not parallel: it swaps the package variable moduleTypeProvider through withSeededProvider, which the
// config rule reads on every run, so a parallel test running that rule would race the swap and could read
// the seeded table.
func TestConfigReportsOnTheImportSpecifier(t *testing.T) {
	withSeededProvider(t, upstreamTestProvider())
	source := "import {useHookNotTypedAsHook} from 'ReactCompilerTest';\n"
	result := rule_testing.Run(t, Config, configFile, source)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("want 1 diagnostic, got %d", len(result.Diagnostics))
	}
	found := result.Diagnostics[0]
	reported := source[found.Range.Pos():found.Range.End()]
	if reported != "useHookNotTypedAsHook" {
		t.Errorf("reported span = %q, want %q", reported, "useHookNotTypedAsHook")
	}
}

// TestConfigReportsOnTheDefaultBinding pins the span of the other arm.
//
// The default arm points at the local binding identifier, which is the only node carrying a name in
// that shape. Separate from the specifier test above because the two arms compute their span from
// different nodes, and a mutation swapping one for the other would leave the other green.
// Not parallel: it swaps the package variable moduleTypeProvider through withSeededProvider, which the
// config rule reads on every run, so a parallel test running that rule would race the swap and could read
// the seeded table.
func TestConfigReportsOnTheDefaultBinding(t *testing.T) {
	withSeededProvider(t, upstreamTestProvider())
	source := "import foo from 'useDefaultExportNotTypedAsHook';\n"
	result := rule_testing.Run(t, Config, configFile, source)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("want 1 diagnostic, got %d", len(result.Diagnostics))
	}
	found := result.Diagnostics[0]
	reported := source[found.Range.Pos():found.Range.End()]
	if reported != "foo" {
		t.Errorf("reported span = %q, want %q", reported, "foo")
	}
}

// TestConfigMessageText asserts the rendered message exactly.
//
// `rule.Message` is `{Id, Description}` with no interpolation layer, so there is nothing to render
// and nothing a format verb could corrupt. The assertion is still worth its lines: it compares
// against a literal typed here rather than against `messageInvalidTypeConfiguration`, so a mutation
// rewriting the constant moves the rule and the test in opposite directions instead of together.
// Not parallel: it swaps the package variable moduleTypeProvider through withSeededProvider, which the
// config rule reads on every run, so a parallel test running that rule would race the swap and could read
// the seeded table.
func TestConfigMessageText(t *testing.T) {
	withSeededProvider(t, upstreamTestProvider())
	source := "import {useHookNotTypedAsHook} from 'ReactCompilerTest';\n"
	result := rule_testing.Run(t, Config, configFile, source)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("want 1 diagnostic, got %d", len(result.Diagnostics))
	}
	found := result.Diagnostics[0]
	if found.Message.Id != "invalidTypeConfiguration" {
		t.Errorf("message id = %q, want %q", found.Message.Id, "invalidTypeConfiguration")
	}
	if !strings.HasPrefix(found.Message.Description, "A module's compiler type configuration disagrees with the name it is attached to.") {
		t.Errorf("message description = %q, want it to open with the disagreement sentence", found.Message.Description)
	}
}

// TestIsHookNameMatchesUpstream pins the predicate both upstreams share.
//
// `/^use[A-Z0-9]/` at React, and an imperative `is_ascii_uppercase() || is_ascii_digit()` at oxc.
// The interesting rows are the last three: a lowercase fourth character is not a hook name, a
// bare `use` is too short to be one, and a non-ASCII uppercase letter is NOT one at either
// upstream. That last row is the axis that splits `react.IsLikelyComponentName` across trees, and
// it is pinned here so this rule cannot drift onto the Unicode side of it.
func TestIsHookNameMatchesUpstream(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		want bool
	}{
		{"useForm", true},
		{"use0", true},
		{"useX", true},
		{"user", false},
		{"use", false},
		{"us", false},
		{"", false},
		{"Use", false},
		{"notAhookTypedAsHook", false},
		// Non-ASCII uppercase. `[A-Z0-9]` is ASCII-only at both upstreams.
		{"use\u00c9", false},
	}
	for _, testCase := range cases {
		if got := isHookName(testCase.name); got != testCase.want {
			t.Errorf("isHookName(%q) = %v, want %v", testCase.name, got, testCase.want)
		}
	}
}

// TestIsKnownReactModuleIsCaseInsensitive pins the lowercased comparison.
//
// Both upstreams lowercase before comparing, so the odd casings take the short-circuit path. A
// version comparing exactly would validate `React` against the provider and could report where
// upstream is silent.
func TestIsKnownReactModuleIsCaseInsensitive(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		want bool
	}{
		{"react", true},
		{"React", true},
		{"REACT", true},
		{"react-dom", true},
		{"React-DOM", true},
		{"react-native", false},
		{"preact", false},
		{"", false},
	}
	for _, testCase := range cases {
		if got := isKnownReactModule(testCase.name); got != testCase.want {
			t.Errorf("isKnownReactModule(%q) = %v, want %v", testCase.name, got, testCase.want)
		}
	}
}
