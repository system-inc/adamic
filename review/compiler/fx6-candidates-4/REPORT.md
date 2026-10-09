Survivor: candidate 3's historical single-filter producer-certificate revert still passes because untaggedCallableABI independently filters direct functions; its adapted two-filter revert is caught. No previously killed mutant now survives.
Built candidate 4 for views task #1fk58py, integrating items 146 and 138 while retaining candidate 3 guards and fixtures, and activating five repaired pending witnesses.
Commits: merge 644b2de3, fixture activation ca88bf56, isolated metadata regression 9006981e; the evidence/delivery commit is the pushed branch tip reported in the delivery response.
Checks: full lowering, selected Node/backend oracles, stage3, reader guard, counts and lane checks pass; 38 intended member/revert obligations caught, including the isolated metadata revert.
Not covered: the full integration gate or full native/oracle packages; the historical equivalent single-filter mutant cannot be claimed killed. Other pending review probes and general callable adapters remain outside this unit.

The exact branch recipe takes precedence over the generic main-start and unlanded-member rules. Started from origin/compiler/fx6-candidates-3 d8ca3f0c and merged origin/compiler/fx7-wrong-aborts 636b7397. Kept this specified ancestry rather than adding later main changes. MERGE.md records the one conflict, both sides, the retained guards and the inspected clean overlaps. No cohere code was copied, no PR was opened, and no protected branch was changed.

All valid fixtures exercised by this unit agree with source Node in JavaScript, release native and sanitized native. The oracle also checks leaks on successful programs. The deliberate lying-input controls continue to reject with their required checks; those are not acceptance fixtures. No production disagreement appeared, so no production fix or production-fix revert commit was necessary.

The five named item-146 pending programs p17, p26, p33, p34 and p35 were explicitly run with their sidecars temporarily removed, restoring the sidecars in finally. All passed, so their stale sidecars were then removed permanently and the existing sources were registered for counts. p18, p20, p68, integration p69 and extended p69 also passed. No source fixture or expected output was weakened.

Observation about the supplemental metadata mutant: removing the new CheckedCast.CheckedFields branch from fieldTypesNeeded passes the original lowered p18 test on this merged stack (checked-cast-metadata.log, exit 0). Candidate 3's read summaries independently enable metadata, masking this edit. TestCheckedCastWithoutReadSummary uses hand-built IR with neither CheckedFields summaries nor Property.View. Its source Node control prints 8. Both release and sanitized native print 8 normally. The same metadata edit now makes the native binary abort at runtime with exit 70 and an unsupported-representation field-read panic; metadata-isolated.log records the kill. This adds a regression test, not a production change. The initial run-extra.py assertion failure is preserved rather than presented as green; run-isolated.py supplies the independent successful proof and the historical masked experiment.

Exact commands are in commands.txt. All test output was written directly to individual logs. Results:

| Check | Observed result |
|---|---|
| Full internal/lower | PASS, 235.209s |
| Selected checked-view, FX6/FX7, scalar/callable and review oracle suite | PASS, 112.842s |
| Explicit named item-146 source/backend oracles | PASS, 14.182s |
| Registered member and wrong-abort fixture sets | PASS, 11.131s |
| Stage3 TestFixturesAssertions from candidate 3 report | PASS, 14.072s |
| Reader guard, first completed run | PASS, 29.747s |
| Reader guard after new test commit | PASS, 1.797s |
| Initial merged counts | PASS, 335.080s, no existing rows moved |
| Counts refresh after five activations | PASS, 105.522s, exactly five new rows |
| Final counts | PASS, 74.156s, no further changes |
| Restored native metadata / lower controls | PASS, native 0.742s, lower 2.433s |
| Four restored constructor oracles | PASS, 1.143s |
| Explicit vet of lower, oracle, JavaScript and native | Exit 0, no diagnostics |
| Final integration lane checks | Exit 0: 18.2s, gofmt/tools on 47 Go files, t.Parallel on 3 test packages, a-check 2 .a files, vet 3 packages |
| git diff --check | Exit 0, no diagnostics |

Initial setup failed because its source census began before the requested checkout and omitted the new member-read file. Its exact missing-method diagnostics remain in setup.log. Stable-tree setup passed using GOPROXY=https://proxy.golang.org|direct and /workspace/adamic-tools/env.sh. Timing lines: Go 0.052s, Node 0.052s, submodules 0.163s, markdown dependencies 0.165s, clang 0.257s, build 128.269s, cache warm 131.325s, total 131.869s. nproc 5; cpu.max 400000 100000. Two early reader commands hit their outer cold-build limits; the successful retries are separate logs. Initial lane vet exceeded its short limit, then explicit vet and the final lane's vet passed. The shared-pool 429 interrupted the turn before delivery; resumption verified both commits, no push and no active mutant, then completed the remaining evidence and checks.

Counts relative to candidate 3: six rows arrive unchanged from wrong-aborts; five rows are newly registered by this unit. All eleven were measured. No pre-existing row changes. A/F/R/L/P/G means allocations/frees/retains/releases/peak/regions.

| Row | A/F/R/L/P/G | Cause |
|---|---|---|
| review native p18 | 4/4/5/9/4/0 | Imported callable checked-read acceptance prints 8 and cleans up. |
| review native p20 | 7/7/6/14/5/0 | Imported checked callable/member fixture executes both successful calls. |
| lower fx7_wrong_aborts/wide.a | 5/5/6/11/5/0 | Imported wider producer fixture boxes its scalar argument for the selected ABI. |
| review views p68 | 4/4/8/14/4/0 | Imported tuple-object view fixture is admitted by JavaScript's tuple marker. Native object representation is unchanged. |
| review views p69 | 3/3/8/13/3/0 | Imported tuple-union equality control now runs in both backends. |
| lower fx7_tuple_recognition/p69.a | 4/4/9/15/4/0 | Imported extended tuple-union fixture also exercises subsequent indexed reads. |
| review native p17 | 11/11/14/19/9/0 | Newly activated escaped callable closures, with two runtime suffix strings. |
| review native p26 | 5/5/8/16/4/0 | Newly activated undefined-to-callable transition checks and call. |
| review native p33 | 4/4/12/18/4/0 | Newly activated Map member of a string/object union, then size read. |
| review native p34 | 4/4/9/15/4/0 | Newly activated Uint8Array union member, then length read. |
| review native p35 | 4/4/10/16/4/0 | Newly activated class-instance union member, then x read. |

The full paths and numeric diff are in internal/oracle/counts.md and the fixture activation commit. The isolated hand-built IR test adds no fixture file or counts row.

Candidate 3 member mutants, all rerun on this stack and caught by the recorded intended diagnostics:

| Mutant | Catcher |
|---|---|
| nullable-receiver | TestCheckedViewOptionalReadBoundary and optional-receiver: optional checked read escaped |
| generic-receiver | TestCheckedViewUntaggedSourceFlows/generic/wrong: exit codes differ |
| union-receiver | Stage3 assertions/19_identifier_kind.a/stage0: gap changed |
| revert-receiver | TestFX6P37: JavaScript backend stdout differs |
| revert-destructure-type-id | TestFX6P53: JavaScript backend stdout differs |
| revert-p70 | TestTupleObjectViewRefused: got nil instead of located refusal |
| element-bypass | TestCheckedViewElementP05: exit 0 instead of ruled exit 70 |
| destructure-bypass | TestCheckedViewDestructuredUnion: exit 0 instead of ruled exit 70 |
| skip-conversion-check | TestScalarUnionViewMisfitNumber/Boolean: native and sanitized exit 0 instead of 70 |
| unfiltered-native | TestCallableProducerRegistryFiltersDirectFunctions: emitted registry contains comparison of distinct pointer types |
| direct-producer-certificate, adapted | TestCallableProducerCertificatesUseClosureThunks: producer certificate includes direct function |
| assignable-direct-certificate | TestCallableProducerCertificatesUseClosureThunks: producer certificate includes direct function |
| exact-identity | TestCallableProducerLiteralReturn: JavaScript backend stdout differs |
| skip-adapter-refusal | FewerParameters/MethodShorthand/ExtraOptional refusal tests: got nil |
| skip-result-registry-bound | TestCallableProducerDiscardedObjectResultRefused: got nil |
| eager-tagged-callable | TestCallableProducerTaggedUnreadMethod: Lower refused acceptance row |
| 147-name-wide-undefined-write | TestFX7P04/P06: native stdout differs |
| 148-name-wide-read | TestFX7P75/P77: backend stdout differs |
| 149-name-wide-null-write | TestFX7P08/P59: native stdout differs |
| final-spread | TestCheckedViewSpread: exit 0 instead of 70 |
| final-in | TestCheckedViewIn: exit 0 instead of 70 |
| final-keys | TestCheckedViewKeys: exit 0 instead of 70 |
| final-keys-alias | TestCheckedViewKeysAlias: exit 0 instead of 70 |
| final-values | TestCheckedViewValues: exit 0 instead of 70, literal member contract bypassed |
| final-entries | TestCheckedViewEntries: exit 0 instead of 70, literal member contract bypassed |
| p19-read | TestCheckedViewElementP19Read: exit 0 instead of 70 |
| p72-input | TestCheckedViewElementP72: compatible string input makes expected misfit stop disappear |
| in-empty-selector | TestViewInEmptyKeyControl: JavaScript stdout differs because an unrelated field is checked |
| spread-method | TestViewSpreadMethodRefused: got nil instead of own-slot refusal |


The constructor-revert additionally fails runtime Node/native exit comparisons for the original four constructor fixtures. No compiler or clang build failure is counted as a kill. Certificate obligations overlap as candidate 3 recorded; the two-filter mutation is run independently for both member obligations.

Wrong-aborts and supplemental evidence:

| Mutant | Observed catcher |
|---|---|
| reject-tuple-objects | TestTupleRecognitionP68, JavaScript stdout differs from source Node |
| missing-tuple-marker | TestTupleRecognitionP68, JavaScript stdout differs from source Node |
| admit-arbitrary-arrays | TestTupleRecognitionRejectsArrayObjectView, checked rejection disappears |
| clear-contract | TestCallableContractChecksBeforeArguments, argument effects run instead of rejecting first |
| source-argument-layout | TestCallableContractWideProducer, native stdout differs from Node |
| missing-producer-mask | TestCallableContractWideProducer, native stdout differs from Node |
| spread-refusal | TestCallableContractSpreadRefused, located refusal becomes nil |
| checked-cast-metadata, isolated | TestCheckedCastWithoutReadSummary, valid native program aborts with exit 70 |
| historical single-filter certificate | Still survives, exit 0, as in candidate 3; redundant ABI filter masks it |

The unadapted p18 metadata attempt also passed; its failed runner assertion and logs remain visible. The isolated rerun catches that same source mutation. Every source was restored before final controls, counts, guard and lane verification. Saved mutant sources are .diff or .patch, never .go.

Final warm leaf durations (seconds), including all inherited new wrong-abort controls:

- TestCheckedCastWithoutReadSummary: 0.72s
- TestCallableContractSpreadRefused: 0.13s
- TestTupleRecognitionRejectsArrayObjectView: 0.61s
- TestCallableContractP18: 1.04s
- TestTupleRecognitionP68: 1.12s
- TestTupleRecognitionP69: 1.15s
- TestCallableContractPlainSpread: 1.21s
- TestCallableContractChecksBeforeArguments: 0.48s
- TestTupleRecognitionOrdinaryArray: 0.90s
- TestCallableContractReceiverBinding: 0.97s
- TestCallableContractWideProducer: 0.96s
- TestCallableContractReceiverEvaluatedOnce: 1.03s
- TestFX7CheckedCastFieldTypes: 0.70s
- TestCallableContractP20: 0.81s

Only TestCheckedCastWithoutReadSummary is a new top-level test added by this integration; it also passed its first run in 0.44s. The five activated review leaves passed in 5.22s, 2.27s, 1.62s, 1.41s and 2.34s respectively in the explicit named run. All are below the unit limit. The full integration gate was not run.
