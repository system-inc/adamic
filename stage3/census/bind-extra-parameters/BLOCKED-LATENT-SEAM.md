# Stopped: latent unit results cannot currently feed faithful whole-program reach

The requested **latent-driver aggregate seam is not implemented or validated**. Its tsc results for both requested lists are **unknown**, not zero. Following the user's explicit instruction, work stops here rather than feeding a fabricated graph to the real check. This blocked result is published before 01:55Z on 2026-11-15.

## What was established through the latent unit path

A scratch-only unit-end hook was added to the unchanged full latent census overlay at head, before latentFinishUnit's optional-tail gate. It records each finished unit's actual local/function IDs, bind closure records, and FunctionTypeTargets population. It neither judges bind acceptance nor changes unit lowering. The head fixture was run with CommonJSBundle, LATENT_FULL=1 and both no-output guards.

The fixture has **9 attempted units**, **3 deferred generic units**, **9 captured unit ends**, and **1 recorded bind site** at main.ts:30:15. Every captured unit has **zero FunctionTypeTargets keys**. The required new reason is not produced through this input-capture path: **the required latent aggregate positive control has not passed**. The successful production-tail control elsewhere in this report is a different path and does not validate this missing bridge.

The captured snapshots demonstrate concrete incompatible namespaces:

| Snapshot unit | Function ID 1 | Local ID 4 | Local ID 5 |
| --- | --- | --- | --- |
| main.ts:29:1 (`narrow = seen`) | seen_value | shape (function 0) | index (function 0) |
| main.ts:30:1 (`bound = narrow.bind(...)`) | bind_factory | bound_value (function 1) | bound_value (function 1) |
| main.ts:31:1 (`found = find(...)`) | find_1 | array (function 1) | predicate (function 1) |

Function ID 1's parameter IDs are respectively `[6,7]`, `[4,5]`, and `[4,5,6]`. A naive union conflates three different functions and unrelated local producers/call arguments. Simply appending functions without rebasing every reference also points at the wrong functions. Splitting every snapshot's globals instead would lose the shared global assignment from `narrow = seen` to `narrow.bind(...)` and the subsequent call.

The actual production graph reads the coherent program's main statements, function bodies, local producers, call/return edges, parameter/optional-argument tables, exports and running-function membership (`view_flow.go:114` onward). The type-wide candidates are built by `resolveCountTypes` (`arguments_length.go:142` onward) from the whole program's callable types and **closureRecords**, including generic instantiations, structural method thunks and forwarding relationships. `FunctionTypeTargets` cannot be obtained by taking the empty unit-end maps or by guessing every source declaration is a runtime target. Object/array allocation identities and method tables also need consistent identities. No validated cross-snapshot linker/finalizer supplies these inputs in the existing overlay.

There is a second, independent loss of information: full mode **rolls back failed statements and returns no IR for them** (`full.go.txt:205–217`), and independently defers uninstantiated generic bodies. Unit snapshots begin separately at `units.go.txt:227`. In the completed head sys.ts full census, 136 units comprise **133 attempted and 3 deferred generic units**; the 93 boundary observations comprise **90 failed-statement rollback boundaries** and **3 signature/prologue boundaries**. These counts do not prove 90 distinct source statements, but they do prove discarded regions exist. Discarded calls, writes and bodies cannot be recovered from retained unit IR. The production projection index scans all main and function IR for opaque calls and mutations; the reach graph scans all writes and call edges. OR-ing observed opaqueCalls booleans and merging incomplete edges is not equivalent to that computation and can change known/unknown reach.

A new, validated IR linker plus an explicit sound treatment of missing bodies/edges would be required before calling this an all-unit production-equivalent measure. Neither was built here. This report does not assert that such a bridge is impossible in future; it states that the current overlay's retained results do not provide one, and this task has stopped as requested.

## Coverage and the two requested lists

- **(a) tsc sites where the head's new stop fires: UNKNOWN.** Zero tsc sites have a valid observation through the requested latent aggregate computation.
- **(b) tsc known-reach sites with extra declared parameters: UNKNOWN.** Zero tsc sites have a valid observation through the requested latent aggregate computation. This is not a count of zero wrong-answer shapes.
- **tsc units run with the unit-input collector: 0.** Only the fixture was used to establish the aggregation blocker. The old sys.ts run attempted 133 units, but omitted the check and captured no aggregate-ready input; those attempts do not count as coverage for either list.
- **Coverage of the requested 12 `.bind` plus 31 `maybeBind` source sites: 0 of 43 verified by the requested seam.** Both counts and the reach/fallback cause remain unmeasured for all 43. No source-to-adapted mapping has been established here.
- The separate syntax inventory in this branch describes **56 direct bind calls in the adapted tree**, after the existing `99-correctness-axolotl-x` maybeBind inlining adaptation, plus 2 calls to binder's unrelated local helper. These are not added to the user's 43 source sites or claimed as an equivalent inventory. They have no validated aggregate-seam coverage. Static classification is being performed separately on `typescript/bind-sites-static`.
- **Original latent sys pair:** byte-identical findings, with the original unproven-zero caveat preserved.
- **Archived production-tail seam:** the fixture passes (head 1 new stop; base no bind-check stop), but the 7 completed tsc root pairs all stop at sys.ts:1385:31 before the seam. These are reachability fallback attempts, not latent all-unit coverage. No further production-root attempt ran after the stop request; sys.ts was not started on that path.

## Receipts and implementation

`UNIT-INPUTS.json` records fixture counts and collision summaries. The raw per-unit captures, latent stream, logs and collision details remain in scratch under `results/unit-input-control/`, identified by SHA-256 and byte size in `receipts.json`. The exact diagnostic hook and its overlay builder are in `instrumentation/unit_inputs.go.txt` and `instrumentation/build-unit-inputs.py`. The normal stream preserves zero findings; the hook has not added a judgment or claimed a positive control.

No compiler or adaptation source in either measured checkout was edited. No changes were pushed outside the report branch. No latent census, production-root attempt, build or seam run remains active after this stop.
