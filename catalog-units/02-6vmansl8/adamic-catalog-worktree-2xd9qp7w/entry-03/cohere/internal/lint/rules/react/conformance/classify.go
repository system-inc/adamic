package react_conformance

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// ShippedRules are the upstream rule names cohere has an implementation for, mapped to the name
// cohere registers that rule under.
//
// Keyed by UPSTREAM's rule name rather than cohere's, because the join to the corpus goes through
// upstream's `getRuleForCategory`.
//
// # The value side was wrong for every entry, and nothing could see it
//
// Until 2026-08-24 every value here read `react/<name>`, on the strength of a doc comment claiming
// the two sides "happen to agree on spelling" and that the map was explicit so a rename would be "a
// visible edit rather than a silent unjoin". Measured against `registry.All()`: not one of those
// eleven strings is a registered rule. The react-compiler rules register BARE — `globals`,
// `rules-of-hooks`, `use-memo` — and the namespace was never part of the name.
//
// The unjoin was silent for the reason the comment did not anticipate: nothing in this package ever
// READ the value. `Classify` asks only `_, found := ShippedRules[ruleName]`, so the right-hand side
// was decorative, and a decorative field that claims to be a join is worse than no field, because
// it reads as checked. `TestShippedRuleNamesAreRegistered` now reads it, which is what makes this
// map's second column a claim rather than a caption. That is the open finding `#5vyycz7` measured:
// the parity guard cannot see a namespaced spelling because this side was never executed.
//
// `hooks` maps to cohere's `rules-of-hooks`, which is the one place the two names differ.
var ShippedRules = map[string]string{
	"config":               "react-hooks/config",
	"error-boundaries":     "react-hooks/error-boundaries",
	"gating":               "react-hooks/gating",
	"globals":              "react-hooks/globals",
	"hooks":                "react-hooks/rules-of-hooks",
	"immutability":         "react-hooks/immutability",
	"incompatible-library": "react-hooks/incompatible-library",
	"purity":               "react-hooks/purity",
	"refs":                 "react-hooks/refs",
	"set-state-in-effect":  "react-hooks/set-state-in-effect",
	"set-state-in-render":  "react-hooks/set-state-in-render",
	"static-components":    "react-hooks/static-components",
	"unsupported-syntax":   "react-hooks/unsupported-syntax",
	"use-memo":             "react-hooks/use-memo",
	"void-use-memo":        "react-hooks/void-use-memo",
}

// TypeAwareRules are the shipped rules whose DECISION depends on resolving React's own types.
//
// This list is what makes `VerdictUnresolvableTypes` a measurement rather than an excuse: only a
// rule that actually reads types can be excused by a fixture that has none. A syntax-only rule
// producing nothing on an import-less fixture is a failure, and must be scored as one. Getting this
// wrong in the generous direction inflates the score with excuses, so the bar for entry is narrow
// and stated below.
//
// # `NeedsTypeChecker` is the wrong test, and it was the obvious one
//
// Nine of the fifteen react-compiler rules declare `NeedsTypeChecker`. Reading membership off that
// flag would put `globals`, `use-memo`, `purity` and `unsupported-syntax` in here and hand each of
// them a blanket excuse on every import-less fixture. Measured at the call sites, the flag covers
// two different questions:
//
//	GetSymbolAtLocation      binding resolution: "is this identifier declared anywhere in scope?"
//	                         Answers correctly with no `@types/react` present, because the question
//	                         is about the fixture's own bindings. `globals.go:616`, `use_memo.go:424`,
//	                         `purity.go:565`, `unsupported_syntax.go:234`. NOT an excuse.
//	GetTypeAtLocation        reads a type and then its alias symbol, which is `any` without React's
//	                         declarations, so the deciding branch cannot be reached.
//	                         `set_state_in_render.go:450`, `refs.go:359`. This IS the excuse.
//
// So membership is by call site, verified rather than declared. `set-state-in-effect` is kept from
// the previous revision on the same grounds (it ships fixtures carrying a local `react.d.ts` at
// `set_state_in_effect_test.go:57` for exactly this reason), and `refs` joins it on the measurement
// above.
//
// `static-components` was in this map and is REMOVED. It declares `NeedsTypeChecker: true` at
// `static_components.go:100` and then never touches `ctx.TypeChecker` anywhere in its body —
// grepped for the selector, zero hits. It was in here on the strength of a declaration it does not
// use, which is the generous direction this comment warns about, arriving through the flag rather
// than through a guess.
var TypeAwareRules = map[string]bool{
	"refs":                true,
	"set-state-in-effect": true,
	"set-state-in-render": true,
}

// reactImportPattern matches an import or require of `react`.
//
// Written to match the module specifier exactly rather than the substring `react`, which would also
// match `react-dom`, `@testing-library/react`, and a local `./react-utils`. That over-match would
// move fixtures out of `VerdictUnresolvableTypes` and into `failed`, inflating the failure count
// with cases that are really corpus artifacts — an error in the direction that looks like rigor.
var reactImportPattern = regexp.MustCompile(`(?m)(?:^|\n)\s*import\b[^\n;]*?['"]react['"]|require\(\s*['"]react['"]\s*\)`)

// hookCallPattern matches a call to something named like a hook.
var hookCallPattern = regexp.MustCompile(`\buse[A-Z]\w*\s*\(`)

// ResolvesReactTypes reports whether a fixture could resolve `@types/react`.
func (f Fixture) ResolvesReactTypes() bool {
	return reactImportPattern.MatchString(f.Source)
}

// CallsHook reports whether a fixture calls anything named like a hook.
func (f Fixture) CallsHook() bool {
	return hookCallPattern.MatchString(f.Source)
}

// StatedDivergence names a fixture cohere deliberately does not reproduce, with where the boundary
// is written down.
//
// Entries are the machine-checked ones. A divergence that lives only in a doc comment is not in
// here, because the entry's whole value is that something else fails when the divergence stops
// being true — otherwise this map is a place to park failures with a nice label on them.
type StatedDivergence struct {
	// Fixture is the corpus fixture, by name.
	Fixture string

	// Boundary is the test that holds the divergence in place.
	Boundary string

	// Reason is what cohere decided and why.
	Reason string
}

// statedDivergences names the fixtures cohere deliberately does not reproduce.
//
// The stated boundaries cohere's react rules carry were all measured against oxc's inline corpus or
// against the executable ESLint rule, and every one of them was checked against this corpus while
// building this file. None of them lands on a fixture in the 325:
//
//   - exhaustive-deps at 15 of 17 kinds (`TestExhaustiveDepsScopeIsStated`, eight named cases): its
//     rule is `exhaustive-deps`, an ESLint rule. The corpus's nearest categories are
//     `memo-dependencies` (14 fixtures) and `exhaustive-effect-dependencies` (4), which are the
//     COMPILER's dependency validators and a different mechanism. Neither is a rule cohere ships.
//   - set-state-in-effect's ref exemption, pinned as a fixture that fails when fixed: that rule has
//     zero error-named fixtures in this corpus, so the divergence cannot be exercised here.
//   - no-unused-vars at 301 of 303: not a react rule and not in this corpus at all.
//
// The eleven entries below are `globals` fixtures, and they are here because the rule was actually
// RUN against them (`internal/react_conformance_score`) rather than because a doc comment predicted
// them. Every one declines on a construct its doc comment already names as out of scope: eight on
// the nested-function subset that needs an inter-procedural effects pass, three on the compilation
// gate that requires JSX or a hook call in the body.
//
// This is the category the brief was most worried about, and the measurement justifies the worry
// exactly: a raw score would have reported `react/globals` at 0 of 11 and read as a broken rule. It
// is a rule with a written, machine-checked boundary meeting a corpus of minimal repros that sit
// entirely outside it.
//
// What that means for the number is worth stating plainly rather than leaving implicit: 0 passed
// and 11 excluded is NOT the same claim as 11 passed. The rule is unmeasured against this corpus,
// not validated by it. The honest reading is that React's error goldens for `globals` and cohere's
// `globals` do not overlap, and closing that gap is a scope decision rather than a bug fix.
var statedDivergences = map[string]StatedDivergence{
	// The nested-function subset. `react/globals` reports only writes lexically inside the
	// compilation root's own body, never inside a nested function, because deciding whether a
	// nested function runs during render is an inter-procedural effects analysis rather than
	// anything syntactic. The rule's own doc comment measures the boundary and shows it is not a
	// statable rule: `const f = () => {g=1}; f()` reports upstream, `const f = () => {g=1};
	// return <Foo cb={f}/>` does not, and `return <Foo>{f}</Foo>` does again. `TestGlobalsBoundary`
	// pins the subset.
	//
	// Each of these was confirmed by running the real rule over the real golden through
	// `internal/react_conformance_score` and reading which construct it declined on, rather than
	// inferred from the doc comment.
	"error.assign-global-in-component-tag-function.js": {
		Fixture:  "error.assign-global-in-component-tag-function.js",
		Boundary: "TestGlobalsBoundary",
		Reason:   "the write is inside a nested arrow, which cohere's globals reports only for the compilation root's own body",
	},
	"error.assign-global-in-jsx-children.js": {
		Fixture:  "error.assign-global-in-jsx-children.js",
		Boundary: "TestGlobalsBoundary",
		Reason:   "the write is inside a nested arrow passed as JSX children, which needs the effects pass to prove it runs during render",
	},
	"error.invalid-global-reassignment-indirect.js": {
		Fixture:  "error.invalid-global-reassignment-indirect.js",
		Boundary: "TestGlobalsBoundary",
		Reason:   "the write is two nested arrows deep, reachable only through inter-procedural aliasing",
	},
	"error.reassign-global-fn-arg.js": {
		Fixture:  "error.reassign-global-fn-arg.js",
		Boundary: "TestGlobalsBoundary",
		Reason:   "the write is inside a nested arrow passed to a function, which needs the effects pass",
	},
	"error.reassignment-to-global-indirect.js": {
		Fixture:  "error.reassignment-to-global-indirect.js",
		Boundary: "TestGlobalsBoundary",
		Reason:   "the write is inside a nested arrow that is then called, which needs the effects pass",
	},

	// The compilation gate. cohere's react rules fire only inside a function React Compiler would
	// compile, which for a `Component`-named function requires JSX or a hook call in the body. The
	// gate was measured over seventeen probe rounds against React's own rule and is shared with
	// `unsupported-syntax`. These fixtures are components by NAME with neither JSX nor a hook, so
	// upstream's test harness compiles them and cohere's gate correctly does not.
	//
	// This is the one divergence here that is arguably worth closing, and it is recorded rather
	// than hidden for exactly that reason: it is invisible in application code, where a component
	// without JSX or hooks is not a component, and visible only against a corpus of minimal
	// repros.
	"error.reassignment-to-global.js": {
		Fixture:  "error.reassignment-to-global.js",
		Boundary: "TestGlobalsGateIsSharedWithUnsupportedSyntax",
		Reason:   "the function has neither JSX nor a hook call, so cohere's compilation gate declines it while upstream's test harness compiles it anyway",
	},
	"new-mutability/error.reassignment-to-global.js": {
		Fixture:  "new-mutability/error.reassignment-to-global.js",
		Boundary: "TestGlobalsGateIsSharedWithUnsupportedSyntax",
		Reason:   "the function has neither JSX nor a hook call, so cohere's compilation gate declines it",
	},
	"new-mutability/error.reassignment-to-global-indirect.js": {
		Fixture:  "new-mutability/error.reassignment-to-global-indirect.js",
		Boundary: "TestGlobalsBoundary",
		Reason:   "the write is inside a nested arrow that is then called, which needs the effects pass",
	},

	// Destructuring assignment to an undeclared target. Upstream reports these under Globals; the
	// write is a destructuring pattern rather than a simple or compound assignment.
	"error.invalid-destructure-assignment-to-global.js": {
		Fixture:  "error.invalid-destructure-assignment-to-global.js",
		Boundary: "TestGlobalsBoundary",
		Reason:   "an array-destructuring assignment whose target is an undeclared global, measured as declined by the rule",
	},
	"error.invalid-destructure-to-local-global-variables.js": {
		Fixture:  "error.invalid-destructure-to-local-global-variables.js",
		Boundary: "TestGlobalsBoundary",
		Reason:   "a mixed destructuring assignment where only one target is a global, measured as declined by the rule",
	},

	// The `config` rule's input does not exist in this tree, by construction rather than by
	// accident. All four of upstream's Config error fixtures import from modules that exist only
	// inside upstream's own test harness — `ReactCompilerTest` and `useDefaultExportNotTypedAsHook`
	// — and the diagnostic is produced by consulting a `moduleTypeProvider`, which React 7.1.1
	// takes as a JavaScript FUNCTION VALUE handed to the Babel plugin programmatically. It is not a
	// file, not a JSON key, and not anything a linter can read off disk; upstream supplies it from
	// `snap/src/sprout/shared-runtime-type-provider.ts`.
	//
	// cohere implements the half whose input is a fact about the compiler rather than a value
	// nobody passes: `defaultModuleTypeProvider`, three hardcoded modules. `configuration.go`'s doc
	// comment measures that half and shows it is structurally unable to fire, because every
	// hardcoded entry pairs a `use`-prefixed name with a `hook` type, so the contradiction the rule
	// looks for is false at every row.
	//
	// So this is a rule that is correct, wired, and silent, meeting four goldens whose input is a
	// test harness. Recorded here rather than counted as failure, because calling it a defect would
	// mean the only way to score would be to grow a channel for a user-supplied function value that
	// this codebase deliberately does not have.
	"error.invalid-type-provider-hook-name-not-typed-as-hook.js": {
		Fixture:  "error.invalid-type-provider-hook-name-not-typed-as-hook.js",
		Boundary: "TestConfigFixturesNeedAModuleTypeProvider",
		Reason:   "the diagnostic needs a `moduleTypeProvider` function value from upstream's test harness, which cohere has no channel to receive; measured as reporting nothing",
	},
	"error.invalid-type-provider-hook-name-not-typed-as-hook-namespace.js": {
		Fixture:  "error.invalid-type-provider-hook-name-not-typed-as-hook-namespace.js",
		Boundary: "TestConfigFixturesNeedAModuleTypeProvider",
		Reason:   "imports the harness-only module `ReactCompilerTest`, whose type configuration cohere has no channel to receive",
	},
	"error.invalid-type-provider-hooklike-module-default-not-hook.js": {
		Fixture:  "error.invalid-type-provider-hooklike-module-default-not-hook.js",
		Boundary: "TestConfigFixturesNeedAModuleTypeProvider",
		Reason:   "imports the harness-only module `useDefaultExportNotTypedAsHook`, whose type configuration cohere has no channel to receive",
	},
	"error.invalid-type-provider-nonhook-name-typed-as-hook.js": {
		Fixture:  "error.invalid-type-provider-nonhook-name-typed-as-hook.js",
		Boundary: "TestConfigFixturesNeedAModuleTypeProvider",
		Reason:   "imports the harness-only module `ReactCompilerTest`, whose type configuration cohere has no channel to receive",
	},

	// Three `set-state-in-render` goldens recorded under an upstream feature flag that defaults to
	// OFF, so the expectation describes a compiler cohere does not implement rather than a finding
	// cohere missed.
	//
	// Measured in React's own unminified bundle rather than inferred:
	// `enableUseKeyedState: z.boolean().default(false)` at line 31626 and
	// `enableTreatSetIdentifiersAsStateSetters: z.boolean().default(false)` at 31647. Without the
	// flag, upstream itself reports nothing on these inputs.
	//
	// This is the narrow, real half of the pragma question. Most pragmas on this corpus restate a
	// default that is already on (`validateNoSetStateInRender` defaults TRUE), which is why the
	// engine deliberately does not gate on pragmas in general. These three are the exception, and
	// they were found by enumerating every wired fixture carrying a pragma measured as
	// `default(false)` — exactly three across all seven wired rules, all in this one rule.
	"error.invalid-setstate-unconditional-with-keyed-state.js": {
		Fixture:  "error.invalid-setstate-unconditional-with-keyed-state.js",
		Boundary: "TestOptInPragmaFixturesAreEnumerated",
		Reason:   "the golden was recorded under `@enableUseKeyedState`, an upstream flag defaulting to false that cohere does not implement",
	},
	"error.invalid-unconditional-set-state-prop-in-render.js": {
		Fixture:  "error.invalid-unconditional-set-state-prop-in-render.js",
		Boundary: "TestOptInPragmaFixturesAreEnumerated",
		Reason:   "the golden was recorded under `@enableTreatSetIdentifiersAsStateSetters`, an upstream flag defaulting to false, which makes any `setX` identifier a state setter",
	},
	"error.invalid-unconditional-set-state-hook-return-in-render.js": {
		Fixture:  "error.invalid-unconditional-set-state-hook-return-in-render.js",
		Boundary: "TestOptInPragmaFixturesAreEnumerated",
		Reason:   "the golden was recorded under `@enableTreatSetIdentifiersAsStateSetters`, an upstream flag defaulting to false",
	},

	// The gate again, in its sharpest form. `useFoo` is hook-NAMED but calls no hook, and cohere's
	// gate requires a hook CALL in the body, not a hook name on the function. The rule's own doc
	// comment states this directly at globals.go:195 (`function useFoo() { useState(0); g = 1; }
	// reports, hook`), and it was re-confirmed by probe here rather than read off that line:
	//
	//	function useFoo() { g += 1; return g; }        silent  <- this fixture
	//	function useFoo() { useState(0); g = 1; }      reports
	//	function Component() { g += 1; return <div/>; } reports
	//
	// The middle and bottom probes are the controls. Without them "the rule is silent" would be
	// consistent with the compound-assignment path being broken, which is a completely different
	// defect and the one worth ruling out — the doc comment makes a specific claim about `+=`
	// pointing at the whole assignment expression, so a silent `+=` had to be shown to be the gate
	// rather than the span logic.
	"error.update-global-should-bailout.tsx": {
		Fixture:  "error.update-global-should-bailout.tsx",
		Boundary: "TestGlobalsGateIsSharedWithUnsupportedSyntax",
		Reason:   "`useFoo` is hook-named but calls no hook, and cohere's compilation gate requires a hook call in the body; probed with controls showing the same write reports from a function that does call one",
	},

	// Three `preserve-manual-memoization` goldens that upstream's own fixture headers call
	// mistakes. Staying silent on them is the correct answer rather than a missed finding, so
	// scoring them as targets makes the rule's number describe the corpus rather than the rule.
	//
	// Quoted from the fixtures rather than characterised:
	//
	//	todo-useCallback-captures-invalidating-value
	//	  "False positive: We currently bail out on this because we don't understand
	//	   that `() => [x]` gets pruned because `x` always invalidates."
	//
	//	false-positive-useMemo-dropped-infer-always-invalidating
	//	  "This is technically a false positive as the useMemo in source was
	//	   effectively a no-op"
	//
	//	todo-repro-unmemoized-callback-captured-in-context-variable
	//	  carries upstream's `todo-repro` prefix, its marker for a known-wrong reproduction
	//
	// This is a different kind of entry from the pragma three above, and the difference is worth
	// stating. Those describe a compiler cohere does not implement. These describe a verdict
	// upstream itself disowns, so reproducing them would mean copying a bug -- which is the one
	// case where matching the answer key is the wrong goal.
	//
	// `error.useMemo-aliased-var.ts` is deliberately NOT here. Its header says "This is technically
	// a false positive, but source is already breaking `exhaustive-deps`", which is upstream
	// describing an input it considers invalid rather than a verdict it considers wrong, and cohere
	// fires on it today. The distinction is the second clause: a fixture upstream calls a mistake
	// outright is excluded, one it blames on the input is not.
	"preserve-memo-validation/error.todo-useCallback-captures-invalidating-value.ts": {
		Fixture:  "preserve-memo-validation/error.todo-useCallback-captures-invalidating-value.ts",
		Boundary: "TestPreserveManualMemoizationAgainstGoldens",
		Reason:   "upstream's own header calls this a false positive: it bails out because it does not understand that `() => [x]` gets pruned when `x` always invalidates",
	},
	"preserve-memo-validation/error.false-positive-useMemo-dropped-infer-always-invalidating.ts": {
		Fixture:  "preserve-memo-validation/error.false-positive-useMemo-dropped-infer-always-invalidating.ts",
		Boundary: "TestPreserveManualMemoizationAgainstGoldens",
		Reason:   "upstream's own header calls this a false positive: the `useMemo` in source was effectively a no-op",
	},
	"error.todo-repro-unmemoized-callback-captured-in-context-variable.tsx": {
		Fixture:  "error.todo-repro-unmemoized-callback-captured-in-context-variable.tsx",
		Boundary: "TestPreserveManualMemoizationAgainstGoldens",
		Reason:   "carries upstream's `todo-repro` prefix, its marker for a reproduction it knows to be wrong",
	},
}

// divergenceRuleIsShipped reports whether any rule a fixture is attributed to is registered.
//
// The gate on answering a stated divergence early. See its call site for why the two cases differ.
func divergenceRuleIsShipped(rules []string) bool {
	// No attribution means there is no unshipped rule to defer to, so the divergence is answered
	// here as it always was. This is the shape `TestStatedDivergenceCategoryCanHoldAnEntry` uses,
	// and it is also the honest answer: deferring would send a fixture nobody owns to a verdict
	// about rule registration.
	if len(rules) == 0 {
		return true
	}
	for _, ruleName := range rules {
		if _, found := ShippedRules[ruleName]; found {
			return true
		}
	}
	return false
}

// StatedDivergenceNames returns the fixtures recorded as deliberate divergences, keyed by name.
//
// Exported so the boundary tests in `react_conformance_score` can assert that a divergence they hold
// in place is actually recorded here. Without it the two halves of a stated divergence — the entry
// and the test named in its `Boundary` field — could drift apart silently, which would leave the
// entry as a label with nothing behind it. That is precisely what the `StatedDivergence` doc
// comment says must not happen.
func StatedDivergenceNames() map[string]StatedDivergence {
	names := make(map[string]StatedDivergence, len(statedDivergences))
	for name, divergence := range statedDivergences {
		names[name] = divergence
	}
	return names
}

// Classify decides which category a fixture belongs to for the rules cohere ships.
//
// Order matters and is defended here, because each step short-circuits the ones below it and a
// different order produces a different number from the same inputs.
//
//  1. Flow first. A fixture cohere cannot parse cannot be judged on any other axis, and calling it
//     anything else would be a claim about a file that was never read.
//  2. Stated divergence next, ahead of everything except parseability. A deliberate decision should
//     not be reclassified as a corpus artifact just because the fixture also happens to lack an
//     import; the decision is the more specific fact.
//  3. Unreachable, then no-rule-shipped. Both produce silence, and the distinction is whether the
//     question is askable at all.
//  4. Unresolvable types, but only for a type-aware rule. A syntax rule gets no such excuse.
//  5. Whatever the comparison says.
func Classify(fixture Fixture, result Result, err error) FixtureVerdict {
	rules, complete := fixture.Rules()
	verdict := FixtureVerdict{
		Name:     fixture.Name,
		Rules:    rules,
		Expected: fixture.Expected.Errors,
		Reported: result.Errors,
	}

	if fixture.RequiresFlow() {
		verdict.Verdict = VerdictFlowSyntax
		verdict.Reason = "requires a Flow parser (`@flow` in source, upstream's own test)"
		return verdict
	}

	// A stated divergence outranks everything below it EXCEPT the question of whether cohere ships a
	// rule that could have been asked at all. Claiming "we deliberately differ" about a fixture no
	// registered rule reaches would be the score taking credit for a judgment nothing acts on, which
	// is the same error `VerdictNoRuleShipped`'s comment warns about from the other side.
	//
	// So the check is split. Divergences for shipped rules are answered here, before attribution;
	// divergences for unshipped ones are answered after the shipped test below, and until the rule
	// registers they land in `no-rule-shipped` with everything else that rule owns.
	if divergence, found := statedDivergences[fixture.Name]; found &&
		divergenceRuleIsShipped(rules) {
		verdict.Verdict = VerdictStatedDivergence
		verdict.Reason = divergence.Reason + " (held by " + divergence.Boundary + ")"
		return verdict
	}

	if !complete {
		// An unattributed diagnostic means the rule that owns it is the unknown, so scoring it
		// against any rule would credit or blame whichever happened to be running. This is
		// unreachable while TestEveryDiagnosticIsAttributed passes, and is kept as the honest
		// answer if it ever stops passing.
		verdict.Verdict = VerdictUnreachableFixture
		verdict.Reason = "not every expected diagnostic could be attributed to an upstream rule"
		return verdict
	}

	shipped := make([]string, 0, len(rules))
	for _, ruleName := range rules {
		if _, found := ShippedRules[ruleName]; found {
			shipped = append(shipped, ruleName)
		}
	}

	if len(shipped) == 0 {
		verdict.Verdict = VerdictNoRuleShipped
		verdict.Reason = "cohere ships no rule for " + strings.Join(rules, ", ")
		return verdict
	}

	if divergence, found := statedDivergences[fixture.Name]; found {
		verdict.Verdict = VerdictStatedDivergence
		verdict.Reason = divergence.Reason + " (held by " + divergence.Boundary + ")"
		return verdict
	}

	if err != nil {
		var unsupported *ErrUnsupported
		if asUnsupported(err, &unsupported) {
			// A decline from the implementation on a fixture a shipped rule owns is a statement
			// about the run, not about the fixture. Naming it `unreachable` would assert the
			// corpus cannot ask this question, which is the opposite of true: these are the only
			// fixtures it can ask.
			verdict.Verdict = VerdictNotScored
			verdict.Reason = unsupported.Reason
			return verdict
		}
		verdict.Verdict = VerdictFailed
		verdict.Reason = "analysis error: " + err.Error()
		return verdict
	}

	typeAware := false
	for _, ruleName := range shipped {
		if TypeAwareRules[ruleName] {
			typeAware = true
		}
	}
	if typeAware && len(result.Errors) == 0 && !fixture.ResolvesReactTypes() && fixture.CallsHook() {
		verdict.Verdict = VerdictUnresolvableTypes
		verdict.Reason = fmt.Sprintf(
			"%s reads types, and this fixture calls a hook without importing react, so every hook resolves to `any`",
			strings.Join(shipped, ", "))
		return verdict
	}

	comparison := Compare(fixture.Expected, result)
	if comparison.Outcome == OutcomePassed {
		verdict.Verdict = VerdictPassed
		return verdict
	}
	verdict.Verdict = VerdictFailed
	verdict.Reason = comparison.Detail
	return verdict
}

// Aggregate runs an implementation over the corpus and produces the categorised report.
func Aggregate(fixtures []Fixture, implementation Implementation) (Report, error) {
	for _, fixture := range fixtures {
		if err := CheckPragmas(fixture.Name, fixture.Pragmas); err != nil {
			return Report{}, err
		}
	}

	report := Report{Considered: len(fixtures), Counts: map[Verdict]int{}}
	for _, fixture := range fixtures {
		var result Result
		var err error
		if !fixture.RequiresFlow() {
			result, err = implementation.Analyze(fixture)
		}
		verdict := Classify(fixture, result, err)
		report.Counts[verdict.Verdict]++
		report.Fixtures = append(report.Fixtures, verdict)
	}

	sort.Slice(report.Fixtures, func(first, second int) bool {
		return report.Fixtures[first].Name < report.Fixtures[second].Name
	})
	if err := report.Check(); err != nil {
		return Report{}, err
	}
	return report, nil
}
