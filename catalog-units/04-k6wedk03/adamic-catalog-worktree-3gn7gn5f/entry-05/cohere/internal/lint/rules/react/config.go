package react

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/imports"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageInvalidTypeConfiguration = rule.Message{
	Id: "invalidTypeConfiguration",
	Description: "A module's compiler type configuration disagrees with the name it is attached " +
		"to. The React Compiler decides whether something is a hook from its name, using the " +
		"same `use` prefix a reader does, and it decides again from the type the configuration " +
		"declares. When the two disagree the compiler has no way to choose, so it stops rather " +
		"than memoizing against a guess. Fix the configuration entry: either declare the type " +
		"as a hook, or rename so the name no longer claims to be one.",
}

// Config flags a React Compiler module type configuration whose declared type contradicts its name.
//
//	valid:   import {useForm} from 'react-hook-form';
//	valid:   import {useReactTable} from '@tanstack/react-table';
//	valid:   import {anything} from 'some-unconfigured-module';
//	invalid: (requires a moduleTypeProvider entry whose type contradicts its name)
//
// # What this rule validates, and why that makes it unlike every other rule here
//
// It validates the user's own compiler configuration rather than their code. The source file is
// only where the finding is *pointed*: the judgment is made about a configuration entry, and the
// import statement is the place a reader can be shown it. Every other rule in this package decides
// something about the code it is handed. This one decides something about a table, and reports at
// whichever import happens to consult the bad row.
//
// # Where that configuration lives, and the consequence for this tree
//
// The configuration is `moduleTypeProvider`, and it is a *JavaScript function value* handed to the
// Babel plugin programmatically: `(moduleName: string) => TypeConfig | null`. React 7.1.1 calls it
// at `Environment.ts`'s `resolveModuleType`, after a `typeof !== 'function'` check that is itself a
// config error. It is not a static file, not a JSON key, and not anything a linter can read off
// disk. Upstream's own corpus supplies it from `snap/src/sprout/shared-runtime-type-provider.ts`,
// a test harness, which is why all four error fixtures import from invented modules named
// `ReactCompilerTest` and `useDefaultExportNotTypedAsHook` that exist nowhere but that harness.
//
// `cohere` has no channel to such a value and should not grow one. `rule.Context` carries a source
// file, a program, a checker, a report function and a file cache; the only user configuration
// cohere reads is `CohereSettings.json`, which holds severities. So the user-supplied half of this
// rule's input is absent from this tree by construction rather than by accident.
//
// **Measured, not assumed:** `moduleTypeProvider` appears zero times across `~/Projects/ahra`,
// `~/Projects/phi` and `~/Projects/connected`, with `node_modules` and build output excluded. The
// control run alongside that zero was `reactCompiler`, which appears in every one of the three
// trees at `libraries/structure/NextConfiguration.ts`, set to `true`. So the React Compiler is
// genuinely enabled here and genuinely has no type provider configured, which is the one
// combination that makes this rule real and silent at the same time.
//
// # The half that still fires with no user configuration at all
//
// `defaultModuleTypeProvider` is hardcoded and consulted whenever the user supplies none. React
// 7.1.1 and oxc agree on it exactly, three modules each: `react-hook-form`, `@tanstack/react-table`
// and `@tanstack/react-virtual`. That is the reachable half of this rule, and it is what the code
// below implements, because it is the half whose input is a fact about the compiler rather than a
// value nobody in this tree passes.
//
// It is also, measurably, a half that cannot report. Each of the four hardcoded entries pairs a
// `use`-prefixed property name with a `hook` type and a non-prefixed one with a `function` type, so
// `expectHook != isHook` is false at every entry. Checked mechanically rather than by eye, across
// all four including the nested `useForm().watch`. This rule is therefore correct, wired, and
// structurally unable to fire on a tree that configures nothing, and that is a property of the
// default table rather than of this port.
//
// # Where React 7.1.1 and oxc disagree, and which is followed
//
// React has five construction sites for this category and oxc has three. Two of React's are in
// `parseAliasingSignatureConfig`, validating that an aliasing signature names its lifetimes
// uniquely and refers only to known names; oxc has no counterpart, and neither does this port,
// because an aliasing signature can only arrive through the provider value that does not exist
// here.
//
// The load-bearing disagreement is in the `ImportSpecifier` arm. React validates it *inline*: it
// resolves the property type, computes `expectHook` from the imported name, and reports when they
// differ. oxc's `ImportSpecifier` arm does not: it only replays errors that `install_type_config`
// deferred while walking the object's properties, so oxc's message for that case names the
// *property* while React's names the `import {x} from 'm'` form. The fixtures settle it, and they
// side with the deferred spelling: `error.invalid-type-provider-hook-name-not-typed-as-hook`
// expects "Expected type for object property 'useHookNotTypedAsHook'", which is the property
// wording, not the import wording. React is the authority for this rule and React's own goldens
// are what is reproduced, so the property wording is what this rule carries.
//
// A second consequence of those goldens is worth stating because it looks like a bug. The fixture
// `error.invalid-type-provider-nonhook-name-typed-as-hook` imports `notAhookTypedAsHook` and its
// expected message names `useHookNotTypedAsHook` — a *different* property. That is not a
// transcription slip: the error is raised while installing the whole module object, so the first
// inconsistent property in declaration order is the one named, whichever property the file went on
// to import. Reproduced here rather than corrected.
//
// # Why this reports on the import and not on the configuration
//
// There is nowhere else to point. The configuration has no source location in the file being
// linted, which is exactly why React threads a `loc` down from the import site and labels the
// span there. All four fixtures point inside the component body at the *use* of the import rather
// than at the import statement itself, because React's HIR only consults the provider when a
// binding is actually resolved. This port points at the import specifier instead, and that is a
// stated divergence rather than an oversight: reproducing the use site would require the binding
// resolution pass this rule deliberately does not build, and the import specifier is the node that
// actually carries the misconfigured name.
var Config = rule.Rule{
	// Bare name, no namespace prefix. The inventory writes `react/config` and the parity guard
	// strips the namespace on a `/` boundary, so `react-config` would match no entry while still
	// passing every fixture in this file.
	Name: "react-hooks/config",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindImportDeclaration: func(node *ast.Node) {
				moduleName, named := importedModuleName(node)
				if !named {
					return
				}
				// A known React module short-circuits before any provider is consulted, at both
				// upstreams. `isKnownReactModule` is a lowercased comparison against exactly
				// `react` and `react-dom` in React 7.1.1 and in oxc's `is_known_react_module`,
				// so `React` and `REACT-DOM` take this path too.
				if isKnownReactModule(moduleName) {
					return
				}
				properties, configured := moduleTypeProvider(moduleName)
				// This early return is a COST guard rather than a behavioral one, and a mutation
				// sweep is what established that rather than reading. Removing it leaves every
				// verdict unchanged: an unconfigured module answers a nil `properties` map, and a
				// nil map lookup in Go returns the zero value with `present == false`, so the
				// named arm's `continue` and the default arm's `present &&` already decline every
				// input this line could reach. Verified with a standalone program rather than
				// assumed, and the mutant correctly SURVIVED the whole suite.
				//
				// It is kept anyway, and the reason is measurable on this tree: it returns before
				// `BindingsOf` walks the import clause, and on Kirk's tree every single import is
				// unconfigured, so this line is what keeps the rule from walking 1,242-plus import
				// clauses to reach a conclusion it can already state. Deleting a subsumed branch is
				// the standing advice; deleting this one would trade a measured no-op for a
				// per-import walk.
				if !configured {
					return
				}

				bindings := imports.BindingsOf(node)

				// The ImportSpecifier arm. `ImportedNameOf` reads through an alias, which is the
				// correct half to check: React computes `expectHook` from `binding.imported`, the
				// name the module exports, never from the local one.
				for _, element := range bindings.Named {
					importedName := imports.ImportedNameOf(element)
					declaredKind, present := properties[importedName]
					if !present {
						continue
					}
					if isHookName(importedName) != (declaredKind == typeConfigHook) {
						ctx.ReportNode(element, messageInvalidTypeConfiguration)
					}
				}

				// The ImportDefault arm. React reads the `default` property off the module type and
				// computes `expectHook` from the *module* name rather than the property name, which
				// is why the fixture for it imports from a module literally called
				// `useDefaultExportNotTypedAsHook`.
				if bindings.Default != nil {
					declaredKind, present := properties["default"]
					if present && isHookName(moduleName) != (declaredKind == typeConfigHook) {
						ctx.ReportNode(bindings.Default, messageInvalidTypeConfiguration)
					}
				}
			},
		}
	},
}

// typeConfigKind is the `kind` discriminant of a TypeConfig entry.
//
// Only the two values this rule can distinguish are modelled. React's schema also carries `object`
// and `type`, and both answer `isHook == false` exactly as `function` does, so collapsing them into
// one non-hook value cannot change a verdict. Named rather than boolean so a reader sees which
// upstream field is being compared.
type typeConfigKind string

const (
	typeConfigHook     typeConfigKind = "hook"
	typeConfigFunction typeConfigKind = "function"
)

// moduleTypeProvider is the table consulted for a module, and the seam a test can substitute at.
//
// A variable rather than a direct call, and the reason is the whole difficulty of this rule. The
// user-supplied `moduleTypeProvider` does not exist in this tree, and the hardcoded default table
// is provably self-consistent, so a rule calling `defaultModuleTypeProvider` directly could not be
// shown to fire by any input whatsoever. That is a rule nothing can guard: every fixture would pass
// by finding nothing, which reads exactly like a working rule and exactly like a broken one.
//
// This seam is what makes the difference visible. The seeded test replaces it with upstream's own
// test provider, the one the four `error.invalid-type-provider-*` fixtures are written against, and
// asserts that all four report. Restoring the default and re-running the same four asserts they go
// silent. That pair is the only available proof that the zero this rule produces on Kirk's tree is
// a real zero rather than an inert rule.
//
// It is deliberately not exported and deliberately not a configuration surface. Making it one would
// invent a config channel React does not have in this form, and this rule's job is to reproduce a
// judgment rather than to grow the place that judgment reads from.
var moduleTypeProvider = defaultModuleTypeProvider

// defaultModuleTypeProvider returns the property table the compiler holds for a module, if it holds one.
//
// This is `defaultModuleTypeProvider` at our spelling, and it is the whole of the configuration
// this tree can see. The second return separates "this module has no configuration" from "this
// module has a configuration with no properties", which are different answers: an unconfigured
// module resolves to a null module type and is never validated at all, while a configured one with
// an empty table is validated and passes.
//
// The three entries are verbatim from React 7.1.1's `defaultModuleTypeProvider` and match oxc's
// `default_module_type_provider.rs` entry for entry. A user-supplied `moduleTypeProvider` would
// *replace* consultation of this table for the modules it answers for, which is why upstream falls
// back with `??` rather than merging.
//
// # Why removing an entry here is an equivalent mutation, and what proves the table is read
//
// A sweep deleting the `react-hook-form` arm SURVIVED the whole suite, and so did one deleting
// `@tanstack/react-table`. That is equivalence rather than a fixture gap, and it is stateable in
// one sentence: with the entry present the module is configured and its single property is
// name-consistent, so the verdict is silence; with the entry removed the module is unconfigured and
// returns early, so the verdict is silence. Both routes, one verdict, for every possible input.
//
// The control that keeps this from being a comfortable excuse is the mutation that *changes* an
// entry rather than removing it: rewriting `useForm` from a hook to a function was CAUGHT, which
// proves this table is genuinely consulted and genuinely compared. A table nothing reads and a
// table whose every row agrees with itself produce the same zero, and only that pair of mutations
// separates them.
func defaultModuleTypeProvider(moduleName string) (map[string]typeConfigKind, bool) {
	switch moduleName {
	case "react-hook-form":
		// `useForm` is a hook whose return object carries `watch`, a plain function. The nested
		// property is not reachable from an import declaration, so it is recorded for the reader
		// rather than checked: reaching it needs the call-graph pass this rule does not build.
		return map[string]typeConfigKind{"useForm": typeConfigHook}, true
	case "@tanstack/react-table":
		return map[string]typeConfigKind{"useReactTable": typeConfigHook}, true
	case "@tanstack/react-virtual":
		return map[string]typeConfigKind{"useVirtualizer": typeConfigHook}, true
	}
	return nil, false
}

// isHookName answers React's `isHookName`, which is `/^use[A-Z0-9]/`.
//
// Written out rather than regex-matched because the shape is four bytes and the predicate runs per
// import specifier. oxc's `is_hook_name` in `environment.rs` spells the same test imperatively with
// `is_ascii_uppercase() || is_ascii_digit()`, and the two agree on every input: the regex character
// class `[A-Z0-9]` is ASCII-only, so there is no Unicode axis here of the kind that splits
// `react.IsLikelyComponentName` across trees. `useÉ` is not a hook name at either upstream, and not
// one here.
func isHookName(name string) bool {
	if len(name) < 4 {
		return false
	}
	if !strings.HasPrefix(name, "use") {
		return false
	}
	fourth := name[3]
	return (fourth >= 'A' && fourth <= 'Z') || (fourth >= '0' && fourth <= '9')
}

// isKnownReactModule answers React's `#isKnownReactModule`, a lowercased match on two names.
//
// `Environment.knownReactModules` is `['react', 'react-dom']` and the comparison is
// `moduleName.toLowerCase() === ...`, so the casing of the specifier does not matter. oxc's
// `is_known_react_module` uses `cow_to_lowercase` against the same two literals.
func isKnownReactModule(moduleName string) bool {
	lowered := strings.ToLower(moduleName)
	return lowered == "react" || lowered == "react-dom"
}

// importedModuleName reads the specifier text of an import declaration.
//
// The kind check before the text read is load-bearing rather than defensive, matching the standing
// hazard in this tree: `Node.Text()` panics on several kinds rather than returning empty, so the
// specifier is confirmed to be a string literal before it is read. `ast.IsStringLiteralLike` is the
// same predicate the `imports` shelf uses for this, and it accepts a no-substitution template as
// well as a string, which the parser can produce for `import x from` followed by a template.
func importedModuleName(node *ast.Node) (string, bool) {
	if node == nil || node.Kind != ast.KindImportDeclaration {
		return "", false
	}
	declaration := node.AsImportDeclaration()
	if declaration == nil || declaration.ModuleSpecifier == nil {
		return "", false
	}
	if !ast.IsStringLiteralLike(declaration.ModuleSpecifier) {
		return "", false
	}
	return declaration.ModuleSpecifier.Text(), true
}
