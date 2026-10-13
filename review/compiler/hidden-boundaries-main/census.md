# Hidden boundary byte remeasurement

The stack reveals **20,245 additional bytes** in the ten assigned, disjoint hidden regions compared with main. Hidden bytes fall from **76,588 to 56,343**, with **zero newly hidden bytes**. This advances step 30. These are latent lowering measurements, not a claim that the complete TypeScript compiler now compiles.

The production stack is `f742bb83173dcd151abb8f4513985693112c1ea8`; the main reference fetched at the start is `68db8ddd145281a62655452496bdc32ef848bdf3`. The audit commit `a383e4d1` changes evidence only. No compiler source changed in this turn.

## Per-region results

Offsets and byte counts use the pinned adapted source, not today's reformatted source. Seconds are main / stack wall seconds for the completed measurement runs.

| Region | Source and byte interval | Main hidden | Stack hidden | Revealed | Newly hidden | Seconds |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| hidden-01-large | `transformers/declarations.ts` [68180, 81805) | 13,625 | 13,625 | 0 | 0 | 16.663 / 29.881 |
| hidden-01-small | `transformers/declarations.ts` [82928, 89827) | 6,899 | 6,899 | 0 | 0 | 16.598 / 28.866 |
| hidden-05-large | `transformers/es2018.ts` [35690, 41768) | 5,583 | 5,583 | 0 | 0 | 4.980 / 83.396 |
| hidden-05-small | `transformers/es2015.ts` [144178, 149926) | 5,420 | 5,420 | 0 | 0 | 3.986 / 4.237 |
| hidden-06 | `transformers/esDecorators.ts` [61324, 72741) | 11,417 | 1,020 | 10,397 | 0 | 2.472 / 13.434 |
| hidden-13 | `transformers/es2017.ts` [30237, 37526) | 7,289 | 6,965 | 324 | 0 | 2.230 / 42.512 |
| hidden-14 | `utilities.ts` [455532, 462634) | 7,102 | 5,008 | 2,094 | 0 | 2.280 / 29.044 |
| hidden-08 | `utilities.ts` [356194, 365984) | 9,790 | 2,360 | 7,430 | 0 | 12.628 / 14.340 |
| hidden-04-large | `checker.ts` [620304, 627479) | 5,441 | 5,441 | 0 | 0 | 13.769 / 15.700 |
| hidden-04-small | `checker.ts` [1623136, 1629022) | 4,022 | 4,022 | 0 | 0 | 21.347 / 30.280 |
| Total | Ten intervals | 76,588 | 56,343 | 20,245 | 0 | |

The seven overload regions include the TNode and receiver-independent method regions; overlapping work is counted once. The other three intervals are never-array lowering and the two optional-array caller regions. This report covers those assigned intervals, not every hidden interval in the historical corpus. In particular, the optional-array fixes add no further revealed bytes versus this main reference.

## Method and evidence

The stock census and adapted-source catalog are pinned to `388096e6a83a4e9d287fb827f793c599ba1bf0ad`. All 82 compiler files match that catalog's byte lengths and SHA-256 hashes. Both workers load and register the complete project. Each query attempts every independent AST unit intersecting its interval; a count and identity assertion checks the attempted units against the independent stock catalog. The original hidden census arithmetic unions blocked checker and dependency boundaries, subtracts independently covered intervals, and clips the result to the query. Revealed bytes are main-hidden minus stack-hidden; newly hidden bytes are the reverse difference.

The project remains checker-rejected. The latent census continues lowering independent units to measure covered and blocked bytes; it does not turn a checker rejection into an admission. [The final machine-readable results](census-evidence/memo/regions.json) retain the unit roster, hidden ranges, remaining reasons, source hashes, and checker-rejected status. [Provenance](census-evidence/provenance.json) records worker and overlay hashes; the exact overlay inputs are included as non-compilable evidence.

The original stack never-array query hit its 85-second process bound, then its separate 90-second retry, without emitting a ledger. Its bounded stack trace repeatedly showed `entriesProgramHasRecords` rescanning the project. For measurement only, both workers memoize this query per lowering instance. Inspection finds that it reads the fixed AST, checker, imports and module roots; that is the reason to infer that memoization preserves the result. The observed check is stronger: **all 19 completed original ledgers are identical as parsed JSON to their memoized counterparts**, including diagnostics and coverage. All 20 memoized queries complete, including never-array in 14.340 seconds. [The equivalence assertion](census-evidence/memo-equivalence.json) records this check. This memo is scratch overlay instrumentation and is not a production change.

Every command was bounded at 90 seconds; query children use an 85-second bound or the remaining batch bound. Go test invocations additionally use `-timeout 90s`. The box has `nproc=5` and a four-CPU cgroup quota. Setup completed in 36.139 seconds, including a 35.921-second Go build. Worker processes used GOMAXPROCS=2 during paired measurements; the standalone retry used 4. The 83.396-second census query is a bounded measurement command, not a newly added gate test.

## Remaining boundaries

The unchanged declaration and callback regions still encounter unsupported captures, generic function values, optional/union field reads with record storage, and failed overload result relations. Hidden-06 retains 1,020 bytes around unsupported immediate function calls, visitor covariance, modifier access, Map keys and object-literal assignment. Hidden-13 retains 6,965 bytes around visitor arguments, generic function values and unresolved visited values. Hidden-14 retains 5,008 bytes around an Expression brand relation, record-field reads, unchecked casts and object iteration. Hidden-08 retains 2,360 bytes around union/optional record-field reads and `target` access. Exact locations and reasons, including reasons in unchanged optional-array regions, are retained in the final JSON rather than treated as successful lowering.

## Checks that can fail

- The eight original interval-arithmetic tests pass. Five arithmetic mutants fail: ignore dependency boundaries, ignore skipped units, double-count overlaps, forget checker children, and forget independent coverage.
- Removing a unit from the attempted roster fails the catalog identity assertion.
- A memo mutant serving `false` instead of the actual record query changes the hidden-14 ledger and fails the equality assertion. This checks the memo's value, independently of the 19-ledger control comparison.
- The separate [counts audit](counts-audit.md) explains all nine existing retain/release changes and holds their output to Node in both backends, including sanitized native. Its wrong-ABI count mutant fails all nine count assertions.

Logs and mutant evidence are under [census-evidence](census-evidence/). `run-regions.py` records the exact bounded worker invocations. Recheck the ledger equivalence with `timeout 90 python3 review/compiler/hidden-boundaries-main/census-evidence/verify-memo.py`; recompute the table with `timeout 90 python3 review/compiler/hidden-boundaries-main/census-evidence/memo/summarize.py`. These scripts refer to the recorded source and scratch worker paths; rebuilding the workers requires the recorded overlays and source pin.

Lane checks pass: `lane checks 1.2 s: gofmt and tools on 44 Go files, t.Parallel on 3 test packages; vet 3 packages`. The first invocation lacked the toolchain PATH and stopped at missing gofmt; sourcing the setup environment corrected it.
