V4 full sweep for #19541xg: merged main and repaired four V4 regressions.
Commits: 4fb2ecbd2 and 22524743f (main merges), 3d7aa1a3c, de06d2b2d, 046a3a5aa, 8900983e0 (fixes), ded33d96c (counts).
Commands: admission checker over 1,997 fixtures; per-name Go tests with -count=1 -timeout 90s and outer timeout 90; counts update and independent verification passed.
Mutants: existing certified-union and receiver-domain omissions, unrelated-allocation-proof omission, and duplicate iterator receiver all caught; details below.
Not covered: Set certificate, rank-165 readonly control, checked construction and runtime-callback source-site blame remain follow-ups; no full repository gate ran.

| Required sweep | Final result | Seconds |
|---|---|---:|
| Every oracle and stage3 .a fixture | 1,997 compared; zero unexpected admission regressions; six ruled refusals; two inherited panics | 314.309 summed main/V4 probe time |
| Linux sanitized native records | All eight correctness/mutant leaves passed, including main's new two-index witness | 100.423 summed final commands; max 38.540 |
| Relevant stage1 gaps | All 31 leaves passed across ten packages | 218.056 summed final commands; max 30.504 |
| TestCountsAreRecorded | Complete update passed; independent verification passed; four moved rows named below | 60.602 update; 56.115 verify |
| Oracle checked-view and V4 suite | 156 top-level leaves: 153 passed, three existing pendings skipped | Per-leaf results in final-results.json; max command 33.366 |
| TestCallTargetReaders | Passed after compiler fixes | 31.769 |
| Lane checks | Passed before evidence packaging; final run recorded separately | See lane-checks.log |

The initial main tip was bdb89962b178f618e2e6d34e9a7ea5c395089faf. The initial merge's only conflict was the call-target reader allowlist: retain both V4's callable emission reader and main's class-static test reader. Main advanced with test-only changes to 7d113268b1903e4e47289f226e94ec2fb609e15b; merge 22524743f incorporates that tip. All affected records and JSON/YAML gaps leaves were rerun (refresh.jsonl). No .a fixture files changed between those main tips.

Setup: GOPROXY='https://proxy.golang.org|direct', bash cloud/setup.sh, source /workspace/adamic-tools/env.sh. Setup took 41.653s (Go build ready at 41.400s); nproc=5, cgroup quota=4 CPUs. setup.log preserves every timing line.

Admission comparison uses load.Load and lower.Lower, not just TypeScript acceptance. fixtures.json lists every path; check.go.txt preserves the helper without adding a Go consumer. Main was built through a Go overlay of its exact internal source/runtime files, with V4-only files excluded; cohere uses the identical pinned submodule. admission-final.jsonl retains all baseline results and rechecks the initially regressed fixtures after repairs, plus the remaining six refusals on the final compiler. Forty-nine fixtures are admitted by V4 that main refuses; 1,427 are admitted by both; 515 are non-admitted by both, including the two inherited panics.

The six permitted new refusals are the .a rule that an escaping callable relation must be proven. Each diagnostic includes source path, the failing read and a fix to prove producer parameters/results or call directly:

- internal/oracle/testdata/review/agree/fxspptb_oct9_views_p22_callable_union_fewer_parameters.a: exact V2 certificate does not prove the widened callable relation (ordinary widening remains outside V4).
- internal/oracle/testdata/review/agree/fxspptb_oct9_views_p23_callable_union_literal_result.a: exact V2 certificate does not prove the widened result relation.
- internal/oracle/testdata/review/agree/fxspptb_oct9_views_p51_callable_union_destructure_wrong.a: misfit producer escapes through destructuring.
- stage3/interface-downcasts/untagged/fixtures/callable-union-nested.a: nested misfit callable escapes.
- stage3/interface-downcasts/untagged/fixtures/callable-union-wrong.a: wrong callable producer escapes.
- stage3/interface-downcasts/v2/callable-producer-wrong.a: wrong callable producer escapes.

Inherited failures, unchanged in main and V4: internal/oracle/testdata/review/agree/fxspptb_iterators_derived_symbol.a and internal/oracle/testdata/review/agree/fxspptb_iterators_override_source.a panic in checkMemberOverrides on Node.Text(*ast.ComputedPropertyName). They are recorded as probe failures, not sound compile refusals. They were not repaired because they are not V4 regressions.

Fixes, one per commit:

- 3d7aa1a3c: preserve escaping callable unions when existing V2 exact producer certificates already prove every reaching producer. No synthetic metadata. Wrong .a producers stay refused; their Node controls and diagnostic pins replace obsolete runtime-read negatives. The admitted producer-control fixture still catches certificate omission and matches Node when that mutant is neutralized, in both backends and native sanitizers. union-proof-controls.log, union-proof-negatives.log and v2-repaired-final.log record the checks.
- de06d2b2d: an explicit method this annotation describes the existing receiver local, rather than a second runtime parameter. Restores the namespace_method_receiver and both stage3 native-namespace-object-receiver admissions. Node, JavaScript, release native and ASan/UBSan/LSan agree. The existing receiver-domain omission mutant remains caught. method-receiver-fix.log records the new leaf at 1.31s.
- 046a3a5aa: unrelated reference fields sharing a checked field name use their ordinary read/write representation only when the closed allocation-flow proof excludes every viewed allocation. Unknown receivers retain checks. The array-write obligation applies to array payload schemas. Restores fxspptb_oct9_native_p08_null_array_write.a. Removing both unrelated-write certificates produces the false-positive exit 70 in both backends and sanitized native, so the regression is pinned. Existing checked-array write refusals remain green. unrelated-array-mutant.log records the new leaf at 1.03s.
- 8900983e0: iterator member lowering already packs the receiver; mark that ABI explicitly so V4 does not prepend another null receiver. Restores user_iterators.a counted execution. Full existing Node comparison passes, including release/JavaScript/native sanitizers (iterator-fix.log, 5.38s). Clearing ReceiverPacked is caught by the native sanitizer; the minimal new leaf is 0.80s (iterator-mutant-final.log). JavaScript's already-correct iterator ABI is unchanged.

Counts changed only these four rows, in column order allocations/frees/retains/releases/peak/in-regions:

| Fixture | Before | After | Cause |
|---|---|---|---|
| internal/oracle/testdata/namespace_method_receiver.a | 7/7/9/16/5/0 | 7/7/7/14/5/0 | Remove duplicate null receiver parameter and its retain/release on each of two method calls. Generated receiver-main.c.txt versus receiver-v4.c.txt proves the cause. |
| stage3/interface-downcasts/untagged/fixtures/callable-union-nested.a | 3/0/6/4/3/0 | No runtime row | Pinned .a escaping-read compile refusal. |
| stage3/interface-downcasts/untagged/fixtures/callable-union-wrong.a | 2/0/5/4/2/0 | No runtime row | Pinned .a escaping-read compile refusal. |
| stage3/interface-downcasts/v2/callable-producer-wrong.a | 2/0/5/4/2/0 | No runtime row | Pinned .a escaping-read compile refusal. |

counts-changes.txt and commit ded33d96c give literal before/after rows. The complete updater, without a filtered fixture list, passed in counts-update-alone.log; counts-verify.log independently verifies the resulting table. The three removed rows' fixtures still run Node/refusal checks in the counts fixture family.

Evidence and limitations: tests.jsonl preserves original failures; final-results.json explicitly links repaired leaves and successful counts logs. admission-premature.jsonl is an abandoned pre-baseline-build run and is excluded from conclusions. Exploratory failed logs remain for audit. Cold compilation and competing jobs caused hard 90s count timeouts. Two later count attempts passed all 934 primary fixture leaves but failed in additional counts because the build-cache filesystem filled; disposable obsolete Go cache entries were removed and the updater reran alone. The final counts run is a pass, not an inferred result. The exact-main isolated user_iterators counts leaf passes; its parent table comparison intentionally fails because only one leaf was selected, so that command is not reported as a whole counts pass.

Existing pendings seen in the full oracle sweep: TestV4EscapeAdapterBlameRuntimeCallback, TestV4Lane5FactoryWiderWritePending, TestV4Lane5FactoryReadonlyTargetControl. The Set certificate and rank-165 readonly control remain deferred as requested. Runtime-callback blame needs source call-site metadata on callback-producing IR and forwarding that site through the native trampoline/JavaScript callback boundary while preserving adapter read-site metadata; runtime callbacks currently lack that source site. No new JavaScript oracle divergence was introduced by these sweep fixes. The opt-in records benchmark was excluded; sanitized record correctness and all six record-mutant units were included.
