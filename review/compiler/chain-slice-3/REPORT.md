# Slice 3

Built five independent lowering members on main 5e33a17b186a8a2218d27b69b21e2de5acc5b750 for #wj4pmt1.
Member commits: 9d23d4bc, 2048b4a3, 76a0da87, 29907656, 1ec1e4c3; owner repairs and evidence follow.
Bounded lowering shards, uncached Node-held member fixtures, recorded mutants, admission delta, counts and call-target guard pass; lane output is recorded separately.
Twelve production overlays fail behaviorally; two spread and six assignment source mutants are caught by their Node-held tests; the cache flag mutant fails its intended compile-contract check.
Not covered: other chain members, full gate, full checker performance or JavaScript support for checker-library calls; integration acceptance remains external.

## Slice-3 member rows

| Member | Source sha | Kept or dropped | Dependency evidence | Tasks closed by this delivery |
|---|---|---|---|---|
| optional-chain-after-call-main | 5028fa69 | kept | Defined.Throws is supplied by this member at internal/ir/defined.go:6; CallMayThrow exists on main. Isolated main-plus-member production build passes. | #ht2nwj5 |
| spread-shorthand | e82c9646 | kept | Array argument expansion is supplied at internal/lower/array_call_arguments.go; existing main array/JSON helpers suffice. Isolated main-plus-member build passes. | #aqh658t, #kqxkvg0 |
| refusal-rulings-main | 8799f518 | kept | SourceLength is supplied throughout IR, emitters and closure runtime by this member; strictViewContract used at internal/lower/merged_fields.go:131 exists on main. Isolated main-plus-member build passes. | #h3tshdq, #ex6eqh1 |
| assignment-proofs-main | a6f05066 | kept | Test-only proofs use existing main oracle helpers and assignment lowering; independent overlay has no production additions. | #2dxms6q, #cvhj5fk |
| tsgo.go-cache | 63f5bb07 | kept | cachedRuntime and cache machinery exist on main; internal/native/tsgo.go:139 calls them with member-specific flags. Isolated main-plus-member build passes. | Source plan gives no individual task attribution; cache portion only |

No external-member symbol, fixture or helper dependency was observed. Shared files are not treated as dependencies. Isolated overlays replace every production file touched by any candidate with main, then add only the selected refined member patch. Build observations are in independence.json; semantic statements above are audit inferences supported by those builds.

## Own ranges and repairs

members.json and the five member .patch files preserve the refined ranges and exclude stack envelopes, old counts and inherited review material. Each member has one source-named commit.

71972e18: take optional Defined throw-summary OR repair and selector/YAML spread gap hunks. Adapt assignment fixture ownership to main's top-level test contract in TestFixturesAssignmentProofs; add no unrelated fixture directories.
a6cb5660: take parser and CSS push gap proofs, and JSON multiplePush closure only. Keep JSON repeatInTry refusal because exceptions-21 is outside this slice.
a233eec3: take markdown push closure and scanner push proof only. Leave unrelated scanner port, runtime stores and supplementary scanner checks out.
167aaf0c reader/counts interaction: preserve main's approved reader table; use CallTargets in spread mutant tests and regenerate counts.
d8e6f59c required-read native repair changes thrown-payload storage for exceptions-21. Main retains the earlier object-pointer exception storage, so its original member code is retained; importing that interaction would create an external dependency.

The additional summary control initially used separate statements and failed to kill the OR mutant because traversal prunes the later statement. The corrected control uses sibling expressions within one concatenation. It passes in 0.00 s and kills the original assignment mutant. Failed preliminary runs are not claimed as passes.

## Validation

All commands source /workspace/adamic-tools/env.sh; GOPROXY=https://proxy.golang.org|direct was exported before setup. setup.log records timeout exit 124 during whole-repository cache warming; tools and submodules were ready. Timing lines: Node 0.089 s, Go 0.091 s, clang 0.679 s, markdown ready 1.194 s, submodules ready 18.033 s, shared cache 23.723 s. nproc=5. Bounded focused retries completed after warming. No full package test or full gate was run.

run-lower-shards.py lists all 277 top-level internal/lower tests and runs nine shards of at most 32, timeout 90, Go timeout 85 s, parallel 4. All pass; final maximum shard is 29.331 s.

final-member-tests.jsonl records uncached go test ./internal/oracle ./internal/flow ./internal/lower ./stage1/cohere/cssstrings with member-name selection, -count=1 -timeout=85s -parallel=4 -json under timeout 120. Admitted fixtures agree with external source Node, generated JavaScript and release/sanitized native; refusal fixtures retain their exact contracts. Native sanitizers include ASan/UBSan and leak detection. Embedded two spread and six assignment mutants fail their expected Node outputs.

repair-tests.jsonl records the bounded selected stage1 selector/YAML/JSON/markdown/parser/scanner leaves and TestFixturesAssignmentProofs, TestFixtureDirectoriesHaveTopLevelTests and parser probes. The assignment owner leaf is 24.61 s; markdown gap owner leaf is under 16 s. CSS final proof is in final-member-tests.jsonl. No new leaf reaches 60 s.

mutant-catalog.json, mutants-results.json and each mutations test.log record all twelve overlays and exact commands. Every final overlay fails a test assertion, without a build failure or timeout. The cache's recorded include-order flag omission is TestTSGoBuildSeesTheProgramsFeatures; its failure is compile-contract evidence, not runtime evidence.

tsgo-cache-build.log and tsgo-cache-runtime.json record actual BuildTSGo first/warm release and first/warm sanitizer products linked to a freshly built checker archive. Their stdout, stderr and exit match source Node and generated JavaScript for the recorded feature witness arguments_length_extended.a. The first archive build timed out; the warmed retry passed. This proves the retained build/cache path, not checker-library execution performance.

admission-inputs.json fixes 1992 existing main inputs. Final slice observations are rebuilt after owner repairs; main observations are unchanged. admission-delta.json contains seven newly admitted spread/push gaps, each attributed to spread-shorthand, and no newly refused input. All seven have explicit Node-held backend gap proofs.

Counts regenerated successfully once by timeout 310 go test ./internal/oracle -run ^TestCountsAreRecorded$ -count=1 -timeout=300s -args -update-counts: 88.903 s. The initial 85 s attempt timed out before writing. counts-attribution.json attributes all 33 numeric added/changed rows: 22 additions and 11 existing changes. The largest existing change, 09_tree.ts, is reproduced exactly by the isolated main-plus-optional count build (33/33/66/69/16/0). Attribution causes are marked as inference where appropriate; numbers are observations and were never hand edited. No fixture was added after regeneration; the optional summary repair only changes throw metadata.

TestCallTargetReaders: timeout 110 go test ./internal/ir -run ^TestCallTargetReaders$ -count=1 -timeout=85s; passes, 20.471 s on final touched tests. No allowlist expansion.

Lane checks run against committed HEAD using the exact integration command. lane-checks.log holds output: lane checks 13.7 s, gofmt and tools on 34 Go files, t.Parallel on 11 test packages; vet skipped over 10 s. A separate timeout 90 go vet over all 11 changed test packages then exited 0, recorded in vet-final.log. Full gate is intentionally deferred to the shared post-push gate.

## Every recorded production mutant

| Mutant | Catching test |
|---|---|
| check-optional-receiver | TestOptionalAfterCall79 |
| javascript-required-panic | TestOptionalAfterCallRequired |
| native-required-panic | TestOptionalAfterCallRequired |
| drop-flow-throw-edge | TestOptionalAfterCallFlowEdges |
| drop-required-propagation | TestOptionalAfterCallPropagation |
| merged-computed-read | TestMergedFieldTypeScriptComputedChecked |
| merged-declaration | TestMergedFieldAdamicUnusedRefused |
| merged-read | TestMergedFieldTypeScriptMissingChecked |
| function-annotation | TestFunctionAdamicUnusedAnnotationRefused |
| function-call | TestFunctionTypeScriptCallRefused |
| function-length | TestFunctionTypeScriptLengthAgreesWithNode |
| erase-earlier-throw | TestOptionalAfterCallPreservesEarlierThrow |

Embedded mutant leaves: TestAssignmentProofChainedMutant, TestAssignmentProofCompoundMutant, TestAssignmentProofIfNarrowingMutant, TestAssignmentProofReturnDefinedMutant, TestAssignmentProofScannerKeywordMutant, TestAssignmentProofWiderTargetMutant, TestCallSpreadEvaluatedTwiceMutant, TestJSONStringifyShorthandWrongBindingMutant. Each passes by rejecting its deliberately changed source/output against its independently captured Node golden.
