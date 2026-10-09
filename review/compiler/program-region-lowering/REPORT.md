Built step 06 compiler Program selection, allocation-site flags, immediate adoption and storage reuse exclusion on runtime core a3f28d97.
Implementation commits: e09a69d7 and 80d6b331; the following validation commit carries fixtures, counts and this report.
Targeted lower, native and oracle checks pass; all 1,045 existing recorded count entries are unchanged.
Both ruling selection mutants are caught by membership assertions; their runtime variants stay safe; reuse and short-size mutants are caught.
Full TypeScript native lowering, graph compiler fallback, asynchronous work and the full gate were not covered. K60 and K144 remain explicit disagreements.

Production lowering discovers concrete types and captures, selects members, then builds final allocations with ProgramRegion already set. There is no reflective rewrite of the production IR. The structural index filters candidates using required and offered property sets, then checks assignability only on that shortlist. Optional properties have a dedicated failing mutant. Only the owning-link adapter was ported from the prototype graph compiler; graph ownership and union-find were not imported.

Adoption precedes field initialization, constructor bodies, capture aliases and canonical cache publication. Emission uses the full allocated size, including optional presence/readiness storage. Dictionary-backed records on this base use their allocated object wrapper size. Member headers also block in-place mapping when the mapper output itself stays counted. Members are excluded from statement regions. Parallel boundaries are refused; asynchronous Program lowering is conservatively refused.

No runtime source files were changed. Two emitter slot-cache reads were adapted to this runtime's packed representation through adamic_slot_index, fixing the base's compiler/runtime mismatch. Lane checks required removing an inherited trailing blank line in each of internal/flow/corpus_units_test.go and internal/fresh/corpus_units_test.go; these are formatting changes only.

Census evidence uses pristine TypeScript 6.0.3, commit 050880ce59e30b356b686bd3144efe24f875ebc8, generated diagnostics, and inventory fcb7451a8239723fd903ae5998f54d2a86715e1f. Independently built checker runs both match the pinned membership ledger: 1,685 records, 1,599 members, 86 counted, zero unresolved container identities. There are 77 uninstantiated declarations awaiting concrete lowering. Selection sees 3,120 candidate types and selects 1,600 checker nodes, distinct from ledger records.

Selection before the shape index: 23.147740696 seconds. After: 11.766952292 seconds. These final independent runs shared CPU with other validation; whole test leaves took 30.95 and 19.23 seconds. Earlier same-checker warm measurements were 8.452791265 and 1.348071880 seconds. Logs preserve both measurements. The initially combined census test exceeded the leaf budget under contention, so the final tests are separate top-level leaves.

All fourteen formerly unresolved containers were reproduced:

| Records | Reading | Reason |
| --- | --- | --- |
| K24, K67, K85, K86, K109 | Member | Selected array element |
| K116, K117, K118 | Member | Selected Map value |
| K130, K131 | Member | Concrete selected Node key; generic value unknown; K131 private generic schema alpha-renamed |
| K90, K91 | Counted | No selected element/key/value |
| K60 | Selector member; recorded compiler expectation counted | readonly ProjectReference[] and its ProjectReference element are selected by the owning graph and structural relation |
| K144 | Selector member; recorded compiler expectation counted | Pending-emit union contains selected tuples and IncrementalBuildInfoFileId; its branded-number intersection exposes a selected object brand while plain number is scalar |

K60 and K144 expectations were not changed. census-split.log records the type flags and selections for both original and indexed runs. These observations reproduce the prototype's conservative structural overapproximation; they do not establish that either container requires membership semantically.

The six cycles mirrors and million.a agree with Node under ASan/UBSan, pass leak checking, and satisfy allocations = frees + regions. Million records 1,000,003 allocations, 3 frees, 1,000,000 regions. Explicit weak relation/symbol mirrors can remain fully counted. Ownership, optional-field constructor and canonical closure witnesses also pass. Flag off keeps cyclic strong fixtures refused and existing accepted fixtures unchanged. counts.md adds fifteen rows without modifying any old entry: 1,030 heap rows and fifteen predicate rows preserved. counts-comparison.txt and counts-final-check.log are the evidence.

Mutants:

| Mutation | Detector | Evidence |
| --- | --- | --- |
| Select acyclic leaf | Membership assertion: acyclic Leaf made a member | extra-leaf-mutant.log |
| Drop cyclic allocation flag | Membership assertion: recursive Node left counted | missing-member-mutant.log |
| Add acyclic leaf to actual emitted membership | Node unchanged, ASan/UBSan and leak clean, member count +1 | program-complete.log |
| Drop one actual cyclic member | Node unchanged, ASan/UBSan and leak clean, member count -1 | program-complete.log |
| Permit member mapper reuse | Actual source/result storage inequality fails with panic under sanitizer build; counted control reuses | program-complete.log |
| Adopt only declared field bytes | ASan heap-buffer-overflow reading optional tail | construction.log and program-complete.log |
| Treat optional property as required in index | Shape-index test fails | optional-shape-portable.log |

The ruling runtime mutations intentionally preserve safety; their selection mistakes are separately detected by inference tests. Reflection appears only in test code that constructs these deliberate mutations. Existing runtime dependency mutants also passed: premature free caught by ASan use-after-free, weak and cell-weak forgetting caught by target assertions, short adoption caught by ASan overflow.

Commands and results (all test output written directly to logs):

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
npm ci --prefix stage3/api
go test ./internal/native -run '^TestProgramRegion(Core|Target|CoreMutants)$' -count=1 -timeout 5m
go test ./internal/lower -run '^TestProgramRegion' -count=1 -v -timeout 5m
ADAMIC_PROGRAM_CENSUS_ROOT=/workspace/scratch/program-region-typescript go test ./internal/lower -run '^TestProgramRegionCensus' -count=1 -v -timeout 5m
go test ./internal/oracle -run '^TestProgramRegion' -count=1 -v -timeout 5m
go test ./internal/oracle -run '^TestNativeAgreesWithNode/(internal/oracle/testdata/)?(cycles_weak_|program_region_map_storage)' -count=1 -v -timeout 5m
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m
go vet ./internal/ir ./internal/lower ./internal/native ./internal/oracle
```

The three compiler inference/index overlays were run with go test -overlay=<review JSON> ./internal/lower -run <focused test> -count=1 -v, each returning the required failure. The portable optional overlay was rerun after strengthening its witness to prevent reverse assignability from masking the defect. Initial count validation failed on missing pinned Node declarations; npm ci resolved it. Initial constructor fixture used unsupported console argument forms; the corrected fixture passes. Earlier failed logs are retained alongside final evidence.

Setup measurements: Go ready 0.207s, Node ready 0.436s, clang ready 1.033s, Markdown ready 2.204s, submodules ready 25.803s, Go build ready 275.115s, cache warm 275.243s, done 275.283s. nproc=5, cgroup allocation four CPUs. Runtime core checks passed in 53.486s; final existing count verification passed in 197.920s; default fixture verification passed in 4.664s; vet passed. Each new focused leaf duration is recorded in the verbose logs; census leaves are below 60 seconds. CountsAreRecorded is an existing aggregate test.

Before delivery the committed branch runs integration's required lane command:

```sh
git fetch -q origin main devtools/fast-gate cloud/merge-tree
git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -
```

Its final output is saved in lane-final.log. Delivery stays on the named runtime core base and does not merge another worker's unlanded branch. No full package suite or full gate was run. The conservative choices are to refuse asynchronous Program use, preserve both census disagreements, and retain counted weak/container cases when the selector finds no selected ownership. No outstanding question was needed to implement this scope.

Final Program oracle leaf durations:

```text
--- PASS: TestProgramRegionCyclesWeakSymbols (0.90s)
--- PASS: TestProgramRegionCyclesGraphParent (1.05s)
--- PASS: TestProgramRegionMapperStorage (0.70s)
--- PASS: TestProgramRegionMapperCountedStorage (0.58s)
--- PASS: TestProgramRegionDisabledRelations (0.09s)
--- PASS: TestProgramRegionDisabledClosure (0.06s)
--- PASS: TestProgramRegionMissingCyclicMemberMutant (1.80s)
--- PASS: TestProgramRegionExtraLeafMutant (1.82s)
--- PASS: TestProgramRegionDisabledConstruction (0.10s)
--- PASS: TestProgramRegionDisabledOwnership (0.12s)
--- PASS: TestProgramRegionDisabledMillion (0.10s)
--- PASS: TestProgramRegionDisabledSymbols (0.08s)
--- PASS: TestProgramRegionDisabledParent (0.12s)
--- PASS: TestProgramRegionProgramRegionOwnership (0.74s)
--- PASS: TestProgramRegionShortObjectMutant (0.43s)
--- PASS: TestProgramRegionCyclesWeakParent (0.75s)
--- PASS: TestProgramRegionClosure (1.00s)
--- PASS: TestProgramRegionCyclesWeakRelations (0.74s)
--- PASS: TestProgramRegionCyclesGraphSymbols (0.93s)
--- PASS: TestProgramRegionCyclesGraphRelations (0.76s)
--- PASS: TestProgramRegionConstruction (0.72s)
--- PASS: TestProgramRegionGraphRegionsMillion (4.57s)
```
