# Unit 4: four checkpoint certificates

Merged `origin/stage1-hir/wip` checkpoint commit `9addb0e8` without rebasing.
Cohere pin: `7945d102a6c18dd36adf9114a758ce646e8b2359`. Exclusive directory: `passes/unit-4/`.
Read PLAN.md, PROGRESS.md, REPORT.md, the arena ruling at docs/memory.md:477,
and the complete shared arena_index.a. No shared file was edited.

## Certificate

| Independent Go before/after checkpoint | Source Node | Sanitized native | Emitted JavaScript |
| --- | ---: | ---: | ---: |
| Primitive property constraints | 1,465 / 1,465 | 1,465 / 1,465 | 1,465 / 1,465 |
| InferReactive | 1,465 / 1,465 | 1,465 / 1,465 | 1,465 / 1,465 |
| InferAliasingEffects | 1,465 / 1,465 | 1,465 / 1,465 | 1,465 / 1,465 |
| InferAliasingEffectsForNested | 1,465 / 1,465 | 1,465 / 1,465 | 1,465 / 1,465 |
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
**five source-Node**, **five emitted-JavaScript** and **five sanitized-native**
mutant comparisons caught. Each mutation executes successfully on the full census;
only the independent Go byte comparison catches it.
Definitions are in `mutants.json`; exact catches are in `validation/certificates.txt`.

The final owned package runs `TestUnit4Census`, `TestArrayNeverGap`,
`TestRecursiveInitializerGap` and `TestRetainedWitnesses`; its receipt is
`validation/certificates.txt`: **4 pass, 0 fail, 0 skip**, **977.088s**. Each top-level test
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

The user authorized the recorded empty-array workaround used by values gap 5
and YAML emptyAlternative. Typed constants replace `?? []`; separately initialized,
explicitly typed arrays precede conditional assignment. Every changed site carries
its gap 5 comment. The constants are only read, never mutated. See
`stage1/cohere/values/GAPS.md` (gap 5), `values/parser.ts` (`noChildren`) and
`stage1/cohere/yaml/GAPS.md` (`emptyAlternative`). `TestArrayNeverGap` uses
`errors.As` to select `*lower.NotYet` and checks `What == "an array of never"`;
it fails the moment the original assignment or coalesce fixture lowers, so the
workarounds can be removed. Node still prints `0` and `1` respectively.

Effects' recursive path search, module-hook resolver and fresh-identifier visitor
are ordinary module-level function declarations with explicit arguments. Their
recursion and traversal order are unchanged. No compiler workaround or shared
file change is involved. Local function declarations are not lowered by this
stage 0, so the declarations live at module scope.

The original local recursive arrow initializer remains in
`gaps/recursive-initializer.a`, naming compiler **#dv99xzy**. Node prints `0`.
`TestRecursiveInitializerGap` selects `*lower.NotYet` with
`What == "a function value that captures the variable its own initializer declares"`
and fails when the original shape closes or changes. The production effects
entrypoint is no longer expected to refuse. Go's root/nested effects entrypoints
are `cohere/internal/lint/ecmascript/high_level_intermediate_representation/effects.go:493`
and `:523`.

Projection still waits for the real Adamic mutation_aliasing import/API, as
instructed. Go `effects.go:1189 ProjectEffects` requires
`MutableRanges.Get(effect.Into.Identifier).End` at :1196. The lane does not invent
that module or substitute recorded ranges for its imported API. Once the module lands, merge the shared branch, rerun every backend and mutant, and
certify projection. This lane is not declared fully certified.

Final machine observation: nproc 5; load average 0.66, 0.95, 0.97.
