# Cast tag readonly investigation: stopped, no landing commits

Task #b5w3ycg, 2026-11-16. Nothing was submitted to Loom. All changes and observations in this directory are evidence, not a landed adaptation.

The bounded (c) trial fails two acceptance conditions:

1. Both TypePredicate casts still fail the unchanged production `castProof` after all four member kinds become readonly. The first rejected writable field is **parameterName**, not type: `string | undefined` through `TypePredicate` does not fit `AssertsIdentifierTypePredicate.parameterName: string`. The guard returns there before examining type. Neither parameterName, parameterIndex, type nor TypePredicateBase was widened or made readonly.
2. The complete generators census exposes a **new lowering NotYet** at `generators.ts:2257:9`, `exception.catchVariable = name`: “writing a checked object field that may hold undefined or null, whose checked write can't change the representation its slot holds”. It is attributed to the independently attempted `endExceptionBlock` unit at `2285:5`. There is also one changed boundary at `2287:9`, whose text changes from the cast refusal to this write failure.

The requested stop on a newly caused finding applies. No landing adapter, ledger row, API sanction, shard-plan hash or test count was changed. Neither `typescript/cast-tags-readonly-1` nor `typescript/mutable-keeps-kind-1` was created or pushed by this investigation. (b) was scouted but not implemented or stacked on the failed (c) trial.

## Pins and setup

- Adamic base: `4a434af87754db2b710ccedaff64d21872edcb39`, fresh detached worktree `/workspace/scratch/cast-tags/work`; the shared checkout's source was not edited.
- Upstream TypeScript v6.0.3: `050880ce59e30b356b686bd3144efe24f875ebc8`, depth-one clone `/workspace/scratch/cast-tags/typescript`.
- Recursive pinned submodules initialized. Stage3 API and upstream locked npm installs completed. TMPDIR, GOTMPDIR and Go cache were scratch paths; subsequent product-cache paths were also scratch paths.
- A runtime interruption killed the original baseline apply and queued processes after `99-correctness-core-multimap`. Its complete whole-tree ref `2d3b45993697c021a9bf08554bbadf7de199aac1` survived. Only the scratch adapted checkout was restored to that ref; remaining unmodified adapters completed. Final adapted **tree** (not commit): `7a63782ed07eb5a7c37d1721028601ccf8757d80`. See [resume provenance](main-resume-status.json), [saved refs](completed-apply-refs.json) and the compressed original/resume logs.
- The final source trial is only five CodeBlock and four TypePredicate interface `kind` modifiers. [c-probe-owners.json](c-probe-owners.json) preserves each complete before/after owner and both SHA-256 hashes. [probe-c.cjs](probe-c.cjs) is an exploratory source trial, not the requested landing adapter contract.

Every adaptation's owners.json, targets.json and adapt.cjs was searched, including nested targets manifests. [pin-scout.json](pin-scout.json) and [pin-scout.log](pin-scout.log) retain all hits.

| Owner | Scout result |
| --- | --- |
| Five CodeBlock interfaces | No whole-interface pin. They are top-level, outside receiver-transformers' whole `transformGenerators` pin. A new (c) directory would be appropriate if proof passed. |
| Four TypePredicate member interfaces and TypePredicateBase | No whole-interface pin. |
| Mutable | Actually declared beside other internal helpers in **utilities.ts:10339**, not types.ts in v6.0.3. No whole-owner pin. It was not changed. |
| createBinder / bindContainer | No whole-owner pin. The 90-void-widening hits concern other declarations/expressions; those do not block a new internal helper or a bindContainer edit. |
| createNodeFactory | **Whole owner pinned** in `91-patcher-discard/proofs/targets.json`; a (b) edit inside it must amend that adaptation. Before SHA-256 `553d211c0d455a6badb5bf74a6555a8bd45189d259ba1b7276d7ac30e24c10a0`; after `5d4de2d08b291cb02f8abb4cbc7a2a6aca135e07ba14a5a85e30aad7b0734104`. No amendment was made. |
| createSyntacticTypeNodeBuilder / tryReuseExistingTypeNode / visitExistingNodeTreeSymbolsWorker | No whole-owner pin. |

## Per-site scout

Requested line labels are retained. Existing adaptations at this exact pin move the generator casts by -1 line and the checker casts by +3 lines. [site-scout.jsonl](site-scout.jsonl) preserves source excerpts; the actual diagnostic positions below come from the completed census or isolated production proof.

| (c) requested site | Measured site | Actual use/writes | Trial result |
| --- | --- | --- | --- |
| generators 2201 | 2200:23, endWithBlock | Reads endLabel; no field write through the cast | Cast refusal gone in refusal_scan and lowering |
| generators 2249 | 2248:27, beginCatchBlock | Writes state, catchVariable, catchLabel; reads endLabel | Cast refusal remains in refusal_scan and lowering |
| generators 2271 | 2270:27, beginFinallyBlock | Writes state, finallyLabel; reads endLabel | Cast refusal remains in refusal_scan and lowering |
| generators 2288 | 2287:27, endExceptionBlock | Reads state/endLabel; writes state=Done | Cast refusal gone; enables the new cross-function checked-field write finding at 2257:9 |
| generators 2342 | 2341:23, endLoopBlock | Reads breakLabel/isScript; no field write through the cast | Cast refusal gone in both phases |
| generators 2381 | 2380:23, endSwitchBlock | Reads breakLabel/isScript; no field write through the cast | Cast refusal gone in both phases |
| generators 2409 | 2408:23, endLabeledBlock | Reads isScript/breakLabel; no field write through the cast | Cast refusal gone in both phases |
| checker 22077 | 22080:44, compareTypePredicateRelatedTo | Reads parameterIndex; no field write | Refused before and after; guarded replay plus isolated production proof |
| checker 22079 | 22082:127, compareTypePredicateRelatedTo | Reads parameterName; no field write | Refused before and after in isolated production proof |

The two forward-looking peekBlock casts remain refusals at this pinned tree. No repair to peekBlock or its optional result was attempted. No listed site writes kind itself. CodeBlock kind assignments occur in fresh object construction.

The ordinary guarded checker replay reaches the first cast, then records a failed-statement boundary covering the second. It does **not** prove an independently reached second-site finding. To measure both, the scratch overlay calls the unchanged production `castProof` on the two actual checker AST nodes, with the real 79-root checker program. It retains both no-output guards and asserts zero checker diagnostics. These results are labeled `isolated_cast_proof`, not full-census findings. Logging added at the original `castUnionWrites` failure branch preserves its return decision. See [c-isolated-casts.json](c-isolated-casts.json), [its trace](c-isolated-casts.log), and the corresponding main files. All Go probe sources are `.go.txt`; no production Go source was edited.

| (b) site | Actual use/writes | Next-stop measurement |
| --- | --- | --- |
| binder 984 | Assigns container/blockScopeContainer references; subsequent narrowed container writes locals. bindContainer's mutable parameter also supports flags and endFlowNode writes. | Not run after (c) stop |
| binder 991 | Assigns blockScopeContainer reference; subsequent narrowed view clears locals. | Not run |
| binder 1008 | Reads asteriskToken; no field write through this cast. | Not run |
| binder 1016 | Stores node reference in currentFlow.node; no mutation of node through this cast. | Not run |
| binder 1031 | Reads body; adjacent direct node writes update flags. | Not run |
| binder 1034 | Writes endFlowNode. | Not run |
| nodeFactory 2560 | Return cast in createUnionTypeNode. Shared createUnionOrIntersectionTypeNode writes types and transformFlags; this return cast writes no field. | Not run |
| nodeFactory 2570 | Same for createIntersectionTypeNode. | Not run |
| expressionToTypeNode 469 | Writes modifiers=undefined through Mutable<ParameterDeclaration>. The preceding line 467 writes type through Mutable<typeof visited>. | Not run |

Thus **all nine (b) next stops are unmeasured**. There is no measured count of id/number-or-undefined NotYets, and no claim of zero. A distributing internal helper was not introduced before (c)'s stop was resolved.

## Completed proof and limits

| Proof | Observation |
| --- | --- |
| Complete generators single-file census | LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, exact LATENT_ONLY_FILE, CommonJSBundle. Both complete header/file/relation-trailer streams retained. 523 -> 498 raw finding occurrences. **1 new, 26 gone, 1 changed** after normalizing only tree prefixes and byte offsets shifted by added readonly tokens. This fails the zero-new/zero-changed gate. |
| Every generator delta | [census-delta-full.json](census-delta-full.json) lists every new/gone/changed finding in full, including phase and attempting unit; [summary](census-delta-summary.json) gives phase/kind totals. Boundaries are included, not hidden. |
| Full checker census | Started in background on both trees. Stopped after the independently measured TypePredicate gate failure, preserving headers and external CPU/output/status checkpoints. Main ran 714.05s; trial 391.05s before SIGTERM. **Incomplete; no full checker comparison or zero-new claim.** The tool's internal progress cursor exists in speculative mode only, so external checkpoints were retained explicitly. |
| Stock TypeScript 6.0.3, 79 roots | 64 diagnostics on each tree, identical, **0 new and 0 changed**. [Summary](stock-diagnostics-summary.json), complete diagnostics compressed. The first harness incorrectly asserted absolute zero after the identical comparison; its failed assertion log is retained. Only that overstrict assertion was corrected to the requested zero-new criterion; no compiler options or source were changed. |
| Production adamic types, 79 sorted roots | Main and trial both exit 0; both stderr files empty. [Status](adamic-types-status.json); complete stdout retained compressed. |
| JavaScript identity | **604 source files; 1,208 paired emits**, comments kept/removed, byte-identical. Only two source files differ. [Result](javascript-identity.json). |
| Upstream declarations | Upstream npm ci and unmodified Hereby dts pass; both adapted main and trial Hereby dts pass with the upstream locked install. |
| typescript.d.ts | Exactly four TypePredicate member readonly-kind modifiers, nothing else. [Exact diff](typescript.d.ts.diff). Main SHA-256 `f03adcf915df559d616fcafd75181d3423eaa0f272237482cedbd025525c9516`; trial `3f39a3407b11916b345ed5b4503bea71849c994c46e0ed43f3e9afe637708346`. Both artifact bytes retained compressed. |
| tsserverlibrary.d.ts | Byte-identical, both SHA-256 `cee9e9bf9a8f67388ee878afd5e172189421bae6625076b774d68563d036fbfd`; own diff file is empty. |
| Restoring mutant | Nine original interfaces restored in a separate source tree; all **747 source-tree files byte-identical to main**. Complete full generator census exits 0; its **523 original finding occurrences are restored exactly, including byte offsets**, and all seven cast sites are again Refused in both refusal_scan and lowering. The five cleared sites return; the two peekBlock sites had never cleared. [Result](restoring-mutant-result.json), [comparison](restoring-mutant-comparison.log). Both isolated TypePredicate proofs again refuse with kind as the first unsafe field; no full checker-mutant census claim. |
| Trial stage3/apply.sh and landing contract controls | Not run / not constructed after failed trial gates. Baseline chain completed through saved refs as described above; do not confuse that with a clean full trial apply pass. |
| Adapted-tree product under stock tsc | Not run after stop. Hereby dts success is not this product test. |
| Whole unfiltered stage3/lane | Not run after stop. No lane-green claim. |
| Sanctions / shard plan / ledger | Unchanged because no adaptation landed. No sanction, runtime hash, count or checker/oracle rule was weakened. |

The evidence branch is `typescript/cast-tags-evidence`, checkpoint-pushed without force. The final chat response names its final commit. The requested landing branches have no commit from this work. Nothing was submitted to Loom.
