Two rows defended in a bounded expanded matrix; SourceFlows not defended after three attempts.
Production starts from origin/main 7d113268b1903e4e47289f226e94ec2fb609e15b; nproc 5. Earlier evidence retained unchanged.
Evidence includes 33 complete top-level rows plus four subcases of TestNativeAgreesWithNode; no full-package uniqueness claim.

Code under test and oracle

SourceFlows: checked-view lowering and values transported through helpers, generic instantiation, callbacks and stored objects; native closure-result emission. Node controls execute and must print true. The expected backend stdout and exact Adamic-only panic text are self-written. The recursive subsumer tests object-graph recursion, without these helper/callback transport histories. Per-test compiler coverage has 893 blocks covered by SourceFlows and not Recursive. Generic instantiation and native callThrough are among the differing paths.

OwnClassData: native checked-view own-slot admission in runtime/view_unions_untagged.c, reached through lowered class-instance data and untagged union membership. Node executes; successful backends compare to Node, while wrong-class data uses a self-written exact panic pin. Unlike Recursive's plain objects, this row supplies an actual class instance whose own numeric kind must be read. Its Go coverage has 557 exclusive blocks against Recursive, including class construction. Go coverage does not instrument embedded runtime C. C1 flips the runtime class is_static condition, rejecting instance slots instead of static-class slots. The good subcase exits 70 instead of 0 in release and sanitized native runs; JavaScript still exits 0 and prints true. Only this top-level row fails in the bounded expanded replay.

ArrayPending: lowering's unsupported array-element refusal, particularly arrayElement in internal/lower/object.go. Oracle is self: NotYet type plus either a views-v3 array-element reason, or exactly an array of never for the empty case. No Node runs in this row. All five execution subcases skip after the diagnostic assertion. Its profile has 73 exclusive Go blocks against OwnClassData. A1 changes the production prefix an array of to an array with elements of. Only the empty subcase fails because it pins the older wording. This defense protects that diagnostic distinction, not runtime array admission or execution. No stronger semantic protection is claimed.

Attempts and broadened replay

F1, generic.go:131: drop the concrete typeMapper assignment. SourceFlows and every replayed row pass. No changed source-level behavior was established; it is a survivor, not a claimed uncovered bug.
F2, emit_functions.go:365: remove ownership bookkeeping from the closure-reference-result return. All original 33 rows pass. TestNativeAgreesWithNode catches it with leaks at oracle_test.go:787 in all four additional fixtures. SourceFlows does not call the separate leaks helper; nativelyUncached ordinarily sets detect_leaks=0, as the unchanged harness shows. The name promises source flows, not leak checking, so this is not a demonstrated name/assertion mismatch.
F3, emit_functions.go:365: return NULL early for closure reference results. SourceFlows/callback/good prints false rather than true in both native backends, failing checked_views_v2_source_test.go:143. TestNativeAgreesWithNode also catches it in all four additional fixtures, including function_values_return.a at oracle_test.go:774 (exit codes differ). This is not unique, so SourceFlows is not defended after the three attempts.
C1, runtime/view_unions_untagged.c:44: flip the own-slot static-class condition. OwnClassData alone fails, at checked_views_v2_source_test.go:168 (native: exit codes differ). Other 32 complete rows and the additional differential-family subcases pass.
A1, object.go:335: change unsupported-array diagnostic prefix. ArrayPending alone fails at checked_views_v2_source_test.go:211: stage 0 can't lower an array with elements of never yet. Other 32 complete rows and additional differential-family subcases pass.

The first 33-row matrix made F3 appear unique. A subsequent reachability check found closure-return fixtures in the package's general differential family. Their clean baseline passed, and replay disproved uniqueness. Both matrices and the correction are retained. No mutation was aimed at or applied to the tests, Node, disagreement helper, or fixture inputs.

Scope and commands

Fresh go test -list . defines current package scope. matrix-rows.json lists all 33 original complete rows; expanded-matrix-rows.json adds TestNativeAgreesWithNode. Only closures.a, function_values_chain.a, function_values_return.a and generic_functions.a were run from that large family. The exact filter is in differential-family.regex. Other subcases and excluded-rows.json outcomes are unknown. This bounded selection is a conservative reached-code replay, not exhaustive package-wide reachability.

Whole-package clean baseline timed out at 90.122 seconds with no earlier top-level failure. The 33-row baseline passed at 15.318 seconds, with eight skipped subcases recorded exactly in baseline-skips.json: five pending-array executions and three comment/flag projections. The additional differential-family baseline passed in 0.681 seconds. Current view/interface/class/generic tests beyond the audit's original slice are included. No new top-level assertion skipped.

Coverage commands use -coverpkg=./internal/lower,./internal/native,./internal/javascript and -run '^Name$', one profile for each target and Recursive. Exclusive block files retain exact source coordinates. OwnClassData is itself the comparison profile for ArrayPending. Runtime C is uninstrumented, so no exclusive C line claim rests on Go profiles.

All matrix commands use timeout 120 go test -json -count=1 -timeout 90s, with the saved scope regex or differential-family regex. ADAMIC_GATE_UNCACHED=1 prevents oracle-result cache reuse; every mutant has its own ADAMIC_BUILD_CACHE_DIR=/tmp/defend-views-flows/cache/ID. Go mutants use scratch overlays, while standalone diffs apply directly without a switch. C1 physically changed the embedded source for its two runs, then was restored. Native runtime archives are content/flags keyed under os.UserCacheDir(), forcing a rebuild for the changed C bytes even though that archive cache does not use ADAMIC_BUILD_CACHE_DIR.

Every diff passed git apply --check against unchanged production at origin/main. F1/A1 passed go vet ./internal/lower with their overlays; F2/F3 passed go vet ./internal/native with theirs. C1 passed release clang using native.Flags(Options{}), go vet ./internal/native, and native/sanitized runtime builds in the matrix. All production and tests now match origin/main. No test was rewritten or weakened.

Timings

Warm env.sh worked; setup rerun 0 seconds. npm ci in stage3/api reported three packages installed in 587 ms, before baseline. Main matrix binary times: F1 14.681 s, F2 15.613 s, F3 14.582 s, A1 14.390 s, C1 38.900 s. Additional family replays take under one second each. Runtime rebuild time is included in C1 and was not separately measured. Raw coverage timing lines and all run outputs are saved.

Brief ambiguities, costs and owner findings

- The requested branch already held a completed defense for different rows. It was preserved, merged with current origin/main without history rewriting, and this run uses flows-ownclass-array/. Production equality to origin/main was checked before baseline and after restoration. Earlier evidence was not substituted for this session's measurements.
- /tmp is only 8.8 GB total, so 15 GB free is impossible. Free space was 3.0 GB in /tmp and 6.8 GB in /workspace initially. Only identifiable earlier /tmp/defend-element scratch (92 KiB) was removed, with no tools/repository deletion. Final free space was 3.0 GB and 6.3 GB. No disk exhaustion occurred.
- The full package cannot complete within 90 seconds. Unique kills for the two defended rows are explicitly bounded; all other rows/subcases remain unknown. Passing-row lists are saved, rather than implying the whole package passed.
- ArrayPending's name and assertions match a pending refusal receipt. Its defense rests on a self-written diagnostic pin; all runtime subcases skip. A wording-only mutant does not prove meaningful runtime support.
- SourceFlows was not defended despite making a real production mutant fail. The differential family also catches that mutant. Its name promises helper/generic/callback/stored transport and its assertions exercise all four; no missing named assertion was demonstrated. Non-defense is not a deletion recommendation.
- F2 drops an ownership call inside a return expression rather than a whole standalone statement. Under a strict whole-statement interpretation of the menu, treat it as supplemental. No defended verdict relies on F2; F3 is the permitted early-return mutation and is demonstrably caught elsewhere.
- Full native runtime C coverage, complete package reachability, and full differential-family replay were not obtained. Broader reachability already changed one tentative verdict, demonstrating why the remaining bounded claims need central replay.
- A network read attempted without escalation could not reach the configured proxy. The requested remote branch fetch worked with authorized escalation. No approval rejection occurred.

Artifacts

rows.json contains per-row verdicts and exact commands; matrix.json contains both replay stages, complete passed-row lists and unknown limits. F1/F2/F3/C1/A1.diff are standalone production changes. Raw logs, vet/clang logs, per-row profiles, exclusive-block files, tests-list.log, baseline-skips.json and provenance.json provide the session evidence. Production sources are restored; no PR or main push is requested.
