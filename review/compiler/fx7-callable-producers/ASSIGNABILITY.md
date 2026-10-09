# Item 135: assignable callable producers

Items 133 and 145 were pushed first at 23b47088291cadf8888511e0e7d69267d6e77c05. This second delivery addresses all four named item 135 shapes.

Producer certification now uses the checker's directional IsTypeAssignableTo rather than exact type identity, together with an explicit native ABI check. Non-receiver closures with the same fixed parameter representations and a registry-supported scalar or void result are certified. Any compatible member suffices for a callable union; other members do not introduce an eager refusal. The original p23 literal-return producer therefore compiles and agrees with Node; no runtime signature broadening or backend representation guessing is needed.

Known assignable implementations with different arity, rest, parameter/result storage, or receiver conventions need an adapter that the current signature registry cannot express. The shared graph's joined local/result sources and allocation projections find known closure implementations at a demanded checked read. Nested object/union field contracts retain that demand path. Tagged object selectors defer unread callable descendants; a new tagged-method control pins that laziness. The original p22, p59 and p62 now return located NotYet diagnostics at main.a:7:18, main.a:8:15 and main.a:9:15 respectively, before either backend emits output. Each diagnostic says to wrap the producer in an arrow with the view member's exact parameters and result. The original source Node controls print true. These review witnesses moved from pending agreement to active refusal.

Exact-signature arrow controls p22_adapted, p59_adapted and p62_adapted compile and agree with Node in JavaScript and native. p23 and the three adapted fixtures are counted execution rows. All new test leaves are top-level and parallel.

Current focused leaf seconds:
--- PASS: TestCallableProducerExtraOptionalRefused (0.18s)
--- PASS: TestCallableProducerFewerParametersRefused (0.19s)
--- PASS: TestCallableProducerLiteralReturn (0.42s)
--- PASS: TestCallableProducerFewerParametersAdapted (0.53s)
--- PASS: TestCallableProducerMethodShorthandAdapted (0.53s)
--- PASS: TestCallableProducerExtraOptionalAdapted (0.55s)
--- PASS: TestCallableProducerCertificatesUseClosureThunks (0.24s)
--- PASS: TestCallableProducerP61 (0.38s)
--- PASS: TestCallableProducerMethodShorthandRefused (0.17s)
--- PASS: TestCallableProducerP60 (0.35s)
--- PASS: TestCallableProducerP21 (0.33s)
--- PASS: TestCallableProducerRegistryFiltersDirectFunctions (0.37s)

The current runner run-assignable-mutants.py supersedes the first-delivery run-mutants.py for the final implementation (the new ABI guard independently rejects direct records too):
- assignable-unfiltered-native.diff removes the native registry guard; stale direct certificates make clang reject comparison of distinct pointer types, as explicitly requested.
- assignable-direct-certificate.diff removes both lowering exclusions of direct functions; the certificate invariant test catches a direct function after Node agreement. The native defensive filter otherwise masks this metadata error.
- exact-identity.diff restores exact checker type identity. TestCallableProducerLiteralReturn fails the JavaScript/Node output comparison because the valid producer no longer receives a certificate.
- skip-adapter-refusal.diff removes only the read-time ABI refusal. Each of the three refusal leaves fails because Lower returns nil. Production sources are restored in finally blocks.

Commands, logs in this directory:
- timeout 180 go test ./internal/lower -run '^TestCallableProducer' -count=1 -v -timeout 90s: PASS 0.906s, assignable-focused.log.
- timeout 1000 python3 review/compiler/fx7-callable-producers/run-assignable-mutants.py: PASS runner, six expected test exits 1.
- timeout 300 go test ./internal/lower -count=1 -timeout 240s: assignable-lower.log.
- timeout 120 go test ./internal/ir -run TestCallTargetReaders -count=1 -timeout 90s: assignable-readers.log.
- timeout 360 go test ./internal/oracle -run 'TestCountsAreRecorded|TestNativeAgreesWithNode/internal/lower/testdata/callable_producers|TestReviewProgramsAgreeWithNode/fxspptb_oct9_(views_p(21|23)|native_p(60|61))_|TestReviewProgramsRefuse/fxspptb_oct9_views_p(22|59|62)_|TestReviewProgramsNoLooseFiles' -count=1 -v -timeout 5m -args -update-counts: assignable-oracle-counts.log.
- Integration lane checks after commit: git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -.

Scope: fixed scalar callable signatures. Arity-changing and receiver adapters remain unimplemented and are narrowly refused for known reaching implementations, not advertised as callable support. Unknown producer frontiers retain existing runtime checks; this change does not claim to prove their logical types. The refusal is conservative across possible union arms and known projected sources. No global field-name rejection was added. No protected emitter/orchestration file was edited, no cohere source was copied, and no full repository gate or PR was run/opened.

Final soundness review also excludes object/maybe results even when the checker permits a discarded void result, since the producer registry cannot record those implementations. TestCallableProducerDiscardedObjectResultRefused observes Node true and requires located NotYet; skip-result-registry-bound.diff makes it fail with nil instead. eager-tagged-callable.diff demands a receiver method during tag-only selection; TestCallableProducerTaggedUnreadMethod rejects that eager refusal. All six mutants are in the final run-assignable-mutants.py.

The stage3 fixture callable-producer-wrong.a is historically misnamed: an implementation taking number can accept the union member's literal 1 argument. Its checked_views_v2_migration_test.go leaf now starts with Node agreement, then removes Functions from certified contracts and pins the resulting unknown-signature stop in all three backends. That existing mutant leaf passed in 0.70s. Its counts row legitimately changes from a runtime stop to successful allocation/free balance. No fixture source was rewritten.

An initial result-bound mutant was masked by the old eager refusal on a second callable-union arm. The final any-compatible-member logic removes that false refusal, so the result-bound mutant now specifically proves the unsupported-result guard. Sources were restored after each run.

Final focused run: PASS 4.509s. All current new leaf timings are in assignable-focused.log; tagged unread method 3.96s, discarded object result refusal 0.17s, all remaining leaves under one second. The earlier 0.906s focused receipt predates the two soundness-boundary leaves.

Final validation on restored final sources:
- Full internal/lower: PASS 72.535s (assignable-lower-final.log).
- TestCallTargetReaders: PASS 23.934s (assignable-readers-final.log).
- Selected oracle originals, adaptations, active refusals, tagged control, existing migration mutant, and TestCountsAreRecorded -args -update-counts: PASS 82.267s (assignable-oracle-counts-final.log). The final command also selects TestCheckedViewV2CallableProducerMutant.
- counts.md: five new execution rows plus the corrected historical valid-producer row.
- New oracle execution leaf seconds: p23 1.68, p22_adapted 1.61, p59_adapted 1.49, p62_adapted 1.43, tagged unread method 18.83. All below the leaf budget.
- git diff --check: PASS.

Final item 135 lane checks PASS 2.0s: gofmt/tools on 26 Go files, parallel rule on 2 test packages, vet 2 packages.
