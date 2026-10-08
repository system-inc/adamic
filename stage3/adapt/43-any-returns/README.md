# Real any returns

Permanent type-only adaptation of five of the eight real-any callees reported in
`stage3/census/any-returns` at fb2b782a. Base: main
`3ffb1a835184713998a34874e86326cd21db971f`; upstream: TypeScript 6.0.3,
`050880ce59e30b356b686bd3144efe24f875ebc8`.

The five sites account for **6,947** credited hidden bytes in the ed6e2975
measurement. This is attribution from that measurement, not a new census or a
claim that these bytes now lower successfully. Two sites, totaling 254 bytes,
remain unchanged. No compiler code or public declaration is edited.

| Callee | Truthful return | Credited bytes | Decision |
| --- | --- | ---: | --- |
| convertConfigFileToObject | AdamicJsonRecoveryObject | 1,545 | Adapt |
| convertToObject | AdamicJsonRecoveryValue or undefined | 150 | Public declaration; decline |
| convertToJson | AdamicJsonRecoveryValue or undefined | 4,862 | Adapt |
| convertToJson.convertObjectLiteralExpressionToJson | AdamicJsonRecoveryObject or undefined | 0 | Adapt |
| convertToJson.convertPropertyValueToJson | AdamicJsonRecoveryValue or undefined | 0 | Adapt |
| setTimeout | NodeJS.Timeout | 89 | Covered by adaptation 41 |
| tryParseJson | Recursive strict JSON value or undefined | 104 | Consumer contracts; decline |
| objectAllocator.getNodeConstructor | ReturnType<ObjectAllocator["getNodeConstructor"]> | 540 | Adapt |

`AdamicJsonRecoveryValue` is string, number, boolean, null, an array of those
values, or `AdamicJsonRecoveryObject`. The object maps string keys to recovery
values **or undefined**: invalid property values are retained, while invalid
array elements are filtered out. Both declarations are module exports marked
`@internal`, keeping them out of the public declaration artifacts. Internal module exports also preserve the upstream namespace emitter's formatting.

The config root converter accepts only object roots; its recovery branch chooses
the first object in an array, otherwise returns `{}`. Its two type-only object
views follow those existing guards. The nested object accumulator becomes typed;
its non-null view follows the same `returnValue` condition that initialized it.
The property converter returns scalars even with `returnValue === false`;
only object/array construction is disabled. The fixtures cover that distinction.

Adaptations apply in number order. Adaptation 41 already gives the private
setTimeout declaration a truthful `handler: () => void` and `NodeJS.Timeout`
return. Adaptation 43 reads that composed tree and leaves the timer untouched;
it is removed from 43's site list and guard. The five remaining adapted sites
and two declined sites retain their exact source guards. The remaining
getNodeConstructor mutant still proves an unreviewed site is rejected before
any output is written. The existing public System API stays unchanged.
The allocator uses the existing ObjectAllocator constructor contract. The Node
function initializes the returned node; its preexisting `as any` bridge remains
because TypeScript does not give an ordinary function a construct signature.
The getter's own return signature is now precise.

Only **convertToObject** would require changing a public declaration to expose
its truthful return. It is deliberately unchanged. **tryParseJson** actually
returns primitives and null: `JSON.parse("7")` yields 7. Its readJsonOrUndefined
caller promises `object | undefined`, and package-json callers access properties
without narrowing. `declines.cjs` asks stock tsc to check the truthful union and
records the resulting consumer errors. Adding an object cast would conceal this
mismatch; adding runtime narrowing would violate the requested JavaScript byte
identity. Its signature therefore remains unchanged pending that consumer work.

## Reproduction

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/real-any-setup.log 2>&1
source /workspace/adamic-tools/env.sh
# In a detached checkout of the base main:
bash stage3/lane/run.sh /tmp/real-any-baseline-lane > /tmp/real-any-baseline-lane.log 2>&1
# In this branch, with new output directories:
NODE_OPTIONS=--max-old-space-size=1400 bash stage3/lane/run.sh /tmp/real-any-verified-lane > /tmp/real-any-verified-lane.log 2>&1
NODE_PATH=$HOME/.cache/adamic-stage3/api/node_modules node stage3/adapt/43-any-returns/proof.cjs /tmp/real-any-baseline-lane /tmp/real-any-verified-lane stage3/adapt/43-any-returns/evidence > /tmp/real-any-proof.log 2>&1
NODE_PATH=$HOME/.cache/adamic-stage3/api/node_modules node stage3/adapt/43-any-returns/declines.cjs /tmp/real-any-verified-lane/adapted-tree stage3/adapt/43-any-returns/evidence > /tmp/real-any-declines.log 2>&1
```

Each lane invokes the entire apply pipeline and `stage3/oracle/run.sh`, with all
upstream runners and eight workers. The permanent proof compares all emitted
JavaScript files and the public typescript declaration artifacts by SHA-256,
checks identical lane counts/verdict/API comparison and baseline-diff bytes,
runs recovery/timer/allocator fixtures on both builds, and checks idempotence.
It changes each of the seven remaining sites' `any` to `string` in scratch input;
the exact reviewed source contract rejects every mutant before writing output.
The two declined sites are guarded too. Separate JavaScript-byte, API-byte and
lane-count mutants prove the comparison checks can fail.

Initial comparisons caught the new aliases changing namespace declaration
formatting. Exporting them internally and marking them `@internal` corrected it. Two attempts lost test workers; the cgroup recorded OOM kills. The accepted run bounds each worker heap to 1,400 MiB inside the 16 GiB cgroup. Its partial result is not accepted as a
passing oracle. Final observations and output hashes are in `evidence/`.

No native lowering claim, public API repair, JSON-consumer narrowing, or additional
allocator getter adaptation is covered. These are adapter fixtures, so no
internal/oracle fixture or counts.md row was added.

## Historical observed result before batch 5

Both lanes pass: 106,366 passing, one failing, zero pending. The one failure is
main's existing public API acknowledgment test; both oracle baseline differences
are exactly `api/typescript.d.ts`, with identical diff bytes. Apply and build exit
zero; the upstream oracle exits one in both trees as expected by the lane.
All ten emitted JavaScript files and both public declaration artifacts have
identical hashes. Stock tsc reports zero diagnostics on both compiler programs;
six return signatures change from any to the types above, and the two declined
signatures stay any. Recovery, timer and allocator fixtures pass on both builds.
All eight site mutants and the three comparison mutants are caught; the adapter
is idempotent. Setup completed in 6.664 seconds; nproc is 5, CPU quota is 4.
Full setup timing lines, lane/oracle reports, compressed test logs, per-site
mutant failures, strict-JSON consumer diagnostics, and artifact hashes are saved
in `evidence/`.
