Survivors: historical-single-filter remains equivalent under the independent ABI filter; exact original-spread-method also survives because creation admission rejects the method. Its two-guard adaptation is caught.
Built: attempted cleanup merge onto the requested train-plus-2 base; no green delivery was produced.
Commits: base 0c6b8384df2465dac8ee1dbdf7e88cd9dc47d196; requested member fb00e043a4a770f3041af6f5ea6ea34ed94d9174; no merge commit or push.
Checks: full lower FAIL (one stale refusal), fixture oracle FAIL (four creation refusals), counts update FAIL before writing; readers and vet PASS; 39 intended mutation obligations caught.
Limits: no successful counts regeneration, committed-tip lane checks, push or shared full gate; the attempted source tree is saved in attempted-merge.patch.

This is a blocked integration result for the views fix-forward work, not a successful delivery. The exact specified base and member override the generic main-start instruction. The work branch was compiler/miscompile-train-plus-2. No other branch was changed or pushed. The unsuccessful merge is aborted after retaining this evidence so the branch remains at its requested base.

Observed blockers on the merged code, with both parents' admission guards retained:

| Activated fixture | Observation | Contract builder cause |
|---|---|---|
| native p26 callable_or_undefined | creation refuses value at 7:14; source Node prints true and 2 | prepareUntaggedCallableUnionRead rejects the undefined member as lacking a producer signature |
| native p33 map_in_union | creation refuses value at 7:16; source Node prints 1 | internMixedUnionViewContract rejects the unavailable Map member contract |
| native p34 typed_array_in_union | creation refuses value at 5:16; source Node prints 3 | same builder rejects the unavailable Uint8Array member contract |
| native p35 class_in_union | creation refuses value at 6:16; source Node prints 3 | same builder rejects the unavailable Point member contract |

The exact nested builder errors and descriptors were obtained through temporary logging, restored immediately (guard-debug.log). The public refusal comes from checkViewMembers. Native untagged emission also turns nominal/unsupported contracts into ViewUnknown. Inference: admitting these probes while retaining the train's completeness guard needs additional contract implementation, not a count resolution or deletion of the guard. That implementation is outside this merge unit. The four failures also prevent a valid counts table and successful backend agreement on all activated fixtures.

Full lower also observes TestUnsupportedViewCallableMember returning a program and no error: candidate 4 now supports this fixed scalar callable union. Its old creation-refusal assertion is stale, not evidence of a wrong output. The full run completed in 39.211s and had no other failing leaf (lower-final.log). An earlier concurrent run reached its 90s package limit and is retained separately; the final run used -timeout 210s.

Conflict resolutions saved in attempted-merge.patch:

- JavaScript untagged membership retains the train's loop-based recursion and optional-union tail handling, plus candidate 4's tuple branding checks for object and fixed-tuple membership.
- view_lazy.go retains checkViewMembers and recursive unsupportedViewContract from the train; deletes only cleanup's unused lazyReadRefusal.
- Cleanup's three source-location tests are blocked by creation admission. Their attempted adaptation first asserts the 5:14 creation refusal, then independently pins readViewMember's ViewWhere to the original read AST location. All three pass in 0.05s each; blanking ViewWhere fails all three. Production creation admission stays intact.
- counts.md was never hand-resolved. The single requested update attempt failed before writing the generated table. The saved patch includes its unresolved conflict for transparency, not an approved table.

Counts: no regenerated row exists to attribute. The command failed in 107.660s on exactly p26/p33/p34/p35, before counts_test.go writes the table. Eleven registered rows are imported from the member: p18/p20/wide from callable repair, p68/p69/extended p69 from tuple recognition, p17/p26/p33/p34/p35 activated by ca88bf56. Their numbers in the conflict are inherited historical numbers, not newly verified merged counts. The conflicting p04 10-versus-8 retain values are the already-attributed stale NULL-retain discrepancy from d3d2e838, documented and measured in review/compiler/miscompile-train-plus/REPORT.md. The base's undefined-read-write.a retains=4 has the same cause. Neither value was changed by hand. No unexplained count movement is claimed; regeneration did not complete.

Verification commands all sourced /workspace/adamic-tools/env.sh and redirected output to logs:

```text
export GOPROXY='https://proxy.golang.org|direct'
timeout 600 bash cloud/setup.sh
PASS: Go ready 0.020s; Node ready 0.019s; submodules ready 0.055s; markdown dependencies ready 0.060s; clang ready 0.141s; build ready 43.769s; test binaries deferred 43.904s; cache warm 43.905s; done 43.932s. nproc=5; cpu.max=400000 100000.
timeout 150 go test ./internal/lower -run 'TestCallableContract|TestTupleRecognition|TestViewDiagnosticP' -count=1 -v -timeout 90s
FAIL: original diagnostics masked by creation. Callable and tuple controls pass (interaction.log).
timeout 120 go test ./internal/lower -run '^TestViewDiagnosticP' -count=1 -v -timeout 90s
PASS 0.063s after test-only isolation (diagnostics.log).
timeout 240 go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 210s -args -update-counts
FAIL 107.660s; four creation refusals; no generated table (counts.log). Only one update attempt.
timeout 150 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -timeout 90s
PASS (readers.log).
timeout 180 go vet ./internal/lower ./internal/native ./internal/javascript ./internal/oracle ./internal/ir
PASS, empty output (vet.log).
timeout 240 go test ./internal/lower -count=1 -v -timeout 210s
FAIL 39.211s, only TestUnsupportedViewCallableMember (lower-final.log).
timeout 180 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/lower/testdata/(fx7_wrong_aborts|fx7_tuple_recognition)|TestNativeAgreesWithNode/internal/oracle/testdata/review/agree/fxspptb_oct9_(native_p(17|18|20|26|33|34|35)_|views_p(68|69)_)' -count=1 -v -timeout 150s
FAIL 1.246s: p26/p33/p34/p35; seven other registered fixtures pass Node/JavaScript/native/sanitized comparisons (fixtures-final.log). The first attempted combined-path regex selected no leaves and is not evidence (fixtures.log).
timeout 150 go test ./internal/native -run '^TestCheckedCastWithoutReadSummary$' -count=1 -v -timeout 90s
PASS 0.252s after restoration (metadata-restored.log).
```

Full lower also exercises every added .a source: p18/p20/wide/p68/p69 in Node and both backends, bad_callable as a deliberate pre-argument rejection, and three diagnostic sources as creation refusals with independently tested read locations. The source-only Node commands for diagnostic p17/p48/p64 print 1; p26/p33/p34/p35 print true+2, 1, 3, 3. No backend agreement is claimed for refused sources. No test source or fixture was weakened in the repository delivered state.

Mutant reruns, all restored in finally, bounded Go children and outer shell commands:

- timeout 840 python3 review/compiler/miscompile-train-plus-2/run-members.py: 29 obligations caught (member-mutants.log and fx6/ logs/patches). nullable-receiver, generic-receiver, union-receiver; revert-receiver, revert-destructure-type-id, revert-p70; element-bypass, destructure-bypass; skip-conversion-check; unfiltered-native, adapted direct-producer-certificate; assignable-direct-certificate, exact-identity, skip-adapter-refusal, skip-result-registry-bound, eager-tagged-callable; name-wide undefined-write/read/null-write; final-spread/in/keys/keys-alias/values/entries; p19-read, p72-input, in-empty-selector, adapted spread-method. Catchers are the same recorded test names and semantic diagnostics as candidate 4, with actual logs here.
- timeout 600 python3 .../run-extra.py then run-extra.py retry: constructor revert and tuple reject/marker/arbitrary-array mutations caught; callable clear-contract/source-argument-layout/missing-producer-mask/spread-refusal caught. The arbitrary-array patch needed only context adaptation from .some to the train's for-loop; both failure and adapted patch are retained. Eight behavioral kills in extra-mutants/results.json.
- timeout 300 python3 .../run-isolated.py: same cast-metadata edit caught by TestCheckedCastWithoutReadSummary, exit 70 (isolated-results.json). Its lowered p18 witness is independently masked by read summaries and passes, as previously recorded; that passing experiment is not counted as the kill.
- historical-single-filter survives with exit 0, as explicitly allowed by the brief. The adapted two-filter certificate removal is caught in the member run.
- timeout 180 python3 .../run-location-mutant.py: blank ViewWhere caught by all P17/P48/P64 location assertions; source restored.
- timeout 150 python3 .../run-original-spread.py: exact candidate-4 spread-method.patch survives, exit 0 (original-spread.json), masked by creation admission. The adapted member run removes both creation and own-slot guards and is caught. This is the same masked exact mutant identified in the previous train-plus delivery, not a claim that the original patch was killed.

All 38 intended candidate-4 obligations and the cleanup location obligation are caught using the recorded isolation/adaptation where necessary. The named masked exact experiments remain visible. No compiler build failure is counted as a mutant kill.

No merge commit was made because counts regeneration and fixture acceptance fail. Therefore the mandatory post-commit lane checkpoint could not be run on a merged tip, and no branch was pushed. Running it against the old HEAD would inspect the base, not this merge. The shared gate was not run. To finish this unit, the four unavailable contracts need a sound supporting dependency (or integration must explicitly withdraw their activation); the obsolete fixed-callable refusal expectation also needs updating. No mid-turn approval or answer was requested.
