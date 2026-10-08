Built: value-producer dynamic-key projections, named membership/mutation frontiers, source controls, and executable-array regressions.
Commits: step 1 918f3785; current main merged in 3e635421; latest views-integration 1d165a7c merged and published as 42c8a414; dynamic-key commit follows this report.
Commands and outputs: focused lower/native/IR checks, full IR/JavaScript, uncached selected Node oracle, ten adapter controls, 35 existing controls and vet PASS; independently audited census table follows.
Mutants: twelve valid Go projection mutants, six built adapter mutants, and one unsafe dynamic read-erasure mutant are caught by semantic assertions; two metadata mutants are caught by the lossless artifact hash; step 1 has three additional caught C adoption mutants.
Not covered: checker-ledger repairs, initialization of the production array-result temporary, precise indexed-write aliases, whole-array contracts, or a green complete gate.

## Scope and assumptions

The source is still the exact isolated `stage3/apply.sh` adaptation at 234ab1aa5f728a5221fb6075c35b94f88a2c6437, with the original 2,936 cast identities (1,758 tagged, 1,178 untagged). This is analysis of a checker-rejected program, not a compiler-admitted program. The loader and lowering refusals remain asserted by the census tool, and no executable source is emitted. New source controls are `.a`. No cohere source was copied or submodule pin changed.

Before and after use the same adapted source, site map and checker inventory. The initial before/after measurement binaries were built on the current-main/integration merge 3e635421; the optimized binary was built after 1d165a7c was integrated. A later views-integration update to 1d165a7c changed checked-read handling and controls, not the shape graph, eraser, IR shape metadata or latent adapter; the owned unit was re-greened after that merge. The rebuilt before run has exactly the same outcome and reason at all 2,936 identities as the initial before run.

Interpretation of “leave the diagnosed bucket”: do not repair diagnosed bodies or the checker ledger, and continue skipping their bodies. Following a newly visible allocation/key dependency can expose a diagnosed or host origin; that is evidence reclassification, not a checker repair. Unknown never proves nonconformance.

## New proof queries

Dynamic record and array reads preserve their receiver **and key producer**. The query follows actual string/number constants, local producers and conditional joins; it never trusts a declared, asserted or generic key type. Finite keys join every selected initializer and visible property store. Open keys join all modeled slots while retaining `dynamic key membership not proven`. Missing slots, spreads, opaque calls/stores, array mutation and depth exhaustion retain specific causes. Executable `ir.ArrayIndex` uses the same helper as the analysis-only source node.

Repeated immutable projection joins initially made the full after census expensive. A bounded cache includes receiver, key and recursion depth, so it never bypasses the depth obligation; exceeding the cache weight only skips storage. An indexed field map removes repeated linear scans and preserves every initializer, including duplicate field-name producers. Unit tests query twice, and a valid cache-corruption mutant drops the saved producers. A separate 20-second profile is retained in the logs. No new proof is inferred from a performance cache.

The old adapter represented every numeric literal as zero and every string literal with index zero because only allocation identities were needed. Dynamic keys require the actual values. Source controls selecting the wrong second slot pin both fixes.

Indexed assignment effects are not yet certified by receiver alias. A whole-program indexed-write marker conservatively retains projection and field-certificate checks, including direct casts after a possible indexed field write. It does not invent an allocation edge from every unrelated indexed store into every cast, nor classify unrelated diagnosed stores as actual dependencies. Precise alias/effect certificates remain future work.

## Census

| Outcome / Unknown cause | Before tagged | Before untagged | Before total | After tagged | After untagged | After total |
|---|---:|---:|---:|---:|---:|---:|
| Free | 0 | 0 | 0 | 0 | 0 | 0 |
| Readiness-only | 0 | 0 | 0 | 0 | 0 | 0 |
| Conforms-if | 0 | 0 | 0 | 0 | 0 | 0 |
| Unknown: unsupported flow | 258 | 248 | 506 | 192 | 202 | 394 |
| Unknown: diagnosed body/dependency | 1500 | 927 | 2427 | 1566 | 973 | 2539 |
| Unknown: host | 0 | 3 | 3 | 0 | 3 | 3 |
| Total | 1758 | 1178 | 2936 | 1758 | 1178 | 2936 |

Diagnostics remain 260. All 112 reclassifications are in the original dynamic cohort: 66 tagged and 46 untagged now expose diagnosed dependencies; no diagnosed body was repaired or emitted. Four tagged and ten untagged cohort sites remain unsupported-flow Unknown. All 126 remain Unknown overall, so this is visibility into the ledger boundary, not 126 newly erased casts. The table above is the independently audited first after run; the independently audited optimized run matches all 2,936 outcome/reason pairs.

The original dynamic-key cohort is exactly 126 overlapping unsupported-flow sites: 70 tagged, 56 untagged. The cohort is identified by the **before** rows, not by counting a renamed diagnostic in the after rows. The retained artifact records every original identity and its before/after classification, so unchanged Unknowns remain visible.

The first after run adds actual slot edges at every original dynamic-key frontier. The remaining 14 sites are nine in tracing.ts (247:43, 256:39, 266:41, 277:42, 286:43, 296:43, 321:63, 322:77, 324:62), two in factory/nodeFactory.ts (7047:42, 7051:43), two in checker.ts (36293:111, 36293:170), and one in commandLineParser.ts (3231:32). Indexed-store effects and opaque-call mutation remain common causes; tracing also has `dynamic projected slot absent or allocation not modeled: 0`. Factory sites retain escaped callback-argument frontiers. These are overlapping causes, not separate exclusive buckets.

## Controls and production boundary

Ten new adapter controls: finiteString and finiteNumber are Free; wrongString and wrongNumber are conforms-if with a `ready: number` versus boolean obligation; openKey and absentKey retain membership/slot causes; hostKey retains JSON.parse provenance; diagnosedKey retains the diagnosed dependency; indexedStore and directStore retain indexed-write effects. The original 35 controls still match their independent audit, including readiness-only, host, diagnosed, callbacks, generics, opaque property stores and whole-array cases.

Two production `.a` fixtures verify that finite numeric parameter producers establish every possible selected allocation. The array-result temporary is separately declared without a tracked initializer and assigned later. Its checked read still reports `missing expression` in allocation provenance, so the view and field reads remain checked. We did not erase that initialization frontier. The ready fixture agrees with source Node in native release, sanitized native, JavaScript and leak checking. The wrong fixture refuses its second `ready` field with the pinned named boolean check. Erasing that read guard yields successful wrong output: native `true\nfalse\n`, JavaScript `true\n0\n`; the assertion catches both.

## Every new mutant and its catcher

All mutants are built valid code, not compile-error sentinels. The scripts restore or overlay the production sources.

| Mutant | Semantic catcher |
|---|---|
| drop-cached-sources | second TestShapeDynamicKeys query loses its cached allocation producers |
| drop-production-array-edge | TestShapeDynamicKeys/production-array loses selected allocations |
| drop-opaque-slot-store | TestShapeDynamicKeys/opaque-store loses the opaque store allocation |
| ignore-projection-depth | TestShapeDynamicKeys/depth loses the exact named recursion limit |
| drop-key-arm | TestShapeDynamicKeys/finite loses one conditional slot |
| ignore-open-key | TestShapeDynamicKeys/open loses Unknown membership |
| ignore-absent-slot | TestShapeDynamicKeys/absent loses the missing-slot obligation |
| ignore-indexed-store | TestShapeDynamicKeys/mutation loses indexed effects |
| drop-slot-store | TestShapeDynamicKeys/store loses the later stored allocation |
| ignore-array-mutation | TestShapeDynamicKeys/array-mutation loses the mutation obligation |
| ignore-record-spread | TestShapeDynamicKeys/spread loses the spread obligation |
| drop-open-slots | TestShapeDynamicKeys/unbounded loses all modeled slot producers |
| zero-numeric-key | wrongNumber source control changes conforms-if to Free |
| zero-string-key | finiteString source control loses its actual selected keys |
| drop-key-provenance | hostKey loses the host reason (diagnosedKey also pins diagnosed origin) |
| ignore-indexed-assignment | indexedStore loses its indexed-effect Unknown |
| ignore-indexed-field-effects | directStore loses the field-certificate effect obligation |
| drop-dynamic-read | finiteString loses its Free proof |
| unsafe dynamic field-read erasure | TestShapeDynamicArrayExecution sees exit 0 with wrong stdout instead of the pinned check in both backends |

The artifact codec interns repeated diagnostic metadata without dropping any row, field, cause, schema or raw diagnostic. Its canonical JSON hash also preserves scalar types: `conflate-true-one` replaces true with integer 1 (Python equality alone would miss it), and `drop-array-item` removes a referenced item. Both valid metadata mutants change the reconstructed JSON hash and are caught. The complete raw censuses are independently audited; the lossless before and first-after artifacts are retained at /tmp/shape-dynamic-before-interned.json.gz and /tmp/shape-dynamic-after-interned.json.gz. The repository publishes all 2,936 before/after transition identities, the exact 126-site cohort, source hashes, and audit logs. Large full artifacts remain local because their GitHub-app transport stalled; they can be regenerated with artifacts.py.

Step 1's three additional mutants and their catchers are in GRAPH-FRONTIERS-REPORT.md.

## Commands and gate limits

```
source /workspace/adamic-tools/env.sh
python3 stage3/shape-conformance/latent/make-overlay.py /tmp/shape-dynamic-overlay
go build -buildvcs=false -overlay=/tmp/shape-dynamic-overlay/overlay.json -o /tmp/shape-dynamic-census ./stage3/shape-conformance/latent/tool
GOMEMLIMIT=3GiB GOGC=50 /tmp/shape-dynamic-census /tmp/shape-stage3-234ab1aa-adapted /tmp/shape-dynamic-map.json /tmp/shape-dynamic-after.json
python3 stage3/shape-conformance/latent/audit.py /tmp/shape-dynamic-after.json /tmp/shape-dynamic-map.json /tmp/shape-stage3-234ab1aa-adapted
python3 stage3/shape-conformance/dynamic-keys/controls.py /tmp/shape-dynamic-census
python3 stage3/shape-conformance/dynamic-key-mutants.py
python3 stage3/shape-conformance/dynamic-frontend-mutants.py
go test ./internal/lower ./internal/ir ./internal/native ./internal/javascript -run 'TestShape|TestAllocationFlow|TestClosureTargets|TestOptionalWidening' -count=1 -timeout 10m
go test ./internal/ir ./internal/javascript -count=1 -timeout 10m
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^(TestShapeGraphReportedFrontiers|TestCheckedViewShapeErasure|TestShapeErasureCountRows|TestShapeCallbackCountRows|TestShapeGenericCountRows|TestCheckedViewDictionaryComponents|TestCheckedViewNullishLiterals)$' -count=1 -v -timeout 10m
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestShapeDynamicArrayExecution$' -count=1 -v -timeout 10m
go vet ./internal/lower ./internal/ir ./internal/native ./internal/javascript ./internal/oracle
```

These focused commands pass; logs are retained under `logs/dynamic-*.log`. An earlier test correctly exposed the production temporary frontier; the final regression pins that conservative outcome instead of claiming erased reads.

The broad `go test ./internal/lower ./internal/ir ./internal/native ./internal/javascript -count=1 -timeout 15m` was stopped while native spent minutes in the unrelated exhaustive TestDecodeASCII. Full IR passed. Lower had seven failures, all reproduced with the pre-change shape_flow.go overlay: TestOverloadedShorthandFunctionValueStaysNotYet, TestCensusPredicateMarkerKeepsProofBoundaries, TestCensusOverloadRelation/result_covariance, TestNestedFunctionGapsAreLoud (optional/default/rest), TestPhantomArrayRequiredCastsAreErased, TestPhantomArrayCastsAreErased, and TestPhantomArrayProofs/cycle. Their expectations conflict with already integrated behavior; this unit did not rewrite them.

Additional native graph/map checks pass MapSmallProfile, MapSmallResidentMutant, MapSmallStorage, GraphUnreleasedAnchorCounted and the new adoption test. Five native graph harnesses remain red: old byte-count expectations in GraphRegionsMillion/LazyRegions/ContainerBoundary/RegionsRuntime, a stale two-argument closure C callback in GraphClosureEnvironment, and the RegionsRuntime leak harness expectation. The new object-size normalization correctly includes current readiness/representation/aligned contract storage; stale byte expectations were not hidden by rewriting shared harnesses. The original graph fixture count guard also remains red at the three named array-write frontiers documented in step 1. No green full-gate claim is made.

Toolchain timings and nproc are recorded in GRAPH-FRONTIERS-REPORT.md. Push through 3e635421 succeeded normally. A workspace restart later left the shell GitHub credential rejected; normal push/retry and the available gh helper failed. The connected GitHub app has repository push permission and published the exact verified merge tree (ca1e5dff4591eedae261ddc439802f85d715ac3b) as 42c8a414, with parents 3e635421 and 1d165a7c. The local branch was synchronized only after checking tree equality. No PR was opened and no main/area ref was moved.
