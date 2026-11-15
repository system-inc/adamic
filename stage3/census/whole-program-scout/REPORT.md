# Whole-program census scout — et209ny

## Status and scope

**BLOCKED for the full census.** Individual tail controls pass after an earlier scan refusal. The real sys.ts program reaches the declaration/body phase but cannot reach the whole-program tail. The scout identifies the recovery seam; it does not claim an all-stop implementation.

Worktree base: main `763d5fc83` (fresh detached worktree, then scout branch). No production source is edited. The prototype is a no-output Go overlay. It brings in Language's exact `function_bind.go` from `ba619e28` and Memory's `cycles.go`, `graph_records.go`, `records.go`, plus the `graphRecords` lowering field from `b529db90`. These are independent rule heads, not a claim that either is merged into main. The AHRA task path is unavailable in this Linux environment; the supplied brief is the task definition. Both blocked census reports were read from their fetched commits.

This is a bounded continuation probe, **not an all-stop lowering implementation**. It uses production lowering for one entry and its complete import closure, records refusal-scan findings and continues the scan, then runs the production declaration/body/tail pipeline until its first subsequent stop. No unit snapshots are unioned. No discarded statement is treated as a complete body. No IR can leave either public lowering API in the overlay. Stops and completion are measurement results, never compilation acceptance.

## One-page pipeline map

All file:line references below name base main unless a head is explicit.

| Order | Production seam and behavior |
| --- | --- |
| Entry | `internal/lower/lower.go:52` Lower delegates to LowerWithOptions (`:56`), which requires **one** entry, acquires one checker, optionally does ProgramRegion discovery (`:69`) before the actual lowering (`:76`). `lowerCheckedWithCycleTypes` (`:93`) retries with all view-element reads checked on key misses; relation overflow is a final stop (`:105`). |
| Preparation | `lower.go:111` constructs one lowering/IR; deferred `hostMemberPending` (`:114`) can substitute an earlier deferred host-write stop. `modules.go:16` computes ESM import order/cycle read proofs; dynamic import gaps stop before declarations. `lower.go:129–140`: collect viewed elements, inheritance, enum initialization, accessor names, omitted optionals, namespace initialization. Any preparation error aborts. |
| Refusal scan | `refusals.go:35`: directives/pragmas, reference/parallel/merged-field/function-annotation/false-predicate preflights, call contracts (`:78`), then the ordinal rule registry (`:104`). First rule error stops the visitor; its result returns to `lower.go:142`, which aborts before all declaration/body lowering. Deferred refusal is returned only if no prior one (`refusals.go:119`). `refuseNode` (`:129`) is a separate subtree check, not automatically made recoverable by changing the module scan. |
| Declaration | `modules.go:109` flattens namespace declarations, registers globals/classes/functions in a shared namespace; generics register templates and instantiate on demand. All ordinary signatures precede bodies; static instances follow signatures, then function bodies. Errors at **registration, representation, signature or prologue** stop the entire program, not just a statement. `functions.go:48` lowerFunction ensures a signature (`:64`), consumes pending signature state, then enters lowerBody (`:291`). |
| Bodies | `functions.go:291`: branded-call judgment, saved function/this/index, parameter defaults/destructuring/constructor stores, block statements or expression return, restore outer state, finish nested environment, commit by function index, generator-body check. `statements.go:13` first lowers nested declarations, then each statement; its first error returns no statement list. `source_identity.go:17` moduleStatements judges calls, lowers statements and records the module boundary only on success. `lower.go:162` also aborts on the sticky `unlowerable` field. |
| Tail, part 1 | `lower.go:165–207`, in exact order: checkOverloadResults; prepend forwarder values; finishClassCalls; certifySourceSlots; finishAccessors; finishToPrimitive; checkIterableViews; exceptions; checkLibrarySpreads; checkAccessorSpreads; checkRecordGuardedReads; checkDynamicKeyedGetters; checkGeneratorRoads; flow.NormalizeAsync. These include transformations, not just diagnostics: accessors create dispatchers, iterable views rewrite generator stepping, exceptions computes throwing fixed points (`exceptions.go:218`), async normalization moves bodies (`internal/flow/async_normalize.go:13`). Async entry source/module boundaries are then repaired (`lower.go:208`). |
| Tail, part 2 | `lower.go:215–245`: **findCycles**, **checkBoundArguments**, checkViewOptionalWrites, checkOptionalMemberPresence, checkViewCopies, checkViewWidenedReads, checkHostMemberWrites, first asyncPayloadFailure, checkPromiseConsumption, checkViewLogicalWrites. `cycles.go:101` discovers shapes/types and concrete instantiations; graphTypes and JSON reference reach must succeed **before** resolveCountTypes (`:159`) fills FunctionTypeTargets and ABI argument slots. It then checks host ABIs and lazy views. `graph_types.go:375` returns the saved record stop; head `graph_records.go:17` supplies it at a seed. Head `function_bind.go:450` reads the coherent closure records and targets; boundValueReach uses allocation flow. `shape_flow.go:254` scans **all** Main/function IR, including opaque calls (`:272`); a missing call changes the answer. |
| Finalization | `lower.go:247–304`: ProgramRegion async restriction; applyProgramMetadata if a plan exists; optional ownership-transfer proof; readiness; borrow; counters; voidDiscards; checkInspectorDoomedContracts; checkFlowLiveness. Only then return the IR. Every fallible tail call aborts on its first error; none has a general recorder/continue contract. |
| Latent recovery | `stage3/census/latent/make_overlay.py:18` removes LowerWithOptions **and the production orchestration** through lowerChecked. `full.go.txt:186` wraps a statement in a deep checkpoint; errors, panic and sticky unlowerable are recorded, `*l = *checkpoint` restores state (`:214`), and both IR and error become nil (`:216`). `units.go.txt:142` prepares a file base, `:227` deep-copies it independently for each unit. Thus unrelated units reuse IDs; even perfect rebasing cannot recover the failed statement's lost producers. Generic units can be deferred (`:207`). |

## Designs and costs

Coverage means observed production-pass behavior, not a source inventory. No design has a measured all-tsc fraction yet. Attempting all roots and proving every program-wide answer are separate quantities.

| Design | Cost and concrete changes | Potential coverage; blind spots |
| --- | --- | --- |
| **A. Overlay-only coherent lowering with placeholders** | No production edit, but rewrites copies of `lower.go`, `refusals.go`, `modules.go`, `functions.go`, `statements.go`, probably `locals.go`/representation and the affected tail passes. Use production registration and one mutable lowering; recover a statement or function *in place*. Reuse the typed state copier, not shallow copies. Reserve/tombstone failed allocations rather than accidentally aliasing subsequent IDs. Add a no-output API/driver. Maintenance cost is repeated AST-aware rewrites and drift checks on hot files; not merely a change to latentFinishUnit. | Can visit all 79 roots, potentially the full import union, but only successfully lowered regions retain exact facts. A throw body carries no source writes, producers, rest/count observation, captures, generic instantiations or hidden opaque calls. Declaration/signature failures precede any replaceable body. Analysis must explicitly mark those gaps unknown; a throwing body alone cannot certify clean bind reach or graph freshness. Deep checkpoints per statement can approach quadratic copying as the coherent IR grows. |
| **B. Production recorder + continue hook, default off** | `lower.go` is hot and needs option plumbing through ProgramRegion and view-key retry. `refusals.go` needs all-stop registry/preflight reporting with stable rule ID/site/words; `modules.go` needs independently recoverable registration/signature/body boundaries; `functions.go` needs signed-body recovery; `census_small.go` needs per-overload recording rather than a first-error loop, and `host_member_write.go` must expose deferred and triggering diagnoses separately; `statements.go` needs transactional recovery and sticky-error handling. A dedicated recovery file owns the log, poisoned identities and checkpoints. `cycles.go`/`graph_types.go` and each later validating pass need a defined continue contract. The off path should do no copying and only test for an absent recorder at boundaries; the on path pays for journals/checkpoints. The IR/flow layer needs explicit unknown producers/effects, unless all answers after a gap are labeled conditional. `internal/lowered` must reject census output/caching: its options are automatically formatted into keys (`lowered.go:160`), so a recorder callback must not become pointer-dependent cache identity. Prefer a separate measurement result type over returning usable `*ir.Program`. | The only maintainable route to 79/79 attempted programs with stable production identities and explicit unknowns. Exact coverage remains the subset whose answer does not depend on missing facts. An option alone supplies **no** missing-body semantics; it is not a shortcut to 100% reliable answers. It also needs an explicit entry-set contract: production accepts one root. An import-union entry changes exports/liveness and is not silently equivalent to passing all 79 roots. |
| **C. Scan-continuation only (this cheapest probe)** | Rewrite the existing refusal visitor; keep production declaration/body lowering and its aborts. Overlay the two rule heads. No snapshots or placeholders. Six existing production file paths are replaced, plus a measurement-only helper and tool. This is a small executable test of the shared-ID route, not a solution to unsupported bodies. | Exact coherent IR for a root *if* its bodies lower; control tail checks are meaningful after a scan finding. Stops before or within bodies still prevent the tail. Measured tsc coverage below; unobserved tail answers are UNKNOWN. |
| **D. Link latent snapshots or source-only approximation** | A new linker must rebase every IR reference, reconcile shared globals, call/return edges, allocation sites, method tables, exports, and instantiations. It still needs source recovery of everything rollback discarded. Static type/source scans avoid this cost but answer a different conditional question. | Cannot produce the requested exhaustive production census from existing retained snapshots. Native seed eligibility, opaque calls, omitted producers and concrete generics remain unknown. Existing blocked reports already demonstrate this loss. |

A/B must also distinguish attempts: ProgramRegion discovery and view-key retries can rebuild the same source sites, and recording must stay outside state rollback while retaining which attempt produced each observation. Event counts must not silently become distinct-site counts.

B needs mutants for: recorder disabled/first-error restored; recovery retaining poisoned state; missing-body unknown effects omitted; failed IDs reused; tail accidentally skipped; count-target construction skipped after a cycle stop; source-generic bodies mistaken for instantiated ones; default-on continuation or usable IR leaking to emission/cache. Both clean twins must remain clean, and controls with failed *bodies/signatures*, not just scan stops, are required before claiming A or B works.

### Exact seam that needs Lower's ownership

`lowerWithElementKey` (`lower.go:111`) is the coherent-program seam. Recovery cannot be installed only at `latentFinishUnit`. It must surround declaration/signature/body operations **before** `declareModule` returns at `lower.go:151`, retain the production namespace, and carry explicit missing-effect state into the tail. Then split `findCycles`' fact construction/finalization from its first-error validation: today `finder.graphTypes` can return a record stop before `resolveCountTypes` ever runs. Simply changing `return nil, err` to record-and-continue makes the following bind pass read an unfinished target table. Each mutating pass needs either a completed-result contract, a transaction, or an explicit blocked-dependent-pass result.

This scout stops short of inventing that contract. A production hook is recommended for durable upkeep; a sufficiently large overlay could implement the same semantics, but a throw placeholder alone does not meet the census goal.

## Controls and sys.ts trial

The executable receipts are archived alongside this report. All fixtures contain an earlier `// @ts-check` refusal (checking stays enabled). The bind fixture is Language's b01 (`narrow_type.a`) as TypeScript, with only that leading pragma; its twin removes `seen`'s extra declared index parameter and prints a constant. Record controls are the field-record positive and array-only clean case from `b529db90`'s oracle fixtures, with the same leading pragma. The mutant restores production's first-error refusal scan. Each runs in a fresh process. Public Lower and LowerWithOptions are checked to return nil IR and an error before the measurement call.

| Control | Scan stops | Target tail stop | Tail completed | First-error mutant |
| --- | ---: | --- | --- | --- |
| b01 / bind-stop | 1 | `checkBoundArguments`: declares 2 parameters; type names 1, line 31:15 | no, target stop | target disappears, aborts at scan |
| bind-clean | 1 | absent | yes | aborts at scan |
| record-stop | 1 | `findCycles` / `noteRecordSlot`, line 6:1 | no, target stop | target disappears, aborts at scan |
| record-clean | 1 | absent | yes | aborts at scan |

All eight assertions passed; raw results, per-process wall time and peak RSS are in `evidence/`. The zero Functions/Locals fields on abort are unset report defaults, **not** a statement that the program had no functions or locals. `Completed` means the pipeline completed over IR after reporting scan findings, not that the compiler accepts the source. The prototype does not continue from the tail stop: the record control does not demonstrate bind after a cycle failure.

The four **combined** programs contain b01 and the record case in the same module, with one earlier scan pragma. Record-positive + bind-positive reports only the record stop; record-clean + bind-positive reports the bind stop; record-positive + bind-clean reports the record stop; both clean completes. `combined.py` asserts these four observations. Thus the stronger **both positive stops in one continued program** control is **BLOCKED**, explicitly not passed. This exercises the exact `findCycles` → `resolveCountTypes` → `checkBoundArguments` dependency described above.

The extra `signature-boundary.ts` inserts an unsupported bigint return signature before b01. It records the scan pragma, then `declareModule` stops at line 2:10, “a function returning bigint”; the bind stop is unobserved. This is an executable example of why body replacement cannot recover signature failures.

### sys.ts: blocked before bodies and tail

Ordinary checked Load succeeded. The final instrumented run completed normally as a **measurement**, with `Completed: false`: **272.6745 s wall**, **1,912,264 KiB peak RSS (1.824 GiB)**. The initial run took **274.7572 s**, **1,953,108 KiB (1.863 GiB)**. Both used the same adapted tree and produced the same **5,542 recorded events by rule**. The second run adds a read-only capture of the error *before* `hostMemberPending` substitutes the returned diagnosis; it does not bypass the overload check.

The **underlying abort** is `declareModule` at **`src/compiler/core.ts:261:1`**:

> overload 1 of filter result U[] cannot be served by implementation result readonly T[] | undefined

`declareModule`'s `censusOverloads(statement)` call is in the registration loop (`modules.go:195`), before its signature/body loops; the validator is `census_small.go:137`, with the result refusal at `:274`. This is a declaration/overload boundary, so replacing a failed body with `throw` cannot get past it. The **returned production diagnosis** instead names the first deferred host write, `moduleNameResolver.ts:1356:5`, “a method read as a value (getOrCreateCacheForModuleName would lose its object, and this with it)”. `host_member_write.go:112` / `lower.go:114` explain that substitution. `UnderlyingAbort` in the final receipt preserves the actual blocker separately; it is not double-counted as a scan event.

**Coverage:** 1 of 79 root programs attempted; **0 of 79** completed through the whole-program tail. The other 78 root programs were not run. The adapted source tree has 79 compiler files, and the one sys.ts import-closure scan reaches findings across that tree; this is not 79 independent root runs. `findCycles`, `noteRecordSlot` and `checkBoundArguments` outcomes for real tsc remain **UNKNOWN, not zero**. No all-tsc fraction of reliable bind/record judgments is established.

The raw `Stops` list has **5,541 scan events plus one substituted return diagnosis**. There are **4,026 unique (rule, site, words) triples** in that list; event counts are not distinct source sites or units. Counts below are the raw recorded rule labels. The 23 `lowering.scoutRefuse` events are predicate-argument contract scan stops (the outer scan label is retained for them); registry labels name the actual registry rule. Preflight/validator routines may themselves still return their first error, so this is not an exhaustive count of every possible refusal.

| Recorded rule / boundary | Events |
| --- | ---: |
| `refusal_scan/castRefusal` | 2532 |
| `refusal_scan/stringWideningRefusal` | 1055 |
| `refusal_scan/enumRefusal` | 962 |
| `refusal_scan/constraintRelationRefusal` | 454 |
| `refusal_scan/refuseOptionalWidening` | 233 |
| `refusal_scan/methodValueRefusal` | 151 |
| `refusal_scan/recordStorageView` | 24 |
| `refusal_scan/refusePromiseIntoVoid` | 24 |
| `lowering.scoutRefuse` | 23 |
| `refusal_scan/predicateArgumentsRefusal` | 23 |
| `refusal_scan/classViewRefusal` | 18 |
| `refusal_scan/narrowedNullishReadGap` | 11 |
| `refusal_scan/unionSlotWriteRefusal` | 5 |
| `refusal_scan/classDeadZoneGap` | 4 |
| `refusal_scan/refuseVoidRead` | 3 |
| `refusal_scan/refuseVoidWidening` | 3 |
| `refusal_scan/enumerationDescriptorRefusal` | 3 |
| `refusal_scan/boundObservation` | 3 |
| `refusal_scan/enumerationIndexedWriteRefusal` | 2 |
| `refusal_scan/narrowedNeverReadGap` | 2 |
| `refusal_scan/nodeLibraryRefusal` | 2 |
| `refusal_scan/refuseAnyMember` | 1 |
| `refusal_scan/phantomRefusal` | 1 |
| `refusal_scan/deleteRefusal` | 1 |
| `refusal_scan/namespaceRefusal` | 1 |
| `lowering.declareModule` | 1 |

The full site and exact words of each event are in `evidence/sys-final.json.gz`; the resource receipt is `evidence/sys-final.time.json`. The final underlying overload error is stored separately in that JSON. No new source adaptation, checker-option relaxation, fake type target, or fabricated missing body was used to pass this boundary. This is the scout's stopping reason: Lower needs a recoverable declaration/signature/body contract and a fact-finalization/validation split before this can become the requested all-stop census.

## Reproduction and review boundary

Run from this branch's worktree, with fresh scratch paths. `heads/*.go.txt` are exact archived sources of the two heads, not files Go discovers as production code. The builder uses the existing tested AST refusal rewriter, removes its per-unit checker-diagnostic exclusions (this probe uses ordinary checked Load), and emits only an overlay outside the repository. It asserts its critical textual seams before rewriting. The archived-head builder reproduced **all tested overlay source bytes identically**. `SCOUT_MUTANT_FIRST_ERROR=1` is the negative control, not a production setting.

```sh
export PATH=/workspace/adamic-tools/go/bin:$PATH
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3
export GOTMPDIR=/workspace/scratch/whole-program-state/tmp
export TMPDIR="$GOTMPDIR"
export GOCACHE=/workspace/scratch/whole-program-state/gocache
export ADAMIC_BUILD_CACHE_DIR=/workspace/scratch/whole-program-state/build-cache
# Initial setup: detached worktree from origin/main, then the scout branch;
# git submodule update --init --recursive --depth 1
# clone TypeScript --branch v6.0.3 --depth 1; verify source.json's full SHA
# npm ci in stage3/api (locked install succeeded: 3 packages)
# stage3/apply.sh /workspace/scratch/whole-program-state/adapted
python3 stage3/census/whole-program-scout/build.py /workspace/scratch/whole-program-state/overlay
go build -overlay /workspace/scratch/whole-program-state/overlay/overlay.json \
  -o /workspace/scratch/whole-program-state/scout \
  ./stage3/census/whole-program-scout/tool/main.go
python3 stage3/census/whole-program-scout/controls.py \
  /workspace/scratch/whole-program-state/scout /workspace/scratch/whole-program-state/controls
python3 stage3/census/whole-program-scout/combined.py \
  /workspace/scratch/whole-program-state/scout /workspace/scratch/whole-program-state/combined
GOMEMLIMIT=10GiB python3 stage3/census/whole-program-scout/measure.py \
  /workspace/scratch/whole-program-state/sys \
  /workspace/scratch/whole-program-state/scout \
  /workspace/scratch/whole-program-state/adapted/src/compiler/sys.ts
```

Setup used Go 1.27.1 and the submodule pins in `evidence/provenance.json`. All Go temporary/cache paths were in scratch. The adapted tree has 79 compiler TypeScript files; their hashes, the generated patch table and compressed apply log are archived. Resource measurements include loading/checking and lowering in the Go child; they exclude adaptation, dependency installation and the cold Go build. RSS is Linux `wait4`'s per-child peak in KiB, not a sampled heap estimate. No GNU time was installed, so `measure.py` collects the kernel's resource usage directly.

The measurement globals support one invocation per process, which the runner enforces; this is not a concurrent production API. The probe retains only the first error from each preflight/tail validator and its production abort. It does **not** claim all individual stops within those validators, continuation after a body error, or completion of dependent passes after an incomplete transform. There is no new emitter or compiler acceptance path, no production edits, no merged change, and no PR. Only the scout directory is committed on the authorized branch.
