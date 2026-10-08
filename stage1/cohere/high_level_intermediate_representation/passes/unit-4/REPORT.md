# Unit 4: source-Node certificate, native array-of-never stop

Merged `origin/stage1-hir/wip` checkpoint commit `9addb0e8` without rebasing.
Cohere pin: `7945d102a6c18dd36adf9114a758ce646e8b2359`. Exclusive directory: `passes/unit-4/`.
Read PLAN.md, PROGRESS.md, REPORT.md, the arena ruling at docs/memory.md:477,
and the complete shared arena_index.a. No shared file was edited.

## Certificate

| Independent Go before/after checkpoint | Source Node | Sanitized native | Emitted JavaScript |
| --- | ---: | ---: | ---: |
| Primitive property constraints | 1,465 / 1,465 | 1,465 / 1,465 | 1,465 / 1,465 |
| InferReactive | 1,465 / 1,465 | 0, compiler refusal | 0, compiler refusal |
| InferAliasingEffects | 1,465 / 1,465 | 0, compiler refusal | 0, compiler refusal |
| InferAliasingEffectsForNested | 1,465 / 1,465 | 0, compiler refusal | 0, compiler refusal |
| ProjectEffects | 0, deferred import | 0, deferred import | 0, deferred import |

All 1,465 original census keys remain, including **23 / 23 Flow graphs** in each
executed matrix row. No original error outcome is filtered. The 72 construction
probes are exported but are not included in these counts. Comparisons include
the complete graph and canonical sidecars, not an instruction count or checksum.
The Go adapter invokes the actual pass on fresh Go state, writes immediately
before and after, and repeats the after encoding to check determinism. Effects
and reactive inference have separate input clones and retain nil checker state.
No earlier Adamic pass or expected analysis answer is read as an input fact.

`primitive.a` uses checked function/identifier handles as two-part constraint
keys. Effects retain sequence, from/into places, value kind, closure captures,
custom-hook signatures and callback mutation summaries. All alias-effect records
and primitive/reactive schemas are lane-owned. `input.unit4-ast` contains actual
identifier text and syntax callee/receiver names; signatures are ported literals,
not Go-computed effect answers. The shared replay module is imported by reference.
No SSA or mutation_aliasing algorithm is copied. No shared file is edited.

## Mutants and verification

Four semantic source changes are caught by Go byte comparison after successful
execution: primitive/other constraints swapped; parameter reactivity omitted;
closure capture changed to immutable capture; custom-hook frozen result changed
to mutable. The capture change is run separately for root and nested effects:
**five source-Node mutant comparisons caught**, plus the primitive mutant on
**sanitized native**. The blocked backends have no claimed mutant certificate.
Definitions are in `mutants.json`; exact catches are in `validation/certificates.txt`.

`TestUnit4Census` and `TestArrayNeverGap`: **2 pass, 0 fail, 0 skip**, 183.204s.
`TestRetainedWitnesses`: **1 pass, 0 fail, 0 skip**, 0.818s. Each top-level test
calls t.Parallel. gofmt and lane vet pass. The original Go census exporter retains
its upstream skipped tests; their classifications belong to the shared census,
not successful unit-4 observations. Retained fixtures are the Go checkpoints
that catch the mutants; provenance names their original upstream calls.

Reproduce from the repository root after sourcing the toolchain environment:

```sh
HIR_UNIT4_CENSUS=/tmp/hir-unit4-census go test -v -count=1 -timeout=40m ./stage1/cohere/high_level_intermediate_representation/passes/unit-4 > /tmp/hir-unit4-tests.log 2>&1
gofmt -l stage1/cohere/high_level_intermediate_representation/passes/unit-4
go vet ./stage1/cohere/high_level_intermediate_representation/passes/unit-4
```

## Requests and stopping point

The replay, record and checker-selector requests are answered by `9addb0e8`.
No further shared record, instruction variant or arena class is requested.

Native and emitted-JavaScript lowering refuse **an array of never**, despite the
source's contextual element type. No workaround is applied. The shortest inputs
are `gaps/array-never.a:2:58` (optional map result assigned an empty array) and
`gaps/array-never-coalesce.a:2:35` (optional map result iterated with an empty-array
fallback). Node prints `0` and `1` respectively. Both backends report exactly:

```text
stage 0 can't lower an array of never yet
```

The affected port sites are `frontiers.a:10:116` and `effects.a:67:116`.
They implement Go `postdominator.go:130`'s nil predecessor slice append and
`effects.go:1454`'s absent instruction effect-list traversal, respectively.
`TestArrayNeverGap` holds the exact refusal and fails when it closes. Full compiler
output is retained in `validation/`; this is a language fix request, not permission
to weaken a comparison. Reactive/effects native and emitted-JavaScript certificates
remain stopped, with their complete source-Node certificates kept separate.

Projection still waits for the real Adamic mutation_aliasing import/API, as
instructed. Go `effects.go:1189 ProjectEffects` requires
`MutableRanges.Get(effect.Into.Identifier).End` at :1196. The lane does not invent
that module or substitute recorded ranges for its imported API. Once the language
fix and module land, merge the shared branch, rerun every backend and mutant, and
certify projection. This lane is not declared fully certified.
