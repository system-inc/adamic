# Delegated predicate proof results

This unit serves roadmap step 09, #tzd3gjg. It builds on compiler/step09-predicates-ahead 223f233a. No other worker branch is merged. The source remains the 8a7ab17e adapted tree: 79 identical source hashes, 320 preserved checker diagnostics. The 115 delegation bodies and all 105 mixed-kind bodies were measured before and after, including their 1,134 resolved calls. These are production proof/admission probes, not an executable build of the TypeScript compiler.

## Proof and admission

Only independently verified helper bodies supply summaries. Reachable direct named boolean helpers are verified alongside annotated predicates; both truth directions start empty and converge from independently proved seeds. Scalar literal sets, multi-literal kind fields, kind switches with fallthrough and ordinary breaks, boolean result aliases, conditionals, conjunction, disjunction and negation preserve the tested argument identity. A scalar helper may observe that argument’s unchanged kind snapshot. Unknown calls and writes refuse the path, conservatively invalidating every earlier fact. Effectful or unrelated helper arguments, defaults, rest parameters, generic or bodyless helpers, arbitrary getters, and unseeded recursion supply no proof. Supported named function bindings cannot be reassigned under the checker.

Tag summaries establish the partition only. Every complete target keeps its checked field views. Extra conditions that reject a valid target do not prove the false direction; their checks remain. Pending views are never labelled admitted.

| Group | Logical body proofs before → after | Fully admitted before → after | .ts checked predicates (calls) before → after | .a refused before → after | .a pending views before → after | .ts pending views before → after |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| delegation or composition of predicate calls (115) | 13 → 24 | 1 → 8 | 10 (13) → 3 (5) | 102 → 91 | 12 → 16 | 78 → 78 |
| kind partitions with control flow or extra conditions (105) | 3 → 18 | 3 → 3 | 0 (0) → 0 (0) | 102 → 87 | 0 → 15 | 87 → 88 |

The delegation .ts partition is 8 proven, 3 checked, 78 pending views, 3 pending other contracts and 23 without a resolved corpus call. The mixed-kind .ts partition is 3 proven, 88 pending views and 14 without a resolved corpus call. One previously uncalled mixed-kind body now proves its tags but exposes a complete-target view stop, explaining its change from unobserved to pending. The .a and .ts columns overlap by source mode and must not be summed. Eight call checks retire, all belonging to seven independently admitted operator-helper predicates.

## Remaining delegation obligations

| Reason for an unproved body | Bodies |
| --- | ---: |
| a delegated helper tests a different argument identity | 4 |
| a helper must test the same argument through a direct call | 2 |
| false return does not exclude the target | 2 |
| opaque test or possible mutation on this return path | 11 |
| the target has an unsupported runtime contract | 37 |
| true return does not establish the target | 1 |
| unconverged or opaque helper on this return path | 34 |

All 115 before/after positions, .a diagnostics, .ts statuses and remaining checked call positions are in [delegation-sites.md](delegation-sites.md). All 220 exact body and call results, including reasons for the 87 remaining mixed-kind refusals, are in evidence/delegation-results.json.gz.

## Remaining checked delegation sites

| Predicate annotation | Checked call positions |
| --- | --- |
| `src/compiler/factory/utilities.ts:1256:58` (isAssignmentOperatorOrHigher) | `src/compiler/factory/utilities.ts:1263:12` |
| `src/compiler/factory/utilities.ts:1262:46` (isBinaryOperator) | `src/compiler/factory/utilities.ts:1268:12` |
| `src/compiler/utilities.ts:5276:60` (isKeywordOrPunctuation) | `src/compiler/parser.ts:2493:22`; `src/compiler/parser.ts:2549:22`; `src/compiler/parser.ts:5775:30` |

## Q5 optional node.original

All exact 138 cases still stop at utilitiesPublic.ts:768:16 with “a checked field alias requiring an optional, accessor, or representation conversion”. They comprise 133 direct kind bodies and five delegation bodies. Their tag proofs remain valid; no complete target is admitted.

V2 is available on main: the non-array read dispatcher lands as 4556d540, with nullable-reference compatibility 4df3a0b7 and the explicitly retained optional read boundary f3190f58. The published V2 dependency is compiler/views-v2-on-c2 6d0ba0a9f06cf4c4dd691e05c1f1de12c18addfc; the inspected main tip is 1ac93ec1. V2’s own dispatch report explicitly says an unused optional view can be admitted lazily while reading its optional alias remains refused. The adapter is not registered on this older 223f233a base.

These 138 no longer wait for V2 to land. They need an optional Node | undefined object-slot read conversion that preserves the per-slot presence bit, distinguishes missing from present undefined, validates present object payloads and retains the complete child view through the saved alias. Q5’s compiler/optional-presence a774d316 dependency and this branch’s 5661342e conversion provide presence and scalar/string conversions, not that object conversion. There is no pushed implementation SHA for the remaining optional-object adapter found in the inspected tips. codex/views-optional-host 3716d57c handles optional host methods, not these object slots. No unlanded views branch was merged or credited as passing.

## Witnesses and mutants

Ten .a witnesses are stored under internal/lower/testdata/predicates_delegated. Four composition witnesses pass as .a and temporary .ts copies against source Node, sanitized native with ASan/UBSan, release native, JavaScript and leak checks. Their default-false helper return mutants compile and exit zero with empty stderr, but disagree with Node output in both backends. The literal witness follows factory/utilities.ts’s operator-helper chain; the switch witness follows utilitiesPublic.ts’s boolean kind helpers without deleting the target’s fields.

Five independent lying witnesses cover an annotated helper whose body lies, an inverted delegation, a finite execution of a helper cycle with no independent proof seed, a mutation after saving a helper result and a different tested argument. All stay refused in .a; .ts exits 70 naming the predicate, call site, true branch and both types. Dropping their checks reproduces Node’s unvalidated output in both backends and is caught. The separate mixed_false witness retains its false-direction check; dropping it changes the named diagnostic, while a downstream union narrowing still stops. It is not claimed to reproduce Node or to be an independent semantic false-guard kill.

Four production proof mutants run independently through scratch Go overlays: trusting a lying helper annotation, seeding a recursive cycle from its annotation, failing to negate the truth direction and ignoring the intervening change call. TestDelegatedProofLyingHelper, TestDelegatedProofSeedlessCycle, TestDelegatedProofNegation and TestDelegatedProofEffect respectively catch them. Direct declaration probes prevent an earlier helper refusal from masking a caller’s unsound proof. The finite cycle witness is the executable Node stand-in; an infinitely recursive pure cycle is tested only for proof refusal, never run.

Four measurement mutants are caught: dropping a body, dropping a call, inventing an admission and altering a pinned source hash. A proposed helper-binding mutation was discarded because the checker rejected its witness with TS2630 before the proof ran. That was not a valid catcher; the redundant new guard and that test were removed, and delivery code was remeasured.

## Verification

Every command writes directly to a log. No whole package or full gate runs. Final focused commands and per-test seconds are below; all new or touched top-level leaves stay below 60 seconds on the four-CPU cgroup.

- `go test ./internal/lower -run '^TestDelegatedProof' -count=1 -parallel=1 -v`: PASS, 0.339s.
- `go test ./internal/lower -run '^TestDelegatedProof|^TestPredicate|^TestConditionAssertionProofMutants$|^TestConditionAssertionAdmission$' -count=1 -parallel=1 -v`: PASS, 8.136s.
- `python3 review/compiler/step09-predicates-ahead/delegation/run-delegation-mutants.py`: four independent proof mutants caught.
- `go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -args -update-counts`: PASS, 45.631s.
- Generate measure-overlay.py and build measure-probe with its hatch_predicate_measurement tag. Run HATCH_SELECTION for the selected 220 bodies; separately run HATCH_BODY_ONLY=1 for the exact 138 Q5 bodies. audit-delegation.py verifies both snapshots; run-delegation-audit-mutants.py catches all four corruptions.

The refreshed counts add ten own allocation rows and ten own predicate-direction rows. The four successful .a rows balance allocations/frees. Six temporary .ts rows record their checked stops and call directions; process-stop rows can retain allocations because exit 70 is terminal. Every previous row remains unchanged. No .ts source file is added. No pending test is added.

Integration’s required lane checks run after the implementation commit and before the single push. This base predates cmd/adamic-gate/parallel_test.go, so its unchanged analyzer from origin/main was also run in a scratch tree containing the three test files added or touched by this unit: 20 top-level tests including the analyzer itself, zero failures, empty serial baseline. The repository-root lane command and its exact output are reported after the commit.

## Setup and remaining design questions

GOPROXY=https://proxy.golang.org|direct; source /workspace/adamic-tools/env.sh. Setup: Node 0.022s, Go 0.021s, markdown ready 0.073s, submodules 0.078s, clang 0.185s, build 38.521s, test binaries deferred 38.627s, warm 38.629s, done 38.662s. nproc=5, cpu.max=400000 100000. Node 24.19.0, Go 1.27.1, clang 20.1.8.

Q2–Q4 and Q6 remain recorded in corpus-report.md: complete negative view membership without consuming a view; indirect/method/generic/callback calls; effects and getter evaluation; complete producer domains and overlapping optional targets. No new ruling is requested for this conservative delegation rule. Optional object aliases still need the conversion described above. Effects currently refuse the whole surviving proof path rather than resume from a fresh fact afterward. Unrelated fields, flags and semantic subset tests are not certified by helper annotations. The remaining 360 corpus bodies outside these two groups are not remeasured by this unit.

## Final test leaf measurements

Final oracle and regression selector: `go test ./internal/oracle -run '^TestDelegatedPredicate|^TestPredicateDirectionCountsAreRecorded$|^TestCheckedPredicateOracle$|^TestPredicateKindProofOracle$|^TestConditionAssertionsOracle$|^TestConditionAssertionReevaluationMutant$' -count=1 -parallel=1 -v`: PASS, 20.851s. `go vet ./internal/lower ./internal/oracle`: PASS.

| Added or touched top-level test | Seconds |
| --- | ---: |
| TestDelegatedProofLyingHelper | 0.03 |
| TestDelegatedProofSeedlessCycle | 0.03 |
| TestDelegatedProofExtraCondition | 0.03 |
| TestDelegatedProofPositiveComposition | 0.15 |
| TestDelegatedProofEffect | 0.03 |
| TestDelegatedProofIdentity | 0.03 |
| TestDelegatedProofNegation | 0.03 |
| TestPredicateDirectionCountsAreRecorded | 0.88 |
| TestDelegatedPredicateMixedFalseOracle | 0.56 |
| TestDelegatedPredicateCountsAreRecorded | 1.03 |
| TestDelegatedPredicateIdentityOracle | 0.49 |
| TestDelegatedPredicateEffectOracle | 0.53 |
| TestDelegatedPredicateCycleOracle | 0.51 |
| TestDelegatedPredicateNegationOracle | 0.52 |
| TestDelegatedPredicateLyingHelperOracle | 0.50 |
| TestDelegatedPredicatePathsOracle | 0.66 |
| TestDelegatedPredicateSwitchOracle | 0.73 |
| TestDelegatedPredicateLiteralOracle | 0.69 |
| TestDelegatedPredicatePrimitiveOracle | 0.64 |

## Committed integration lane checks

Implementation commit: `28f26bb3d`. The first check named an inherited gofmt error in `internal/javascript/view_test_runtime_test.go`. Commit `1cf272d55` removes its extra blank line; this helper-only file declares no Test functions. The required repository-root command then passed:

```text
lane checks 8.4 s: gofmt and tools on 92 Go files, t.Parallel on 9 test packages; no t.Parallel analyzer on this tree; vet 9 packages
```

Command: `git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -`. The missing analyzer notice is retained verbatim; the independent unchanged integration analyzer run described above verifies all added or touched test leaves. No main or area branch was pushed or merged.
