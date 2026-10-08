# Unit 4: stopped on shared checkpoint inputs

Base: `3714b3198397502297cd298d3d497835dcdeaa49`, fetched from
`origin/stage1-hir/wip`. Exclusive directory: `passes/unit-4/`.
Read PLAN.md, PROGRESS.md, REPORT.md, the arena ruling at docs/memory.md:477,
and the complete shared arena_index.a. No shared file was edited.

## Certificate

Native: **0 / 1,465**. Source Node: **0 / 1,465**.
Emitted JavaScript: **0 / 1,465**. Flow replay: **0 / 23**.
No pass comparison or semantic mutant was executed. These are unexecuted
observations, not passing or skipped tests. No implementation is certified.
No independent language gap was observed.

## Requests to hir-01

1. Land the checkpoint bundle framing and decoder under `replay/`, with
   canonical roundtrip and corrupt-index mutant. Export all 1,465 original census
   keys, including the 23 Flow graphs and error outcomes, as Go input. Preserve
   original nesting paths, source kind/span handles, function-local identities,
   nil-checker availability and table order. Provide before/after boundaries for
   `effects.go:493 InferAliasingEffects`, `:523 InferAliasingEffectsForNested`,
   `:1189 ProjectEffects`, and `reactive.go:189 InferReactive` independently.
2. Extend generic replay fact selectors for identifier source handles, hook and
   module-export origins, callee signature inputs and the exact type-name facts
   consumed by `reactive.go:662 stableTypeName` (`GetTypeAtLocation` at :670,
   alias at :674, symbol at :679). Input facts must remain separate from the
   expected reactive/effect answers. Preserve the original nil checker path.
   I will own alias-effect, primitive-constraint and reactive sidecar records;
   please publish the framing/import API before those records are implemented.
3. Land the existing Adamic `mutation_aliasing` module by reference and publish
   its exports corresponding to Go `AliasingEffect[P]`, `AliasingEffectKind`,
   `EffectValueKind`, `CreateEffect`, `FlowEffect`, `MutationEffect`, and
   `MutableRanges.Get`. Do not copy the Go analysis into this lane.
   `effects.go:1189 ProjectEffects` needs the actual imported mutable ranges;
   its aliasing branch at :1196 tests `ranges.Get(effect.Into.Identifier).End`.

No new arena class or instruction variant is requested yet: that would be
speculation before the checkpoint schema exists. Existing concrete index
classes remain the contract, with private minting and checked arena reads.
SSA will be imported from the existing module, never copied.

## Reproducer and stopping point

From the repository root, run:

```sh
git ls-tree -r --name-only HEAD stage1/cohere/high_level_intermediate_representation/replay
git ls-tree -r --name-only HEAD stage1/cohere/mutation_aliasing
```

Both commands produce no paths on the recorded base. `BLOCKERS.log` retains
this inventory. The construction printer is not an analysis checkpoint decoder;
using it would omit the very effects/reactive state this lane must compare.
Stopped before fabricating a private replay codec, checker stand-in or copied
mutation-aliasing implementation. Once the owner lands these inputs, merge the
shared branch without rebasing and resume the full isolated pass certificate.
