package high_level_intermediate_representation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	shimchecker "github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/rules/react/conformance"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// TestPreserveManualMemoizationAgainstGoldens scores the rule against upstream's own answers.
//
// This is the first independent oracle in phase 6. Every other stage was verified by conservation
// properties, mutation sweeps and cross-checks of my own design; this one has an answer key written
// by the people who wrote the rule.
//
// # What it asserts, and what it deliberately does not
//
// It does NOT assert that all 33 pass. Scoring a fresh port against a golden set and demanding a
// perfect number is how a rule ends up shaped like its fixtures: the failures get patched
// individually until the count is right, and the result matches the corpus rather than the
// specification.
//
// What it asserts is that the rule RUNS on the corpus and produces a non-degenerate answer -- it
// fires on some fixtures and not others -- plus the current pass count as a floor, so the number can
// only be driven up. The count is logged rather than pinned exactly, and the floor moves when
// someone improves it.
func TestPreserveManualMemoizationAgainstGoldens(t *testing.T) {
	t.Parallel()

	fixtures, err := react_conformance.Load("../../rules/react/conformance/testdata/fixtures")
	if err != nil {
		t.Fatalf("loading the vendored corpus: %v", err)
	}

	// The fixtures whose golden carries this rule's message. Everything else is another rule's
	// business and scoring against it would measure the wrong thing.
	const ruleMessage = "Existing memoization could not be preserved"
	var mine []react_conformance.Fixture
	var flowOnly int
	for _, fixture := range fixtures {
		if !strings.Contains(fixture.Expected.Raw, ruleMessage) {
			continue
		}
		// Flow's `component Component(id) { ... }` declaration form is EXCLUDED, deliberately,
		// because the vendored TypeScript parser does not model it and no work in this package can
		// recover it.
		//
		// # What the parser actually produces, measured rather than assumed
		//
		// There is no `KindComponentDeclaration` anywhere in `typescript-go`, and the parser does
		// not reject the syntax either -- it RECOVERS, into three unrelated top-level nodes:
		//
		//	component            ->  ExpressionStatement(Identifier)
		//	Component(id)        ->  ExpressionStatement(CallExpression)
		//	{ ... }              ->  Block                       (a detached block, not a body)
		//
		// So `ForEachFunctionLike` finds no function for the component. It descends into that
		// detached `Block` and reaches the `useMemo`/`useCallback` CALLBACKS, which are ordinary
		// arrow functions, and lowers each one as though it were itself a top-level component. That
		// is why these two fixtures report `lowered=true` with reactive scopes and yet record zero
		// written dependencies: the `useMemo` CALL SITES live in the detached block, which is never
		// lowered, so `DropManualMemoization` is never handed a call to recognise and no
		// `StartMemoize` marker is ever constructed. The identical program written as
		// `function Component(id)` lowers to one function with two markers carrying their deps,
		// which is the control.
		//
		// # Why this is not repairable downstream
		//
		// Lowering is keyed on `*ast.Symbol` identity throughout -- `builder.declarations`,
		// `.identifiers` and `.captured` are all symbol-keyed. The binder never created a parameter
		// binding for `id`, because there is no function to bind one into, so
		// `GetSymbolAtLocation(id)` returns nil at its declaration site AND at its use inside the
		// dependency array, while resolving at an unrelated use. A dependency-array entry with no
		// symbol cannot be built into a `ManualMemoDependency` and cannot be compared against an
		// inferred dependency, so synthesising a function node here would produce a marker whose
		// operands name nothing. The gap is in the parser, above `Lower`, and closing it means
		// teaching `typescript-go` a Flow production.
		//
		// # Why they are counted as unsupported rather than filtered out
		//
		// Dropping them would move the denominator to 31 and hide the exclusion inside a number.
		// They stay in `len(mine)` and land in `unsupported`, which is the bucket this test already
		// uses for a program it declines to judge, so the 33 keeps matching the corpus and the cost
		// of the exclusion stays visible on the log line.
		//
		// Both are named `todo-repro` upstream, and the second's own comments describe a capture
		// chain widening a mutable range through an invoked function -- work `#t28gsec` addresses --
		// so it would likely stay silent even with a parser that handled the declaration.
		//
		// `RequiresFlow` is upstream's own test (the substring `@flow`), and on THIS population it
		// is exact: it selects these two fixtures and no others.
		if fixture.RequiresFlow() {
			flowOnly++
		}
		mine = append(mine, fixture)
	}
	if len(mine) == 0 {
		t.Fatal("no fixture in the vendored corpus carries this rule's message; the corpus or the " +
			"message string is wrong, and a zero score below would be meaningless")
	}

	fired, silent, unsupported, disowned := 0, 0, 0, 0
	divergences := react_conformance.StatedDivergenceNames()
	for _, fixture := range mine {
		// The Flow-declaration exclusion documented above. This is checked BEFORE lowering rather
		// than after, because these fixtures do lower -- into the wrong thing -- and so the `!ok`
		// decline below cannot see them.
		if fixture.RequiresFlow() {
			unsupported++
			continue
		}
		// Fixtures upstream's own headers call mistakes. Staying silent on one is the right answer,
		// so counting it as a target would make this number describe the corpus rather than the
		// rule. The three are enumerated in `statedDivergences` with the header text quoted, and
		// `TestPreserveManualMemoizationAgainstGoldens` is named there as the boundary that holds
		// them -- which is this test, so an entry that stops being true fails here.
		if divergence, disowns := divergences[fixture.Name]; disowns &&
			divergence.Boundary == "TestPreserveManualMemoizationAgainstGoldens" {
			disowned++
			continue
		}
		findings, ok := findingsForSource(t, fixture.Source)
		if !ok {
			unsupported++
			continue
		}
		if len(findings) > 0 {
			fired++
		} else {
			silent++
		}
	}

	// The denominator, spelled out rather than left to be reconstructed. `fixtures` is every golden
	// carrying this rule's message; `scoreable` is what remains once the two populations that
	// cannot be scored are removed, and it is the number the pass count should be read against.
	// Printing only `fixtures=33 fired=28` reads as five failures, and three of those five are the
	// rule being right.
	scoreable := len(mine) - unsupported - disowned
	t.Logf("fixtures=%d scoreable=%d fired=%d silent=%d "+
		"(excluded: %d Flow, %d disowned upstream)",
		len(mine), scoreable, fired, silent, unsupported, disowned)

	// Pinned to the exact population, for the same reason as the Flow count below: an exclusion
	// that silently widens is how a denominator rots. A fourth entry naming this boundary, or one
	// of these three being rewritten upstream, has to be re-justified rather than inherited.
	const knownDisownedFixtures = 3
	if disowned != knownDisownedFixtures {
		t.Errorf("fixtures upstream disowns = %d, want %d; the exclusions in `statedDivergences` "+
			"naming this test as their boundary have changed and the judgment has to be re-made",
			disowned, knownDisownedFixtures)
	}

	// The exclusion is pinned to the exact population it was measured against. If the corpus is
	// re-vendored and a THIRD Flow fixture appears -- or one of these two is rewritten in the
	// ordinary `function` form -- this fires, and whoever sees it has to re-make the judgment above
	// rather than inherit it. An exclusion that silently widens is how a denominator rots.
	const knownFlowDeclarationFixtures = 2
	if flowOnly != knownFlowDeclarationFixtures {
		t.Errorf("Flow-declaration fixtures = %d, want %d; the exclusion documented above was "+
			"measured against a population that has changed, and it has to be re-justified rather "+
			"than widened", flowOnly, knownFlowDeclarationFixtures)
	}

	if fired == 0 {
		t.Errorf("the rule fired on none of %d fixtures whose golden expects it; every one of "+
			"these is a program where upstream reports a lost memoization", len(mine))
	}

	// The count is pinned, and `fired > 0` above is not enough on its own.
	//
	// # This assertion exists because its absence nearly shipped a regression as an improvement
	//
	// The clean-fixture rate below is pinned exactly and this side was not, so the two instruments
	// were asymmetric in the one direction that matters: a change trading true positives away for
	// false-positive removals moved the pinned number DOWN, which reads as the rule improving, while
	// this side stayed green because it only asked for a non-zero.
	//
	// That is not hypothetical. Freezing component parameters in the aliasing graph took the clean
	// rate from 31 to 19 and this number from 15 to 9 in the same run. Reported through the
	// instruments as they stood, that is "false positives down 39%" with no signal at all that six
	// programs upstream reports on had gone silent.
	//
	// The six were `error.useMemo-aliased-var` plus five optional-member-expression fixtures, and
	// what they show is that the repair was in the wrong pass. Freezing a parameter suppressed the
	// walk that carries a mutation ONWARD, so `x.push(props?.items)` stopped widening `x`, which is
	// mutated, rather than `props`, which is not. Upstream does not do this: oxc's `NodeValue` in
	// `infer_mutation_aliasing_ranges.rs:72` is `Object | Phi` with no third element, and the
	// parameter freeze lives in `InferMutationAliasingEffects` as a `MutateFrozen` EFFECT rather
	// than as a node value this pass reads.
	//
	// So both directions are pinned, and a change moving either one has to say which it moved and
	// why. Silence is not the right answer on ANY of these 33: each is a program where upstream
	// reports a lost memoization.
	// Raised from 15 by turning on the third firing condition, which had been gated since it was
	// written. The gate's own unblock instruction was "turn this on when the hoistable analysis
	// lands", and it did: `CollectHoistablePropertyLoads` populates the tree, the collector descends
	// into nested functions and into callbacks it can assume are invoked, and 595 dependencies carry
	// a path against 2,092 that do not -- where the gate was written, 171 did against 1,435.
	//
	// The four recovered are the fixtures whose golden says the inferred dependency did not match
	// the written one, which is the only condition that can report them.
	// Raised from 19 by two changes in `hoistable.go` that are inert apart and worth six together:
	// the developer's own dependency array stops seeding the non-null set, and an invoked callback
	// contributes what its entry block proves rather than what any of its blocks prove.
	//
	// The first is a pipeline difference. Upstream drops manual memoization at `Pipeline.ts:168`,
	// eliminates dead code at 230 and analyses dependencies at 428, and its elimination keeps the
	// memo markers while pruning `PropertyLoad` and `ArrayExpression`. So the loads that built the
	// array are gone before anything reads them. This tree keeps them, and reading them treats the
	// developer's declared `propB.x.y` as proof that `propB.x` is non-null -- then infers a deeper
	// dependency than they wrote and reports the disagreement it just manufactured.
	//
	// The second is `CollectHoistablePropertyLoads.ts:449`, which takes one block's answer from the
	// nested analysis. That block's set is post-propagation, and propagation intersects across
	// neighbours, so it holds exactly what the callback reads on every path. It is the only thing
	// that separates `useMemo-conditional-access-noAlloc.ts`, whose body reads `propB?.x.y`
	// unconditionally and where upstream keeps the deep path, from
	// `useMemo-infer-less-specific-conditional-access.ts`, whose body reads it under an `if` and
	// where upstream infers bare `propB`. Their outer functions are nearly identical; only the
	// callback's control flow tells them apart.
	//
	// Measured apart: the array exclusion alone is 20, the entry-block read alone is 19, together
	// 25. The `dependencies.go` block gate in the same commit reads flat on its own and drops this
	// to 20 when removed from the combined state, so it is latent rather than inert -- it only pays
	// once the hoistable set stops over-approximating.
	// Lowered from 25 to 24 by the parameter exemption recorded at `knownFalsePositives`, and the
	// cost is one fixture, named because lowering this number is otherwise a retreat:
	// `error.useMemo-aliased-var.ts`.
	//
	//	const aliasedX = x; const aliasedProp = x.y.z;
	//	useMemo(() => [x, x.y.z], [aliasedX, aliasedProp])
	//
	// Its own header reads "This is technically a false positive, but source is already breaking
	// `exhaustive-deps` lint rule (and can be considered invalid)" -- upstream reports it while
	// saying it should not have to, the same category as the `todo-repro` fixtures elsewhere in this
	// corpus. Ten false positives for that one.
	// Raised from 24 by defining a context binding once in SSA rather than versioning every write.
	// Both fixtures this rule's scope conditions were missing now fire:
	//
	//	new-mutability/error.invalid-useCallback-captures-reassigned-context.js
	//	preserve-memo-validation/error.invalid-useCallback-captures-reassigned-context.ts
	//
	// See the note at `defineIn` in `ssa.go`. Upstream's two writes to such a binding name one
	// identifier, which is what puts the declaration and the reassignment in one disjoint class so
	// the scope spans the memo block between them.
	//
	// Costs two fixtures, both of which upstream itself flags:
	// `error.todo-repro-unmemoized-callback-captured-in-context-variable.tsx` carries the
	// `todo-repro` prefix, and `error.useMemo-aliased-var.ts` says in its own header "This is
	// technically a false positive, but source is already breaking `exhaustive-deps`". Two real
	// detections for two upstream calls a mistake.
	//
	// 27 once a ref's stability is decided the way upstream decides it, in `reactive.go`. The two
	// that close are `error.ref-like-name-not-a-ref.js` and `error.ref-like-name-not-Ref.js`, and
	// they are a matched pair with two clean fixtures that differ only in the variable's name.
	// False positives are unmoved at 11, `under` holds at 0 fixtures / 0 scopes, and the dependency
	// oracle is byte-identical at 74 matched with ours 119.
	//
	// 28 once a ref's STABILITY is seeded only by a hook call, separately from its shape. The one
	// that closes is `error.preserve-use-memo-ref-missing-reactive.ts`, whose `ref` is a phi of two
	// `useRef` results and is named `ref`. False positives hold at 11, `under` at 0 fixtures / 0
	// scopes, and the dependency oracle is byte-identical at 74 matched with ours 119. Memo blocks
	// improve: 20 all-empty to 17 and 7 contradicted to 5.
	//
	// 29 once an alias into a named binding stops resolving through, in
	// `drop_manual_memoization.go`. The one that closes is
	// `useCallback-alias-property-load-dep.ts`, where `const x = propB.x.y` is written as
	// `[propA.x, x]`: the written `x` was resolving to `propB.x.y` while the inferred side kept
	// `x`, so the two could never match. False positives hold at 11, `under` at 0 fixtures / 0
	// scopes, and both other oracles are byte-identical.
	//
	// 28 with the local zero-argument callee resolution in `effects.go`, and the one that leaves is
	// `error.false-positive-useMemo-dropped-infer-always-invalidating.ts`, whose own header says
	// what it is: "This is technically a false positive as the useMemo in source was effectively a
	// no-op". Upstream reports it and calls the report a mistake, which is the same reasoning
	// already recorded above for the two `todo` fixtures.
	//
	// Taken because the trade is measured and one-sided everywhere else: false positives 11 to 8
	// with three real fixtures closing and none added, dependency `matched` 74 to 75 -- a golden row
	// GAINED on the strongest oracle -- and `under` held at 0 fixtures / 0 scopes.
	const knownTruePositives = 28
	if fired != knownTruePositives {
		t.Errorf("true positives = %d, want %d; if this went UP the rule improved and this number "+
			"should be raised deliberately, and if it went DOWN the rule stopped reporting programs "+
			"upstream reports on, which the clean-fixture rate cannot see", fired, knownTruePositives)
	}
	// Non-degeneracy in the other direction is not asserted: a rule that fires on all 33 could be
	// correct, since all 33 are error fixtures. Over-firing is caught by the clean-fixture check
	// below instead, where silence is the right answer.
}

// TestPreserveManualMemoizationFalsePositiveRate is the over-reporting half, measured directly.
//
// The 33 error fixtures can only show under-reporting: a rule that fires on everything passes all of
// them. This is the population where silence is the right answer, so a finding here is a defect and
// nothing else in the phase can see one.
//
// # It took three attempts to define this population, and the first two are why the number is pinned
//
// The first spelling took every fixture whose `Expected.Raw` lacks this rule's message and which
// mentions `useMemo`. That found 30 fixtures and 11 firings, which read as an 11-false-positive rate.
// The filter was wrong: `Expectation` parses only the `## Error` block, so `Raw` is empty for every
// fixture expecting clean output, and the complement was made almost entirely of error fixtures for
// other rules. Firing on those is not obviously wrong.
//
// The second concluded the corpus held zero fixtures that use manual memoization and expect no error
// at all, and therefore that this rule's over-reporting could not be measured here. That was true of
// the corpus and false about the cause: `Load` rejected every fixture not named `error.*`, so the
// tree was error-only by construction. Upstream ships 70 of these. The loader excluded them, and the
// exclusion was reasoned about for hours as a property of upstream.
//
// So the rate is now a real measurement rather than an upper bound, and it is high: 31 of 70. The
// count is asserted exactly rather than as a ceiling, because a ceiling cannot distinguish a fix
// from a rule that went quiet, and this rule's registration is being held precisely on the strength
// of this number.
func TestPreserveManualMemoizationFalsePositiveRate(t *testing.T) {
	t.Parallel()

	fixtures, err := react_conformance.Load("../../rules/react/conformance/testdata/fixtures")
	if err != nil {
		t.Fatalf("loading the vendored corpus: %v", err)
	}

	const ruleMessage = "Existing memoization could not be preserved"
	clean, fired, silent, unsupported, logsOnly := 0, 0, 0, 0, 0
	for _, fixture := range fixtures {
		base := filepath.Base(fixture.Name)
		if strings.HasPrefix(base, "error.") || strings.HasPrefix(base, "todo.error.") {
			continue
		}
		if upstreamReportsInLogs(t, fixture.ExpectPath, ruleMessage) {
			// Upstream reports this rule on the fixture, in its `## Logs` block rather than in a
			// `## Error` section. The population above is chosen by FILENAME, and `Expectation`
			// parses only `## Error`, so neither route can see it: the fixture reads as clean while
			// upstream reports on it, and firing here was scored as a false positive.
			//
			// Measured: exactly one fixture corpus-wide,
			// `gating/dynamic-gating-bailout-nopanic.js`, whose source is
			// `useMemo(() => identity(value), [])` -- an empty dependency array over a callback that
			// reads `value`, which is a real violation both compilers agree on. Its log line carries
			// `"kind":"CompileError"` and this rule's message.
			//
			// Excluded here rather than in the loader because `ExpectedCleanFixtureCount` is pinned
			// against the filename split and is read by other tests.
			logsOnly++
			continue
		}
		clean++
		findings, ok := findingsForSource(t, fixture.Source)
		if !ok {
			unsupported++
			continue
		}
		if len(findings) > 0 {
			fired++
			continue
		}
		silent++
	}

	const knownLogsOnly = 1
	if logsOnly != knownLogsOnly {
		t.Errorf("%d fixtures report this rule only in `## Logs`, want %d; that population is "+
			"invisible to both the filename split and to `Expectation`, so a change in it moves "+
			"the false-positive count for a reason unrelated to the rule", logsOnly, knownLogsOnly)
	}
	if clean+logsOnly != react_conformance.ExpectedCleanFixtureCount {
		t.Fatalf("scored %d clean fixtures, want %d; the population this test measures is not the "+
			"one it believes it is", clean, react_conformance.ExpectedCleanFixtureCount)
	}
	// A declined fixture is not a correct silence. In particular, pin this at zero now that the
	// false-positive count is zero: otherwise a lowering or pipeline regression could move a clean
	// fixture from `silent` to `unsupported` while leaving the headline score green.
	if unsupported != 0 {
		t.Errorf("unsupported clean fixtures = %d, want 0; every clean fixture in the denominator "+
			"must reach validation before zero false positives is a meaningful result", unsupported)
	}
	t.Logf("cleanFixtures=%d silentCorrect=%d falsePositives=%d unsupported=%d (%.0f%% false-positive rate)",
		clean, silent, fired, unsupported, 100*float64(fired)/float64(clean))

	// Every one of these is the rule reporting a lost memoization on a program upstream compiles
	// clean. A gate that does this gets switched off, which is why the registration is held.
	// Lowered from 31 by `PruneNonEscapingScopes` marking `FinishMemoize` markers whose value was
	// never memoized, which upstream does in the same pass and this tree did not do at all.
	//
	// The move is smaller than the measured ceiling of 23 and that is the informative part rather
	// than a shortfall. Forcing every marker to read as pruned reaches 23, so 7 of those 8 belong to
	// fixtures that ALSO fire the `StartMemoize` condition and stay in the count on that other
	// finding. The marker pass itself prunes 36 of 68 markers on this corpus and leaves zero
	// unpruned markers whose value carries no scope, so it is not under-firing.
	// Lowered from 30 by declaring the four `Object` statics in the effect table. Three fixtures,
	// exactly the three whose subject they are: `object-keys`, `object-values` and
	// `repro-object-fromEntries-entries`, all now silent.
	// The DENOMINATOR moved from 70 to 69 without this number moving, which is worth stating
	// because the two are usually read together. `gating/dynamic-gating-bailout-nopanic.js` is now
	// excluded: upstream reports this rule on it, in `## Logs` rather than in a `## Error` section,
	// so it was counted as clean while both compilers agree it is a violation. It was not firing
	// under the current gate, so 27 is unchanged; ungated it was one of the false positives and is
	// no longer counted as one.
	// Raised from 27 by the same change, and the cost is exactly one fixture:
	// `useCallback-alias-property-load-dep.ts`. Traced to a scope boundary rather than to the
	// comparison -- `const x = propB.x.y` is declared inside our scope 1, so `checkValidDependency`
	// rejects `x` on the rule this tree shares with upstream verbatim
	// (`decl.order < scopeRange.Start`), while upstream's scope begins after that declaration and
	// names `x` in its compiled output. That belongs to the scope work.
	//
	// Four true positives for one false positive is the first net-positive reading this trade has
	// ever had. It was plus six for thirteen when the ungating task was re-measured, and every step
	// between is recorded on that task.
	// Raised from 28 by the change described at `knownTruePositives`, and the cost is one fixture
	// family rather than three defects:
	//
	//	useMemo-conditional-access-alloc.ts
	//	useMemo-conditional-access-noAlloc.ts
	//	useMemo-conditional-access-own-scope.ts
	//
	// All three read `propB?.x.y`, and an optional load proves nothing about its object, so
	// `maybeNonNullInInstruction` refuses it on purpose. Upstream reaches the deep path a different
	// way: `collectOptionalChainSidemap` keys its hoistable set by optional block, and that pass is
	// driven by `Optional` terminals this lowering does not produce. The refusal is correct and the
	// route around it is missing, which is the gap already recorded at that arm.
	//
	// `own-scope` is also the single row the dependency oracle loses in this change, `propB.x.y`,
	// so that oracle's fall and one of these three are the same fixture rather than two findings.
	//
	// Six true positives for three false positives, with `under` in the scope oracle unmoved at
	// 5 fixtures / 7 scopes and the corpus depth distribution moving 595 to 574 rather than
	// collapsing the way every truncation attempt on this cluster did.
	// Lowered from 31 by exempting a dependency whose scope the walk never entered. See the note at
	// `check`: 35 of 55 checks corpus-wide are on such a scope, and upstream cannot reach that state
	// because every scope it asks about is in the tree it walks.
	//
	// The six, and they are one shape:
	//
	//	prune-nonescaping-useMemo.ts
	//	prune-nonescaping-useMemo-mult-returns.ts
	//	prune-nonescaping-useMemo-mult-returns-primitive.ts
	//	maybe-invalid-useMemo-no-memoblock-sideeffect.ts
	//	optional-member-expression-inverted-optionals-parallel-paths.js
	//	propagate-scope-deps-hir-fork/optional-member-expression-inverted-optionals-parallel-paths.js
	//
	// The first four are the zero-cache-slot group: upstream emits no `_c(` at all for them, having
	// deleted the bodies, and `prune-nonescaping-useMemo.ts` calls reporting here "technically a
	// false positive" in its own header. Nothing newly fires, and golden holds at 25 with all three
	// oracles byte identical.
	// Lowered from 25 by exempting a component or hook parameter from the dependency condition. See
	// the note at `check`: the parameter's scope is inherited from the callback that captured it,
	// and upstream's own comment calls this an edge case of mutable range inference.
	//
	// The ten, which are the dependency-condition cluster this rule has carried all along:
	//
	//	memoize-primitive-function-calls.js
	//	optional-member-expression-single-with-unconditional.js
	//	propagate-scope-deps-hir-fork/optional-member-expression-single-with-unconditional.js
	//	repro-maybe-invalid-useMemo-read-maybeRef.ts
	//	repro-preserve-memoization-inner-destructured-value-mistaken-as-dependency.js
	//	useMemo-in-other-reactive-block.ts
	//	useMemo-infer-fewer-deps.ts
	//	useMemo-infer-more-specific.ts
	//	useMemo-infer-nonallocating.ts
	//	useMemo-inner-decl.ts
	//
	// Nothing newly fires; every oracle and every structural invariant is unchanged.
	// # Lowered to 13 when capture-free function expressions began being outlined
	//
	// A `FunctionExpression` closing over nothing is the same function on every render, so upstream
	// lifts it to a module-level declaration before scopes are assigned and it never gets a reactive
	// scope. `repro-object-fromEntries-entries.expect.md` shows both of its callbacks compiled to
	// `_temp` and `_temp2` at module scope.
	//
	// The two that stop firing are `allow-global-mutation-in-effect-indirect-usecallback.ts` and
	// `array-pattern-spread-creates-array.ts`, named by logging the fixture per false positive and
	// diffing rather than inferred from the fixture that motivated the pass. Both were the rule
	// validating a memoization against a scope that should never have existed.
	//
	// Goldens hold at 25 and all three oracles are byte-identical: scope survived 108 with exact 30
	// and under-production at 5 fixtures / 7 scopes, dependency 77 matched against 121 produced,
	// memo blocks 101 with 65 carrying.
	//
	// Lowered to 11 by the frozen-capture rule in `ranges.go`. The two that stop firing are
	// `useMemo-constant-prop.ts` and `todo-ensure-constant-prop-decls-get-removed.ts`, named the
	// same way, by logging the fixture per false positive and diffing. Both memoize over a constant,
	// so upstream emits a `memo_cache_sentinel` rather than a dependency comparison, and both were
	// the rule validating against a scope that had fused with its neighbour across a closure
	// capture.
	//
	// Goldens hold at 25 and `under` holds at 5 fixtures / 7 scopes. The other oracles move and each
	// carries its reasoning: dependency 74 matched held with ours 116 to 119, scope survived 109 to
	// 110 with exact 28 held, memo blocks 20 all-empty to 22.
	//
	// Lowered to 8 by the local zero-argument callee resolution in `effects.go`. The three that
	// close are `hoisting-setstate-captured-indirectly-jsx.js`,
	// `useCallback-nonescaping-invoked-callback-escaping-return.ts` and
	// `repro-missing-memoization-lack-of-phi-types.js`, named by logging the fixture per false
	// positive and diffing. None is added.
	//
	// Lowered from 8 to 1 by giving optional chains their upstream control-flow shape and sidemap,
	// preserving that shape through memo-callback inlining, deferring every instruction consumed by
	// the optional traversal, and making hoistability respect the branch that proved a path safe.
	// Goldens rose to and held at 28 while each clean-fixture removal was checked against the same
	// production pipeline rather than against a fixture-specific comparison exemption.
	//
	// Lowered from 1 to 0 by closing the abstract-value chain underneath the last optional fixture.
	// Destructure now carries a component parameter's Frozen kind into each ordinary binding, a phi
	// whose inputs are all non-mutable retains that fact, and Alias/MaybeAlias candidates from Frozen
	// sources refine to range no-ops as upstream's ImmutableCapture does. That leaves the optional
	// argument outside the allocating call's singleton scope and also prevents later mutations of a
	// call result from widening a frozen source. The final board is 28/28 goldens and 69/69 clean.
	const knownFalsePositives = 0
	if fired != knownFalsePositives {
		t.Errorf("false positives = %d, want %d; if this went DOWN the rule improved and this "+
			"number should be lowered deliberately, and if it went UP something regressed",
			fired, knownFalsePositives)
	}

	// Non-degeneracy. A rule silent on everything scores zero false positives and is worthless, and
	// the error-fixture test above is what would catch that -- restated here so this test cannot
	// pass alone by the rule ceasing to report.
	if silent == 0 {
		t.Error("the rule fired on every clean fixture, which is unconditional reporting")
	}
}

// findingsForSource runs the whole pipeline over one fixture and returns what the rule reported.
//
// Returns false when the source does not lower, which is a decline rather than a silent zero -- the
// distinction the conformance harness draws between `Declined` and `Failed`, and for the same
// reason: an empty result is a claim, and a claim about a program that never parsed is a wrong one.
func findingsForSource(t *testing.T, source string) ([]PreserveManualMemoizationFinding, bool) {
	t.Helper()

	var all []PreserveManualMemoizationFinding
	lowered := false

	probe := rule.Rule{
		Name:             "preserve-manual-memoization-score",
		NeedsTypeChecker: true,
		Run: func(ctx rule.Context, options any) rule.Listeners {
			return rule.Listeners{
				ast.KindSourceFile: func(node *ast.Node) {
					if ctx.TypeChecker == nil {
						return
					}
					forEachFunctionLike(node, func(functionNode *ast.Node) {
						function := Lower(functionNode, ctx.TypeChecker)
						if function == nil {
							return
						}
						Construct(function)
						lowered = true
						all = append(all, pipelineFindings(function, ctx.TypeChecker)...)
					})
				},
			}
		},
	}
	rule_testing.RunTypedFiles(t, probe, map[string]string{
		"/react.d.ts":  reactiveDeclarations,
		"/fixture.tsx": source,
	}, "/fixture.tsx")

	return all, lowered
}

// pipelineFindings runs every phase-6 pass in order and then the rule.
//
// The order is upstream's pipeline order, and it is load-bearing: the rule reads `Pruned` and
// `Merged`, both of which are written by passes that must have run.
func pipelineFindings(function *Function, checker *shimchecker.Checker) []PreserveManualMemoizationFinding {
	return AnalyzePreservedManualMemoization(function, checker)
}

// upstreamReportsInLogs reports whether a fixture's expectation carries this rule's message in its
// `## Logs` block rather than in a `## Error` section.
//
// `Expectation` parses only `## Error`, deliberately -- see its own comment. A fixture compiled
// under `@panicThreshold:"none"` still records the diagnostic, but as a `CompileError` log line, and
// nothing else in this package can see that. Reading the file directly is the narrow answer.
//
// The `## Logs` scoping is intent rather than a measured necessity, and saying so is the honest
// version: measured over every clean fixture that mentions this rule's message, all zero of the
// mentions fall outside `## Logs`, so searching the whole file would answer identically today. A
// mutation removing the scoping survives for that reason. It is kept because the question being
// asked is specifically "did upstream RECORD a diagnostic here", and a fixture whose `## Input`
// happened to contain the sentence in a comment would answer yes to a whole-file search and be
// silently dropped from the population.
func upstreamReportsInLogs(t *testing.T, expectPath string, ruleMessage string) bool {
	t.Helper()
	body, err := os.ReadFile(expectPath)
	if err != nil {
		return false
	}
	text := string(body)
	start := strings.Index(text, "## Logs")
	if start < 0 {
		return false
	}
	return strings.Contains(text[start:], ruleMessage)
}
