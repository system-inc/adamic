# Shared schema batch: stopped on compiler gap 3

Completed landing: **hir/land-08 e5ac6a3c**. React construction/clone/replay,
static-components and the CFG meet in one concrete arena home. Its complete
certificates are recorded in that branch's short REPORT.md. This branch merges
that landing without rebasing and preserves all planning/evidence.

## Requests supplied

Read latest reports: unit 3 a7e23a96; 4 6c380a18; 6 c9c38057; 7 44a7e743;
8 7521b125; 9 2de68d81; 10 20e126a4. Workers import `../../replay/index.ts`.
Scope, Sequence and MaybeThrow now have central typed records and codecs,
cloning/remapping and graph edges. `fn.addScope(goId)`, `scopeAt(goId)` and
`scopeId(index)` translate Go IDs through function-owned checked ScopeIndex
arenas; Scope 17 is deliberately not its slot. Missing exception handlers are
undefined and serialize as Go InvalidBlock 0. `finalizeHIR(arena,index)` calls
imported SSA reverse postorder, predecessors and evaluation order, recursively,
without reconstructing SSA. DependencyPathIndex/DependencyTreeIndex join the
single private-minting arena home. Raw `input.types` value-present and flags
facts, with InputFacts selectors, preserve checker/node/type absence separately.
Old frames round-trip; type-flag consumers must regenerate actual Go inputs.

Unit 3's persistent Instruction[] is replaced by InstructionArena columns and
checked InstructionIndex views. No new index class is necessary. Getter values
must be read into a local once before narrowing; unit 3 owns that adjustment
in optional_sources.ts:70. Its five Go-oracle mismatches remain #6d8y0pf.

## Current certificates and stop

Node construction/clone: **1,442/1,442 non-Flow originals +72 probes**.
Node fresh decode/dump/re-encode: **1,465/1,465 originals +72 probes**, all Flow.
Go/Node terminals, clone and existing-phi finalization: **15/15**; three Node
semantic mutants caught. Dependency indices and their private/brand/off-by-one
checks pass on native and Node. The exact accessor gap selector passes.
Native IR certificates are **compile stopped**, not matching certificates:
GAPS.md gap 3 retains the three-line input and exact lower.NotYet.What.
No workaround is applied. Mandatory native tests stay red until compiler fixes
accessor dispatch; no compiler refusal is counted as a semantic mutant.

**Declined:** mutation_aliasing import/API remains absent from latest lint area
bda1026b; no substitute is invented. ScopeIdentity, ScopeDependencies (including
temporaries) and reactive-tree record schemas belong to lanes 6, 7 and 8;
those lanes must publish their public codecs for 9/10. This batch supplies the
shared core contracts, not their analysis algorithms or expected answers.

Latest native allocation measurement is the finished landing's 12-function
probe: **4,087 allocations = 4,087 frees**, +875 from 3,212, -77 from 4,164.
Current view/column allocation delta is unmeasurable while gap 3 stops native;
no historical count is presented as a current shared-branch measurement.

---

The following bootstrap certificate is historical, before the schema/arena changes:

# Unit 2 complete and shared replay bootstrap

Roadmap steps 08 and 28. This finished unit is pushed once to `stage1-hir/wip`.
Cohere remains pinned at `7945d102a6c18dd36adf9114a758ce646e8b2359`; no parser
or Go production source is changed. The previous area merge is retained.

## Coverage

| Certificate | Native | Node |
| --- | ---: | ---: |
| Construction and cached ForFunction originals | 1,442 / 1,465 | 1,442 / 1,465 |
| Admitted non-Flow originals | 1,442 / 1,442 | 1,442 / 1,442 |
| CloneFunction admitted originals | 1,442 / 1,442 | 1,442 / 1,442 |
| Construction and clone path probes | 72 / 72 | 72 / 72 |
| Fresh-arena checkpoint decode / dump / re-encode originals | 1,465 / 1,465 | 1,465 / 1,465 |
| Checkpoint path probes | 72 / 72 | 72 / 72 |
| Static-components upstream / owned cases | 22 / 22 + 3 / 3 | 22 / 22 + 3 / 3 |

Every admitted constructed and cloned graph matches Go hir-v1 byte for byte.
The 23 Flow graphs remain excluded from AST construction and catalogued; they
are included in replay, which reads Go-produced input without parsing Flow.
The 45 original skipped Go tests retain the classifications in SKIPPED.md.
Static-components findings also match in emitted JavaScript.

The spread/static-constructor compiler ruling is applied as explicit typed
copies, with `gap 2 (GAPS.md)` comments. The shortest proving program and exact
selector-style refusal test remain: that test fails when the compiler fixes
lowering. Gap 1's presence/value workaround and proving test are also retained.
Neither ruling weakens private minting or checked arena reads.

## Shared pass contract

Workers import `../../replay/index.ts` from their pass directory. README.md in
replay/ documents the framing, identity reader, fresh decoder, output encoder,
AST/type/checker selectors and Go overlay hooks. Canonical hir-checkpoint-v1
bundles contain unchanged hir-v1 graphs plus ordered namespaced sidecars, linked
by nesting path and existing identifiers, blocks and instructions; scope and
reactive identities are registered explicitly. Nil and empty memo dependencies,
outlined identities, detached/nil blocks and allocator high-water marks survive
round-trip. Internal invalid UTF-8 checker-name bytes use reversible surrogate
escapes, preventing Node's replacement character from changing Go bytes.

All 43 Go instruction variants are central, including DeclareContext and
StartMemoize/FinishMemoize. ScopeIndex and ReactiveIndex have private minting and
checked reads in the single shared arena home. The decoder's JSON syntax graph
also uses a checked PayloadIndex arena. Copy/remap, instruction replacement,
existing declaration reuse and the distinct manual-memo cache contract are
public APIs. The cache invokes worker-supplied drop/inclusive-inline callbacks
at Go's boundaries; this bootstrap does not implement those algorithms.

Requests read: unit 3 STOPPED.md at 9cc68732 and unit 4 REPORT.md at 9519c348.
The framing, memo/outlined records, copy/rewrite/cache hooks, checkpoint decoder
and checker-fact selectors are supplied. Pass workers retain ownership of their
before/after exports, analysis sidecars and implementations; the shared Go
encoder and construction-observation hook connect their tagged adapters without
inventing a pass order. No pass algorithm's certificate is claimed here.

**Declined:** landing or inventing mutation_aliasing's port/API. Latest fetched
lint area 6bf7bcec still has no stage1/cohere/mutation_aliasing directory. That
module belongs to the existing analysis lane and must be imported, never copied.
Unit 4's dependent imports/projection and unit 5 need its actual landed API.
Later scope/reactive analysis record codecs remain worker-owned, with requests
for shared core schema changes routed to this owner.

## Mutants and local verification

Fresh certificates in validation/replay-bootstrap/ cover the full checkpoint
census on both runtimes, its corrupt instruction-index mutant, central Go memo
variants, typed declaration/copy/rewrite/cache contracts, the full construction
and clone census, clone storage-alias mutant, return-lowering mutant, FunctionIndex
mutant, both compiler gap tests, cache-hit mutant, and static-components creation
mutant. ScopeIndex and ReactiveIndex off-by-one, private-constructor and cross-brand
checks pass. The prior 79 / 79 HIR semantic matrix remains recorded in
validation/unit2-step28/; this push reruns focused mutants and full unmutated
corpus certificates rather than claiming a fresh run of all 79.

## Allocations

The same byte-identical 12-function native probe has **4,164 allocations and
4,164 frees**, **+952** from original 3,212, **+295** from shared-index 3,869,
and **+40** from typed-checkpoint 4,124. Retains/releases: **16,323 / 14,901**;
peak live: **240**; regions: **0**. Previous measurement was 4,161: this bootstrap
adds three index classes, but the +3 is not claimed as an isolated attribution.
Known causes of earlier growth are boxed concrete indices, canonical handle
arrays, owner-bound SSA callbacks and pattern node/handle arrays. The brand is
preserved. The old 4,300,331 full-corpus allocation measurement in the prior
certificate is historical; no current full-corpus delta is claimed.
