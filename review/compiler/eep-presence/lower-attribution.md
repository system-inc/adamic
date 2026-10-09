# Historical lower failures

The isolated command at each checkout was `timeout 95 go test -buildvcs=false ./internal/lower -run '^NAME$' -timeout 90s -count=1 -v`. Authoritative rows and timings are in `followup/three-revisions-warm.json`; the corresponding `*-warm.log` files contain the complete output. Cold attempts in `three-revisions.json` include build failures from temporary-storage exhaustion and are not test verdicts.

| Test | 50654a40 | 897d0e79 | eep-presence 5f08a92c | First failing tested revision |
| --- | --- | --- | --- | --- |
| TestUndecidedCycleReadsUseReadyChecks | PASS | FAIL | FAIL | 897d0e79 |
| TestOptionalIndexingMapShapeRefused | PASS | FAIL | FAIL | 897d0e79 |
| TestOptionalIndexingKeepsUnsupportedStorageNotYet | PASS | FAIL | FAIL | 897d0e79 |

Observations: the cycle test on 897d0e79 expects exit 70 and an internal panic message; native instead returns the ruled uncaught-exception exit 1 with empty output. Four early-read subtests fail; initialized-read subtests pass. The Map-shape test gets successful lowering instead of its structural Map receiver refusal. The unsupported-storage test still gets located NotYet diagnostics, but two expected message substrings are stale.

History: `a6cb566048492864d497d20ba055c9351a82059e` (Close lowering chain reds and restore the structural Map boundary) contains all three corrections: the structural receiver guard in `internal/lower/optional_chain.go`, exception-contract reconciliation in `internal/lower/import_cycle_ready_test.go`, and updated diagnostic expectations in `internal/lower/optional_indexing_test.go`. It is an ancestor of 50654a40 and is absent from 897d0e79. The two requested historical tips are not in one linear ancestry: their merge base is 2391c655a4287fee747af74d31f5781e99be19e5.

Inference: these reds are inherited from the optional-presence-chain dependency at 897d0e79, rather than introduced by the eep-presence guard. “First failing” above means the first failing requested comparison point, not a bisected introduction commit. Per the unit instruction, this branch does not repair or merge the dependency. In particular, the missing structural Map boundary remains a dependency defect.
